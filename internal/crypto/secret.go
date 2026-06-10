// Package crypto provides at-rest encryption for sensitive settings (SMTP password,
// Jira token, webhook URLs). It is OPTIONAL and backward-compatible:
//
//   - With APP_ENCRYPTION_KEY set (32-byte key, base64 or hex), Encrypt() returns an
//     "enc:" prefixed AES-256-GCM ciphertext.
//   - Without a key, Encrypt() is a no-op (returns plaintext) so local/dev installs
//     keep working unchanged.
//   - Decrypt() handles both: it decrypts "enc:" values and passes plaintext through,
//     so an existing DB with plaintext secrets keeps working after a key is added.
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io"
	"log"
	"os"
	"strings"
)

const prefix = "enc:"

var key []byte // nil = encryption disabled

// Init loads the key from APP_ENCRYPTION_KEY (base64 or hex, decoding to 32 bytes).
// Call once at startup. A missing key disables encryption (with a warning).
func Init() {
	raw := strings.TrimSpace(os.Getenv("APP_ENCRYPTION_KEY"))
	if raw == "" {
		log.Println("🔓 APP_ENCRYPTION_KEY not set — secrets stored in plaintext (set it to encrypt at rest)")
		return
	}
	k, err := decodeKey(raw)
	if err != nil {
		log.Printf("🔓 APP_ENCRYPTION_KEY invalid (%v) — secrets stored in plaintext", err)
		return
	}
	key = k
	log.Println("🔒 Secret encryption enabled (AES-256-GCM)")
}

func decodeKey(raw string) ([]byte, error) {
	if b, err := base64.StdEncoding.DecodeString(raw); err == nil && len(b) == 32 {
		return b, nil
	}
	if b, err := hex.DecodeString(raw); err == nil && len(b) == 32 {
		return b, nil
	}
	return nil, errors.New("key must decode (base64 or hex) to exactly 32 bytes")
}

// Enabled reports whether encryption is active.
func Enabled() bool { return key != nil }

// Encrypt returns an "enc:"-prefixed ciphertext, or the plaintext unchanged when
// encryption is disabled or the input is empty/already encrypted.
func Encrypt(plaintext string) string {
	if key == nil || plaintext == "" || strings.HasPrefix(plaintext, prefix) {
		return plaintext
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return plaintext
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return plaintext
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return plaintext
	}
	ct := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return prefix + base64.StdEncoding.EncodeToString(ct)
}

// Decrypt reverses Encrypt. Non-prefixed values (legacy plaintext) are returned as-is.
func Decrypt(value string) string {
	if !strings.HasPrefix(value, prefix) {
		return value
	}
	if key == nil {
		return value // can't decrypt without the key; leave as-is
	}
	data, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(value, prefix))
	if err != nil {
		return value
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return value
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil || len(data) < gcm.NonceSize() {
		return value
	}
	nonce, ct := data[:gcm.NonceSize()], data[gcm.NonceSize():]
	pt, err := gcm.Open(nil, nonce, ct, nil)
	if err != nil {
		return value
	}
	return string(pt)
}
