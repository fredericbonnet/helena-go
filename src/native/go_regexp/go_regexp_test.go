package go_regexp_test

import (
	"regexp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"helena/core"
	"helena/native/go_regexp"
)

var _ = Describe("Go regexp", func() {
	var tokenizer core.Tokenizer
	var parser *core.Parser
	var variableResolver *mockVariableResolver
	var commandResolver *mockCommandResolver
	var evaluator core.Evaluator

	parse := func(script string) *core.Script {
		return parser.ParseTokens(tokenizer.Tokenize(script), nil).Script
	}
	execute := func(script string) core.Result {
		return evaluator.EvaluateScript(*parse(script))
	}
	evaluate := func(script string) core.Value { return execute(script).Value }

	BeforeEach(func() {
		tokenizer = core.Tokenizer{}
		parser = core.NewParser(nil)
		variableResolver = newMockVariableResolver()
		commandResolver = newMockCommandResolver()
		evaluator = core.NewCompilingEvaluator(
			variableResolver,
			commandResolver,
			nil,
			nil,
		)
		commandResolver.register("go:regexp", &go_regexp.RegexpCmd{})
	})

	Describe("go:RegexpValue", func() {
		Specify("type should be custom", func() {
			value := go_regexp.NewRegexpValue(regexp.MustCompile("."))
			Expect(core.IsCustomValue(value, go_regexp.RegexpValueType)).To(BeTrue())
		})
		Specify("display", func() {
			re := regexp.MustCompile("a")
			value := go_regexp.NewRegexpValue(re)
			Expect(value.Display(nil)).To(Equal(`{#{Regexp ` + re.String() + `}#}`))
		})
	})

	Describe("go:regexp", func() {
		Describe("`QuoteMeta`", func() {
			Specify("Go documentation example", func() {
				// https://pkg.go.dev/regexp#example-QuoteMeta
				Expect(evaluate(`go:regexp QuoteMeta """Escaping symbols like: .+*?()|[]{}^$"""`)).To(Equal(
					STR(`Escaping symbols like: \.\+\*\?\(\)\|\[\]\{\}\^\$`),
				))
			})
			Describe("Exceptions", func() {
				Specify("wrong arity", func() {
					Expect(execute("go:regexp QuoteMeta")).To(Equal(
						ERROR(`wrong # args: should be "regexp QuoteMeta s"`),
					))
					Expect(execute("go:regexp QuoteMeta a b")).To(Equal(
						ERROR(`wrong # args: should be "regexp QuoteMeta s"`),
					))
				})
				Specify("invalid string value", func() {
					Expect(execute("go:regexp QuoteMeta []")).To(Equal(
						ERROR("value has no string representation"),
					))
				})
			})
		})
		Describe("`Compile`", func() {
			It("should return a RegexpValue", func() {
				Expect(core.IsCustomValue(evaluate(`go:regexp Compile ""`), go_regexp.RegexpValueType)).To(BeTrue())
			})
			Describe("Exceptions", func() {
				Specify("wrong arity", func() {
					Expect(execute("go:regexp Compile")).To(Equal(
						ERROR(`wrong # args: should be "regexp Compile expr"`),
					))
					Expect(execute("go:regexp Compile a b")).To(Equal(
						ERROR(`wrong # args: should be "regexp Compile expr"`),
					))
				})
				Specify("invalid expression value", func() {
					Expect(execute("go:regexp Compile []")).To(Equal(
						ERROR("value has no string representation"),
					))
				})
				Specify("invalid regular expression", func() {
					Expect(execute(`go:regexp Compile """["""`)).To(Equal(
						ERROR("error parsing regexp: missing closing ]: `[`"),
					))
				})
			})
		})
		Describe("`CompilePOSIX`", func() {
			It("should return a RegexpValue", func() {
				Expect(core.IsCustomValue(evaluate(`go:regexp CompilePOSIX ""`), go_regexp.RegexpValueType)).To(BeTrue())
			})
			Describe("Exceptions", func() {
				Specify("wrong arity", func() {
					Expect(execute("go:regexp CompilePOSIX")).To(Equal(
						ERROR(`wrong # args: should be "regexp CompilePOSIX expr"`),
					))
					Expect(execute("go:regexp CompilePOSIX a b")).To(Equal(
						ERROR(`wrong # args: should be "regexp CompilePOSIX expr"`),
					))
				})
				Specify("invalid expression value", func() {
					Expect(execute("go:regexp CompilePOSIX []")).To(Equal(
						ERROR("value has no string representation"),
					))
				})
				Specify("invalid regular expression", func() {
					Expect(execute(`go:regexp CompilePOSIX """["""`)).To(Equal(
						ERROR("error parsing regexp: missing closing ]: `[`"),
					))
				})
			})
		})
		Describe("`FindAllString`", func() {
			Specify("Go documentation example", func() {
				// https://pkg.go.dev/regexp#example-Regexp.FindAllString
				variableResolver.register(
					"re",
					evaluate(`go:regexp Compile "a."`),
				)
				Expect(evaluate("go:regexp FindAllString $re paranormal -1")).To(Equal(
					LIST([]core.Value{STR("ar"), STR("an"), STR("al")}),
				))
				Expect(evaluate("go:regexp FindAllString $re paranormal 2")).To(Equal(
					LIST([]core.Value{STR("ar"), STR("an")}),
				))
				Expect(evaluate("go:regexp FindAllString $re graal -1")).To(Equal(
					LIST([]core.Value{STR("aa")}),
				))
				Expect(evaluate("go:regexp FindAllString $re none -1")).To(Equal(
					NIL,
				))
			})
			Describe("Exceptions", func() {
				Specify("wrong arity", func() {
					Expect(execute("go:regexp FindAllString a b")).To(Equal(
						ERROR(`wrong # args: should be "regexp FindAllString re s n"`),
					))
					Expect(execute("go:regexp FindAllString a b c d")).To(Equal(
						ERROR(`wrong # args: should be "regexp FindAllString re s n"`),
					))
				})
				Specify("invalid regexp value", func() {
					Expect(execute("go:regexp FindAllString a b 1")).To(Equal(
						ERROR("invalid regexp value"),
					))
				})
				Specify("invalid string value", func() {
					Expect(execute("go:regexp FindAllString [go:regexp Compile {}] [] 1")).To(Equal(
						ERROR("value has no string representation"),
					))
				})
				Specify("invalid n value", func() {
					Expect(execute("go:regexp FindAllString [go:regexp Compile {}] a b")).To(Equal(
						ERROR(`invalid integer "b"`),
					))
				})
			})
		})
		Describe("`FindAllStringIndex`", func() {
			Specify("Go documentation example", func() {
				// Adapted from https://pkg.go.dev/regexp#example-Regexp.FindAllIndex
				variableResolver.register(
					"re",
					evaluate(`go:regexp Compile "o."`),
				)
				Expect(evaluate("go:regexp FindAllStringIndex $re London 1")).To(Equal(
					LIST([]core.Value{
						core.LIST([]core.Value{INT(1), INT(3)}),
					}),
				))
				Expect(evaluate("go:regexp FindAllStringIndex $re London -1")).To(Equal(
					LIST([]core.Value{
						core.LIST([]core.Value{INT(1), INT(3)}),
						core.LIST([]core.Value{INT(4), INT(6)}),
					}),
				))
			})
			Describe("Exceptions", func() {
				Specify("wrong arity", func() {
					Expect(execute("go:regexp FindAllStringIndex a b")).To(Equal(
						ERROR(`wrong # args: should be "regexp FindAllStringIndex re s n"`),
					))
					Expect(execute("go:regexp FindAllStringIndex a b c d")).To(Equal(
						ERROR(`wrong # args: should be "regexp FindAllStringIndex re s n"`),
					))
				})
				Specify("invalid regexp value", func() {
					Expect(execute("go:regexp FindAllStringIndex a b 1")).To(Equal(
						ERROR("invalid regexp value"),
					))
				})
				Specify("invalid string value", func() {
					Expect(execute("go:regexp FindAllStringIndex [go:regexp Compile {}] [] 1")).To(Equal(
						ERROR("value has no string representation"),
					))
				})
				Specify("invalid n value", func() {
					Expect(execute("go:regexp FindAllStringIndex [go:regexp Compile {}] a b")).To(Equal(
						ERROR(`invalid integer "b"`),
					))
				})
			})
		})
		Describe("`FindAllStringSubmatch`", func() {
			Specify("Go documentation example", func() {
				// https://pkg.go.dev/regexp#example-Regexp.FindAllStringSubmatch
				variableResolver.register(
					"re",
					evaluate(`go:regexp Compile "a(x*)b"`),
				)
				Expect(evaluate("go:regexp FindAllStringSubmatch $re -ab- -1")).To(Equal(
					LIST([]core.Value{
						core.LIST([]core.Value{STR("ab"), STR("")}),
					}),
				))
				Expect(evaluate("go:regexp FindAllStringSubmatch $re -axxb- -1")).To(Equal(
					LIST([]core.Value{
						core.LIST([]core.Value{STR("axxb"), STR("xx")}),
					}),
				))
				Expect(evaluate("go:regexp FindAllStringSubmatch $re -ab-axb- -1")).To(Equal(
					LIST([]core.Value{
						core.LIST([]core.Value{STR("ab"), STR("")}),
						core.LIST([]core.Value{STR("axb"), STR("x")}),
					}),
				))
				Expect(evaluate("go:regexp FindAllStringSubmatch $re -axxb-ab- -1")).To(Equal(
					LIST([]core.Value{
						core.LIST([]core.Value{STR("axxb"), STR("xx")}),
						core.LIST([]core.Value{STR("ab"), STR("")}),
					}),
				))
			})
			Describe("Exceptions", func() {
				Specify("wrong arity", func() {
					Expect(execute("go:regexp FindAllStringSubmatch a b")).To(Equal(
						ERROR(`wrong # args: should be "regexp FindAllStringSubmatch re s n"`),
					))
					Expect(execute("go:regexp FindAllStringSubmatch a b c d")).To(Equal(
						ERROR(`wrong # args: should be "regexp FindAllStringSubmatch re s n"`),
					))
				})
				Specify("invalid regexp value", func() {
					Expect(execute("go:regexp FindAllStringSubmatch a b 1")).To(Equal(
						ERROR("invalid regexp value"),
					))
				})
				Specify("invalid string value", func() {
					Expect(execute("go:regexp FindAllStringSubmatch [go:regexp Compile {}] [] 1")).To(Equal(
						ERROR("value has no string representation"),
					))
				})
				Specify("invalid n value", func() {
					Expect(execute("go:regexp FindAllStringSubmatch [go:regexp Compile {}] a b")).To(Equal(
						ERROR(`invalid integer "b"`),
					))
				})
			})
		})
		Describe("`FindAllStringSubmatchIndex`", func() {
			Specify("Go documentation example", func() {
				// https://pkg.go.dev/regexp#example-Regexp.FindAllStringSubmatchIndex
				variableResolver.register(
					"re",
					evaluate(`go:regexp Compile "a(x*)b"`),
				)
				Expect(evaluate("go:regexp FindAllStringSubmatchIndex $re -ab- -1")).To(Equal(
					LIST([]core.Value{
						core.LIST([]core.Value{INT(1), INT(3), INT(2), INT(2)}),
					}),
				))
				Expect(evaluate("go:regexp FindAllStringSubmatchIndex $re -axxb- -1")).To(Equal(
					LIST([]core.Value{
						core.LIST([]core.Value{INT(1), INT(5), INT(2), INT(4)}),
					}),
				))
				Expect(evaluate("go:regexp FindAllStringSubmatchIndex $re -ab-axb- -1")).To(Equal(
					LIST([]core.Value{
						core.LIST([]core.Value{INT(1), INT(3), INT(2), INT(2)}),
						core.LIST([]core.Value{INT(4), INT(7), INT(5), INT(6)}),
					}),
				))
				Expect(evaluate("go:regexp FindAllStringSubmatchIndex $re -axxb-ab- -1")).To(Equal(
					LIST([]core.Value{
						core.LIST([]core.Value{INT(1), INT(5), INT(2), INT(4)}),
						core.LIST([]core.Value{INT(6), INT(8), INT(7), INT(7)}),
					}),
				))
				Expect(evaluate("go:regexp FindAllStringSubmatchIndex $re foo -1")).To(Equal(
					NIL,
				))
			})
			Describe("Exceptions", func() {
				Specify("wrong arity", func() {
					Expect(execute("go:regexp FindAllStringSubmatchIndex a b")).To(Equal(
						ERROR(`wrong # args: should be "regexp FindAllStringSubmatchIndex re s n"`),
					))
					Expect(execute("go:regexp FindAllStringSubmatchIndex a b c d")).To(Equal(
						ERROR(`wrong # args: should be "regexp FindAllStringSubmatchIndex re s n"`),
					))
				})
				Specify("invalid regexp value", func() {
					Expect(execute("go:regexp FindAllStringSubmatchIndex a b 1")).To(Equal(
						ERROR("invalid regexp value"),
					))
				})
				Specify("invalid string value", func() {
					Expect(execute("go:regexp FindAllStringSubmatchIndex [go:regexp Compile {}] [] 1")).To(Equal(
						ERROR("value has no string representation"),
					))
				})
				Specify("invalid n value", func() {
					Expect(execute("go:regexp FindAllStringSubmatchIndex [go:regexp Compile {}] a b")).To(Equal(
						ERROR(`invalid integer "b"`),
					))
				})
			})
		})
		Describe("`FindString`", func() {
			Specify("Go documentation example", func() {
				// https://pkg.go.dev/regexp#example-Regexp.FindString
				variableResolver.register(
					"re",
					evaluate(`go:regexp Compile "foo.?"`),
				)
				Expect(evaluate(`go:regexp FindString $re "seafood fool"`)).To(Equal(
					STR("food"),
				))
				Expect(evaluate("go:regexp FindString $re meat")).To(Equal(
					STR(""),
				))
			})
			Describe("Exceptions", func() {
				Specify("wrong arity", func() {
					Expect(execute("go:regexp FindString a")).To(Equal(
						ERROR(`wrong # args: should be "regexp FindString re s"`),
					))
					Expect(execute("go:regexp FindString a b c")).To(Equal(
						ERROR(`wrong # args: should be "regexp FindString re s"`),
					))
				})
				Specify("invalid regexp value", func() {
					Expect(execute("go:regexp FindString a b")).To(Equal(
						ERROR("invalid regexp value"),
					))
				})
				Specify("invalid string value", func() {
					Expect(execute("go:regexp FindString [go:regexp Compile {}] []")).To(Equal(
						ERROR("value has no string representation"),
					))
				})
			})
		})
		Describe("`FindStringIndex`", func() {
			Specify("Go documentation example", func() {
				// https://pkg.go.dev/regexp#example-Regexp.FindStringIndex
				variableResolver.register(
					"re",
					evaluate(`go:regexp Compile "ab?"`),
				)
				Expect(evaluate(`go:regexp FindStringIndex $re tablett`)).To(Equal(
					LIST([]core.Value{INT(1), INT(3)}),
				))
				Expect(evaluate("go:regexp FindStringIndex $re foo")).To(Equal(
					NIL,
				))
			})
			Describe("Exceptions", func() {
				Specify("wrong arity", func() {
					Expect(execute("go:regexp FindStringIndex a")).To(Equal(
						ERROR(`wrong # args: should be "regexp FindStringIndex re s"`),
					))
					Expect(execute("go:regexp FindStringIndex a b c")).To(Equal(
						ERROR(`wrong # args: should be "regexp FindStringIndex re s"`),
					))
				})
				Specify("invalid regexp value", func() {
					Expect(execute("go:regexp FindStringIndex a b")).To(Equal(
						ERROR("invalid regexp value"),
					))
				})
				Specify("invalid string value", func() {
					Expect(execute("go:regexp FindStringIndex [go:regexp Compile {}] []")).To(Equal(
						ERROR("value has no string representation"),
					))
				})
			})
		})
		Describe("`FindStringSubmatch`", func() {
			Specify("Go documentation example", func() {
				// https://pkg.go.dev/regexp#example-Regexp.FindStringSubmatch
				variableResolver.register(
					"re",
					evaluate(`go:regexp Compile "a(x*)b(y|z)c"`),
				)
				Expect(evaluate(`go:regexp FindStringSubmatch $re -axxxbyc-`)).To(Equal(
					LIST([]core.Value{STR("axxxbyc"), STR("xxx"), STR("y")}),
				))
				Expect(evaluate("go:regexp FindStringSubmatch $re -abzc-")).To(Equal(
					LIST([]core.Value{STR("abzc"), STR(""), STR("z")}),
				))
			})
			Describe("Exceptions", func() {
				Specify("wrong arity", func() {
					Expect(execute("go:regexp FindStringSubmatch a")).To(Equal(
						ERROR(`wrong # args: should be "regexp FindStringSubmatch re s"`),
					))
					Expect(execute("go:regexp FindStringSubmatch a b c")).To(Equal(
						ERROR(`wrong # args: should be "regexp FindStringSubmatch re s"`),
					))
				})
				Specify("invalid regexp value", func() {
					Expect(execute("go:regexp FindStringSubmatch a b")).To(Equal(
						ERROR("invalid regexp value"),
					))
				})
				Specify("invalid string value", func() {
					Expect(execute("go:regexp FindStringSubmatch [go:regexp Compile {}] []")).To(Equal(
						ERROR("value has no string representation"),
					))
				})
			})
		})
		Describe("`FindStringSubmatchIndex`", func() {
			Specify("Go documentation example", func() {
				// Adapted from https://pkg.go.dev/regexp#example-Regexp.FindSubmatchIndex
				variableResolver.register(
					"re",
					evaluate(`go:regexp Compile "a(x*)b"`),
				)
				Expect(evaluate(`go:regexp FindStringSubmatchIndex $re -ab-`)).To(Equal(
					LIST([]core.Value{INT(1), INT(3), INT(2), INT(2)}),
				))
				Expect(evaluate(`go:regexp FindStringSubmatchIndex $re -axxb-`)).To(Equal(
					LIST([]core.Value{INT(1), INT(5), INT(2), INT(4)}),
				))
				Expect(evaluate(`go:regexp FindStringSubmatchIndex $re -ab-axb-`)).To(Equal(
					LIST([]core.Value{INT(1), INT(3), INT(2), INT(2)}),
				))
				Expect(evaluate(`go:regexp FindStringSubmatchIndex $re -axxb-ab-`)).To(Equal(
					LIST([]core.Value{INT(1), INT(5), INT(2), INT(4)}),
				))
				Expect(evaluate(`go:regexp FindStringSubmatchIndex $re -foo-`)).To(Equal(
					NIL,
				))
			})
			Describe("Exceptions", func() {
				Specify("wrong arity", func() {
					Expect(execute("go:regexp FindStringSubmatchIndex a")).To(Equal(
						ERROR(`wrong # args: should be "regexp FindStringSubmatchIndex re s"`),
					))
					Expect(execute("go:regexp FindStringSubmatchIndex a b c")).To(Equal(
						ERROR(`wrong # args: should be "regexp FindStringSubmatchIndex re s"`),
					))
				})
				Specify("invalid regexp value", func() {
					Expect(execute("go:regexp FindStringSubmatchIndex a b")).To(Equal(
						ERROR("invalid regexp value"),
					))
				})
				Specify("invalid string value", func() {
					Expect(execute("go:regexp FindStringSubmatchIndex [go:regexp Compile {}] []")).To(Equal(
						ERROR("value has no string representation"),
					))
				})
			})
		})
		Describe("`LiteralPrefix`", func() {
			Specify("complete", func() {
				variableResolver.register(
					"re",
					evaluate(`go:regexp Compile "abc"`),
				)
				Expect(evaluate(`go:regexp LiteralPrefix $re`)).To(Equal(
					TUPLE([]core.Value{STR("abc"), TRUE}),
				))
			})
			Specify("not complete", func() {
				variableResolver.register(
					"re",
					evaluate(`go:regexp Compile "a(x+)b"`),
				)
				Expect(evaluate(`go:regexp LiteralPrefix $re`)).To(Equal(
					TUPLE([]core.Value{STR("ax"), FALSE}),
				))
			})
			Describe("Exceptions", func() {
				Specify("wrong arity", func() {
					Expect(execute("go:regexp LiteralPrefix")).To(Equal(
						ERROR(`wrong # args: should be "regexp LiteralPrefix re"`),
					))
					Expect(execute("go:regexp LiteralPrefix a b")).To(Equal(
						ERROR(`wrong # args: should be "regexp LiteralPrefix re"`),
					))
				})
				Specify("invalid regexp value", func() {
					Expect(execute("go:regexp LiteralPrefix a")).To(Equal(
						ERROR("invalid regexp value"),
					))
				})
			})
		})
		Describe("`Longest`", func() {
			Specify("Go documentation example", func() {
				// https://pkg.go.dev/regexp#example-Regexp.Longest
				variableResolver.register(
					"re",
					evaluate(`go:regexp Compile "a(|b)"`),
				)
				Expect(evaluate(`go:regexp FindString $re ab`)).To(Equal(
					STR("a"),
				))
				evaluate(`go:regexp Longest $re`)
				Expect(evaluate(`go:regexp FindString $re ab`)).To(Equal(
					STR("ab"),
				))
			})
			Describe("Exceptions", func() {
				Specify("wrong arity", func() {
					Expect(execute("go:regexp Longest")).To(Equal(
						ERROR(`wrong # args: should be "regexp Longest re"`),
					))
					Expect(execute("go:regexp Longest a b")).To(Equal(
						ERROR(`wrong # args: should be "regexp Longest re"`),
					))
				})
				Specify("invalid regexp value", func() {
					Expect(execute("go:regexp Longest a")).To(Equal(
						ERROR("invalid regexp value"),
					))
				})
			})
		})
		Describe("`MatchString`", func() {
			Specify("Go documentation example", func() {
				// https://pkg.go.dev/regexp#example-Regexp.MatchString
				variableResolver.register(
					"re",
					evaluate(`go:regexp Compile "(gopher){2}"`),
				)
				Expect(evaluate(`go:regexp MatchString $re gopher`)).To(Equal(
					FALSE,
				))
				Expect(evaluate(`go:regexp MatchString $re gophergopher`)).To(Equal(
					TRUE,
				))
				Expect(evaluate(`go:regexp MatchString $re gophergophergopher`)).To(Equal(
					TRUE,
				))
			})
			Describe("Exceptions", func() {
				Specify("wrong arity", func() {
					Expect(execute("go:regexp MatchString a")).To(Equal(
						ERROR(`wrong # args: should be "regexp MatchString re s"`),
					))
					Expect(execute("go:regexp MatchString a b c")).To(Equal(
						ERROR(`wrong # args: should be "regexp MatchString re s"`),
					))
				})
				Specify("invalid regexp value", func() {
					Expect(execute("go:regexp MatchString a b")).To(Equal(
						ERROR("invalid regexp value"),
					))
				})
				Specify("invalid string value", func() {
					Expect(execute("go:regexp MatchString [go:regexp Compile {}] []")).To(Equal(
						ERROR("value has no string representation"),
					))
				})
			})
		})
		Describe("`NumSubexp`", func() {
			Specify("Go documentation example", func() {
				// https://pkg.go.dev/regexp#example-Regexp.NumSubexp
				variableResolver.register(
					"re0",
					evaluate(`go:regexp Compile "a."`),
				)
				Expect(evaluate(`go:regexp NumSubexp $re0`)).To(Equal(
					INT(0),
				))

				variableResolver.register(
					"re",
					evaluate(`go:regexp Compile "(.*)((a)b)(.*)a"`),
				)
				Expect(evaluate(`go:regexp NumSubexp $re`)).To(Equal(
					INT(4),
				))
			})
			Describe("Exceptions", func() {
				Specify("wrong arity", func() {
					Expect(execute("go:regexp NumSubexp")).To(Equal(
						ERROR(`wrong # args: should be "regexp NumSubexp re"`),
					))
					Expect(execute("go:regexp NumSubexp a b")).To(Equal(
						ERROR(`wrong # args: should be "regexp NumSubexp re"`),
					))
				})
				Specify("invalid regexp value", func() {
					Expect(execute("go:regexp NumSubexp a")).To(Equal(
						ERROR("invalid regexp value"),
					))
				})
			})
		})
		Describe("`ReplaceAllLiteralString`", func() {
			Specify("Go documentation example", func() {
				// https://pkg.go.dev/regexp#example-Regexp.ReplaceAllLiteralString
				variableResolver.register(
					"re",
					evaluate(`go:regexp Compile "a(x*)b"`),
				)
				Expect(evaluate(`go:regexp ReplaceAllLiteralString $re -ab-axxb- T`)).To(Equal(
					STR("-T-T-"),
				))
				Expect(evaluate(`go:regexp ReplaceAllLiteralString $re -ab-axxb- """$1"""`)).To(Equal(
					STR("-$1-$1-"),
				))
				Expect(evaluate(`go:regexp ReplaceAllLiteralString $re -ab-axxb- """${1}"""`)).To(Equal(
					STR("-${1}-${1}-"),
				))
			})
			Describe("Exceptions", func() {
				Specify("wrong arity", func() {
					Expect(execute("go:regexp ReplaceAllLiteralString a b")).To(Equal(
						ERROR(`wrong # args: should be "regexp ReplaceAllLiteralString re src repl"`),
					))
					Expect(execute("go:regexp ReplaceAllLiteralString a b c d")).To(Equal(
						ERROR(`wrong # args: should be "regexp ReplaceAllLiteralString re src repl"`),
					))
				})
				Specify("invalid regexp value", func() {
					Expect(execute("go:regexp ReplaceAllLiteralString a b c")).To(Equal(
						ERROR("invalid regexp value"),
					))
				})
				Specify("invalid string value", func() {
					Expect(execute("go:regexp ReplaceAllLiteralString [go:regexp Compile {}] [] a")).To(Equal(
						ERROR("value has no string representation"),
					))
					Expect(execute("go:regexp ReplaceAllLiteralString [go:regexp Compile {}] a []")).To(Equal(
						ERROR("value has no string representation"),
					))
				})
			})
		})
		Describe("`ReplaceAllString`", func() {
			Specify("Go documentation example", func() {
				// https://pkg.go.dev/regexp#example-Regexp.ReplaceAllString
				variableResolver.register(
					"re",
					evaluate(`go:regexp Compile "a(x*)b"`),
				)
				Expect(evaluate(`go:regexp ReplaceAllString $re -ab-axxb- T`)).To(Equal(
					STR("-T-T-"),
				))
				Expect(evaluate(`go:regexp ReplaceAllString $re -ab-axxb- """$1"""`)).To(Equal(
					STR("--xx-"),
				))
				Expect(evaluate(`go:regexp ReplaceAllString $re -ab-axxb- """$1W"""`)).To(Equal(
					STR("---"),
				))
				Expect(evaluate(`go:regexp ReplaceAllString $re -ab-axxb- """${1}W"""`)).To(Equal(
					STR("-W-xxW-"),
				))

				variableResolver.register(
					"re2",
					evaluate(`go:regexp Compile "a(?P<1W>x*)b"`),
				)
				Expect(evaluate(`go:regexp ReplaceAllString $re2 -ab-axxb- """$1W"""`)).To(Equal(
					STR("--xx-"),
				))
				Expect(evaluate(`go:regexp ReplaceAllString $re2 -ab-axxb- """${1}W"""`)).To(Equal(
					STR("-W-xxW-"),
				))
			})
			Describe("Exceptions", func() {
				Specify("wrong arity", func() {
					Expect(execute("go:regexp ReplaceAllString a b")).To(Equal(
						ERROR(`wrong # args: should be "regexp ReplaceAllString re src repl"`),
					))
					Expect(execute("go:regexp ReplaceAllString a b c d")).To(Equal(
						ERROR(`wrong # args: should be "regexp ReplaceAllString re src repl"`),
					))
				})
				Specify("invalid regexp value", func() {
					Expect(execute("go:regexp ReplaceAllString a b c")).To(Equal(
						ERROR("invalid regexp value"),
					))
				})
				Specify("invalid string value", func() {
					Expect(execute("go:regexp ReplaceAllString [go:regexp Compile {}] [] a")).To(Equal(
						ERROR("value has no string representation"),
					))
					Expect(execute("go:regexp ReplaceAllString [go:regexp Compile {}] a []")).To(Equal(
						ERROR("value has no string representation"),
					))
				})
			})
		})
		Describe("`Split`", func() {
			Specify("Go documentation example", func() {
				// https://pkg.go.dev/regexp#example-Regexp.Split
				variableResolver.register(
					"a",
					evaluate(`go:regexp Compile "a"`),
				)
				Expect(evaluate("go:regexp Split $a banana -1")).To(Equal(
					LIST([]core.Value{STR("b"), STR("n"), STR("n"), STR("")}),
				))
				Expect(evaluate("go:regexp Split $a banana 0")).To(Equal(
					LIST([]core.Value{}),
				))
				Expect(evaluate("go:regexp Split $a banana 1")).To(Equal(
					LIST([]core.Value{STR("banana")}),
				))
				Expect(evaluate("go:regexp Split $a banana 2")).To(Equal(
					LIST([]core.Value{STR("b"), STR("nana")}),
				))

				variableResolver.register(
					"zp",
					evaluate(`go:regexp Compile "z+"`),
				)
				Expect(evaluate("go:regexp Split $zp pizza -1")).To(Equal(
					LIST([]core.Value{STR("pi"), STR("a")}),
				))
				Expect(evaluate("go:regexp Split $zp pizza 0")).To(Equal(
					LIST([]core.Value{}),
				))
				Expect(evaluate("go:regexp Split $zp pizza 1")).To(Equal(
					LIST([]core.Value{STR("pizza")}),
				))
				Expect(evaluate("go:regexp Split $zp pizza 2")).To(Equal(
					LIST([]core.Value{STR("pi"), STR("a")}),
				))
			})
			Describe("Exceptions", func() {
				Specify("wrong arity", func() {
					Expect(execute("go:regexp Split a b")).To(Equal(
						ERROR(`wrong # args: should be "regexp Split re s n"`),
					))
					Expect(execute("go:regexp Split a b c d")).To(Equal(
						ERROR(`wrong # args: should be "regexp Split re s n"`),
					))
				})
				Specify("invalid regexp value", func() {
					Expect(execute("go:regexp Split a b 1")).To(Equal(
						ERROR("invalid regexp value"),
					))
				})
				Specify("invalid string value", func() {
					Expect(execute("go:regexp Split [go:regexp Compile {}] [] 1")).To(Equal(
						ERROR("value has no string representation"),
					))
				})
				Specify("invalid n value", func() {
					Expect(execute("go:regexp Split [go:regexp Compile {}] a b")).To(Equal(
						ERROR(`invalid integer "b"`),
					))
				})
			})
		})
		Describe("`String`", func() {
			Specify("Basic test", func() {
				variableResolver.register(
					"re",
					evaluate(`go:regexp Compile "a(x*)b"`),
				)
				Expect(evaluate("go:regexp String $re")).To(Equal(
					STR("a(x*)b"),
				))
			})
			Describe("Exceptions", func() {
				Specify("wrong arity", func() {
					Expect(execute("go:regexp String")).To(Equal(
						ERROR(`wrong # args: should be "regexp String re"`),
					))
					Expect(execute("go:regexp String a b")).To(Equal(
						ERROR(`wrong # args: should be "regexp String re"`),
					))
				})
				Specify("invalid regexp value", func() {
					Expect(execute("go:regexp String a")).To(Equal(
						ERROR("invalid regexp value"),
					))
				})
			})
		})
		Describe("`SubexpIndex`", func() {
			Specify("Go documentation example", func() {
				// https://pkg.go.dev/regexp#example-Regexp.SubexpIndex
				variableResolver.register(
					"re",
					evaluate(`go:regexp Compile """(?P<first>[a-zA-Z]+) (?P<last>[a-zA-Z]+)"""`),
				)
				Expect(evaluate(`go:regexp MatchString $re "Alan Turing"`)).To(Equal(
					TRUE,
				))
				matches := evaluate(`go:regexp FindStringSubmatch $re "Alan Turing"`)
				lastIndex := evaluate(`go:regexp SubexpIndex $re last`)
				Expect(lastIndex).To(Equal(INT(2)))
				Expect(matches.(core.ListValue).Values[lastIndex.(core.IntegerValue).Value]).To(Equal(STR("Turing")))
			})
			Describe("Exceptions", func() {
				Specify("wrong arity", func() {
					Expect(execute("go:regexp SubexpIndex a")).To(Equal(
						ERROR(`wrong # args: should be "regexp SubexpIndex re s"`),
					))
					Expect(execute("go:regexp SubexpIndex a b c")).To(Equal(
						ERROR(`wrong # args: should be "regexp SubexpIndex re s"`),
					))
				})
				Specify("invalid regexp value", func() {
					Expect(execute("go:regexp SubexpIndex a b")).To(Equal(
						ERROR("invalid regexp value"),
					))
				})
				Specify("invalid string value", func() {
					Expect(execute("go:regexp SubexpIndex [go:regexp Compile {}] []")).To(Equal(
						ERROR("value has no string representation"),
					))
				})
			})
		})
		Describe("`SubexpNames`", func() {
			Specify("Go documentation example", func() {
				// https://pkg.go.dev/regexp#example-Regexp.SubexpNames
				variableResolver.register(
					"re",
					evaluate(`go:regexp Compile """(?P<first>[a-zA-Z]+) (?P<last>[a-zA-Z]+)"""`),
				)
				Expect(evaluate(`go:regexp MatchString $re "Alan Turing"`)).To(Equal(
					TRUE,
				))
				Expect(evaluate(`go:regexp SubexpNames $re`)).To(Equal(
					core.LIST([]core.Value{STR(""), STR("first"), STR("last")}),
				))
				Expect(evaluate(`go:regexp ReplaceAllString $re "Alan Turing" """${last} ${first}"""`)).To(Equal(
					core.STR("Turing Alan"),
				))
			})
			Describe("Exceptions", func() {
				Specify("wrong arity", func() {
					Expect(execute("go:regexp SubexpNames")).To(Equal(
						ERROR(`wrong # args: should be "regexp SubexpNames re"`),
					))
					Expect(execute("go:regexp SubexpNames a b")).To(Equal(
						ERROR(`wrong # args: should be "regexp SubexpNames re"`),
					))
				})
				Specify("invalid regexp value", func() {
					Expect(execute("go:regexp SubexpNames a")).To(Equal(
						ERROR("invalid regexp value"),
					))
				})
			})
		})

		Describe("Exceptions", func() {
			Specify("wrong arity", func() {
				Expect(execute("go:regexp")).To(Equal(
					ERROR(`wrong # args: should be "regexp method ?arg ...?"`),
				))
			})
			Specify("unknown method", func() {
				Expect(execute("go:regexp unknownMethod")).To(Equal(
					ERROR(`unknown method "unknownMethod"`),
				))
			})
		})

	})
})
