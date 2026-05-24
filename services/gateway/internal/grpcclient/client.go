package grpcclient

import (
	"fmt"
	"log"
	"math"
	"os"

	pb "github.com/MatveySotnikov/fireprotect/gen/calculatorpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var Client pb.CalcServiceClient

// Init инициализирует gRPC клиент
func Init() {
	serviceConfig := `{
		"methodConfig": [{
			"name": [{"service": "calculatorpb.CalcService"}],
			"retryPolicy": {
				"maxAttempts": 3,
				"initialBackoff": "0.1s",
				"maxBackoff": "1s",
				"backoffMultiplier": 2.0,
				"retryableStatusCodes": ["UNAVAILABLE"]
			}
		}]
	}`

	addr := os.Getenv("CALCULATOR_ADDR")
	if addr == "" {
		addr = "localhost:50051"
	}

	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(serviceConfig),
	)

	if err != nil {
		log.Fatalf("Failed to connect to gRPC calculator: %v", err)
	}
	Client = pb.NewCalcServiceClient(conn)
	log.Println("gRPC client initialized")
}

// ComputeMassVolume вычисляет массу и объём по исходным данным
func ComputeMassVolume(area, slopeAngle float64, areaType string, usedNormativeRate, usedDensity float64) (float64, float64, error) {
	if area <= 0 {
		return 0, 0, fmt.Errorf("площадь должна быть положительной")
	}
	var effectiveArea float64
	switch areaType {
	case "projection":
		if slopeAngle < 0 || slopeAngle >= 90 {
			return 0, 0, fmt.Errorf("некорректный угол уклона")
		}
		angleRad := slopeAngle * math.Pi / 180.0
		cosA := math.Cos(angleRad)
		if cosA == 0 {
			return 0, 0, fmt.Errorf("cos(slope_angle) равен нулю")
		}
		effectiveArea = area / cosA
	case "slope":
		effectiveArea = area
	default:
		return 0, 0, fmt.Errorf("неизвестный тип площади: %s", areaType)
	}

	totalMass := effectiveArea * usedNormativeRate
	totalVolume := totalMass / usedDensity
	return totalMass, totalVolume, nil
}
