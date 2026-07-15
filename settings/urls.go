package settings

import (
	"event_proxy/handlers"
	"event_proxy/routes"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func importURLS(dbHandler *handlers.Handler) *chi.Mux {
	router := chi.NewRouter()

	router.Use(middleware.Logger)

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Register all route groups
	routes.RegisterCategoryRoutes(router, dbHandler)
	routes.RegisterEventListsRoutes(router, dbHandler)
	routes.RegisterEventDetailRoute(router, dbHandler)
	routes.RegisterBookmarkedRoutes(router, dbHandler)
	routes.RegisterUserBookedRoutes(router, dbHandler)
	return router
}
