package task

import (
	"crypto/sha256"
	"crypto/subtle"
)

type ComparePasswordTask struct {
	expected [sha256.Size]byte
}

func NewComparePasswordTask(correctPassword string) *ComparePasswordTask {
	return &ComparePasswordTask{expected: sha256.Sum256([]byte(correctPassword))}
}

func (t *ComparePasswordTask) Execute(password string) bool {
	provided := sha256.Sum256([]byte(password))

	return subtle.ConstantTimeCompare(provided[:], t.expected[:]) == 1
}
