package core_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "helena/core"
)

// import { MorphemeType, Script, SyntaxChecker, Word, WordType } from "./syntax";
// import { Parser } from "./parser";
// import { Tokenizer } from "./tokenizer";

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
			//	    const script = parse(_LITERAL + "$" + _BLOCK);
			//	    const word = firstWord(script);
			//	    Expect(word.morphemes).to.have.length(3);
			//	    Expect(checker.checkWord(word)).to.eq(WordType.COMPOUND);
		})
		Specify("expression prefix", func() {
			//	    const script = parse(_EXPRESSION + _LITERAL);
			//	    const word = firstWord(script);
			//	    Expect(word.morphemes).to.have.length(2);
			//	    Expect(checker.checkWord(word)).to.eq(WordType.COMPOUND);
		})
		Specify("substitution prefix", func() {
			//	    const script = parse("$" + _BLOCK + _TUPLE + _LITERAL);
			//	    const word = firstWord(script);
			//	    Expect(word.morphemes).to.have.length(4);
			//	    Expect(checker.checkWord(word)).to.eq(WordType.COMPOUND);
		})
		Specify("complex case", func() {
			//	    const script = parse(_LITERAL + "$" + _BLOCK + _EXPRESSION + "$" + _LITERAL);
			//	    const word = firstWord(script);
			//	    Expect(word.morphemes).to.have.length(6);
			//	    Expect(checker.checkWord(word)).to.eq(WordType.COMPOUND);
		})
		Describe("exceptions", func() {
			for _, v := range []testMorpheme{
				{"tuple", _TUPLE},
				{"block", _BLOCK},
			} {
				Specify(v.type_+"/literal", func() {
					//	        const script = parse(value + _LITERAL);
					//	        const word = firstWord(script);
					//	        Expect(word.morphemes).to.have.length(2);
					//	        Expect(checker.checkWord(word)).To(Equal(WordType.INVALID);)
				})
				Specify(v.type_+"/substitution", func() {
					//	        const script = parse(value + "$" + _LITERAL);
					//	        const word = firstWord(script);
					//	        Expect(word.morphemes).to.have.length(3);
					//	        Expect(checker.checkWord(word)).To(Equal(WordType.INVALID);)
				})
				Specify("expression/"+v.type_, func() {
					//	        const script = parse(_EXPRESSION + value);
					//	        const word = firstWord(script);
					//	        Expect(word.morphemes).to.have.length(2);
					//	        Expect(checker.checkWord(word)).To(Equal(WordType.INVALID);)
				})
				Specify("literal/"+v.type_+"/literal", func() {
					//	        const script = parse(_LITERAL + value + _LITERAL);
					//	        const word = firstWord(script);
					//	        Expect(word.morphemes).to.have.length(3);
					//	        Expect(checker.checkWord(word)).To(Equal(WordType.INVALID);)
				})
				Specify("literal/"+v.type_+"/substitution", func() {
					//	        const script = parse(_LITERAL + value + "$" + _LITERAL);
					//	        const word = firstWord(script);
					//	        Expect(word.morphemes).to.have.length(4);
					//	        Expect(checker.checkWord(word)).To(Equal(WordType.INVALID);)
				})
			}
			Specify("literal/tuple substitution", func() {
				//	      const script = parse(_LITERAL + "$" + _TUPLE);
				//	      const word = firstWord(script);
				//	      Expect(word.morphemes).to.have.length(3);
				//	      Expect(checker.checkWord(word)).To(Equal(WordType.INVALID);)
			})
			Specify("tuple substitution/literal", func() {
				//	      const script = parse("$" + _TUPLE + _LITERAL);
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
					//	        const script = parse("$" + value + _EXPRESSION);
					//	        const word = firstWord(script);
					//	        Expect(word.morphemes).to.have.length(3);
					//	        Expect(checker.checkWord(word)).to.eq(WordType.SUBSTITUTION);
				})
				Specify("keyed selector", func() {
					//	        const script = parse("$" + value + _TUPLE);
					//	        const word = firstWord(script);
					//	        Expect(word.morphemes).to.have.length(3);
					//	        Expect(checker.checkWord(word)).to.eq(WordType.SUBSTITUTION);
				})
				Specify("generic selector", func() {
					//	        const script = parse("$" + value + _BLOCK);
					//	        const word = firstWord(script);
					//	        Expect(word.morphemes).to.have.length(3);
					//	        Expect(checker.checkWord(word)).to.eq(WordType.SUBSTITUTION);
				})
				Specify("multiple selectors", func() {
					//	        const script = parse(
					//	          "$" + value + _TUPLE + _BLOCK + _EXPRESSION + _TUPLE + _EXPRESSION
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
					//	        const script = parse(value + _EXPRESSION);
					//	        const word = firstWord(script);
					//	        Expect(word.morphemes).to.have.length(2);
					//	        Expect(checker.checkWord(word)).to.eq(WordType.QUALIFIED);
				})
				Specify("keyed selector", func() {
					//	        const script = parse(value + _TUPLE);
					//	        const word = firstWord(script);
					//	        Expect(word.morphemes).to.have.length(2);
					//	        Expect(checker.checkWord(word)).to.eq(WordType.QUALIFIED);
				})
				Specify("generic selector", func() {
					//	        const script = parse(value + _BLOCK);
					//	        const word = firstWord(script);
					//	        Expect(word.morphemes).to.have.length(2);
					//	        Expect(checker.checkWord(word)).to.eq(WordType.QUALIFIED);
				})
				Specify("multiple selectors", func() {
					//	        const script = parse(
					//	          value + _TUPLE + _BLOCK + _EXPRESSION + _TUPLE + _EXPRESSION
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
				//	        _LITERAL + _TUPLE + _TUPLE + _EXPRESSION + _TUPLE + _EXPRESSION + _LITERAL
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
