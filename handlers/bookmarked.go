package handlers

import (
	"encoding/json"
	"event_proxy/utils"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) GetUserBookmarks(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "user_id")

	limitStr := r.URL.Query().Get("limit")
	pageStr := r.URL.Query().Get("page")

	limit := 10
	page := 0

	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil {
			limit = parsedLimit
		}
	}
	if pageStr != "" {
		if parsedPage, err := strconv.Atoi(pageStr); err == nil {
			page = parsedPage - 1
		}
	}
	if page < 0 {
		page = 0
	}

	offset := page * limit

	var total int
	countQuery := `SELECT COUNT(*) FROM event_bookmark b 
JOIN res_partner r ON r.id = b.user_id 
LEFT JOIN event_event e ON e.id = b.event_id 
CROSS JOIN ir_config_parameter icp WHERE icp.key = 'image.base.url'
AND r.app_user_id = $1;`

	err := h.DB.QueryRow(countQuery, userID).Scan(&total)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	totalPages := 0
	if limit > 0 {
		totalPages = (total + limit - 1) / limit
	}

	hasPrev := page > 0
	hasNext := false
	if limit > 0 {
		hasNext = (page+1)*limit < total
	}

	query := `
SELECT COALESCE(json_agg(t), '[]'::json) FROM (
	SELECT b.id,
		e.id AS event_id,
		e.name->>'en_US' AS event_name, 
		e.event_location AS location,    
		icp.value || '/api/v1/image/event.event/' || e.id || '/badge_image' AS pic,
		e.date_begin AS event_date_start,
		e.date_end AS event_date_end
	FROM event_bookmark b 
	JOIN res_partner r ON r.id = b.user_id 
	LEFT JOIN event_event e ON e.id = b.event_id 
	CROSS JOIN ir_config_parameter icp WHERE icp.key = 'image.base.url'
	AND r.app_user_id = $1
	LIMIT $2
	OFFSET $3
) t;`

	var jsonData []byte
	err = h.DB.QueryRow(query, userID, limit, offset).Scan(&jsonData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	pagination := map[string]interface{}{
		"page":        page + 1,
		"limit":       limit,
		"total":       total,
		"total_pages": totalPages,
		"has_next":    hasNext,
		"has_prev":    hasPrev,
	}
	response := utils.ResponseFormat(200, "success", json.RawMessage(jsonData), pagination, nil)
	w.Write(response)
}