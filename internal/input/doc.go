// Package input contains operating-system adapters for remote input and volume
// control.
package input

// Adapter identifies the implementation that will handle control commands.
// Behavioral methods will be added with the first input feature.
type Adapter interface {
	Name() string
}

type namedAdapter string

func (a namedAdapter) Name() string { return string(a) }

func mockAdapter() Adapter { return namedAdapter("mock") }
