package main

import (
	"github.com/OddOneOutApp/backend/internal/config"
	"github.com/OddOneOutApp/backend/internal/database"
	"github.com/OddOneOutApp/backend/internal/http"
	"github.com/OddOneOutApp/backend/internal/utils"
)

func main() {
	utils.InitializeLogger()
	utils.Logger.Infoln("Starting OddOneOut Backend...")

	cfg := config.Load()
	db := database.New()

	http.Initialize(db, cfg)
}
