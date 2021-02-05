package reward

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapKey(t *testing.T) {
	type x struct {
		a, b int
	}
	x1, x2 := x{1, 2}, x{1, 2}
	assert.True(t, x1 == x2)

	m := map[x]int{
		x1: 123,
	}
	m[x2] = 233
	assert.Equal(t, 233, m[x1])
}
