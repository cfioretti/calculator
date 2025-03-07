package grpc_test

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	"github.com/cfioretti/calculator/pkg/application"
	grpcServer "github.com/cfioretti/calculator/pkg/infrastructure/grpc"
	pb "github.com/cfioretti/calculator/pkg/infrastructure/grpc/proto/generated"
)

const bufSize = 1024 * 1024

// setup in memory gRPC Server
func setupGRPCServer(t *testing.T) (*grpc.ClientConn, func()) {
	lis := bufconn.Listen(bufSize)

	calculatorService := application.NewCalculatorService()
	server := grpcServer.NewServer(calculatorService)
	grpcNewServer := grpc.NewServer()

	pb.RegisterDoughCalculatorServer(grpcNewServer, server)
	go func() {
		if err := grpcNewServer.Serve(lis); err != nil {
			t.Errorf("Failed to serve: %v", err)
			return
		}
	}()

	dialer := func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	conn, err := grpc.DialContext(
		ctx,
		"bufnet",
		grpc.WithContextDialer(dialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)

	cleanup := func() {
		err := conn.Close()
		if err != nil {
			return
		}
		grpcNewServer.Stop()
	}
	return conn, cleanup
}

func TestTotalDoughWeightByPans(t *testing.T) {
	conn, cleanup := setupGRPCServer(t)
	defer cleanup()

	client := pb.NewDoughCalculatorClient(conn)

	testCases := []struct {
		name     string
		input    *pb.PansProto
		expected *pb.PansProto
	}{
		{
			name: "Single round pan",
			input: &pb.PansProto{
				Pans: []*pb.PanProto{
					{
						Shape: "round",
						Measures: &pb.MeasuresProto{
							Diameter: func() *int32 { d := int32(28); return &d }(),
						},
						Name: "round 28 cm",
						Area: 615.75,
					},
				},
				TotalArea: 615.75,
			},
			expected: &pb.PansProto{
				Pans: []*pb.PanProto{
					{
						Shape: "round",
						Measures: &pb.MeasuresProto{
							Diameter: func() *int32 { d := int32(28); return &d }(),
						},
						Name: "round 28 cm",
						Area: 615.75,
					},
				},
				TotalArea: 615.75,
			},
		},
		{
			name: "Multiple pans of different shapes",
			input: &pb.PansProto{
				Pans: []*pb.PanProto{
					{
						Shape: "round",
						Measures: &pb.MeasuresProto{
							Diameter: func() *int32 { d := int32(28); return &d }(),
						},
						Name: "round 28 cm",
						Area: 615.75,
					},
					{
						Shape: "rectangular",
						Measures: &pb.MeasuresProto{
							Width:  func() *int32 { w := int32(30); return &w }(),
							Length: func() *int32 { l := int32(40); return &l }(),
						},
						Name: "rectangular 30 x 40 cm",
						Area: 1200.0,
					},
					{
						Shape: "square",
						Measures: &pb.MeasuresProto{
							Edge: func() *int32 { e := int32(25); return &e }(),
						},
						Name: "square 25 cm",
						Area: 625.0,
					},
				},
				TotalArea: 2440.75,
			},
			expected: &pb.PansProto{
				Pans: []*pb.PanProto{
					{
						Shape: "round",
						Measures: &pb.MeasuresProto{
							Diameter: func() *int32 { d := int32(28); return &d }(),
						},
						Name: "round 28 cm",
						Area: 615.75,
					},
					{
						Shape: "rectangular",
						Measures: &pb.MeasuresProto{
							Width:  func() *int32 { w := int32(30); return &w }(),
							Length: func() *int32 { l := int32(40); return &l }(),
						},
						Name: "rectangular 30 x 40 cm",
						Area: 1200.0,
					},
					{
						Shape: "square",
						Measures: &pb.MeasuresProto{
							Edge: func() *int32 { e := int32(25); return &e }(),
						},
						Name: "square 25 cm",
						Area: 625.0,
					},
				},
				TotalArea: 2440.75,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			request := &pb.PansRequest{
				Pans: tc.input,
			}

			response, err := client.TotalDoughWeightByPans(ctx, request)
			require.NoError(t, err)

			assert.NotNil(t, response)
			assert.NotNil(t, response.Pans)
			assert.Equal(t, len(tc.expected.Pans), len(response.Pans.Pans))
			assert.Equal(t, tc.expected.TotalArea, response.Pans.TotalArea)

			for i, expectedPan := range tc.expected.Pans {
				actualPan := response.Pans.Pans[i]
				assert.Equal(t, expectedPan.Shape, actualPan.Shape)
				assert.Equal(t, expectedPan.Name, actualPan.Name)
				assert.Equal(t, expectedPan.Area, actualPan.Area)

				if expectedPan.Measures.Diameter != nil {
					assert.NotNil(t, actualPan.Measures.Diameter)
					assert.Equal(t, *expectedPan.Measures.Diameter, *actualPan.Measures.Diameter)
				}
				if expectedPan.Measures.Edge != nil {
					assert.NotNil(t, actualPan.Measures.Edge)
					assert.Equal(t, *expectedPan.Measures.Edge, *actualPan.Measures.Edge)
				}
				if expectedPan.Measures.Width != nil {
					assert.NotNil(t, actualPan.Measures.Width)
					assert.Equal(t, *expectedPan.Measures.Width, *actualPan.Measures.Width)
				}
				if expectedPan.Measures.Length != nil {
					assert.NotNil(t, actualPan.Measures.Length)
					assert.Equal(t, *expectedPan.Measures.Length, *actualPan.Measures.Length)
				}
			}
		})
	}
}
