package startup

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"sync"

	"github.com/c12s/oort/internal/configs"
	"github.com/c12s/oort/internal/domain"
	"github.com/c12s/oort/internal/repos/rhabac/neo4j"
	"github.com/c12s/oort/internal/servers"
	"github.com/c12s/oort/internal/services"
	"github.com/c12s/oort/pkg/api"
	"github.com/c12s/oort/pkg/messaging"
	"github.com/c12s/oort/pkg/messaging/nats"
	natsgo "github.com/nats-io/nats.go"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type app struct {
	config                    configs.Config
	grpcServer                *grpc.Server
	administratorAsyncServer  *servers.AdministratorAsyncServer
	administratorGrpcServer   api.OortAdministratorServer
	evaluatorGrpcServer       api.OortEvaluatorServer
	administrationService     *services.AdministrationService
	evaluationService         *services.EvaluationService
	publisher                 messaging.Publisher
	administratorSubscriber   messaging.Subscriber
	rhabacRepo                domain.RHABACRepo
	shutdownProcesses         []func()
	gracefulShutdownProcesses []func(wg *sync.WaitGroup)
}

func NewAppWithConfig(config configs.Config) (*app, error) {
	if config == nil {
		return nil, errors.New("config is nil")
	}
	return &app{
		config:                    config,
		shutdownProcesses:         make([]func(), 0),
		gracefulShutdownProcesses: make([]func(wg *sync.WaitGroup), 0),
	}, nil
}

func (a *app) Start() error {
	ctx := context.Background()
	shutdownTracing := initTracing(
		ctx,
		os.Getenv("JAEGER_HOST")+":"+os.Getenv("JAEGER_GRPC_PORT"),
	)
	a.shutdownProcesses = append(a.shutdownProcesses, shutdownTracing)

	a.init()

	err := a.startAdministratorAsyncServer()
	if err != nil {
		return err
	}
	return a.startGrpcServer()
}

func (a *app) GracefulStop(ctx context.Context) {
	// call all shutdown processes after a timeout or graceful shutdown processes completion
	defer a.shutdown()

	// wait for all graceful shutdown processes to complete
	wg := &sync.WaitGroup{}
	wg.Add(len(a.gracefulShutdownProcesses))

	for _, gracefulShutdownProcess := range a.gracefulShutdownProcesses {
		go gracefulShutdownProcess(wg)
	}

	// notify when graceful shutdown processes are done
	gracefulShutdownDone := make(chan struct{})
	go func() {
		wg.Wait()
		gracefulShutdownDone <- struct{}{}
	}()

	// wait for graceful shutdown processes to complete or for ctx timeout
	select {
	case <-ctx.Done():
		log.Println("ctx timeout ... shutting down")
	case <-gracefulShutdownDone:
		log.Println("app gracefully stopped")
	}
}

func (a *app) init() {

	natsConn, err := newNatsConn(a.config.Nats().Uri())
	if err != nil {
		log.Fatalln(err)
	}
	a.shutdownProcesses = append(a.shutdownProcesses, func() {
		log.Println("closing nats conn")
		natsConn.Close()
	})

	manager, err := neo4j.NewTransactionManager(
		a.config.Neo4j().Uri(),
		a.config.Neo4j().DbName())
	if err != nil {
		log.Fatalln(err)
	}
	a.shutdownProcesses = append(a.shutdownProcesses, func() {
		log.Println("closing neo4j conn")
		manager.Stop()
	})

	a.initNatsPublisher(natsConn)
	a.initAdministrationNatsSubscriber(natsConn)

	a.initRhabacNeo4jRepo(manager)

	a.initAdministratorService()
	a.initEvaluatorService()

	a.initAdministratorAsyncServer()
	a.initAdministratorGrpcServer()
	a.initEvaluatorGrpcServer()
	a.initGrpcServer()
}

