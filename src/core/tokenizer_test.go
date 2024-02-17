package core_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "helena/core"
)

func toType(token Token) TokenType  { return token.Type }
func toIndex(token Token) uint      { return token.Position.Index }
func toLine(token Token) uint       { return token.Position.Line }
func toColumn(token Token) uint     { return token.Position.Column }
func toLiteral(token Token) string  { return token.Literal }
func toSequence(token Token) string { return token.Sequence }

func mapToType(tokens []Token) []TokenType {
	result := make([]TokenType, len(tokens))
	for i, token := range tokens {
		result[i] = toType(token)
	}
	return result
}
func mapToLiteral(tokens []Token) []string {
	result := make([]string, len(tokens))
	for i, token := range tokens {
		result[i] = toLiteral(token)
	}
	return result
}
func mapToIndex(tokens []Token) []uint {
	result := make([]uint, len(tokens))
	for i, token := range tokens {
		result[i] = toIndex(token)
	}
	return result
}
func mapToLine(tokens []Token) []uint {
	result := make([]uint, len(tokens))
	for i, token := range tokens {
		result[i] = toLine(token)
	}
	return result
}
func mapToColumn(tokens []Token) []uint {
	result := make([]uint, len(tokens))
	for i, token := range tokens {
		result[i] = toColumn(token)
	}
	return result
}
func mapToSequence(tokens []Token) []string {
	result := make([]string, len(tokens))
	for i, token := range tokens {
		result[i] = toSequence(token)
	}
	return result
}

