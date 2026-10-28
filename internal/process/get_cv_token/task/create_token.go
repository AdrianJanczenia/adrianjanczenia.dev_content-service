package task

import (
	"crypto/rand"
	"math/big"
)

type CreateTokenTask struct{}

func NewCreateTokenTask() *CreateTokenTask {
	return &CreateTokenTask{}
}

func (t *CreateTokenTask) Execute() (string, error) {
	const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"

	result := make([]byte, 32)

	for i := range result {
		index, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
		if err != nil {
			return "", err
		}

		result[i] = alphabet[index.Int64()]
	}

	return string(result), nil
}
