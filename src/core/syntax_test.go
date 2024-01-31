package core_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "helena/core"
)

// import { MorphemeType, Script, SyntaxChecker, Word, WordType } from "./syntax";
// import { Parser } from "./parser";
// import { Tokenizer } from "./tokenizer";

const LITERAL = "literal"
const TUPLE = "(word1 word2)"
const BLOCK = "{word1 word2}"
const EXPRESSION = "[word1 word2]"
const STRING = `"$some [string]"`
const HERE_STRING = `"""some here "" string"""`
const TAGGED_STRING = "\"\"TAG\nsome tagged \" string\nTAG\"\""
const LINE_COMMENT = "# some [{comment"
const BLOCK_COMMENT = "#{ some [block {comment }#"

type TestMorpheme struct {
	type_ string
	value string
}

var roots = []TestMorpheme{
	{"literal", LITERAL},
	{"tuple", TUPLE},
	{"block", BLOCK},
	{"expression", EXPRESSION},
}
var qualifiedSources = []TestMorpheme{
	{"literal", LITERAL},
	{"tuple", TUPLE},
	{"block", BLOCK},
}
var monomorphemes = []TestMorpheme{
	{"string", STRING},
	{"here-string", HERE_STRING},
	{"tagged string", TAGGED_STRING},
}
var ignored = []TestMorpheme{
	{"line comment", LINE_COMMENT},
	{"block comment", BLOCK_COMMENT},
}

