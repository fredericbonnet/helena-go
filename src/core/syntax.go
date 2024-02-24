//
// Helena syntax checking and AST
//

package core

//
// Generic syntax error
//
type SyntaxError struct {
	message string
}

func (e SyntaxError) Error() string {
	return e.message
}

// Thrown when a word has an invalid structure
var InvalidWordStructureError = SyntaxError{"invalid word structure"}

// Thrown when a word contains an unexpected morpheme
var UnexpectedMorphemeError = SyntaxError{"unexpected morpheme"}

//
// Helena script
//
// Scripts are lists of sentences
//
type Script struct {
	// Sentences that compose the script
	Sentences []Sentence
}

//
// Helena sentence
//
// Sentences are lists of words or values
//
type Sentence struct {
	// Words that compose the sentence
	// Words: (Word | Value)[] = [];
	Words []Word
}

//
// Helena word
//
// Words are made of morphemes
//
type Word struct {
	// Morphemes that compose the word
	Morphemes []Morpheme
}

// Helena morpheme type
type MorphemeType int8

const (
	MorphemeType_LITERAL MorphemeType = iota
	MorphemeType_TUPLE
	MorphemeType_BLOCK
	MorphemeType_EXPRESSION
	MorphemeType_STRING
	MorphemeType_HERE_STRING
	MorphemeType_TAGGED_STRING
	MorphemeType_LINE_COMMENT
	MorphemeType_BLOCK_COMMENT
	MorphemeType_SUBSTITUTE_NEXT
)

//
// Helena morpheme
//
// Morphemes are the basic constituents of words
//
type Morpheme interface {
	// Type identifier
	Type() MorphemeType
}

//
// Literal morpheme
//
// Literals are plain strings
//
type LiteralMorpheme struct {
	// Literal string value
	Value string
}

func (morpheme LiteralMorpheme) Type() MorphemeType {
	return MorphemeType_LITERAL
}

//
// Tuple morpheme
//
// Tuples are scripts between tuple delimiters
//
type TupleMorpheme struct {
	// Tuple script content
	Subscript Script
}

func (morpheme TupleMorpheme) Type() MorphemeType {
	return MorphemeType_TUPLE
}

//
// Block morpheme
//
// Blocks are scripts or strings between block delimiters
//
type BlockMorpheme struct {
	// Block script content
	Subscript Script

	// Block string value
	Value string
}

func (morpheme BlockMorpheme) Type() MorphemeType {
	return MorphemeType_BLOCK
}

//
// Expression morpheme
//
// Expressions are scripts between expression delimiters
//
type ExpressionMorpheme struct {
	// Expression script content
	Subscript Script
}

func (morpheme ExpressionMorpheme) Type() MorphemeType {
	return MorphemeType_EXPRESSION
}

//
// String morpheme
//
// Strings are made of morphemes between single string delimiters
//
type StringMorpheme struct {
	// String content
	Morphemes []Morpheme
}

func (morpheme StringMorpheme) Type() MorphemeType {
	return MorphemeType_STRING
}

//
// Here-string morpheme
//
// Here-strings are plain strings between three or more string delimiters
//
type HereStringMorpheme struct {
	// Here-string value
	Value string

	// Number of string delimiters around content
	DelimiterLength uint
}

func (morpheme HereStringMorpheme) Type() MorphemeType {
	return MorphemeType_HERE_STRING
}

//
// Tagged string morpheme
//
// Tagged strings are plain strings between two string delimiters and an
// arbitrary tag
//
type TaggedStringMorpheme struct {
	// Tagged string value
	Value string

	// Tag
	Tag string
}

func (morpheme TaggedStringMorpheme) Type() MorphemeType {
	return MorphemeType_TAGGED_STRING
}

//
// Line comment morpheme
//
type LineCommentMorpheme struct {
	// Line comment content
	Value string

	// Number of comment characters before content
	DelimiterLength uint
}

func (morpheme LineCommentMorpheme) Type() MorphemeType {
	return MorphemeType_LINE_COMMENT
}

//
// Block comment morpheme
//
type BlockCommentMorpheme struct {
	// Block comment content
	Value string

	// Number of comment characters around content
	DelimiterLength uint
}

func (morpheme BlockCommentMorpheme) Type() MorphemeType {
	return MorphemeType_BLOCK_COMMENT
}

//
// Substitute Next morpheme
//
// Always followed by a sequence of morphemes to substitute; stale substitutions
// (substitution characters with no such sequence) are always converted to
// `LiteralMorpheme`}` and should not appear in a well-formed AST
//
type SubstituteNextMorpheme struct {
	// Simple or expanded substitution flag
	Expansion bool

	// Number of substitutions to perform
	Levels uint

	// Literal value; can be safely ignored
	Value string
}

func (morpheme SubstituteNextMorpheme) Type() MorphemeType {
	return MorphemeType_SUBSTITUTE_NEXT
}

//
// Helena word type
//
// Valid word types must respect strict syntactic rules
//
type WordType int8

