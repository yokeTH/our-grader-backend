package main

import (
	"fmt"
	"log"

	"github.com/yokeTH/our-grader-backend/api/internal/core/domain"
	"github.com/yokeTH/our-grader-backend/api/internal/database"
	"github.com/yokeTH/our-grader-backend/api/pkg/config"
)

func main() {
	config := config.Load()

	db, err := database.NewPostgresDB(config.PSQL)
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}

	if err := db.AutoMigrate(
		&domain.Book{},
	); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	fmt.Println("Migration completed")
}