var _ = Describe("SyntaxChecker", func() {
	//   let tokenizer: Tokenizer;
	//   let parser: Parser;
	var checker SyntaxChecker

	//   const parse = (script: string) =>
	//     parser.parse(tokenizer.tokenize(script)).script;
	//   const firstWord = (script: Script) => script.sentences[0].words[0] as Word;

	BeforeEach(func() {
		//     tokenizer = new Tokenizer();
		//     parser = new Parser();
		checker = SyntaxChecker{}
	})

	Describe("roots", func() {
		for _, v := range append(roots, monomorphemes...) {
			Specify(v.type_+" root", func() {
				//	      const script = parse(value);
				//	      const word = firstWord(script);
				//	      Expect(word.morphemes).to.have.length(1);
				//	      Expect(checker.checkWord(word)).to.eq(WordType.ROOT);
			})
		}
	})
	//
	Describe("compounds", func() {
		Specify("literal prefix", func() {
			//	    const script = parse(LITERAL + "$" + BLOCK);
			//	    const word = firstWord(script);
			//	    Expect(word.morphemes).to.have.length(3);
			//	    Expect(checker.checkWord(word)).to.eq(WordType.COMPOUND);
		})
		Specify("expression prefix", func() {
			//	    const script = parse(EXPRESSION + LITERAL);
			//	    const word = firstWord(script);
			//	    Expect(word.morphemes).to.have.length(2);
			//	    Expect(checker.checkWord(word)).to.eq(WordType.COMPOUND);
		})
		Specify("substitution prefix", func() {
			//	    const script = parse("$" + BLOCK + TUPLE + LITERAL);
			//	    const word = firstWord(script);
			//	    Expect(word.morphemes).to.have.length(4);
			//	    Expect(checker.checkWord(word)).to.eq(WordType.COMPOUND);
		})
		Specify("complex case", func() {
			//	    const script = parse(LITERAL + "$" + BLOCK + EXPRESSION + "$" + LITERAL);
			//	    const word = firstWord(script);
			//	    Expect(word.morphemes).to.have.length(6);
			//	    Expect(checker.checkWord(word)).to.eq(WordType.COMPOUND);
		})
		Describe("exceptions", func() {
			for _, v := range []TestMorpheme{
				{"tuple", TUPLE},
				{"block", BLOCK},
			} {
				Specify(v.type_+"/literal", func() {
					//	        const script = parse(value + LITERAL);
					//	        const word = firstWord(script);
					//	        Expect(word.morphemes).to.have.length(2);
					//	        Expect(checker.checkWord(word)).To(Equal(WordType.INVALID);)
				})
				Specify(v.type_+"/substitution", func() {
					//	        const script = parse(value + "$" + LITERAL);
					//	        const word = firstWord(script);
					//	        Expect(word.morphemes).to.have.length(3);
					//	        Expect(checker.checkWord(word)).To(Equal(WordType.INVALID);)
				})
				Specify("expression/"+v.type_, func() {
					//	        const script = parse(EXPRESSION + value);
					//	        const word = firstWord(script);
					//	        Expect(word.morphemes).to.have.length(2);
					//	        Expect(checker.checkWord(word)).To(Equal(WordType.INVALID);)
				})
				Specify("literal/"+v.type_+"/literal", func() {
					//	        const script = parse(LITERAL + value + LITERAL);
					//	        const word = firstWord(script);
					//	        Expect(word.morphemes).to.have.length(3);
					//	        Expect(checker.checkWord(word)).To(Equal(WordType.INVALID);)
				})
				Specify("literal/"+v.type_+"/substitution", func() {
					//	        const script = parse(LITERAL + value + "$" + LITERAL);
					//	        const word = firstWord(script);
					//	        Expect(word.morphemes).to.have.length(4);
					//	        Expect(checker.checkWord(word)).To(Equal(WordType.INVALID);)
				})
			}
			Specify("literal/tuple substitution", func() {
				//	      const script = parse(LITERAL + "$" + TUPLE);
				//	      const word = firstWord(script);
				//	      Expect(word.morphemes).to.have.length(3);
				//	      Expect(checker.checkWord(word)).To(Equal(WordType.INVALID);)
			})
			Specify("tuple substitution/literal", func() {
				//	      const script = parse("$" + TUPLE + LITERAL);
				//	      const word = firstWord(script);
				//	      Expect(word.morphemes).to.have.length(3);
				//	      Expect(checker.checkWord(word)).To(Equal(WordType.INVALID);)
			})
		})
	})
	//
	Describe("substitutions", func() {
		for _, v := range roots {
			Describe(v.type_+" source", func() {
				Specify("simple", func() {
					//	        const script = parse("$" + value);
					//	        const word = firstWord(script);
					//	        Expect(word.morphemes).to.have.length(2);
					//	        Expect(checker.checkWord(word)).to.eq(WordType.SUBSTITUTION);
				})
				Specify("double", func() {
					//	        const script = parse("$$" + value);
					//	        const word = firstWord(script);
					//	        Expect(word.morphemes).to.have.length(2);
					//	        Expect(checker.checkWord(word)).to.eq(WordType.SUBSTITUTION);
				})
				Specify("indexed selector", func() {
					//	        const script = parse("$" + value + EXPRESSION);
					//	        const word = firstWord(script);
					//	        Expect(word.morphemes).to.have.length(3);
					//	        Expect(checker.checkWord(word)).to.eq(WordType.SUBSTITUTION);
				})
				Specify("keyed selector", func() {
					//	        const script = parse("$" + value + TUPLE);
					//	        const word = firstWord(script);
					//	        Expect(word.morphemes).to.have.length(3);
					//	        Expect(checker.checkWord(word)).to.eq(WordType.SUBSTITUTION);
				})
				Specify("generic selector", func() {
					//	        const script = parse("$" + value + BLOCK);
					//	        const word = firstWord(script);
					//	        Expect(word.morphemes).to.have.length(3);
					//	        Expect(checker.checkWord(word)).to.eq(WordType.SUBSTITUTION);
				})
				Specify("multiple selectors", func() {
					//	        const script = parse(
					//	          "$" + value + TUPLE + BLOCK + EXPRESSION + TUPLE + EXPRESSION
					//	        );
					//	        const word = firstWord(script);
					//	        Expect(word.morphemes).to.have.length(7);
					//	        Expect(checker.checkWord(word)).to.eq(WordType.SUBSTITUTION);
				})
			})
		}
	})
	//
	Describe("qualified words", func() {
		for _, v := range qualifiedSources {
			Describe(v.type_+" source", func() {
				Specify("indexed selector", func() {
					//	        const script = parse(value + EXPRESSION);
					//	        const word = firstWord(script);
					//	        Expect(word.morphemes).to.have.length(2);
					//	        Expect(checker.checkWord(word)).to.eq(WordType.QUALIFIED);
				})
				Specify("keyed selector", func() {
					//	        const script = parse(value + TUPLE);
					//	        const word = firstWord(script);
					//	        Expect(word.morphemes).to.have.length(2);
					//	        Expect(checker.checkWord(word)).to.eq(WordType.QUALIFIED);
				})
				Specify("generic selector", func() {
					//	        const script = parse(value + BLOCK);
					//	        const word = firstWord(script);
					//	        Expect(word.morphemes).to.have.length(2);
					//	        Expect(checker.checkWord(word)).to.eq(WordType.QUALIFIED);
				})
				Specify("multiple selectors", func() {
					//	        const script = parse(
					//	          value + TUPLE + BLOCK + EXPRESSION + TUPLE + EXPRESSION
					//	        );
					//	        const word = firstWord(script);
					//	        Expect(word.morphemes).to.have.length(6);
					//	        Expect(checker.checkWord(word)).to.eq(WordType.QUALIFIED);
				})
			})
		}
		Describe("exceptions", func() {
			Specify("trailing morphemes", func() {
				//	      const script = parse(
				//	        LITERAL + TUPLE + TUPLE + EXPRESSION + TUPLE + EXPRESSION + LITERAL
				//	      );
				//	      const word = firstWord(script);
				//	      Expect(word.morphemes).to.have.length(7);
				//	      Expect(checker.checkWord(word)).To(Equal(WordType.INVALID);)
			})
		})
	})
	//
	Describe("ignored words", func() {
		for _, v := range ignored {
			Specify(v.type_, func() {
				//	      const script = parse(value);
				//	      const word = firstWord(script);
				//	      Expect(word.morphemes).to.have.length(1);
				//	      Expect(checker.checkWord(word)).to.eq(WordType.IGNORED);
			})
		}
	})

	Describe("impossible cases", func() {
		Specify("empty word", func() {
			word := Word{}
			Expect(checker.CheckWord(word)).To(Equal(INVALID))
		})
		Specify("empty substitution", func() {
			word := Word{}
			word.Morphemes = append(word.Morphemes, Morpheme{Type: MorphemeType_SUBSTITUTE_NEXT})
			Expect(checker.CheckWord(word)).To(Equal(INVALID))
		})
		Describe("incompatible morphemes", func() {
			for _, v1 := range append(monomorphemes, ignored...) {
				for _, v2 := range append(roots, monomorphemes...) {
					Specify(v1.type_+"/"+v2.type_, func() {
						//   const word = Word{};
						//   word.morphemes = append(word.morphemes,
						// 	firstWord(parse(value1)).morphemes[0]),
						// 	firstWord(parse(value2)).morphemes[0]),
						//   )
						//   Expect(checker.checkWord(word)).To(Equal(WordType.INVALID);)
					})
					Specify(v2.type_+"/"+v1.type_, func() {
						//	          const word = new Word();
						//	          word.morphemes.push(firstWord(parse(value2)).morphemes[0]);
						//	          word.morphemes.push(firstWord(parse(value1)).morphemes[0]);
						//	          Expect(checker.checkWord(word)).To(Equal(WordType.INVALID);)
					})
				}
			}
			Specify("substitution", func() {
				word := Word{}
				word.Morphemes = append(word.Morphemes,
					Morpheme{Type: MorphemeType_SUBSTITUTE_NEXT},
					Morpheme{Type: MorphemeType_BLOCK_COMMENT})
				Expect(checker.CheckWord(word)).To(Equal(INVALID))
			})
		})
	})
})
