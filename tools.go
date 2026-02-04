//go:build tools

package tools

import (
	_ "github.com/samber/oops"
	_ "github.com/spf13/cobra"
	_ "github.com/spf13/viper"
	_ "github.com/stretchr/testify"
	_ "github.com/zalando/go-keyring"
)
