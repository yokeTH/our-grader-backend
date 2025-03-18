package domain

type Profile struct {
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	FamilyName    string `json:"family_name"`
	GivenName     string `json:"given_name"`
	HD            string `json:"hd"`
	Picture       string `json:"picture"`
	Sub           string `json:"sub"`
}
