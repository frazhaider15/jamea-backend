package main

import (
	"os"

	"go.uber.org/zap"

	"github.com/jamea/conf"
	"github.com/jamea/logger"
	"github.com/jamea/routes"
	"github.com/jamea/store"
)

func main() {
	logger.InitLogger()
	conf.LoadEnvVariables()

	// Initialize the in-memory user store
	store.Init()

	// Starting REST server
	addr := os.Getenv("REST_SERVER_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	r := routes.SetupRouter()
	if err := r.Run(addr); err != nil {
		logger.Logger.Error("Failed to run server", zap.Error(err))
	}
}
