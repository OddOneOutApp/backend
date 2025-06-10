package random

import (
	"math/rand"
	"time"
)

const charset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ" + "0123456789"

var seededRand *rand.Rand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

func RandomStringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func RandomString(length int) string {
	return RandomStringWithCharset(length, charset)
}

func RandomSelect[T any](input []T, count int) []T {
	localRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	n := len(input)
	if count >= n {
		return input
	}
	indices := localRand.Perm(n)[:count]
	result := make([]T, count)
	for i, idx := range indices {
		result[i] = input[idx]
	}
	return result
}