const (
	// Roots are monomorphemic words
	WordType_ROOT WordType = iota

	// Compounds are words made of several stems, that don't fit in the other categories
	WordType_COMPOUND

	// Substitions are root or qualified words prefixed by a substitute morpheme
	WordType_SUBSTITUTION

	// Qualified words are root words followed by selectors
	WordType_QUALIFIED

	// Ignored words are line and block comments
	WordType_IGNORED

	// Invalid word structure
	WordType_INVALID
)

//
// Helena syntax checker
//
// This class validates syntactic rules on words and determines their type
//
type SyntaxChecker struct{}

// Check word syntax and return its type
func (checker SyntaxChecker) CheckWord(word Word) WordType {
	if len(word.Morphemes) == 0 {
		return WordType_INVALID
	}
	switch word.Morphemes[0].Type() {
	case MorphemeType_LITERAL:
		{
			type_ := checker.checkQualifiedWord(word)
			if type_ == WordType_INVALID {
				return checker.checkCompoundWord(word)
			}
			return type_
		}
	case MorphemeType_EXPRESSION:
		return checker.checkCompoundWord(word)
	case MorphemeType_TUPLE,
		MorphemeType_BLOCK:
		return checker.checkQualifiedWord(word)
	case MorphemeType_STRING,
		MorphemeType_HERE_STRING,
		MorphemeType_TAGGED_STRING:
		return checker.checkRootWord(word)
	case MorphemeType_LINE_COMMENT,
		MorphemeType_BLOCK_COMMENT:
		return checker.checkIgnoredWord(word)
	case MorphemeType_SUBSTITUTE_NEXT:
		return checker.checkSubstitutionWord(word)
	}
	panic("CANTHAPPEN")
}

func (checker SyntaxChecker) checkRootWord(word Word) WordType {
	if len(word.Morphemes) != 1 {
		return WordType_INVALID
	}
	return WordType_ROOT
}

func (checker SyntaxChecker) checkCompoundWord(word Word) WordType {
	/* Lone morphemes are roots */
	if len(word.Morphemes) == 1 {
		return WordType_ROOT
	}

	if checker.checkStems(word.Morphemes) < 0 {
		return WordType_INVALID
	}
	return WordType_COMPOUND
}

func (checker SyntaxChecker) checkQualifiedWord(word Word) WordType {
	/* Lone morphemes are roots */
	if len(word.Morphemes) == 1 {
		return WordType_ROOT
	}

	selectors := checker.skipSelectors(word.Morphemes, 1)
	if selectors != len(word.Morphemes) {
		return WordType_INVALID
	}
	return WordType_QUALIFIED
}

func (checker SyntaxChecker) checkSubstitutionWord(word Word) WordType {
	if len(word.Morphemes) < 2 {
		return WordType_INVALID
	}
	nbStems := checker.checkStems(word.Morphemes)
	if nbStems < 0 {
		return WordType_INVALID
	}
	if nbStems > 1 {
		return WordType_COMPOUND
	}
	return WordType_SUBSTITUTION
}

func (checker SyntaxChecker) checkIgnoredWord(word Word) WordType {
	if len(word.Morphemes) != 1 {
		return WordType_INVALID
	}
	return WordType_IGNORED
}

// Check stem sequence in a compound or substitution word
//
// Returns number of stems, or < 0 if error
func (checker SyntaxChecker) checkStems(morphemes []Morpheme) int {
	nbStems := 0
	substitute := false
	hasTuples := false
	for i := 0; i < len(morphemes); i++ {
		morpheme := morphemes[i]
		if substitute {
			/* Expect valid root followed by selectors */
			switch morpheme.Type() {
			case MorphemeType_TUPLE:
				hasTuples = true
				i = checker.skipSelectors(morphemes, i+1) - 1
				substitute = false

			case MorphemeType_LITERAL,
				MorphemeType_BLOCK,
				MorphemeType_EXPRESSION:
				i = checker.skipSelectors(morphemes, i+1) - 1
				substitute = false

			default:
				return -1
			}
		} else {
			switch morpheme.Type() {
			case MorphemeType_SUBSTITUTE_NEXT:
				nbStems++
				substitute = true

			case MorphemeType_LITERAL,
				MorphemeType_EXPRESSION:
				nbStems++
				substitute = false

			default:
				return -1
			}
		}
	}
	/* Tuples are invalid in compound words */
	if hasTuples && nbStems > 1 {
		return -1
	}

	return nbStems
}

// Skip all the selectors following a stem root starting at first
//
// Returns index after selector sequence
func (checker SyntaxChecker) skipSelectors(morphemes []Morpheme, first int) int {
	for i := first; i < len(morphemes); i++ {
		morpheme := morphemes[i]
		switch morpheme.Type() {
		case MorphemeType_TUPLE,
			MorphemeType_BLOCK,
			MorphemeType_EXPRESSION:
			/* Eat up valid selector */

		default:
			/* Stop there */
			return i
		}
	}
	return len(morphemes)
}
