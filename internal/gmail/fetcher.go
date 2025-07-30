package gmail

import (
	"encoding/base64"
	"strings"

	"google.golang.org/api/gmail/v1"
)

func ExtractPlainText(payload *gmail.MessagePart) string {
	if payload == nil {
		return ""
	}

	if payload.MimeType == "text/plain" && payload.Body != nil && payload.Body.Data != "" {
		return DecodeBase64URL(payload.Body.Data)
	}

	// マルチパートの場合はパーツを再帰的に探索
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
