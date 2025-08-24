package oauth

import (
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
)

const (
	// Gemini APIに必要なスコープ
	GenerativeLanguageScope = "https://www.googleapis.com/auth/generative-language.retriever"
)

func OAuth2ConfigFromEnv() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		Endpoint:     google.Endpoint,
		Scopes:       []string{"openid", "email", "profile", gmail.GmailReadonlyScope, GenerativeLanguageScope},
	}
}
