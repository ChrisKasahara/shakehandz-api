# shakehandz-api

## 起動方法

```sh
go run ./cmd/server
```

## ディレクトリ構成（抜粋）

- cmd/server/main.go ... エントリポイント
- internal/gmail/ ... Gmail 関連
- internal/gemini/ ... Gemini クライアント
- internal/humanresource/ ... 人事ドメイン
- internal/project/ ... プロジェクトドメイン
- internal/auth/ ... 認証
- internal/shared/ ... DB・ロガー等
