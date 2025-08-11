package auth

import (
	"net/http"
	"os"
	"strings"
	"time"

	"shakehandz-api/internal/shared/crypto"

	"errors" // ← 追加

	"github.com/gin-gonic/gin"
	"github.com/google/uuid" // ← 追加
	"google.golang.org/api/idtoken"
	"gorm.io/gorm"
)

// 受信JSON
type consentRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type GoogleConsentHandler struct {
	DB *gorm.DB
}

func NewGoogleConsentHandler(db *gorm.DB) *GoogleConsentHandler {
	return &GoogleConsentHandler{DB: db}
}

// 期待ヘッダ：
//   - X-App-Auth: <BACKEND_INTERNAL_TOKEN>（ミドルウェアで検証）
//   - Authorization: Bearer <id_token>（本人性検証）
func (h *GoogleConsentHandler) SaveRefreshToken(c *gin.Context) {
	// 1) id_token 検証
	authz := c.GetHeader("Authorization")
	if !strings.HasPrefix(authz, "Bearer ") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing id_token"})
		return
	}
	idToken := strings.TrimPrefix(authz, "Bearer ")
	aud := os.Getenv("GOOGLE_CLIENT_ID")
	payload, err := idtoken.Validate(c.Request.Context(), idToken, aud)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid id_token"})
		return
	}
	sub, _ := payload.Claims["sub"].(string)
	email, _ := payload.Claims["email"].(string)
	if sub == "" && email == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid claims"})
		return
	}

	// 2) リクエストBody
	var req consentRequest
	if err := c.BindJSON(&req); err != nil || req.RefreshToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "refresh_token required"})
		return
	}

	// 3) ユーザー取得（なければ作成: Upsert）
	var user User
	if err := h.DB.Where("google_id = ?", sub).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// payloadから任意で名前/アイコンも拾う
			name, _ := payload.Claims["name"].(string)
			picture, _ := payload.Claims["picture"].(string)

			// 新規作成
			user = User{
				ID:       uuid.New(),
				GoogleID: sub,
				Email:    email,
				Name:     name,
				Picture:  picture,
			}
			if err3 := h.DB.Create(&user).Error; err3 != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "user create failed"})
				return
			}

		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
			return
		}
	}

	// 4) 暗号化
	enc, err := crypto.EncryptToBytes(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 5) Upsert（(user_id, provider) 一意 & (provider, sub) 一意の方針）
	const provider = "google"
	now := time.Now()

	var token OAuthToken
	err = h.DB.Where("user_id = ? AND provider = ?", user.ID, provider).First(&token).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}

	scope := "openid email profile https://www.googleapis.com/auth/gmail.readonly"

	if token.ID == 0 {
		token = OAuthToken{
			UserID:       user.ID,
			Provider:     provider,
			Sub:          sub,
			Scope:        &scope,
			RefreshToken: enc,
			TokenType:    "Bearer",
		}
		if err := h.DB.Create(&token).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "create failed"})
			return
		}
	} else {
		// 既存更新
		token.Sub = sub
		token.RefreshToken = enc
		token.RevokedAt = nil
		token.UpdatedAt = now
		if token.Scope == nil || *token.Scope != scope {
			token.Scope = &scope
		}
		if err := h.DB.Save(&token).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "update failed"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
