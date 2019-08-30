package localStorage

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRandomKeyGen_Generate(t *testing.T) {
	gen := RandomKeyGen{
		Seed: 1,
	}
	firstKey := gen.Generate()

	gen2 := RandomKeyGen{
		Seed: 1,
	}
	secondKey := gen2.Generate()

	// If the Seed field is equal, generated key must be equal.
	assert.Equal(t, firstKey, secondKey)
}
