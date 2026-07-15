package settings

import (
	"context"
	"database/sql"
	"event_proxy/database"
	"event_proxy/handlers"
	"fmt"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	router http.Handler
	DB     *sql.DB
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func NewConfig() (*Config, error) {
	// Load from .env if present; ignore error if file doesn't exist (env vars may be set directly)
	_ = godotenv.Load()

	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5210")
	user := getEnv("DB_USER", "cbe_event")
	password := getEnv("DB_PASSWORD", "cbe_event")
	dbname := getEnv("DB_NAME", "you8e")

	dbcon, err := database.Connect(host, port, user, password, dbname)
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

	fmt.Println("server running at http://localhost:8090")
	if err := server.ListenAndServe(); err != nil {
		return fmt.Errorf("Failed to start server: %w", err)
	}

	return nil
}
