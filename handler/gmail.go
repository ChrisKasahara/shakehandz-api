package handler

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/sync/errgroup"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

func GmailMessagesHandler(c *gin.Context) {
	ctx := context.Background()

	// Authorization ヘッダーから access_token を取得
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing"})
		return
	}

	var accessToken string
	_, err := fmt.Sscanf(authHeader, "Bearer %s", &accessToken)
	if err != nil || accessToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Authorization header"})
		return
	}

	token := &oauth2.Token{AccessToken: accessToken}
	tokenSource := oauth2.StaticTokenSource(token)

	srv, err := gmail.NewService(ctx, option.WithTokenSource(tokenSource))
	if err != nil {
		log.Printf("Error creating Gmail service: %v", err) // ログ追加
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to create Gmail service"})
		return
	}

	// メッセージID一覧取得 (has:attachment クエリはそのまま維持)
	msgsList, err := srv.Users.Messages.List("me").MaxResults(5).Q("has:attachment").Do()
	if err != nil {
		log.Printf("Error listing messages: %v", err) // ログ追加
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to list messages"})
		return
	}

	g := new(errgroup.Group)
	var mu sync.Mutex // detailedMessages を保護するためのMutex

	var detailedMessages []map[string]interface{} // 取得した詳細なメッセージ情報を格納するスライス

	// 各メッセージの詳細を並行して取得
	for _, msg := range msgsList.Messages {
		msg := msg // ループ変数のキャプチャ
		g.Go(func() error {
			// "full" フォーマットでメッセージの詳細を取得
			fullMsg, err := srv.Users.Messages.Get("me", msg.Id).Format("full").Do()
			if err != nil {
				// 個々のメッセージ取得エラーは致命的ではない場合が多いため、ログに記録し、nilを返す
				log.Printf("Failed to get full message for ID %s: %v", msg.Id, err)
				return nil
			}

			// 抽出する情報の初期化
			var subject, from, date, to, cc, replyTo string
			var plainTextBody, htmlBody string
			attachments := []map[string]interface{}{} // 添付ファイルの情報を格納

			// --- ヘッダー情報の抽出 ---
			if fullMsg.Payload != nil && fullMsg.Payload.Headers != nil {
				for _, h := range fullMsg.Payload.Headers {
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
						// 他に必要なヘッダーがあればここに追加
					}
				}
			}

			// --- 本文と添付ファイルの抽出（再帰的なヘルパー関数） ---
			var extractParts func(parts []*gmail.MessagePart)
			extractParts = func(parts []*gmail.MessagePart) {
				for _, part := range parts {
					// 本文の取得 (text/plain または text/html)
					if part.Body != nil && part.Body.Data != "" {
						decodedData, decodeErr := base64.URLEncoding.DecodeString(part.Body.Data)
						if decodeErr != nil {
							log.Printf("Failed to decode part data for message %s: %v", msg.Id, decodeErr)
						} else {
							if part.MimeType == "text/plain" {
								plainTextBody = string(decodedData)
							} else if part.MimeType == "text/html" {
								htmlBody = string(decodedData)
							}
						}
					}

					// 添付ファイルの取得
					// Filenameがあり、BodyにAttachmentIdがある場合は添付ファイル
					if part.Filename != "" && part.Body != nil && part.Body.AttachmentId != "" {
						attachments = append(attachments, map[string]interface{}{
							"filename":     part.Filename,
							"mimeType":     part.MimeType,
							"size":         part.Body.Size,         // 添付ファイルのサイズ
							"attachmentId": part.Body.AttachmentId, // 添付ファイルの内容を別途取得する際に使用
						})
					}

					// ネストされたパートがある場合、再帰的に処理
					if part.Parts != nil && len(part.Parts) > 0 {
						extractParts(part.Parts)
					}
				}
			}

			// メッセージのペイロードを処理
			if fullMsg.Payload != nil {
				if fullMsg.Payload.Parts != nil {
					extractParts(fullMsg.Payload.Parts)
				} else if fullMsg.Payload.Body != nil && fullMsg.Payload.Body.Data != "" {
					// Partsがない場合のシンプルな本文（単一パートのメッセージ）
					decodedData, decodeErr := base64.URLEncoding.DecodeString(fullMsg.Payload.Body.Data)
					if decodeErr != nil {
						log.Printf("Failed to decode simple body data for message %s: %v", msg.Id, decodeErr)
					} else {
						if fullMsg.Payload.MimeType == "text/plain" {
							plainTextBody = string(decodedData)
						} else if fullMsg.Payload.MimeType == "text/html" {
							htmlBody = string(decodedData)
						}
					}
				}
			}

			// --- 取得した全ての情報をマップにまとめる ---
			mu.Lock()
			detailedMessages = append(detailedMessages, map[string]interface{}{
				"id":            fullMsg.Id,
				"threadId":      fullMsg.ThreadId,     // スレッドID
				"snippet":       fullMsg.Snippet,      // スニペット（短いプレビューテキスト）
				"labelIds":      fullMsg.LabelIds,     // 適用されているラベルのID
				"historyId":     fullMsg.HistoryId,    // 最新の変更履歴ID
				"sizeEstimate":  fullMsg.SizeEstimate, // 推定サイズ（バイト）
				"internalDate":  fullMsg.InternalDate, // Gmail内部での受信日時（Unixエポックミリ秒）
				"subject":       subject,              // 件名
				"from":          from,                 // 送信元
				"date":          date,                 // 送信日時
				"to":            to,                   // 宛先
				"cc":            cc,                   // CC
				"replyTo":       replyTo,              // 返信先
				"plainTextBody": plainTextBody,        // プレーンテキスト本文
				"htmlBody":      htmlBody,             // HTML本文
				"attachments":   attachments,          // 添付ファイル情報リスト
			})
			mu.Unlock()
			return nil
		})
	}

	// 全メッセージの取得が終わるのを待つ
	if err := g.Wait(); err != nil {
		log.Printf("Error during fetching message details: %v", err) // ログ追加
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching message details"})
		return
	}

	// 日付で新しい順にソート
	sort.Slice(detailedMessages, func(i, j int) bool {
		layouts := []string{
			time.RFC1123Z,                          // "Mon, 02 Jan 2006 15:04:05 -0700"
			time.RFC1123,                           // "Mon, 02 Jan 2006 15:04:05 MST"
			time.RFC822Z,                           // "02 Jan 06 15:04 -0700"
			time.RFC822,                            // "02 Jan 06 15:04 MST"
			time.RFC3339,                           // "2006-01-02T15:04:05Z07:00"
			"Mon, 2 Jan 2006 15:04:05 -0700 (MST)", // 特定のGMTオフセット形式に対応
			"Mon, 2 Jan 2006 15:04:05 MST",         // タイムゾーン名のみの場合
		}

		parse := func(s string) time.Time {
			for _, layout := range layouts {
				if t, err := time.Parse(layout, s); err == nil {
					return t
				}
			}
			// 全てのレイアウトでパースできなかった場合
			log.Printf("Could not parse date string: %s", s)
			return time.Time{} // 無効なTimeオブジェクトを返す
		}

		// 型アサーションが失敗する可能性を考慮
		dateStrI, okI := detailedMessages[i]["date"].(string)
		dateStrJ, okJ := detailedMessages[j]["date"].(string)

		if !okI || !okJ {
			// 日付が取得できなかった場合は順番を維持しないか、エラーをログに記録
			return false
		}

		ti := parse(dateStrI)
		tj := parse(dateStrJ)

		// 無効な時間オブジェクトが返された場合も考慮
		if ti.IsZero() || tj.IsZero() {
			return false
		}

		return tj.Before(ti) // 新しい順 (降順)
	})

	c.JSON(http.StatusOK, detailedMessages)
}
