# システム構成ドキュメント

## ディレクトリ・ファイル構成と主な責務

- `cmd/server/main.go`

  - サーバのエントリーポイント。Gin, Redis, ルーティング初期化。

- `internal/auth/`

  - `handler.go` ... 認証系 API ハンドラ
  - `idtoken_user.go` ... ID トークンからユーザー情報取得
  - `middleware.go` ... Gin 用認証ミドルウェア
  - `model.go` ... ユーザーモデル定義
  - `oauth_repo.go` ... Google リフレッシュトークン取得
  - `upsert_user_with_token.go` ... トークン付きユーザー登録・更新

- `internal/gemini/`

  - `client.go` ... Gemini API クライアント
  - `handler.go` ... Gemini 連携 API ハンドラ
  - `message.go` ... 未処理メッセージ取得・処理
  - `service.go` ... Gemini 連携サービス本体
  - `slice.go` ... メッセージ配列分割ユーティリティ
  - `status_handler.go` ... Gemini ステータス API ハンドラ
  - `text.go` ... Gemini レスポンスからテキスト抽出

- `internal/humanresource/`

  - `handler.go` ... 要員情報 API ハンドラ
  - `model.go` ... 要員情報モデル定義

- `internal/project/`

  - `handler.go` ... プロジェクト情報 API ハンドラ
  - `model.go` ... プロジェクト情報モデル定義

- `internal/shared/`

  - `db.go` ... GORM による DB 初期化・マイグレーション
  - `auth/verified.go` ... ユーザー認証済み判定
  - `cache/client.go` ... Redis クライアント生成
  - `cache/ai/model.go` ... AI キャッシュモデル
  - `cache/ai/redis.go` ... AI キャッシュ用 Redis 操作
  - `crypto/crypto.go` ... 暗号化・復号化ユーティリティ
  - `google/oauth.go` ... Google OAuth ユーティリティ
  - `message/MessageIF.go` ... メッセージ IF 定義
  - `message/gmail/client.go` ... Gmail API クライアント
  - `message/gmail/fetcher.go` ... Gmail メッセージ取得
  - `message/gmail/fetcher_msg_detail.go` ... メッセージ詳細取得
  - `message/gmail/fetcher_msg_ids.go` ... メッセージ ID 取得
  - `message/gmail/GmailMessageIF.go` ... Gmail メッセージ IF
  - `message/gmail/msg_heper.go` ... メッセージ解析・添付抽出

- `internal/router/router.go`

  - ルーティング設定

- `prompts/`

  - プロンプト・指示文格納

- `utils/`
  - 汎用ユーティリティ

## 主要モデル

### HumanResource

- ID, MessageID（ユニーク）
- 添付ファイル情報
- 基本情報（氏名イニシャル・年齢・国籍）
- 経験領域・役割（JSON 配列）
- スキル（JSON 配列）
- 雇用体系・勤務スタイル
- 直請けフラグ

### Project

- ID（主キー）
- メール情報（ID, 件名, 送信者, 受信日時）
- プロジェクト開始月
- 勤務地・都道府県
- リモート頻度・勤務時間
- 必須スキル・単価
- 業務内容・制約・優先人材
- 抽出信頼度・備考

## DB 構成・マイグレーション

- `internal/shared/db.go` で GORM による DB 初期化・マイグレーション
- `Project`, `HumanResource` モデルを自動マイグレーション

## ER 図

ER 図は [`ai-generated/docs/ER/er_diagram.md`](ai-generated/docs/ER/er_diagram.md) を参照
