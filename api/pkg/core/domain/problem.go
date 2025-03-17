package domain

import (
	"gorm.io/gorm"
)

type Problem struct {
	gorm.Model
	Name           string
	Description    string
	AllowLanguage  []Language
	TestcaseNum    uint
	EditableFile   []TemplateFile
	ProjectZipFile string
}
