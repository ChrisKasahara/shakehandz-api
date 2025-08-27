package gmail

import (
	"encoding/base64"
	"errors"
	msg "shakehandz-api/internal/shared/message"
	"strings"

	"google.golang.org/api/gmail/v1"
)

// メッセージ詳細→DTO
func ParseMessage(gmsg *gmail.Message) (*msg.Message, error) {
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
func ExtractAttachments(payload *gmail.MessagePart) []msg.Attachment {
	var atts []msg.Attachment
	if payload == nil {
		return atts
	}
	if payload.Filename != "" && payload.Body != nil && payload.Body.AttachmentId != "" {
		atts = append(atts, msg.Attachment{
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
