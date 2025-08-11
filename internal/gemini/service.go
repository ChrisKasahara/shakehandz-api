package gemini

import (
	"encoding/json"
	"fmt"
	"log"
	"shakehandz-api/internal/humanresource"
	msg "shakehandz-api/internal/shared/message"
	"shakehandz-api/prompts"
	"sort"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/google/generative-ai-go/genai"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
	"gorm.io/gorm"
)

// 　Gmailメッセージ取得→Gemini解析→（将来）DB保存
type Service struct {
	Fetcher msg.MessageFetcher
	Gemini  *Client
	DB      *gorm.DB
}

func NewService(f msg.MessageFetcher, g *Client, db *gorm.DB) *Service {
	return &Service{Fetcher: f, Gemini: g, DB: db}
}

func (s *Service) Run(c *gin.Context, token string) ([]humanresource.HumanResource, error) {
	ctx := c.Request.Context()
	fmt.Println("NegoはGmailを取得中")

	// DB既存のメッセージIDを除外した未処理メッセージを最大N件取得
	msgs, err := s.fetchUnprocessedMessages(c, token, 20, s.DB)
	if err != nil {
		return nil, err
	}

	// 未処理メッセージがない場合は空の結果を返す
	if len(msgs) == 0 {
		fmt.Println("最新のメールはすべて処理済みです")
		return []humanresource.HumanResource{}, nil
	}

	fmt.Println("Gmail取得を完了。今回の解析件数は", len(msgs), "件です。Negoにプロンプトを送信中")

	// chunkArrayで分割（JSON文字列の配列として）
	chunkedMsgs := chunkArray(msgs, 5)

	// 共有モデルの浅いコピーを作成してSystemInstructionを一度だけ設定
	localModel := *s.Gemini.Model
	localModel.SystemInstruction = &genai.Content{
		Role:  "system",
		Parts: []genai.Part{genai.Text(prompts.HRInstruction)},
	}
	fmt.Println("Negoは準備完了。続いて変換処理へ移行")

	g, ctx := errgroup.WithContext(ctx)
	var mu sync.Mutex
	sem := semaphore.NewWeighted(5)
	var humanResources []humanresource.HumanResource

	// SystemInstruction設定済みのローカルモデルを各ゴルーチンで使用
	for _, cmsg := range chunkedMsgs {
		chunk := cmsg // range変数のクロージャ捕捉対策
		if len(chunk) == 0 {
			continue
		}

		if err := sem.Acquire(ctx, 1); err != nil {
			return nil, fmt.Errorf("セマフォの取得に失敗: %w", err)
		}

		g.Go(func() error {
			defer sem.Release(1)

			// 事前設定されたローカルモデルでGenerateContentを呼び出し
			geminiResponse, geminiResErr := localModel.GenerateContent(ctx, genai.Text(chunk))
			if geminiResErr != nil {
				log.Printf("Gemini API 呼び出し失敗: %v", geminiResErr)
				return fmt.Errorf("Gemini API 呼び出し失敗: %w", geminiResErr)
			}

			if geminiResponse == nil {
				log.Printf("Gemini レスポンスが nil です")
				return fmt.Errorf("Gemini レスポンスが nil です")
			}

			geminiResponsePart, ok := ExtractText(geminiResponse)
			if !ok {
				log.Printf("Gemini レスポンスデータの文字列変換不正: %v", geminiResponsePart)
				return fmt.Errorf("Gemini レスポンスデータの文字列変換不正: %s", geminiResponsePart)
			}

			trimmedResponse := TrimPrefixAndSuffixGeminiResponse(geminiResponsePart)

			ChunkHumanResources := []humanresource.HumanResource{}

			if err := json.Unmarshal([]byte(trimmedResponse), &ChunkHumanResources); err != nil {
				log.Printf("JSON Unmarshal失敗: %v", err)
				return fmt.Errorf("JSON Unmarshal失敗: %w", err)
			}

			for _, hr := range ChunkHumanResources {
				mu.Lock()
				humanResources = append(humanResources, hr)
				mu.Unlock()
			}

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		log.Printf("fetcher: detail fetch error: %v", err)
		return nil, err
	}

	// 念の為、MessageIDの重複を除外
	seen := make(map[string]struct{}, len(humanResources))
	uniq := make([]humanresource.HumanResource, 0, len(humanResources))

	for _, hr := range humanResources {
		mid := strings.TrimSpace(hr.MessageID)
		if mid == "" {
			uniq = append(uniq, hr)
			continue
		}
		if _, ok := seen[mid]; ok {
			continue
		}
		seen[mid] = struct{}{}
		uniq = append(uniq, hr)
	}

	humanResources = uniq

	fmt.Println("Negoは全ての変換を完了しました。総件数：", len(humanResources), "件です。最後の整形を行なっています")

	// 日付で降順
	sort.Slice(humanResources, func(i, j int) bool {
		return humanResources[i].CreatedAt.Unix() > humanResources[j].CreatedAt.Unix()
	})

	fmt.Println("Negoは作業を保存中")
	// DB保存処理
	if len(humanResources) > 0 {
		if err := s.DB.Create(&humanResources).Error; err != nil {
			return nil, fmt.Errorf("DB保存失敗: %w", err)
		}
	}
	return humanResources, nil
}
