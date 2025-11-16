package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	MongoURI    string
	Database    string
	JWTSecret   string
	TokenExpiry int
}

func LoadConfig() *Config {
	// Load .env file automatically
	_ = godotenv.Load()

	uri := os.Getenv("MONGODB_URI")
	db := os.Getenv("MONGODB_DBNAME")

	if uri == "" || db == "" {
		log.Fatal("Missing MONGODB_URI or MONGODB_DBNAME environment variables")
	}

	return &Config{
		MongoURI:    uri,
		Database:    db,
		JWTSecret:   os.Getenv("JWT_SECRET"),
		TokenExpiry: 72,
	}
}
