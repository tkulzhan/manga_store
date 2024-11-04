package helpers

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
)

var bytes = []byte{35, 46, 57, 24, 85, 35, 24, 74, 87, 35, 88, 98, 66, 32, 14, 5}

func Encode(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}

func Decode(s string) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func ensureKeyLength(secret string) []byte {
	if len(secret) == 24 {
		return []byte(secret)
	}
	if len(secret) < 24 {
		return []byte(secret + string(make([]byte, 24-len(secret))))
	}
	return []byte(secret[:24])
}

func Encrypt(text string) string {
	secret := GetEnv("SECRET", "secret_key")
	key := ensureKeyLength(secret)

	block, _ := aes.NewCipher(key)

	plainText := []byte(text)
	cfb := cipher.NewCFBEncrypter(block, bytes)
	cipherText := make([]byte, len(plainText))
	cfb.XORKeyStream(cipherText, plainText)
	return Encode(cipherText)
}

func Decrypt(encodedText string) string {
	secret := GetEnv("SECRET", "secret_key")
	key := ensureKeyLength(secret)

	block, _ := aes.NewCipher(key)
	cipherText, _ := Decode(encodedText)

	cfb := cipher.NewCFBDecrypter(block, bytes)
	plainText := make([]byte, len(cipherText))
	cfb.XORKeyStream(plainText, cipherText)
	return string(plainText)
}
