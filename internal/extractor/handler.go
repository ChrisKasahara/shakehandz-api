package extractor

import (
	"fmt"
	"shakehandz-api/internal/shared/llm/gemini"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"google.golang.org/api/gmail/v1"
)

type ClientData struct {
	UserID     uuid.UUID
	gmail_svc  *gmail.Service
	gemini_cli *gemini.Client
	UpdatedAt  time.Time
}

var (
	userData  = make(map[uuid.UUID]*ClientData)
	userMutex = sync.RWMutex{}
)

func UpdateClientData(userID uuid.UUID, newGmailService *gmail.Service, newGeminiClient *gemini.Client) {
	userMutex.Lock()
	defer userMutex.Unlock()

	if existing, exists := userData[userID]; exists {
		existing.gmail_svc = newGmailService
		existing.gemini_cli = newGeminiClient
		existing.UpdatedAt = time.Now()
	} else {
		userData[userID] = &ClientData{
			UserID:     userID,
			gmail_svc:  newGmailService,
			gemini_cli: newGeminiClient,
			UpdatedAt:  time.Now(),
		}
	}
}

func GetClientData(userID uuid.UUID) (*ClientData, bool) {
	userMutex.RLock()
	defer userMutex.RUnlock()
	data, exists := userData[userID]
	return data, exists
}

func RefreshExtractorTokenHandler(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := svc.Run(c)

		if err != nil {
			fmt.Println("Error in RefreshExtractorTokenHandler:", err)
			return
		}

	}
}
