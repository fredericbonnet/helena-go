//
// Helena value display
//

package core

import (
	"regexp"
)

// Value display function
type DisplayFunction func(displayable Displayable) string

// Default display function
func DefaultDisplayFunction(displayable Displayable) string {
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
	// UnUndisplayable objects will optionally use display function fn
	display(fn DisplayFunction) string
}

// /**
//  * Display an object by dispatching to its `display()` function if it exists,
//  * else use the provided function.
//  *
//  * @param displayable - Object to display
//  * @param fn          - Display function for undisplayable objects
//  *
//  * @returns             Displayable string
//  */
// export function display(
//   displayable: Displayable,
//   fn: DisplayFunction = defaultDisplayFunction
// ): string {
//   return displayable.display?.(fn) ?? fn(displayable);
// }

// Return a displayable string as a single literal or as a quoted string if the
// string contains special characters
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

// Return a displayable string as a single literal or as a block if the string
// contains special characters
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

// /**
//  * Display a list of objects
//  *
//  * Useful for sentences, tuples expressions, lists...
//  *
//  * @param displayables - Objects to display
//  * @param fn           - Display function for undisplayable objects
//  *
//  * @returns              Displayable string
//  */
// export function displayList(
//   displayables: Displayable[],
//   fn: DisplayFunction = defaultDisplayFunction
// ): string {
//   return `${displayables
//     .map((displayable) => display(displayable, fn))
//     .join(" ")}`;
// }
