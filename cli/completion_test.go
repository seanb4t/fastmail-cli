package cli

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCompletionCommand(t *testing.T) {
	cmd := newCompletionCommand()
	assert.Equal(t, "completion [bash|zsh|fish|powershell]", cmd.Use)
	assert.Equal(t, "Generate shell completion script", cmd.Short)
	assert.Equal(t, []string{"bash", "zsh", "fish", "powershell"}, cmd.ValidArgs)
}

func TestCompletionOutput(t *testing.T) {
	shells := []struct {
		name    string
		contain string
	}{
		{"bash", "bash completion V2"},
		{"zsh", "zsh completion"},
		{"fish", "fish completion"},
		{"powershell", "Register-ArgumentCompleter"},
	}

	for _, sh := range shells {
		t.Run(sh.name, func(t *testing.T) {
			root := NewRootCommand()
			out := &bytes.Buffer{}
			root.SetOut(out)
			root.SetArgs([]string{"completion", sh.name})

			err := root.Execute()
			require.NoError(t, err)
			assert.Contains(t, out.String(), sh.contain)
		})
	}
}

func TestCompletionInvalidShell(t *testing.T) {
	root := NewRootCommand()
	out := &bytes.Buffer{}
	root.SetOut(out)
	root.SetErr(out)
	root.SetArgs([]string{"completion", "nushell"})

	err := root.Execute()
	assert.Error(t, err)
}

func TestCompletionNoArgs(t *testing.T) {
	root := NewRootCommand()
	out := &bytes.Buffer{}
	root.SetOut(out)
	root.SetErr(out)
	root.SetArgs([]string{"completion"})

	err := root.Execute()
	assert.Error(t, err)
}
