package main

import (
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"rainbow-backend/internal/config"
	"rainbow-backend/internal/model"
	"rainbow-backend/internal/router"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	gin.SetMode(cfg.GinMode())

	db, err := model.OpenDB(cfg.Database)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}

	if err := model.AutoMigrate(db); err != nil {
		log.Fatalf("auto migrate: %v", err)
	}

	if err := model.SeedAdmin(db, cfg.AdminUsername, cfg.AdminPassword); err != nil {
		log.Fatalf("seed admin: %v", err)
	}

	engine := router.New(cfg, db)
	server := &http.Server{
		Addr:              cfg.Address(),
		Handler:           engine,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("server listening on %s", cfg.Address())
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("server stopped: %v", err)
	}
}
