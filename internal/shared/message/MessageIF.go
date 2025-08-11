package message

import "context"

type Message struct {
	Id          string       `json:"id"`
	Subject     string       `json:"subject"`
	From        string       `json:"from"`
	Date        string       `json:"date"`
	PlainBody   string       `json:"plain_body"`
	HtmlBody    string       `json:"html_body"`
	To          string       `json:"to"`
	Cc          string       `json:"cc"`
	ReplyTo     string       `json:"reply_to"`
	Attachments []Attachment `json:"attachments,omitempty"`
}

type Attachment struct {
	Filename     string `json:"filename"`
	Size         int64  `json:"size"`
	AttachmentID string `json:"attachment_id"`
}

type MessageFetcher interface {
	FetchMsg(ctx context.Context, token, query string, max int64) ([]*Message, error)
	FetchMsgWithPaging(ctx context.Context, token, query string, pageSize int64, pageToken string) ([]*Message, string, error)
}
