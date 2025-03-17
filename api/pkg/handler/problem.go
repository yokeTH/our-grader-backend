package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/yokeTH/our-grader-backend/api/pkg/apperror"
	"github.com/yokeTH/our-grader-backend/api/pkg/core/port"
	"github.com/yokeTH/our-grader-backend/api/pkg/dto"
)

type ProblemHandler struct {
	problemService port.ProblemService
}

func NewProblemHandler(problemService port.ProblemService) port.ProblemHandler {
	return &ProblemHandler{problemService: problemService}
}

func (h *ProblemHandler) CreateProblem(ctx *fiber.Ctx) error {
	zipFile, err := ctx.FormFile("zip")
	if err != nil {
		return apperror.BadRequestError(err, "invalid request body")
	}
	body := new(dto.ProblemRequestFrom)
	if err := ctx.BodyParser(body); err != nil {
		return apperror.BadRequestError(err, "invalid request body")
	}
	problem, err := h.problemService.CreateProblem(ctx.Context(), *body, zipFile)
	if err != nil {
		if apperror.IsAppError(err) {
			return err
		}
		return apperror.InternalServerError(err, "create problem error")
	}
	return ctx.Status(201).JSON(dto.Success(problem))
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
