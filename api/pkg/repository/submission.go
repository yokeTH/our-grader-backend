package repository

import (
	"github.com/yokeTH/our-grader-backend/api/pkg/core/domain"
	"github.com/yokeTH/our-grader-backend/api/pkg/database"
)

type SubmissionRepository struct {
	db *database.Database
}

func NewSubmissionRepository(db *database.Database) *SubmissionRepository {
	return &SubmissionRepository{db: db}
}

func (r *SubmissionRepository) Create(s *domain.Submission) error {
	if err := r.db.Create(s).Error; err != nil {
		return err
	}
	return nil
}

func (r *SubmissionRepository) GetSubmissionsByUserIDAndProblemID(email string, pid uint, limit int, page int) ([]domain.Submission, int, int, error) {
	var submissions []domain.Submission
	query := r.db.Preload("SubmissionFile").
		Preload("Language").
		Preload("Problem").
		Where("submission_by = ?", email).
		Where("problem_id = ?", pid)
	lastPage, total, err := r.db.Paginate(&submissions, query, limit, page, "id ASC")
	if err != nil {
		return nil, 0, 0, err
	}
	return submissions, lastPage, total, nil
}