func (a *app) initGrpcServer() {
	if a.administratorGrpcServer == nil {
		log.Fatalln("admin grpc server is nil")
	}
	if a.evaluatorGrpcServer == nil {
		log.Fatalln("eval grpc server is nil")
	}
	s := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
	)
	api.RegisterOortAdministratorServer(s, a.administratorGrpcServer)
	api.RegisterOortEvaluatorServer(s, a.evaluatorGrpcServer)
	reflection.Register(s)
	a.grpcServer = s
}

func (a *app) initAdministratorGrpcServer() {
	if a.administrationService == nil {
		log.Fatalln("admin service is nil")
	}
	server, err := servers.NewOortAdministratorGrpcServer(*a.administrationService)
	if err != nil {
		log.Fatalln(err)
	}
	a.administratorGrpcServer = server
}

func (a *app) initEvaluatorGrpcServer() {
	if a.evaluationService == nil {
		log.Fatalln("eval service is nil")
	}
	server, err := servers.NewOortEvaluatorGrpcServer(*a.evaluationService)
	if err != nil {
		log.Fatalln(err)
	}
	a.evaluatorGrpcServer = server
}

func (a *app) initAdministratorAsyncServer() {
	if a.administrationService == nil {
		log.Fatalln("admin service is nil")
	}
	if a.publisher == nil {
		log.Fatalln("publisher is nil")
	}
	if a.administratorSubscriber == nil {
		log.Fatalln("administration subscriber is nil")
	}
	server, err := servers.NewAdministratorAsyncServer(a.administratorSubscriber, a.publisher, *a.administrationService)
	if err != nil {
		log.Fatalln(err)
	}
	a.administratorAsyncServer = server
}

func (a *app) initEvaluatorService() {
	if a.rhabacRepo == nil {
		log.Fatalln("rhabac repo is nil")
	}
	evaluatorService, err := services.NewEvaluationService(a.rhabacRepo)
	if err != nil {
		log.Fatalln(err)
	}
	a.evaluationService = evaluatorService
}

func (a *app) initAdministratorService() {
	if a.rhabacRepo == nil {
		log.Fatalln("rhabac repo is nil")
	}
	administratorService, err := services.NewAdministrationService(a.rhabacRepo)
	if err != nil {
		log.Fatalln(err)
	}
	a.administrationService = administratorService
}

func (a *app) initNatsPublisher(conn *natsgo.Conn) {
	publisher, err := nats.NewPublisher(conn)
	if err != nil {
		log.Fatalln(err)
	}
	a.publisher = publisher
}

func (a *app) initAdministrationNatsSubscriber(conn *natsgo.Conn) {
	administrationSubscriber, err := nats.NewSubscriber(conn, api.AdministrationReqSubject, "oort")
	if err != nil {
		log.Fatalln(err)
	}
	a.administratorSubscriber = administrationSubscriber
}

func (a *app) initRhabacNeo4jRepo(manager *neo4j.TransactionManager) {
	a.rhabacRepo = neo4j.NewRHABACRepo(manager, neo4j.NewSimpleCypherFactory())
}

func (a *app) startAdministratorAsyncServer() error {
	err := a.administratorAsyncServer.Serve()
	if err != nil {
		return err
	}
	a.gracefulShutdownProcesses = append(a.gracefulShutdownProcesses, func(wg *sync.WaitGroup) {
		a.administratorAsyncServer.GracefulStop()
		log.Println("registration server gracefully stopped")
		wg.Done()
	})
	return nil
}

func (a *app) startGrpcServer() error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", a.config.Server().Port()))
	if err != nil {
		return err
	}
	go func() {
		log.Printf("server listening at %v", lis.Addr())
		if err := a.grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()
	a.gracefulShutdownProcesses = append(a.gracefulShutdownProcesses, func(wg *sync.WaitGroup) {
		a.grpcServer.GracefulStop()
		log.Println("oort server gracefully stopped")
		wg.Done()
	})
	return nil
}

func (a *app) shutdown() {
	for _, shutdownProcess := range a.shutdownProcesses {
		shutdownProcess()
	}
}
