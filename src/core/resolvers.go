//
// Helena resolvers
//

package core

//
// Variable resolver
//
type VariableResolver interface {
	// Resolve a value from its name
	Resolve(name string) (value Value, ok bool)
}

//
// Command resolver
//
type CommandResolver interface {
	// Resolve a command from its name
	Resolve(name Value) (value Command, ok bool)
}

//
// Selector resolver
//
type SelectorResolver interface {
	// Resolve a selector from a set of rules
	Resolve(rules []Value) (result TypedResult[Selector], ok bool)
}
