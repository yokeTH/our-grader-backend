package handler

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/yokeTH/our-grader-backend/api/pkg/apperror"
	"github.com/yokeTH/our-grader-backend/api/pkg/core/service"
	"github.com/yokeTH/our-grader-backend/api/pkg/dto"
)

type SubmissionHandler struct {
	problemService *service.SubmissionService
}

func NewSubmissionHandler(ps *service.SubmissionService) *SubmissionHandler {
	return &SubmissionHandler{problemService: ps}
}

func (h *SubmissionHandler) Submit(c *fiber.Ctx) error {
	body := new(dto.SubmissionRequest)
	if err := c.BodyParser(body); err != nil {
		return apperror.BadRequestError(errors.New("request body invalid"), "request body invalid")
	}

	if err := h.problemService.Create(c.Context(), "6530162621@student.chula.ac.th", *body); err != nil {
		return err
	}

	return c.SendStatus(201)
}
