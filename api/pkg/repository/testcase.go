package repository

import (
	"github.com/yokeTH/our-grader-backend/api/pkg/core/domain"
	"github.com/yokeTH/our-grader-backend/api/pkg/database"
)

type TestcaseRepository struct {
	db *database.Database
}

func NewTestcaseRepository(db *database.Database) *TestcaseRepository {
	return &TestcaseRepository{
		db: db,
	}
}

func (r *TestcaseRepository) CreateTestcase(testcase *domain.Testcase) error {
	result := r.db.Create(testcase)
	err := result.Error
	if err != nil {
		return err
	}
	return nil
}

func (r *TestcaseRepository) GetTestcase(testcaseID uint) (domain.Testcase, error) {
	var testcase domain.Testcase
	err := r.db.First(&testcase, testcaseID).Error
	if err != nil {
		return testcase, err
	}
	return testcase, nil
}

func (r *TestcaseRepository) GetAllTestcases() ([]domain.Testcase, error) {
	var testcases []domain.Testcase
	err := r.db.Find(&testcases).Error
	if err != nil {
		return testcases, err
	}
	return testcases, nil
}

func (r *TestcaseRepository) GetTestcasesBySubmissionID(submissionID uint) ([]domain.Testcase, error) {
	var testcases []domain.Testcase
	err := r.db.Where("submission_id = ?", submissionID).Find(&testcases).Error
	if err != nil {
		return testcases, err
	}
	return testcases, nil
}

func (r *TestcaseRepository) UpdateTestcase(testcase *domain.Testcase) error {
	err := r.db.Model(&domain.Testcase{}).Where("id = ?", testcase.ID).Updates(testcase).Error
	if err != nil {
		return err
	}
	return nil
}
