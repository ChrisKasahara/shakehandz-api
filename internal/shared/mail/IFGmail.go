package mail

import "context"

type Message struct {
	ID        string
	Subject   string
	From      string
	Date      string
	PlainBody string
	HTMLBody  string
}

type Fetcher interface {
	Fetch(ctx context.Context, token, query string, max int64) ([]*Message, error)
}
