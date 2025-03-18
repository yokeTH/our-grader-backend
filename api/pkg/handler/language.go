package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/yokeTH/our-grader-backend/api/pkg/apperror"
	"github.com/yokeTH/our-grader-backend/api/pkg/core/service"
	"github.com/yokeTH/our-grader-backend/api/pkg/dto"
)

type LanguageHandler struct {
	languageService *service.LanguageService
}

func NewLanguageHandler(languageService *service.LanguageService) *LanguageHandler {
	return &LanguageHandler{languageService: languageService}
}

func (h *LanguageHandler) Create(c *fiber.Ctx) error {
	l := new(dto.LanguageCreateRequest)

	if err := c.BodyParser(l); err != nil {
		return apperror.BadRequestError(err, "invalid request body")
	}

	language, err := h.languageService.Create(l.Name)
	if err != nil {
		if apperror.IsAppError(err) {
			return err
		}
		return apperror.InternalServerError(err, "create language error")
	}

	return c.Status(201).JSON(dto.Success(language.ToDTO()))
}

func (h *LanguageHandler) GetAll(c *fiber.Ctx) error {
	languages, last, total, err := h.languageService.GetAll(100, 1)

	if err != nil {
		if apperror.IsAppError(err) {
			return err
		}
		return apperror.InternalServerError(err, "get all language error")
	}

	data := make([]dto.LanguageResponse, len(languages))
	for i, v := range languages {
		data[i] = v.ToDTO()
	}

	return c.JSON(dto.SuccessPagination(data, dto.Pagination{
		CurrentPage: 1,
		LastPage:    last,
		Limit:       100,
		Total:       total,
	}))
}
