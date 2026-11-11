package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/r4stypasty/gobuild-api/internal/models"
)

// GetMaterials godoc
// @Summary      Get all materials
// @Description  Returns a list of all materials in stock
// @Tags         materials
// @Produce      json
// @Success      200  {array}   models.Material
// @Failure      500  {object}  map[string]string
// @Router       /materials [get]
// @Security ApiKeyAuth
// @Param        limit   query     int  false  "Number of items to return (default 10, max 100)"
// @Param        offset  query     int  false  "Number of items to skip (default 0)"
// GET /materials
func (h *MaterialHandler) GetMaterials(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 10 // Default limit
	if limitStr != "" {
		l, err := strconv.Atoi(limitStr)
		if err != nil {
			h.respondWithError(w, http.StatusBadRequest, "Invalid limit format")
			return
		}
		if l <= 0 {
			h.respondWithError(w, http.StatusBadRequest, "limit must be a positive integer")
			return
		}
		if l > 100 {
			h.respondWithError(w, http.StatusBadRequest, "limit cannot exceed 100")
			return
		}
		limit = l
	}

	offset := 0 // Default offset
	if offsetStr != "" {
		o, err := strconv.Atoi(offsetStr)
		if err != nil {
			h.respondWithError(w, http.StatusBadRequest, "Invalid offset format")
			return
		}
		if o < 0 {
			h.respondWithError(w, http.StatusBadRequest, "offset must be a non-negative integer")
			return
		}
		offset = o
	}

	rows, err := h.DB.QueryContext(r.Context(), "SELECT id, name, unit, current_stock, min_stock FROM materials ORDER BY id LIMIT ? OFFSET ?", limit, offset)
	if err != nil {
		log.Printf("[ERROR] Database query failed: %v", err)
		h.respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	defer rows.Close()

	materials := make([]models.Material, 0)

	for rows.Next() {
		var m models.Material
		err := rows.Scan(&m.ID, &m.Name, &m.Unit, &m.CurrentStock, &m.MinStock)
		if err != nil {
			log.Printf("[ERROR] Scan row failed: %v", err)
			h.respondWithError(w, http.StatusInternalServerError, "Internal server error")
			return
		}
		materials = append(materials, m)
	}

	if err = rows.Err(); err != nil {
		log.Printf("[ERROR] Rows iteration failed: %v", err)
		h.respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(materials); err != nil {
		log.Printf("[ERROR] JSON encoding failed: %v", err)
	}
}

// GetAlerts godoc
// @Summary      Get low stock alerts
// @Description  Returns materials whose current stock is below or equal to the minimum stock
// @Tags         alerts
// @Produce      json
// @Success      200  {array}   map[string]interface{}
// @Failure      500  {object}  map[string]string
// @Router       /alerts [get]
// Alerting for low stock levels
// @Security ApiKeyAuth
func (h *MaterialHandler) GetAlerts(w http.ResponseWriter, r *http.Request) {
	rows, err := h.DB.QueryContext(r.Context(), "SELECT id, name, unit, current_stock, min_stock FROM materials WHERE current_stock <= min_stock")
	if err != nil {
		log.Printf("[ERROR] Database query failed: %v", err)
		h.respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	defer rows.Close()

	alerts := []map[string]interface{}{}
	for rows.Next() {
		var id int
		var current_stock, min_stock float64
		var name, unit string
		if err := rows.Scan(&id, &name, &unit, &current_stock, &min_stock); err != nil {
			log.Printf("[ERROR] Failed to scan row: %v", err)
			h.respondWithError(w, http.StatusInternalServerError, "Internal server error")
			return
		}
		alerts = append(alerts, map[string]interface{}{
			"id":            id,
			"name":          name,
			"unit":          unit,
			"current_stock": current_stock,
			"min_stock":     min_stock,
		})
	}
	if err = rows.Err(); err != nil {
		log.Printf("[ERROR] Rows iteration failed: %v", err)
		h.respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(alerts)
}
