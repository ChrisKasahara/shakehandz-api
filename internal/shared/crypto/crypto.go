package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

func getKey() ([]byte, error) {
	kb64 := strings.TrimSpace(os.Getenv("GOOGLE_TOKEN_ENC_KEY_BASE64"))
	if kb64 == "" {
		return nil, errors.New("GOOGLE_TOKEN_ENC_KEY_BASE64 not set")
	}
	// 改行・空白を除去（コピペ対策）
	kb64 = strings.ReplaceAll(kb64, "\n", "")
	kb64 = strings.ReplaceAll(kb64, "\r", "")
	kb64 = strings.ReplaceAll(kb64, " ", "")

	// まず標準Base64
	key, err := base64.StdEncoding.DecodeString(kb64)
	if err != nil {
		// URL-safe やパディング無しの可能性に対応
		if k2, err2 := base64.RawStdEncoding.DecodeString(kb64); err2 == nil {
			key = k2
		} else if k3, err3 := base64.RawURLEncoding.DecodeString(kb64); err3 == nil {
			key = k3
		} else {
			return nil, fmt.Errorf("decode base64 failed: %w", err)
		}
	}
	// 長さ検証（AES: 16/24/32）
	if n := len(key); n != 16 && n != 24 && n != 32 {
		return nil, fmt.Errorf("invalid key length: %d (must be 16/24/32 bytes)", n)
	}
	return key, nil
}

func EncryptToBytes(plain string) ([]byte, error) {
	key, err := getKey()
	if err != nil {
		return nil, fmt.Errorf("getKey: %w", err)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("newCipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("newGCM: %w", err)
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("nonce: %w", err)
	}
	out := gcm.Seal(nonce, nonce, []byte(plain), nil)
	return out, nil
}

func DecryptFromBytes(raw []byte) (string, error) {
	key, err := getKey()
	if err != nil {
		return "", fmt.Errorf("getKey: %w", err)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("newCipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("newGCM: %w", err)
	}
	if len(raw) < gcm.NonceSize() {
		return "", errors.New("ciphertext too short")
	}
	nonce, ct := raw[:gcm.NonceSize()], raw[gcm.NonceSize():]
	pt, err := gcm.Open(nil, nonce, ct, nil)
	if err != nil {
		return "", fmt.Errorf("open: %w", err)
	}
	return string(pt), nil
}
