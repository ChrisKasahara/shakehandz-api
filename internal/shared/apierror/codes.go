// Package apierror は、APIで利用するエラーコード、HTTPステータス、メッセージを一元管理します。
package apierror

import (
	"net/http"
)

// Code はアプリケーション固有のエラーコードを表す型です。
type Code string

// --- エラーコードの定義 (構造体でグループ化) ---

// commonErrors は共通エラーのコードを保持します。
type commonErrors struct {
	Unknown          Code
	InvalidRequest   Code
	ValidationFailed Code
	JSONParseFailed  Code
	DatabaseError    Code
}

// Common は共通エラーコードのインスタンスです。
var Common = commonErrors{
	Unknown:          "CM00_0001",
	InvalidRequest:   "CM01_0001",
	ValidationFailed: "CM01_0002",
	JSONParseFailed:  "CM01_0003",
	DatabaseError:    "CM01_0004",
}

// authErrors は認証・認可関連エラーのコードを保持します。
type authErrors struct {
	Unauthorized     Code
	TokenExpired     Code
	PermissionDenied Code
}

// Auth は認証・認可関連エラーコードのインスタンスです。
var Auth = authErrors{
	Unauthorized:     "UA01_0001",
	TokenExpired:     "UA01_0002",
	PermissionDenied: "UA02_0001",
}

// redisErrors はRedis関連エラーのコードを保持します。
type redisErrors struct {
	SessionNotFound  Code
	GetDataFailed    Code
	UpdateDataFailed Code
}

// Redis はRedis関連エラーコードのインスタンスです。
var Redis = redisErrors{
	SessionNotFound:  "RD01_0001",
	GetDataFailed:    "RD02_0001",
	UpdateDataFailed: "RD03_0001",
}

// --- エラーコードと情報の紐付け ---

// ErrorInfo は各エラーコードに紐づく情報（HTTPステータスとデフォルトメッセージ）を保持します。
type ErrorInfo struct {
	HTTPStatus int
	Message    string
}

// errorMap はエラーコードとエラー情報の対応表です。
var errorMap = map[Code]ErrorInfo{
	// 共通エラー
	Common.Unknown:          {http.StatusInternalServerError, "不明なエラーが発生しました。"},
	Common.InvalidRequest:   {http.StatusBadRequest, "無効なリクエストです。"},
	Common.ValidationFailed: {http.StatusUnprocessableEntity, "入力内容が正しくありません。"},
	Common.JSONParseFailed:  {http.StatusBadRequest, "JSONの解析に失敗しました。"},
	Common.DatabaseError:    {http.StatusInternalServerError, "データベースエラーが発生しました。"},

	// 認証・認可関連エラー
	Auth.Unauthorized:     {http.StatusUnauthorized, "認証に失敗しました。"},
	Auth.TokenExpired:     {http.StatusUnauthorized, "セッションの有効期限が切れました。"},
	Auth.PermissionDenied: {http.StatusForbidden, "この操作を行う権限がありません。"},

	// Redis関連エラー
	Redis.SessionNotFound:  {http.StatusNotFound, "セッションが見つかりませんでした。"},
	Redis.GetDataFailed:    {http.StatusInternalServerError, "データの取得に失敗しました。"},
	Redis.UpdateDataFailed: {http.StatusInternalServerError, "データの更新に失敗しました。"},
}

// GetInfo はエラーコードに対応するErrorInfoを取得します。
func GetInfo(code Code) ErrorInfo {
	info, ok := errorMap[code]
	if !ok {
		// 未定義の場合は共通のUnknownエラーを返す
		return errorMap[Common.Unknown]
	}
	return info
}
