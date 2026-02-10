package messaging

import (
	"context"

	"go.opentelemetry.io/otel/propagation"
)

var Propagator = propagation.TraceContext{}

type Subscriber interface {
	Subscribe(handler func(ctx context.Context, msg []byte, replySubject string)) error
	Unsubscribe() error
}

type Publisher interface {
	Publish(ctx context.Context, msg []byte, subject string) error
	Request(ctx context.Context, msg []byte, subject, replySubject string) error
	GenerateReplySubject() string
}
