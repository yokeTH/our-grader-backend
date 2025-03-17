package domain

import "gorm.io/gorm"

type TemplateFile struct {
	gorm.Model
	Name string
	Key  string
}
