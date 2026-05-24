package handler

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
	"github.com/MatveySotnikov/fireprotect/services/gateway/internal/grpcclient"
	"github.com/MatveySotnikov/fireprotect/services/gateway/internal/model"
)

func Calc(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req model.CalcRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var usedNormativeRate, usedDensity float64
	var materialID *uint
	var lossFactor float64

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
			http.Error(w, "Invalid target_group", http.StatusBadRequest)
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

	resp, err := grpcclient.Client.Compute(ctx, grpcReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("Calculation error: %v", err), http.StatusInternalServerError)
		return
	}

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

	jsonResp := model.CalcResponse{
		TotalMass:   resp.GetTotalMass(),
		TotalVolume: resp.GetTotalVolume(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(jsonResp)
}
