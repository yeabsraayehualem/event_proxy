package utils

import (
	"encoding/json"
	"net/http"
)

var statusMessages = map[int]string{
	http.StatusOK:                  "Success",
	http.StatusCreated:             "Created successfully",
	http.StatusBadRequest:          "Bad request",
	http.StatusUnauthorized:        "Unauthorized",
	http.StatusForbidden:           "Forbidden",
	http.StatusNotFound:            "Resource not found",
	http.StatusConflict:            "Conflict",
	http.StatusInternalServerError: "Internal server error",
}

func statusMessage(code int) string {
	if msg, ok := statusMessages[code]; ok {
		return msg
	}
	return "Unknown status"
}

func ResponseFormat(
	statusCode int,
	message string,
	data interface{},
	pagination interface{},
	extra map[string]interface{},
) []byte {

	if message == "" {
		message = statusMessage(statusCode)
	}

	payload := map[string]interface{}{
		"status":  statusCode,
		"message": message,
		"data":    data,
	}

	if payload["data"] == nil {
		payload["data"] = map[string]interface{}{}
	}

	if pagination != nil {
		payload["pagination"] = pagination
	}

	for k, v := range extra {
		payload[k] = v
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return []byte(`{"status":500,"message":"Internal server error","data":{}}`)
	}

	return jsonData
}

