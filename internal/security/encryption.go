package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
)

type DataEncryptor struct {
	block cipher.Block
}

func NewDataEncryptor(key []byte) (*DataEncryptor, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return &DataEncryptor{block: block}, nil
}

func (e *DataEncryptor) Encrypt(data []byte) ([]byte, error) {
	// Generate a random IV
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	// Create a CBC mode encrypter
	stream := cipher.NewCFBDecrypter(e.block, iv)

	// XOR the data
	encrypted := make([]byte, len(data))
	stream.XORKeyStream(encrypted, data)

	// Prepend IV to encrypted data
	return append(iv, encrypted...), nil
}

func (e *DataEncryptor) Decrypt(data []byte) ([]byte, error) {
	if len(data) < aes.BlockSize {
		return nil, fmt.Errorf("data too short to contain IV")
	}

	iv := data[:aes.BlockSize]
	encrypted := data[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(e.block, iv)
	decrypted := make([]byte, len(encrypted))
	stream.XORKeyStream(decrypted, encrypted)

	return decrypted, nil
}
