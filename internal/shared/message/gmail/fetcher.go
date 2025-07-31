package message_gmail

import (
	"context"
	"encoding/base64"
	"errors"
	"log"
	"sort"
	"strings"
	"sync"

	msg "shakehandz-api/internal/shared/message"

	"golang.org/x/sync/errgroup"
	"google.golang.org/api/gmail/v1"
)

type Attachment struct {
	Filename     string `json:"filename"`
	Size         int64  `json:"size"`
	AttachmentID string `json:"attachmentId"`
}

type fetcher struct{}

func New() msg.MessageFetcher {
	return &fetcher{}
}

func (fetcher *fetcher) FetchMsg(ctx context.Context, token, query string, max int64) ([]*msg.Message, error) {
	if token == "" {
		return nil, errors.New("token is required")
	}
	if max <= 0 {
		max = 10 // デフォルト値
	}
	return fetcher.Fetch(ctx, token, query, max)
}

// Fetch: メッセージID一覧取得→詳細取得→DTO化
func (fetcher *fetcher) Fetch(ctx context.Context, token, query string, max int64) ([]*msg.Message, error) {
	srv, err := NewService(ctx, token)
	if err != nil {
		return nil, err
	}
	msgsList, err := srv.Users.Messages.List("me").MaxResults(max).Q(query).Do()
	if err != nil {
		return nil, err
	}
	if len(msgsList.Messages) == 0 {
		return []*msg.Message{}, nil
	}

	g := new(errgroup.Group)
	var mu sync.Mutex
	var result []*msg.Message

	for _, m := range msgsList.Messages {
		mid := m.Id
		g.Go(func() error {
			msg, err := srv.Users.Messages.Get("me", mid).Format("full").Do()
			if err != nil {
				return err
			}
			dto, err := parseMessage(msg)
			if err != nil {
				return err
			}
			mu.Lock()
			result = append(result, dto)
			mu.Unlock()
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		log.Printf("fetcher: detail fetch error: %v", err)
		return nil, err
	}
	// 日付降順
	sort.Slice(result, func(i, j int) bool {
		return result[i].Date > result[j].Date
	})
	return result, nil
}

// メッセージ詳細→DTO
func parseMessage(gmsg *gmail.Message) (*msg.Message, error) {
	if gmsg == nil || gmsg.Payload == nil {
		return nil, errors.New("empty message")
	}
	var subject, from, date, to, cc, replyTo string

	for _, h := range gmsg.Payload.Headers {
		switch h.Name {
		case "Subject":
			subject = h.Value
		case "From":
			from = h.Value
		case "Date":
			date = h.Value
		case "To":
			to = h.Value
		case "Cc":
			cc = h.Value
		case "Reply-To": // 返信先
			replyTo = h.Value
		}
	}
	plainBody := ExtractPlainText(gmsg.Payload)
	htmlBody := "" // 必要ならHTML抽出ロジック追加
	return &msg.Message{
		Id:        gmsg.Id,
		Subject:   subject,
		From:      from,
		Date:      date,
		PlainBody: plainBody,
		HtmlBody:  htmlBody,
		To:        to,
		Cc:        cc,
		ReplyTo:   replyTo,
	}, nil
}

// 添付ファイル抽出
func extractAttachments(payload *gmail.MessagePart) []Attachment {
	var atts []Attachment
	if payload == nil {
		return atts
	}
	if payload.Filename != "" && payload.Body != nil && payload.Body.AttachmentId != "" {
		atts = append(atts, Attachment{
			Filename:     payload.Filename,
			Size:         payload.Body.Size,
			AttachmentID: payload.Body.AttachmentId,
		})
	}
	for _, part := range payload.Parts {
		atts = append(atts, extractAttachments(part)...)
	}
	return atts
}

// 本文抽出（text/plain優先）
func ExtractPlainText(payload *gmail.MessagePart) string {
	if payload == nil {
		return ""
	}
	if payload.MimeType == "text/plain" && payload.Body != nil && payload.Body.Data != "" {
		return DecodeBase64URL(payload.Body.Data)
	}
	for _, part := range payload.Parts {
		result := ExtractPlainText(part)
		if result != "" {
			return result
		}
	}
	return ""
}

func DecodeBase64URL(data string) string {
	decoded, err := base64.URLEncoding.DecodeString(strings.ReplaceAll(data, "-", "+"))
	if err != nil {
		return ""
	}
	return string(decoded)
}
