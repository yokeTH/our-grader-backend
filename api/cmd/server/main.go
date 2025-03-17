package main

import (
	"context"
	"log"

	"github.com/yokeTH/our-grader-backend/api/pkg/config"
	"github.com/yokeTH/our-grader-backend/api/pkg/core/service"
	"github.com/yokeTH/our-grader-backend/api/pkg/database"
	"github.com/yokeTH/our-grader-backend/api/pkg/handler"
	"github.com/yokeTH/our-grader-backend/api/pkg/repository"
	"github.com/yokeTH/our-grader-backend/api/pkg/server"
	"github.com/yokeTH/our-grader-backend/api/pkg/storage"
)

func main() {
	ctx, stop := context.WithCancel(context.Background())
	defer stop()

	config := config.Load()

	db, err := database.NewPostgresDB(config.PSQL)
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	store, err := storage.NewR2Storage(config.R2)
	if err != nil {
		log.Fatalf("failed to create storage: %v", err)
	}

	problemRepo := repository.NewProblemRepository(db)
	problemService := service.NewProblemService(problemRepo, store)
	problemHandler := handler.NewProblemHandler(problemService)

	languageRepo := repository.NewLanguageRepository(db)
	languageService := service.NewLanguageService(languageRepo)
	languageHandler := handler.NewLanguageHandler(languageService)

	s := server.New(
		server.WithName(config.Server.Name),
		server.WithBodyLimitMB(config.Server.BodyLimitMB),
		server.WithPort(config.Server.Port),
	)

	problemRoute := s.App.Group("/problems")
	problemRoute.Post("/", problemHandler.CreateProblem)

	languageRoute := s.App.Group("/languages")
	languageRoute.Get("/", languageHandler.GetAll)
	languageRoute.Post("/", languageHandler.Create)

	s.Start(ctx, stop)
}
