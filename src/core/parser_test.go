package core_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "helena/core"
)

type _M_SCRIPT struct {
	sentences []_M_SENTENCE
	position  *SourcePosition
}
type _M_SENTENCE struct {
	words    []_M_WORD
	position *SourcePosition
}
type _M_WORD struct {
	morphemes []any
	position  *SourcePosition
}

type _M_LITERAL struct {
	LITERAL  string
	position *SourcePosition
}
type _M_TUPLE struct {
	subscript _M_SCRIPT
	position  *SourcePosition
}
type _M_BLOCK struct {
	subscript _M_SCRIPT
	position  *SourcePosition
}
type _M_EXPRESSION struct {
	subscript _M_SCRIPT
	position  *SourcePosition
}
type _M_STRING struct {
	morphemes []any
	position  *SourcePosition
}
type _M_HERE_STRING struct {
	HERE_STRING string
	position    *SourcePosition
}
type _M_TAGGED_STRING struct {
	TAGGED_STRING string
	position      *SourcePosition
}
type _M_LINE_COMMENT struct {
	LINE_COMMENT string
	position     *SourcePosition
}
type _M_BLOCK_COMMENT struct {
	BLOCK_COMMENT string
	position      *SourcePosition
}
type _M_SUBSTITUTE_NEXT struct {
	SUBSTITUTE_NEXT string
	position        *SourcePosition
}
type _M_EXPAND_NEXT struct {
	EXPAND_NEXT string
	position    *SourcePosition
}

func M_SCRIPT(s ..._M_SENTENCE) _M_SCRIPT { return M_SCRIPT_P(nil, s...) }
func M_SCRIPT_P(p *SourcePosition, s ..._M_SENTENCE) _M_SCRIPT {
	if len(s) > 0 {
		return _M_SCRIPT{s, p}
	} else {
		return _M_SCRIPT{[]_M_SENTENCE{}, p}
	}
}
func M_SENTENCE(w ..._M_WORD) _M_SENTENCE { return M_SENTENCE_P(nil, w...) }
func M_SENTENCE_P(p *SourcePosition, w ..._M_WORD) _M_SENTENCE {
	if len(w) > 0 {
		return _M_SENTENCE{w, p}
	} else {
		return _M_SENTENCE{[]_M_WORD{}, p}
	}
}
func M_WORD(m ...any) _M_WORD { return M_WORD_P(nil, m...) }
func M_WORD_P(p *SourcePosition, m ...any) _M_WORD {
	if len(m) > 0 {
		return _M_WORD{m, p}
	} else {
		return _M_WORD{[]any{}, p}
	}
}

func M_LITERAL(s string) any                            { return _M_LITERAL{s, nil} }
func M_LITERAL_P(p *SourcePosition, s string) any       { return _M_LITERAL{s, p} }
func M_TUPLE(s ..._M_SENTENCE) any                      { return M_TUPLE_P(nil, M_SCRIPT(s...)) }
func M_TUPLE_P(p *SourcePosition, s _M_SCRIPT) any      { return _M_TUPLE{s, p} }
func M_BLOCK(s ..._M_SENTENCE) any                      { return M_BLOCK_P(nil, M_SCRIPT(s...)) }
func M_BLOCK_P(p *SourcePosition, s _M_SCRIPT) any      { return _M_BLOCK{s, p} }
func M_EXPRESSION(s ..._M_SENTENCE) any                 { return M_EXPRESSION_P(nil, M_SCRIPT(s...)) }
func M_EXPRESSION_P(p *SourcePosition, s _M_SCRIPT) any { return _M_EXPRESSION{s, p} }
func M_STRING(m ...any) any                             { return M_STRING_P(nil, m...) }
func M_STRING_P(p *SourcePosition, m ...any) any {
	if len(m) > 0 {
		return _M_STRING{m, p}
	} else {
		return _M_STRING{[]any{}, p}
	}
}
func M_HERE_STRING(s string) any                          { return _M_HERE_STRING{s, nil} }
func M_HERE_STRING_P(p *SourcePosition, s string) any     { return _M_HERE_STRING{s, p} }
func M_TAGGED_STRING(s string) any                        { return _M_TAGGED_STRING{s, nil} }
func M_TAGGED_STRING_P(p *SourcePosition, s string) any   { return _M_TAGGED_STRING{s, p} }
func M_LINE_COMMENT(s string) any                         { return _M_LINE_COMMENT{s, nil} }
func M_LINE_COMMENT_P(p *SourcePosition, s string) any    { return _M_LINE_COMMENT{s, p} }
func M_BLOCK_COMMENT(s string) any                        { return _M_BLOCK_COMMENT{s, nil} }
func M_BLOCK_COMMENT_P(p *SourcePosition, s string) any   { return _M_BLOCK_COMMENT{s, p} }
func M_SUBSTITUTE_NEXT(s string) any                      { return _M_SUBSTITUTE_NEXT{s, nil} }
func M_SUBSTITUTE_NEXT_P(p *SourcePosition, s string) any { return _M_SUBSTITUTE_NEXT{s, p} }
func M_EXPAND_NEXT(s string) any                          { return _M_EXPAND_NEXT{s, nil} }
func M_EXPAND_NEXT_P(p *SourcePosition, s string) any     { return _M_EXPAND_NEXT{s, p} }

func mapMorpheme(morpheme Morpheme) any {
	switch morpheme.Type() {
	case MorphemeType_LITERAL:
		{
			m := morpheme.(LiteralMorpheme)
			return _M_LITERAL{m.Value, m.Position()}
		}

	case MorphemeType_TUPLE:
		{
			m := morpheme.(TupleMorpheme)
			return _M_TUPLE{toTree(&m.Subscript), m.Position()}
		}

	case MorphemeType_BLOCK:
		{
			m := morpheme.(BlockMorpheme)
			return _M_BLOCK{toTree(&m.Subscript), m.Position()}
		}

	case MorphemeType_EXPRESSION:
		{
			m := morpheme.(ExpressionMorpheme)
			return _M_EXPRESSION{toTree(&m.Subscript), m.Position()}
		}

	case MorphemeType_STRING:
		{
			m := morpheme.(StringMorpheme)
			morphemes := make([]any, len(m.Morphemes))
			for k, morpheme := range m.Morphemes {
				morphemes[k] = mapMorpheme(morpheme)
			}
			return _M_STRING{morphemes, m.Position()}
		}

	case MorphemeType_HERE_STRING:
		{
			m := morpheme.(HereStringMorpheme)
			return _M_HERE_STRING{m.Value, m.Position()}
		}

	case MorphemeType_TAGGED_STRING:
		{
			m := morpheme.(TaggedStringMorpheme)
			return _M_TAGGED_STRING{m.Value, m.Position()}
		}

	case MorphemeType_LINE_COMMENT:
		{
			m := morpheme.(LineCommentMorpheme)
			return _M_LINE_COMMENT{m.Value, m.Position()}
		}

	case MorphemeType_BLOCK_COMMENT:
		{
			m := morpheme.(BlockCommentMorpheme)
			return _M_BLOCK_COMMENT{m.Value, m.Position()}
		}

	case MorphemeType_SUBSTITUTE_NEXT:
		{
			m := morpheme.(SubstituteNextMorpheme)
			if m.Expansion {
				return _M_EXPAND_NEXT{m.Value, m.Position()}
			} else {
				return _M_SUBSTITUTE_NEXT{m.Value, m.Position()}
			}
		}

	default:
		panic("CANTHAPPEN")
	}
}

func toTree(script *Script) _M_SCRIPT {
	sentences := make([]_M_SENTENCE, len(script.Sentences))
	for i, sentence := range script.Sentences {
		words := make([]_M_WORD, len(sentence.Words))
		for j, word := range sentence.Words {
			morphemes := make([]any, len(word.Word.Morphemes))
			for k, morpheme := range word.Word.Morphemes {
				morphemes[k] = mapMorpheme(morpheme)
			}
			words[j] = _M_WORD{morphemes, word.Word.Position}
		}
		sentences[i] = _M_SENTENCE{words, sentence.Position}
	}
	return _M_SCRIPT{sentences, script.Position}
}

