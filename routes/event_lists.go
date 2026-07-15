package routes

import (
	"event_proxy/handlers"

	"github.com/go-chi/chi/v5"
)

func RegisterEventListsRoutes(router chi.Router, h *handlers.Handler) {
	router.Get("/api/v1/events", h.GetEvents)
	router.Get("/api/v1/events/popular_events", h.GetPopularEvents)
	router.Get("/api/v1/events/{category_id}", h.GetEventByCategory)

}