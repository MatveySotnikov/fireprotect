package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	pb "github.com/MatveySotnikov/fireprotect/gen/calculatorpb"
	"github.com/MatveySotnikov/fireprotect/services/gateway/internal/auth"
	"github.com/MatveySotnikov/fireprotect/services/gateway/internal/db"
	"github.com/MatveySotnikov/fireprotect/services/gateway/internal/pdf"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/joho/godotenv"
)

var grpcClient pb.CalcServiceClient

// ---------- Request / Response structs ----------

// Новый формат запроса на расчёт
type calcRequest struct {
	Area              float64 `json:"area"`
	AreaType          string  `json:"area_type"` // "projection" или "slope"
	SlopeAngle        float64 `json:"slope_angle"`
	TargetGroup       string  `json:"target_group"`          // "1_group" или "2_group"
	ApplicationMethod string  `json:"application_method"`    // "brush", "spray_indoor", "spray_outdoor"
	MaterialID        *uint   `json:"material_id,omitempty"` // указатель, может быть null
	// Поля для ручного ввода, если material_id не указан
	NormativeRate *float64 `json:"normative_rate,omitempty"`
	Density       *float64 `json:"density,omitempty"`
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

// ---------- Вспомогательные функции ----------

// computeMassVolume вычисляет массу и объём по исходным данным (без учёта площади)
func computeMassVolume(area, slopeAngle float64, areaType string, usedNormativeRate, usedDensity float64) (float64, float64, error) {
	if area <= 0 {
		return 0, 0, fmt.Errorf("площадь должна быть положительной")
	}
	// Определяем эффективную площадь с учётом типа
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

	// Определяем расход и плотность
	var usedNormativeRate, usedDensity float64
	var materialID *uint
	var lossFactor float64 // разница (например, 0.2 для spray_indoor)

	if req.MaterialID != nil {
		var mat db.Material
		if err := db.DB.First(&mat, *req.MaterialID).Error; err != nil {
			http.Error(w, "Material not found", http.StatusBadRequest)
			return
		}

		var baseConsumption float64
		switch req.TargetGroup {
		case "1_group":
			baseConsumption = mat.Group1Consumption
		case "2_group":
			baseConsumption = mat.Group2Consumption
		default:
			http.Error(w, "Invalid target_group, must be '1_group' or '2_group'", http.StatusBadRequest)
			return
		}

		switch req.ApplicationMethod {
		case "brush":
			lossFactor = mat.BrushLoss - 1.0
		case "spray_indoor":
			lossFactor = mat.SprayIndoorLoss - 1.0
		case "spray_outdoor":
			lossFactor = mat.SprayOutdoorLoss - 1.0
		default:
			http.Error(w, "Invalid application_method", http.StatusBadRequest)
			return
		}

		usedNormativeRate = baseConsumption * (1 + lossFactor)
		usedDensity = mat.DefaultDensity
		materialID = req.MaterialID
	} else {
		if req.NormativeRate == nil || req.Density == nil {
			http.Error(w, "Either material_id or both normative_rate and density must be provided", http.StatusBadRequest)
			return
		}
		usedNormativeRate = *req.NormativeRate
		usedDensity = *req.Density
		materialID = nil
		lossFactor = 0
	}

	if req.Area <= 0 || usedNormativeRate <= 0 || usedDensity <= 0 {
		http.Error(w, "Invalid input values", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	grpcReq := &pb.ComputeRequest{
		Area:          req.Area,
		NormativeRate: usedNormativeRate,
		Layers:        1,
		SlopeAngle:    req.SlopeAngle,
		LossFactor:    0,
		Density:       usedDensity,
	}

	resp, err := grpcClient.Compute(ctx, grpcReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("Calculation error: %v", err), http.StatusInternalServerError)
		return
	}

	// Сохранение в БД
	lossMultiplier := 1.0 + lossFactor

	calc := db.Calculation{
		UserID:            claims.UserID,
		MaterialID:        materialID,
		Area:              req.Area,
		AreaType:          req.AreaType,
		SlopeAngle:        req.SlopeAngle,
		TargetGroup:       req.TargetGroup,
		ApplicationMethod: req.ApplicationMethod,
		LossFactor:        lossMultiplier,
		Layers:            1,
		UsedNormativeRate: usedNormativeRate,
		UsedDensity:       usedDensity,
	}

	if err := db.DB.Create(&calc).Error; err != nil {
		log.Printf("Failed to save calculation: %v", err)
	} else {
		log.Printf("Calculation saved with ID %d for user %d", calc.ID, claims.UserID)
	}

	jsonResp := calcResponse{
		TotalMass:   resp.GetTotalMass(),
		TotalVolume: resp.GetTotalVolume(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(jsonResp)
}

// Обработчик истории расчётов
func calculationsHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/calculations")
	path = strings.TrimSuffix(path, "/")

	// GET /calculations — список
	if path == "" {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var calcs []db.Calculation
		db.DB.Preload("User").Preload("Material").Where("user_id = ?", claims.UserID).Order("created_at desc").Find(&calcs)

		// Обогащаем каждую запись вычисленными массой и объёмом
		type calcEnriched struct {
			db.Calculation
			TotalMass   float64 `json:"total_mass"`
			TotalVolume float64 `json:"total_volume"`
		}
		result := make([]calcEnriched, len(calcs))
		for i, c := range calcs {
			mass, vol, err := computeMassVolume(c.Area, c.SlopeAngle, c.AreaType, c.UsedNormativeRate, c.UsedDensity)
			if err != nil {
				log.Printf("Error computing mass/volume for calc %d: %v", c.ID, err)
				mass, vol = 0, 0
			}
			result[i] = calcEnriched{c, mass, vol}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
		return
	}

	// Дальше путь вида /{id} или /{id}/download
	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
	if len(parts) == 0 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	idStr := parts[0]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid calculation ID", http.StatusBadRequest)
		return
	}

	// GET /calculations/{id}/download
	if len(parts) == 2 && parts[1] == "download" {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var calc db.Calculation
		if result := db.DB.Preload("User").Where("id = ? AND user_id = ?", id, claims.UserID).First(&calc); result.Error != nil {
			http.Error(w, "Calculation not found", http.StatusNotFound)
			return
		}

		mass, vol, err := computeMassVolume(calc.Area, calc.SlopeAngle, calc.AreaType, calc.UsedNormativeRate, calc.UsedDensity)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to compute result: %v", err), http.StatusInternalServerError)
			return
		}

		data := pdf.ActData{
			UserName:      calc.User.Name,
			UserEmail:     calc.User.Email,
			Area:          calc.Area,
			NormativeRate: calc.UsedNormativeRate,
			Layers:        calc.Layers,
			SlopeAngle:    calc.SlopeAngle,
			LossFactor:    calc.LossFactor,
			Density:       calc.UsedDensity,
			TotalMass:     mass,
			TotalVolume:   vol,
			CalcDate:      calc.CreatedAt,
		}
		pdfBytes, err := pdf.GenerateAct(data)
		if err != nil {
			http.Error(w, fmt.Sprintf("PDF generation error: %v", err), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=act_%d.pdf", calc.ID))
		w.WriteHeader(http.StatusOK)
		w.Write(pdfBytes)
		return
	}

	// GET /calculations/{id}
	if len(parts) == 1 {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var calc db.Calculation
		if result := db.DB.Preload("User").Preload("Material").Where("id = ? AND user_id = ?", id, claims.UserID).First(&calc); result.Error != nil {
			http.Error(w, "Calculation not found", http.StatusNotFound)
			return
		}

		mass, vol, err := computeMassVolume(calc.Area, calc.SlopeAngle, calc.AreaType, calc.UsedNormativeRate, calc.UsedDensity)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to compute result: %v", err), http.StatusInternalServerError)
			return
		}

		type calcDetail struct {
			db.Calculation
			TotalMass   float64 `json:"total_mass"`
			TotalVolume float64 `json:"total_volume"`
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(calcDetail{calc, mass, vol})
		return
	}

	http.Error(w, "Not found", http.StatusNotFound)
}

// Обработчик справочника материалов
func materialsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var materials []db.Material
	db.DB.Find(&materials)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(materials)
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	if err := db.Init(); err != nil {
		log.Fatalf("Database initialization failed: %v", err)
	}
	log.Println("Database connected and migrated")

	// Маршруты
	http.HandleFunc("/auth/register", registerHandler)
	http.HandleFunc("/auth/login", loginHandler)
	http.HandleFunc("/calc", auth.AuthMiddleware(calcHandler))
	http.HandleFunc("/calculations/", auth.AuthMiddleware(calculationsHandler))
	http.HandleFunc("/materials", materialsHandler) // общедоступный
	log.Println("Gateway HTTP server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
