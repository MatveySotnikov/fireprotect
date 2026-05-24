package grpcclient

import (
	"context"
	"fmt"
	"log"
	"os"

	pb "github.com/MatveySotnikov/fireprotect/gen/pdfpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var PdfClient pb.PdfServiceClient

// InitPdfClient инициализирует gRPC клиент для PDF-сервиса
func InitPdfClient() {
	addr := os.Getenv("PDF_SERVICE_ADDR")
	if addr == "" {
		addr = "localhost:50052"
	}

	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("Failed to connect to PDF service: %v", err)
	}
	PdfClient = pb.NewPdfServiceClient(conn)
	log.Println("PDF gRPC client initialized")
}

// GenerateAct вызывает PDF-сервис для генерации акта и возвращает PDF-байты
func GenerateAct(ctx context.Context, req *pb.PdfRequest) ([]byte, error) {
	resp, err := PdfClient.GenerateAct(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("PDF generation failed: %w", err)
	}
	return resp.GetPdfData(), nil
}
