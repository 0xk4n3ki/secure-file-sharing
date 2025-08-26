package models

type GoogleUser struct {
	ID             string `json:"id" validate:"required"`
	Email          string `json:"email" validate:"required"`
	Verified_email bool   `json:"verified_email" validate:"required"`
	Name           string `json:"name"`
	Given_name     string `json:"given_name"`
	Family_name    string `json:"family_name"`
	Picture        string `json:"picture"`
	Locale         string `json:"locale"`
}
