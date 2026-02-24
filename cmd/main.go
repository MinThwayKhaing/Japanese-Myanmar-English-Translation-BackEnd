package main

import (
	"USDT_BackEnd/config"
	"USDT_BackEnd/db"
	"USDT_BackEnd/middleware"
	"USDT_BackEnd/routes"
	"fmt"
	"net/http"
	"os"
)

func main() {
	cfg := config.LoadConfig()
	db.ConnectDB(cfg)

	mux := http.NewServeMux()
	routes.RegisterRoutes(mux, cfg)
	handler := middleware.AppVersionMiddleware(cfg)(mux)

	// Optional: Serve static frontend
	fileServer := http.FileServer(http.Dir("./web"))
	mux.Handle("/", fileServer)

	// Get port from environment variable or default to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("🚀 Server running at http://localhost:%s\n", port)
	http.ListenAndServe(":"+port, handler)
}
