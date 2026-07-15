package routes

import (
	"event_proxy/handlers"

	"github.com/go-chi/chi/v5"
)


func RegisterUserBookedRoutes(router chi.Router, h *handlers.Handler){
	router.Get("/api/v1/user/{user_id}/events",h.GetUserBookedEvents)
	router.Get("/api/v1/{user_id}/tickets/{ticket_id}",h.GetBookedTicketsByEvent)
}