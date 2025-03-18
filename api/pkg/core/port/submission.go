package port

import "github.com/yokeTH/our-grader-backend/api/pkg/core/domain"

type SubmissionRepository interface {
	Create(s *domain.Submission) error
	GetSubmissionsByID(id uint) (domain.Submission, error)
	GetSubmissionsByUserIDAndProblemID(email string, pid uint, limit int, page int) ([]domain.Submission, int, int, error)
}
