package helena_dialect_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"helena/core"
	. "helena/helena_dialect"
)

var _ = Describe("Helena strings", func() {
	var rootScope *Scope

	var tokenizer core.Tokenizer
	var parser *core.Parser

	parse := func(script string) *core.Script {
		return parser.ParseTokens(tokenizer.Tokenize(script), nil).Script
	}
	prepareScript := func(script string) *Process {
		return rootScope.PrepareProcess(rootScope.Compile(*parse(script)))
	}
	execute := func(script string) core.Result {
		return prepareScript(script).Run()
	}
	evaluate := func(script string) core.Value {
		return execute(script).Value
	}
	init := func() {
		rootScope = NewRootScope(nil)
		InitCommands(rootScope)

		tokenizer = core.Tokenizer{}
		parser = core.NewParser(nil)
	}

	example := specifyExample(func(spec exampleSpec) core.Result { return execute(spec.script) })

	BeforeEach(init)

	Describe("string", func() {
		Describe("String creation and conversion", func() {
			It("should return string value", func() {
				Expect(evaluate("string example")).To(Equal(STR("example")))
			})
			It("should convert non-string values to strings", func() {
				Expect(evaluate("string [+ 1 3]")).To(Equal(STR("4")))
			})

			Describe("Exceptions", func() {
				Specify("values with no string representation", func() {
					Expect(execute("string []")).To(Equal(
						ERROR("value has no string representation"),
					))
					Expect(execute("string ()")).To(Equal(
						ERROR("value has no string representation"),
					))
				})
			})
		})

		Describe("Subcommands", func() {
			Describe("Introspection", func() {
				Describe("`subcommands`", func() {
					Specify("usage", func() {
						Expect(evaluate(`help string "" subcommands`)).To(Equal(
							STR("string value subcommands"),
						))
					})

					It("should return list of subcommands", func() {
						Expect(evaluate(`list [string "" subcommands] sort`)).To(Equal(
							evaluate(
								"list (subcommands length at range append remove insert replace == != > >= < <=) sort",
							),
						))
					})

					Describe("Exceptions", func() {
						Specify("wrong arity", func() {
							Expect(execute(`string "" subcommands a`)).To(Equal(
								ERROR(`wrong # args: should be "string value subcommands"`),
							))
							Expect(execute(`help string "" subcommands a`)).To(Equal(
								ERROR(`wrong # args: should be "string value subcommands"`),
							))
						})
					})
				})
			})

			Describe("Accessors", func() {
				Describe("`length`", func() {
					Specify("usage", func() {
						Expect(evaluate(`help string "" length`)).To(Equal(
							STR("string value length"),
						))
					})

					It("should return the string length", func() {
						Expect(evaluate(`string "" length`)).To(Equal(INT(0)))
						Expect(evaluate("string example length")).To(Equal(INT(7)))
					})

					Describe("Exceptions", func() {
						Specify("wrong arity", func() {
							Expect(execute("string example length a")).To(Equal(
								ERROR(`wrong # args: should be "string value length"`),
							))
							Expect(execute("help string example length a")).To(Equal(
								ERROR(`wrong # args: should be "string value length"`),
							))
						})
					})
				})

				Describe("`at`", func() {
					Specify("usage", func() {
						Expect(evaluate(`help string "" at`)).To(Equal(
							STR("string value at index ?default?"),
						))
					})

					It("should return the character at `index`", func() {
						Expect(evaluate("string example at 1")).To(Equal(STR("x")))
					})
					It("should return the default value for an out-of-range `index`", func() {
						Expect(evaluate("string example at 10 default")).To(Equal(
							STR("default"),
						))
					})
					Specify("`at` <-> indexed selector equivalence", func() {
						rootScope.SetNamedVariable("v", STR("example"))
						evaluate("set s (string $v)")

						Expect(execute("string $v at 2")).To(Equal(execute("idem $v[2]")))
						Expect(execute("$s at 2")).To(Equal(execute("idem $v[2]")))
						Expect(execute("idem $[$s][2]")).To(Equal(execute("idem $v[2]")))

						Expect(execute("string $v at -1")).To(Equal(execute("idem $v[-1]")))
						Expect(execute("$s at -1")).To(Equal(execute("idem $v[-1]")))
						Expect(execute("idem $[$s][-1]")).To(Equal(execute("idem $v[-1]")))
					})

					Describe("Exceptions", func() {
						Specify("wrong arity", func() {
							Expect(execute("string example at")).To(Equal(
								ERROR(
									`wrong # args: should be "string value at index ?default?"`,
								),
							))
							Expect(execute("string example at a b c")).To(Equal(
								ERROR(
									`wrong # args: should be "string value at index ?default?"`,
								),
							))
							Expect(execute("help string example at a b c")).To(Equal(
								ERROR(
									`wrong # args: should be "string value at index ?default?"`,
								),
							))
						})
						Specify("invalid `index`", func() {
							Expect(execute("string example at a")).To(Equal(
								ERROR(`invalid integer "a"`),
							))
						})
						Specify("`index` out of range", func() {
							Expect(execute("string example at -1")).To(Equal(
								ERROR(`index out of range "-1"`),
							))
							Expect(execute("string example at 10")).To(Equal(
								ERROR(`index out of range "10"`),
							))
						})
					})
				})
			})

			Describe("Operations", func() {
				Describe("`range`", func() {
					Specify("usage", func() {
						Expect(evaluate(`help string "" range`)).To(Equal(
							STR("string value range first ?last?"),
						))
					})

					It("should return the string included within [`first`, `last`]", func() {
						Expect(evaluate("string example range 1 3")).To(Equal(STR("xam")))
					})
					It("should return the remainder of the string when given `first` only", func() {
						Expect(evaluate("string example range 2")).To(Equal(STR("ample")))
					})
					It("should truncate out of range boundaries", func() {
						Expect(evaluate("string example range -1")).To(Equal(STR("example")))
						Expect(evaluate("string example range -10 1")).To(Equal(STR("ex")))
						Expect(evaluate("string example range 2 10")).To(Equal(STR("ample")))
						Expect(evaluate("string example range -2 10")).To(Equal(
							STR("example"),
						))
					})
					It("should return an empty string when last is before `first`", func() {
						Expect(evaluate("string example range 2 0")).To(Equal(STR("")))
					})
					It("should return an empty string when `first` is past the string length", func() {
						Expect(evaluate("string example range 10 12")).To(Equal(STR("")))
					})
					It("should return an empty string when `last` is negative", func() {
						Expect(evaluate("string example range -3 -1")).To(Equal(STR("")))
					})

					Describe("Exceptions", func() {
						Specify("wrong arity", func() {
							Expect(execute("string example range")).To(Equal(
								ERROR(
									`wrong # args: should be "string value range first ?last?"`,
								),
							))
							Expect(execute("string example range a b c")).To(Equal(
								ERROR(
									`wrong # args: should be "string value range first ?last?"`,
								),
							))
							Expect(execute("help string example range a b c")).To(Equal(
								ERROR(
									`wrong # args: should be "string value range first ?last?"`,
								),
							))
						})
						Specify("invalid `index`", func() {
							Expect(execute("string example range a")).To(Equal(
								ERROR(`invalid integer "a"`),
							))
							Expect(execute("string example range 1 b")).To(Equal(
								ERROR(`invalid integer "b"`),
							))
						})
					})
				})

				Describe("`remove`", func() {
					Specify("usage", func() {
						Expect(evaluate(`help string "" remove`)).To(Equal(
							STR("string value remove first last"),
						))
					})

					It("should remove the range included within [`first`, `last`]", func() {
						Expect(evaluate("string example remove 1 3")).To(Equal(STR("eple")))
					})
					It("should truncate out of range boundaries", func() {
						Expect(evaluate("string example remove -10 1")).To(Equal(
							STR("ample"),
						))
						Expect(evaluate("string example remove 2 10")).To(Equal(STR("ex")))
						Expect(evaluate("string example remove -2 10")).To(Equal(STR("")))
					})
					It("should do nothing when `last` is before `first`", func() {
						Expect(evaluate("string example remove 2 0")).To(Equal(
							STR("example"),
						))
					})
					It("should do nothing when `last` is negative", func() {
						Expect(evaluate("string example remove -3 -1")).To(Equal(
							STR("example"),
						))
					})
					It("should do nothing when `first` is past the string length", func() {
						Expect(evaluate("string example remove 10 12")).To(Equal(
							STR("example"),
						))
					})

					Describe("Exceptions", func() {
						Specify("wrong arity", func() {
							Expect(execute("string example remove a")).To(Equal(
								ERROR(
									`wrong # args: should be "string value remove first last"`,
								),
							))
							Expect(execute("string example remove a b c d")).To(Equal(
								ERROR(
									`wrong # args: should be "string value remove first last"`,
								),
							))
							Expect(execute("help string example remove a b c d")).To(Equal(
								ERROR(
									`wrong # args: should be "string value remove first last"`,
								),
							))
						})
						Specify("invalid `index`", func() {
							Expect(execute("string example remove a b")).To(Equal(
								ERROR(`invalid integer "a"`),
							))
							Expect(execute("string example remove 1 b")).To(Equal(
								ERROR(`invalid integer "b"`),
							))
						})
					})
				})

				Describe("`append`", func() {
					Specify("usage", func() {
						Expect(evaluate(`help string "" append`)).To(Equal(
							STR("string value append ?string ...?"),
						))
					})

					It("should append two strings", func() {
						Expect(evaluate("string example append foo")).To(Equal(
							STR("examplefoo"),
						))
					})
					It("should accept several strings", func() {
						Expect(evaluate("string example append foo bar baz")).To(Equal(
							STR("examplefoobarbaz"),
						))
					})
					It("should accept zero string", func() {
						Expect(evaluate("string example append")).To(Equal(STR("example")))
					})

					Describe("Exceptions", func() {
						Specify("values with no string representation", func() {
							Expect(execute("string example append []")).To(Equal(
								ERROR("value has no string representation"),
							))
							Expect(execute("string example append ()")).To(Equal(
								ERROR("value has no string representation"),
							))
						})
					})
				})

				Describe("`insert`", func() {
					Specify("usage", func() {
						Expect(evaluate(`help string "" insert`)).To(Equal(
							STR("string value insert index value2"),
						))
					})

					It("should insert `string` at `index`", func() {
						Expect(evaluate("string example insert 1 foo")).To(Equal(
							STR("efooxample"),
						))
					})
					It("should prepend `string` when `index` is negative", func() {
						Expect(evaluate("string example insert -10 foo")).To(Equal(
							STR("fooexample"),
						))
					})
					It("should append `string` when `index` is past the target string length", func() {
						Expect(evaluate("string example insert 10 foo")).To(Equal(
							STR("examplefoo"),
						))
					})

					Describe("Exceptions", func() {
						Specify("wrong arity", func() {
							Expect(execute("string example insert a")).To(Equal(
								ERROR(
									`wrong # args: should be "string value insert index value2"`,
								),
							))
							Expect(execute("string example insert a b c")).To(Equal(
								ERROR(
									`wrong # args: should be "string value insert index value2"`,
								),
							))
							Expect(execute("help string example insert a b c")).To(Equal(
								ERROR(
									`wrong # args: should be "string value insert index value2"`,
								),
							))
						})
						Specify("invalid `index`", func() {
							Expect(execute("string example insert a b")).To(Equal(
								ERROR(`invalid integer "a"`),
							))
						})
						Specify("values with no string representation", func() {
							Expect(execute("string example insert 1 []")).To(Equal(
								ERROR("value has no string representation"),
							))
							Expect(execute("string example insert 1 ()")).To(Equal(
								ERROR("value has no string representation"),
							))
						})
					})
				})

				Describe("`replace`", func() {
					Specify("usage", func() {
						Expect(evaluate(`help string "" replace`)).To(Equal(
							STR("string value replace first last value2"),
						))
					})

					It("should replace the range included within [`first`, `last`] with `string`", func() {
						Expect(evaluate("string example replace 1 3 foo")).To(Equal(
							STR("efoople"),
						))
					})
					It("should truncate out of range boundaries", func() {
						Expect(evaluate("string example replace -10 1 foo")).To(Equal(
							STR("fooample"),
						))
						Expect(evaluate("string example replace 2 10 foo")).To(Equal(
							STR("exfoo"),
						))
						Expect(evaluate("string example replace -2 10 foo")).To(Equal(
							STR("foo"),
						))
					})
					It("should insert `string` at `first` index when `last` is before `first`", func() {
						Expect(evaluate("string example replace 2 0 foo")).To(Equal(
							STR("exfooample"),
						))
					})
					It("should prepend `string` when `last` is negative", func() {
						Expect(evaluate("string example replace -3 -1 foo")).To(Equal(
							STR("fooexample"),
						))
					})
					It("should append `string` when `first` is past the target string length", func() {
						Expect(evaluate("string example replace 10 12 foo")).To(Equal(
							STR("examplefoo"),
						))
					})

					Describe("Exceptions", func() {
						Specify("wrong arity", func() {
							Expect(execute("string example replace a b")).To(Equal(
								ERROR(
									`wrong # args: should be "string value replace first last value2"`,
								),
							))
							Expect(execute("string example replace a b c d")).To(Equal(
								ERROR(
									`wrong # args: should be "string value replace first last value2"`,
								),
							))
							Expect(execute("help string example replace a b c d")).To(Equal(
								ERROR(
									`wrong # args: should be "string value replace first last value2"`,
								),
							))
						})
						Specify("invalid index", func() {
							Expect(execute("string example replace a b c")).To(Equal(
								ERROR(`invalid integer "a"`),
							))
							Expect(execute("string example replace 1 b c")).To(Equal(
								ERROR(`invalid integer "b"`),
							))
						})
						Specify("values with no string representation", func() {
							Expect(execute("string example replace 1 3 []")).To(Equal(
								ERROR("value has no string representation"),
							))
							Expect(execute("string example replace 1 3 ()")).To(Equal(
								ERROR("value has no string representation"),
							))
						})
					})
				})
			})

			Describe("String comparisons", func() {
				Describe("`==`", func() {
					Specify("usage", func() {
						Expect(evaluate(`help string "" ==`)).To(Equal(
							STR("string value1 == value2"),
						))
					})

					It("should compare two strings", func() {
						Expect(evaluate("string example == foo")).To(Equal(FALSE))
						Expect(evaluate("string example == example")).To(Equal(TRUE))
						Expect(evaluate("set var example; string $var == $var")).To(Equal(
							TRUE,
						))
					})

					Describe("Exceptions", func() {
						Specify("wrong arity", func() {
							Expect(execute("string a ==")).To(Equal(
								ERROR(`wrong # operands: should be "string value1 == value2"`),
							))
							Expect(execute("string a == b c")).To(Equal(
								ERROR(`wrong # operands: should be "string value1 == value2"`),
							))
							Expect(execute("help string a == b c")).To(Equal(
								ERROR(`wrong # operands: should be "string value1 == value2"`),
							))
						})
						Specify("values with no string representation", func() {
							Expect(execute("string example == []")).To(Equal(
								ERROR("value has no string representation"),
							))
							Expect(execute("string example == ()")).To(Equal(
								ERROR("value has no string representation"),
							))
						})
					})
				})

				Describe("`!=`", func() {
					Specify("usage", func() {
						Expect(evaluate(`help string "" !=`)).To(Equal(
							STR("string value1 != value2"),
						))
					})

					It("should compare two strings", func() {
						Expect(evaluate("string example != foo")).To(Equal(TRUE))
						Expect(evaluate("string example != example")).To(Equal(FALSE))
						Expect(evaluate("set var example; string $var != $var")).To(Equal(
							FALSE,
						))
					})

					Describe("Exceptions", func() {
						Specify("wrong arity", func() {
							Expect(execute("string a !=")).To(Equal(
								ERROR(`wrong # operands: should be "string value1 != value2"`),
							))
							Expect(execute("string a != b c")).To(Equal(
								ERROR(`wrong # operands: should be "string value1 != value2"`),
							))
							Expect(execute("help string a != b c")).To(Equal(
								ERROR(`wrong # operands: should be "string value1 != value2"`),
							))
						})
						Specify("values with no string representation", func() {
							Expect(execute("string example != []")).To(Equal(
								ERROR("value has no string representation"),
							))
							Expect(execute("string example != ()")).To(Equal(
								ERROR("value has no string representation"),
							))
						})
					})
				})

				Describe("`>`", func() {
					Specify("usage", func() {
						Expect(evaluate(`help string "" >`)).To(Equal(
							STR("string value1 > value2"),
						))
					})

					It("should compare two strings", func() {
						Expect(evaluate("string example > foo")).To(Equal(FALSE))
						Expect(evaluate("string example > example")).To(Equal(FALSE))
						Expect(evaluate("set var example; string $var > $var")).To(Equal(
							FALSE,
						))
					})

					Describe("Exceptions", func() {
						Specify("wrong arity", func() {
							Expect(execute("string a >")).To(Equal(
								ERROR(`wrong # operands: should be "string value1 > value2"`),
							))
							Expect(execute("string a > b c")).To(Equal(
								ERROR(`wrong # operands: should be "string value1 > value2"`),
							))
							Expect(execute("help string a > b c")).To(Equal(
								ERROR(`wrong # operands: should be "string value1 > value2"`),
							))
						})
						Specify("values with no string representation", func() {
							Expect(execute("string example > []")).To(Equal(
								ERROR("value has no string representation"),
							))
							Expect(execute("string example > ()")).To(Equal(
								ERROR("value has no string representation"),
							))
						})
					})
				})

				Describe("`>=`", func() {
					Specify("usage", func() {
						Expect(evaluate(`help string "" >=`)).To(Equal(
							STR("string value1 >= value2"),
						))
					})

					It("should compare two strings", func() {
						Expect(evaluate("string example >= foo")).To(Equal(FALSE))
						Expect(evaluate("string example >= example")).To(Equal(TRUE))
						Expect(evaluate("set var example; string $var >= $var")).To(Equal(
							TRUE,
						))
					})

					Describe("Exceptions", func() {
						Specify("wrong arity", func() {
							Expect(execute("string a >=")).To(Equal(
								ERROR(`wrong # operands: should be "string value1 >= value2"`),
							))
							Expect(execute("string a >= b c")).To(Equal(
								ERROR(`wrong # operands: should be "string value1 >= value2"`),
							))
							Expect(execute("help string a >= b c")).To(Equal(
								ERROR(`wrong # operands: should be "string value1 >= value2"`),
							))
						})
						Specify("values with no string representation", func() {
							Expect(execute("string example >= []")).To(Equal(
								ERROR("value has no string representation"),
							))
							Expect(execute("string example >= ()")).To(Equal(
								ERROR("value has no string representation"),
							))
						})
					})
				})

				Describe("`<`", func() {
					Specify("usage", func() {
						Expect(evaluate(`help string "" <`)).To(Equal(
							STR("string value1 < value2"),
						))
					})

					It("should compare two strings", func() {
						Expect(evaluate("string example < foo")).To(Equal(TRUE))
						Expect(evaluate("string example < example")).To(Equal(FALSE))
						Expect(evaluate("set var example; string $var < $var")).To(Equal(
							FALSE,
						))
					})

					Describe("Exceptions", func() {
						Specify("wrong arity", func() {
							Expect(execute("string a <")).To(Equal(
								ERROR(`wrong # operands: should be "string value1 < value2"`),
							))
							Expect(execute("string a < b c")).To(Equal(
								ERROR(`wrong # operands: should be "string value1 < value2"`),
							))
							Expect(execute("help string a < b c")).To(Equal(
								ERROR(`wrong # operands: should be "string value1 < value2"`),
							))
						})
						Specify("values with no string representation", func() {
							Expect(execute("string example < []")).To(Equal(
								ERROR("value has no string representation"),
							))
							Expect(execute("string example < ()")).To(Equal(
								ERROR("value has no string representation"),
							))
						})
					})
				})

				Describe("`<=`", func() {
					Specify("usage", func() {
						Expect(evaluate(`help string "" <=`)).To(Equal(
							STR("string value1 <= value2"),
						))
					})

					It("should compare two strings", func() {
						Expect(evaluate("string example <= foo")).To(Equal(TRUE))
						Expect(evaluate("string example <= example")).To(Equal(TRUE))
						Expect(evaluate("set var example; string $var <= $var")).To(Equal(
							TRUE,
						))
					})

					Describe("Exceptions", func() {
						Specify("wrong arity", func() {
							Expect(execute("string a <=")).To(Equal(
								ERROR(`wrong # operands: should be "string value1 <= value2"`),
							))
							Expect(execute("string a <= b c")).To(Equal(
								ERROR(`wrong # operands: should be "string value1 <= value2"`),
							))
							Expect(execute("help string a <= b c")).To(Equal(
								ERROR(`wrong # operands: should be "string value1 <= value2"`),
							))
						})
						Specify("values with no string representation", func() {
							Expect(execute("string example <= []")).To(Equal(
								ERROR("value has no string representation"),
							))
							Expect(execute("string example <= ()")).To(Equal(
								ERROR("value has no string representation"),
							))
						})
					})
				})
			})

			Describe("Exceptions", func() {
				Specify("unknown subcommand", func() {
					Expect(execute(`string "" unknownSubcommand`)).To(Equal(
						ERROR(`unknown subcommand "unknownSubcommand"`),
					))
				})
				Specify("invalid subcommand name", func() {
					Expect(execute(`string "" []`)).To(Equal(
						ERROR("invalid subcommand name"),
					))
				})
			})
		})

		Describe("Examples", func() {
			Specify("Currying and encapsulation", func() {
				example([]exampleSpec{
					{
						script: "set s (string example)",
					},
					{
						script: "$s",
						result: STR("example"),
					},
					{
						script: "$s length",
						result: INT(7),
					},
					{
						script: "$s at 2",
						result: STR("a"),
					},
					{
						script: "$s range 3 5",
						result: STR("mpl"),
					},
					{
						script: "$s == example",
						result: TRUE,
					},
					{
						script: "$s > exercise",
						result: FALSE,
					},
				})
			})
			Specify("Argument type guard", func() {
				example([]exampleSpec{
					{
						script: "macro len ( (string s) ) {string $s length}",
					},
					{
						script: "len example",
						result: INT(7),
					},
					{
						script: "len (invalid value)",
						result: ERROR("value has no string representation"),
					},
				})
			})
		})

		Describe("Ensemble command", func() {
			It("should return its ensemble metacommand when called with no argument", func() {
				Expect(evaluate("string").Type()).To(Equal(core.ValueType_COMMAND))
			})
			It("should be extensible", func() {
				evaluate(`
					[string] eval {
						macro foo {value} {idem bar}
					}
        		`)
				Expect(evaluate("string example foo")).To(Equal(STR("bar")))
			})
			It("should support help for custom subcommands", func() {
				evaluate(`
          [string] eval {
            macro foo {value a b} {idem bar}
          }
        `)
				Expect(evaluate("help string example foo")).To(Equal(
					STR("string value foo a b"),
				))
				Expect(execute("help string example foo 1 2 3")).To(Equal(
					ERROR(`wrong # args: should be "string value foo a b"`),
				))
			})

			Describe("Examples", func() {
				Specify("Adding a `last` subcommand", func() {
					example([]exampleSpec{
						{
							script: `
								[string] eval {
									macro last {value} {
										string $value at [- [string $value length] 1]
									}
								}
							`,
						},
						{
							script: "string example last",
							result: STR("e"),
						},
					})
				})
				Specify("Adding a `+` operator", func() {
					example([]exampleSpec{
						{
							script: `
								[string] eval {
									macro + {str1 str2} {idem $str1$str2}
								}
							`,
						},
						{
							script: "string s1 + s2",
							result: STR("s1s2"),
						},
					})
				})
			})
		})
	})
})
