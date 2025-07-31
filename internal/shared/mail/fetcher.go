package mail

import "context"

// Message DTO（必要最小限で OK。追加は後続タスクで）
type Message struct {
	ID        string
	Subject   string
	From      string
	Date      string
	PlainBody string
	HTMLBody  string
	// Attachments など必要なら追記
}

// メール取得インターフェイス
type Fetcher interface {
	Fetch(ctx context.Context, token, query string, max int64) ([]*Message, error)
}
