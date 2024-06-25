package helena_dialect_test

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"helena/core"
	. "helena/helena_dialect"
)

var _ = Describe("Helena logic operations", func() {
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

	Describe("Booleans", func() {
		It("are valid commands", func() {
			Expect(evaluate("true")).To(Equal(TRUE))
			Expect(evaluate("false")).To(Equal(FALSE))
		})
		It("are idempotent", func() {
			Expect(evaluate("[true]")).To(Equal(TRUE))
			Expect(evaluate("[false]")).To(Equal(FALSE))
		})

		Describe("Infix operators", func() {
			Describe("Conditional", func() {
				Describe("`?`", func() {
					Describe("`true`", func() {
						It("should return first argument", func() {
							Expect(evaluate("true ? a b")).To(Equal(STR("a")))
						})
						It("should support a single argument", func() {
							Expect(evaluate("true ? a")).To(Equal(STR("a")))
						})
					})
					Describe("`false`", func() {
						It("should return nil if no second argument is given", func() {
							Expect(evaluate("false ? a")).To(Equal(NIL))
						})
						It("should return second argument", func() {
							Expect(evaluate("false ? a b")).To(Equal(STR("b")))
						})
					})

					Describe("Exceptions", func() {
						Specify("wrong arity", func() {
							Expect(execute("true ?")).To(Equal(
								ERROR(`wrong # args: should be "true ? arg ?arg?"`),
							))
							Expect(execute("true ? a b c")).To(Equal(
								ERROR(`wrong # args: should be "true ? arg ?arg?"`),
							))
							Expect(execute("help true ? a b c")).To(Equal(
								ERROR(`wrong # args: should be "true ? arg ?arg?"`),
							))
							Expect(execute("false ?")).To(Equal(
								ERROR(`wrong # args: should be "false ? arg ?arg?"`),
							))
							Expect(execute("false ? a b c")).To(Equal(
								ERROR(`wrong # args: should be "false ? arg ?arg?"`),
							))
							Expect(execute("help false ? a b c")).To(Equal(
								ERROR(`wrong # args: should be "false ? arg ?arg?"`),
							))
						})
					})
				})

				Describe("`!?`", func() {
					Describe("`true`", func() {
						It("should return nil if no second argument is given", func() {
							Expect(evaluate("true !? a")).To(Equal(NIL))
						})
						It("should return second argument", func() {
							Expect(evaluate("true !? a b")).To(Equal(STR("b")))
						})
					})
					Describe("`false`", func() {
						It("should return first argument", func() {
							Expect(evaluate("false !? a b")).To(Equal(STR("a")))
						})
						It("should support a single argument", func() {
							Expect(evaluate("false !? a")).To(Equal(STR("a")))
						})
					})

					Describe("Exceptions", func() {
						Specify("wrong arity", func() {
							Expect(execute("true !?")).To(Equal(
								ERROR(`wrong # args: should be "true !? arg ?arg?"`),
							))
							Expect(execute("true !? a b c")).To(Equal(
								ERROR(`wrong # args: should be "true !? arg ?arg?"`),
							))
							Expect(execute("help true !? a b c")).To(Equal(
								ERROR(`wrong # args: should be "true !? arg ?arg?"`),
							))
							Expect(execute("false !?")).To(Equal(
								ERROR(`wrong # args: should be "false !? arg ?arg?"`),
							))
							Expect(execute("false !? a b c")).To(Equal(
								ERROR(`wrong # args: should be "false !? arg ?arg?"`),
							))
							Expect(execute("help false !? a b c")).To(Equal(
								ERROR(`wrong # args: should be "false !? arg ?arg?"`),
							))
						})
					})
				})
			})
		})

		Describe("Subcommands", func() {
			Describe("Introspection", func() {
				Describe("`subcommands`", func() {
					It("should return list of subcommands", func() {
						Expect(evaluate("true subcommands")).To(Equal(
							evaluate("list (subcommands ? !?)"),
						))
						Expect(evaluate("false subcommands")).To(Equal(
							evaluate("list (subcommands ? !?)"),
						))
					})

					Describe("Exceptions", func() {
						Specify("wrong arity", func() {
							Expect(execute("true subcommands a")).To(Equal(
								ERROR(`wrong # args: should be "true subcommands"`),
							))
							Expect(execute("help true subcommands a")).To(Equal(
								ERROR(`wrong # args: should be "true subcommands"`),
							))
							Expect(execute("false subcommands a")).To(Equal(
								ERROR(`wrong # args: should be "false subcommands"`),
							))
							Expect(execute("help false subcommands a")).To(Equal(
								ERROR(`wrong # args: should be "false subcommands"`),
							))
						})
					})
				})
			})
		})

		Describe("Exceptions", func() {
			Specify("unknown subcommand", func() {
				Expect(execute("true unknownSubcommand")).To(Equal(
					ERROR(`unknown subcommand "unknownSubcommand"`),
				))
				Expect(execute("false unknownSubcommand")).To(Equal(
					ERROR(`unknown subcommand "unknownSubcommand"`),
				))
			})
			Specify("invalid subcommand name", func() {
				Expect(execute("true []")).To(Equal(ERROR("invalid subcommand name")))
				Expect(execute("false []")).To(Equal(ERROR("invalid subcommand name")))
			})
		})
	})

	Describe("bool", func() {
		Describe("Boolean conversion", func() {
			It("should return boolean value", func() {
				Expect(evaluate("bool true")).To(Equal(TRUE))
				Expect(evaluate("bool false")).To(Equal(FALSE))
			})

			Describe("Exceptions", func() {
				Specify("values with no string representation", func() {
					Expect(execute("bool []")).To(Equal(
						ERROR("value has no string representation"),
					))
					Expect(execute("bool ()")).To(Equal(
						ERROR("value has no string representation"),
					))
				})
				Specify("invalid values", func() {
					Expect(execute("bool a")).To(Equal(ERROR(`invalid boolean "a"`)))
				})
			})
		})

		Describe("Subcommands", func() {
			Describe("Introspection", func() {
				Describe("`subcommands`", func() {
					Specify("usage", func() {
						Expect(evaluate("help bool true subcommands")).To(Equal(
							STR("bool value subcommands"),
						))
					})

					It("should return list of subcommands", func() {
						Expect(evaluate("bool true subcommands")).To(Equal(
							evaluate("list (subcommands)"),
						))
					})

					Describe("Exceptions", func() {
						Specify("wrong arity", func() {
							Expect(execute("bool true subcommands a")).To(Equal(
								ERROR(`wrong # args: should be "bool value subcommands"`),
							))
							Expect(execute("help bool true subcommands a")).To(Equal(
								ERROR(`wrong # args: should be "bool value subcommands"`),
							))
						})
					})
				})
			})

			Describe("Exceptions", func() {
				Specify("unknown subcommand", func() {
					Expect(execute("bool true unknownSubcommand")).To(Equal(
						ERROR(`unknown subcommand "unknownSubcommand"`),
					))
				})
				Specify("invalid subcommand name", func() {
					Expect(execute("bool true []")).To(Equal(
						ERROR("invalid subcommand name"),
					))
				})
			})
		})

		Describe("Ensemble command", func() {
			It("should return its ensemble metacommand when called with no argument", func() {
				Expect(evaluate("bool").Type()).To(Equal(core.ValueType_COMMAND))
			})
			It("should be extensible", func() {
				evaluate(`
					[bool] eval {
						macro foo {value} {idem bar}
					}
        `)
				Expect(evaluate("bool example foo")).To(Equal(STR("bar")))
			})
			It("should support help for custom subcommands", func() {
				evaluate(`
					[bool] eval {
						macro foo {value a b} {idem bar}
					}
				`)
				Expect(evaluate("help bool true foo")).To(Equal(
					STR("bool value foo a b"),
				))
				Expect(execute("help bool true foo 1 2 3")).To(Equal(
					ERROR(`wrong # args: should be "bool value foo a b"`),
				))
			})

			Describe("Examples", func() {
				Specify("Adding a `xor` subcommand", func() {
					example([]exampleSpec{
						{
							script: `
								[bool] eval {
									macro xor {(bool value1) (bool value2)} {
										$value1 ? [! $value2] $value2
									}
								}
							`,
						},
						{
							script: "bool true xor false",
							result: TRUE,
						},
						{
							script: "bool true xor true",
							result: FALSE,
						},
						{
							script: "bool false xor false",
							result: FALSE,
						},
						{
							script: "bool false xor true",
							result: TRUE,
						},
					})
				})
			})
		})
	})

	Describe("Prefix operators", func() {
		Describe("`!`", func() {
			Describe("Specifications", func() {
				Specify("usage", func() {
					Expect(evaluate("help !")).To(Equal(STR("! arg")))
				})

				It("should invert boolean values", func() {
					Expect(evaluate("! true")).To(Equal(FALSE))
					Expect(evaluate("! false")).To(Equal(TRUE))
				})
				It("should accept script expressions", func() {
					Expect(evaluate("! {idem true}")).To(Equal(FALSE))
					Expect(evaluate("! {idem false}")).To(Equal(TRUE))
				})
			})

			Describe("Exceptions", func() {
				Specify("wrong arity", func() {
					Expect(execute("!")).To(Equal(ERROR(`wrong # args: should be "! arg"`)))
					Expect(execute("! a b")).To(Equal(
						ERROR(`wrong # args: should be "! arg"`),
					))
					Expect(execute("help ! a b")).To(Equal(
						ERROR(`wrong # args: should be "! arg"`),
					))
				})
				Specify("invalid value", func() {
					Expect(execute("! 1")).To(Equal(ERROR(`invalid boolean "1"`)))
					Expect(execute("! 1.23")).To(Equal(ERROR(`invalid boolean "1.23"`)))
					Expect(execute("! a")).To(Equal(ERROR(`invalid boolean "a"`)))
				})
			})

			Describe("Control flow", func() {
				Describe("`return`", func() {
					It("should interrupt expression with `RETURN` code", func() {
						Expect(execute("! {return value; unreachable}")).To(Equal(
							RETURN(STR("value")),
						))
					})
				})
				Describe("`tailcall`", func() {
					It("should interrupt expression with `RETURN` code", func() {
						Expect(execute("! {tailcall {idem value}; unreachable}")).To(Equal(
							RETURN(STR("value")),
						))
					})
				})
				Describe("`yield`", func() {
					It("should interrupt expression with `YIELD` code", func() {
						result := execute("! {yield value; true}")
						Expect(result.Code).To(Equal(core.ResultCode_YIELD))
						Expect(result.Value).To(Equal(STR("value")))
					})
					It("should provide a resumable state", func() {
						process := prepareScript(
							"! {yield val1; yield val2}",
						)

						result := process.Run()
						Expect(result.Code).To(Equal(core.ResultCode_YIELD))
						Expect(result.Value).To(Equal(STR("val1")))

						result = process.Run()
						Expect(result.Code).To(Equal(core.ResultCode_YIELD))
						Expect(result.Value).To(Equal(STR("val2")))

						process.YieldBack(TRUE)
						result = process.Run()
						Expect(result).To(Equal(OK(FALSE)))
					})
				})
				Describe("`error`", func() {
					It("should interrupt expression with `ERROR` code", func() {
						Expect(execute("! {error msg; false}")).To(Equal(ERROR("msg")))
					})
				})
				Describe("`break`", func() {
					It("should interrupt expression with `BREAK` code", func() {
						Expect(execute("! {break; unreachable}")).To(Equal(BREAK(NIL)))
					})
				})
				Describe("`continue`", func() {
					It("should interrupt expression with `CONTINUE` code", func() {
						Expect(execute("! {continue; false}")).To(Equal(CONTINUE(NIL)))
					})
				})
			})
		})

		Describe("`&&`", func() {
			Describe("Specifications", func() {
				Specify("usage", func() {
					Expect(evaluate("help &&")).To(Equal(STR("&& arg ?arg ...?")))
				})

				It("should accept one boolean", func() {
					Expect(evaluate("&& false")).To(Equal(FALSE))
					Expect(evaluate("&& true")).To(Equal(TRUE))
				})
				It("should accept two booleans", func() {
					Expect(evaluate("&& false false")).To(Equal(FALSE))
					Expect(evaluate("&& false true")).To(Equal(FALSE))
					Expect(evaluate("&& true false")).To(Equal(FALSE))
					Expect(evaluate("&& true true")).To(Equal(TRUE))
				})
				It("should accept several booleans", func() {
					Expect(evaluate("&&" + strings.Repeat(" true", 3))).To(Equal(TRUE))
					Expect(evaluate("&&" + strings.Repeat(" true", 3) + " false")).To(Equal(FALSE))
				})
				It("should accept script expressions", func() {
					Expect(evaluate("&& {idem false}")).To(Equal(FALSE))
					Expect(evaluate("&& {idem true}")).To(Equal(TRUE))
				})
				It("should short-circuit on `false`", func() {
					Expect(evaluate("&& false {unreachable}")).To(Equal(FALSE))
				})
			})

			Describe("Exceptions", func() {
				Specify("wrong arity", func() {
					Expect(execute("&&")).To(Equal(
						ERROR(`wrong # args: should be "&& arg ?arg ...?"`),
					))
				})
				Specify("invalid value", func() {
					Expect(execute("&& a")).To(Equal(ERROR(`invalid boolean "a"`)))
				})
			})

			Describe("Control flow", func() {
				Describe("`return`", func() {
					It("should interrupt expression with `RETURN` code", func() {
						Expect(execute("&& true {return value; unreachable} false")).To(Equal(
							RETURN(STR("value")),
						))
					})
				})
				Describe("`tailcall`", func() {
					It("should interrupt expression with `RETURN` code", func() {
						Expect(
							execute("&& true {tailcall {idem value}; unreachable} false"),
						).To(Equal(RETURN(STR("value"))))
					})
				})
				Describe("`yield`", func() {
					It("should interrupt expression with `YIELD` code", func() {
						result := execute("&& true {yield value; true}")
						Expect(result.Code).To(Equal(core.ResultCode_YIELD))
						Expect(result.Value).To(Equal(STR("value")))
					})
					It("should provide a resumable state", func() {
						process := prepareScript(
							"&& {yield val1} {yield val2}",
						)

						result := process.Run()
						Expect(result.Code).To(Equal(core.ResultCode_YIELD))
						Expect(result.Value).To(Equal(STR("val1")))

						process.YieldBack(TRUE)
						result = process.Run()
						Expect(result.Code).To(Equal(core.ResultCode_YIELD))
						Expect(result.Value).To(Equal(STR("val2")))

						process.YieldBack(FALSE)
						result = process.Run()
						Expect(result).To(Equal(OK(FALSE)))
					})
				})
				Describe("`error`", func() {
					It("should interrupt expression with `ERROR` code", func() {
						Expect(execute("&& true {error msg; true} false")).To(Equal(
							ERROR("msg"),
						))
					})
				})
				Describe("`break`", func() {
					It("should interrupt expression with `BREAK` code", func() {
						Expect(execute("&& true {break; unreachable} false")).To(Equal(
							BREAK(NIL),
						))
					})
				})
				Describe("`continue`", func() {
					It("should interrupt expression with `CONTINUE` code", func() {
						Expect(execute("&& true {continue; unreachable} false")).To(Equal(
							CONTINUE(NIL),
						))
					})
				})
			})
		})

		Describe("`||`", func() {
			Describe("Specifications", func() {
				Specify("usage", func() {
					Expect(evaluate("help ||")).To(Equal(STR("|| arg ?arg ...?")))
				})

				It("should accept one boolean", func() {
					Expect(evaluate("|| false")).To(Equal(FALSE))
					Expect(evaluate("|| true")).To(Equal(TRUE))
				})
				It("should accept two booleans", func() {
					Expect(evaluate("|| false false")).To(Equal(FALSE))
					Expect(evaluate("|| false true")).To(Equal(TRUE))
					Expect(evaluate("|| true false")).To(Equal(TRUE))
					Expect(evaluate("|| true true")).To(Equal(TRUE))
				})
				It("should accept several booleans", func() {
					Expect(evaluate("||" + strings.Repeat(" false", 3))).To(Equal(FALSE))
					Expect(evaluate("||" + strings.Repeat(" false", 3) + " true")).To(Equal(TRUE))
				})
				It("should accept script expressions", func() {
					Expect(evaluate("|| {idem false}")).To(Equal(FALSE))
					Expect(evaluate("|| {idem true}")).To(Equal(TRUE))
				})
				It("should short-circuit on `true`", func() {
					Expect(evaluate("|| true {unreachable}")).To(Equal(TRUE))
				})
			})

			Describe("Exceptions", func() {
				Specify("wrong arity", func() {
					Expect(execute("||")).To(Equal(
						ERROR(`wrong # args: should be "|| arg ?arg ...?"`),
					))
				})
				Specify("invalid value", func() {
					Expect(execute("|| a")).To(Equal(ERROR(`invalid boolean "a"`)))
				})
			})

			Describe("Control flow", func() {
				Describe("`return`", func() {
					It("should interrupt expression with `RETURN` code", func() {
						Expect(execute("|| false {return value; unreachable} true")).To(Equal(
							RETURN(STR("value")),
						))
					})
				})
				Describe("`tailcall`", func() {
					It("should interrupt expression with `RETURN` code", func() {
						Expect(
							execute("|| false {tailcall {idem value}; unreachable} true"),
						).To(Equal(RETURN(STR("value"))))
					})
				})
				Describe("`yield`", func() {
					It("should interrupt expression with `YIELD` code", func() {
						result := execute("|| false {yield value; false}")
						Expect(result.Code).To(Equal(core.ResultCode_YIELD))
						Expect(result.Value).To(Equal(STR("value")))
					})
					It("should provide a resumable state", func() {
						process := prepareScript(
							"|| {yield val1} {yield val2}",
						)

						result := process.Run()
						Expect(result.Code).To(Equal(core.ResultCode_YIELD))
						Expect(result.Value).To(Equal(STR("val1")))

						process.YieldBack(FALSE)
						result = process.Run()
						Expect(result.Code).To(Equal(core.ResultCode_YIELD))
						Expect(result.Value).To(Equal(STR("val2")))

						process.YieldBack(TRUE)
						result = process.Run()
						Expect(result).To(Equal(OK(TRUE)))
					})
				})
				Describe("`error`", func() {
					It("should interrupt expression with `ERROR` code", func() {
						Expect(execute("|| false {error msg; true} true")).To(Equal(
							ERROR("msg"),
						))
					})
				})
				Describe("`break`", func() {
					It("should interrupt expression with `BREAK` code", func() {
						Expect(execute("|| false {break; unreachable} true")).To(Equal(
							BREAK(NIL),
						))
					})
				})
				Describe("`continue`", func() {
					It("should interrupt expression with `CONTINUE` code", func() {
						Expect(execute("|| false {continue; unreachable} true")).To(Equal(
							CONTINUE(NIL),
						))
					})
				})
			})
		})
	})
})
