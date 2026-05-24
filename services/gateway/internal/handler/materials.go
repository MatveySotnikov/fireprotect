package handler

import (
	"encoding/json"
	"net/http"

	"github.com/MatveySotnikov/fireprotect/services/gateway/internal/db"
)

func Materials(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var materials []db.Material
	db.DB.Find(&materials)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(materials)
}
