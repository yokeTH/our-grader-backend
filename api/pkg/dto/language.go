package dto

type LanguageCreateRequest struct {
	Name string `json:"name"`
}

type LanguageResponse struct {
	Name string `json:"name"`
}
