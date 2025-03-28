package domain

import "gorm.io/gorm"

type SubmissionFile struct {
	gorm.Model
	TemplateFile   TemplateFile `gorm:"foreignKey:TemplateFileID"`
	TemplateFileID uint
	SubmissionID   uint
	Submission     Submission `gorm:"foreignKey:SubmissionID"`
}
