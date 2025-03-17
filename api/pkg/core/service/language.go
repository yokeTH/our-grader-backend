package service

import (
	"github.com/yokeTH/our-grader-backend/api/pkg/core/domain"
	"github.com/yokeTH/our-grader-backend/api/pkg/repository"
)

type LanguageService struct {
	languageRepository *repository.LanguageRepository
}

func NewLanguageService(repo *repository.LanguageRepository) *LanguageService {
	return &LanguageService{languageRepository: repo}
}

func (s *LanguageService) Create(l string) (domain.Language, error) {
	language := domain.Language{
		Name: l,
	}

	if err := s.languageRepository.Create(&language); err != nil {
		return language, err
	}

	return language, nil
}

func (s *LanguageService) GetAll(limit int, page int) ([]domain.Language, int, int, error) {
	return s.languageRepository.GetLanguages(limit, page)
}
