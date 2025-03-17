package domain

import (
	"gorm.io/gorm"
)

type Submission struct {
	gorm.Model
	SubmissionBy    string
	SubmissionFile  []SubmissionFile
	LanguageName    string
	Language        Language `gorm:"foreignKey:LanguageName;references:Name"`
	StdoutObjectKey string
	Additional      string
	ProblemID       uint
	Problem         Problem `gorm:"foreignKey:ProblemID"`
	MemoryUsageMB   uint
	Testcases       []Testcase
}
