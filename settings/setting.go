package settings

import (
	"context"
	"database/sql"
	"event_proxy/database"
	"event_proxy/handlers"
	"fmt"
	"net/http"
)

type Config struct {
	router http.Handler
	DB     *sql.DB
}

func NewConfig() (*Config, error) {
	// TODO: Change this to read from .env first then from Vault  to get the odoo events dbhost, port dbname,db user and password
	dbcon, err := database.Connect(
		"localhost", "5210", "cbe_event", "cbe_event", "you8e",
	)
	if err != nil {

		fmt.Printf("Unable to connect db %v", err)
		return nil, err
	}

	handler := &handlers.Handler{
		DB: dbcon,
	}
	app := &Config{
		router: importURLS(handler),
		DB:     dbcon,
	}

	return app, nil
}

func (c *Config) Start(ctx context.Context) error {

	server := &http.Server{
		Addr:    ":8090",
		Handler: c.router,
	}

	if err := server.ListenAndServe(); err != nil {
		return fmt.Errorf("Failed to strt server : %w", err)
	}
	fmt.Println("server running at http://localhost:8090")

	return nil
}
