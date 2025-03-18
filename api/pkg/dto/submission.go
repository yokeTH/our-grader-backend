package dto

type SubmissionRequest struct {
	Language  string     `json:"language"`
	ProblemID uint       `json:"problem_id"`
	Codes     []CodeFile `json:"codes"`
}

type CodeFile struct {
	Code             string `json:"code"`
	TemplateFileName string `json:"template_name"`
}
