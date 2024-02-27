//
// Helena commands
//

package core

// Helena command
type Command interface {
	// Execute the command by passing a list of argument values
	//
	// An optional opaque context can be passed
	Execute(args []Value, context any) Result
}

type ResumableCommand interface {
	// Resume the previously yielded command with a result to yield back
	//
	// An optional opaque context can be passed
	Resume(result Result, context any) Result
}

type CommandHelpOptions struct {
	Prefix string
	Skip   uint
}
type CommandWithHelp interface {
	// Return help for the command and a list of arguments
	//
	// Provided arguments will be validated against the command signature
	//
	// Help formating options can be provided
	Help(args []Value, options CommandHelpOptions, context any) Result
}
