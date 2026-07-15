package handlers

import (
	"encoding/json"
	"event_proxy/utils"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) GetEvents(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	pageStr := r.URL.Query().Get("page")
	userID := r.URL.Query().Get("user_id")

	limit := 10
	page := 0

	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil {
			limit = parsedLimit
		}
	}
	if pageStr != "" {
		if parsedPage, err := strconv.Atoi(pageStr); err == nil {
			page = parsedPage -1
		}
	}

	offset := page * limit

	var total int
	countQuery := `SELECT COUNT(*) FROM event_event e JOIN res_company c ON e.company_id = c.id CROSS JOIN ir_config_parameter icp
WHERE icp.key = 'image.base.url' AND e.x_event_approval_status = 'published' AND e.is_published = TRUE AND e.date_end >= NOW() AND c.cps_enabled = TRUE;`

	err := h.DB.QueryRow(countQuery).Scan(&total)
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
	SELECT
		e.id,
		e.name->>'en_US' AS name,
		json_build_object(
			'id', c.id,
			'name', c.name,
			'merchant_id', c.merchant
		) AS organizer,
		e.date_begin,
		e.date_end,
		e.event_location,
		icp.value || '/api/v1/image/event.event/' || e.id || '/badge_image' AS pic,
		CASE
			WHEN eb.id IS NOT NULL THEN TRUE
			ELSE FALSE
		END AS is_bookmarked
	FROM event_event e
	JOIN res_company c
		ON e.company_id = c.id
	LEFT JOIN event_bookmark eb 
		ON eb.event_id = e.id
		AND eb.user_id = (
			SELECT id 
			FROM res_partner 
			WHERE app_user_id = $1 
		)
	CROSS JOIN ir_config_parameter icp
	WHERE icp.key = 'image.base.url'
	  AND e.x_event_approval_status = 'published'
	  AND e.is_published = TRUE
	  AND e.date_end >= NOW()
	  AND c.cps_enabled = TRUE
	ORDER BY e.date_begin ASC, e.id ASC
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
		"page":        page,
		"limit":       limit,
		"total":       total,
		"total_pages": totalPages,
		"has_next":    hasNext,
		"has_prev":    hasPrev,
	}
	response := utils.ResponseFormat(200, "success", json.RawMessage(jsonData), pagination, nil)
	w.Write(response)
}

func (h *Handler) GetPopularEvents(w http.ResponseWriter, r *http.Request) {
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

	offset := page * limit

	var total int
	countQuery := `
SELECT COUNT(*)
FROM event_event e
JOIN res_company c
    ON e.company_id = c.id
CROSS JOIN ir_config_parameter icp
LEFT JOIN (
    SELECT
        event_id,
        COUNT(*) AS seats_taken
    FROM event_registration
    WHERE state IN ('open', 'done')
    GROUP BY event_id
) r
    ON r.event_id = e.id
WHERE icp.key = 'image.base.url'
  AND e.x_event_approval_status = 'published'
  AND e.is_published = TRUE
  AND COALESCE(r.seats_taken, 0) > 0
  AND e.date_end >= NOW()
  AND c.cps_enabled = TRUE;`

	err := h.DB.QueryRow(countQuery).Scan(&total)
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
	SELECT
	  e.id,
    e.name->>'en_US' AS name,
  
    e.date_begin AS date_start,
    e.event_location AS location,
	e.lat_location AS lat, 
	e.lng_location AS long,
    icp.value || '/api/v1/image/event.event/' || e.id || '/badge_image' AS event_picture
	FROM event_event e
	JOIN res_company c
		ON e.company_id = c.id
	CROSS JOIN ir_config_parameter icp
	LEFT JOIN (
		SELECT
			event_id,
			COUNT(*) AS seats_taken
		FROM event_registration
		WHERE state IN ('open', 'done')
		GROUP BY event_id
	) r
		ON r.event_id = e.id
	WHERE icp.key = 'image.base.url'
	  AND e.x_event_approval_status = 'published'
	  AND e.is_published = TRUE
	  AND COALESCE(r.seats_taken, 0) > 0
	  AND e.date_end >= NOW()
	  AND c.cps_enabled = TRUE
	ORDER BY
		COALESCE(r.seats_taken, 0) DESC,
		e.date_begin ASC
	LIMIT $1
	OFFSET $2
) t;`

	var jsonData []byte
	err = h.DB.QueryRow(query, limit, offset).Scan(&jsonData)
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

func (h *Handler) GetEventByCategory(w http.ResponseWriter, r *http.Request) {
	categoryIDStr := chi.URLParam(r, "category_id")
	categoryID, err := strconv.Atoi(categoryIDStr)
	if err != nil {
		http.Error(w, "Invalid category ID", http.StatusBadRequest)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	pageStr := r.URL.Query().Get("page")
	userID := r.URL.Query().Get("user_id")

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

	offset := page * limit

	var total int
	countQuery := `SELECT COUNT(*) FROM event_event e JOIN res_company c ON e.company_id = c.id CROSS JOIN ir_config_parameter icp
WHERE icp.key = 'image.base.url' AND e.x_event_approval_status = 'published' AND e.is_published = TRUE AND e.date_end >= NOW() AND c.cps_enabled = TRUE AND e.event_type_id = $1;`

	err = h.DB.QueryRow(countQuery, categoryID).Scan(&total)
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
	SELECT
		e.id,
		e.name->>'en_US' AS name,
		json_build_object(
			'id', c.id,
			'name', c.name,
			'merchant_id', c.merchant
		) AS organizer,
		e.date_begin,
		e.date_end,
		e.event_location,
		icp.value || '/api/v1/image/event.event/' || e.id || '/badge_image' AS pic,
		CASE
			WHEN eb.id IS NOT NULL THEN TRUE
			ELSE FALSE
		END AS is_bookmarked
	FROM event_event e
	JOIN res_company c
		ON e.company_id = c.id
	LEFT JOIN event_bookmark eb 
		ON eb.event_id = e.id
		AND eb.user_id = (
			SELECT id 
			FROM res_partner 
			WHERE app_user_id = $1 
		)
	CROSS JOIN ir_config_parameter icp
	WHERE icp.key = 'image.base.url'
	  AND e.x_event_approval_status = 'published'
	  AND e.is_published = TRUE
	  AND e.date_end >= NOW()
	  AND c.cps_enabled = TRUE
	  AND e.event_type_id = $2
	ORDER BY e.date_begin ASC, e.id ASC
	LIMIT $3
	OFFSET $4
) t;`

	var jsonData []byte
	err = h.DB.QueryRow(query, userID, categoryID, limit, offset).Scan(&jsonData)
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