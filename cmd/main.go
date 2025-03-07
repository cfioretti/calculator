package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"

	"github.com/cfioretti/calculator/pkg/application"
	grpcServer "github.com/cfioretti/calculator/pkg/infrastructure/grpc"
	pb "github.com/cfioretti/calculator/pkg/infrastructure/grpc/proto/generated"
)

const defaultPort = ":50051"

func main() {
	port := getPort()
	calculatorService := application.NewCalculatorService()
	server := grpcServer.NewServer(calculatorService)
	grpcInstance := grpc.NewServer()

	pb.RegisterDoughCalculatorServer(grpcInstance, server)

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	go handleShutdown(grpcInstance)

	log.Printf("Server listening on %s", port)
	if err := grpcInstance.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

func getPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		return defaultPort
	}
	return ":" + port
}

func handleShutdown(server *grpc.Server) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	<-sigCh
	log.Println("Shutting down server...")
	server.GracefulStop()
	log.Println("Server stopped")
}
