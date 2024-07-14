//
// Helena errors
//

package core

// Helena error stack level
type ErrorStackLevel struct {
	// Frame where the error occurred
	Frame *[]Value

	// Source where the error occurred
	Source *Source

	// Position where the error occurred
	Position *SourcePosition
}

//
// Helena error stack
//
// This class is used to propagate error info with results
//
type ErrorStack struct {
	// Error stack from its occurrence to its callers
	stack []ErrorStackLevel
}

func NewErrorStack() *ErrorStack {
	return &ErrorStack{[]ErrorStackLevel{}}
}

// Return depth of the stack, i.e. number of levels
func (errorStack *ErrorStack) Depth() uint {
	return uint(len(errorStack.stack))
}

// Push an error stack level
func (errorStack *ErrorStack) Push(level ErrorStackLevel) {
	errorStack.stack = append(errorStack.stack, level)
}

// Clear the error stack
func (errorStack *ErrorStack) Clear() {
	errorStack.stack = errorStack.stack[:0]
}

// Return the given error stack level
func (errorStack *ErrorStack) Level(level uint) ErrorStackLevel {
	return errorStack.stack[level]
}
