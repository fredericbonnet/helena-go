//
// Helena value display
//

package core

import (
	"regexp"
)

// Value display function
type DisplayFunction func(displayable any) string

// Default display function
func DefaultDisplayFunction(displayable any) string {
	return UndisplayableValue()
}

// Undisplayable values represented as a block comment within a block.
func UndisplayableValue() string {
	return `{#{undisplayable value}#}`
}

// Undisplayable values represented as a label in block comment within a block.
func UndisplayableValueWithLabel(label string) string {
	return `{#{` + label + `}#}`
}

// Helena displayable object
//
// A displayable value will produce a string that, when evaluated as a word,
// will give a value that is isomorphic with the source value.
//
// Undisplayable values will use a placeholder.
type Displayable interface {
	// Return displayable string value
	//
	// Undisplayable objects will optionally use display function fn
	Display(fn DisplayFunction) string
}

// Return the display string of an object by dispatching to its `Display()`
// method if it implements `Displayable`, else use fn (defaults to
// `DefaultDisplayFunction`)
func Display(
	displayable any,
	fn DisplayFunction,
) string {
	if fn == nil {
		fn = DefaultDisplayFunction
	}
	switch o := displayable.(type) {
	case Displayable:
		return o.Display(fn)
	default:
		return fn(displayable)
	}
}

// Return the display string of str as a single literal or as a quoted string if
// the string contains special characters
func DisplayLiteralOrString(str string) string {
	tokenizer := Tokenizer{}
	tokens := tokenizer.Tokenize(str)
	if len(tokens) == 0 {
		return `""`
	}
	if len(tokens) == 1 && tokens[0].Type == TokenType_TEXT {
		return tokens[0].Literal
	}
	re := regexp.MustCompile(`[\\$"({[]`)
	return `"` + re.ReplaceAllStringFunc(str, func(c string) string { return `\` + c }) + `"`
}

// Return the display string of str as a single literal or as a block if the
// string contains special characters
//
// Mostly used to display qualified value sources
func DisplayLiteralOrBlock(str string) string {
	tokenizer := Tokenizer{}
	tokens := tokenizer.Tokenize(str)
	if len(tokens) == 0 {
		return `{}`
	}
	if len(tokens) == 1 && tokens[0].Type == TokenType_TEXT {
		return tokens[0].Literal
	}
	re := regexp.MustCompile(`[\\$"#(){}[\]]`)
	return `{` + re.ReplaceAllStringFunc(str, func(c string) string { return `\` + c }) + `}`
}

// Return a displayable list of objects
//
// Useful for sentences, tuples expressions, lists...
func DisplayList[T any](displayables []T, fn DisplayFunction) string {
	output := ""
	for i, displayable := range displayables {
		if i != 0 {
			output += " "
		}
		output += Display(displayable, fn)
	}
	return output
}
