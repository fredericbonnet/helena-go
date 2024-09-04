//
// Helena intrinsic commands
//

package core

// Intrinsic command
type intrinsicCommand int8

func (intrinsicCommand) Execute(_ []Value, _ any) Result {
	return ERROR("intrinsic command cannot be called directly")
}

const (
	// Return the last result of the current program
	LAST_RESULT intrinsicCommand = iota

	// Get the result of the last frame of the current program, shift it right,
	// and evaluate the sentence again.
	SHIFT_LAST_FRAME_RESULT
)
