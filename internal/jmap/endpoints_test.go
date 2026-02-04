package jmap

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultSessionURL(t *testing.T) {
	assert.Equal(t, "https://api.fastmail.com/jmap/session", DefaultSessionURL)
	assert.True(t, strings.HasPrefix(DefaultSessionURL, "https://"))
}
