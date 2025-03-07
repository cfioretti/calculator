package test

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/cfioretti/calculator/pkg/application"
	grpcServer "github.com/cfioretti/calculator/pkg/infrastructure/grpc"
	pb "github.com/cfioretti/calculator/pkg/infrastructure/grpc/proto/generated"
)

func TestRealNetworkConnection(t *testing.T) {
	lis, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)
	port := lis.Addr().(*net.TCPAddr).Port
	lis.Close()

	go func() {
		lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
		if err != nil {
			t.Errorf("Failed to listen: %v", err)
			return
		}

		calculatorService := application.NewCalculatorService()
		calculatorServiceServer := grpcServer.NewServer(calculatorService)
		grpcNewServer := grpc.NewServer()
		pb.RegisterDoughCalculatorServer(grpcNewServer, calculatorServiceServer)

		if err := grpcNewServer.Serve(lis); err != nil {
			t.Errorf("Failed to serve: %v", err)
		}
	}()
	time.Sleep(100 * time.Millisecond)

	conn, err := grpc.NewClient(
		fmt.Sprintf("localhost:%d", port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	defer conn.Close()

	client := pb.NewDoughCalculatorClient(conn)

	edge := int32(20)
	request := &pb.PansRequest{
		Pans: &pb.PansProto{
			Pans: []*pb.PanProto{
				{
					Shape: "square",
					Measures: &pb.MeasuresProto{
						Edge: &edge,
					},
					Name: "square 20 cm",
					Area: 400.0,
				},
			},
			TotalArea: 400.0,
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	response, err := client.TotalDoughWeightByPans(ctx, request)
	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, request.Pans.TotalArea, response.Pans.TotalArea)
}
