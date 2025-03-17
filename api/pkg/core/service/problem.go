package service

import (
	"github.com/yokeTH/our-grader-backend/api/pkg/core/domain"
	"github.com/yokeTH/our-grader-backend/api/pkg/core/port"
	"github.com/yokeTH/our-grader-backend/api/pkg/dto"
	"github.com/yokeTH/our-grader-backend/api/pkg/storage"
)

type ProblemService struct {
	ProblemRepository port.ProblemRepository
	Storage           storage.IStorage
}

func NewProblemService(p port.ProblemRepository, s storage.IStorage) port.ProblemService {
	return &ProblemService{ProblemRepository: p, Storage: s}
}

func (s *ProblemService) CreateProblem(problem dto.ProblemRequestFrom) (domain.Problem, error) {
	return domain.Problem{}, nil
}

func (s *ProblemService) GetProblems(limit int, page int) ([]domain.Problem, error) {
	return []domain.Problem{}, nil
}
func (s *ProblemService) GetProblemByID() (domain.Problem, error) {
	return domain.Problem{}, nil
}
func (s *ProblemService) UpdateProblem(id uint, problem domain.Problem) (domain.Problem, error) {
	return domain.Problem{}, nil
}
func (s *ProblemService) DeleteProblem() error {
	return nil
}
