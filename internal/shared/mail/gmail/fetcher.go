package gmailfetcher

import (
	"context"
	"encoding/base64"
	"errors"
	"log"
	"sort"
	"strings"
	"sync"

	mail "shakehandz-api/internal/shared/mail"

	"golang.org/x/sync/errgroup"
	"google.golang.org/api/gmail/v1"
)

type Attachment struct {
	Filename     string `json:"filename"`
	Size         int64  `json:"size"`
	AttachmentID string `json:"attachmentId"`
}

type fetcher struct{}

func New() mail.Fetcher {
	return &fetcher{}
}

// Fetch: メッセージID一覧取得→詳細取得→DTO化
func (f *fetcher) Fetch(ctx context.Context, token, query string, max int64) ([]*mail.Message, error) {
	srv, err := NewService(ctx, token)
	if err != nil {
		return nil, err
	}
	msgsList, err := srv.Users.Messages.List("me").MaxResults(max).Q(query).Do()
	if err != nil {
		return nil, err
	}
	if len(msgsList.Messages) == 0 {
		return []*mail.Message{}, nil
	}

	g := new(errgroup.Group)
	var mu sync.Mutex
	var result []*mail.Message

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
func parseMessage(msg *gmail.Message) (*mail.Message, error) {
	if msg == nil || msg.Payload == nil {
		return nil, errors.New("empty message")
	}
	subject, from, date := "", "", ""
	for _, h := range msg.Payload.Headers {
		switch h.Name {
		case "Subject":
			subject = h.Value
		case "From":
			from = h.Value
		case "Date":
			date = h.Value
		}
	}
	plainBody := ExtractPlainText(msg.Payload)
	htmlBody := "" // 必要ならHTML抽出ロジック追加
	return &mail.Message{
		ID:        msg.Id,
		Subject:   subject,
		From:      from,
		Date:      date,
		PlainBody: plainBody,
		HTMLBody:  htmlBody,
		// Attachments等は後続タスクで
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
