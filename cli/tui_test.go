package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTUICommand(t *testing.T) {
	cmd := newTUICommand()

	assert.Equal(t, "tui", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
	assert.NotEmpty(t, cmd.Long)
	assert.NotNil(t, cmd.RunE)
}
