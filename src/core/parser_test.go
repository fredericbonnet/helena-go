package core_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "helena/core"
)

type MS any
type MW any
type MM any
type M_LITERAL struct{ string }
type M_TUPLE []MS
type M_BLOCK []MS
type M_EXPRESSION []MS
type M_STRING []MM
type M_HERE_STRING struct{ string }
type M_TAGGED_STRING struct{ string }
type M_LINE_COMMENT struct{ string }
type M_BLOCK_COMMENT struct{ string }
type M_SUBSTITUTE_NEXT struct{ string }
type M_EXPAND_NEXT struct{ string }

func mapMorpheme(morpheme Morpheme) any {
	switch morpheme.Type() {
	case MorphemeType_LITERAL:
		return M_LITERAL{morpheme.(LiteralMorpheme).Value}

	case MorphemeType_TUPLE:
		{
			subscript := morpheme.(TupleMorpheme).Subscript
			return M_TUPLE{toTree(&subscript)}
		}

	case MorphemeType_BLOCK:
		{
			subscript := morpheme.(BlockMorpheme).Subscript
			return M_BLOCK{toTree(&subscript)}
		}

	case MorphemeType_EXPRESSION:
		{
			subscript := morpheme.(ExpressionMorpheme).Subscript
			return M_EXPRESSION{toTree(&subscript)}
		}

	case MorphemeType_STRING:
		{
			morphemes := make([]MM, len(morpheme.(StringMorpheme).Morphemes))
			for k, morpheme := range morpheme.(StringMorpheme).Morphemes {
				morphemes[k] = mapMorpheme(morpheme)
			}
			return M_STRING{morphemes}
		}

	case MorphemeType_HERE_STRING:
		return M_HERE_STRING{morpheme.(HereStringMorpheme).Value}

	case MorphemeType_TAGGED_STRING:
		return M_TAGGED_STRING{morpheme.(TaggedStringMorpheme).Value}

	case MorphemeType_LINE_COMMENT:
		return M_LINE_COMMENT{morpheme.(LineCommentMorpheme).Value}

	case MorphemeType_BLOCK_COMMENT:
		return M_BLOCK_COMMENT{morpheme.(BlockCommentMorpheme).Value}

	case MorphemeType_SUBSTITUTE_NEXT:
		if morpheme.(SubstituteNextMorpheme).Expansion {
			return M_EXPAND_NEXT{morpheme.(SubstituteNextMorpheme).Value}
		} else {
			return M_SUBSTITUTE_NEXT{morpheme.(SubstituteNextMorpheme).Value}
		}

	default:
		panic("CANTHAPPEN")
	}
}

func toTree(script *Script) []MS {
	sentences := make([]MS, len(script.Sentences))
	for i, sentence := range script.Sentences {
		words := make([]MW, len(sentence.Words))
		for j, word := range sentence.Words {
			morphemes := make([]MM, len(word.Word.Morphemes))
			for k, morpheme := range word.Word.Morphemes {
				morphemes[k] = mapMorpheme(morpheme)
			}
			words[j] = morphemes
		}
		sentences[i] = words
	}
	return sentences
}

