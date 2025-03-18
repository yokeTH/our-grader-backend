package dto

type ProblemRequestFrom struct {
	Name         string   `form:"name" validate:"required,min=2,max=40"`
	Description  string   `form:"description" validate:"required,omitempty"`
	Language     []string `form:"language" validate:"required,min=1"`
	TestcaseNum  uint     `form:"testcase_num" validate:"required,min=1"`
	EditableFile []string `form:"editable_file" validate:"required"`
}