var _ = Describe("Parser", func() {
	var tokenizer Tokenizer
	var parser *Parser

	parse := func(script string) *Script {
		return parser.Parse(tokenizer.Tokenize(script), nil).Script
	}

	BeforeEach(func() {
		tokenizer = Tokenizer{}
		parser = NewParser(nil)
	})

	Describe("scripts", func() {
		Specify("empty script", func() {
			script := parse("")
			Expect(script.Sentences).To(BeEmpty())
		})
		Specify("blank lines", func() {
			script := parse(" \n\n    \n")
			Expect(script.Sentences).To(BeEmpty())
		})
		Specify("single sentence", func() {
			script := parse("sentence")
			Expect(toTree(script)).To(Equal(M_SCRIPT(M_SENTENCE(M_WORD(M_LITERAL("sentence"))))))
		})
		Specify("single sentence surrounded by blank lines", func() {
			script := parse("  \nsentence\n  ")
			Expect(toTree(script)).To(Equal(M_SCRIPT(M_SENTENCE(M_WORD(M_LITERAL("sentence"))))))
		})
		Specify("two sentences separated by newline", func() {
			script := parse("sentence1\nsentence2")
			Expect(toTree(script)).To(Equal(M_SCRIPT(
				M_SENTENCE(M_WORD(M_LITERAL("sentence1"))),
				M_SENTENCE(M_WORD(M_LITERAL("sentence2"))),
			)))
		})
		Specify("two sentences separated by semicolon", func() {
			script := parse("sentence1;sentence2")
			Expect(toTree(script)).To(Equal(M_SCRIPT(
				M_SENTENCE(M_WORD(M_LITERAL("sentence1"))),
				M_SENTENCE(M_WORD(M_LITERAL("sentence2"))),
			)))
		})
		Specify("blank sentences are ignored", func() {
			script := parse("\nsentence1;; \t  ;\n\n \t   \nsentence2\n")
			Expect(toTree(script)).To(Equal(M_SCRIPT(
				M_SENTENCE(M_WORD(M_LITERAL("sentence1"))),
				M_SENTENCE(M_WORD(M_LITERAL("sentence2"))),
			)))
		})
	})
	Describe("words", func() {
		Describe("literals", func() {
			Specify("single literal", func() {
				script := parse("word")
				Expect(toTree(script)).To(Equal(M_SCRIPT(M_SENTENCE(M_WORD(M_LITERAL("word"))))))
			})
			Specify("single literal surrounded by spaces", func() {
				script := parse(" word ")
				Expect(toTree(script)).To(Equal(M_SCRIPT(M_SENTENCE(M_WORD(M_LITERAL("word"))))))
			})
			Specify("single literal with escape sequences", func() {
				script := parse("one\\tword")
				Expect(toTree(script)).To(Equal(M_SCRIPT(M_SENTENCE(M_WORD(M_LITERAL("one\tword"))))))
			})
			Specify("two literals separated by whitespace", func() {
				script := parse("word1 word2")
				Expect(toTree(script)).To(Equal(M_SCRIPT(
					M_SENTENCE(M_WORD(M_LITERAL("word1")), M_WORD(M_LITERAL("word2"))),
				)))
			})
			Specify("two literals separated by continuation", func() {
				script := parse("word1\\\nword2")
				Expect(toTree(script)).To(Equal(M_SCRIPT(
					M_SENTENCE(M_WORD(M_LITERAL("word1")), M_WORD(M_LITERAL("word2"))),
				)))
			})
		})
		Describe("tuples", func() {
			Specify("empty tuple", func() {
				script := parse("()")
				Expect(toTree(script)).To(Equal(M_SCRIPT(M_SENTENCE(M_WORD(M_TUPLE())))))
			})
			Specify("tuple with one word", func() {
				script := parse("(word)")
				Expect(toTree(script)).To(Equal(M_SCRIPT(
					M_SENTENCE(M_WORD(M_TUPLE(M_SENTENCE(M_WORD(M_LITERAL("word")))))),
				)))
			})
			Specify("tuple with two levels", func() {
				script := parse("(word1 (subword1 subword2) word2)")
				Expect(toTree(script)).To(Equal(M_SCRIPT(
					M_SENTENCE(
						M_WORD(
							M_TUPLE(
								M_SENTENCE(
									M_WORD(M_LITERAL("word1")),
									M_WORD(
										M_TUPLE(
											M_SENTENCE(
												M_WORD(M_LITERAL("subword1")),
												M_WORD(M_LITERAL("subword2")),
											),
										),
									),
									M_WORD(M_LITERAL("word2")),
								),
							),
						),
					),
				)))
			})
			Describe("exceptions", func() {
				Specify("unterminated tuple", func() {
					tokens := tokenizer.Tokenize("(")
					Expect(parser.Parse(tokens, nil)).To(Equal(
						PARSE_ERROR("unmatched left parenthesis")),
					)
				})
				Specify("unmatched right parenthesis", func() {
					tokens := tokenizer.Tokenize(")")
					Expect(parser.Parse(tokens, nil)).To(Equal(
						PARSE_ERROR("unmatched right parenthesis")),
					)
				})
				Specify("mismatched right brace", func() {
					tokens := tokenizer.Tokenize("(}")
					Expect(parser.Parse(tokens, nil)).To(Equal(
						PARSE_ERROR("mismatched right brace")),
					)
				})
				Specify("mismatched right bracket", func() {
					tokens := tokenizer.Tokenize("(]")
					Expect(parser.Parse(tokens, nil)).To(Equal(
						PARSE_ERROR("mismatched right bracket")),
					)
				})
			})
		})
		Describe("blocks", func() {
			Specify("empty block", func() {
				script := parse("{}")
				Expect(toTree(script)).To(Equal(M_SCRIPT(M_SENTENCE(M_WORD(M_BLOCK())))))
			})
			Specify("block with one word", func() {
				script := parse("{word}")
				Expect(toTree(script)).To(Equal(M_SCRIPT(
					M_SENTENCE(M_WORD(M_BLOCK(M_SENTENCE(M_WORD(M_LITERAL("word")))))),
				)))
			})
			Specify("block with two levels", func() {
				script := parse("{word1 {subword1 subword2} word2}")
				Expect(toTree(script)).To(Equal(M_SCRIPT(
					M_SENTENCE(
						M_WORD(
							M_BLOCK(
								M_SENTENCE(
									M_WORD(M_LITERAL("word1")),
									M_WORD(
										M_BLOCK(
											M_SENTENCE(
												M_WORD(M_LITERAL("subword1")),
												M_WORD(M_LITERAL("subword2")),
											),
										),
									),
									M_WORD(M_LITERAL("word2")),
								),
							),
						),
					),
				),
				))
			})
			Describe("string value", func() {
				getBlock := func(script *Script, wordIndex uint) BlockMorpheme {
					return script.Sentences[0].Words[wordIndex].Word.Morphemes[0].(BlockMorpheme)
				}
				Specify("empty", func() {
					script := parse("{}")
					block := getBlock(script, 0)
					Expect(block.Value).To(Equal(""))
				})
				Specify("one word", func() {
					script := parse("{word}")
					block := getBlock(script, 0)
					Expect(block.Value).To(Equal("word"))
				})
				Specify("two levels", func() {
					script := parse("{word1 {word2 word3} word4}")
					block := getBlock(script, 0)
					Expect(block.Value).To(Equal("word1 {word2 word3} word4"))
					subblock := getBlock(&block.Subscript, 1)
					Expect(subblock.Value).To(Equal("word2 word3"))
				})
				Specify("space preservation", func() {
					script := parse("{ word1  \nword2\t}")
					block := getBlock(script, 0)
					Expect(block.Value).To(Equal(" word1  \nword2\t"))
				})
				Specify("continuations", func() {
					script := parse("{word1 \\\n \t  word2}")
					block := getBlock(script, 0)
					Expect(block.Value).To(Equal("word1  word2"))
				})
			})
			Describe("exceptions", func() {
				Specify("unterminated block", func() {
					tokens := tokenizer.Tokenize("{")
					Expect(parser.Parse(tokens, nil)).To(Equal(
						PARSE_ERROR("unmatched left brace"),
					))
				})
				Specify("unmatched right brace", func() {
					tokens := tokenizer.Tokenize("}")
					Expect(parser.Parse(tokens, nil)).To(Equal(
						PARSE_ERROR("unmatched right brace"),
					))
				})
				Specify("mismatched right parenthesis", func() {
					tokens := tokenizer.Tokenize("{)")
					Expect(parser.Parse(tokens, nil)).To(Equal(
						PARSE_ERROR("mismatched right parenthesis"),
					))
				})
				Specify("mismatched right bracket", func() {
					tokens := tokenizer.Tokenize("{]")
					Expect(parser.Parse(tokens, nil)).To(Equal(
						PARSE_ERROR("mismatched right bracket"),
					))
				})
			})
		})
		Describe("expressions", func() {
			Specify("empty expression", func() {
				script := parse("[]")
				Expect(toTree(script)).To(Equal(M_SCRIPT(M_SENTENCE(M_WORD(M_EXPRESSION())))))
			})
			Specify("expression with one word", func() {
				script := parse("[word]")
				Expect(toTree(script)).To(Equal(M_SCRIPT(
					M_SENTENCE(M_WORD(M_EXPRESSION(M_SENTENCE(M_WORD(M_LITERAL("word")))))),
				)))
			})
			Specify("expression with two levels", func() {
				script := parse("[word1 [subword1 subword2] word2]")
				Expect(toTree(script)).To(Equal(M_SCRIPT(
					M_SENTENCE(
						M_WORD(
							M_EXPRESSION(
								M_SENTENCE(
									M_WORD(M_LITERAL("word1")),
									M_WORD(
										M_EXPRESSION(
											M_SENTENCE(
												M_WORD(M_LITERAL("subword1")),
												M_WORD(M_LITERAL("subword2")),
											),
										),
									),
									M_WORD(M_LITERAL("word2")),
								),
							),
						),
					),
				)))
			})
			Describe("exceptions", func() {
				Specify("unterminated expression", func() {
					tokens := tokenizer.Tokenize("[")
					Expect(parser.Parse(tokens, nil)).To(Equal(
						PARSE_ERROR("unmatched left bracket"),
					))
				})
				Specify("unmatched right bracket", func() {
					tokens := tokenizer.Tokenize("]")
					Expect(parser.Parse(tokens, nil)).To(Equal(
						PARSE_ERROR("unmatched right bracket"),
					))
				})
				Specify("mismatched right parenthesis", func() {
					tokens := tokenizer.Tokenize("[)")
					Expect(parser.Parse(tokens, nil)).To(Equal(
						PARSE_ERROR("mismatched right parenthesis"),
					))
				})
				Specify("mismatched right brace", func() {
					tokens := tokenizer.Tokenize("[}")
					Expect(parser.Parse(tokens, nil)).To(Equal(
						PARSE_ERROR("mismatched right brace"),
					))
				})
			})
		})
		Describe("strings", func() {
			Specify("empty string", func() {
				script := parse(`""`)
				Expect(toTree(script)).To(Equal(M_SCRIPT(M_SENTENCE(M_WORD(M_STRING())))))
			})
			Specify("simple string", func() {
				script := parse(`"string"`)
				Expect(toTree(script)).To(Equal(M_SCRIPT(
					M_SENTENCE(M_WORD(M_STRING(M_LITERAL("string")))),
				)))
			})
			Specify("longer string", func() {
				script := parse(`"this is a string"`)
				Expect(toTree(script)).To(Equal(M_SCRIPT(
					M_SENTENCE(M_WORD(M_STRING(M_LITERAL("this is a string")))),
				)))
			})
			Specify("string with whitespaces and continuations", func() {
				script := parse("\"this  \t  is\r\f a   \\\n  \t  string\"")
				Expect(toTree(script)).To(Equal(M_SCRIPT(
					M_SENTENCE(M_WORD(M_STRING(M_LITERAL("this  \t  is\r\f a    string")))),
				)))
			})
			Specify("string with special characters", func() {
				script := parse(`"this {is (a #string"`)
				Expect(toTree(script)).To(Equal(M_SCRIPT(
					M_SENTENCE(M_WORD(M_STRING(M_LITERAL("this {is (a #string")))),
				)))
				Expect(toTree(parse(`"("`))).To(Equal(M_SCRIPT(
					M_SENTENCE(M_WORD(M_STRING(M_LITERAL("(")))),
				)))
				Expect(toTree(parse(`"{"`))).To(Equal(M_SCRIPT(
					M_SENTENCE(M_WORD(M_STRING(M_LITERAL("{")))),
				)))
			})
			Describe("expressions", func() {
				Specify("empty expression", func() {
					script := parse(`"[]"`)
					Expect(toTree(script)).To(Equal(M_SCRIPT(M_SENTENCE(M_WORD(M_STRING(M_EXPRESSION()))))))
				})
				Specify("expression with one word", func() {
					script := parse(`"[word]"`)
					Expect(toTree(script)).To(Equal(M_SCRIPT(
						M_SENTENCE(M_WORD(M_STRING(M_EXPRESSION(M_SENTENCE(M_WORD(M_LITERAL("word"))))))),
					)))
				})
				Specify("expression with two levels", func() {
					script := parse(`"[word1 [subword1 subword2] word2]"`)
					Expect(toTree(script)).To(Equal(M_SCRIPT(
						M_SENTENCE(
							M_WORD(
								M_STRING(
									M_EXPRESSION(
										M_SENTENCE(
											M_WORD(M_LITERAL("word1")),
											M_WORD(
												M_EXPRESSION(
													M_SENTENCE(
														M_WORD(M_LITERAL("subword1")),
														M_WORD(M_LITERAL("subword2")),
													),
												),
											),
											M_WORD(M_LITERAL("word2")),
										),
									),
								),
							),
						),
					),
					))
				})
				Describe("exceptions", func() {
					Specify("unterminated expression", func() {
						tokens := tokenizer.Tokenize(`"[`)
						Expect(parser.Parse(tokens, nil)).To(Equal(
							PARSE_ERROR("unmatched left bracket"),
						))
					})
					Specify("mismatched right parenthesis", func() {
						tokens := tokenizer.Tokenize(`"[)"`)
						Expect(parser.Parse(tokens, nil)).To(Equal(
							PARSE_ERROR("mismatched right parenthesis"),
						))
					})
					Specify("mismatched right brace", func() {
						tokens := tokenizer.Tokenize(`"[}"`)
						Expect(parser.Parse(tokens, nil)).To(Equal(
							PARSE_ERROR("mismatched right brace"),
						))
					})
				})
			})
			Describe("substitutions", func() {
				Specify("lone dollar", func() {
					script := parse(`"$"`)
					Expect(toTree(script)).To(Equal(M_SCRIPT(M_SENTENCE(M_WORD(M_STRING(M_LITERAL("$")))))))
				})
				Specify("simple variable", func() {
					script := parse(`"$a"`)
					Expect(toTree(script)).To(Equal(M_SCRIPT(
						M_SENTENCE(M_WORD(M_STRING(M_SUBSTITUTE_NEXT("$"), M_LITERAL("a")))),
					)))
				})
				Specify("Unicode variable name", func() {
					script := parse("\"$a\u1234\"")
					Expect(toTree(script)).To(Equal(M_SCRIPT(
						M_SENTENCE(M_WORD(M_STRING(M_SUBSTITUTE_NEXT("$"), M_LITERAL("a\u1234")))),
					)))
				})
				Specify("block", func() {
					script := parse(`"${a}"`)
					Expect(toTree(script)).To(Equal(M_SCRIPT(
						M_SENTENCE(
							M_WORD(
								M_STRING(
									M_SUBSTITUTE_NEXT("$"),
									M_BLOCK(M_SENTENCE(M_WORD(M_LITERAL("a")))),
								),
							),
						),
					)))
				})
				Specify("expression", func() {
					script := parse(`"$[a]"`)
					Expect(toTree(script)).To(Equal(M_SCRIPT(
						M_SENTENCE(
							M_WORD(
								M_STRING(
									M_SUBSTITUTE_NEXT("$"),
									M_EXPRESSION(M_SENTENCE(M_WORD(M_LITERAL("a")))),
								),
							),
						),
					)))
				})
				Specify("multiple substitution", func() {
					script := parse(`"$$a $$$b $$$$[c]"`)
					Expect(toTree(script)).To(Equal(M_SCRIPT(
						M_SENTENCE(
							M_WORD(
								M_STRING(
									M_SUBSTITUTE_NEXT("$"),
									M_SUBSTITUTE_NEXT("$"),
									M_LITERAL("a"),
									M_LITERAL(" "),
									M_SUBSTITUTE_NEXT("$"),
									M_SUBSTITUTE_NEXT("$"),
									M_SUBSTITUTE_NEXT("$"),
									M_LITERAL("b"),
									M_LITERAL(" "),
									M_SUBSTITUTE_NEXT("$"),
									M_SUBSTITUTE_NEXT("$"),
									M_SUBSTITUTE_NEXT("$"),
									M_SUBSTITUTE_NEXT("$"),
									M_EXPRESSION(M_SENTENCE(M_WORD(M_LITERAL("c")))),
								),
							),
						),
					)))
				})
				Specify("expansion", func() {
					script := parse(`"$*$$*a $*$[b]"`)
					Expect(toTree(script)).To(Equal(M_SCRIPT(
						M_SENTENCE(
							M_WORD(
								M_STRING(
									M_EXPAND_NEXT("$*"),
									M_SUBSTITUTE_NEXT("$"),
									M_SUBSTITUTE_NEXT("$*"),
									M_LITERAL("a"),
									M_LITERAL(" "),
									M_EXPAND_NEXT("$*"),
									M_SUBSTITUTE_NEXT("$"),
									M_EXPRESSION(M_SENTENCE(M_WORD(M_LITERAL("b")))),
								),
							),
						),
					)))
				})
				Describe("variable name delimiters", func() {
					Specify("trailing dollars", func() {
						script := parse(`"a$ b$*$ c$$*$"`)
						Expect(toTree(script)).To(Equal(M_SCRIPT(
							M_SENTENCE(M_WORD(M_STRING(M_LITERAL("a$ b$*$ c$$*$")))),
						)))
					})
					Specify("escapes", func() {
						script := parse("\"$a\\x62 $c\\d\"")
						Expect(toTree(script)).To(Equal(M_SCRIPT(
							M_SENTENCE(
								M_WORD(
									M_STRING(
										M_SUBSTITUTE_NEXT("$"),
										M_LITERAL("a"),
										M_LITERAL("b "),
										M_SUBSTITUTE_NEXT("$"),
										M_LITERAL("c"),
										M_LITERAL("d"),
									),
								),
							),
						)))
					})
					Specify("parentheses", func() {
						script := parse(`"$(a"`)
						Expect(toTree(script)).To(Equal(M_SCRIPT(
							M_SENTENCE(
								M_WORD(
									M_STRING(M_LITERAL("$(a")),
								),
							),
						)))
					})
					Specify("special characters", func() {
						script := parse("$a# $b*")
						Expect(toTree(script)).To(Equal(M_SCRIPT(
							M_SENTENCE(
								M_WORD(M_SUBSTITUTE_NEXT("$"), M_LITERAL("a"), M_LITERAL("#")),
								M_WORD(M_SUBSTITUTE_NEXT("$"), M_LITERAL("b"), M_LITERAL("*")),
							),
						)))
					})
				})
				Describe("selectors", func() {
					Describe("indexed selectors", func() {
						Specify("single", func() {
							script := parse(`"$name[index1] $[expression][index2]"`)
							Expect(toTree(script)).To(Equal(M_SCRIPT(
								M_SENTENCE(
									M_WORD(
										M_STRING(
											M_SUBSTITUTE_NEXT("$"),
											M_LITERAL("name"),
											M_EXPRESSION(M_SENTENCE(M_WORD(M_LITERAL("index1")))),
											M_LITERAL(" "),
											M_SUBSTITUTE_NEXT("$"),
											M_EXPRESSION(M_SENTENCE(M_WORD(M_LITERAL("expression")))),
											M_EXPRESSION(M_SENTENCE(M_WORD(M_LITERAL("index2")))),
										),
									),
								),
							)))
						})
						Specify("chained", func() {
							script := parse(
								`"$name[index1][index2][index3] $[expression][index4][index5][index6]"`,
							)
							Expect(toTree(script)).To(Equal(M_SCRIPT(
								M_SENTENCE(
									M_WORD(
										M_STRING(
											M_SUBSTITUTE_NEXT("$"),
											M_LITERAL("name"),
											M_EXPRESSION(M_SENTENCE(M_WORD(M_LITERAL("index1")))),
											M_EXPRESSION(M_SENTENCE(M_WORD(M_LITERAL("index2")))),
											M_EXPRESSION(M_SENTENCE(M_WORD(M_LITERAL("index3")))),
											M_LITERAL(" "),
											M_SUBSTITUTE_NEXT("$"),
											M_EXPRESSION(M_SENTENCE(M_WORD(M_LITERAL("expression")))),
											M_EXPRESSION(M_SENTENCE(M_WORD(M_LITERAL("index4")))),
											M_EXPRESSION(M_SENTENCE(M_WORD(M_LITERAL("index5")))),
											M_EXPRESSION(M_SENTENCE(M_WORD(M_LITERAL("index6")))),
										),
									),
								),
							)))
						})
					})
					Describe("keyed selectors", func() {
						Specify("single", func() {
							script := parse(`"$name(key1) $[expression](key2)"`)
							Expect(toTree(script)).To(Equal(M_SCRIPT(
								M_SENTENCE(
									M_WORD(
										M_STRING(
											M_SUBSTITUTE_NEXT("$"),
											M_LITERAL("name"),
											M_TUPLE(M_SENTENCE(M_WORD(M_LITERAL("key1")))),
											M_LITERAL(" "),
											M_SUBSTITUTE_NEXT("$"),
											M_EXPRESSION(M_SENTENCE(M_WORD(M_LITERAL("expression")))),
											M_TUPLE(M_SENTENCE(M_WORD(M_LITERAL("key2")))),
										),
									),
								),
							)))
						})
						Specify("multiple", func() {
							script := parse(
								`"$name(key1 key2) $[expression](key3 key4)"`,
							)
							Expect(toTree(script)).To(Equal(M_SCRIPT(
								M_SENTENCE(
									M_WORD(
										M_STRING(
											M_SUBSTITUTE_NEXT("$"),
											M_LITERAL("name"),
											M_TUPLE(
												M_SENTENCE(M_WORD(M_LITERAL("key1")), M_WORD(M_LITERAL("key2"))),
											),
											M_LITERAL(" "),
											M_SUBSTITUTE_NEXT("$"),
											M_EXPRESSION(M_SENTENCE(M_WORD(M_LITERAL("expression")))),
											M_TUPLE(
												M_SENTENCE(M_WORD(M_LITERAL("key3")), M_WORD(M_LITERAL("key4"))),
											),
										),
									),
								),
							)))
						})
						Specify("chained", func() {
							script := parse(
								`"$name(key1)(key2 key3)(key4) $[expression](key5 key6)(key7)(key8 key9)"`,
							)
							Expect(toTree(script)).To(Equal(M_SCRIPT(
								M_SENTENCE(
									M_WORD(
										M_STRING(
											M_SUBSTITUTE_NEXT("$"),
											M_LITERAL("name"),
											M_TUPLE(M_SENTENCE(M_WORD(M_LITERAL("key1")))),
											M_TUPLE(
												M_SENTENCE(M_WORD(M_LITERAL("key2")), M_WORD(M_LITERAL("key3"))),
											),
											M_TUPLE(M_SENTENCE(M_WORD(M_LITERAL("key4")))),
											M_LITERAL(" "),
											M_SUBSTITUTE_NEXT("$"),
											M_EXPRESSION(M_SENTENCE(M_WORD(M_LITERAL("expression")))),
											M_TUPLE(
												M_SENTENCE(M_WORD(M_LITERAL("key5")), M_WORD(M_LITERAL("key6"))),
											),
											M_TUPLE(M_SENTENCE(M_WORD(M_LITERAL("key7")))),
											M_TUPLE(
												M_SENTENCE(M_WORD(M_LITERAL("key8")), M_WORD(M_LITERAL("key9"))),
											),
										),
									),
								),
							)))
						})
					})
					Describe("generic selectors", func() {
						Specify("single", func() {
							script := parse(
								`"$name{selector1 arg1} $[expression]{selector2 arg2}"`,
							)
							Expect(toTree(script)).To(Equal(M_SCRIPT(
								M_SENTENCE(
									M_WORD(
										M_STRING(
											M_SUBSTITUTE_NEXT("$"),
											M_LITERAL("name"),
											M_BLOCK(
												M_SENTENCE(M_WORD(M_LITERAL("selector1")), M_WORD(M_LITERAL("arg1"))),
											),
											M_LITERAL(" "),
											M_SUBSTITUTE_NEXT("$"),
											M_EXPRESSION(M_SENTENCE(M_WORD(M_LITERAL("expression")))),
											M_BLOCK(
												M_SENTENCE(M_WORD(M_LITERAL("selector2")), M_WORD(M_LITERAL("arg2"))),
											),
										),
									),
								),
							)))
						})
						Specify("chained", func() {
							script := parse(
								`"$name{selector1 arg1}{selector2}{selector3 arg2 arg3} $[expression]{selector4}{selector5 arg4 arg5}{selector6 arg6}"`,
							)
							Expect(toTree(script)).To(Equal(M_SCRIPT(
								M_SENTENCE(
									M_WORD(
										M_STRING(
											M_SUBSTITUTE_NEXT("$"),
											M_LITERAL("name"),
											M_BLOCK(
												M_SENTENCE(M_WORD(M_LITERAL("selector1")), M_WORD(M_LITERAL("arg1"))),
											),
											M_BLOCK(M_SENTENCE(M_WORD(M_LITERAL("selector2")))),
											M_BLOCK(
												M_SENTENCE(
													M_WORD(M_LITERAL("selector3")),
													M_WORD(M_LITERAL("arg2")),
													M_WORD(M_LITERAL("arg3")),
												),
											),
											M_LITERAL(" "),
											M_SUBSTITUTE_NEXT("$"),
											M_EXPRESSION(M_SENTENCE(M_WORD(M_LITERAL("expression")))),
											M_BLOCK(M_SENTENCE(M_WORD(M_LITERAL("selector4")))),
											M_BLOCK(
												M_SENTENCE(
													M_WORD(M_LITERAL("selector5")),
													M_WORD(M_LITERAL("arg4")),
													M_WORD(M_LITERAL("arg5")),
												),
											),
											M_BLOCK(
												M_SENTENCE(M_WORD(M_LITERAL("selector6")), M_WORD(M_LITERAL("arg6"))),
											),
										),
									),
								),
							)))
						})
					})
					Specify("mixed selectors", func() {
						script := parse(
							`"$name(key1 key2){selector1}(key3){selector2 selector3} $[expression]{selector4 selector5}(key4 key5){selector6}(key6)"`,
						)
						Expect(toTree(script)).To(Equal(M_SCRIPT(
							M_SENTENCE(
								M_WORD(
									M_STRING(
										M_SUBSTITUTE_NEXT("$"),
										M_LITERAL("name"),
										M_TUPLE(M_SENTENCE(M_WORD(M_LITERAL("key1")), M_WORD(M_LITERAL("key2")))),
										M_BLOCK(M_SENTENCE(M_WORD(M_LITERAL("selector1")))),
										M_TUPLE(M_SENTENCE(M_WORD(M_LITERAL("key3")))),
										M_BLOCK(
											M_SENTENCE(
												M_WORD(M_LITERAL("selector2")),
												M_WORD(M_LITERAL("selector3")),
											),
										),
										M_LITERAL(" "),
										M_SUBSTITUTE_NEXT("$"),
										M_EXPRESSION(M_SENTENCE(M_WORD(M_LITERAL("expression")))),
										M_BLOCK(
											M_SENTENCE(
												M_WORD(M_LITERAL("selector4")),
												M_WORD(M_LITERAL("selector5")),
											),
										),
										M_TUPLE(M_SENTENCE(M_WORD(M_LITERAL("key4")), M_WORD(M_LITERAL("key5")))),
										M_BLOCK(M_SENTENCE(M_WORD(M_LITERAL("selector6")))),
										M_TUPLE(M_SENTENCE(M_WORD(M_LITERAL("key6")))),
									),
								),
							),
						)))
					})
					Specify("nested selectors", func() {
						script := parse(
							`"$name1(key1 $name2{selector1} $[expression1](key2)) $[expression2]{selector2 $name3(key3)}"`,
						)
						Expect(toTree(script)).To(Equal(M_SCRIPT(
							M_SENTENCE(
								M_WORD(
									M_STRING(
										M_SUBSTITUTE_NEXT("$"),
										M_LITERAL("name1"),
										M_TUPLE(
											M_SENTENCE(
												M_WORD(M_LITERAL("key1")),
												M_WORD(
													M_SUBSTITUTE_NEXT("$"),
													M_LITERAL("name2"),
													M_BLOCK(M_SENTENCE(M_WORD(M_LITERAL("selector1")))),
												),
												M_WORD(
													M_SUBSTITUTE_NEXT("$"),
													M_EXPRESSION(M_SENTENCE(M_WORD(M_LITERAL("expression1")))),
													M_TUPLE(M_SENTENCE(M_WORD(M_LITERAL("key2")))),
												),
											),
										),
										M_LITERAL(" "),
										M_SUBSTITUTE_NEXT("$"),
										M_EXPRESSION(M_SENTENCE(M_WORD(M_LITERAL("expression2")))),
										M_BLOCK(
											M_SENTENCE(
												M_WORD(M_LITERAL("selector2")),
												M_WORD(
													M_SUBSTITUTE_NEXT("$"),
													M_LITERAL("name3"),
													M_TUPLE(M_SENTENCE(M_WORD(M_LITERAL("key3")))),
												),
											),
										),
									),
								),
							),
						)))
					})
				})
			})
			Describe("exceptions", func() {
				Specify("unterminated string", func() {
					tokens := tokenizer.Tokenize(`"`)
					Expect(parser.Parse(tokens, nil)).To(Equal(
						PARSE_ERROR("unmatched string delimiter"),
					))
				})
				Specify("extra quotes", func() {
					tokens := tokenizer.Tokenize(`"hello""`)
					Expect(parser.Parse(tokens, nil)).To(Equal(
						PARSE_ERROR("extra characters after string delimiter"),
					))
				})
			})
		})
		Describe("here-strings", func() {
			Specify("3-quote delimiter", func() {
				script := parse(
					"\"\"\"some \" \\\n    $arbitrary [character\n  \"\" sequence\"\"\"",
				)
				Expect(toTree(script)).To(Equal(M_SCRIPT(
					M_SENTENCE(
						M_WORD(
							M_HERE_STRING(
								"some \" \\\n    $arbitrary [character\n  \"\" sequence",
							)),
					),
				)))
			})
			Specify("4-quote delimiter", func() {
				script := parse(`""""here is """ some text""""`)
				Expect(toTree(script)).To(Equal(M_SCRIPT(
					M_SENTENCE(
						M_WORD(
							M_HERE_STRING(`here is """ some text`),
						),
					),
				)))
			})
			Specify("4-quote sequence between 3-quote delimiters", func() {
				script := parse(
					`""" <- 3 quotes here / 4 quotes there -> """" / 3 quotes here -> """`,
				)
				Expect(toTree(script)).To(Equal(M_SCRIPT(
					M_SENTENCE(
						M_WORD(
							M_HERE_STRING(
								` <- 3 quotes here / 4 quotes there -> """" / 3 quotes here -> `,
							),
						),
					),
				)))
			})
			Describe("exceptions", func() {
				Specify("unterminated here-string", func() {
					tokens := tokenizer.Tokenize(`"""hello`)
					Expect(parser.Parse(tokens, nil)).To(Equal(
						PARSE_ERROR("unmatched here-string delimiter"),
					))
				})
				Specify("extra quotes", func() {
					tokens := tokenizer.Tokenize(
						`""" <- 3 quotes here / 4 quotes there -> """"`,
					)
					Expect(parser.Parse(tokens, nil)).To(Equal(
						PARSE_ERROR("unmatched here-string delimiter"),
					))
				})
			})
		})
		Describe("tagged strings", func() {
			Specify("empty tagged string", func() {
				script := parse("\"\"EOF\nEOF\"\"")
				Expect(toTree(script)).To(Equal(M_SCRIPT(M_SENTENCE(M_WORD(M_TAGGED_STRING(""))))))
			})
			Specify("single empty line", func() {
				script := parse("\"\"EOF\n\nEOF\"\"")
				Expect(toTree(script)).To(Equal(M_SCRIPT(M_SENTENCE(M_WORD(M_TAGGED_STRING("\n"))))))
			})
			Specify("extra characters after open delimiter", func() {
				script := parse(
					"\"\"EOF some $arbitrary[ }text\\\n (with continuation\nfoo\nbar\nEOF\"\"",
				)
				Expect(toTree(script)).To(Equal(M_SCRIPT(M_SENTENCE(M_WORD(M_TAGGED_STRING("foo\nbar\n"))))))
			})
			Specify("tag within string", func() {
				script := parse("\"\"EOF\nEOF \"\"\nEOF\"\"")
				Expect(toTree(script)).To(Equal(M_SCRIPT(M_SENTENCE(M_WORD(M_TAGGED_STRING("EOF \"\"\n"))))))
			})
			Specify("continuations", func() {
				script := parse("\"\"EOF\nsome\\\n   string\nEOF\"\"")
				Expect(toTree(script)).To(Equal(M_SCRIPT(
					M_SENTENCE(M_WORD(M_TAGGED_STRING("some\\\n   string\n"))),
				)))
			})
			Specify("indentation", func() {
				script := parse(`""EOF
			          #include <stdio.h>

			          int main(void) {
			            printf("Hello, world!");
			            return 0;
			          }
			          EOF""`)
				Expect(toTree(script)).To(Equal(M_SCRIPT(
					M_SENTENCE(
						M_WORD(
							M_TAGGED_STRING(`#include <stdio.h>

int main(void) {
  printf("Hello, world!");
  return 0;
}
`),
						),
					),
				)))
			})
			Specify("line prefix", func() {
				script := parse(`""EOF
			1  #include <stdio.h>
			2
			3  int main(void) {
			4    printf("Hello, world!");
			5    return 0;
			6  }
			   EOF""`)
				Expect(toTree(script)).To(Equal(M_SCRIPT(
					M_SENTENCE(
						M_WORD(M_TAGGED_STRING(`#include <stdio.h>

int main(void) {
  printf("Hello, world!");
  return 0;
}
`),
						),
					),
				)))
			})
			Specify("prefix with shorter lines", func() {
				script := parse(`""TAG
			          $ prompt

			          > result
			          > TAG""`)
				Expect(toTree(script)).To(Equal(M_SCRIPT(
					M_SENTENCE(M_WORD(M_TAGGED_STRING("prompt\n\nresult\n"))),
				)))
			})
			Describe("exceptions", func() {
				Specify("unterminated tagged string", func() {
					tokens := tokenizer.Tokenize("\"\"EOF\nhello")
					Expect(parser.Parse(tokens, nil)).To(Equal(
						PARSE_ERROR("unmatched tagged string delimiter"),
					))
				})
				Specify("extra quotes", func() {
					tokens := tokenizer.Tokenize("\"\"EOF\nhello\nEOF\"\"\"")
					Expect(parser.Parse(tokens, nil)).To(Equal(
						PARSE_ERROR("unmatched tagged string delimiter"),
					))
				})
			})
		})
		Describe("line comments", func() {
			Specify("empty line comment", func() {
				script := parse("#")
				Expect(toTree(script)).To(Equal(M_SCRIPT(M_SENTENCE(M_WORD(M_LINE_COMMENT(""))))))
			})
			Specify("simple line comment", func() {
				script := parse("# this is a comment")
				Expect(toTree(script)).To(Equal(M_SCRIPT(
					M_SENTENCE(M_WORD(M_LINE_COMMENT(" this is a comment"))),
				)))
			})
			Specify("line comment with special characters", func() {
				script := parse("# this ; is$ (a [comment{")
				Expect(toTree(script)).To(Equal(M_SCRIPT(
					M_SENTENCE(M_WORD(M_LINE_COMMENT(" this ; is$ (a [comment{"))),
				)))
			})
			Specify("line comment with continuation", func() {
				script := parse("# this is\\\na comment")
				Expect(toTree(script)).To(Equal(M_SCRIPT(
					M_SENTENCE(M_WORD(M_LINE_COMMENT(" this is a comment"))),
				)))
			})
			Specify("line comment with escapes", func() {
				script := parse("# hello \\x41\\t")
				Expect(toTree(script)).To(Equal(M_SCRIPT(
					M_SENTENCE(M_WORD(M_LINE_COMMENT(" hello A\t"))),
				)))
			})
			Specify("line comment with multiple hashes", func() {
				script := parse("### this is a comment")
				Expect(toTree(script)).To(Equal(M_SCRIPT(
					M_SENTENCE(M_WORD(M_LINE_COMMENT(" this is a comment"))),
				)))
			})
		})
		Describe("block comments", func() {
			Specify("empty block comment", func() {
				script := parse("#{}#")
				Expect(toTree(script)).To(Equal(M_SCRIPT(M_SENTENCE(M_WORD(M_BLOCK_COMMENT(""))))))
			})
			Specify("simple block comment", func() {
				script := parse("#{comment}#")
				Expect(toTree(script)).To(Equal(M_SCRIPT(M_SENTENCE(M_WORD(M_BLOCK_COMMENT("comment"))))))
			})
			Specify("multiple line block comment", func() {
				script := parse("#{\ncomment\n}#")
				Expect(toTree(script)).To(Equal(M_SCRIPT(M_SENTENCE(M_WORD(M_BLOCK_COMMENT("\ncomment\n"))))))
			})
			Specify("block comment with continuation", func() {
				script := parse("#{this is\\\na comment}#")
				Expect(toTree(script)).To(Equal(M_SCRIPT(
					M_SENTENCE(M_WORD(M_BLOCK_COMMENT("this is\\\na comment"))),
				)))
			})
			Specify("block comment with escapes", func() {
				script := parse("#{hello \\x41\\t}#")
				Expect(toTree(script)).To(Equal(M_SCRIPT(
					M_SENTENCE(M_WORD(M_BLOCK_COMMENT("hello \\x41\\t"))),
				)))
			})
			Specify("block comment with multiple hashes", func() {
				script := parse("##{comment}##")
				Expect(toTree(script)).To(Equal(M_SCRIPT(M_SENTENCE(M_WORD(M_BLOCK_COMMENT("comment"))))))
			})
			Specify("nested block comments", func() {
				script := parse("##{comment ##{}##}##")
				Expect(toTree(script)).To(Equal(M_SCRIPT(
					M_SENTENCE(M_WORD(M_BLOCK_COMMENT("comment ##{}##"))),
				)))
			})
			Specify("nested block comments with different prefixes", func() {
				script := parse("##{comment #{}##")
				Expect(toTree(script)).To(Equal(M_SCRIPT(M_SENTENCE(M_WORD(M_BLOCK_COMMENT("comment #{"))))))
			})
			Describe("exceptions", func() {
				Specify("unterminated block comment", func() {
					tokens := tokenizer.Tokenize("#{hello")
					Expect(parser.Parse(tokens, nil)).To(Equal(
						PARSE_ERROR("unmatched block comment delimiter"),
					))
				})
				Specify("extra hashes", func() {
					tokens := tokenizer.Tokenize(
						"#{ <- 1 hash here / 2 hashes there -> }##",
					)
					Expect(parser.Parse(tokens, nil)).To(Equal(
						PARSE_ERROR("unmatched block comment delimiter"),
					))
				})
			})
		})
		Describe("substitutions", func() {
			Specify("lone dollar", func() {
				script := parse("$")
				Expect(toTree(script)).To(Equal(M_SCRIPT(M_SENTENCE(M_WORD(M_LITERAL("$"))))))
			})
			Specify("simple variable", func() {
				script := parse("$a")
				Expect(toTree(script)).To(Equal(M_SCRIPT(
					M_SENTENCE(M_WORD(M_SUBSTITUTE_NEXT("$"), M_LITERAL("a"))),
				)))
			})
			Specify("Unicode variable name", func() {
				script := parse("$a\u1234")
				Expect(toTree(script)).To(Equal(M_SCRIPT(
					M_SENTENCE(M_WORD(M_SUBSTITUTE_NEXT("$"), M_LITERAL("a\u1234"))),
				)))
			})
			Specify("tuple", func() {
				script := parse("$(a)")
				Expect(toTree(script)).To(Equal(M_SCRIPT(
					M_SENTENCE(M_WORD(M_SUBSTITUTE_NEXT("$"), M_TUPLE(M_SENTENCE(M_WORD(M_LITERAL("a")))))),
				)))
			})
			Specify("block", func() {
				script := parse("${a}")
				Expect(toTree(script)).To(Equal(M_SCRIPT(
					M_SENTENCE(M_WORD(M_SUBSTITUTE_NEXT("$"), M_BLOCK(M_SENTENCE(M_WORD(M_LITERAL("a")))))),
				)))
			})
			Specify("expression", func() {
				script := parse("$[a]")
				Expect(toTree(script)).To(Equal(M_SCRIPT(
					M_SENTENCE(M_WORD(M_SUBSTITUTE_NEXT("$"), M_EXPRESSION(M_SENTENCE(M_WORD(M_LITERAL("a")))))),
				)))
			})
			Specify("multiple substitution", func() {
				script := parse("$$a $$$b $$$$[c]")
				Expect(toTree(script)).To(Equal(M_SCRIPT(
					M_SENTENCE(
						M_WORD(M_SUBSTITUTE_NEXT("$"),
							M_SUBSTITUTE_NEXT("$"), M_LITERAL("a")),
						M_WORD(M_SUBSTITUTE_NEXT("$"),
							M_SUBSTITUTE_NEXT("$"),
							M_SUBSTITUTE_NEXT("$"), M_LITERAL("b")),
						M_WORD(
							M_SUBSTITUTE_NEXT("$"),
							M_SUBSTITUTE_NEXT("$"),
							M_SUBSTITUTE_NEXT("$"),
							M_SUBSTITUTE_NEXT("$"),
							M_EXPRESSION(M_SENTENCE(M_WORD(M_LITERAL("c")))),
						),
					),
				)))
			})
			Specify("expansion", func() {
				script := parse("$*$$*a $*$[b]")
				Expect(toTree(script)).To(Equal(M_SCRIPT(
					M_SENTENCE(
						M_WORD(
							M_EXPAND_NEXT("$*"),
							M_SUBSTITUTE_NEXT("$"),
							M_SUBSTITUTE_NEXT("$*"),
							M_LITERAL("a"),
						),
						M_WORD(
							M_EXPAND_NEXT("$*"),
							M_SUBSTITUTE_NEXT("$"),
							M_EXPRESSION(M_SENTENCE(M_WORD(M_LITERAL("b")))),
						),
					),
				)))
			})
			Describe("variable name delimiters", func() {
				Specify("trailing dollars", func() {
					script := parse("a$ b$*$ c$$*$")
					Expect(toTree(script)).To(Equal(M_SCRIPT(
						M_SENTENCE(
							M_WORD(M_LITERAL("a$")),
							M_WORD(M_LITERAL("b$*$")),
							M_WORD(M_LITERAL("c$$*$")),
						),
					)))
				})
				Specify("escapes", func() {
					script := parse("$a\\x62 $c\\d")
					Expect(toTree(script)).To(Equal(M_SCRIPT(
						M_SENTENCE(
							M_WORD(M_SUBSTITUTE_NEXT("$"), M_LITERAL("a"), M_LITERAL("b")),
							M_WORD(M_SUBSTITUTE_NEXT("$"), M_LITERAL("c"), M_LITERAL("d")),
						),
					)))
				})
				Specify("special characters", func() {
					script := parse("$a# $b*")
					Expect(toTree(script)).To(Equal(M_SCRIPT(
						M_SENTENCE(
							M_WORD(M_SUBSTITUTE_NEXT("$"), M_LITERAL("a"), M_LITERAL("#")),
							M_WORD(M_SUBSTITUTE_NEXT("$"), M_LITERAL("b"), M_LITERAL("*")),
						),
					)))
				})
				Describe("exceptions", func() {
					Specify("leading hash", func() {
						tokens := tokenizer.Tokenize("$#")
						Expect(parser.Parse(tokens, nil)).To(Equal(
							PARSE_ERROR("unexpected comment delimiter"),
						))
					})
					Specify("leading quote", func() {
						tokens := tokenizer.Tokenize(`$"`)
						Expect(parser.Parse(tokens, nil)).To(Equal(
							PARSE_ERROR("unexpected string delimiter"),
						))
					})
				})
			})
			Describe("selectors", func() {
				Describe("indexed selectors", func() {
					Specify("single", func() {
						script := parse("$name[index1] $[expression][index2]")
						Expect(toTree(script)).To(Equal(M_SCRIPT(
							M_SENTENCE(
								M_WORD(
									M_SUBSTITUTE_NEXT("$"),
									M_LITERAL("name"),
									M_EXPRESSION(M_SENTENCE(M_WORD(M_LITERAL("index1")))),
								),
								M_WORD(
									M_SUBSTITUTE_NEXT("$"),
									M_EXPRESSION(M_SENTENCE(M_WORD(M_LITERAL("expression")))),
									M_EXPRESSION(M_SENTENCE(M_WORD(M_LITERAL("index2")))),
								),
							),
						)))
					})
					Specify("chained", func() {
						script := parse(
							"$name[index1][index2][index3] $[expression][index4][index5][index6]",
						)
						Expect(toTree(script)).To(Equal(M_SCRIPT(
							M_SENTENCE(
								M_WORD(
									M_SUBSTITUTE_NEXT("$"),
									M_LITERAL("name"),
									M_EXPRESSION(M_SENTENCE(M_WORD(M_LITERAL("index1")))),
									M_EXPRESSION(M_SENTENCE(M_WORD(M_LITERAL("index2")))),
									M_EXPRESSION(M_SENTENCE(M_WORD(M_LITERAL("index3")))),
								),
								M_WORD(
									M_SUBSTITUTE_NEXT("$"),
									M_EXPRESSION(M_SENTENCE(M_WORD(M_LITERAL("expression")))),
									M_EXPRESSION(M_SENTENCE(M_WORD(M_LITERAL("index4")))),
									M_EXPRESSION(M_SENTENCE(M_WORD(M_LITERAL("index5")))),
									M_EXPRESSION(M_SENTENCE(M_WORD(M_LITERAL("index6")))),
								),
							),
						)))
					})
				})
				Describe("keyed selectors", func() {
					Specify("single", func() {
						script := parse("$name(key1) $[expression](key2)")
						Expect(toTree(script)).To(Equal(M_SCRIPT(
							M_SENTENCE(
								M_WORD(
									M_SUBSTITUTE_NEXT("$"),
									M_LITERAL("name"),
									M_TUPLE(M_SENTENCE(M_WORD(M_LITERAL("key1")))),
								),
								M_WORD(
									M_SUBSTITUTE_NEXT("$"),
									M_EXPRESSION(M_SENTENCE(M_WORD(M_LITERAL("expression")))),
									M_TUPLE(M_SENTENCE(M_WORD(M_LITERAL("key2")))),
								),
							),
						)))
					})
					Specify("multiple", func() {
						script := parse("$name(key1 key2) $[expression](key3 key4)")
						Expect(toTree(script)).To(Equal(M_SCRIPT(
							M_SENTENCE(
								M_WORD(
									M_SUBSTITUTE_NEXT("$"),
									M_LITERAL("name"),
									M_TUPLE(M_SENTENCE(M_WORD(M_LITERAL("key1")), M_WORD(M_LITERAL("key2")))),
								),
								M_WORD(
									M_SUBSTITUTE_NEXT("$"),
									M_EXPRESSION(M_SENTENCE(M_WORD(M_LITERAL("expression")))),
									M_TUPLE(M_SENTENCE(M_WORD(M_LITERAL("key3")), M_WORD(M_LITERAL("key4")))),
								),
							),
						)))
					})
					Specify("chained", func() {
						script := parse(
							"$name(key1)(key2 key3)(key4) $[expression](key5 key6)(key7)(key8 key9)",
						)
						Expect(toTree(script)).To(Equal(M_SCRIPT(
							M_SENTENCE(
								M_WORD(
									M_SUBSTITUTE_NEXT("$"),
									M_LITERAL("name"),
									M_TUPLE(M_SENTENCE(M_WORD(M_LITERAL("key1")))),
									M_TUPLE(M_SENTENCE(M_WORD(M_LITERAL("key2")), M_WORD(M_LITERAL("key3")))),
									M_TUPLE(M_SENTENCE(M_WORD(M_LITERAL("key4")))),
								),
								M_WORD(
									M_SUBSTITUTE_NEXT("$"),
									M_EXPRESSION(M_SENTENCE(M_WORD(M_LITERAL("expression")))),
									M_TUPLE(M_SENTENCE(M_WORD(M_LITERAL("key5")), M_WORD(M_LITERAL("key6")))),
									M_TUPLE(M_SENTENCE(M_WORD(M_LITERAL("key7")))),
									M_TUPLE(M_SENTENCE(M_WORD(M_LITERAL("key8")), M_WORD(M_LITERAL("key9")))),
								),
							),
						)))
					})
				})
				Describe("generic selectors", func() {
					Specify("single", func() {
						script := parse("$name{selector1} $[expression]{selector2}")
						Expect(toTree(script)).To(Equal(M_SCRIPT(
							M_SENTENCE(
								M_WORD(
									M_SUBSTITUTE_NEXT("$"),
									M_LITERAL("name"),
									M_BLOCK(M_SENTENCE(M_WORD(M_LITERAL("selector1")))),
								),
								M_WORD(
									M_SUBSTITUTE_NEXT("$"),
									M_EXPRESSION(M_SENTENCE(M_WORD(M_LITERAL("expression")))),
									M_BLOCK(M_SENTENCE(M_WORD(M_LITERAL("selector2")))),
								),
							),
						)))
					})
					Specify("chained", func() {
						script := parse(
							"$name{selector1}{selector2 arg1}{selector3} $[expression]{selector4 arg2 arg3}{selector5}{selector6 arg4}",
						)
						Expect(toTree(script)).To(Equal(M_SCRIPT(
							M_SENTENCE(
								M_WORD(
									M_SUBSTITUTE_NEXT("$"),
									M_LITERAL("name"),
									M_BLOCK(M_SENTENCE(M_WORD(M_LITERAL("selector1")))),
									M_BLOCK(
										M_SENTENCE(M_WORD(M_LITERAL("selector2")), M_WORD(M_LITERAL("arg1"))),
									),
									M_BLOCK(M_SENTENCE(M_WORD(M_LITERAL("selector3")))),
								),
								M_WORD(
									M_SUBSTITUTE_NEXT("$"),
									M_EXPRESSION(M_SENTENCE(M_WORD(M_LITERAL("expression")))),
									M_BLOCK(
										M_SENTENCE(
											M_WORD(M_LITERAL("selector4")),
											M_WORD(M_LITERAL("arg2")),
											M_WORD(M_LITERAL("arg3")),
										),
									),
									M_BLOCK(M_SENTENCE(M_WORD(M_LITERAL("selector5")))),
									M_BLOCK(
										M_SENTENCE(M_WORD(M_LITERAL("selector6")), M_WORD(M_LITERAL("arg4"))),
									),
								),
							),
						)))
					})
				})
				Specify("mixed selectors", func() {
					script := parse(
						"$name(key1 key2){selector1}(key3){selector2 selector3} $[expression]{selector4 selector5}(key4 key5){selector6}(key6)",
					)
					Expect(toTree(script)).To(Equal(M_SCRIPT(
						M_SENTENCE(
							M_WORD(
								M_SUBSTITUTE_NEXT("$"),
								M_LITERAL("name"),
								M_TUPLE(M_SENTENCE(M_WORD(M_LITERAL("key1")), M_WORD(M_LITERAL("key2")))),
								M_BLOCK(M_SENTENCE(M_WORD(M_LITERAL("selector1")))),
								M_TUPLE(M_SENTENCE(M_WORD(M_LITERAL("key3")))),
								M_BLOCK(
									M_SENTENCE(M_WORD(M_LITERAL("selector2")), M_WORD(M_LITERAL("selector3"))),
								),
							),
							M_WORD(
								M_SUBSTITUTE_NEXT("$"),
								M_EXPRESSION(M_SENTENCE(M_WORD(M_LITERAL("expression")))),
								M_BLOCK(
									M_SENTENCE(M_WORD(M_LITERAL("selector4")), M_WORD(M_LITERAL("selector5"))),
								),
								M_TUPLE(M_SENTENCE(M_WORD(M_LITERAL("key4")), M_WORD(M_LITERAL("key5")))),
								M_BLOCK(M_SENTENCE(M_WORD(M_LITERAL("selector6")))),
								M_TUPLE(M_SENTENCE(M_WORD(M_LITERAL("key6")))),
							),
						),
					)))
				})
				Specify("nested selectors", func() {
					script := parse(
						"$name1(key1 $name2{selector1} $[expression1](key2)) $[expression2]{selector2 $name3(key3)}",
					)
					Expect(toTree(script)).To(Equal(M_SCRIPT(
						M_SENTENCE(
							M_WORD(
								M_SUBSTITUTE_NEXT("$"),
								M_LITERAL("name1"),
								M_TUPLE(
									M_SENTENCE(
										M_WORD(M_LITERAL("key1")),
										M_WORD(
											M_SUBSTITUTE_NEXT("$"),
											M_LITERAL("name2"),
											M_BLOCK(M_SENTENCE(M_WORD(M_LITERAL("selector1")))),
										),
										M_WORD(
											M_SUBSTITUTE_NEXT("$"),
											M_EXPRESSION(M_SENTENCE(M_WORD(M_LITERAL("expression1")))),
											M_TUPLE(M_SENTENCE(M_WORD(M_LITERAL("key2")))),
										),
									),
								),
							),
							M_WORD(
								M_SUBSTITUTE_NEXT("$"),
								M_EXPRESSION(M_SENTENCE(M_WORD(M_LITERAL("expression2")))),
								M_BLOCK(
									M_SENTENCE(
										M_WORD(M_LITERAL("selector2")),
										M_WORD(
											M_SUBSTITUTE_NEXT("$"),
											M_LITERAL("name3"),
											M_TUPLE(M_SENTENCE(M_WORD(M_LITERAL("key3")))),
										),
									),
								),
							),
						),
					)))
				})
			})
		})
		Describe("qualified words", func() {
			Describe("indexed selectors", func() {
				Specify("single", func() {
					script := parse("name[index]")
					Expect(toTree(script)).To(Equal(M_SCRIPT(
						M_SENTENCE(M_WORD(M_LITERAL("name"), M_EXPRESSION(M_SENTENCE(M_WORD(M_LITERAL("index")))))),
					)))
				})
				Specify("chained", func() {
					script := parse("name[index1][index2][index3]")
					Expect(toTree(script)).To(Equal(M_SCRIPT(
						M_SENTENCE(
							M_WORD(
								M_LITERAL("name"),
								M_EXPRESSION(M_SENTENCE(M_WORD(M_LITERAL("index1")))),
								M_EXPRESSION(M_SENTENCE(M_WORD(M_LITERAL("index2")))),
								M_EXPRESSION(M_SENTENCE(M_WORD(M_LITERAL("index3")))),
							),
						),
					)))
				})
			})
			Describe("keyed selectors", func() {
				Specify("single", func() {
					script := parse("name(key)")
					Expect(toTree(script)).To(Equal(M_SCRIPT(
						M_SENTENCE(M_WORD(M_LITERAL("name"), M_TUPLE(M_SENTENCE(M_WORD(M_LITERAL("key")))))),
					)))
				})
				Specify("multiple", func() {
					script := parse("name(key1 key2)")
					Expect(toTree(script)).To(Equal(M_SCRIPT(
						M_SENTENCE(
							M_WORD(
								M_LITERAL("name"),
								M_TUPLE(M_SENTENCE(M_WORD(M_LITERAL("key1")), M_WORD(M_LITERAL("key2")))),
							),
						),
					)))
				})
				Specify("chained", func() {
					script := parse("name(key1)(key2 key3)(key4)")
					Expect(toTree(script)).To(Equal(M_SCRIPT(
						M_SENTENCE(
							M_WORD(
								M_LITERAL("name"),
								M_TUPLE(M_SENTENCE(M_WORD(M_LITERAL("key1")))),
								M_TUPLE(M_SENTENCE(M_WORD(M_LITERAL("key2")), M_WORD(M_LITERAL("key3")))),
								M_TUPLE(M_SENTENCE(M_WORD(M_LITERAL("key4")))),
							),
						),
					)))
				})
			})
			Describe("generic selectors", func() {
				Specify("single", func() {
					script := parse("name{selector}")
					Expect(toTree(script)).To(Equal(M_SCRIPT(
						M_SENTENCE(M_WORD(M_LITERAL("name"), M_BLOCK(M_SENTENCE(M_WORD(M_LITERAL("selector")))))),
					)))
				})
				Specify("multiple", func() {
					script := parse("name{selector1 selector2}")
					Expect(toTree(script)).To(Equal(M_SCRIPT(
						M_SENTENCE(
							M_WORD(
								M_LITERAL("name"),
								M_BLOCK(
									M_SENTENCE(M_WORD(M_LITERAL("selector1")), M_WORD(M_LITERAL("selector2"))),
								),
							),
						),
					)))
				})
				Specify("chained", func() {
					script := parse(
						"name{selector1}{selector2 selector3}{selector4}",
					)
					Expect(toTree(script)).To(Equal(M_SCRIPT(
						M_SENTENCE(
							M_WORD(
								M_LITERAL("name"),
								M_BLOCK(M_SENTENCE(M_WORD(M_LITERAL("selector1")))),
								M_BLOCK(
									M_SENTENCE(M_WORD(M_LITERAL("selector2")), M_WORD(M_LITERAL("selector3"))),
								),
								M_BLOCK(M_SENTENCE(M_WORD(M_LITERAL("selector4")))),
							),
						),
					)))
				})
			})
			Specify("mixed selectors", func() {
				script := parse(
					"name(key1 key2){selector1}(key3){selector2 selector3}",
				)
				Expect(toTree(script)).To(Equal(M_SCRIPT(
					M_SENTENCE(
						M_WORD(
							M_LITERAL("name"),
							M_TUPLE(M_SENTENCE(M_WORD(M_LITERAL("key1")), M_WORD(M_LITERAL("key2")))),
							M_BLOCK(M_SENTENCE(M_WORD(M_LITERAL("selector1")))),
							M_TUPLE(M_SENTENCE(M_WORD(M_LITERAL("key3")))),
							M_BLOCK(
								M_SENTENCE(M_WORD(M_LITERAL("selector2")), M_WORD(M_LITERAL("selector3"))),
							),
						),
					),
				)))
			})
			Specify("nested selectors", func() {
				script := parse("name1(key1 name2{selector1})")
				Expect(toTree(script)).To(Equal(M_SCRIPT(
					M_SENTENCE(
						M_WORD(
							M_LITERAL("name1"),
							M_TUPLE(
								M_SENTENCE(
									M_WORD(M_LITERAL("key1")),
									M_WORD(
										M_LITERAL("name2"),
										M_BLOCK(M_SENTENCE(M_WORD(M_LITERAL("selector1")))),
									),
								),
							),
						),
					),
				)))
			})
		})
	})

	Describe("capturePositions", func() {
		BeforeEach(func() {
			parser = NewParser(&ParserOptions{CapturePositions: true})
		})
		Specify("literals", func() {
			script := parse("this is a list\nof literals")
			Expect(toTree(script)).To(Equal(M_SCRIPT_P(
				&SourcePosition{Index: 0, Line: 0, Column: 0},
				M_SENTENCE_P(
					&SourcePosition{Index: 0, Line: 0, Column: 0},
					M_WORD_P(
						&SourcePosition{Index: 0, Line: 0, Column: 0},
						M_LITERAL_P(
							&SourcePosition{Index: 0, Line: 0, Column: 0},
							"this",
						),
					),
					M_WORD_P(
						&SourcePosition{Index: 5, Line: 0, Column: 5},
						M_LITERAL_P(
							&SourcePosition{Index: 5, Line: 0, Column: 5},
							"is",
						),
					),
					M_WORD_P(
						&SourcePosition{Index: 8, Line: 0, Column: 8},
						M_LITERAL_P(
							&SourcePosition{Index: 8, Line: 0, Column: 8},
							"a",
						),
					),
					M_WORD_P(
						&SourcePosition{Index: 10, Line: 0, Column: 10},
						M_LITERAL_P(
							&SourcePosition{Index: 10, Line: 0, Column: 10},
							"list",
						),
					),
				),
				M_SENTENCE_P(
					&SourcePosition{Index: 15, Line: 1, Column: 0},
					M_WORD_P(
						&SourcePosition{Index: 15, Line: 1, Column: 0},
						M_LITERAL_P(
							&SourcePosition{Index: 15, Line: 1, Column: 0},
							"of",
						),
					),
					M_WORD_P(
						&SourcePosition{Index: 18, Line: 1, Column: 3},
						M_LITERAL_P(
							&SourcePosition{Index: 18, Line: 1, Column: 3},
							"literals",
						),
					),
				),
			)))
		})
		Specify("tuples", func() {
			script := parse("( ; ) ( ( \n ) ())")
			Expect(toTree(script)).To(Equal(M_SCRIPT_P(
				&SourcePosition{Index: 0, Line: 0, Column: 0},
				M_SENTENCE_P(
					&SourcePosition{Index: 0, Line: 0, Column: 0},
					M_WORD_P(
						&SourcePosition{Index: 0, Line: 0, Column: 0},
						M_TUPLE_P(
							&SourcePosition{Index: 0, Line: 0, Column: 0},
							M_SCRIPT_P(
								&SourcePosition{Index: 0, Line: 0, Column: 0},
							),
						),
					),
					M_WORD_P(
						&SourcePosition{Index: 6, Line: 0, Column: 6},
						M_TUPLE_P(
							&SourcePosition{Index: 6, Line: 0, Column: 6},
							M_SCRIPT_P(
								&SourcePosition{Index: 6, Line: 0, Column: 6},
								M_SENTENCE_P(
									&SourcePosition{Index: 8, Line: 0, Column: 8},
									M_WORD_P(
										&SourcePosition{Index: 8, Line: 0, Column: 8},
										M_TUPLE_P(
											&SourcePosition{Index: 8, Line: 0, Column: 8},
											M_SCRIPT_P(
												&SourcePosition{Index: 8, Line: 0, Column: 8},
											),
										),
									),
									M_WORD_P(
										&SourcePosition{Index: 14, Line: 1, Column: 3},
										M_TUPLE_P(
											&SourcePosition{Index: 14, Line: 1, Column: 3},
											M_SCRIPT_P(
												&SourcePosition{Index: 14, Line: 1, Column: 3},
											),
										),
									),
								),
							),
						),
					),
				),
			)))
		})
		Specify("blocks", func() {
			script := parse("{ ; } { { \n } {}}")
			Expect(toTree(script)).To(Equal(M_SCRIPT_P(
				&SourcePosition{Index: 0, Line: 0, Column: 0},
				M_SENTENCE_P(
					&SourcePosition{Index: 0, Line: 0, Column: 0},
					M_WORD_P(
						&SourcePosition{Index: 0, Line: 0, Column: 0},
						M_BLOCK_P(
							&SourcePosition{Index: 0, Line: 0, Column: 0},
							M_SCRIPT_P(
								&SourcePosition{Index: 0, Line: 0, Column: 0},
							),
						),
					),
					M_WORD_P(
						&SourcePosition{Index: 6, Line: 0, Column: 6},
						M_BLOCK_P(
							&SourcePosition{Index: 6, Line: 0, Column: 6},
							M_SCRIPT_P(
								&SourcePosition{Index: 6, Line: 0, Column: 6},

								M_SENTENCE_P(
									&SourcePosition{Index: 8, Line: 0, Column: 8},
									M_WORD_P(
										&SourcePosition{Index: 8, Line: 0, Column: 8},
										M_BLOCK_P(
											&SourcePosition{Index: 8, Line: 0, Column: 8},
											M_SCRIPT_P(
												&SourcePosition{Index: 8, Line: 0, Column: 8},
											),
										),
									),
									M_WORD_P(
										&SourcePosition{Index: 14, Line: 1, Column: 3},
										M_BLOCK_P(
											&SourcePosition{Index: 14, Line: 1, Column: 3},
											M_SCRIPT_P(
												&SourcePosition{Index: 14, Line: 1, Column: 3},
											),
										),
									),
								),
							),
						),
					),
				),
			)))
		})
		Specify("expressions", func() {
			script := parse("[ ; ] [ [ \n ] []]")
			Expect(toTree(script)).To(Equal(M_SCRIPT_P(
				&SourcePosition{Index: 0, Line: 0, Column: 0},
				M_SENTENCE_P(
					&SourcePosition{Index: 0, Line: 0, Column: 0},
					M_WORD_P(
						&SourcePosition{Index: 0, Line: 0, Column: 0},
						M_EXPRESSION_P(
							&SourcePosition{Index: 0, Line: 0, Column: 0},
							M_SCRIPT_P(
								&SourcePosition{Index: 0, Line: 0, Column: 0},
							),
						),
					),
					M_WORD_P(
						&SourcePosition{Index: 6, Line: 0, Column: 6},
						M_EXPRESSION_P(
							&SourcePosition{Index: 6, Line: 0, Column: 6},
							M_SCRIPT_P(
								&SourcePosition{Index: 6, Line: 0, Column: 6},
								M_SENTENCE_P(
									&SourcePosition{Index: 8, Line: 0, Column: 8},
									M_WORD_P(
										&SourcePosition{Index: 8, Line: 0, Column: 8},
										M_EXPRESSION_P(
											&SourcePosition{Index: 8, Line: 0, Column: 8},
											M_SCRIPT_P(
												&SourcePosition{Index: 8, Line: 0, Column: 8},
											),
										),
									),
									M_WORD_P(
										&SourcePosition{Index: 14, Line: 1, Column: 3},
										M_EXPRESSION_P(
											&SourcePosition{Index: 14, Line: 1, Column: 3},
											M_SCRIPT_P(
												&SourcePosition{Index: 14, Line: 1, Column: 3},
											),
										),
									),
								),
							),
						),
					),
				),
			)))
		})
		Specify("strings", func() {
			script := parse("\"this\" \"is\" \"a\nlist $of\" \"\" \"strings\"")
			Expect(toTree(script)).To(Equal(M_SCRIPT_P(
				&SourcePosition{Index: 0, Line: 0, Column: 0},
				M_SENTENCE_P(
					&SourcePosition{Index: 0, Line: 0, Column: 0},
					M_WORD_P(
						&SourcePosition{Index: 0, Line: 0, Column: 0},
						M_STRING_P(
							&SourcePosition{Index: 0, Line: 0, Column: 0},
							M_LITERAL_P(
								&SourcePosition{Index: 1, Line: 0, Column: 1},
								"this",
							),
						),
					),
					M_WORD_P(
						&SourcePosition{Index: 7, Line: 0, Column: 7},
						M_STRING_P(
							&SourcePosition{Index: 7, Line: 0, Column: 7},
							M_LITERAL_P(
								&SourcePosition{Index: 8, Line: 0, Column: 8},
								"is",
							),
						),
					),
					M_WORD_P(
						&SourcePosition{Index: 12, Line: 0, Column: 12},
						M_STRING_P(
							&SourcePosition{Index: 12, Line: 0, Column: 12},
							M_LITERAL_P(
								&SourcePosition{Index: 13, Line: 0, Column: 13},
								"a\nlist ",
							),
							M_SUBSTITUTE_NEXT_P(
								&SourcePosition{Index: 20, Line: 1, Column: 5},
								"$",
							),
							M_LITERAL_P(
								&SourcePosition{Index: 21, Line: 1, Column: 6},
								"of",
							),
						),
					),
					M_WORD_P(
						&SourcePosition{Index: 25, Line: 1, Column: 10},
						M_STRING_P(
							&SourcePosition{Index: 25, Line: 1, Column: 10},
						),
					),
					M_WORD_P(
						&SourcePosition{Index: 28, Line: 1, Column: 13},
						M_STRING_P(
							&SourcePosition{Index: 28, Line: 1, Column: 13},
							M_LITERAL_P(
								&SourcePosition{Index: 29, Line: 1, Column: 14},
								"strings",
							),
						),
					),
				),
			)))
		})
		Specify("here-strings", func() {
			script := parse(
				"\"\"\"this is\"\"\" \"\"\"a\nlist\"\"\"\n  \"\"\"of\"\"\" \"\"\"here-strings\"\"\"",
			)
			Expect(toTree(script)).To(Equal(M_SCRIPT_P(
				&SourcePosition{Index: 0, Line: 0, Column: 0},
				M_SENTENCE_P(
					&SourcePosition{Index: 0, Line: 0, Column: 0},
					M_WORD_P(
						&SourcePosition{Index: 0, Line: 0, Column: 0},
						M_HERE_STRING_P(
							&SourcePosition{Index: 0, Line: 0, Column: 0},
							"this is",
						),
					),
					M_WORD_P(
						&SourcePosition{Index: 14, Line: 0, Column: 14},
						M_HERE_STRING_P(
							&SourcePosition{Index: 14, Line: 0, Column: 14},
							"a\nlist",
						),
					),
				),
				M_SENTENCE_P(
					&SourcePosition{Index: 29, Line: 2, Column: 2},
					M_WORD_P(
						&SourcePosition{Index: 29, Line: 2, Column: 2},
						M_HERE_STRING_P(
							&SourcePosition{Index: 29, Line: 2, Column: 2},
							"of",
						),
					),
					M_WORD_P(
						&SourcePosition{Index: 38, Line: 2, Column: 11},
						M_HERE_STRING_P(
							&SourcePosition{Index: 38, Line: 2, Column: 11},
							"here-strings",
						),
					),
				),
			)))
		})
		Specify("tagged strings", func() {
			script := parse(
				"\"\"A\nthis is\nA\"\" \"\"B\na\nlist\nB\"\"\n  \"\"C\nof\nC\"\" \"\"D\ntagged strings\nD\"\"",
			)
			Expect(toTree(script)).To(Equal(M_SCRIPT_P(
				&SourcePosition{Index: 0, Line: 0, Column: 0},
				M_SENTENCE_P(
					&SourcePosition{Index: 0, Line: 0, Column: 0},
					M_WORD_P(
						&SourcePosition{Index: 0, Line: 0, Column: 0},
						M_TAGGED_STRING_P(
							&SourcePosition{Index: 0, Line: 0, Column: 0},
							"this is\n",
						),
					),
					M_WORD_P(
						&SourcePosition{Index: 16, Line: 2, Column: 4},
						M_TAGGED_STRING_P(
							&SourcePosition{Index: 16, Line: 2, Column: 4},
							"a\nlist\n",
						),
					),
				),
				M_SENTENCE_P(
					&SourcePosition{Index: 33, Line: 6, Column: 2},
					M_WORD_P(
						&SourcePosition{Index: 33, Line: 6, Column: 2},
						M_TAGGED_STRING_P(
							&SourcePosition{Index: 33, Line: 6, Column: 2},
							"of\n",
						),
					),
					M_WORD_P(
						&SourcePosition{Index: 44, Line: 8, Column: 4},
						M_TAGGED_STRING_P(
							&SourcePosition{Index: 44, Line: 8, Column: 4},
							"tagged strings\n",
						),
					),
				),
			)))
		})
		Specify("line comments", func() {
			script := parse(
				"# this is\n# a\\\nlist\n  ###of\n    ## line comments",
			)
			Expect(toTree(script)).To(Equal(M_SCRIPT_P(
				&SourcePosition{Index: 0, Line: 0, Column: 0},
				M_SENTENCE_P(
					&SourcePosition{Index: 0, Line: 0, Column: 0},
					M_WORD_P(
						&SourcePosition{Index: 0, Line: 0, Column: 0},
						M_LINE_COMMENT_P(
							&SourcePosition{Index: 0, Line: 0, Column: 0},
							" this is",
						),
					),
				),
				M_SENTENCE_P(
					&SourcePosition{Index: 10, Line: 1, Column: 0},
					M_WORD_P(
						&SourcePosition{Index: 10, Line: 1, Column: 0},
						M_LINE_COMMENT_P(
							&SourcePosition{Index: 10, Line: 1, Column: 0},
							" a list",
						),
					),
				),
				M_SENTENCE_P(
					&SourcePosition{Index: 22, Line: 3, Column: 2},
					M_WORD_P(
						&SourcePosition{Index: 22, Line: 3, Column: 2},
						M_LINE_COMMENT_P(
							&SourcePosition{Index: 22, Line: 3, Column: 2},
							"of",
						),
					),
				),
				M_SENTENCE_P(
					&SourcePosition{Index: 32, Line: 4, Column: 4},
					M_WORD_P(
						&SourcePosition{Index: 32, Line: 4, Column: 4},
						M_LINE_COMMENT_P(
							&SourcePosition{Index: 32, Line: 4, Column: 4},
							" line comments",
						),
					),
				),
			)))
		})
		Specify("block comments", func() {
			script := parse(
				"#{this is}# ##{a\nlist}##\n  #{of}# ###{block comments}###",
			)
			Expect(toTree(script)).To(Equal(M_SCRIPT_P(
				&SourcePosition{Index: 0, Line: 0, Column: 0},
				M_SENTENCE_P(
					&SourcePosition{Index: 0, Line: 0, Column: 0},
					M_WORD_P(
						&SourcePosition{Index: 0, Line: 0, Column: 0},
						M_BLOCK_COMMENT_P(
							&SourcePosition{Index: 0, Line: 0, Column: 0},
							"this is",
						),
					),
					M_WORD_P(
						&SourcePosition{Index: 12, Line: 0, Column: 12},
						M_BLOCK_COMMENT_P(
							&SourcePosition{Index: 12, Line: 0, Column: 12},
							"a\nlist",
						),
					),
				),
				M_SENTENCE_P(
					&SourcePosition{Index: 27, Line: 2, Column: 2},
					M_WORD_P(
						&SourcePosition{Index: 27, Line: 2, Column: 2},
						M_BLOCK_COMMENT_P(
							&SourcePosition{Index: 27, Line: 2, Column: 2},
							"of",
						),
					),
					M_WORD_P(
						&SourcePosition{Index: 34, Line: 2, Column: 9},
						M_BLOCK_COMMENT_P(
							&SourcePosition{Index: 34, Line: 2, Column: 9},
							"block comments",
						),
					),
				),
			)))
		})
		Specify("substitutions", func() {
			script := parse("$this $${is a\nlist} $*$$*$(of substitutions)")
			Expect(toTree(script)).To(Equal(M_SCRIPT_P(
				&SourcePosition{Index: 0, Line: 0, Column: 0},
				M_SENTENCE_P(
					&SourcePosition{Index: 0, Line: 0, Column: 0},
					M_WORD_P(
						&SourcePosition{Index: 0, Line: 0, Column: 0},
						M_SUBSTITUTE_NEXT_P(
							&SourcePosition{Index: 0, Line: 0, Column: 0},
							"$",
						),
						M_LITERAL_P(
							&SourcePosition{Index: 1, Line: 0, Column: 1},
							"this",
						),
					),
					M_WORD_P(
						&SourcePosition{Index: 6, Line: 0, Column: 6},
						M_SUBSTITUTE_NEXT_P(
							&SourcePosition{Index: 6, Line: 0, Column: 6},
							"$",
						),
						M_SUBSTITUTE_NEXT_P(
							&SourcePosition{Index: 7, Line: 0, Column: 7},
							"$",
						),
						M_BLOCK_P(
							&SourcePosition{Index: 8, Line: 0, Column: 8},
							M_SCRIPT_P(
								&SourcePosition{Index: 8, Line: 0, Column: 8},
								M_SENTENCE_P(
									&SourcePosition{Index: 9, Line: 0, Column: 9},
									M_WORD_P(
										&SourcePosition{Index: 9, Line: 0, Column: 9},
										M_LITERAL_P(
											&SourcePosition{Index: 9, Line: 0, Column: 9},
											"is",
										),
									),
									M_WORD_P(
										&SourcePosition{Index: 12, Line: 0, Column: 12},
										M_LITERAL_P(
											&SourcePosition{Index: 12, Line: 0, Column: 12},
											"a",
										),
									),
								),
								M_SENTENCE_P(
									&SourcePosition{Index: 14, Line: 1, Column: 0},
									M_WORD_P(
										&SourcePosition{Index: 14, Line: 1, Column: 0},
										M_LITERAL_P(
											&SourcePosition{Index: 14, Line: 1, Column: 0},
											"list",
										),
									),
								),
							),
						),
					),
					M_WORD_P(
						&SourcePosition{Index: 20, Line: 1, Column: 6},
						M_EXPAND_NEXT_P(
							&SourcePosition{Index: 20, Line: 1, Column: 6},
							"$*",
						),
						M_SUBSTITUTE_NEXT_P(
							&SourcePosition{Index: 22, Line: 1, Column: 8},
							"$",
						),
						M_SUBSTITUTE_NEXT_P(
							&SourcePosition{Index: 23, Line: 1, Column: 9},
							"$*",
						),
						M_SUBSTITUTE_NEXT_P(
							&SourcePosition{Index: 25, Line: 1, Column: 11},
							"$",
						),
						M_TUPLE_P(
							&SourcePosition{Index: 26, Line: 1, Column: 12},
							M_SCRIPT_P(
								&SourcePosition{Index: 26, Line: 1, Column: 12},
								M_SENTENCE_P(
									&SourcePosition{Index: 27, Line: 1, Column: 13},
									M_WORD_P(
										&SourcePosition{Index: 27, Line: 1, Column: 13},
										M_LITERAL_P(
											&SourcePosition{Index: 27, Line: 1, Column: 13},
											"of",
										),
									),
									M_WORD_P(
										&SourcePosition{Index: 30, Line: 1, Column: 16},
										M_LITERAL_P(
											&SourcePosition{Index: 30, Line: 1, Column: 16},
											"substitutions",
										),
									),
								),
							),
						),
					),
				),
			)))
		})
	})
})
