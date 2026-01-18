package main

import (
	"log"
	"os"

	"github.com/aLittlecrocodile/devops-practice/internal/server"
)

func main() {
	port := getEnv("PORT", "8080")
	srv := server.New(port)
	if err := srv.Start(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
