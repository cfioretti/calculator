package logging

import (
	"context"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	CorrelationIDKey = "correlation_id"
	ServiceNameKey   = "service_name"
	VersionKey       = "version"

	CorrelationIDMetadataKey = "x-correlation-id"
)

type Logger struct {
	*logrus.Logger
	serviceName string
	version     string
}

func NewLogger(serviceName, version string) *Logger {
	logger := logrus.New()

	logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339,
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "timestamp",
			logrus.FieldKeyLevel: "level",
			logrus.FieldKeyMsg:   "message",
			logrus.FieldKeyFunc:  "function",
			logrus.FieldKeyFile:  "file",
		},
	})

	level := os.Getenv("LOG_LEVEL")
	if level == "" {
		level = "info"
	}

	logLevel, err := logrus.ParseLevel(level)
	if err != nil {
		logLevel = logrus.InfoLevel
	}
	logger.SetLevel(logLevel)

	logger.SetOutput(os.Stdout)

	return &Logger{
		Logger:      logger,
		serviceName: serviceName,
		version:     version,
	}
}

func (l *Logger) WithContext(ctx context.Context) *logrus.Entry {
	entry := l.Logger.WithFields(logrus.Fields{
		ServiceNameKey: l.serviceName,
		VersionKey:     l.version,
	})

	if correlationID := ctx.Value(CorrelationIDKey); correlationID != nil {
		entry = entry.WithField(CorrelationIDKey, correlationID)
	}

	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if correlationIDs := md.Get(CorrelationIDMetadataKey); len(correlationIDs) > 0 {
			entry = entry.WithField(CorrelationIDKey, correlationIDs[0])
		}
	}

	return entry
}

func (l *Logger) WithField(key string, value interface{}) *logrus.Entry {
	return l.Logger.WithFields(logrus.Fields{
		ServiceNameKey: l.serviceName,
		VersionKey:     l.version,
		key:            value,
	})
}

func (l *Logger) WithFields(fields logrus.Fields) *logrus.Entry {
	fields[ServiceNameKey] = l.serviceName
	fields[VersionKey] = l.version
	return l.Logger.WithFields(fields)
}

func (l *Logger) WithError(err error) *logrus.Entry {
	return l.Logger.WithFields(logrus.Fields{
		ServiceNameKey: l.serviceName,
		VersionKey:     l.version,
		"error":        err.Error(),
	})
}

func GetCorrelationID(ctx context.Context) string {
	if correlationID := ctx.Value(CorrelationIDKey); correlationID != nil {
		return correlationID.(string)
	}

	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if correlationIDs := md.Get(CorrelationIDMetadataKey); len(correlationIDs) > 0 {
			return correlationIDs[0]
		}
	}

	return uuid.New().String()
}

func NewContextWithCorrelationID(ctx context.Context, correlationID string) context.Context {
	return context.WithValue(ctx, CorrelationIDKey, correlationID)
}

func (l *Logger) GRPCUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		correlationID := GetCorrelationID(ctx)

		ctx = NewContextWithCorrelationID(ctx, correlationID)

		start := time.Now()
		l.WithFields(logrus.Fields{
			CorrelationIDKey: correlationID,
			"method":         info.FullMethod,
			"type":           "grpc_request",
		}).Info("gRPC request started")

		resp, err := handler(ctx, req)

		duration := time.Since(start)
		logEntry := l.WithFields(logrus.Fields{
			CorrelationIDKey: correlationID,
			"method":         info.FullMethod,
			"type":           "grpc_request",
			"duration_ms":    duration.Milliseconds(),
		})

		if err != nil {
			logEntry.WithError(err).Error("gRPC request failed")
		} else {
			logEntry.Info("gRPC request completed")
		}

		return resp, err
	}
}

func (l *Logger) GRPCStreamInterceptor() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		correlationID := GetCorrelationID(ss.Context())

		l.WithFields(logrus.Fields{
			CorrelationIDKey: correlationID,
			"method":         info.FullMethod,
			"type":           "grpc_stream",
		}).Info("gRPC stream started")

		err := handler(srv, ss)

		logEntry := l.WithFields(logrus.Fields{
			CorrelationIDKey: correlationID,
			"method":         info.FullMethod,
			"type":           "grpc_stream",
		})

		if err != nil {
			logEntry.WithError(err).Error("gRPC stream failed")
		} else {
			logEntry.Info("gRPC stream completed")
		}

		return err
	}
}
