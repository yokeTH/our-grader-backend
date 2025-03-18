package domain

import "github.com/yokeTH/our-grader-backend/api/pkg/dto"

type Language struct {
	Name string `gorm:"primarykey"`
}

func (l *Language) ToDTO() dto.LanguageResponse {
	return dto.LanguageResponse{Name: l.Name}
}
