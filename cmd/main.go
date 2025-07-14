package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"

	"github.com/cfioretti/calculator/internal/infrastructure/grpc/middleware"
	httpHandlers "github.com/cfioretti/calculator/internal/infrastructure/http"
	"github.com/cfioretti/calculator/internal/infrastructure/logging"
	prometheusMetrics "github.com/cfioretti/calculator/internal/infrastructure/metrics"
	"github.com/cfioretti/calculator/internal/infrastructure/tracing"
	"github.com/cfioretti/calculator/pkg/application"
	grpcServer "github.com/cfioretti/calculator/pkg/infrastructure/grpc"
	pb "github.com/cfioretti/calculator/pkg/infrastructure/grpc/proto/generated"
)

const (
	defaultGRPCPort = ":50051"
	defaultHTTPPort = ":8080"
	serviceName     = "calculator"
	version         = "1.0.0"
)

var logger *logging.Logger

func main() {
	logger = logging.NewLogger(serviceName, version)

	ctx := context.Background()
	logger.WithContext(ctx).Info("Starting calculator service")

	if err := tracing.InitTracing(nil); err != nil {
		logger.WithError(err).Fatal("Failed to initialize tracing")
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := tracing.ShutdownTracing(ctx); err != nil {
			logger.WithError(err).Error("Failed to shutdown tracing")
		}
	}()

	prometheusMetrics := prometheusMetrics.NewPrometheusMetrics()

	grpcPort := getGRPCPort()
	httpPort := getHTTPPort()
	logger.WithField("grpc_port", grpcPort).WithField("http_port", httpPort).Info("Server configuration loaded")

	calculatorService := application.NewCalculatorService()
	server := grpcServer.NewServer(calculatorService)

	metricsMiddleware := middleware.NewMetricsMiddleware(prometheusMetrics, prometheusMetrics)

	grpcInstance := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.UnaryInterceptor(func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
			loggedHandler := logger.GRPCUnaryInterceptor()
			metricsHandler := metricsMiddleware.UnaryServerInterceptor()

			return loggedHandler(ctx, req, info, func(ctx context.Context, req interface{}) (interface{}, error) {
				return metricsHandler(ctx, req, info, handler)
			})
		}),
		grpc.StreamInterceptor(func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
			loggedHandler := logger.GRPCStreamInterceptor()
			metricsHandler := metricsMiddleware.StreamServerInterceptor()

			return loggedHandler(srv, ss, info, func(srv interface{}, stream grpc.ServerStream) error {
				return metricsHandler(srv, stream, info, handler)
			})
		}),
	)

	pb.RegisterDoughCalculatorServer(grpcInstance, server)
	logger.Info("gRPC service registered successfully")

	httpServer := setupHTTPServer(httpPort)
	go func() {
		logger.WithField("port", httpPort).Info("HTTP server starting")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Fatal("HTTP server failed")
		}
	}()

	lis, err := net.Listen("tcp", grpcPort)
	if err != nil {
		logger.WithError(err).Fatal("Failed to listen on gRPC port")
	}

	go handleShutdown(grpcInstance, httpServer)

	logger.WithField("port", grpcPort).Info("gRPC server starting")
	if err := grpcInstance.Serve(lis); err != nil {
		logger.WithError(err).Fatal("Failed to serve gRPC server")
	}
}

func getGRPCPort() string {
	port := os.Getenv("GRPC_PORT")
	if port == "" {
		logger.WithField("default_port", defaultGRPCPort).Info("Using default gRPC port")
		return defaultGRPCPort
	}

	fullPort := ":" + port
	logger.WithField("configured_port", fullPort).Info("Using configured gRPC port")
	return fullPort
}

func getHTTPPort() string {
	port := os.Getenv("HTTP_PORT")
	if port == "" {
		logger.WithField("default_port", defaultHTTPPort).Info("Using default HTTP port")
		return defaultHTTPPort
	}

	fullPort := ":" + port
	logger.WithField("configured_port", fullPort).Info("Using configured HTTP port")
	return fullPort
}

func setupHTTPServer(port string) *http.Server {
	mux := http.NewServeMux()

	metricsHandler := httpHandlers.NewMetricsHandler()
	metricsHandler.RegisterRoutes(mux)

	healthHandler := httpHandlers.NewHealthHandler()
	healthHandler.RegisterRoutes(mux)

	return &http.Server{
		Addr:    port,
		Handler: mux,
	}
}

func handleShutdown(grpcServer *grpc.Server, httpServer *http.Server) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	sig := <-sigCh
	logger.WithField("signal", sig.String()).Info("Received shutdown signal")

	logger.Info("Shutting down servers gracefully...")

	// Shutdown HTTP server
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		logger.WithError(err).Error("HTTP server shutdown error")
	}

	// Shutdown gRPC server
	grpcServer.GracefulStop()
	logger.Info("Servers stopped successfully")
}
