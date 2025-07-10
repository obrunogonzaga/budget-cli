package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	MongoDB MongoDBConfig
}

type MongoDBConfig struct {
	URI      string
	Database string
}

func Load() (*Config, error) {
	godotenv.Load()

	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}

	mongoDatabase := os.Getenv("MONGODB_DATABASE")
	if mongoDatabase == "" {
		mongoDatabase = "financli"
	}

	return &Config{
		MongoDB: MongoDBConfig{
			URI:      mongoURI,
			Database: mongoDatabase,
		},
	}, nil
}
