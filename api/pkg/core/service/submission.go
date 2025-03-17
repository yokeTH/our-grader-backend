package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/yokeTH/our-grader-backend/api/pkg/apperror"
	"github.com/yokeTH/our-grader-backend/api/pkg/core/domain"
	"github.com/yokeTH/our-grader-backend/api/pkg/core/port"
	"github.com/yokeTH/our-grader-backend/api/pkg/dto"
	"github.com/yokeTH/our-grader-backend/api/pkg/storage"
)

type SubmissionService struct {
	storage        storage.IStorage
	submissionRepo port.SubmissionRepository
	problemRepo    port.ProblemRepository
}

func NewSubmissionService(storage storage.IStorage, submissionRepo port.SubmissionRepository, problemRepo port.ProblemRepository) *SubmissionService {
	return &SubmissionService{
		storage:        storage,
		problemRepo:    problemRepo,
		submissionRepo: submissionRepo,
	}
}

func (s *SubmissionService) Create(ctx context.Context, by string, body dto.SubmissionRequest) error {
	submissionFiles := make([]*domain.SubmissionFile, len(body.Codes))
	for i, v := range body.Codes {
		submissionFiles[i] = &domain.SubmissionFile{
			TemplateFileID: v.TemplateFileID,
		}
	}

	submission := domain.Submission{
		SubmissionBy:   by,
		SubmissionFile: submissionFiles,
		LanguageName:   body.Language,
		ProblemID:      body.ProblemID,
	}

	if err := s.submissionRepo.Create(&submission); err != nil {
		return apperror.InternalServerError(err, "create submission error")
	}

	for _, v := range body.Codes {
		data := strings.NewReader(v.Code)
		key := fmt.Sprintf("submissions/%d/%d", submission.ID, v.TemplateFileID)
		if err := s.storage.UploadFile(ctx, key, "text/plain", data); err != nil {
			return apperror.InternalServerError(err, "upload error")
		}
	}

	return nil
}
