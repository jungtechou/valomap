package ctx

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// ContextKey type is used to avoid collisions with context values
type ContextKey string

// Common context keys
const (
	RequestIDKey ContextKey = "request_id"
	StartTimeKey ContextKey = "start_time"
	UserIDKey    ContextKey = "user_id"
	ClientIPKey  ContextKey = "client_ip"
)

// CTX extends Google's context to support logging and request tracking
type CTX struct {
	context.Context
	logrus.FieldLogger
	requestID string
}

// Background returns a non-nil, empty Context with a UUID as request ID
// It is typically used for initialization and as the top-level Context for incoming requests
func Background() CTX {
	requestID := uuid.New().String()
	logger := logrus.StandardLogger().WithField(string(RequestIDKey), requestID)

	ctx := context.Background()
	ctx = context.WithValue(ctx, RequestIDKey, requestID)
	ctx = context.WithValue(ctx, StartTimeKey, time.Now())

	return CTX{
		Context:     ctx,
		FieldLogger: logger,
		requestID:   requestID,
	}
}

// WithValue returns a copy of parent in which the value associated with key is val.
// It also adds the key-value pair to the logger for consistent tracing.
func WithValue(parent CTX, k ContextKey, v interface{}) CTX {
	return CTX{
		Context:     context.WithValue(parent.Context, k, v),
		FieldLogger: parent.FieldLogger.WithField(string(k), v),
		requestID:   parent.requestID,
	}
}

// WithValues returns a copy of parent with multiple key-value pairs added
func WithValues(parent CTX, kvs map[string]interface{}) CTX {
	ctx := parent.Context
	logger := parent.FieldLogger

	for k, v := range kvs {
		ctx = context.WithValue(ctx, ContextKey(k), v)
		logger = logger.WithField(k, v)
	}

	return CTX{
		Context:     ctx,
		FieldLogger: logger,
		requestID:   parent.requestID,
	}
}

// WithCancel returns a copy of parent with added cancel function
func WithCancel(parent CTX) (CTX, context.CancelFunc) {
	newCtx, cFunc := context.WithCancel(parent.Context)
	return CTX{
		Context:     newCtx,
		FieldLogger: parent.FieldLogger,
		requestID:   parent.requestID,
	}, cFunc
}

// WithTimeout returns a copy of parent with timeout condition and cancel function
func WithTimeout(parent CTX, d time.Duration) (CTX, context.CancelFunc) {
	newCtx, cFunc := context.WithTimeout(parent.Context, d)
	return CTX{
		Context:     newCtx,
		FieldLogger: parent.FieldLogger.WithField("timeout_ms", d.Milliseconds()),
		requestID:   parent.requestID,
	}, cFunc
}

// WithDeadline returns a copy of parent with deadline condition and cancel function
func WithDeadline(parent CTX, deadline time.Time) (CTX, context.CancelFunc) {
	newCtx, cFunc := context.WithDeadline(parent.Context, deadline)
	return CTX{
		Context:     newCtx,
		FieldLogger: parent.FieldLogger.WithField("deadline", deadline),
		requestID:   parent.requestID,
	}, cFunc
}

// RequestID returns the request ID from the context
func (c CTX) RequestID() string {
	return c.requestID
}

// ElapsedTime returns the elapsed time since the context was created
func (c CTX) ElapsedTime() time.Duration {
	startTime, ok := c.Value(StartTimeKey).(time.Time)
	if !ok {
		return 0
	}
	return time.Since(startTime)
}

// LogElapsed logs the elapsed time with a custom message
func (c CTX) LogElapsed(msg string) {
	c.WithField("elapsed_ms", c.ElapsedTime().Milliseconds()).Info(msg)
}

// FromContext creates a CTX from a standard context.Context
// If the context contains a request ID, it will be used
func FromContext(ctx context.Context) CTX {
	requestID, _ := ctx.Value(RequestIDKey).(string)
	if requestID == "" {
		requestID = uuid.New().String()
	}

	logger := logrus.StandardLogger().WithField(string(RequestIDKey), requestID)

	return CTX{
		Context:     ctx,
		FieldLogger: logger,
		requestID:   requestID,
	}
}
