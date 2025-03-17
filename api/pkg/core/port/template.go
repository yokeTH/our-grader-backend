package port

import "github.com/yokeTH/our-grader-backend/api/pkg/core/domain"

type TemplateRepository interface {
	Create(template *domain.TemplateFile) error
	CreateMany(template []*domain.TemplateFile) error
}
