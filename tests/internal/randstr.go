package internal

import (
	"crypto/rand"
	"encoding/base64"
)

var encoder = base64.StdEncoding.WithPadding(base64.NoPadding)

func RandomString(n int) string {
	randomBytes := make([]byte, n)
	_, err := rand.Read(randomBytes)
	if err != nil {
		panic(err)
	}
	dst := make([]byte, base64.StdEncoding.EncodedLen(len(randomBytes)))

	encoder.Encode(dst, randomBytes)

	// str := string(dst[:n]) + "%v %v"
	// return str
	return string(dst[:n])
}

func RandomStrings(n int) []string {
	s := make([]string, 1000)

	for i := 0; i < len(s); i++ {
		s[i] = RandomString(n)
	}

	return s
}

func RandomStringWithFormatting(n int) string {
	s := RandomString(n)

	s += " %v"
	return s
}

func RandomStringsWithFormatting(n int) []string {
	s := make([]string, 1000)

	for i := 0; i < len(s); i++ {
		s[i] = RandomStringWithFormatting(n)
	}

	return s
}
