package tzcrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
)

//base64Encoder for encoding the encrypted slice of byte to base64 string
func base64Encoder(phrase []byte) string {
	return base64.StdEncoding.EncodeToString(phrase)
}

// base64Decoder for decoding base64 string to the slice of bytes
func base64Decoder(phrase string) (data []byte, valid bool) {
	data, err := base64.StdEncoding.DecodeString(phrase)

	//if it is not a valid base64 string
	if err != nil {
		return data, false
	}

	return data, true
}

// Encrypter is for encrypting the given phrase
func Encrypter(phrase, secreteKey string, iv []byte) string {
	//  create aes cipher
	block, err := aes.NewCipher([]byte(secreteKey))
	if err != nil {
		panic(err)
	}

	// convert the phrase to []bytes
	plainText := []byte(phrase)
	cfb := cipher.NewCFBEncrypter(block, iv)
	cipherText := make([]byte, len(plainText))
	cfb.XORKeyStream(cipherText, plainText)

	// return the base64Encoded string
	return base64Encoder(cipherText)
}

// Decrypter is for extracting the text from the encoded string
func Decrypter(phrase, secreteKey string, iv []byte) (string, bool) {
	// create the aes cipher block
	block, err := aes.NewCipher([]byte(secreteKey))
	if err != nil {
		panic(err)
	}

	// decode the given base64 string to slice of bytes
	cipherText, valid := base64Decoder(phrase)

	if !valid {
		return "", false
	}

	cfb := cipher.NewCFBDecrypter(block, iv)
	plainText := make([]byte, len(cipherText))
	cfb.XORKeyStream(plainText, cipherText)

	// return the decrypted
	return string(plainText), true
}

//RandString is for generating random string
func RandString(len int) (randomString string) {
	randomBytes := make([]byte, 32)
	_, err := rand.Read(randomBytes)
	if err != nil {
		panic(err)
	}

	// return base64 Encodeed string of the length provided by the caller
	return base64.StdEncoding.EncodeToString(randomBytes)[:len]
}

// RandBytes is for generating random []byte    (Useful for creating initiliazing vectors)
func RandBytes(len int) (randomBytes []byte) {
	randBytes := make([]byte, 32)

	_, err := rand.Read(randBytes)
	if err != nil {
		panic(err)
	}

	return randBytes[:len]
}
