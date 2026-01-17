package utils

import (
	"fmt"
	"testing"
)

func TestTokenGenerate(t *testing.T) {
	for i := 0; i < 30; i++ {
		token := RandomToken(15)

		fmt.Println(token)
	}
}
