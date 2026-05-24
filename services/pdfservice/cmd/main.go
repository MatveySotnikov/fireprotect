package main

import (
	"log"
	"net"

	pb "github.com/MatveySotnikov/fireprotect/gen/pdfpb"
	"github.com/MatveySotnikov/fireprotect/services/pdfservice/internal"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func main() {
	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterPdfServiceServer(s, &internal.PdfServiceServer{})

	// Health check
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(s, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	log.Println("PDF service gRPC server listening on :50052")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
