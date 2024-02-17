//
// Helena syntax checking and AST
//

package core

// import { Value } from "./values";

// /**
//  * Generic syntax error
//  */
// export class SyntaxError extends Error {
//   /**
//    *
//    * @param message - Error message
//    */
//   constructor(message) {
//     super(message);
//     this.name = "CompilationError";
//   }
// }

// /**
//  * Thrown when a word has an invalid structure
//  */
// export class InvalidWordStructureError extends SyntaxError {
//   /**
//    *
//    * @param message - Error message
//    */
//   constructor(message) {
//     super(message);
//     this.name = "InvalidWordStructureError";
//   }
// }

// /**
//  * Thrown when a word contains an unexpected morpheme
//  */
// export class UnexpectedMorphemeError extends SyntaxError {
//   /**
//    *
//    * @param message - Error message
//    */
//   constructor(message) {
//     super(message);
//     this.name = "UnexpectedMorphemeError";
//   }
// }

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

//
// Helena morpheme type
//
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

// Helena morpheme
//
// Morphemes are the basic constituents of words
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
	INVALID
)

//
// Helena syntax checker
//
// This class validates syntactic rules on words and determines their type
//
type SyntaxChecker struct{}

// Check word syntax and return its type
func (checker SyntaxChecker) CheckWord(word Word) WordType {
	//     if (word.morphemes.length == 0) return WordType.INVALID;
	//     switch (word.morphemes[0].type) {
	//       case MorphemeType.LITERAL: {
	//         const type = this.checkQualifiedWord(word);
	//         return type == WordType.INVALID ? this.checkCompoundWord(word) : type;
	//       }
	//       case MorphemeType.EXPRESSION:
	//         return this.checkCompoundWord(word);
	//       case MorphemeType.TUPLE:
	//       case MorphemeType.BLOCK:
	//         return this.checkQualifiedWord(word);
	//       case MorphemeType.STRING:
	//       case MorphemeType.HERE_STRING:
	//       case MorphemeType.TAGGED_STRING:
	//         return this.checkRootWord(word);
	//       case MorphemeType.LINE_COMMENT:
	//       case MorphemeType.BLOCK_COMMENT:
	//         return this.checkIgnoredWord(word);
	//       case MorphemeType.SUBSTITUTE_NEXT:
	//         return this.checkSubstitutionWord(word);
	//     }
	// TODO
	return INVALID
}

//   private checkRootWord(word: Word): WordType {
//     if (word.morphemes.length != 1) return WordType.INVALID;
//     return WordType.ROOT;
//   }

//   private checkCompoundWord(word: Word): WordType {
//     /* Lone morphemes are roots */
//     if (word.morphemes.length == 1) return WordType.ROOT;

//     if (this.checkStems(word.morphemes) < 0) return WordType.INVALID;
//     return WordType.COMPOUND;
//   }

//   private checkQualifiedWord(word: Word): WordType {
//     /* Lone morphemes are roots */
//     if (word.morphemes.length == 1) return WordType.ROOT;

//     const selectors = this.skipSelectors(word.morphemes, 1);
//     if (selectors != word.morphemes.length) return WordType.INVALID;
//     return WordType.QUALIFIED;
//   }

//   private checkSubstitutionWord(word: Word): WordType {
//     if (word.morphemes.length < 2) return WordType.INVALID;
//     const nbStems = this.checkStems(word.morphemes);
//     return nbStems < 0
//       ? WordType.INVALID
//       : nbStems > 1
//       ? WordType.COMPOUND
//       : WordType.SUBSTITUTION;
//   }

//   private checkIgnoredWord(word: Word): WordType {
//     if (word.morphemes.length != 1) return WordType.INVALID;
//     return WordType.IGNORED;
//   }

//   /**
//    * Check stem sequence in a compound or substitution word
//    *
//    * @param morphemes - Morphemes to check
//    *
//    * @returns           Number of stems, or < 0 if error
//    */
//   private checkStems(morphemes: Morpheme[]): number {
//     let nbStems = 0;
//     let substitute = false;
//     let hasTuples = false;
//     for (let i = 0; i < morphemes.length; i++) {
//       const morpheme = morphemes[i];
//       if (substitute) {
//         /* Expect valid root followed by selectors */
//         switch (morpheme.type) {
//           case MorphemeType.TUPLE:
//             hasTuples = true;
//           /* continued */
//           // eslint-disable-next-line no-fallthrough
//           case MorphemeType.LITERAL:
//           case MorphemeType.BLOCK:
//           case MorphemeType.EXPRESSION:
//             i = this.skipSelectors(morphemes, i + 1) - 1;
//             substitute = false;
//             break;

//           default:
//             return -1;
//         }
//       } else {
//         switch (morpheme.type) {
//           case MorphemeType.SUBSTITUTE_NEXT:
//             nbStems++;
//             substitute = true;
//             break;

//           case MorphemeType.LITERAL:
//           case MorphemeType.EXPRESSION:
//             nbStems++;
//             substitute = false;
//             break;

//           default:
//             return -1;
//         }
//       }
//     }
//     /* Tuples are invalid in compound words */
//     if (hasTuples && nbStems > 1) return -1;

//     return nbStems;
//   }

//   /**
//    * Skip all the selectors following a stem root
//    *
//    * @param morphemes - Morphemes to check
//    * @param first     - Index of first expected selector
//    *
//    * @returns           Index after selector sequence
//    */
//   private skipSelectors(morphemes: Morpheme[], first: number): number {
//     for (let i = first; i < morphemes.length; i++) {
//       const morpheme = morphemes[i];
//       switch (morpheme.type) {
//         case MorphemeType.TUPLE:
//         case MorphemeType.BLOCK:
//         case MorphemeType.EXPRESSION:
//           /* Eat up valid selector */
//           break;

//         default:
//           /* Stop there */
//           return i;
//       }
//     }
//     return morphemes.length;
//   }
