package extractor

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"shakehandz-api/internal/auth"
	"shakehandz-api/internal/humanresource"
	"shakehandz-api/internal/shared/llm/gemini"
	"shakehandz-api/internal/shared/options"
	"shakehandz-api/prompts"
	"strings"
	"sync"

	"github.com/google/generative-ai-go/genai"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
	"google.golang.org/api/gmail/v1"
)

func Extract(ctx context.Context, user auth.User, client *gemini.Client, gmail_svc *gmail.Service, s *Service, currentBatch ExtractorBatchExecution) (bool, error) {
	fmt.Println("kmoaiはGmailを取得中")

	// DB既存のメッセージIDを除外した未処理メッセージを最大N件取得
	msgs, err := s.fetchUnprocessedMessages(ctx, user, gmail_svc, MaxMessages)
	if err != nil {
		fmt.Println("fetchUnprocessedMessages error:", err)
		return false, err
	}

	// 未処理メッセージがない場合は空の結果を返す
	if len(msgs) == 0 {
		fmt.Println("最新のメールはすべて処理済みです")
		return false, nil
	}

	fmt.Println("Gmail取得を完了。今回の解析件数は", len(msgs), "件です。kmoaiにプロンプトを送信中")

	// chunkArrayで分割（JSON文字列の配列として）
	chunkedMsgs := chunkArray(msgs, GeminiChunkSize)

	// 共有モデルの浅いコピーを作成してSystemInstructionを一度だけ設定
	localModel := client.Model
	localModel.SystemInstruction = &genai.Content{
		Role:  "system",
		Parts: []genai.Part{genai.Text(prompts.HRInstruction)},
	}
	fmt.Println("kmoaiは準備完了。続いて変換処理へ移行")

	g, ctx := errgroup.WithContext(ctx)
	var mu sync.Mutex
	sem := semaphore.NewWeighted(MaxGoroutine)

	// 最終結果を格納するスライス
	var humanResources []humanresource.HumanResource

	// SystemInstruction設定済みのローカルモデルを各ゴルーチンで使用
	for _, cmsg := range chunkedMsgs {
		chunk := cmsg
		if len(chunk) == 0 {
			continue
		}

		if err := sem.Acquire(ctx, 1); err != nil {
			log.Printf("セマフォの取得に失敗: %v", err)
			return false, fmt.Errorf("セマフォの取得に失敗: %w", err)
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

			// Geminiのレスポンスから文字列を抽出
			geminiResponsePart, ok := gemini.ExtractText(geminiResponse)
			if !ok {
				log.Printf("Gemini レスポンスデータの文字列変換不正: %v", geminiResponsePart)
				return fmt.Errorf("Gemini レスポンスデータの文字列変換不正: %s", geminiResponsePart)
			}

			// Geminiのレスポンスから前後の不要な文字列をトリム
			trimmedResponse := gemini.TrimPrefixAndSuffixGeminiResponse(geminiResponsePart)

			ChunkHumanResources := []humanresource.HumanResource{}

			if err := json.Unmarshal([]byte(trimmedResponse), &ChunkHumanResources); err != nil {
				return fmt.Errorf("JSON Unmarshal失敗: %w", err)
			}

			// 念の為、MessageIDの重複を除外
			seen := make(map[string]struct{}, len(ChunkHumanResources))
			uniq := make([]humanresource.HumanResource, 0, len(ChunkHumanResources))

			for _, hr := range ChunkHumanResources {
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

			ChunkHumanResources = uniq

			// DB保存
			fmt.Println("変換完了。kmoaiは", len(ChunkHumanResources), "件の変換を保存中")
			saved, err := SaveExtractedHumanResources(ChunkHumanResources, user, s)

			if err != nil {
				return fmt.Errorf("データベースの保存に失敗 次の項目へ進む: %w", err)
			}

			if saved {
				fmt.Printf("kmoaiは%d件の変換を保存しました\n", len(ChunkHumanResources))

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
		log.Printf("ERROR: 並列処理中にエラー発生: %v", err)
		return false, err
	}

	fmt.Println("kmoaiは全ての変換を完了しました。総件数：", len(humanResources), "件です。最後の整形を行なっています")

	// 登録されたすべてのスキルをまとめる
	var allSkills []string
	for _, hr := range humanResources {
		for _, skill := range hr.MainSkills {
			allSkills = append(allSkills, skill)
		}
		for _, skill := range hr.SubSkills {
			allSkills = append(allSkills, skill)
		}
	}

	fmt.Println("kmoaiは作業を保存中")

	// 登場スキルを保存
	if len(allSkills) > 0 {
		if err := options.SaveSkills(s.DB, allSkills); err != nil {
			log.Printf("ERROR: Failed to save skills: %v", err)
			return false, err
		}
	}
	return true, nil
}
