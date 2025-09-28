package extractor

import "time"

const (
	// バッチ処理ステータス
	// 1回のバッチで処理する最大メール件数
	MaxMessages = 3
	// Geminiに1度に渡すメールの件数
	GeminiChunkSize = 3
	// goroutineの最大同時実行数
	MaxGoroutine = 3
	// 処理可能メールの最大に達した場合の次回バッチ実行までの待機時間
	MaxExecutionDuration = 10 * time.Minute
	// バッチ失敗時の自動復旧可能数
	MaxRetryCount = 3

	// Gmail API関連定数
	// Gmailの1ページあたりの取得件数
	PageSize = 50
	// Gmailの最大ページ数（PageSize x MaxPages 件まで取得）
	MaxPages = 10

	// Geminiモデル
	GeminiModel = "models/gemini-2.5-flash"

	// メール本文の構造化処理
	// 有効期限
	// MessageTTL = 5 * 24 * time.Hour // 5日
	MessageTTL = 30 * time.Minute // 30分
)
