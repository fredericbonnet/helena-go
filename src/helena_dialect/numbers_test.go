package helena_dialect_test

import (
	"math"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"helena/core"
	. "helena/helena_dialect"
)

var _ = Describe("Helena numbers", func() {
	var rootScope *Scope

	var tokenizer core.Tokenizer
	var parser *core.Parser

	parse := func(script string) *core.Script {
		return parser.Parse(tokenizer.Tokenize(script), nil).Script
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

	Describe("Number commands", func() {
		Describe("Integer numbers", func() {
			It("are valid commands", func() {
				Expect(evaluate("1")).To(Equal(INT(1)))
			})
			It("are idempotent", func() {
				Expect(evaluate("[1]")).To(Equal(INT(1)))
			})
			It("can be expressed as strings", func() {
				Expect(evaluate(`"123"`)).To(Equal(INT(123)))
			})
			It("should not take precedence over named commands", func() {
				Expect(evaluate("123")).To(Equal(INT(123)))
				Expect(evaluate("macro 123 {} {idem value}; 123")).To(Equal(STR("value")))
			})

			Describe("Exceptions", func() {
				Specify("unknown subcommand", func() {
					Expect(execute("1 unknownSubcommand")).To(Equal(
						ERROR(`unknown subcommand "unknownSubcommand"`),
					))
				})
				Specify("invalid subcommand name", func() {
					Expect(execute("1 []")).To(Equal(ERROR("invalid subcommand name")))
				})
			})
		})

		Describe("Real numbers", func() {
			It("are valid commands", func() {
				Expect(evaluate("1.25")).To(Equal(REAL(1.25)))
			})
			It("are idempotent", func() {
				Expect(evaluate("[1.25]")).To(Equal(REAL(1.25)))
			})
			It("can be expressed as strings", func() {
				Expect(evaluate(`"0.5"`)).To(Equal(REAL(0.5)))
			})
			It("should not take precedence over named commands", func() {
				Expect(evaluate("12.3")).To(Equal(REAL(12.3)))
				Expect(evaluate("macro 12.3 {} {idem value}; 12.3")).To(Equal(STR("value")))
			})

			Describe("Exceptions", func() {
				Specify("unknown subcommand", func() {
					Expect(execute("1.23 unknownSubcommand")).To(Equal(
						ERROR(`unknown subcommand "unknownSubcommand"`),
					))
				})
				Specify("invalid subcommand name", func() {
					Expect(execute("1.23 []")).To(Equal(ERROR("invalid subcommand name")))
				})
			})
		})

		Describe("Infix operators", func() {
			Describe("Arithmetic", func() {
				Specify("`+`", func() {
					Expect(evaluate("1 + 2")).To(Equal(INT(3)))
					Expect(evaluate("1 + 2 + 3 + 4")).To(Equal(INT(10)))
				})

				Specify("`-`", func() {
					Expect(evaluate("1 - 2")).To(Equal(INT(-1)))
					Expect(evaluate("1 - 2 - 3 - 4")).To(Equal(INT(-8)))
				})

				Specify("`*`", func() {
					Expect(evaluate("1 * 2")).To(Equal(INT(2)))
					Expect(evaluate("1 * 2 * 3 * 4")).To(Equal(INT(24)))
				})

				Specify("`/`", func() {
					Expect(evaluate("1 / 2")).To(Equal(REAL(0.5)))
					Expect(evaluate("1 / 2 / 4 / 8")).To(Equal(REAL(0.015625)))
					Expect(evaluate("1 / 0")).To(Equal(REAL(math.Inf(1))))
					Expect(evaluate("-1 / 0")).To(Equal(REAL(math.Inf(-1))))
					Expect(math.IsNaN(evaluate("0 / 0").(core.RealValue).Value))
				})

				Specify("Precedence rules", func() {
					Expect(evaluate("1 + 2 * 3 * 4 + 5")).To(Equal(INT(30)))
					Expect(evaluate("1 * 2 + 3 * 4 + 5 + 6 * 7")).To(Equal(INT(61)))
					Expect(evaluate("1 - 2 * 3 * 4 + 5")).To(Equal(INT(-18)))
					Expect(evaluate("1 - 2 * 3 / 4 + 5 * 6 / 10")).To(Equal(REAL(2.5)))
					Expect(evaluate("10 / 2 / 5")).To(Equal(INT(1)))
				})

				Specify("Conversions", func() {
					Expect(evaluate("1 + 2.3")).To(Equal(REAL(3.3)))
					Expect(evaluate("1.5 + 2.5")).To(Equal(INT(4)))
				})

				Describe("Exceptions", func() {
					Specify("wrong arity", func() {
						Expect(execute("1 +")).To(Equal(
							ERROR(
								`wrong # operands: should be "operand ?operator operand? ?...?"`,
							),
						))
					})
					Specify("invalid value", func() {
						Expect(execute("1 + a")).To(Equal(ERROR(`invalid number "a"`)))
					})
					Specify("unknown operator", func() {
						Expect(execute("1 + 2 a 3")).To(Equal(ERROR(`invalid operator "a"`)))
					})
					Specify("invalid operator", func() {
						Expect(execute("1 + 2 [] 3")).To(Equal(ERROR("invalid operator")))
					})
				})
			})

			Describe("Comparisons", func() {
				Describe("`==`", func() {
					It("should compare two numbers", func() {
						Expect(evaluate(`"123" == -34`)).To(Equal(FALSE))
						Expect(evaluate(`56 == "56.0"`)).To(Equal(TRUE))
						Expect(evaluate("set var 1; $var == $var")).To(Equal(TRUE))
					})

					Describe("Exceptions", func() {
						Specify("wrong arity", func() {
							Expect(execute("1 ==")).To(Equal(
								ERROR(`wrong # operands: should be "operand1 == operand2"`),
							))
							Expect(execute("1 == 2 3")).To(Equal(
								ERROR(`wrong # operands: should be "operand1 == operand2"`),
							))
						})
						Specify("invalid value", func() {
							Expect(execute("1 == a")).To(Equal(ERROR(`invalid number "a"`)))
						})
					})
				})

				Describe("`!=`", func() {
					It("should compare two numbers", func() {
						Expect(evaluate(`"123" != -34`)).To(Equal(TRUE))
						Expect(evaluate(`56 != "56.0"`)).To(Equal(FALSE))
						Expect(evaluate("set var 1; $var != $var")).To(Equal(FALSE))
					})

					Describe("Exceptions", func() {
						Specify("wrong arity", func() {
							Expect(execute("1 !=")).To(Equal(
								ERROR(`wrong # operands: should be "operand1 != operand2"`),
							))
							Expect(execute("1 != 2 3")).To(Equal(
								ERROR(`wrong # operands: should be "operand1 != operand2"`),
							))
						})
						Specify("invalid value", func() {
							Expect(execute("1 != a")).To(Equal(ERROR(`invalid number "a"`)))
						})
					})
				})

				Describe("`>`", func() {
					It("should compare two numbers", func() {
						Expect(evaluate("12 > -34")).To(Equal(TRUE))
						Expect(evaluate(`56 > "56.0"`)).To(Equal(FALSE))
						Expect(evaluate("set var 1; $var > $var")).To(Equal(FALSE))
					})

					Describe("Exceptions", func() {
						Specify("wrong arity", func() {
							Expect(execute("1 >")).To(Equal(
								ERROR(`wrong # operands: should be "operand1 > operand2"`),
							))
							Expect(execute("1 > 2 3")).To(Equal(
								ERROR(`wrong # operands: should be "operand1 > operand2"`),
							))
						})
						Specify("invalid value", func() {
							Expect(execute("1 > a")).To(Equal(ERROR(`invalid number "a"`)))
						})
					})
				})

				Describe("`>=`", func() {
					It("should compare two numbers", func() {
						Expect(evaluate("12 >= -34")).To(Equal(TRUE))
						Expect(evaluate(`56 >= "56.0"`)).To(Equal(TRUE))
						Expect(evaluate("set var 1; $var >= $var")).To(Equal(TRUE))
					})

					Describe("Exceptions", func() {
						Specify("wrong arity", func() {
							Expect(execute("1 >=")).To(Equal(
								ERROR(`wrong # operands: should be "operand1 >= operand2"`),
							))
							Expect(execute("1 >= 2 3")).To(Equal(
								ERROR(`wrong # operands: should be "operand1 >= operand2"`),
							))
						})
						Specify("invalid value", func() {
							Expect(execute("1 >= a")).To(Equal(ERROR(`invalid number "a"`)))
						})
					})
				})

				Describe("`<`", func() {
					It("should compare two numbers", func() {
						Expect(evaluate("12 < -34")).To(Equal(FALSE))
						Expect(evaluate(`56 < "56.0"`)).To(Equal(FALSE))
						Expect(evaluate("set var 1; $var < $var")).To(Equal(FALSE))
					})

					Describe("Exceptions", func() {
						Specify("wrong arity", func() {
							Expect(execute("1 <")).To(Equal(
								ERROR(`wrong # operands: should be "operand1 < operand2"`),
							))
							Expect(execute("1 < 2 3")).To(Equal(
								ERROR(`wrong # operands: should be "operand1 < operand2"`),
							))
						})
						Specify("invalid value", func() {
							Expect(execute("1 < a")).To(Equal(ERROR(`invalid number "a"`)))
						})
					})
				})

				Describe("`<=`", func() {
					It("should compare two numbers", func() {
						Expect(evaluate("12 <= -34")).To(Equal(FALSE))
						Expect(evaluate(`56 <= "56.0"`)).To(Equal(TRUE))
						Expect(evaluate("set var 1; $var <= $var")).To(Equal(TRUE))
					})

					Describe("Exceptions", func() {
						Specify("wrong arity", func() {
							Expect(execute("1 <=")).To(Equal(
								ERROR(`wrong # operands: should be "operand1 <= operand2"`),
							))
							Expect(execute("1 <= 2 3")).To(Equal(
								ERROR(`wrong # operands: should be "operand1 <= operand2"`),
							))
						})
						Specify("invalid value", func() {
							Expect(execute("1 <= a")).To(Equal(ERROR(`invalid number "a"`)))
						})
					})
				})
			})
		})

		Describe("Subcommands", func() {
			Describe("Introspection", func() {
				Describe("`subcommands`", func() {
					It("should return list of subcommands", func() {
						Expect(evaluate("1 subcommands")).To(Equal(
							evaluate("list (subcommands + - * / == != > >= < <=)"),
						))
						Expect(evaluate("1.2 subcommands")).To(Equal(
							evaluate("list (subcommands + - * / == != > >= < <=)"),
						))
					})

					Describe("Exceptions", func() {
						Specify("wrong arity", func() {
							Expect(execute("1 subcommands a")).To(Equal(
								ERROR(`wrong # args: should be "<number> subcommands"`),
							))
							Expect(execute("1.2 subcommands a")).To(Equal(
								ERROR(`wrong # args: should be "<number> subcommands"`),
							))
						})
					})
				})
			})
		})
	})

	Describe("int", func() {
		Describe("Integer conversion", func() {
			It("should return integer value", func() {
				Expect(evaluate("int 0")).To(Equal(INT(0)))
			})

			Describe("Exceptions", func() {
				Specify("values with no string representation", func() {
					Expect(execute("int []")).To(Equal(
						ERROR("value has no string representation"),
					))
					Expect(execute("int ()")).To(Equal(
						ERROR("value has no string representation"),
					))
				})
				Specify("invalid values", func() {
					Expect(execute("int a")).To(Equal(ERROR(`invalid integer "a"`)))
				})
				Specify("real values", func() {
					Expect(execute("int 1.1")).To(Equal(ERROR(`invalid integer "1.1"`)))
				})
			})
		})

		Describe("Subcommands", func() {
			Describe("Introspection", func() {
				Describe("`subcommands`", func() {
					Specify("usage", func() {
						Expect(evaluate("help int 0 subcommands")).To(Equal(
							STR("int value subcommands"),
						))
					})

					It("should return list of subcommands", func() {
						Expect(evaluate("int 0 subcommands")).To(Equal(
							evaluate("list (subcommands)"),
						))
					})

					Describe("Exceptions", func() {
						Specify("wrong arity", func() {
							Expect(execute("int 0 subcommands a")).To(Equal(
								ERROR(`wrong # args: should be "int value subcommands"`),
							))
							Expect(execute("help int 0 subcommands a")).To(Equal(
								ERROR(`wrong # args: should be "int value subcommands"`),
							))
						})
					})
				})
			})

			Describe("Exceptions", func() {
				Specify("unknown subcommand", func() {
					Expect(execute("int 0 unknownSubcommand")).To(Equal(
						ERROR(`unknown subcommand "unknownSubcommand"`),
					))
				})
				Specify("invalid subcommand name", func() {
					Expect(execute("int 0 []")).To(Equal(ERROR("invalid subcommand name")))
				})
			})
		})

		Describe("Ensemble command", func() {
			It("should return its ensemble metacommand when called with no argument", func() {
				Expect(evaluate("int").Type()).To(Equal(core.ValueType_COMMAND))
			})
			It("should be extensible", func() {
				evaluate(`
					[int] eval {
						macro foo {value} {idem bar}
					}
				`)
				Expect(evaluate("int example foo")).To(Equal(STR("bar")))
			})
			It("should support help for custom subcommands", func() {
				evaluate(`
					[int] eval {
						macro foo {value a b} {idem bar}
					}
				`)
				Expect(evaluate("help int 0 foo")).To(Equal(STR("int value foo a b")))
				Expect(execute("help int 0 foo 1 2 3")).To(Equal(
					ERROR(`wrong # args: should be "int value foo a b"`),
				))
			})

			Describe("Examples", func() {
				Specify("Adding a `positive` subcommand", func() {
					example([]exampleSpec{
						{
							script: `
								[int] eval {
									macro positive {value} {
										$value > 0
									}
								}
							`,
						},
						{
							script: "int 1 positive",
							result: TRUE,
						},
						{
							script: "int 0 positive",
							result: FALSE,
						},
						{
							script: "int -1 positive",
							result: FALSE,
						},
					})
				})
			})
		})
	})

	Describe("real", func() {
		Describe("Real conversion", func() {
			It("should return real value", func() {
				Expect(evaluate("real 0")).To(Equal(REAL(0)))
			})

			Describe("Exceptions", func() {
				Specify("values with no string representation", func() {
					Expect(execute("real []")).To(Equal(
						ERROR("value has no string representation"),
					))
					Expect(execute("real ()")).To(Equal(
						ERROR("value has no string representation"),
					))
				})
				Specify("invalid values", func() {
					Expect(execute("real a")).To(Equal(ERROR(`invalid number "a"`)))
				})
			})
		})

		Describe("Subcommands", func() {
			Describe("Introspection", func() {
				Describe("`subcommands`", func() {
					Specify("usage", func() {
						Expect(evaluate("help real 0 subcommands")).To(Equal(
							STR("real value subcommands"),
						))
					})

					It("should return list of subcommands", func() {
						Expect(evaluate("real 0 subcommands")).To(Equal(
							evaluate("list (subcommands)"),
						))
					})

					Describe("Exceptions", func() {
						Specify("wrong arity", func() {
							Expect(execute("real 0 subcommands a")).To(Equal(
								ERROR(`wrong # args: should be "real value subcommands"`),
							))
							Expect(execute("help real 0 subcommands a")).To(Equal(
								ERROR(`wrong # args: should be "real value subcommands"`),
							))
						})
					})
				})
			})

			Describe("Exceptions", func() {
				Specify("unknown subcommand", func() {
					Expect(execute("real 0 unknownSubcommand")).To(Equal(
						ERROR(`unknown subcommand "unknownSubcommand"`),
					))
				})
				Specify("invalid subcommand name", func() {
					Expect(execute("real 0 []")).To(Equal(ERROR("invalid subcommand name")))
				})
			})
		})

		Describe("Ensemble command", func() {
			It("should return its ensemble metacommand when called with no argument", func() {
				Expect(evaluate("real").Type()).To(Equal(core.ValueType_COMMAND))
			})
			It("should be extensible", func() {
				evaluate(`
					[real] eval {
						macro foo {value} {idem bar}
					}
				`)
				Expect(evaluate("real 0 foo")).To(Equal(STR("bar")))
			})
			It("should support help for custom subcommands", func() {
				evaluate(`
					[real] eval {
						macro foo {value a b} {idem bar}
					}
				`)
				Expect(evaluate("help real 0 foo")).To(Equal(STR("real value foo a b")))
				Expect(execute("help real 0 foo 1 2 3")).To(Equal(
					ERROR(`wrong # args: should be "real value foo a b"`),
				))
			})

			Describe("Examples", func() {
				Specify("Adding a `positive` subcommand", func() {
					example([]exampleSpec{
						{
							script: `
								[real] eval {
									macro positive {value} {
										$value > 0
									}
								}
							`,
						},
						{
							script: "real 0.1 positive",
							result: TRUE,
						},
						{
							script: "real 0 positive",
							result: FALSE,
						},
						{
							script: "real -1 positive",
							result: FALSE,
						},
					})
				})
			})
		})
	})
})
