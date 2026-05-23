package main

import (
	"log"
	"net"

	pb "github.com/MatveySotnikov/fireprotect/gen/calculatorpb"
	"github.com/MatveySotnikov/fireprotect/services/calculator/internal"

	"google.golang.org/grpc"
)

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterCalcServiceServer(s, &internal.CalcServiceServer{})

	log.Println("Calculator gRPC server listening on :50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
