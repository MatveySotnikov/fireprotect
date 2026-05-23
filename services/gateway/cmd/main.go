package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	pb "github.com/MatveySotnikov/fireprotect/gen/calculatorpb"
	"github.com/MatveySotnikov/fireprotect/services/gateway/internal/auth"
	"github.com/MatveySotnikov/fireprotect/services/gateway/internal/db"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/joho/godotenv"
)

var grpcClient pb.CalcServiceClient

// ---------- Request / Response structs ----------
type calcRequest struct {
	Area          float64 `json:"area"`
	NormativeRate float64 `json:"normative_rate"`
	Layers        int32   `json:"layers"`
	SlopeAngle    float64 `json:"slope_angle"`
	LossFactor    float64 `json:"loss_factor"`
}

type calcResponse struct {
	TotalMass   float64 `json:"total_mass"`
	TotalVolume float64 `json:"total_volume"`
}

type registerRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type tokenResponse struct {
	Token string `json:"token"`
}

// ---------- Инициализация gRPC клиента ----------
func init() {
	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to gRPC calculator: %v", err)
	}
	grpcClient = pb.NewCalcServiceClient(conn)
}

// ---------- Обработчики ----------
func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if req.Name == "" || req.Email == "" || req.Password == "" {
		http.Error(w, "Name, email and password are required", http.StatusBadRequest)
		return
	}

	// Проверка на существующий email
	var existing db.User
	if result := db.DB.Where("email = ?", req.Email).First(&existing); result.Error == nil {
		http.Error(w, "Email already registered", http.StatusConflict)
		return
	}

	hashed, err := auth.HashPassword(req.Password)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	user := db.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: hashed,
	}
	if result := db.DB.Create(&user); result.Error != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":    user.ID,
		"email": user.Email,
		"name":  user.Name,
	})
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var user db.User
	if result := db.DB.Where("email = ?", req.Email).First(&user); result.Error != nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	if !auth.CheckPassword(req.Password, user.Password) {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	token, err := auth.GenerateToken(user.ID, user.Email)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(tokenResponse{Token: token})
}

func calcHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req calcRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	grpcReq := &pb.ComputeRequest{
		Area:          req.Area,
		NormativeRate: req.NormativeRate,
		Layers:        req.Layers,
		SlopeAngle:    req.SlopeAngle,
		LossFactor:    req.LossFactor,
	}

	resp, err := grpcClient.Compute(ctx, grpcReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("Calculation error: %v", err), http.StatusInternalServerError)
		return
	}

	// Пока не сохраняем в БД (будет на шаге 6), просто выводим userID в логи
	log.Printf("User %d (email: %s) requested calculation", claims.UserID, claims.Email)

	jsonResp := calcResponse{
		TotalMass:   resp.GetTotalMass(),
		TotalVolume: resp.GetTotalVolume(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(jsonResp)
}

func main() {
	// Загружаем .env
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	// Подключаем БД
	if err := db.Init(); err != nil {
		log.Fatalf("Database initialization failed: %v", err)
	}
	log.Println("Database connected and migrated")

	http.HandleFunc("/auth/register", registerHandler)
	http.HandleFunc("/auth/login", loginHandler)
	http.HandleFunc("/calc", auth.AuthMiddleware(calcHandler)) // защищён

	log.Println("Gateway HTTP server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
