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

func (r *SubmissionRepository) GetSubmissionsByID(id uint) (domain.Submission, error) {
	var submissions domain.Submission

	if err := r.db.Preload("SubmissionFile").
		Preload("SubmissionFile.TemplateFile").
		Preload("Language").
		Preload("Problem").
		Preload("Testcases").
		Where("id = ?", id).
		First(&submissions).Error; err != nil {
		return submissions, err
	}
	return submissions, nil
}

func (r *SubmissionRepository) GetSubmissionsByUserIDAndProblemID(email string, pid uint, limit int, page int) ([]domain.Submission, int, int, error) {
	var submissions []domain.Submission
	query := r.db.Preload("SubmissionFile").
		Preload("Language").
		Preload("Problem").
		Preload("Testcases").
		Where("submission_by = ?", email).
		Where("problem_id = ?", pid)
	lastPage, total, err := r.db.Paginate(&submissions, query, limit, page, "id DESC")
	if err != nil {
		return nil, 0, 0, err
	}
	return submissions, lastPage, total, nil
}

func (r *SubmissionRepository) Update(data *domain.Submission) error {
	if err := r.db.Updates(data).Error; err != nil {
		return err
	}
	return nil
}
