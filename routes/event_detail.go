package routes

import (
	"event_proxy/handlers"

	"github.com/go-chi/chi/v5"
)


func RegisterEventDetailRoute(router chi.Router, h *handlers.Handler){
	router.Get("/api/v1/event/{event_id}", h.GetEventDetail)
	router.Get("/api/v1/event_customs/{event_id}", h.GetEventCustomForms)
}