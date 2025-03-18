package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/yokeTH/our-grader-backend/api/pkg/apperror"
	"github.com/yokeTH/our-grader-backend/api/pkg/core/domain"
	"github.com/yokeTH/our-grader-backend/api/pkg/core/port"
	"github.com/yokeTH/our-grader-backend/api/pkg/dto"
	"github.com/yokeTH/our-grader-backend/api/pkg/storage"
	"github.com/yokeTH/our-grader-backend/proto/verilog"
)

type SubmissionService struct {
	storage        storage.IStorage
	submissionRepo port.SubmissionRepository
	problemRepo    port.ProblemRepository
	client         verilog.SomeServiceClient
}

func NewSubmissionService(storage storage.IStorage, submissionRepo port.SubmissionRepository, problemRepo port.ProblemRepository, client verilog.SomeServiceClient) *SubmissionService {
	return &SubmissionService{
		storage:        storage,
		problemRepo:    problemRepo,
		submissionRepo: submissionRepo,
		client:         client,
	}
}

func (s *SubmissionService) Create(ctx context.Context, by string, body dto.SubmissionRequest) error {
	submissionFiles := make([]*domain.SubmissionFile, len(body.Codes))
	problem, err := s.problemRepo.GetProblemByID(body.ProblemID)
	if err != nil {
		return err
	}
	for i, v := range body.Codes {
		var id uint = 0
		for _, file := range problem.EditableFile {
			if strings.HasSuffix(file.Name, v.TemplateFileName) {
				id = file.ID
				break
			}
		}
		if id == 0 {
			return apperror.BadRequestError(errors.New("mismatch"), "invalid request")
		}
		submissionFiles[i] = &domain.SubmissionFile{
			TemplateFileID: id,
		}
	}

	submission := domain.Submission{
		SubmissionBy:   by,
		SubmissionFile: submissionFiles,
		LanguageName:   body.Language,
		ProblemID:      body.ProblemID,
		Testcases:      make([]domain.Testcase, problem.TestcaseNum),
	}

	if err := s.submissionRepo.Create(&submission); err != nil {
		return apperror.InternalServerError(err, "create submission error")
	}

	for i, v := range body.Codes {
		data := strings.NewReader(v.Code)
		key := fmt.Sprintf("submissions/%d/%d", submission.ID, submissionFiles[i].TemplateFileID)
		if err := s.storage.UploadFile(ctx, key, "text/plain", data); err != nil {
			return apperror.InternalServerError(err, "upload error")
		}
	}

	s.client.Run(ctx, &verilog.VerilogRequest{
		SubmissionID: uint32(submission.ID),
	})

	return nil
}

func (s *SubmissionService) GetSubmissionsByUserIDAndProblemID(email string, pid uint, limit int, page int) ([]domain.Submission, int, int, error) {
	return s.submissionRepo.GetSubmissionsByUserIDAndProblemID(email, pid, limit, page)
}
