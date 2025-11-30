package types

import (
	"crypto/rand"
	"math/big"
)

// Alphabet matches the TypeScript nanoid alphabet
const alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// GenerateID generates a 12-character nanoid-like string
// Matches TypeScript: nid = customAlphabet(alphabet, maxSize)
func GenerateID() string {
	const length = 12
	const alphabetLen = 62 // len(alphabet)

	result := make([]byte, length)
	for i := range result {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(alphabetLen)))
		result[i] = alphabet[n.Int64()]
	}

	return string(result)
}
