package repository

import (
	"github.com/yokeTH/our-grader-backend/api/pkg/apperror"
	"github.com/yokeTH/our-grader-backend/api/pkg/core/domain"
	"github.com/yokeTH/our-grader-backend/api/pkg/database"
)

type LanguageRepository struct {
	db *database.Database
}

func NewLanguageRepository(db *database.Database) *LanguageRepository {
	return &LanguageRepository{db: db}
}

func (r *LanguageRepository) Create(language *domain.Language) error {
	if err := r.db.Create(language).Error; err != nil {
		return apperror.InternalServerError(err, "can't create language")
	}
	return nil
}

func (r *LanguageRepository) GetLanguages(limit int, page int) ([]domain.Language, int, int, error) {
	var languages []domain.Language
	lastPage, total, err := r.db.Paginate(&languages, r.db.DB, limit, page, "name ASC")
	if err != nil {
		return nil, 0, 0, apperror.InternalServerError(err, "can't get languages")
	}
	return languages, lastPage, total, nil
}
