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
	msgs, err := s.Fetcher.FetchMsg(ctx, token, "has:attachment", 10)
	if err != nil {
		return nil, err
	}

	fmt.Println("Gmail取得を完了。今回の解析件数は", len(msgs), "件です。Negoにプロンプトを送信中")

	// chunkArrayで分割（JSON文字列の配列として）
	chunkedMsgs := chunkArray(msgs, 3)

	chat := s.Gemini.Model.StartChat()

	readyResp, err := chat.SendMessage(ctx, genai.Text(prompts.HRInstruction))

	// Geminiに変換用の指示プロンプトを記憶させる
	readyStr, ok := ExtractText(readyResp)
	if err != nil {
		log.Printf("Gemini API 呼び出し失敗: %v", err)
		c.JSON(500, gin.H{"error": "Gemini API 呼び出し失敗"})
		return nil, fmt.Errorf("Gemini API 呼び出し失敗")
	}

	if !ok || strings.ToLower(strings.TrimSpace(readyStr)) != "ready" {
		c.JSON(500, gin.H{"error": "'ready' が返ってきませんでした"})
		return nil, fmt.Errorf("geminiからの取得メッセージ: %s", readyStr)
	}

	// Geminiから「ready」が返ってきたことを確認
	fmt.Println("Nego「I'm ", readyStr, "!」。Negoは準備完了。続いて変換処理へ移行")

	g, ctx := errgroup.WithContext(ctx)
	var mu sync.Mutex
	sem := semaphore.NewWeighted(10)
	var humanResources []humanresource.HumanResource
	// 最初のチャンクをGeminiに送信
	for _, cmsg := range chunkedMsgs {
		if len(cmsg) == 0 {
			continue
		}

		if err := sem.Acquire(ctx, 1); err != nil {
			return nil, fmt.Errorf("セマフォの取得に失敗: %w", err)
		}

		g.Go(func() error {
			defer sem.Release(1)

			geminiResponse, geminiResErr := chat.SendMessage(ctx, genai.Text(cmsg))
			geminiResponsePart, ok := ExtractText(geminiResponse)
			if geminiResErr != nil {
				log.Printf("Gemini レスポンスデータ不正: %v", geminiResErr)
				return nil
			}
			if !ok {
				log.Printf("Gemini レスポンスデータの文字列変換不正: %v", geminiResponsePart)
				return nil
			}

			trimmedResponse := TrimPrefixAndSuffixGeminiResponse(geminiResponsePart)

			ChunkHumanResources := []humanresource.HumanResource{}

			if err := json.Unmarshal([]byte(trimmedResponse), &ChunkHumanResources); err != nil {
				return nil
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

	fmt.Println("Negoは全ての変換を完了しました。総件数：", len(humanResources), "件です。最後の整形を行なっています")

	// 日付降順
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
