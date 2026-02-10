package servers

import (
	"context"
	"log"

	"github.com/c12s/oort/internal/domain"
	"github.com/c12s/oort/internal/mappers/proto"
	"github.com/c12s/oort/internal/services"
	"github.com/c12s/oort/pkg/api"
	"github.com/c12s/oort/pkg/messaging"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type AdministratorAsyncServer struct {
	service    services.AdministrationService
	publisher  messaging.Publisher
	subscriber messaging.Subscriber
}

func NewAdministratorAsyncServer(subscriber messaging.Subscriber, publisher messaging.Publisher, service services.AdministrationService) (*AdministratorAsyncServer, error) {
	return &AdministratorAsyncServer{
		service:    service,
		publisher:  publisher,
		subscriber: subscriber,
	}, nil
}

func (s *AdministratorAsyncServer) Serve() error {
	return s.subscriber.Subscribe(s.serve)
}

func (s *AdministratorAsyncServer) serve(
	ctx context.Context,
	adminReqMarshalled []byte,
	replySubject string,
) {
	tracer := otel.Tracer("oort-admin-async-server")

	ctx, span := tracer.Start(
		ctx,
		"Handle Administration Request",
		trace.WithAttributes(
			attribute.String("messaging.system", "nats"),
			attribute.String("messaging.operation", "process"),
			attribute.String("reply.subject", replySubject),
		),
	)
	defer span.End()

	adminReq := &api.AdministrationAsyncReq{}
	if err := adminReq.Unmarshal(adminReqMarshalled); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return
	}

	var domainResp domain.AdministrationResp

	switch adminReq.Kind {

	case api.AdministrationAsyncReq_CreateResource:
		req := &api.CreateResourceReq{}
		if err := req.Unmarshal(adminReq.ReqMarshalled); err != nil {
			span.RecordError(err)
			return
		}

		reqDomain, err := proto.CreateResourceReqToDomain(req)
		if err != nil {
			span.RecordError(err)
			return
		}

		domainResp = s.service.CreateResource(ctx, *reqDomain)

	case api.AdministrationAsyncReq_DeleteResource:
		req := &api.DeleteResourceReq{}
		if err := req.Unmarshal(adminReq.ReqMarshalled); err != nil {
			span.RecordError(err)
			return
		}

		reqDomain, err := proto.DeleteResourceReqToDomain(req)
		if err != nil {
			span.RecordError(err)
			return
		}

		domainResp = s.service.DeleteResource(ctx, *reqDomain)

	case api.AdministrationAsyncReq_PutAttribute:
		req := &api.PutAttributeReq{}
		if err := req.Unmarshal(adminReq.ReqMarshalled); err != nil {
			span.RecordError(err)
			return
		}

		reqDomain, err := proto.PutAttributeReqToDomain(req)
		if err != nil {
			span.RecordError(err)
			return
		}

		domainResp = s.service.PutAttribute(ctx, *reqDomain)

	case api.AdministrationAsyncReq_DeleteAttribute:
		req := &api.DeleteAttributeReq{}
		if err := req.Unmarshal(adminReq.ReqMarshalled); err != nil {
			span.RecordError(err)
			return
		}

		reqDomain, err := proto.DeleteAttributeReqToDomain(req)
		if err != nil {
			span.RecordError(err)
			return
		}

		domainResp = s.service.DeleteAttribute(ctx, *reqDomain)

	case api.AdministrationAsyncReq_CreateInheritanceRel:
		req := &api.CreateInheritanceRelReq{}
		if err := req.Unmarshal(adminReq.ReqMarshalled); err != nil {
			span.RecordError(err)
			return
		}

		reqDomain, err := proto.CreateInheritanceRelReqToDomain(req)
		if err != nil {
			span.RecordError(err)
			return
		}

		domainResp = s.service.CreateInheritanceRel(ctx, *reqDomain)

	case api.AdministrationAsyncReq_DeleteInheritanceRel:
		req := &api.DeleteInheritanceRelReq{}
		if err := req.Unmarshal(adminReq.ReqMarshalled); err != nil {
			span.RecordError(err)
			return
		}

		reqDomain, err := proto.DeleteInheritanceRelReqToDomain(req)
		if err != nil {
			span.RecordError(err)
			return
		}

		domainResp = s.service.DeleteInheritanceRel(ctx, *reqDomain)

	case api.AdministrationAsyncReq_CreatePolicy:
		req := &api.CreatePolicyReq{}
		if err := req.Unmarshal(adminReq.ReqMarshalled); err != nil {
			span.RecordError(err)
			return
		}

		reqDomain, err := proto.CreatePolicyReqToDomain(req)
		if err != nil {
			span.RecordError(err)
			return
		}

		domainResp = s.service.CreatePolicy(ctx, *reqDomain)

	case api.AdministrationAsyncReq_DeletePolicy:
		req := &api.DeletePolicyReq{}
		if err := req.Unmarshal(adminReq.ReqMarshalled); err != nil {
			span.RecordError(err)
			return
		}

		reqDomain, err := proto.DeletePolicyReqToDomain(req)
		if err != nil {
			span.RecordError(err)
			return
		}

		domainResp = s.service.DeletePolicy(ctx, *reqDomain)

	default:
		span.SetStatus(codes.Error, "unknown administration request kind")
		return
	}

	resp, err := proto.AdministrationAsyncRespFromDomain(domainResp)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return
	}

	respMarshalled, err := resp.Marshal()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return
	}

	if err := s.publisher.Publish(ctx, respMarshalled, replySubject); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
}

func (s *AdministratorAsyncServer) GracefulStop() {
	err := s.subscriber.Unsubscribe()
	if err != nil {
		log.Println(err)
	}
}
