package grpc

import (
	"context"

	"github.com/cfioretti/calculator/pkg/application"
	"github.com/cfioretti/calculator/pkg/domain"
	pb "github.com/cfioretti/calculator/pkg/infrastructure/grpc/proto/generated"
)

type Server struct {
	pb.UnimplementedDoughCalculatorServer
	calculatorService *application.DoughCalculatorService
}

func NewServer(calculatorService *application.DoughCalculatorService) *Server {
	return &Server{
		calculatorService: calculatorService,
	}
}

func (s *Server) TotalDoughWeightByPans(ctx context.Context, req *pb.PansRequest) (*pb.PansResponse, error) {
	domainPans := toDomainPans(req.Pans)

	result, err := s.calculatorService.TotalDoughWeightByPans(domainPans)
	if err != nil {
		return nil, err
	}

	responseProto := toProtoMessage(result)

	return &pb.PansResponse{
		Pans: responseProto,
	}, nil
}

func toDomainPans(protoMessage *pb.PansProto) domain.Pans {
	pans := make([]domain.Pan, 0, len(protoMessage.Pans))

	for _, p := range protoMessage.Pans {
		pan := domain.Pan{
			Shape: p.Shape,
			Measures: domain.Measures{
				Diameter: toPointer(p.Measures.Diameter),
				Edge:     toPointer(p.Measures.Edge),
				Width:    toPointer(p.Measures.Width),
				Length:   toPointer(p.Measures.Length),
			},
			Name: p.Name,
			Area: p.Area,
		}
		pans = append(pans, pan)
	}

	return domain.Pans{
		Pans:      pans,
		TotalArea: protoMessage.TotalArea,
	}
}

func toProtoMessage(domainPans *domain.Pans) *pb.PansProto {
	panProtos := make([]*pb.PanProto, 0, len(domainPans.Pans))

	for _, p := range domainPans.Pans {
		panProto := &pb.PanProto{
			Shape: p.Shape,
			Measures: &pb.MeasuresProto{
				Diameter: fromPointer(p.Measures.Diameter),
				Edge:     fromPointer(p.Measures.Edge),
				Width:    fromPointer(p.Measures.Width),
				Length:   fromPointer(p.Measures.Length),
			},
			Name: p.Name,
			Area: p.Area,
		}
		panProtos = append(panProtos, panProto)
	}

	return &pb.PansProto{
		Pans:      panProtos,
		TotalArea: domainPans.TotalArea,
	}
}

func toPointer(value *int32) *int {
	if value == nil {
		return nil
	}
	val := int(*value)
	return &val
}

func fromPointer(value *int) *int32 {
	if value == nil {
		return nil
	}
	val := int32(*value)
	return &val
}
