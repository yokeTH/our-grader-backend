package main

import (
	"context"
	"fmt"
	"log"

	"github.com/yokeTH/our-grader-backend/api/pkg/config"
	"github.com/yokeTH/our-grader-backend/api/pkg/core/service"
	"github.com/yokeTH/our-grader-backend/api/pkg/database"
	"github.com/yokeTH/our-grader-backend/api/pkg/handler"
	"github.com/yokeTH/our-grader-backend/api/pkg/middleware"
	"github.com/yokeTH/our-grader-backend/api/pkg/repository"
	"github.com/yokeTH/our-grader-backend/api/pkg/server"
	"github.com/yokeTH/our-grader-backend/api/pkg/storage"
	"github.com/yokeTH/our-grader-backend/proto/verilog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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
	conn, err := grpc.NewClient(fmt.Sprintf("%s:50051", config.VerilogServer), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		{
			panic(err)
		}
	}
	verilogGradingServer := verilog.NewSomeServiceClient(conn)

	auth := middleware.NewAuthMiddleware()

	templateRepo := repository.NewTemplateFileRepository(db)
	problemRepo := repository.NewProblemRepository(db)
	problemService := service.NewProblemService(problemRepo, templateRepo, store)
	problemHandler := handler.NewProblemHandler(problemService)

	languageRepo := repository.NewLanguageRepository(db)
	languageService := service.NewLanguageService(languageRepo)
	languageHandler := handler.NewLanguageHandler(languageService)

	submissionRepo := repository.NewSubmissionRepository(db)
	submissionService := service.NewSubmissionService(store, submissionRepo, problemRepo, verilogGradingServer)
	submissionHandler := handler.NewSubmissionHandler(submissionService)

	s := server.New(
		server.WithName(config.Server.Name),
		server.WithBodyLimitMB(config.Server.BodyLimitMB),
		server.WithPort(config.Server.Port),
	)

	problemRoute := s.App.Group("/problems")
	problemRoute.Get("/", auth.Auth, problemHandler.GetProblems)
	problemRoute.Get("/:id", auth.Auth, problemHandler.GetProblemByID)
	problemRoute.Post("/", auth.Auth, auth.Owner, problemHandler.CreateProblem)

	languageRoute := s.App.Group("/languages")
	languageRoute.Get("/", auth.Auth, auth.Owner, languageHandler.GetAll)
	languageRoute.Post("/", auth.Auth, auth.Owner, languageHandler.Create)

	submissionRoute := s.App.Group("/submissions")
	submissionRoute.Post("/", auth.Auth, submissionHandler.Submit)
	submissionRoute.Get("/problem/:problemID", auth.Auth, submissionHandler.GetSubmissions)

	s.Start(ctx, stop)
}
