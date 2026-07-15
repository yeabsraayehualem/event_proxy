package handlers

import (
	"encoding/json"
	"event_proxy/utils"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) GetEventDetail(w http.ResponseWriter, r *http.Request) {
	eventIDStr := chi.URLParam(r, "event_id")
	eventID, err := strconv.Atoi(eventIDStr)
	if err != nil {
		http.Error(w, "Invalid event ID", http.StatusBadRequest)
		return
	}

	query := `
SELECT COALESCE(row_to_json(t), '{}'::json) FROM (
	SELECT
		e.id,
		e.name->>'en_US' AS name,
		icp.value || '/api/v1/image/event.event/' || e.id || '/badge_image' AS pic,
		e.event_location,
		e.date_begin,
		e.date_end,
		regexp_replace(
			e.description->>'en_US',
			'<[^>]*>',
			'',
			'g'
		) AS description,

		json_build_object(
			'name', ven.name,
			'lat', e.lat_location,
			'lng', e.lng_location
		) AS venue,

		json_build_object(
			'id', et.id,
			'name', et.name->>'en_US'
		) AS category,

		json_build_object(
			'id', c.id,
			'name', c.name,
			'merchant_id', c.merchant,
			'logo', icp.value || '/api/v1/image/res.company/' || c.id || '/logo'
		) AS organizer,

		(
			SELECT COALESCE(
				json_agg(
					json_build_object(
						'id', eg.id,
						'image', eg.image_url
					)
					ORDER BY eg.id
				),
				'[]'::json
			)
			FROM event_galary eg
			WHERE eg.event_id = e.id
		) AS gallery,

		(
			SELECT COALESCE(
				json_agg(
					json_build_object(
						'id', ev.id,
						'name', ev.name
					)
					ORDER BY ev.id
				),
				'[]'::json
			)
			FROM events_variant ev
			WHERE ev.event_id = e.id
		) AS custom_form,
		json_build_object(
			'id', tic.id, 'name', tic.name->>'en_US', 'price', tic.price, 'description', tic.description, 'icon', tic.icon_url
		) AS tickets
	FROM event_event e
	JOIN res_company c
		ON e.company_id = c.id
	LEFT JOIN res_partner ven
		ON e.address_id = ven.id
	LEFT JOIN event_type et
		ON e.event_type_id = et.id
	LEFT JOIN event_event_ticket as tic ON tic.event_id = e.id
	CROSS JOIN ir_config_parameter icp
	WHERE icp.key = 'image.base.url'
	  AND e.id = $1
) t;`

	var jsonData []byte
	err = h.DB.QueryRow(query, eventID).Scan(&jsonData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := utils.ResponseFormat(200, "success", json.RawMessage(jsonData), nil, nil)
	w.Write(response)
}



func (h *Handler) GetEventCustomForms(w http.ResponseWriter, r *http.Request) {
	eventIDStr := chi.URLParam(r, "event_id")
	eventID, err := strconv.Atoi(eventIDStr)
	if err != nil {
		http.Error(w, "Invalid event ID", http.StatusBadRequest)
		return
	}

	query := `
SELECT COALESCE(json_agg(t), '[]'::json) FROM (
	SELECT
		ev.id,
		ev.name,
		COALESCE(
			json_agg(
				json_build_object(
					'id', ea.id,
					'name', ea.name
				)
			) FILTER (WHERE ea.id IS NOT NULL),
			'[]'::json
		) AS attributes
	FROM events_variant ev
	LEFT JOIN event_variant_attributes_events_variant_rel rel
		ON rel.events_variant_id = ev.id
	LEFT JOIN event_variant_attributes ea
		ON ea.id = rel.event_variant_attributes_id
	WHERE ev.event_id = $1
	GROUP BY ev.id, ev.name
) t;`

	var jsonData []byte
	err = h.DB.QueryRow(query, eventID).Scan(&jsonData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := utils.ResponseFormat(200, "success", json.RawMessage(jsonData), nil, nil)
	w.Write(response)
}
