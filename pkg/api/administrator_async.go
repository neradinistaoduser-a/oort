package api

import (
	context "context"
	"fmt"

	"github.com/c12s/oort/pkg/messaging"
	"github.com/c12s/oort/pkg/messaging/nats"
	natsgo "github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type AdministrationAsyncClient struct {
	publisher         messaging.Publisher
	subscriberFactory func(subject string) messaging.Subscriber
}

func NewAdministrationAsyncClient(natsAddress string) (*AdministrationAsyncClient, error) {
	conn, err := natsgo.Connect(fmt.Sprintf("nats://%s", natsAddress))
	if err != nil {
		return nil, err
	}
	publisher, err := nats.NewPublisher(conn)
	if err != nil {
		return nil, err
	}
	subscriberFactory := func(subject string) messaging.Subscriber {
		subscriber, _ := nats.NewSubscriber(conn, subject, "")
		return subscriber
	}
	return &AdministrationAsyncClient{
		publisher:         publisher,
		subscriberFactory: subscriberFactory,
	}, nil
}

func (n *AdministrationAsyncClient) SendRequest(
	ctx context.Context,
	req AdministrationReq,
	callback AdministrationCallback,
) error {

	tracer := otel.Tracer("oort-admin-async-client")

	ctx, span := tracer.Start(
		ctx,
		"Send Administration Request",
		trace.WithAttributes(
			attribute.String("messaging.system", "nats"),
			attribute.String("messaging.destination", AdministrationReqSubject),
			attribute.String("admin.kind", req.Kind().String()),
		),
	)
	defer span.End()

	reqMarshalled, err := req.Marshal()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	adminReq := &AdministrationAsyncReq{
		Kind:          req.Kind(),
		ReqMarshalled: reqMarshalled,
	}

	adminReqMarshalled, err := adminReq.Marshal()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	replySubject := n.publisher.GenerateReplySubject()

	subscriber := n.subscriberFactory(replySubject)
	err = subscriber.Subscribe(func(ctx context.Context, msg []byte, _ string) {
		resp := &AdministrationAsyncResp{}
		if err := resp.Unmarshal(msg); err != nil {
			span.RecordError(err)
			return
		}
		callback(resp)
	})
	if err != nil {
		span.RecordError(err)
		return err
	}

	err = n.publisher.Request(
		ctx,
		adminReqMarshalled,
		AdministrationReqSubject,
		replySubject,
	)
	if err != nil {
		_ = subscriber.Unsubscribe()
		span.RecordError(err)
		return err
	}

	return nil
}

type AdministrationCallback func(resp *AdministrationAsyncResp)
