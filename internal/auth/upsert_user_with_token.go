package auth

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"shakehandz-api/internal/shared/crypto"
)

// AuthServiceはユーザー関連のビジネスロジックを担当します
type AuthService struct {
	db *gorm.DB
}

// NewAuthServiceは新しいAuthServiceを生成します
func NewAuthService(db *gorm.DB) *AuthService {
	return &AuthService{db: db}
}

// UpsertUserWithTokenはGORMのトランザクションを使い、ユーザーとOAuthトークンを作成または更新します
func (s *AuthService) UpsertUserWithToken(req UpsertUserRequest) error {
	// トランザクションを開始します。中の処理でエラーが発生した場合、自動的にロールバックされます。
	return s.db.Transaction(func(tx *gorm.DB) error {
		// --- 1. OAuthTokenを起点にユーザーを検索または準備 ---
		var token OauthToken
		// ProviderとSub(GoogleID)で既存のトークンを検索、なければメモリ上に準備
		tx.Where(OauthToken{Provider: "google", Sub: req.GoogleID}).FirstOrInit(&token)

		var user User
		// トークンが既に存在する場合（既存ユーザー）
		if token.UserID != uuid.Nil {
			// トークンに紐づくUserIDを使ってユーザー情報を取得
			if err := tx.First(&user, token.UserID).Error; err != nil {
				return err // ユーザーが見つからない場合はエラー（データ不整合の可能性）
			}
		}

		// --- 2. UserテーブルのUpsert ---
		// ユーザー情報をリクエストデータで更新
		user.Email = req.Email
		user.Name = req.Name
		user.Image = req.Image

		// ユーザーが新規作成の場合、UUIDを付与
		if user.ID == uuid.Nil {
			user.ID = uuid.New()
		}

		// ユーザー情報を保存
		if err := tx.Save(&user).Error; err != nil {
			return err // エラーが発生したらロールバック
		}

		// --- 3. OAuthTokenテーブルのUpsert ---
		// リフレッシュトークンを暗号化
		encRefreshToken, err := crypto.EncryptToBytes(req.RefreshToken)
		if err != nil {
			return err
		}

		// トークン情報を更新
		token.UserID = user.ID
		token.Provider = "google"
		token.Sub = req.GoogleID
		token.RefreshToken = encRefreshToken
		token.Scope = &req.Scope
		token.RevokedAt = nil

		// トークン情報を保存
		if err := tx.Save(&token).Error; err != nil {
			return err // エラーが発生したらロールバック
		}

		// 全て成功したらコミット
		return nil
	})
}
