package domain

import "gorm.io/gorm"

type TemplateFile struct {
	gorm.Model
	Name      string
	Key       string
	ProblemID uint
	Problem   Problem `gorm:"foreignKey:ProblemID"`
}
