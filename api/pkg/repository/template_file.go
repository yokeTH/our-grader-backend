package repository

import (
	"github.com/yokeTH/our-grader-backend/api/pkg/core/domain"
	"github.com/yokeTH/our-grader-backend/api/pkg/database"
)

type TemplateFileRepository struct {
	db *database.Database
}

func NewTemplateFileRepository(db *database.Database) *TemplateFileRepository {
	return &TemplateFileRepository{db: db}
}

func (r *TemplateFileRepository) Create(template *domain.TemplateFile) error {
	if err := r.db.Create(template).Error; err != nil {
		return err
	}
	return nil
}

func (r *TemplateFileRepository) CreateMany(template []*domain.TemplateFile) error {
	if err := r.db.Create(template).Error; err != nil {
		return err
	}
	return nil
}
