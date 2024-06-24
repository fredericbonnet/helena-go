package core_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "helena/core"
)

const _LITERAL = "literal"
const _TUPLE = "(word1 word2)"
const _BLOCK = "{word1 word2}"
const _EXPRESSION = "[word1 word2]"
const _STRING = `"$some [string]"`
const _HERE_STRING = `"""some here "" string"""`
const _TAGGED_STRING = "\"\"TAG\nsome tagged \" string\nTAG\"\""
const _LINE_COMMENT = "# some [{comment"
const _BLOCK_COMMENT = "#{ some [block {comment }#"

type testMorpheme struct {
	type_ string
	value string
}

var roots = []testMorpheme{
	{"literal", _LITERAL},
	{"tuple", _TUPLE},
	{"block", _BLOCK},
	{"expression", _EXPRESSION},
}
var qualifiedSources = []testMorpheme{
	{"literal", _LITERAL},
	{"tuple", _TUPLE},
	{"block", _BLOCK},
}
var monomorphemes = []testMorpheme{
	{"string", _STRING},
	{"here-string", _HERE_STRING},
	{"tagged string", _TAGGED_STRING},
}
var ignored = []testMorpheme{
	{"line comment", _LINE_COMMENT},
	{"block comment", _BLOCK_COMMENT},
}

