package routes

import (
	"event_proxy/handlers"

	"github.com/go-chi/chi/v5"
)


func RegisterBookmarkedRoutes(router chi.Router, h *handlers.Handler){
	router.Get("/api/v1/get_bookmarks/{user_id}", h.GetUserBookmarks)
}