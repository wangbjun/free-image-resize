package common

import (
	"github.com/stretchr/testify/assert"
	_ "image/gif"
	_ "image/png"
	"testing"
)

func TestResize_Resize(t *testing.T) {
	t.Run("test new", func(t *testing.T) {
		resize := New()
		assert.NotNil(t, resize)

		_, err := resize.Resize("test.jpg")
		assert.Nil(t, err)
	})
}
