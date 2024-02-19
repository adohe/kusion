package generator

import v1 "kusionstack.io/kusion/pkg/apis/core/v1"

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
