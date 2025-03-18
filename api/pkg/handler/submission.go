package handler

import (
	"errors"
	"math"

	"github.com/gofiber/fiber/v2"
	"github.com/yokeTH/our-grader-backend/api/pkg/apperror"
	"github.com/yokeTH/our-grader-backend/api/pkg/core/domain"
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
	profile := c.Locals("profile").(*domain.Profile)
	if err := c.BodyParser(body); err != nil {
		return apperror.BadRequestError(errors.New("request body invalid"), "request body invalid")
	}

	if err := h.problemService.Create(c.Context(), profile.Email, *body); err != nil {
		return err
	}

	return c.Status(201).JSON(fiber.Map{"success": true})
}

func (h *SubmissionHandler) GetSubmissions(c *fiber.Ctx) error {
	profile := c.Locals("profile").(domain.Profile)
	limit := math.Min(float64(c.QueryInt("limit", 10)), 50)
	page := c.QueryInt("limit", 1)
	problemID, err := c.ParamsInt("problemID")
	if err != nil {
		return apperror.BadRequestError(err, "invalid problem ID")
	}

	data, last, total, err := h.problemService.GetSubmissionsByUserIDAndProblemID(profile.Email, uint(problemID), int(limit), page)
	if err != nil {
		return apperror.InternalServerError(err, "get submission error")
	}

	return c.JSON(dto.SuccessPagination(data, dto.Pagination{
		CurrentPage: page,
		LastPage:    last,
		Limit:       int(limit),
		Total:       total,
	}))
}
