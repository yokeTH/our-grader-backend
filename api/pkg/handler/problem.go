package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/yokeTH/our-grader-backend/api/pkg/core/port"
)

type ProblemHandler struct {
	problemService port.ProblemService
}

func NewProblemHandler(problemService port.ProblemService) port.ProblemHandler {
	return &ProblemHandler{problemService: problemService}
}

func (h *ProblemHandler) CreateProblem(ctx *fiber.Ctx) error {
	return nil
}

func (h *ProblemHandler) GetProblems(ctx *fiber.Ctx) error {
	return nil
}

func (h *ProblemHandler) GetProblemByID(ctx *fiber.Ctx) error {
	return nil
}

func (h *ProblemHandler) UpdateProblem(ctx *fiber.Ctx) error {
	return nil
}

func (h *ProblemHandler) DeleteProblem(ctx *fiber.Ctx) error {
	return nil
}
