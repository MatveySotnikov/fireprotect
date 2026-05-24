package internal

import (
	"context"
	"fmt"
	"math"

	pb "github.com/MatveySotnikov/fireprotect/gen/calculatorpb"
)

type CalcServiceServer struct {
	pb.UnimplementedCalcServiceServer
}

func (s *CalcServiceServer) Compute(ctx context.Context, req *pb.ComputeRequest) (*pb.ComputeResponse, error) {
	area := req.GetArea()
	normRate := req.GetNormativeRate()
	layers := req.GetLayers()
	slopeAngle := req.GetSlopeAngle()
	lossFactor := req.GetLossFactor()

	// Проверка входных данных
	if area <= 0 || normRate <= 0 || layers <= 0 || slopeAngle < 0 || slopeAngle >= 90 || lossFactor < 0 {
		return nil, fmt.Errorf("invalid input parameters")
	}

	density := req.GetDensity()
	if density <= 0 {
		density = 1.2
	}

	// Коэффициент наклона: k1 = 1 / cos(угол в радианах)
	angleRad := slopeAngle * math.Pi / 180.0
	cosA := math.Cos(angleRad)
	if cosA == 0 {
		return nil, fmt.Errorf("cos(slope_angle) is zero")
	}
	slopeCoeff := 1.0 / cosA

	// Расход на 1 м²
	Q := normRate * float64(layers) * (1 + lossFactor) * slopeCoeff
	totalMass := area * Q // кг

	// Объём в литрах
	totalVolume := totalMass / density

	return &pb.ComputeResponse{
		TotalMass:   totalMass,
		TotalVolume: totalVolume,
	}, nil
}
