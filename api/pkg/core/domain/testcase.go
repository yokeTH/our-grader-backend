package domain

import (
	"gorm.io/gorm"
)

type TestcaseResult string

const (
	TestResultNotStarted TestcaseResult = "NOT_STARTED"
	TestResultPass       TestcaseResult = "PASS"
	TestResultFail       TestcaseResult = "FAIL"
	TestResultCompile    TestcaseResult = "COMPILE_ERROR"
)

type Testcase struct {
	gorm.Model
	Name         string
	SubmissionID uint
	Submission   Submission     `gorm:"foreignKey:SubmissionID"`
	Result       TestcaseResult `gorm:"default:NOT_STARTED"`
}
