package port

import (
	"mime/multipart"

	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/yokeTH/our-grader-backend/api/pkg/core/domain"
	"github.com/yokeTH/our-grader-backend/api/pkg/dto"
)

type ProblemHandler interface {
	CreateProblem(ctx *fiber.Ctx) error
	GetProblems(ctx *fiber.Ctx) error
	GetProblemByID(ctx *fiber.Ctx) error
	UpdateProblem(ctx *fiber.Ctx) error
	DeleteProblem(ctx *fiber.Ctx) error
}

type ProblemService interface {
	CreateProblem(ctx context.Context, problem dto.ProblemRequestFrom, zip *multipart.FileHeader) (domain.Problem, error)
	GetProblemByID(id uint) (domain.Problem, error)
	GetProblems(limit int, page int) ([]domain.Problem, int, int, error)
	UpdateProblem(id uint, problem domain.Problem) (domain.Problem, error)
	DeleteProblem() error
}

type ProblemRepository interface {
	CreateProblem(problem *domain.Problem) error
	GetProblems(limit int, page int) ([]domain.Problem, int, int, error)
	GetProblemByID(id uint) (domain.Problem, error)
	UpdateProblem(id uint, problem domain.Problem) (domain.Problem, error)
	DeleteProblem(id uint) error
}
