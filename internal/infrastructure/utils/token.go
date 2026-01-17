package utils

import (
	"crypto/sha1"
	"fmt"
	"math/rand"
	"time"
)

var characterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

var rd *rand.Rand

func init() {
	rd = rand.New(rand.NewSource(time.Now().UnixNano()))
}

// NewSHA1Hash generates a new SHA1 hash based on
// a random number of characters.
func GenerateRefreshToken(prefix string, n ...int) string {
	noRandomCharacters := 32

	if len(n) > 0 {
		noRandomCharacters = n[0]
	}

	randString := RandomString(noRandomCharacters)

	hash := sha1.New()
	hash.Write([]byte(prefix + randString))
	bs := hash.Sum(nil)

	return fmt.Sprintf("%x", bs)
}

// RandomString generates a random string of n length
func RandomString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = characterRunes[rd.Intn(len(characterRunes))]
	}
	return string(b)
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandomToken(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func GenerateVerificationCode() string {
	code := rand.Intn(900000) + 100000
	return fmt.Sprintf("%06d", code)
}

// SHA1 加密
func Sha1(data string) string {
	h := sha1.New()
	h.Write([]byte(data))
	return fmt.Sprintf("%x", h.Sum(nil))
}
