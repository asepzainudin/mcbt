package passwordutil

import (
	"crypto/rand"
	"math/big"
)

const charset = "abcdefghjkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789"

func Generate(length int) (string, error) {
	if length < 8 {
		length = 8
	}
	out := make([]byte, length)
	max := big.NewInt(int64(len(charset)))
	for i := range out {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		out[i] = charset[n.Int64()]
	}
	return string(out), nil
}

const DefaultPassword = "McBT@1234"
