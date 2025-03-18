package domain

import (
	"gorm.io/gorm"
)

type Problem struct {
	gorm.Model
	Name           string
	Description    string
	AllowLanguage  []Language `gorm:"many2many:support_languages;"`
	TestcaseNum    uint
	EditableFile   []TemplateFile
	ProjectZipFile string
}