var _ = Describe("Parser", func() {
	var tokenizer Tokenizer
	var parser *Parser

	parse := func(script string) *Script {
		return parser.Parse(tokenizer.Tokenize(script)).Script
	}

	BeforeEach(func() {
		tokenizer = Tokenizer{}
		parser = &Parser{}
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
			Expect(toTree(script)).To(Equal([]MS{[]MW{[]MM{M_LITERAL{"sentence"}}}}))
		})
		Specify("single sentence surrounded by blank lines", func() {
			script := parse("  \nsentence\n  ")
			Expect(toTree(script)).To(Equal([]MS{[]MW{[]MM{M_LITERAL{"sentence"}}}}))
		})
		Specify("two sentences separated by newline", func() {
			script := parse("sentence1\nsentence2")
			Expect(toTree(script)).To(Equal([]MS{
				[]MW{[]MM{M_LITERAL{"sentence1"}}},
				[]MW{[]MM{M_LITERAL{"sentence2"}}},
			}))
		})
		Specify("two sentences separated by semicolon", func() {
			script := parse("sentence1;sentence2")
			Expect(toTree(script)).To(Equal([]MS{
				[]MW{[]MM{M_LITERAL{"sentence1"}}},
				[]MW{[]MM{M_LITERAL{"sentence2"}}},
			}))
		})
		Specify("blank sentences are ignored", func() {
			script := parse("\nsentence1;; \t  ;\n\n \t   \nsentence2\n")
			Expect(toTree(script)).To(Equal([]MS{
				[]MW{[]MM{M_LITERAL{"sentence1"}}},
				[]MW{[]MM{M_LITERAL{"sentence2"}}},
			}))
		})
	})
	Describe("words", func() {
		Describe("literals", func() {
			Specify("single literal", func() {
				script := parse("word")
				Expect(toTree(script)).To(Equal([]MS{[]MW{[]MM{M_LITERAL{"word"}}}}))
			})
			Specify("single literal surrounded by spaces", func() {
				script := parse(" word ")
				Expect(toTree(script)).To(Equal([]MS{[]MW{[]MM{M_LITERAL{"word"}}}}))
			})
			Specify("single literal with escape sequences", func() {
				script := parse("one\\tword")
				Expect(toTree(script)).To(Equal([]MS{[]MW{[]MM{M_LITERAL{"one\tword"}}}}))
			})
			Specify("two literals separated by whitespace", func() {
				script := parse("word1 word2")
				Expect(toTree(script)).To(Equal([]MS{
					[]MW{[]MM{M_LITERAL{"word1"}}, []MM{M_LITERAL{"word2"}}},
				}))
			})
			Specify("two literals separated by continuation", func() {
				script := parse("word1\\\nword2")
				Expect(toTree(script)).To(Equal([]MS{
					[]MW{[]MM{M_LITERAL{"word1"}}, []MM{M_LITERAL{"word2"}}},
				}))
			})
		})
		Describe("tuples", func() {
			Specify("empty tuple", func() {
				script := parse("()")
				Expect(toTree(script)).To(Equal([]MS{[]MW{[]MM{M_TUPLE{[]MS{}}}}}))
			})
			Specify("tuple with one word", func() {
				script := parse("(word)")
				Expect(toTree(script)).To(Equal([]MS{
					[]MW{[]MM{M_TUPLE{[]MS{[]MW{[]MM{M_LITERAL{"word"}}}}}}},
				}))
			})
			Specify("tuple with two levels", func() {
				script := parse("(word1 (subword1 subword2) word2)")
				Expect(toTree(script)).To(Equal([]MS{
					[]MW{
						[]MM{
							M_TUPLE{[]MS{
								[]MW{
									[]MM{M_LITERAL{"word1"}},
									[]MM{
										M_TUPLE{[]MS{
											[]MW{
												[]MM{M_LITERAL{"subword1"}},
												[]MM{M_LITERAL{"subword2"}},
											},
										}},
									},
									[]MM{M_LITERAL{"word2"}},
								},
							}},
						},
					},
				}))
			})
			Describe("exceptions", func() {
				Specify("unterminated tuple", func() {
					tokens := tokenizer.Tokenize("(")
					Expect(parser.Parse(tokens)).To(Equal(
						PARSE_ERROR("unmatched left parenthesis")),
					)
				})
				Specify("unmatched right parenthesis", func() {
					tokens := tokenizer.Tokenize(")")
					Expect(parser.Parse(tokens)).To(Equal(
						PARSE_ERROR("unmatched right parenthesis")),
					)
				})
				Specify("mismatched right brace", func() {
					tokens := tokenizer.Tokenize("(}")
					Expect(parser.Parse(tokens)).To(Equal(
						PARSE_ERROR("mismatched right brace")),
					)
				})
				Specify("mismatched right bracket", func() {
					tokens := tokenizer.Tokenize("(]")
					Expect(parser.Parse(tokens)).To(Equal(
						PARSE_ERROR("mismatched right bracket")),
					)
				})
			})
		})
		Describe("blocks", func() {
			Specify("empty block", func() {
				script := parse("{}")
				Expect(toTree(script)).To(Equal([]MS{[]MW{[]MM{M_BLOCK{[]MS{}}}}}))
			})
			Specify("block with one word", func() {
				script := parse("{word}")
				Expect(toTree(script)).To(Equal([]MS{
					[]MW{[]MM{M_BLOCK{[]MS{[]MW{[]MM{M_LITERAL{"word"}}}}}}},
				}))
			})
			Specify("block with two levels", func() {
				script := parse("{word1 {subword1 subword2} word2}")
				Expect(toTree(script)).To(Equal([]MS{
					[]MW{
						[]MM{
							M_BLOCK{[]MS{
								[]MW{
									[]MM{M_LITERAL{"word1"}},
									[]MM{
										M_BLOCK{[]MS{
											[]MW{
												[]MM{M_LITERAL{"subword1"}},
												[]MM{M_LITERAL{"subword2"}},
											},
										}},
									},
									[]MM{M_LITERAL{"word2"}},
								},
							}},
						},
					},
				},
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
					Expect(parser.Parse(tokens)).To(Equal(
						PARSE_ERROR("unmatched left brace"),
					))
				})
				Specify("unmatched right brace", func() {
					tokens := tokenizer.Tokenize("}")
					Expect(parser.Parse(tokens)).To(Equal(
						PARSE_ERROR("unmatched right brace"),
					))
				})
				Specify("mismatched right parenthesis", func() {
					tokens := tokenizer.Tokenize("{)")
					Expect(parser.Parse(tokens)).To(Equal(
						PARSE_ERROR("mismatched right parenthesis"),
					))
				})
				Specify("mismatched right bracket", func() {
					tokens := tokenizer.Tokenize("{]")
					Expect(parser.Parse(tokens)).To(Equal(
						PARSE_ERROR("mismatched right bracket"),
					))
				})
			})
		})
		Describe("expressions", func() {
			Specify("empty expression", func() {
				script := parse("[]")
				Expect(toTree(script)).To(Equal([]MS{[]MW{[]MM{M_EXPRESSION{[]MS{}}}}}))
			})
			Specify("expression with one word", func() {
				script := parse("[word]")
				Expect(toTree(script)).To(Equal([]MS{
					[]MW{[]MM{M_EXPRESSION{[]MS{[]MW{[]MM{M_LITERAL{"word"}}}}}}},
				}))
			})
			Specify("expression with two levels", func() {
				script := parse("[word1 [subword1 subword2] word2]")
				Expect(toTree(script)).To(Equal([]MS{
					[]MW{
						[]MM{
							M_EXPRESSION{[]MS{
								[]MW{
									[]MM{M_LITERAL{"word1"}},
									[]MM{
										M_EXPRESSION{[]MS{
											[]MW{
												[]MM{M_LITERAL{"subword1"}},
												[]MM{M_LITERAL{"subword2"}},
											},
										}},
									},
									[]MM{M_LITERAL{"word2"}},
								}},
							},
						},
					},
				}))
			})
			Describe("exceptions", func() {
				Specify("unterminated expression", func() {
					tokens := tokenizer.Tokenize("[")
					Expect(parser.Parse(tokens)).To(Equal(
						PARSE_ERROR("unmatched left bracket"),
					))
				})
				Specify("unmatched right bracket", func() {
					tokens := tokenizer.Tokenize("]")
					Expect(parser.Parse(tokens)).To(Equal(
						PARSE_ERROR("unmatched right bracket"),
					))
				})
				Specify("mismatched right parenthesis", func() {
					tokens := tokenizer.Tokenize("[)")
					Expect(parser.Parse(tokens)).To(Equal(
						PARSE_ERROR("mismatched right parenthesis"),
					))
				})
				Specify("mismatched right brace", func() {
					tokens := tokenizer.Tokenize("[}")
					Expect(parser.Parse(tokens)).To(Equal(
						PARSE_ERROR("mismatched right brace"),
					))
				})
			})
		})
		Describe("strings", func() {
			Specify("empty string", func() {
				script := parse(`""`)
				Expect(toTree(script)).To(Equal([]MS{[]MW{[]MM{M_STRING{[]MM{}}}}}))
			})
			Specify("simple string", func() {
				script := parse(`"string"`)
				Expect(toTree(script)).To(Equal([]MS{
					[]MW{[]MM{M_STRING{[]MM{M_LITERAL{"string"}}}}},
				}))
			})
			Specify("longer string", func() {
				script := parse(`"this is a string"`)
				Expect(toTree(script)).To(Equal([]MS{
					[]MW{[]MM{M_STRING{[]MM{M_LITERAL{"this is a string"}}}}},
				}))
			})
			Specify("string with whitespaces and continuations", func() {
				script := parse("\"this  \t  is\r\f a   \\\n  \t  string\"")
				Expect(toTree(script)).To(Equal([]MS{
					[]MW{[]MM{M_STRING{[]MM{M_LITERAL{"this  \t  is\r\f a    string"}}}}},
				}))
			})
			Specify("string with special characters", func() {
				script := parse(`"this {is (a #string"`)
				Expect(toTree(script)).To(Equal([]MS{
					[]MW{[]MM{M_STRING{[]MM{M_LITERAL{"this {is (a #string"}}}}},
				}))
				Expect(toTree(parse(`"("`))).To(Equal([]MS{
					[]MW{[]MM{M_STRING{[]MM{M_LITERAL{"("}}}}},
				}))
				Expect(toTree(parse(`"{"`))).To(Equal([]MS{
					[]MW{[]MM{M_STRING{[]MM{M_LITERAL{"{"}}}}},
				}))
			})
			Describe("expressions", func() {
				Specify("empty expression", func() {
					script := parse(`"[]"`)
					Expect(toTree(script)).To(Equal([]MS{[]MW{[]MM{M_STRING{[]MM{M_EXPRESSION{[]MS{}}}}}}}))
				})
				Specify("expression with one word", func() {
					script := parse(`"[word]"`)
					Expect(toTree(script)).To(Equal([]MS{
						[]MW{[]MM{M_STRING{[]MM{M_EXPRESSION{[]MS{[]MW{[]MM{M_LITERAL{"word"}}}}}}}}},
					}))
				})
				Specify("expression with two levels", func() {
					script := parse(`"[word1 [subword1 subword2] word2]"`)
					Expect(toTree(script)).To(Equal([]MS{
						[]MW{
							[]MM{
								M_STRING{[]MM{
									M_EXPRESSION{[]MS{
										[]MW{
											[]MM{M_LITERAL{"word1"}},
											[]MM{
												M_EXPRESSION{[]MS{
													[]MW{
														[]MM{M_LITERAL{"subword1"}},
														[]MM{M_LITERAL{"subword2"}},
													},
												}},
											},
											[]MM{M_LITERAL{"word2"}},
										},
									}},
								}},
							},
						},
					},
					))
				})
				Describe("exceptions", func() {
					Specify("unterminated expression", func() {
						tokens := tokenizer.Tokenize(`"[`)
						Expect(parser.Parse(tokens)).To(Equal(
							PARSE_ERROR("unmatched left bracket"),
						))
					})
					Specify("mismatched right parenthesis", func() {
						tokens := tokenizer.Tokenize(`"[)"`)
						Expect(parser.Parse(tokens)).To(Equal(
							PARSE_ERROR("mismatched right parenthesis"),
						))
					})
					Specify("mismatched right brace", func() {
						tokens := tokenizer.Tokenize(`"[}"`)
						Expect(parser.Parse(tokens)).To(Equal(
							PARSE_ERROR("mismatched right brace"),
						))
					})
				})
			})
			Describe("substitutions", func() {
				Specify("lone dollar", func() {
					script := parse(`"$"`)
					Expect(toTree(script)).To(Equal([]MS{[]MW{[]MM{M_STRING{[]MM{M_LITERAL{"$"}}}}}}))
				})
				Specify("simple variable", func() {
					script := parse(`"$a"`)
					Expect(toTree(script)).To(Equal([]MS{
						[]MW{[]MM{M_STRING{[]MM{M_SUBSTITUTE_NEXT{"$"}, M_LITERAL{"a"}}}}},
					}))
				})
				Specify("Unicode variable name", func() {
					script := parse("\"$a\u1234\"")
					Expect(toTree(script)).To(Equal([]MS{
						[]MW{[]MM{M_STRING{[]MM{M_SUBSTITUTE_NEXT{"$"}, M_LITERAL{"a\u1234"}}}}},
					}))
				})
				Specify("block", func() {
					script := parse(`"${a}"`)
					Expect(toTree(script)).To(Equal([]MS{
						[]MW{
							[]MM{
								M_STRING{[]MM{
									M_SUBSTITUTE_NEXT{"$"},
									M_BLOCK{[]MS{[]MW{[]MM{M_LITERAL{"a"}}}}},
								}},
							},
						},
					}))
				})
				Specify("expression", func() {
					script := parse(`"$[a]"`)
					Expect(toTree(script)).To(Equal([]MS{
						[]MW{
							[]MM{
								M_STRING{[]MM{
									M_SUBSTITUTE_NEXT{"$"},
									M_EXPRESSION{[]MS{[]MW{[]MM{M_LITERAL{"a"}}}}},
								}},
							},
						},
					}))
				})
				Specify("multiple substitution", func() {
					script := parse(`"$$a $$$b $$$$[c]"`)
					Expect(toTree(script)).To(Equal([]MS{
						[]MW{
							[]MM{
								M_STRING{[]MM{
									M_SUBSTITUTE_NEXT{"$"},
									M_SUBSTITUTE_NEXT{"$"},
									M_LITERAL{"a"},
									M_LITERAL{" "},
									M_SUBSTITUTE_NEXT{"$"},
									M_SUBSTITUTE_NEXT{"$"},
									M_SUBSTITUTE_NEXT{"$"},
									M_LITERAL{"b"},
									M_LITERAL{" "},
									M_SUBSTITUTE_NEXT{"$"},
									M_SUBSTITUTE_NEXT{"$"},
									M_SUBSTITUTE_NEXT{"$"},
									M_SUBSTITUTE_NEXT{"$"},
									M_EXPRESSION{[]MS{[]MW{[]MM{M_LITERAL{"c"}}}}},
								}},
							},
						},
					}))
				})
				Specify("expansion", func() {
					script := parse(`"$*$$*a $*$[b]"`)
					Expect(toTree(script)).To(Equal([]MS{
						[]MW{
							[]MM{
								M_STRING{[]MM{
									M_EXPAND_NEXT{"$*"},
									M_SUBSTITUTE_NEXT{"$"},
									M_SUBSTITUTE_NEXT{"$*"},
									M_LITERAL{"a"},
									M_LITERAL{" "},
									M_EXPAND_NEXT{"$*"},
									M_SUBSTITUTE_NEXT{"$"},
									M_EXPRESSION{[]MS{[]MW{[]MM{M_LITERAL{"b"}}}}},
								}},
							},
						},
					}))
				})
				Describe("variable name delimiters", func() {
					Specify("trailing dollars", func() {
						script := parse(`"a$ b$*$ c$$*$"`)
						Expect(toTree(script)).To(Equal([]MS{
							[]MW{[]MM{M_STRING{[]MM{M_LITERAL{"a$ b$*$ c$$*$"}}}}},
						}))
					})
					Specify("escapes", func() {
						script := parse("\"$a\\x62 $c\\d\"")
						Expect(toTree(script)).To(Equal([]MS{
							[]MW{
								[]MM{
									M_STRING{[]MM{
										M_SUBSTITUTE_NEXT{"$"},
										M_LITERAL{"a"},
										M_LITERAL{"b "},
										M_SUBSTITUTE_NEXT{"$"},
										M_LITERAL{"c"},
										M_LITERAL{"d"},
									}},
								},
							},
						}))
					})
					Specify("parentheses", func() {
						script := parse(`"$(a"`)
						Expect(toTree(script)).To(Equal([]MS{
							[]MW{
								[]MM{
									M_STRING{[]MM{M_LITERAL{"$(a"}}},
								},
							},
						}))
					})
					Specify("special characters", func() {
						script := parse("$a# $b*")
						Expect(toTree(script)).To(Equal([]MS{
							[]MW{
								[]MM{M_SUBSTITUTE_NEXT{"$"}, M_LITERAL{"a"}, M_LITERAL{"#"}},
								[]MM{M_SUBSTITUTE_NEXT{"$"}, M_LITERAL{"b"}, M_LITERAL{"*"}},
							},
						}))
					})
				})
				Describe("selectors", func() {
					Describe("indexed selectors", func() {
						Specify("single", func() {
							script := parse(`"$name[index1] $[expression][index2]"`)
							Expect(toTree(script)).To(Equal([]MS{
								[]MW{
									[]MM{
										M_STRING{[]MM{
											M_SUBSTITUTE_NEXT{"$"},
											M_LITERAL{"name"},
											M_EXPRESSION{[]MS{[]MW{[]MM{M_LITERAL{"index1"}}}}},
											M_LITERAL{" "},
											M_SUBSTITUTE_NEXT{"$"},
											M_EXPRESSION{[]MS{[]MW{[]MM{M_LITERAL{"expression"}}}}},
											M_EXPRESSION{[]MS{[]MW{[]MM{M_LITERAL{"index2"}}}}},
										}},
									},
								},
							}))
						})
						Specify("chained", func() {
							script := parse(
								`"$name[index1][index2][index3] $[expression][index4][index5][index6]"`,
							)
							Expect(toTree(script)).To(Equal([]MS{
								[]MW{
									[]MM{
										M_STRING{[]MM{
											M_SUBSTITUTE_NEXT{"$"},
											M_LITERAL{"name"},
											M_EXPRESSION{[]MS{[]MW{[]MM{M_LITERAL{"index1"}}}}},
											M_EXPRESSION{[]MS{[]MW{[]MM{M_LITERAL{"index2"}}}}},
											M_EXPRESSION{[]MS{[]MW{[]MM{M_LITERAL{"index3"}}}}},
											M_LITERAL{" "},
											M_SUBSTITUTE_NEXT{"$"},
											M_EXPRESSION{[]MS{[]MW{[]MM{M_LITERAL{"expression"}}}}},
											M_EXPRESSION{[]MS{[]MW{[]MM{M_LITERAL{"index4"}}}}},
											M_EXPRESSION{[]MS{[]MW{[]MM{M_LITERAL{"index5"}}}}},
											M_EXPRESSION{[]MS{[]MW{[]MM{M_LITERAL{"index6"}}}}},
										}},
									},
								},
							}))
						})
					})
					Describe("keyed selectors", func() {
						Specify("single", func() {
							script := parse(`"$name(key1) $[expression](key2)"`)
							Expect(toTree(script)).To(Equal([]MS{
								[]MW{
									[]MM{
										M_STRING{[]MM{
											M_SUBSTITUTE_NEXT{"$"},
											M_LITERAL{"name"},
											M_TUPLE{[]MS{[]MW{[]MM{M_LITERAL{"key1"}}}}},
											M_LITERAL{" "},
											M_SUBSTITUTE_NEXT{"$"},
											M_EXPRESSION{[]MS{[]MW{[]MM{M_LITERAL{"expression"}}}}},
											M_TUPLE{[]MS{[]MW{[]MM{M_LITERAL{"key2"}}}}},
										}},
									},
								},
							}))
						})
						Specify("multiple", func() {
							script := parse(
								`"$name(key1 key2) $[expression](key3 key4)"`,
							)
							Expect(toTree(script)).To(Equal([]MS{
								[]MW{
									[]MM{
										M_STRING{[]MM{
											M_SUBSTITUTE_NEXT{"$"},
											M_LITERAL{"name"},
											M_TUPLE{[]MS{
												[]MW{[]MM{M_LITERAL{"key1"}}, []MM{M_LITERAL{"key2"}}},
											}},
											M_LITERAL{" "},
											M_SUBSTITUTE_NEXT{"$"},
											M_EXPRESSION{[]MS{[]MW{[]MM{M_LITERAL{"expression"}}}}},
											M_TUPLE{[]MS{
												[]MW{[]MM{M_LITERAL{"key3"}}, []MM{M_LITERAL{"key4"}}},
											}},
										}},
									},
								},
							}))
						})
						Specify("chained", func() {
							script := parse(
								`"$name(key1)(key2 key3)(key4) $[expression](key5 key6)(key7)(key8 key9)"`,
							)
							Expect(toTree(script)).To(Equal([]MS{
								[]MW{
									[]MM{
										M_STRING{[]MM{
											M_SUBSTITUTE_NEXT{"$"},
											M_LITERAL{"name"},
											M_TUPLE{[]MS{[]MW{[]MM{M_LITERAL{"key1"}}}}},
											M_TUPLE{[]MS{
												[]MW{[]MM{M_LITERAL{"key2"}}, []MM{M_LITERAL{"key3"}}},
											}},
											M_TUPLE{[]MS{[]MW{[]MM{M_LITERAL{"key4"}}}}},
											M_LITERAL{" "},
											M_SUBSTITUTE_NEXT{"$"},
											M_EXPRESSION{[]MS{[]MW{[]MM{M_LITERAL{"expression"}}}}},
											M_TUPLE{[]MS{
												[]MW{[]MM{M_LITERAL{"key5"}}, []MM{M_LITERAL{"key6"}}},
											}},
											M_TUPLE{[]MS{[]MW{[]MM{M_LITERAL{"key7"}}}}},
											M_TUPLE{[]MS{
												[]MW{[]MM{M_LITERAL{"key8"}}, []MM{M_LITERAL{"key9"}}},
											}},
										}},
									},
								},
							}))
						})
					})
					Describe("generic selectors", func() {
						Specify("single", func() {
							script := parse(
								`"$name{selector1 arg1} $[expression]{selector2 arg2}"`,
							)
							Expect(toTree(script)).To(Equal([]MS{
								[]MW{
									[]MM{
										M_STRING{[]MM{
											M_SUBSTITUTE_NEXT{"$"},
											M_LITERAL{"name"},
											M_BLOCK{[]MS{
												[]MW{[]MM{M_LITERAL{"selector1"}}, []MM{M_LITERAL{"arg1"}}},
											}},
											M_LITERAL{" "},
											M_SUBSTITUTE_NEXT{"$"},
											M_EXPRESSION{[]MS{[]MW{[]MM{M_LITERAL{"expression"}}}}},
											M_BLOCK{[]MS{
												[]MW{[]MM{M_LITERAL{"selector2"}}, []MM{M_LITERAL{"arg2"}}},
											}},
										}},
									},
								},
							}))
						})
						Specify("chained", func() {
							script := parse(
								`"$name{selector1 arg1}{selector2}{selector3 arg2 arg3} $[expression]{selector4}{selector5 arg4 arg5}{selector6 arg6}"`,
							)
							Expect(toTree(script)).To(Equal([]MS{
								[]MW{
									[]MM{
										M_STRING{[]MM{
											M_SUBSTITUTE_NEXT{"$"},
											M_LITERAL{"name"},
											M_BLOCK{[]MS{
												[]MW{[]MM{M_LITERAL{"selector1"}}, []MM{M_LITERAL{"arg1"}}},
											}},
											M_BLOCK{[]MS{[]MW{[]MM{M_LITERAL{"selector2"}}}}},
											M_BLOCK{[]MS{
												[]MW{
													[]MM{M_LITERAL{"selector3"}},
													[]MM{M_LITERAL{"arg2"}},
													[]MM{M_LITERAL{"arg3"}},
												},
											}},
											M_LITERAL{" "},
											M_SUBSTITUTE_NEXT{"$"},
											M_EXPRESSION{[]MS{[]MW{[]MM{M_LITERAL{"expression"}}}}},
											M_BLOCK{[]MS{[]MW{[]MM{M_LITERAL{"selector4"}}}}},
											M_BLOCK{[]MS{
												[]MW{
													[]MM{M_LITERAL{"selector5"}},
													[]MM{M_LITERAL{"arg4"}},
													[]MM{M_LITERAL{"arg5"}},
												},
											}},
											M_BLOCK{[]MS{
												[]MW{[]MM{M_LITERAL{"selector6"}}, []MM{M_LITERAL{"arg6"}}},
											}},
										}},
									},
								},
							}))
						})
					})
					Specify("mixed selectors", func() {
						script := parse(
							`"$name(key1 key2){selector1}(key3){selector2 selector3} $[expression]{selector4 selector5}(key4 key5){selector6}(key6)"`,
						)
						Expect(toTree(script)).To(Equal([]MS{
							[]MW{
								[]MM{
									M_STRING{[]MM{
										M_SUBSTITUTE_NEXT{"$"},
										M_LITERAL{"name"},
										M_TUPLE{[]MS{[]MW{[]MM{M_LITERAL{"key1"}}, []MM{M_LITERAL{"key2"}}}}},
										M_BLOCK{[]MS{[]MW{[]MM{M_LITERAL{"selector1"}}}}},
										M_TUPLE{[]MS{[]MW{[]MM{M_LITERAL{"key3"}}}}},
										M_BLOCK{[]MS{
											[]MW{
												[]MM{M_LITERAL{"selector2"}},
												[]MM{M_LITERAL{"selector3"}},
											},
										},
										},
										M_LITERAL{" "},
										M_SUBSTITUTE_NEXT{"$"},
										M_EXPRESSION{[]MS{[]MW{[]MM{M_LITERAL{"expression"}}}}},
										M_BLOCK{[]MS{
											[]MW{
												[]MM{M_LITERAL{"selector4"}},
												[]MM{M_LITERAL{"selector5"}},
											},
										},
										},
										M_TUPLE{[]MS{[]MW{[]MM{M_LITERAL{"key4"}}, []MM{M_LITERAL{"key5"}}}}},
										M_BLOCK{[]MS{[]MW{[]MM{M_LITERAL{"selector6"}}}}},
										M_TUPLE{[]MS{[]MW{[]MM{M_LITERAL{"key6"}}}}},
									}},
								},
							},
						}))
					})
					Specify("nested selectors", func() {
						script := parse(
							`"$name1(key1 $name2{selector1} $[expression1](key2)) $[expression2]{selector2 $name3(key3)}"`,
						)
						Expect(toTree(script)).To(Equal([]MS{
							[]MW{
								[]MM{
									M_STRING{[]MM{
										M_SUBSTITUTE_NEXT{"$"},
										M_LITERAL{"name1"},
										M_TUPLE{[]MS{
											[]MW{
												[]MM{M_LITERAL{"key1"}},
												[]MM{
													M_SUBSTITUTE_NEXT{"$"},
													M_LITERAL{"name2"},
													M_BLOCK{[]MS{[]MW{[]MM{M_LITERAL{"selector1"}}}}},
												},
												[]MM{
													M_SUBSTITUTE_NEXT{"$"},
													M_EXPRESSION{[]MS{[]MW{[]MM{M_LITERAL{"expression1"}}}}},
													M_TUPLE{[]MS{[]MW{[]MM{M_LITERAL{"key2"}}}}},
												},
											},
										}},
										M_LITERAL{" "},
										M_SUBSTITUTE_NEXT{"$"},
										M_EXPRESSION{[]MS{[]MW{[]MM{M_LITERAL{"expression2"}}}}},
										M_BLOCK{[]MS{
											[]MW{
												[]MM{M_LITERAL{"selector2"}},
												[]MM{
													M_SUBSTITUTE_NEXT{"$"},
													M_LITERAL{"name3"},
													M_TUPLE{[]MS{[]MW{[]MM{M_LITERAL{"key3"}}}}},
												},
											},
										}},
									}},
								},
							},
						}))
					})
				})
			})
			Describe("exceptions", func() {
				Specify("unterminated string", func() {
					tokens := tokenizer.Tokenize(`"`)
					Expect(parser.Parse(tokens)).To(Equal(
						PARSE_ERROR("unmatched string delimiter"),
					))
				})
				Specify("extra quotes", func() {
					tokens := tokenizer.Tokenize(`"hello""`)
					Expect(parser.Parse(tokens)).To(Equal(
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
				Expect(toTree(script)).To(Equal([]MS{
					[]MW{
						[]MM{
							M_HERE_STRING{
								"some \" \\\n    $arbitrary [character\n  \"\" sequence",
							}},
					},
				}))
			})
			Specify("4-quote delimiter", func() {
				script := parse(`""""here is """ some text""""`)
				Expect(toTree(script)).To(Equal([]MS{
					[]MW{
						[]MM{
							M_HERE_STRING{`here is """ some text`},
						},
					},
				}))
			})
			Specify("4-quote sequence between 3-quote delimiters", func() {
				script := parse(
					`""" <- 3 quotes here / 4 quotes there -> """" / 3 quotes here -> """`,
				)
				Expect(toTree(script)).To(Equal([]MS{
					[]MW{
						[]MM{
							M_HERE_STRING{
								` <- 3 quotes here / 4 quotes there -> """" / 3 quotes here -> `,
							},
						},
					},
				}))
			})
			Describe("exceptions", func() {
				Specify("unterminated here-string", func() {
					tokens := tokenizer.Tokenize(`"""hello`)
					Expect(parser.Parse(tokens)).To(Equal(
						PARSE_ERROR("unmatched here-string delimiter"),
					))
				})
				Specify("extra quotes", func() {
					tokens := tokenizer.Tokenize(
						`""" <- 3 quotes here / 4 quotes there -> """"`,
					)
					Expect(parser.Parse(tokens)).To(Equal(
						PARSE_ERROR("unmatched here-string delimiter"),
					))
				})
			})
		})
		Describe("tagged strings", func() {
			Specify("empty tagged string", func() {
				script := parse("\"\"EOF\nEOF\"\"")
				Expect(toTree(script)).To(Equal([]MS{[]MW{[]MM{M_TAGGED_STRING{""}}}}))
			})
			Specify("single empty line", func() {
				script := parse("\"\"EOF\n\nEOF\"\"")
				Expect(toTree(script)).To(Equal([]MS{[]MW{[]MM{M_TAGGED_STRING{"\n"}}}}))
			})
			Specify("extra characters after open delimiter", func() {
				script := parse(
					"\"\"EOF some $arbitrary[ }text\\\n (with continuation\nfoo\nbar\nEOF\"\"",
				)
				Expect(toTree(script)).To(Equal([]MS{[]MW{[]MM{M_TAGGED_STRING{"foo\nbar\n"}}}}))
			})
			Specify("tag within string", func() {
				script := parse("\"\"EOF\nEOF \"\"\nEOF\"\"")
				Expect(toTree(script)).To(Equal([]MS{[]MW{[]MM{M_TAGGED_STRING{"EOF \"\"\n"}}}}))
			})
			Specify("continuations", func() {
				script := parse("\"\"EOF\nsome\\\n   string\nEOF\"\"")
				Expect(toTree(script)).To(Equal([]MS{
					[]MW{[]MM{M_TAGGED_STRING{"some\\\n   string\n"}}},
				}))
			})
			Specify("indentation", func() {
				script := parse(`""EOF
			          #include <stdio.h>

			          int main(void) {
			            printf("Hello, world!");
			            return 0;
			          }
			          EOF""`)
				Expect(toTree(script)).To(Equal([]MS{
					[]MW{
						[]MM{
							M_TAGGED_STRING{`#include <stdio.h>

int main(void) {
  printf("Hello, world!");
  return 0;
}
`},
						},
					},
				}))
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
				Expect(toTree(script)).To(Equal([]MS{
					[]MW{
						[]MM{
							M_TAGGED_STRING{`#include <stdio.h>

int main(void) {
  printf("Hello, world!");
  return 0;
}
`},
						},
					},
				}))
			})
			Specify("prefix with shorter lines", func() {
				script := parse(`""TAG
			          $ prompt

			          > result
			          > TAG""`)
				Expect(toTree(script)).To(Equal([]MS{
					[]MW{[]MM{M_TAGGED_STRING{"prompt\n\nresult\n"}}},
				}))
			})
			Describe("exceptions", func() {
				Specify("unterminated tagged string", func() {
					tokens := tokenizer.Tokenize("\"\"EOF\nhello")
					Expect(parser.Parse(tokens)).To(Equal(
						PARSE_ERROR("unmatched tagged string delimiter"),
					))
				})
				Specify("extra quotes", func() {
					tokens := tokenizer.Tokenize("\"\"EOF\nhello\nEOF\"\"\"")
					Expect(parser.Parse(tokens)).To(Equal(
						PARSE_ERROR("unmatched tagged string delimiter"),
					))
				})
			})
		})
		Describe("line comments", func() {
			Specify("empty line comment", func() {
				script := parse("#")
				Expect(toTree(script)).To(Equal([]MS{[]MW{[]MM{M_LINE_COMMENT{""}}}}))
			})
			Specify("simple line comment", func() {
				script := parse("# this is a comment")
				Expect(toTree(script)).To(Equal([]MS{
					[]MW{[]MM{M_LINE_COMMENT{" this is a comment"}}},
				}))
			})
			Specify("line comment with special characters", func() {
				script := parse("# this ; is$ (a [comment{")
				Expect(toTree(script)).To(Equal([]MS{
					[]MW{[]MM{M_LINE_COMMENT{" this ; is$ (a [comment{"}}},
				}))
			})
			Specify("line comment with continuation", func() {
				script := parse("# this is\\\na comment")
				Expect(toTree(script)).To(Equal([]MS{
					[]MW{[]MM{M_LINE_COMMENT{" this is a comment"}}},
				}))
			})
			Specify("line comment with escapes", func() {
				script := parse("# hello \\x41\\t")
				Expect(toTree(script)).To(Equal([]MS{
					[]MW{[]MM{M_LINE_COMMENT{" hello A\t"}}},
				}))
			})
			Specify("line comment with multiple hashes", func() {
				script := parse("### this is a comment")
				Expect(toTree(script)).To(Equal([]MS{
					[]MW{[]MM{M_LINE_COMMENT{" this is a comment"}}},
				}))
			})
		})
		Describe("block comments", func() {
			Specify("empty block comment", func() {
				script := parse("#{}#")
				Expect(toTree(script)).To(Equal([]MS{[]MW{[]MM{M_BLOCK_COMMENT{""}}}}))
			})
			Specify("simple block comment", func() {
				script := parse("#{comment}#")
				Expect(toTree(script)).To(Equal([]MS{[]MW{[]MM{M_BLOCK_COMMENT{"comment"}}}}))
			})
			Specify("multiple line block comment", func() {
				script := parse("#{\ncomment\n}#")
				Expect(toTree(script)).To(Equal([]MS{[]MW{[]MM{M_BLOCK_COMMENT{"\ncomment\n"}}}}))
			})
			Specify("block comment with continuation", func() {
				script := parse("#{this is\\\na comment}#")
				Expect(toTree(script)).To(Equal([]MS{
					[]MW{[]MM{M_BLOCK_COMMENT{"this is\\\na comment"}}},
				}))
			})
			Specify("block comment with escapes", func() {
				script := parse("#{hello \\x41\\t}#")
				Expect(toTree(script)).To(Equal([]MS{
					[]MW{[]MM{M_BLOCK_COMMENT{"hello \\x41\\t"}}},
				}))
			})
			Specify("block comment with multiple hashes", func() {
				script := parse("##{comment}##")
				Expect(toTree(script)).To(Equal([]MS{[]MW{[]MM{M_BLOCK_COMMENT{"comment"}}}}))
			})
			Specify("nested block comments", func() {
				script := parse("##{comment ##{}##}##")
				Expect(toTree(script)).To(Equal([]MS{
					[]MW{[]MM{M_BLOCK_COMMENT{"comment ##{}##"}}},
				}))
			})
			Specify("nested block comments with different prefixes", func() {
				script := parse("##{comment #{}##")
				Expect(toTree(script)).To(Equal([]MS{[]MW{[]MM{M_BLOCK_COMMENT{"comment #{"}}}}))
			})
			Describe("exceptions", func() {
				Specify("unterminated block comment", func() {
					tokens := tokenizer.Tokenize("#{hello")
					Expect(parser.Parse(tokens)).To(Equal(
						PARSE_ERROR("unmatched block comment delimiter"),
					))
				})
				Specify("extra hashes", func() {
					tokens := tokenizer.Tokenize(
						"#{ <- 1 hash here / 2 hashes there -> }##",
					)
					Expect(parser.Parse(tokens)).To(Equal(
						PARSE_ERROR("unmatched block comment delimiter"),
					))
				})
			})
		})
		Describe("substitutions", func() {
			Specify("lone dollar", func() {
				script := parse("$")
				Expect(toTree(script)).To(Equal([]MS{[]MW{[]MM{M_LITERAL{"$"}}}}))
			})
			Specify("simple variable", func() {
				script := parse("$a")
				Expect(toTree(script)).To(Equal([]MS{
					[]MW{[]MM{M_SUBSTITUTE_NEXT{"$"}, M_LITERAL{"a"}}},
				}))
			})
			Specify("Unicode variable name", func() {
				script := parse("$a\u1234")
				Expect(toTree(script)).To(Equal([]MS{
					[]MW{[]MM{M_SUBSTITUTE_NEXT{"$"}, M_LITERAL{"a\u1234"}}},
				}))
			})
			Specify("tuple", func() {
				script := parse("$(a)")
				Expect(toTree(script)).To(Equal([]MS{
					[]MW{[]MM{M_SUBSTITUTE_NEXT{"$"}, M_TUPLE{[]MS{[]MW{[]MM{M_LITERAL{"a"}}}}}}},
				}))
			})
			Specify("block", func() {
				script := parse("${a}")
				Expect(toTree(script)).To(Equal([]MS{
					[]MW{[]MM{M_SUBSTITUTE_NEXT{"$"}, M_BLOCK{[]MS{[]MW{[]MM{M_LITERAL{"a"}}}}}}},
				}))
			})
			Specify("expression", func() {
				script := parse("$[a]")
				Expect(toTree(script)).To(Equal([]MS{
					[]MW{[]MM{M_SUBSTITUTE_NEXT{"$"}, M_EXPRESSION{[]MS{[]MW{[]MM{M_LITERAL{"a"}}}}}}},
				}))
			})
			Specify("multiple substitution", func() {
				script := parse("$$a $$$b $$$$[c]")
				Expect(toTree(script)).To(Equal([]MS{
					[]MW{
						[]MM{M_SUBSTITUTE_NEXT{"$"},
							M_SUBSTITUTE_NEXT{"$"}, M_LITERAL{"a"}},
						[]MM{M_SUBSTITUTE_NEXT{"$"},
							M_SUBSTITUTE_NEXT{"$"},
							M_SUBSTITUTE_NEXT{"$"}, M_LITERAL{"b"}},
						[]MM{
							M_SUBSTITUTE_NEXT{"$"},
							M_SUBSTITUTE_NEXT{"$"},
							M_SUBSTITUTE_NEXT{"$"},
							M_SUBSTITUTE_NEXT{"$"},
							M_EXPRESSION{[]MS{[]MW{[]MM{M_LITERAL{"c"}}}}},
						},
					},
				}))
			})
			Specify("expansion", func() {
				script := parse("$*$$*a $*$[b]")
				Expect(toTree(script)).To(Equal([]MS{
					[]MW{
						[]MM{
							M_EXPAND_NEXT{"$*"},
							M_SUBSTITUTE_NEXT{"$"},
							M_SUBSTITUTE_NEXT{"$*"},
							M_LITERAL{"a"},
						},
						[]MM{
							M_EXPAND_NEXT{"$*"},
							M_SUBSTITUTE_NEXT{"$"},
							M_EXPRESSION{[]MS{[]MW{[]MM{M_LITERAL{"b"}}}}},
						},
					},
				}))
			})
			Describe("variable name delimiters", func() {
				Specify("trailing dollars", func() {
					script := parse("a$ b$*$ c$$*$")
					Expect(toTree(script)).To(Equal([]MS{
						[]MW{
							[]MM{M_LITERAL{"a$"}},
							[]MM{M_LITERAL{"b$*$"}},
							[]MM{M_LITERAL{"c$$*$"}},
						},
					}))
				})
				Specify("escapes", func() {
					script := parse("$a\\x62 $c\\d")
					Expect(toTree(script)).To(Equal([]MS{
						[]MW{
							[]MM{M_SUBSTITUTE_NEXT{"$"}, M_LITERAL{"a"}, M_LITERAL{"b"}},
							[]MM{M_SUBSTITUTE_NEXT{"$"}, M_LITERAL{"c"}, M_LITERAL{"d"}},
						},
					}))
				})
				Specify("special characters", func() {
					script := parse("$a# $b*")
					Expect(toTree(script)).To(Equal([]MS{
						[]MW{
							[]MM{M_SUBSTITUTE_NEXT{"$"}, M_LITERAL{"a"}, M_LITERAL{"#"}},
							[]MM{M_SUBSTITUTE_NEXT{"$"}, M_LITERAL{"b"}, M_LITERAL{"*"}},
						},
					}))
				})
				Describe("exceptions", func() {
					Specify("leading hash", func() {
						tokens := tokenizer.Tokenize("$#")
						Expect(parser.Parse(tokens)).To(Equal(
							PARSE_ERROR("unexpected comment delimiter"),
						))
					})
					Specify("leading quote", func() {
						tokens := tokenizer.Tokenize(`$"`)
						Expect(parser.Parse(tokens)).To(Equal(
							PARSE_ERROR("unexpected string delimiter"),
						))
					})
				})
			})
			Describe("selectors", func() {
				Describe("indexed selectors", func() {
					Specify("single", func() {
						script := parse("$name[index1] $[expression][index2]")
						Expect(toTree(script)).To(Equal([]MS{
							[]MW{
								[]MM{
									M_SUBSTITUTE_NEXT{"$"},
									M_LITERAL{"name"},
									M_EXPRESSION{[]MS{[]MW{[]MM{M_LITERAL{"index1"}}}}},
								},
								[]MM{
									M_SUBSTITUTE_NEXT{"$"},
									M_EXPRESSION{[]MS{[]MW{[]MM{M_LITERAL{"expression"}}}}},
									M_EXPRESSION{[]MS{[]MW{[]MM{M_LITERAL{"index2"}}}}},
								},
							},
						}))
					})
					Specify("chained", func() {
						script := parse(
							"$name[index1][index2][index3] $[expression][index4][index5][index6]",
						)
						Expect(toTree(script)).To(Equal([]MS{
							[]MW{
								[]MM{
									M_SUBSTITUTE_NEXT{"$"},
									M_LITERAL{"name"},
									M_EXPRESSION{[]MS{[]MW{[]MM{M_LITERAL{"index1"}}}}},
									M_EXPRESSION{[]MS{[]MW{[]MM{M_LITERAL{"index2"}}}}},
									M_EXPRESSION{[]MS{[]MW{[]MM{M_LITERAL{"index3"}}}}},
								},
								[]MM{
									M_SUBSTITUTE_NEXT{"$"},
									M_EXPRESSION{[]MS{[]MW{[]MM{M_LITERAL{"expression"}}}}},
									M_EXPRESSION{[]MS{[]MW{[]MM{M_LITERAL{"index4"}}}}},
									M_EXPRESSION{[]MS{[]MW{[]MM{M_LITERAL{"index5"}}}}},
									M_EXPRESSION{[]MS{[]MW{[]MM{M_LITERAL{"index6"}}}}},
								},
							},
						}))
					})
				})
				Describe("keyed selectors", func() {
					Specify("single", func() {
						script := parse("$name(key1) $[expression](key2)")
						Expect(toTree(script)).To(Equal([]MS{
							[]MW{
								[]MM{
									M_SUBSTITUTE_NEXT{"$"},
									M_LITERAL{"name"},
									M_TUPLE{[]MS{[]MW{[]MM{M_LITERAL{"key1"}}}}},
								},
								[]MM{
									M_SUBSTITUTE_NEXT{"$"},
									M_EXPRESSION{[]MS{[]MW{[]MM{M_LITERAL{"expression"}}}}},
									M_TUPLE{[]MS{[]MW{[]MM{M_LITERAL{"key2"}}}}},
								},
							},
						}))
					})
					Specify("multiple", func() {
						script := parse("$name(key1 key2) $[expression](key3 key4)")
						Expect(toTree(script)).To(Equal([]MS{
							[]MW{
								[]MM{
									M_SUBSTITUTE_NEXT{"$"},
									M_LITERAL{"name"},
									M_TUPLE{[]MS{[]MW{[]MM{M_LITERAL{"key1"}}, []MM{M_LITERAL{"key2"}}}}},
								},
								[]MM{
									M_SUBSTITUTE_NEXT{"$"},
									M_EXPRESSION{[]MS{[]MW{[]MM{M_LITERAL{"expression"}}}}},
									M_TUPLE{[]MS{[]MW{[]MM{M_LITERAL{"key3"}}, []MM{M_LITERAL{"key4"}}}}},
								},
							},
						}))
					})
					Specify("chained", func() {
						script := parse(
							"$name(key1)(key2 key3)(key4) $[expression](key5 key6)(key7)(key8 key9)",
						)
						Expect(toTree(script)).To(Equal([]MS{
							[]MW{
								[]MM{
									M_SUBSTITUTE_NEXT{"$"},
									M_LITERAL{"name"},
									M_TUPLE{[]MS{[]MW{[]MM{M_LITERAL{"key1"}}}}},
									M_TUPLE{[]MS{[]MW{[]MM{M_LITERAL{"key2"}}, []MM{M_LITERAL{"key3"}}}}},
									M_TUPLE{[]MS{[]MW{[]MM{M_LITERAL{"key4"}}}}},
								},
								[]MM{
									M_SUBSTITUTE_NEXT{"$"},
									M_EXPRESSION{[]MS{[]MW{[]MM{M_LITERAL{"expression"}}}}},
									M_TUPLE{[]MS{[]MW{[]MM{M_LITERAL{"key5"}}, []MM{M_LITERAL{"key6"}}}}},
									M_TUPLE{[]MS{[]MW{[]MM{M_LITERAL{"key7"}}}}},
									M_TUPLE{[]MS{[]MW{[]MM{M_LITERAL{"key8"}}, []MM{M_LITERAL{"key9"}}}}},
								},
							},
						}))
					})
				})
				Describe("generic selectors", func() {
					Specify("single", func() {
						script := parse("$name{selector1} $[expression]{selector2}")
						Expect(toTree(script)).To(Equal([]MS{
							[]MW{
								[]MM{
									M_SUBSTITUTE_NEXT{"$"},
									M_LITERAL{"name"},
									M_BLOCK{[]MS{[]MW{[]MM{M_LITERAL{"selector1"}}}}},
								},
								[]MM{
									M_SUBSTITUTE_NEXT{"$"},
									M_EXPRESSION{[]MS{[]MW{[]MM{M_LITERAL{"expression"}}}}},
									M_BLOCK{[]MS{[]MW{[]MM{M_LITERAL{"selector2"}}}}},
								},
							},
						}))
					})
					Specify("chained", func() {
						script := parse(
							"$name{selector1}{selector2 arg1}{selector3} $[expression]{selector4 arg2 arg3}{selector5}{selector6 arg4}",
						)
						Expect(toTree(script)).To(Equal([]MS{
							[]MW{
								[]MM{
									M_SUBSTITUTE_NEXT{"$"},
									M_LITERAL{"name"},
									M_BLOCK{[]MS{[]MW{[]MM{M_LITERAL{"selector1"}}}}},
									M_BLOCK{[]MS{
										[]MW{[]MM{M_LITERAL{"selector2"}}, []MM{M_LITERAL{"arg1"}}},
									}},
									M_BLOCK{[]MS{[]MW{[]MM{M_LITERAL{"selector3"}}}}},
								},
								[]MM{
									M_SUBSTITUTE_NEXT{"$"},
									M_EXPRESSION{[]MS{[]MW{[]MM{M_LITERAL{"expression"}}}}},
									M_BLOCK{[]MS{
										[]MW{
											[]MM{M_LITERAL{"selector4"}},
											[]MM{M_LITERAL{"arg2"}},
											[]MM{M_LITERAL{"arg3"}},
										},
									},
									},
									M_BLOCK{[]MS{[]MW{[]MM{M_LITERAL{"selector5"}}}}},
									M_BLOCK{[]MS{
										[]MW{[]MM{M_LITERAL{"selector6"}}, []MM{M_LITERAL{"arg4"}}},
									}},
								},
							},
						}))
					})
				})
				Specify("mixed selectors", func() {
					script := parse(
						"$name(key1 key2){selector1}(key3){selector2 selector3} $[expression]{selector4 selector5}(key4 key5){selector6}(key6)",
					)
					Expect(toTree(script)).To(Equal([]MS{
						[]MW{
							[]MM{
								M_SUBSTITUTE_NEXT{"$"},
								M_LITERAL{"name"},
								M_TUPLE{[]MS{[]MW{[]MM{M_LITERAL{"key1"}}, []MM{M_LITERAL{"key2"}}}}},
								M_BLOCK{[]MS{[]MW{[]MM{M_LITERAL{"selector1"}}}}},
								M_TUPLE{[]MS{[]MW{[]MM{M_LITERAL{"key3"}}}}},
								M_BLOCK{[]MS{
									[]MW{[]MM{M_LITERAL{"selector2"}}, []MM{M_LITERAL{"selector3"}}},
								}},
							},
							[]MM{
								M_SUBSTITUTE_NEXT{"$"},
								M_EXPRESSION{[]MS{[]MW{[]MM{M_LITERAL{"expression"}}}}},
								M_BLOCK{[]MS{
									[]MW{[]MM{M_LITERAL{"selector4"}}, []MM{M_LITERAL{"selector5"}}},
								}},
								M_TUPLE{[]MS{[]MW{[]MM{M_LITERAL{"key4"}}, []MM{M_LITERAL{"key5"}}}}},
								M_BLOCK{[]MS{[]MW{[]MM{M_LITERAL{"selector6"}}}}},
								M_TUPLE{[]MS{[]MW{[]MM{M_LITERAL{"key6"}}}}},
							},
						},
					}))
				})
				Specify("nested selectors", func() {
					script := parse(
						"$name1(key1 $name2{selector1} $[expression1](key2)) $[expression2]{selector2 $name3(key3)}",
					)
					Expect(toTree(script)).To(Equal([]MS{
						[]MW{
							[]MM{
								M_SUBSTITUTE_NEXT{"$"},
								M_LITERAL{"name1"},
								M_TUPLE{[]MS{
									[]MW{
										[]MM{M_LITERAL{"key1"}},
										[]MM{
											M_SUBSTITUTE_NEXT{"$"},
											M_LITERAL{"name2"},
											M_BLOCK{[]MS{[]MW{[]MM{M_LITERAL{"selector1"}}}}},
										},
										[]MM{
											M_SUBSTITUTE_NEXT{"$"},
											M_EXPRESSION{[]MS{[]MW{[]MM{M_LITERAL{"expression1"}}}}},
											M_TUPLE{[]MS{[]MW{[]MM{M_LITERAL{"key2"}}}}},
										},
									},
								}},
							},
							[]MM{
								M_SUBSTITUTE_NEXT{"$"},
								M_EXPRESSION{[]MS{[]MW{[]MM{M_LITERAL{"expression2"}}}}},
								M_BLOCK{[]MS{
									[]MW{
										[]MM{M_LITERAL{"selector2"}},
										[]MM{
											M_SUBSTITUTE_NEXT{"$"},
											M_LITERAL{"name3"},
											M_TUPLE{[]MS{[]MW{[]MM{M_LITERAL{"key3"}}}}},
										},
									},
								}},
							},
						},
					}))
				})
			})
		})
		Describe("qualified words", func() {
			Describe("indexed selectors", func() {
				Specify("single", func() {
					script := parse("name[index]")
					Expect(toTree(script)).To(Equal([]MS{
						[]MW{[]MM{M_LITERAL{"name"}, M_EXPRESSION{[]MS{[]MW{[]MM{M_LITERAL{"index"}}}}}}},
					}))
				})
				Specify("chained", func() {
					script := parse("name[index1][index2][index3]")
					Expect(toTree(script)).To(Equal([]MS{
						[]MW{
							[]MM{
								M_LITERAL{"name"},
								M_EXPRESSION{[]MS{[]MW{[]MM{M_LITERAL{"index1"}}}}},
								M_EXPRESSION{[]MS{[]MW{[]MM{M_LITERAL{"index2"}}}}},
								M_EXPRESSION{[]MS{[]MW{[]MM{M_LITERAL{"index3"}}}}},
							},
						},
					}))
				})
			})
			Describe("keyed selectors", func() {
				Specify("single", func() {
					script := parse("name(key)")
					Expect(toTree(script)).To(Equal([]MS{
						[]MW{[]MM{M_LITERAL{"name"}, M_TUPLE{[]MS{[]MW{[]MM{M_LITERAL{"key"}}}}}}},
					}))
				})
				Specify("multiple", func() {
					script := parse("name(key1 key2)")
					Expect(toTree(script)).To(Equal([]MS{
						[]MW{
							[]MM{
								M_LITERAL{"name"},
								M_TUPLE{[]MS{[]MW{[]MM{M_LITERAL{"key1"}}, []MM{M_LITERAL{"key2"}}}}},
							},
						},
					}))
				})
				Specify("chained", func() {
					script := parse("name(key1)(key2 key3)(key4)")
					Expect(toTree(script)).To(Equal([]MS{
						[]MW{
							[]MM{
								M_LITERAL{"name"},
								M_TUPLE{[]MS{[]MW{[]MM{M_LITERAL{"key1"}}}}},
								M_TUPLE{[]MS{[]MW{[]MM{M_LITERAL{"key2"}}, []MM{M_LITERAL{"key3"}}}}},
								M_TUPLE{[]MS{[]MW{[]MM{M_LITERAL{"key4"}}}}},
							},
						},
					}))
				})
			})
			Describe("generic selectors", func() {
				Specify("single", func() {
					script := parse("name{selector}")
					Expect(toTree(script)).To(Equal([]MS{
						[]MW{[]MM{M_LITERAL{"name"}, M_BLOCK{[]MS{[]MW{[]MM{M_LITERAL{"selector"}}}}}}},
					}))
				})
				Specify("multiple", func() {
					script := parse("name{selector1 selector2}")
					Expect(toTree(script)).To(Equal([]MS{
						[]MW{
							[]MM{
								M_LITERAL{"name"},
								M_BLOCK{[]MS{
									[]MW{[]MM{M_LITERAL{"selector1"}}, []MM{M_LITERAL{"selector2"}}},
								}},
							},
						},
					}))
				})
				Specify("chained", func() {
					script := parse(
						"name{selector1}{selector2 selector3}{selector4}",
					)
					Expect(toTree(script)).To(Equal([]MS{
						[]MW{
							[]MM{
								M_LITERAL{"name"},
								M_BLOCK{[]MS{[]MW{[]MM{M_LITERAL{"selector1"}}}}},
								M_BLOCK{[]MS{
									[]MW{[]MM{M_LITERAL{"selector2"}}, []MM{M_LITERAL{"selector3"}}},
								}},
								M_BLOCK{[]MS{[]MW{[]MM{M_LITERAL{"selector4"}}}}},
							},
						},
					}))
				})
			})
			Specify("mixed selectors", func() {
				script := parse(
					"name(key1 key2){selector1}(key3){selector2 selector3}",
				)
				Expect(toTree(script)).To(Equal([]MS{
					[]MW{
						[]MM{
							M_LITERAL{"name"},
							M_TUPLE{[]MS{[]MW{[]MM{M_LITERAL{"key1"}}, []MM{M_LITERAL{"key2"}}}}},
							M_BLOCK{[]MS{[]MW{[]MM{M_LITERAL{"selector1"}}}}},
							M_TUPLE{[]MS{[]MW{[]MM{M_LITERAL{"key3"}}}}},
							M_BLOCK{[]MS{
								[]MW{[]MM{M_LITERAL{"selector2"}}, []MM{M_LITERAL{"selector3"}}},
							}},
						},
					},
				}))
			})
			Specify("nested selectors", func() {
				script := parse("name1(key1 name2{selector1})")
				Expect(toTree(script)).To(Equal([]MS{
					[]MW{
						[]MM{
							M_LITERAL{"name1"},
							M_TUPLE{[]MS{
								[]MW{
									[]MM{M_LITERAL{"key1"}},
									[]MM{
										M_LITERAL{"name2"},
										M_BLOCK{[]MS{[]MW{[]MM{M_LITERAL{"selector1"}}}}},
									},
								},
							}},
						},
					},
				}))
			})
		})
	})
})
