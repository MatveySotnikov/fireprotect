package main

import (
	"log"
	"net/http"

	"github.com/MatveySotnikov/fireprotect/services/gateway/internal/auth"
	"github.com/MatveySotnikov/fireprotect/services/gateway/internal/db"
	"github.com/MatveySotnikov/fireprotect/services/gateway/internal/grpcclient"
	"github.com/MatveySotnikov/fireprotect/services/gateway/internal/handler"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	if err := db.Init(); err != nil {
		log.Fatalf("Database initialization failed: %v", err)
	}
	log.Println("Database connected and migrated")

	grpcclient.Init()

	http.HandleFunc("/auth/register", handler.Register)
	http.HandleFunc("/auth/login", handler.Login)
	http.HandleFunc("/calc", auth.AuthMiddleware(handler.Calc))
	http.HandleFunc("/calculations/", auth.AuthMiddleware(handler.Calculations))
	http.HandleFunc("/materials", handler.Materials)

	log.Println("Gateway HTTP server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
