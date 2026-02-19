package main

import (
	"os"

	"go.uber.org/zap"

	"github.com/jamea/conf"
	"github.com/jamea/db"
	"github.com/jamea/logger"
	"github.com/jamea/routes"
)

func main() {
	logger.InitLogger()
	conf.LoadEnvVariables()

	// Initialize Database
	db.Connect()
	db.Migrate()
	db.Seed()

	// Starting REST server
	addr := os.Getenv("REST_SERVER_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	r := routes.SetupRouter()
	logger.Logger.Info("Server listening", zap.String("addr", addr))
	if err := r.Run(addr); err != nil {
		logger.Logger.Error("Failed to run server", zap.Error(err))
	}
}
