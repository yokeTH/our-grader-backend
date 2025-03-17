package repository

import (
	"github.com/yokeTH/our-grader-backend/api/pkg/core/domain"
	"github.com/yokeTH/our-grader-backend/api/pkg/database"
)

type ProblemRepository struct {
	db *database.Database
}

func NewProblemRepository(db *database.Database) *ProblemRepository {
	return &ProblemRepository{db: db}
}

func (r *ProblemRepository) CreateProblem(problem *domain.Problem) error {
	if err := r.db.Create(problem).Error; err != nil {
		return err
	}
	return nil
}

func (r *ProblemRepository) GetProblems(limit int, page int) ([]domain.Problem, int, int, error) {
	var problems []domain.Problem
	query := r.db.Preload("EditableFile").
		Preload("AllowLanguage")
	lastPage, total, err := r.db.Paginate(&problems, query, limit, page, "id ASC")
	if err != nil {
		return nil, 0, 0, err
	}
	return problems, lastPage, total, nil

}

func (r *ProblemRepository) GetProblemByID(id uint) (domain.Problem, error) {
	var problem *domain.Problem
	if err := r.db.Model(&problem).First(&problem).Where("id = ?", id).Error; err != nil {
		return *problem, err
	}
	return *problem, nil
}

func (r *ProblemRepository) UpdateProblem(id uint, updateProblem domain.Problem) (domain.Problem, error) {
	var problem domain.Problem
	if err := r.db.Model(&domain.Problem{}).Preload("EditableFile").Preload("AllowLanguage").Where("id = ?", id).Updates(updateProblem).First(&problem).Error; err != nil {
		return problem, err
	}
	return problem, nil
}

func (r *ProblemRepository) DeleteProblem(id uint) error {
	if err := r.db.Delete(&domain.Problem{}, id).Error; err != nil {
		return err
	}
	return nil
}
