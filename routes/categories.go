package routes

import (
	"event_proxy/handlers"

	"github.com/go-chi/chi/v5"
)

func RegisterCategoryRoutes(router chi.Router, h *handlers.Handler) {
	router.Get("/api/v1/event/categories", h.GetCategories)
}
