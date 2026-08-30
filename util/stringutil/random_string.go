package stringutil

import (
	"crypto/rand"
	"math/big"
)

func RandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	charsetLen := big.NewInt(int64(len(charset)))
	result := make([]byte, length)
	for i := range result {
		n, err := rand.Int(rand.Reader, charsetLen)
		if err != nil {
			panic("failed to generate random string: " + err.Error())
		}
		result[i] = charset[n.Int64()]
	}
	return string(result)
}
