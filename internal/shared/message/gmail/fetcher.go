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
	"golang.org/x/sync/semaphore"
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

// FetchMsg: メッセージID一覧取得→詳細取得→DTO化（refresh対応）
func (fetcher *fetcher) FetchMsg(ctx context.Context, encRefresh []byte, query string, max int64) ([]*msg.Message, error) {
	if len(encRefresh) == 0 {
		return nil, errors.New("refresh token is required")
	}
	if max <= 0 {
		max = 10 // デフォルト値
	}
	return fetcher.Fetch(ctx, encRefresh, query, max)
}

// FetchMsgWithPaging: ページング対応のメッセージ取得（refresh対応）
func (fetcher *fetcher) FetchMsgWithPaging(ctx context.Context, encRefresh []byte, query string, pageSize int64, pageToken string) ([]*msg.Message, string, error) {
	if len(encRefresh) == 0 {
		return nil, "", errors.New("refresh token is required")
	}
	if pageSize <= 0 {
		pageSize = 50 // デフォルト値
	}
	return fetcher.FetchWithPaging(ctx, encRefresh, query, pageSize, pageToken)
}

// Fetch: メッセージID一覧取得→詳細取得→DTO化（refresh対応）
func (fetcher *fetcher) Fetch(ctx context.Context, encRefresh []byte, query string, max int64) ([]*msg.Message, error) {
	srv, err := NewServiceWithRefresh(ctx, encRefresh)
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
	return fetcher.fetchMessageDetails(ctx, srv, msgsList.Messages)
}

// FetchWithPaging: ページング対応のメッセージ取得（refresh対応）
func (fetcher *fetcher) FetchWithPaging(ctx context.Context, encRefresh []byte, query string, pageSize int64, pageToken string) ([]*msg.Message, string, error) {
	srv, err := NewServiceWithRefresh(ctx, encRefresh)
	if err != nil {
		return nil, "", err
	}

	listCall := srv.Users.Messages.List("me").MaxResults(pageSize).Q(query)
	if pageToken != "" {
		listCall = listCall.PageToken(pageToken)
	}

	msgsList, err := listCall.Do()
	if err != nil {
		return nil, "", err
	}
	if len(msgsList.Messages) == 0 {
		return []*msg.Message{}, msgsList.NextPageToken, nil
	}

	result, err := fetcher.fetchMessageDetails(ctx, srv, msgsList.Messages)
	if err != nil {
		return nil, "", err
	}
	return result, msgsList.NextPageToken, nil
}

// fetchMessageDetails: メッセージ詳細の並列取得処理
func (fetcher *fetcher) fetchMessageDetails(ctx context.Context, srv *gmail.Service, messages []*gmail.Message) ([]*msg.Message, error) {
	g, ctx := errgroup.WithContext(ctx) // コンテキスト付きのerrgroupを使用
	var mu sync.Mutex
	var result []*msg.Message

	// 同時に実行するリクエスト数を10に制限
	sem := semaphore.NewWeighted(10)

	for _, m := range messages {
		mid := m.Id
		if err := sem.Acquire(ctx, 1); err != nil {
			log.Printf("Failed to acquire semaphore: %v", err)
			break
		}
		g.Go(func() error {
			defer sem.Release(1)

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
		case "Reply-To":
			replyTo = h.Value
		}
	}
	// 2. 本文の抽出 (プレーンテキストとHTMLの両方)
	plainBody := ExtractBody(gmsg.Payload, "text/plain")
	htmlBody := ExtractBody(gmsg.Payload, "text/html")

	var msgAttachments []msg.Attachment
	for _, att := range ExtractAttachments(gmsg.Payload) {
		msgAttachments = append(msgAttachments, msg.Attachment{
			Filename:     att.Filename,
			Size:         att.Size,
			AttachmentID: att.AttachmentID,
		})
	}
	return &msg.Message{
		Id:          gmsg.Id,
		Subject:     subject,
		From:        from,
		Date:        date,
		PlainBody:   plainBody,
		HtmlBody:    htmlBody,
		To:          to,
		Cc:          cc,
		ReplyTo:     replyTo,
		Attachments: msgAttachments,
	}, nil
}

// 添付ファイル抽出
func ExtractAttachments(payload *gmail.MessagePart) []Attachment {
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
		atts = append(atts, ExtractAttachments(part)...)
	}
	return atts
}

// ExtractBody は、指定されたMIMEタイプの本文を再帰的に探し、デコードして返します。
func ExtractBody(part *gmail.MessagePart, mimeType string) string {
	// fmt.Printf("Extracting body for MIME type: %s\n", mimeType)
	// fmt.Printf("Part: %#v\n", part)
	if part.MimeType == mimeType && part.Body != nil && part.Body.Data != "" {
		data, err := base64.URLEncoding.DecodeString(part.Body.Data)
		if err == nil {
			return string(data)
		}
	}

	if strings.HasPrefix(part.MimeType, "multipart/") {
		for _, p := range part.Parts {
			if body := ExtractBody(p, mimeType); body != "" {
				return body
			}
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
