package handlers

import (
	"encoding/json"
	"event_proxy/utils"
	"net/http"
)

func (h *Handler) GetCategories(w http.ResponseWriter, r *http.Request) {

	query := `SELECT COALESCE(json_agg(json_build_object('id', et.id,'name', et.name->>'en_US','icon', icp.value || '/api/v1/image/event.type/' || et.id || '/icon')),'[]'::json) FROM event_type et
CROSS JOIN ir_config_parameter icp
WHERE icp.key = 'image.base.url';
`

	var jsonData []byte

	err := h.DB.QueryRow(query).Scan(&jsonData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := utils.ResponseFormat(200, "success", json.RawMessage(jsonData), nil, nil)
	w.Write(response)
}