var _ = Describe("Tokenizer", func() {
	var tokenizer Tokenizer
	BeforeEach(func() {
		tokenizer = Tokenizer{}
	})

	Specify("empty string", func() {
		Expect(tokenizer.Tokenize("")).To(BeEmpty())
	})

	Describe("types", func() {
		Specify("whitespace", func() {
			Expect(mapToType(tokenizer.Tokenize(" "))).To(Equal([]TokenType{
				TokenType_WHITESPACE,
			}))
			Expect(mapToType(tokenizer.Tokenize("\t"))).To(Equal([]TokenType{
				TokenType_WHITESPACE,
			}))
			Expect(mapToType(tokenizer.Tokenize("\r"))).To(Equal([]TokenType{
				TokenType_WHITESPACE,
			}))
			Expect(mapToType(tokenizer.Tokenize("\f"))).To(Equal([]TokenType{
				TokenType_WHITESPACE,
			}))
			Expect(mapToType(tokenizer.Tokenize("  "))).To(Equal([]TokenType{
				TokenType_WHITESPACE,
			}))
			Expect(mapToType(tokenizer.Tokenize("   \t\f  \r\r "))).To(Equal([]TokenType{
				TokenType_WHITESPACE,
			}))

			Expect(mapToType(tokenizer.Tokenize("\\ "))).To(Equal([]TokenType{TokenType_ESCAPE}))
			Expect(mapToType(tokenizer.Tokenize("\\\t"))).To(Equal([]TokenType{TokenType_ESCAPE}))
			Expect(mapToType(tokenizer.Tokenize("\\\r"))).To(Equal([]TokenType{TokenType_ESCAPE}))
			Expect(mapToType(tokenizer.Tokenize("\\\f"))).To(Equal([]TokenType{TokenType_ESCAPE}))
		})

		Specify("newline", func() {
			Expect(mapToType(tokenizer.Tokenize("\n"))).To(Equal([]TokenType{TokenType_NEWLINE}))
			Expect(mapToType(tokenizer.Tokenize("\n\n"))).To(Equal([]TokenType{
				TokenType_NEWLINE,
				TokenType_NEWLINE,
			}))
		})

		Describe("escape sequences", func() {
			Specify("backslash", func() {
				Expect(mapToType(tokenizer.Tokenize("\\"))).To(Equal([]TokenType{TokenType_TEXT}))
			})
			Specify("continuation", func() {
				Expect(mapToType(tokenizer.Tokenize("\\\n"))).To(Equal([]TokenType{
					TokenType_CONTINUATION,
				}))
				Expect(mapToType(tokenizer.Tokenize("\\\n   "))).To(Equal([]TokenType{
					TokenType_CONTINUATION,
				}))
				Expect(mapToType(tokenizer.Tokenize("\\\n \t\r\f "))).To(Equal([]TokenType{
					TokenType_CONTINUATION,
				}))
				Expect(mapToType(tokenizer.Tokenize("\\\n \t \\\n  "))).To(Equal([]TokenType{
					TokenType_CONTINUATION,
					TokenType_CONTINUATION,
				}))
			})
			Specify("control characters", func() {
				Expect(mapToType(tokenizer.Tokenize("\\a"))).To(Equal([]TokenType{
					TokenType_ESCAPE,
				}))
				Expect(mapToType(tokenizer.Tokenize("\\b"))).To(Equal([]TokenType{
					TokenType_ESCAPE,
				}))
				Expect(mapToType(tokenizer.Tokenize("\\f"))).To(Equal([]TokenType{
					TokenType_ESCAPE,
				}))
				Expect(mapToType(tokenizer.Tokenize("\\n"))).To(Equal([]TokenType{
					TokenType_ESCAPE,
				}))
				Expect(mapToType(tokenizer.Tokenize("\\r"))).To(Equal([]TokenType{
					TokenType_ESCAPE,
				}))
				Expect(mapToType(tokenizer.Tokenize("\\t"))).To(Equal([]TokenType{
					TokenType_ESCAPE,
				}))
				Expect(mapToType(tokenizer.Tokenize("\\v"))).To(Equal([]TokenType{
					TokenType_ESCAPE,
				}))
				Expect(mapToType(tokenizer.Tokenize("\\\\"))).To(Equal([]TokenType{
					TokenType_ESCAPE,
				}))
			})
			Specify("octal sequence", func() {
				Expect(mapToType(tokenizer.Tokenize("\\1"))).To(Equal([]TokenType{
					TokenType_ESCAPE,
				}))
				Expect(mapToType(tokenizer.Tokenize("\\123"))).To(Equal([]TokenType{
					TokenType_ESCAPE,
				}))
				Expect(mapToType(tokenizer.Tokenize("\\1234"))).To(Equal([]TokenType{
					TokenType_ESCAPE,
					TokenType_TEXT,
				}))
				Expect(mapToType(tokenizer.Tokenize("\\0x"))).To(Equal([]TokenType{
					TokenType_ESCAPE,
					TokenType_TEXT,
				}))
			})
			Specify("hexadecimal sequence", func() {
				Expect(mapToType(tokenizer.Tokenize("\\x1"))).To(Equal([]TokenType{
					TokenType_ESCAPE,
				}))
				Expect(mapToType(tokenizer.Tokenize("\\x12"))).To(Equal([]TokenType{
					TokenType_ESCAPE,
				}))
				Expect(mapToType(tokenizer.Tokenize("\\x123"))).To(Equal([]TokenType{
					TokenType_ESCAPE,
					TokenType_TEXT,
				}))
				Expect(mapToType(tokenizer.Tokenize("\\x1f"))).To(Equal([]TokenType{
					TokenType_ESCAPE,
				}))
				Expect(mapToType(tokenizer.Tokenize("\\x1F"))).To(Equal([]TokenType{
					TokenType_ESCAPE,
				}))
				Expect(mapToType(tokenizer.Tokenize("\\x1g"))).To(Equal([]TokenType{
					TokenType_ESCAPE,
					TokenType_TEXT,
				}))
			})
			Specify("unicode sequence", func() {
				Expect(mapToType(tokenizer.Tokenize("\\u1"))).To(Equal([]TokenType{
					TokenType_ESCAPE,
				}))
				Expect(mapToType(tokenizer.Tokenize("\\U1"))).To(Equal([]TokenType{
					TokenType_ESCAPE,
				}))
				Expect(mapToType(tokenizer.Tokenize("\\u12"))).To(Equal([]TokenType{
					TokenType_ESCAPE,
				}))
				Expect(mapToType(tokenizer.Tokenize("\\U12"))).To(Equal([]TokenType{
					TokenType_ESCAPE,
				}))
				Expect(mapToType(tokenizer.Tokenize("\\u123456"))).To(Equal([]TokenType{
					TokenType_ESCAPE,
					TokenType_TEXT,
				}))
				Expect(mapToType(tokenizer.Tokenize("\\U12345"))).To(Equal([]TokenType{
					TokenType_ESCAPE,
				}))
				Expect(mapToType(tokenizer.Tokenize("\\U0123456789"))).To(Equal([]TokenType{
					TokenType_ESCAPE,
					TokenType_TEXT,
				}))
				Expect(mapToType(tokenizer.Tokenize("\\u1f"))).To(Equal([]TokenType{
					TokenType_ESCAPE,
				}))
				Expect(mapToType(tokenizer.Tokenize("\\u1F"))).To(Equal([]TokenType{
					TokenType_ESCAPE,
				}))
				Expect(mapToType(tokenizer.Tokenize("\\U1f"))).To(Equal([]TokenType{
					TokenType_ESCAPE,
				}))
				Expect(mapToType(tokenizer.Tokenize("\\U1F"))).To(Equal([]TokenType{
					TokenType_ESCAPE,
				}))
				Expect(mapToType(tokenizer.Tokenize("\\u1g"))).To(Equal([]TokenType{
					TokenType_ESCAPE,
					TokenType_TEXT,
				}))
				Expect(mapToType(tokenizer.Tokenize("\\U1g"))).To(Equal([]TokenType{
					TokenType_ESCAPE,
					TokenType_TEXT,
				}))
			})
			Specify("unrecognized sequences", func() {
				Expect(mapToType(tokenizer.Tokenize("\\8"))).To(Equal([]TokenType{
					TokenType_ESCAPE,
				}))
				Expect(mapToType(tokenizer.Tokenize("\\9"))).To(Equal([]TokenType{
					TokenType_ESCAPE,
				}))
				Expect(mapToType(tokenizer.Tokenize("\\c"))).To(Equal([]TokenType{
					TokenType_ESCAPE,
				}))
				Expect(mapToType(tokenizer.Tokenize("\\d"))).To(Equal([]TokenType{
					TokenType_ESCAPE,
				}))
				Expect(mapToType(tokenizer.Tokenize("\\e"))).To(Equal([]TokenType{
					TokenType_ESCAPE,
				}))
				Expect(mapToType(tokenizer.Tokenize("\\x"))).To(Equal([]TokenType{
					TokenType_ESCAPE,
				}))
				Expect(mapToType(tokenizer.Tokenize("\\xg"))).To(Equal([]TokenType{
					TokenType_ESCAPE,
					TokenType_TEXT,
				}))
				Expect(mapToType(tokenizer.Tokenize("\\u"))).To(Equal([]TokenType{
					TokenType_ESCAPE,
				}))
				Expect(mapToType(tokenizer.Tokenize("\\ug"))).To(Equal([]TokenType{
					TokenType_ESCAPE,
					TokenType_TEXT,
				}))
				Expect(mapToType(tokenizer.Tokenize("\\U"))).To(Equal([]TokenType{
					TokenType_ESCAPE,
				}))
				Expect(mapToType(tokenizer.Tokenize("\\Ug"))).To(Equal([]TokenType{
					TokenType_ESCAPE,
					TokenType_TEXT,
				}))
			})
		})

		Specify("comments", func() {
			Expect(mapToType(tokenizer.Tokenize("#"))).To(Equal([]TokenType{TokenType_COMMENT}))
			Expect(mapToType(tokenizer.Tokenize("###"))).To(Equal([]TokenType{TokenType_COMMENT}))

			Expect(mapToType(tokenizer.Tokenize("\\#"))).To(Equal([]TokenType{TokenType_ESCAPE}))
		})

		Specify("tuples", func() {
			Expect(mapToType(tokenizer.Tokenize("("))).To(Equal([]TokenType{
				TokenType_OPEN_TUPLE,
			}))
			Expect(mapToType(tokenizer.Tokenize(")"))).To(Equal([]TokenType{
				TokenType_CLOSE_TUPLE,
			}))

			Expect(mapToType(tokenizer.Tokenize("\\("))).To(Equal([]TokenType{TokenType_ESCAPE}))
			Expect(mapToType(tokenizer.Tokenize("\\)"))).To(Equal([]TokenType{TokenType_ESCAPE}))
		})

		Specify("blocks", func() {
			Expect(mapToType(tokenizer.Tokenize("{"))).To(Equal([]TokenType{
				TokenType_OPEN_BLOCK,
			}))
			Expect(mapToType(tokenizer.Tokenize("}"))).To(Equal([]TokenType{
				TokenType_CLOSE_BLOCK,
			}))

			Expect(mapToType(tokenizer.Tokenize("\\{"))).To(Equal([]TokenType{TokenType_ESCAPE}))
			Expect(mapToType(tokenizer.Tokenize("\\}"))).To(Equal([]TokenType{TokenType_ESCAPE}))
		})

		Specify("expressions", func() {
			Expect(mapToType(tokenizer.Tokenize("["))).To(Equal([]TokenType{
				TokenType_OPEN_EXPRESSION,
			}))
			Expect(mapToType(tokenizer.Tokenize("]"))).To(Equal([]TokenType{
				TokenType_CLOSE_EXPRESSION,
			}))

			Expect(mapToType(tokenizer.Tokenize("\\["))).To(Equal([]TokenType{TokenType_ESCAPE}))
			Expect(mapToType(tokenizer.Tokenize("\\]"))).To(Equal([]TokenType{TokenType_ESCAPE}))
		})

		Specify("strings", func() {
			Expect(mapToType(tokenizer.Tokenize(`"`))).To(Equal([]TokenType{
				TokenType_STRING_DELIMITER,
			}))

			Expect(mapToType(tokenizer.Tokenize(`\"`))).To(Equal([]TokenType{TokenType_ESCAPE}))
		})

		Specify("dollar", func() {
			Expect(mapToType(tokenizer.Tokenize("$"))).To(Equal([]TokenType{TokenType_DOLLAR}))

			Expect(mapToType(tokenizer.Tokenize("\\$"))).To(Equal([]TokenType{TokenType_ESCAPE}))
		})

		Specify("semicolon", func() {
			Expect(mapToType(tokenizer.Tokenize(";"))).To(Equal([]TokenType{TokenType_SEMICOLON}))

			Expect(mapToType(tokenizer.Tokenize("\\;"))).To(Equal([]TokenType{TokenType_ESCAPE}))
		})

		Specify("asterisk", func() {
			Expect(mapToType(tokenizer.Tokenize("*"))).To(Equal([]TokenType{TokenType_ASTERISK}))

			Expect(mapToType(tokenizer.Tokenize("\\*"))).To(Equal([]TokenType{TokenType_ESCAPE}))
		})
	})

	Describe("positions", func() {
		It("should track index", func() {
			Expect(mapToIndex(tokenizer.Tokenize("a b c"))).To(Equal([]uint{0, 1, 2, 3, 4}))
			Expect(mapToIndex(tokenizer.Tokenize("abc \r\f de\tf"))).To(Equal([]uint{
				0, 3, 7, 9, 10,
			}))
		})
		It("should track line", func() {
			Expect(mapToLine(tokenizer.Tokenize("a b c"))).To(Equal([]uint{0, 0, 0, 0, 0}))
			Expect(mapToLine(tokenizer.Tokenize("a\nbcd e\nfg  h"))).To(Equal([]uint{
				0, 0, 1, 1, 1, 1, 2, 2, 2,
			}))
		})
		It("should track column", func() {
			Expect(mapToColumn(tokenizer.Tokenize("a b c"))).To(Equal([]uint{0, 1, 2, 3, 4}))
			Expect(mapToColumn(tokenizer.Tokenize("a\nbcd e\nfg  h"))).To(Equal([]uint{
				0, 1, 0, 3, 4, 5, 0, 2, 4,
			}))
		})
	})

	Describe("literals", func() {
		Specify("text", func() {
			Expect(mapToLiteral(tokenizer.Tokenize("abcd"))).To(Equal([]string{"abcd"}))
		})
		Specify("escape", func() {
			Expect(
				mapToLiteral(tokenizer.Tokenize("\\a\\b\\f\\n\\r\\t\\v\\\\")),
			).To(Equal([]string{
				"\x07", "\b", "\f", "\n", "\r", "\t", "\v", "\\",
			}))
			Expect(mapToLiteral(tokenizer.Tokenize("\\123"))).To(Equal([]string{"S"}))
			Expect(mapToLiteral(tokenizer.Tokenize("\\xA5"))).To(Equal([]string{"¥"}))
			Expect(mapToLiteral(tokenizer.Tokenize("\\u1234"))).To(Equal([]string{"\u1234"}))
			Expect(
				mapToLiteral(tokenizer.Tokenize("\\U00012345\\U0006789A")),
			).To(Equal([]string{string(rune(0x12345)), string(rune(0x6789a))}))
			Expect(
				mapToLiteral(tokenizer.Tokenize("\\8\\9\\c\\d\\e\\x\\xg\\u\\ug\\U\\Ug")),
			).To(Equal([]string{
				"8",
				"9",
				"c",
				"d",
				"e",
				"x",
				"x",
				"g",
				"u",
				"u",
				"g",
				"U",
				"U",
				"g",
			}))
		})
		Specify("continuation", func() {
			Expect(mapToLiteral(tokenizer.Tokenize("\\\n"))).To(Equal([]string{" "}))
			Expect(mapToLiteral(tokenizer.Tokenize("\\\n   "))).To(Equal([]string{" "}))
			Expect(mapToLiteral(tokenizer.Tokenize("\\\n \t\r\f "))).To(Equal([]string{" "}))
			Expect(mapToLiteral(tokenizer.Tokenize("\\\n \t \\\n  "))).To(Equal([]string{
				" ",
				" ",
			}))
		})
	})
	Describe("sequences", func() {
		Specify("text", func() {
			Expect(mapToSequence(tokenizer.Tokenize("abcd"))).To(Equal([]string{"abcd"}))
		})
		Specify("escape", func() {
			Expect(
				mapToSequence(tokenizer.Tokenize("\\a\\b\\f\\n\\r\\t\\v\\\\")),
			).To(Equal([]string{"\\a", "\\b", "\\f", "\\n", "\\r", "\\t", "\\v", "\\\\"}))
			Expect(mapToSequence(tokenizer.Tokenize("\\123"))).To(Equal([]string{"\\123"}))
			Expect(mapToSequence(tokenizer.Tokenize("\\xA5"))).To(Equal([]string{"\\xA5"}))
			Expect(mapToSequence(tokenizer.Tokenize("\\u1234"))).To(Equal([]string{"\\u1234"}))
			Expect(
				mapToSequence(tokenizer.Tokenize("\\U00012345\\U0006789A")),
			).To(Equal([]string{"\\U00012345", "\\U0006789A"}))
		})
		Specify("continuation", func() {
			Expect(mapToSequence(tokenizer.Tokenize("\\\n"))).To(Equal([]string{"\\\n"}))
			Expect(mapToSequence(tokenizer.Tokenize("\\\n   "))).To(Equal([]string{"\\\n   "}))
			Expect(mapToSequence(tokenizer.Tokenize("\\\n \t\r\f "))).To(Equal([]string{
				"\\\n \t\r\f ",
			}))
			Expect(mapToSequence(tokenizer.Tokenize("\\\n \t \\\n  "))).To(Equal([]string{
				"\\\n \t ",
				"\\\n  ",
			}))
		})
	})

	Specify("incremental", func() {
		source := "foo (bar) \\\n {$baz [sprong]}"
		tokens := tokenizer.Tokenize(source)
		incrementalTokens := []Token{}
		input := NewStringStream(source)
		tokenizer.Begin(input)
		for !tokenizer.End() {
			incrementalTokens = append(incrementalTokens, *tokenizer.Next())
		}
		Expect(incrementalTokens).To(Equal(tokens))
	})
})