var _ = Describe("SyntaxChecker", func() {
	var tokenizer Tokenizer
	var parser *Parser
	var checker SyntaxChecker

	parse := func(script string) *Script {
		return parser.Parse(tokenizer.Tokenize(script), nil).Script
	}
	firstWord := func(script *Script) Word {
		return script.Sentences[0].Words[0].Word
	}

	BeforeEach(func() {
		tokenizer = Tokenizer{}
		parser = NewParser(nil)
		checker = SyntaxChecker{}
	})

	Describe("roots", func() {
		for _, tm := range append(append([]testMorpheme{}, roots...), monomorphemes...) {
			type_ := tm.type_
			value := tm.value
			Specify(type_+" root", func() {
				script := parse(value)
				word := firstWord(script)
				Expect(word.Morphemes).To(HaveLen(1))
				Expect(checker.CheckWord(word)).To(Equal(WordType_ROOT))
			})
		}
	})
	Describe("compounds", func() {
		Specify("literal prefix", func() {
			script := parse(_LITERAL + "$" + _BLOCK)
			word := firstWord(script)
			Expect(word.Morphemes).To(HaveLen(3))
			Expect(checker.CheckWord(word)).To(Equal(WordType_COMPOUND))
		})
		Specify("expression prefix", func() {
			script := parse(_EXPRESSION + _LITERAL)
			word := firstWord(script)
			Expect(word.Morphemes).To(HaveLen(2))
			Expect(checker.CheckWord(word)).To(Equal(WordType_COMPOUND))
		})
		Specify("substitution prefix", func() {
			script := parse("$" + _BLOCK + _TUPLE + _LITERAL)
			word := firstWord(script)
			Expect(word.Morphemes).To(HaveLen(4))
			Expect(checker.CheckWord(word)).To(Equal(WordType_COMPOUND))
		})
		Specify("complex case", func() {
			script := parse(_LITERAL + "$" + _BLOCK + _EXPRESSION + "$" + _LITERAL)
			word := firstWord(script)
			Expect(word.Morphemes).To(HaveLen(6))
			Expect(checker.CheckWord(word)).To(Equal(WordType_COMPOUND))
		})
		Describe("exceptions", func() {
			for _, tm := range []testMorpheme{
				{"tuple", _TUPLE},
				{"block", _BLOCK},
			} {
				type_ := tm.type_
				value := tm.value
				Specify(type_+"/literal", func() {
					script := parse(value + _LITERAL)
					word := firstWord(script)
					Expect(word.Morphemes).To(HaveLen(2))
					Expect(checker.CheckWord(word)).To(Equal(WordType_INVALID))
				})
				Specify(type_+"/substitution", func() {
					script := parse(value + "$" + _LITERAL)
					word := firstWord(script)
					Expect(word.Morphemes).To(HaveLen(3))
					Expect(checker.CheckWord(word)).To(Equal(WordType_INVALID))
				})
				Specify("expression/"+type_, func() {
					script := parse(_EXPRESSION + value)
					word := firstWord(script)
					Expect(word.Morphemes).To(HaveLen(2))
					Expect(checker.CheckWord(word)).To(Equal(WordType_INVALID))
				})
				Specify("literal/"+type_+"/literal", func() {
					script := parse(_LITERAL + value + _LITERAL)
					word := firstWord(script)
					Expect(word.Morphemes).To(HaveLen(3))
					Expect(checker.CheckWord(word)).To(Equal(WordType_INVALID))
				})
				Specify("literal/"+type_+"/substitution", func() {
					script := parse(_LITERAL + value + "$" + _LITERAL)
					word := firstWord(script)
					Expect(word.Morphemes).To(HaveLen(4))
					Expect(checker.CheckWord(word)).To(Equal(WordType_INVALID))
				})
			}
			Specify("literal/tuple substitution", func() {
				script := parse(_LITERAL + "$" + _TUPLE)
				word := firstWord(script)
				Expect(word.Morphemes).To(HaveLen(3))
				Expect(checker.CheckWord(word)).To(Equal(WordType_INVALID))
			})
			Specify("tuple substitution/literal", func() {
				script := parse("$" + _TUPLE + _LITERAL)
				word := firstWord(script)
				Expect(word.Morphemes).To(HaveLen(3))
				Expect(checker.CheckWord(word)).To(Equal(WordType_INVALID))
			})
		})
	})
	Describe("substitutions", func() {
		for _, tm := range roots {
			type_ := tm.type_
			value := tm.value
			Describe(type_+" source", func() {
				Specify("simple", func() {
					script := parse("$" + value)
					word := firstWord(script)
					Expect(word.Morphemes).To(HaveLen(2))
					Expect(checker.CheckWord(word)).To(Equal(WordType_SUBSTITUTION))
				})
				Specify("double", func() {
					script := parse("$$" + value)
					word := firstWord(script)
					Expect(word.Morphemes).To(HaveLen(3))
					Expect(checker.CheckWord(word)).To(Equal(WordType_SUBSTITUTION))
				})
				Specify("indexed selector", func() {
					script := parse("$" + value + _EXPRESSION)
					word := firstWord(script)
					Expect(word.Morphemes).To(HaveLen(3))
					Expect(checker.CheckWord(word)).To(Equal(WordType_SUBSTITUTION))
				})
				Specify("keyed selector", func() {
					script := parse("$" + value + _TUPLE)
					word := firstWord(script)
					Expect(word.Morphemes).To(HaveLen(3))
					Expect(checker.CheckWord(word)).To(Equal(WordType_SUBSTITUTION))
				})
				Specify("generic selector", func() {
					script := parse("$" + value + _BLOCK)
					word := firstWord(script)
					Expect(word.Morphemes).To(HaveLen(3))
					Expect(checker.CheckWord(word)).To(Equal(WordType_SUBSTITUTION))
				})
				Specify("multiple selectors", func() {
					script := parse(
						"$" + value + _TUPLE + _BLOCK + _EXPRESSION + _TUPLE + _EXPRESSION,
					)
					word := firstWord(script)
					Expect(word.Morphemes).To(HaveLen(7))
					Expect(checker.CheckWord(word)).To(Equal(WordType_SUBSTITUTION))
				})
			})
		}
	})
	Describe("qualified words", func() {
		for _, tm := range qualifiedSources {
			value := tm.value
			Describe(tm.type_+" source", func() {
				Specify("indexed selector", func() {
					script := parse(value + _EXPRESSION)
					word := firstWord(script)
					Expect(word.Morphemes).To(HaveLen(2))
					Expect(checker.CheckWord(word)).To(Equal(WordType_QUALIFIED))
				})
				Specify("keyed selector", func() {
					script := parse(value + _TUPLE)
					word := firstWord(script)
					Expect(word.Morphemes).To(HaveLen(2))
					Expect(checker.CheckWord(word)).To(Equal(WordType_QUALIFIED))
				})
				Specify("generic selector", func() {
					script := parse(value + _BLOCK)
					word := firstWord(script)
					Expect(word.Morphemes).To(HaveLen(2))
					Expect(checker.CheckWord(word)).To(Equal(WordType_QUALIFIED))
				})
				Specify("multiple selectors", func() {
					script := parse(
						value + _TUPLE + _BLOCK + _EXPRESSION + _TUPLE + _EXPRESSION,
					)
					word := firstWord(script)
					Expect(word.Morphemes).To(HaveLen(6))
					Expect(checker.CheckWord(word)).To(Equal(WordType_QUALIFIED))
				})
			})
		}
		Describe("exceptions", func() {
			Specify("trailing morphemes", func() {
				script := parse(
					_LITERAL + _TUPLE + _TUPLE + _EXPRESSION + _TUPLE + _EXPRESSION + _LITERAL,
				)
				word := firstWord(script)
				Expect(word.Morphemes).To(HaveLen(7))
				Expect(checker.CheckWord(word)).To(Equal(WordType_INVALID))
			})
		})
	})
	Describe("ignored words", func() {
		for _, tm := range ignored {
			type_ := tm.type_
			value := tm.value
			Specify(type_, func() {
				script := parse(value)
				word := firstWord(script)
				Expect(word.Morphemes).To(HaveLen(1))
				Expect(checker.CheckWord(word)).To(Equal(WordType_IGNORED))
			})
		}
	})

	Describe("impossible cases", func() {
		Specify("empty word", func() {
			word := Word{}
			Expect(checker.CheckWord(word)).To(Equal(WordType_INVALID))
		})
		Specify("empty substitution", func() {
			word := Word{}
			word.Morphemes = append(word.Morphemes, SubstituteNextMorpheme{})
			Expect(checker.CheckWord(word)).To(Equal(WordType_INVALID))
		})
		Describe("incompatible morphemes", func() {
			for _, tm1 := range append(append([]testMorpheme{}, monomorphemes...), ignored...) {
				type1 := tm1.type_
				value1 := tm1.value
				for _, tm2 := range append(append([]testMorpheme{}, roots...), monomorphemes...) {
					type2 := tm2.type_
					value2 := tm2.value
					Specify(type1+"/"+type2, func() {
						word := Word{}
						word.Morphemes = append(word.Morphemes,
							firstWord(parse(value1)).Morphemes[0],
							firstWord(parse(value2)).Morphemes[0],
						)
						Expect(checker.CheckWord(word)).To(Equal(WordType_INVALID))
					})
					Specify(type2+"/"+type1, func() {
						word := Word{}
						word.Morphemes = append(word.Morphemes,
							firstWord(parse(value2)).Morphemes[0],
							firstWord(parse(value1)).Morphemes[0],
						)
					})
				}
			}
			Specify("substitution", func() {
				word := Word{}
				word.Morphemes = append(word.Morphemes,
					SubstituteNextMorpheme{},
					BlockCommentMorpheme{},
				)
				Expect(checker.CheckWord(word)).To(Equal(WordType_INVALID))
			})
		})
	})
})
