package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"io"
	"os"
	"sync"

	"github.com/spf13/viper"
)

var (
	encryptionKey []byte
	cryptoOnce    sync.Once
)

// getSecretKey retrieves the AES encryption key from the environment lazily.
func getSecretKey() []byte {
	cryptoOnce.Do(func() {
		key := viper.GetString("ENCRYPTION_KEY")
		if key == "" {
			key = os.Getenv("ENCRYPTION_KEY")
		}
		if len(key) != 32 {
			panic("FALHA CRÍTICA DE SEGURANÇA: ENCRYPTION_KEY não definida ou inválida. Deve ter exatamente 32 bytes.")
		}
		encryptionKey = []byte(key)
	})
	return encryptionKey
}

func HashDeterministic(text string) (string, error) {
	if text == "" {
		return "", nil
	}
	
	// Utilizando HMAC com a mesma chave secreta para agir como Pepper
	// Isso garante que o hash seja determinístico e protegido.
	h := hmac.New(sha256.New, getSecretKey())
	h.Write([]byte(text))
	return hex.EncodeToString(h.Sum(nil)), nil
}

func EncryptAES(text string) (string, error) {
	if text == "" {
		return "", nil
	}

	block, err := aes.NewCipher(getSecretKey())
	if err != nil {
		return "", err
	}

	plaintext := []byte(text)
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	return base64.URLEncoding.EncodeToString(ciphertext), nil
}

func DecryptAES(cryptoText string) (string, error) {
	if cryptoText == "" {
		return "", nil
	}

	ciphertext, err := base64.URLEncoding.DecodeString(cryptoText)
	if err != nil {
		// Se não conseguir decodificar base64, assumimos que não estava criptografado (dados antigos)
		return cryptoText, nil
	}

	block, err := aes.NewCipher(getSecretKey())
	if err != nil {
		return "", err
	}

	if len(ciphertext) < aes.BlockSize {
		return cryptoText, nil // Dados corrompidos ou não criptografados
	}

	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)

	return string(ciphertext), nil
}

func EncryptAESGCM(text string) (string, error) {
	if text == "" {
		return "", nil
	}

	block, err := aes.NewCipher(getSecretKey())
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(text), nil)
	return base64.URLEncoding.EncodeToString(ciphertext), nil
}

func DecryptAESGCM(cryptoText string) (string, error) {
	if cryptoText == "" {
		return "", nil
	}

	ciphertext, err := base64.URLEncoding.DecodeString(cryptoText)
	if err != nil {
		return cryptoText, nil
	}

	block, err := aes.NewCipher(getSecretKey())
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return cryptoText, nil
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return cryptoText, nil
	}

	return string(plaintext), nil
}
