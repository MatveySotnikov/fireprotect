package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	pb "github.com/MatveySotnikov/fireprotect/gen/pdfpb"
	"github.com/MatveySotnikov/fireprotect/services/gateway/internal/auth"
	"github.com/MatveySotnikov/fireprotect/services/gateway/internal/db"
	"github.com/MatveySotnikov/fireprotect/services/gateway/internal/grpcclient"
	"github.com/MatveySotnikov/fireprotect/services/gateway/internal/model"
)

func Calculations(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/calculations")
	path = strings.TrimSuffix(path, "/")

	// Список
	if path == "" {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var calcs []db.Calculation
		db.DB.Preload("User").Preload("Material").Where("user_id = ?", claims.UserID).Order("created_at desc").Find(&calcs)

		result := make([]model.CalcEnriched, len(calcs))
		for i, c := range calcs {
			mass, vol, err := grpcclient.ComputeMassVolume(c.Area, c.SlopeAngle, c.AreaType, c.UsedNormativeRate, c.UsedDensity)
			if err != nil {
				log.Printf("Error computing mass/volume for calc %d: %v", c.ID, err)
				mass, vol = 0, 0
			}
			result[i] = model.CalcEnriched{
				ID:                c.ID,
				CreatedAt:         c.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
				UpdatedAt:         c.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
				DeletedAt:         nil,
				UserID:            c.UserID,
				MaterialID:        c.MaterialID,
				Area:              c.Area,
				AreaType:          c.AreaType,
				SlopeAngle:        c.SlopeAngle,
				TargetGroup:       c.TargetGroup,
				ApplicationMethod: c.ApplicationMethod,
				LossFactor:        c.LossFactor,
				Layers:            c.Layers,
				UsedNormativeRate: c.UsedNormativeRate,
				UsedDensity:       c.UsedDensity,
				User:              c.User,
				Material:          c.Material,
				TotalMass:         mass,
				TotalVolume:       vol,
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
		return
	}

	// /{id} или /{id}/download
	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
	if len(parts) == 0 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(parts[0])
	if err != nil {
		http.Error(w, "Invalid calculation ID", http.StatusBadRequest)
		return
	}

	// PDF
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

		mass, vol, err := grpcclient.ComputeMassVolume(calc.Area, calc.SlopeAngle, calc.AreaType, calc.UsedNormativeRate, calc.UsedDensity)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to compute result: %v", err), http.StatusInternalServerError)
			return
		}

		pdfReq := &pb.PdfRequest{
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
			CalcDate:      calc.CreatedAt.Format(time.RFC3339),
		}

		pdfBytes, err := grpcclient.GenerateAct(r.Context(), pdfReq)
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

	// Одиночный расчёт
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

		mass, vol, err := grpcclient.ComputeMassVolume(calc.Area, calc.SlopeAngle, calc.AreaType, calc.UsedNormativeRate, calc.UsedDensity)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to compute result: %v", err), http.StatusInternalServerError)
			return
		}

		detail := model.CalcEnriched{
			ID:                calc.ID,
			CreatedAt:         calc.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:         calc.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
			DeletedAt:         nil,
			UserID:            calc.UserID,
			MaterialID:        calc.MaterialID,
			Area:              calc.Area,
			AreaType:          calc.AreaType,
			SlopeAngle:        calc.SlopeAngle,
			TargetGroup:       calc.TargetGroup,
			ApplicationMethod: calc.ApplicationMethod,
			LossFactor:        calc.LossFactor,
			Layers:            calc.Layers,
			UsedNormativeRate: calc.UsedNormativeRate,
			UsedDensity:       calc.UsedDensity,
			User:              calc.User,
			Material:          calc.Material,
			TotalMass:         mass,
			TotalVolume:       vol,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(detail)
		return
	}

	http.Error(w, "Not found", http.StatusNotFound)
}
