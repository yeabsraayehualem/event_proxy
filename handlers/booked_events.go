package handlers

import (
	"encoding/json"
	"event_proxy/utils"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) GetUserBookedEvents(w http.ResponseWriter, r *http.Request) {
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
	countQuery := `
SELECT COUNT(*)
FROM event_event ev
JOIN res_company c
    ON ev.company_id = c.id
LEFT JOIN (
    SELECT
        event_id,
        partner_id,
        COUNT(*) AS ticket_count
    FROM event_registration
    GROUP BY event_id, partner_id
) user_tickets
ON user_tickets.event_id = ev.id
LEFT JOIN res_partner r
    ON r.id = user_tickets.partner_id
CROSS JOIN ir_config_parameter icp
WHERE icp.key = 'image.base.url'
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
	SELECT
		ev.id AS id,
		ev.name->>'en_US' AS name,

		icp.value || '/api/v1/image/event.event/' || ev.id || '/badge_image' AS pic,

		json_build_object(
			'name', venue.name,
			'street', venue.street,
			'state', cnt.name,
			'phone', venue.phone,
			'email', venue.email
		) AS venue,

		ev.event_location AS event_location,

		(
			SELECT COALESCE(
				json_agg(
					json_build_object(
						'id', tag.id,
						'name', tag.name
					)
					ORDER BY tag.id
				),
				'[]'::json
			)
			FROM event_event_event_tag_rel rel
			JOIN event_tag tag
				ON tag.id = rel.event_tag_id
			WHERE rel.event_event_id = ev.id
		) AS tags,

		ev.date_begin AS date_start,
		ev.date_end,

		CASE
			WHEN NOW() < ev.date_begin THEN 'upcoming'
			WHEN NOW() BETWEEN ev.date_begin AND ev.date_end THEN 'active'
			ELSE 'past'
		END AS status,

		json_build_object(
			'id', et.id,
			'name', et.name->>'en_US'
		) AS category,

		c.merchant AS organizer,

		COALESCE(seats.seats_taken, 0) AS total_attendees,

		COALESCE(user_tickets.ticket_count, 0) AS ticket

	FROM event_event ev

	JOIN res_company c
		ON ev.company_id = c.id

	LEFT JOIN event_type et
		ON ev.event_type_id = et.id

	LEFT JOIN res_partner venue
		ON ev.address_id = venue.id

	LEFT JOIN res_country_state cnt
		ON cnt.id = venue.state_id

	LEFT JOIN (
		SELECT
			event_id,
			COUNT(*) AS seats_taken
		FROM event_registration
		WHERE state IN ('open', 'done')
		GROUP BY event_id
	) seats
	ON seats.event_id = ev.id

	LEFT JOIN (
		SELECT
			event_id,
			partner_id,
			COUNT(*) AS ticket_count
		FROM event_registration
		GROUP BY event_id, partner_id
	) user_tickets
	ON user_tickets.event_id = ev.id

	LEFT JOIN res_partner r
		ON r.id = user_tickets.partner_id

	CROSS JOIN ir_config_parameter icp

	WHERE icp.key = 'image.base.url'
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



func (h *Handler) GetBookedTicketsByEvent(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "user_id")
	eventIDStr := chi.URLParam(r, "ticket_id")

	eventID, err := strconv.Atoi(eventIDStr)
	if err != nil {
		http.Error(w, "Invalid event ID", http.StatusBadRequest)
		return
	}

	query := `
SELECT COALESCE(json_agg(t), '[]'::json) FROM (
	SELECT
		er.id,
		er.name,
		eet.name->>'en_US' AS ticket_type,
		er.barcode
	FROM event_registration er
	JOIN event_event ev ON er.event_id = ev.id
	LEFT JOIN event_event_ticket eet ON eet.id = er.event_ticket_id
	LEFT JOIN res_partner r ON r.id = er.partner_id
	WHERE r.app_user_id = $1
	AND ev.id = $2
) t;`

	var jsonData []byte
	err = h.DB.QueryRow(query, userID, eventID).Scan(&jsonData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := utils.ResponseFormat(200, "success", json.RawMessage(jsonData), nil, nil)
	w.Write(response)
}