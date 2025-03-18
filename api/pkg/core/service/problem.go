package service

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"sync"

	"github.com/yokeTH/our-grader-backend/api/pkg/apperror"
	"github.com/yokeTH/our-grader-backend/api/pkg/core/domain"
	"github.com/yokeTH/our-grader-backend/api/pkg/core/port"
	"github.com/yokeTH/our-grader-backend/api/pkg/dto"
	"github.com/yokeTH/our-grader-backend/api/pkg/storage"
)

type ProblemService struct {
	ProblemRepository  port.ProblemRepository
	TemplateRepository port.TemplateRepository
	Storage            storage.IStorage
}

func NewProblemService(p port.ProblemRepository, t port.TemplateRepository, s storage.IStorage) port.ProblemService {
	return &ProblemService{ProblemRepository: p, TemplateRepository: t, Storage: s}
}

func (s *ProblemService) CreateProblem(ctx context.Context, problemBody dto.ProblemRequestFrom, zipFile *multipart.FileHeader) (domain.Problem, error) {
	// Convert languages from the request body to the domain structure
	language := make([]domain.Language, len(problemBody.Language))
	for i, v := range problemBody.Language {
		language[i] = domain.Language{Name: v}
	}

	// Initialize the problem struct
	problem := domain.Problem{
		Name:           problemBody.Name,
		Description:    problemBody.Description,
		AllowLanguage:  language,
		TestcaseNum:    problemBody.TestcaseNum,
		EditableFile:   []domain.TemplateFile{}, // will be populated later
		ProjectZipFile: "",                      // to be set after uploading
	}

	// Save the problem to the repository
	if err := s.ProblemRepository.CreateProblem(&problem); err != nil {
		return domain.Problem{}, apperror.InternalServerError(err, "create problem error")
	}

	// Open the zip file
	fileData, err := zipFile.Open()
	if err != nil {
		return problem, apperror.InternalServerError(err, "can't open zip file")
	}
	defer fileData.Close()

	// Validate that the uploaded file is a ZIP file
	contentType := zipFile.Header.Get("Content-Type")
	if contentType != "application/zip" {
		return problem, apperror.BadRequestError(nil, "Uploaded file is not a ZIP file")
	}

	// Generate a unique key for the project zip file
	fileKey := fmt.Sprintf("problems/%d/zip.zip", problem.ID)

	// Upload the zip file to storage (it is uploaded as is)
	if err := s.Storage.UploadFile(ctx, fileKey, contentType, fileData); err != nil {
		return problem, apperror.InternalServerError(err, "upload zip file error")
	}

	// Read the contents of the zip file
	zipReader, err := zip.NewReader(fileData, zipFile.Size)
	if err != nil {
		return problem, apperror.InternalServerError(err, "failed to read zip file")
	}

	// Prepare for storing file keys for editable files
	var editableFile []domain.TemplateFile
	uploadErrors := make([]error, 0)

	// Use a goroutine to upload editable files concurrently
	var wg sync.WaitGroup
	for _, file := range zipReader.File {
		// Iterate over editable files in problemBody and check if it matches the file in the zip
		for i, editableFileName := range problemBody.EditableFile {
			if file.Name == editableFileName {
				wg.Add(1)
				go func(i int, file *zip.File) {
					defer wg.Done()

					// Open the file from the zip archive
					fileData, err := file.Open()
					if err != nil {
						uploadErrors = append(uploadErrors, fmt.Errorf("failed to open editable file '%s': %v", file.Name, err))
						return
					}
					defer fileData.Close()

					// Uncompress the file (i.e., extract it) and upload it uncompressed
					fileKey := fmt.Sprintf("problems/%d/template/%s", problem.ID, file.Name)

					// Use a buffer to handle uncompressed file data
					buf := new(bytes.Buffer)
					_, err = io.Copy(buf, fileData) // Uncompress by copying the file data to a buffer
					if err != nil {
						uploadErrors = append(uploadErrors, fmt.Errorf("failed to decompress file '%s': %v", file.Name, err))
						return
					}

					// Determine the content type based on the file extension or actual content
					contentType := mime.TypeByExtension(filepath.Ext(file.Name))
					if contentType == "" {
						bufBytes := buf.Bytes()
						contentType = http.DetectContentType(bufBytes)
					}

					// Upload the uncompressed file data to storage (S3)
					if err := s.Storage.UploadFile(ctx, fileKey, contentType, buf); err != nil {
						uploadErrors = append(uploadErrors, fmt.Errorf("failed to upload editable file '%s': %v", file.Name, err))
						return
					}

					// Add the uploaded file to the editable files slice
					editableFile = append(editableFile, domain.TemplateFile{
						Name:      file.Name,
						Key:       fileKey,
						ProblemID: problem.ID,
					})
				}(i, file)
			}
		}
	}

	// Wait for all file uploads to complete
	wg.Wait()

	// If there were any errors during file uploads, return them with context
	if len(uploadErrors) > 0 {
		// Aggregate all errors into a single message for easier debugging
		errorMessage := "The following errors occurred during file uploads:\n"
		for _, err := range uploadErrors {
			errorMessage += fmt.Sprintf("- %v\n", err)
		}
		return problem, apperror.InternalServerError(errors.New(errorMessage), "multiple file upload errors occurred")
	}

	// Set the editable files and project zip file in the problem struct
	problem.EditableFile = editableFile
	problem.ProjectZipFile = fileKey

	// Save Template to database
	// Convert editableFile to a slice of pointers
	editableFilePtrs := make([]*domain.TemplateFile, len(editableFile))
	for i := range editableFile {
		editableFilePtrs[i] = &editableFile[i]
	}
	if err := s.TemplateRepository.CreateMany(editableFilePtrs); err != nil {
		return problem, err
	}

	// Update the problem with editable files
	problem, err = s.ProblemRepository.UpdateProblem(problem.ID, problem)
	if err != nil {
		return problem, apperror.InternalServerError(err, "can't update problem")
	}

	// Return the created/updated problem
	return problem, nil
}

func (s *ProblemService) GetProblems(limit int, page int) ([]domain.Problem, int, int, error) {
	return s.ProblemRepository.GetProblems(limit, page)
}
func (s *ProblemService) GetProblemByID(id uint) (domain.Problem, error) {
	return s.ProblemRepository.GetProblemByID(id)
}
func (s *ProblemService) UpdateProblem(id uint, problem domain.Problem) (domain.Problem, error) {
	return domain.Problem{}, nil
}
func (s *ProblemService) DeleteProblem() error {
	return nil
}
