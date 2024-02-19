package generate

import (
	"github.com/spf13/cobra"

	"kusionstack.io/kusion/pkg/apis/core/v1"
)

var (
	generateLong = ``

	generateExample = ``
)

// GenerateFlags directly reflect the information that CLI is gathering via flags. They will be converted to
// GenerateOptions, which reflect the runtime requirements for the command.
//
// This structure reduces the transformation to wiring and makes the logic itself easy to unit test.
type GenerateFlags struct {
}

// GenerateOptions defines flags and other configuration parameters for the `generate` command.
type GenerateOptions struct {
}

// NewGenerateFlags returns a default GenerateFlags
func NewGenerateFlags() *GenerateFlags {
	return &GenerateFlags{}
}

// NewCmdGenerate creates the `generate` command.
func NewCmdGenerate() *cobra.Command {
	cmd := &cobra.Command{
		Long:    generateLong,
		Example: generateExample,
	}
	return cmd
}

// Validate verifies if GenerateOptions are valid and without conflicts.
func (o *GenerateOptions) Validate() error {
	return nil
}

// Run executes the `generate` command.
func (o *GenerateOptions) Run() error {
	return nil
}

// Generator is an interface for things that can generate versioned Intent from
// configuration code.
type Generator interface {
	Generate() (*v1.Intent, error)
}

// DefaultGenerator is the default Generator implementation.
type DefaultGenerator struct {
}

func (g *DefaultGenerator) Generate() (*v1.Intent, error) {
	return nil, nil
}
