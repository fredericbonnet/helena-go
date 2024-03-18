package helena_dialect_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"helena/core"
	. "helena/helena_dialect"
)

var _ = Describe("Helena closures", func() {
	var rootScope *Scope

	var tokenizer core.Tokenizer
	var parser *core.Parser

	parse := func(script string) *core.Script {
		return parser.Parse(tokenizer.Tokenize(script)).Script
	}
	execute := func(script string) core.Result {
		return rootScope.ExecuteScript(*parse(script))
	}
	evaluate := func(script string) core.Value {
		return execute(script).Value
	}
	init := func() {
		rootScope = NewScope(nil, false)
		InitCommands(rootScope)

		tokenizer = core.Tokenizer{}
		parser = &core.Parser{}
	}

	example := specifyExample(func(spec exampleSpec) core.Result { return execute(spec.script) })

	BeforeEach(init)

	Describe("closure", func() {

		Describe("Specifications", func() {
			Specify("usage", func() {
				Expect(evaluate("help closure")).To(Equal(
					STR("closure ?name? argspec body"),
				))
				Expect(evaluate("help closure args")).To(Equal(
					STR("closure ?name? argspec body"),
				))
				Expect(evaluate("help closure args {}")).To(Equal(
					STR("closure ?name? argspec body"),
				))
				Expect(evaluate("help closure cmd args {}")).To(Equal(
					STR("closure ?name? argspec body"),
				))
			})

			It("should define a new command", func() {
				evaluate("closure cmd {} {}")
				Expect(rootScope.Context.Commands["cmd"]).NotTo(BeNil())
			})
			It("should replace existing commands", func() {
				evaluate("closure cmd {} {}")
				Expect(execute("closure cmd {} {}").Code).To(Equal(core.ResultCode_OK))
			})
		})

		Describe("Exceptions", func() {
			Specify("wrong arity", func() {
				Expect(execute("closure")).To(Equal(
					ERROR(`wrong # args: should be "closure ?name? argspec body"`),
				))
				Expect(execute("closure a")).To(Equal(
					ERROR(`wrong # args: should be "closure ?name? argspec body"`),
				))
				Expect(execute("closure a b c d")).To(Equal(
					ERROR(`wrong # args: should be "closure ?name? argspec body"`),
				))
				Expect(execute("help closure a b c d")).To(Equal(
					ERROR(`wrong # args: should be "closure ?name? argspec body"`),
				))
			})
			Specify("invalid `argspec`", func() {
				Expect(execute("closure a {}")).To(Equal(ERROR("invalid argument list")))
				Expect(execute("closure cmd a {}")).To(Equal(
					ERROR("invalid argument list"),
				))
			})
			Specify("invalid `name`", func() {
				Expect(execute("closure [] {} {}")).To(Equal(
					ERROR("invalid command name"),
				))
			})
			Specify("non-script body", func() {
				Expect(execute("closure a b")).To(Equal(ERROR("body must be a script")))
				Expect(execute("closure a b c")).To(Equal(ERROR("body must be a script")))
			})
		})

		Describe("Metacommand", func() {
			It("should return a metacommand", func() {
				Expect(evaluate("closure {} {}").Type()).To(Equal(core.ValueType_COMMAND))
				Expect(evaluate("closure cmd {} {}").Type()).To(Equal(core.ValueType_COMMAND))
			})
			Specify("the metacommand should return the closure", func() {
				value := evaluate("set cmd [closure {val} {idem _${val}_}]")
				Expect(evaluate("$cmd").Type()).To(Equal(core.ValueType_COMMAND))
				Expect(evaluate("$cmd")).NotTo(Equal(value))
				Expect(evaluate("[$cmd] arg")).To(Equal(STR("_arg_")))
			})

			Describe("Examples", func() {
				Specify("Calling closure through its wrapped metacommand", func() {
					example([]exampleSpec{
						{
							script: `
								set cmd [closure double {val} {* 2 $val}]
								[$cmd] 3
							`,
							result: INT(6),
						},
						{
							script: `
								double 3
							`,
							result: INT(6),
						},
					})
				})
			})

			Describe("Subcommands", func() {
				Describe("`subcommands`", func() {
					It("should return list of subcommands", func() {
						// Expect(evaluate("[closure {} {}] subcommands")).To(Equal(
						// 	evaluate("list (subcommands argspec)"),
						// ))
					})

					Describe("Exceptions", func() {
						Specify("wrong arity", func() {
							Expect(execute("[closure {} {}] subcommands a")).To(Equal(
								ERROR(`wrong # args: should be "<closure> subcommands"`),
							))
						})
					})
				})

				Describe("`argspec`", func() {
					It("should return the closure's argspec", func() {
						example([]exampleSpec{
							{
								script: `
									[closure {a b} {}] argspec
								`,
								result: evaluate("argspec {a b}"),
							},
							{
								script: `
									argspec {a b}
								`,
							},
						})
					})

					Describe("Exceptions", func() {
						Specify("wrong arity", func() {
							Expect(execute("[closure {} {}] argspec a")).To(Equal(
								ERROR(`wrong # args: should be "<closure> argspec"`),
							))
						})
					})
				})

				Describe("Exceptions", func() {
					Specify("unknown subcommand", func() {
						Expect(execute("[closure {} {}] unknownSubcommand")).To(Equal(
							ERROR(`unknown subcommand "unknownSubcommand"`),
						))
					})
					Specify("invalid subcommand name", func() {
						Expect(execute("[closure {} {}] []")).To(Equal(
							ERROR("invalid subcommand name"),
						))
					})
				})
			})
		})
	})

	Describe("Closure commands", func() {
		Describe("Help", func() {
			Specify("zero", func() {
				evaluate("closure cmd {} {}")
				Expect(evaluate("help cmd")).To(Equal(STR("cmd")))
				Expect(execute("help cmd foo")).To(Equal(
					ERROR(`wrong # args: should be "cmd"`),
				))
			})
			Specify("one", func() {
				evaluate("closure cmd {a} {}")
				Expect(evaluate("help cmd")).To(Equal(STR("cmd a")))
				Expect(evaluate("help cmd foo")).To(Equal(STR("cmd a")))
				Expect(execute("help cmd foo bar")).To(Equal(
					ERROR(`wrong # args: should be "cmd a"`),
				))
			})
			Specify("two", func() {
				evaluate("closure cmd {a b} {}")
				Expect(evaluate("help cmd")).To(Equal(STR("cmd a b")))
				Expect(evaluate("help cmd foo")).To(Equal(STR("cmd a b")))
				Expect(evaluate("help cmd foo bar")).To(Equal(STR("cmd a b")))
				Expect(execute("help cmd foo bar baz")).To(Equal(
					ERROR(`wrong # args: should be "cmd a b"`),
				))
			})
			Specify("optional", func() {
				evaluate("closure cmd {?a} {}")
				Expect(evaluate("help cmd")).To(Equal(STR("cmd ?a?")))
				Expect(evaluate("help cmd foo")).To(Equal(STR("cmd ?a?")))
				Expect(execute("help cmd foo bar")).To(Equal(
					ERROR(`wrong # args: should be "cmd ?a?"`),
				))
			})
			Specify("remainder", func() {
				evaluate("closure cmd {a *} {}")
				Expect(evaluate("help cmd")).To(Equal(STR("cmd a ?arg ...?")))
				Expect(evaluate("help cmd foo")).To(Equal(STR("cmd a ?arg ...?")))
				Expect(evaluate("help cmd foo bar")).To(Equal(STR("cmd a ?arg ...?")))
				Expect(evaluate("help cmd foo bar baz")).To(Equal(STR("cmd a ?arg ...?")))
			})
			Specify("anonymous", func() {
				evaluate("set cmd [closure {a ?b} {}]")
				Expect(evaluate("help [$cmd]")).To(Equal(STR("<closure> a ?b?")))
				Expect(evaluate("help [$cmd] foo")).To(Equal(STR("<closure> a ?b?")))
				Expect(evaluate("help [$cmd] foo bar")).To(Equal(STR("<closure> a ?b?")))
				Expect(execute("help [$cmd] foo bar baz")).To(Equal(
					ERROR(`wrong # args: should be "<closure> a ?b?"`),
				))
			})
		})

		Describe("Arguments", func() {
			It("should shadow scope variables", func() {
				evaluate("set var val")
				evaluate("closure cmd {var} {idem $var}")
				Expect(evaluate("cmd val2")).To(Equal(STR("val2")))
			})
			It("should be closure-local", func() {
				evaluate("set var val")
				evaluate("closure cmd {var} {[[closure {} {idem $var}]]}")
				Expect(evaluate("cmd val2")).To(Equal(STR("val")))
			})

			Describe("Exceptions", func() {
				Specify("wrong arity", func() {
					evaluate("closure cmd {a} {}")
					Expect(execute("cmd")).To(Equal(
						ERROR(`wrong # args: should be "cmd a"`),
					))
					Expect(execute("cmd 1 2")).To(Equal(
						ERROR(`wrong # args: should be "cmd a"`),
					))
					Expect(execute("[[closure {a} {}]]")).To(Equal(
						ERROR(`wrong # args: should be "<closure> a"`),
					))
					Expect(execute("[[closure cmd {a} {}]]")).To(Equal(
						ERROR(`wrong # args: should be "<closure> a"`),
					))
				})
			})
		})

		Describe("Command calls", func() {
			It("should return nil for empty body", func() {
				evaluate("closure cmd {} {}")
				Expect(evaluate("cmd")).To(Equal(NIL))
			})
			It("should return the result of the last command", func() {
				evaluate("closure cmd {} {idem val1; idem val2}")
				Expect(execute("cmd")).To(Equal(OK(STR("val2"))))
			})
			Describe("should evaluate in the closure parent scope", func() {
				Specify("global scope", func() {
					evaluate(
						"closure cmd {} {let cst val1; set var val2; macro cmd2 {} {idem val3}}",
					)
					evaluate("cmd")
					Expect(rootScope.Context.Constants["cst"]).To(Equal(STR("val1")))
					Expect(rootScope.Context.Variables["var"]).To(Equal(STR("val2")))
					Expect(rootScope.Context.Commands["cmd2"]).NotTo(BeNil())
				})
				Specify("child scope", func() {
					evaluate(
						"closure cmd {} {let cst val1; set var val2; macro cmd2 {} {idem val3}}",
					)
					evaluate("scope scp {cmd}")
					Expect(rootScope.Context.Constants["cst"]).To(Equal(STR("val1")))
					Expect(rootScope.Context.Variables["var"]).To(Equal(STR("val2")))
					Expect(rootScope.Context.Commands["cmd2"]).NotTo(BeNil())
				})
				Specify("scoped closure", func() {
					evaluate(
						"scope scp1 {set cmd [closure {} {let cst val1; set var val2; macro cmd2 {} {idem val3}}]}",
					)
					evaluate("scope scp2 {[[scp1 eval {get cmd}]]}")
					Expect(evaluate("scp1 eval {get cst}")).To(Equal(STR("val1")))
					Expect(evaluate("scp1 eval {get var}")).To(Equal(STR("val2")))
					Expect(evaluate("scp1 eval {cmd2}")).To(Equal(STR("val3")))
					Expect(execute("scp2 eval {get cst}").Code).To(Equal(core.ResultCode_ERROR))
					Expect(execute("scp2 eval {get var}").Code).To(Equal(core.ResultCode_ERROR))
					Expect(execute("scp2 eval {cmd2}").Code).To(Equal(core.ResultCode_ERROR))
				})
			})
		})

		Describe("Return guards", func() {
			It("should apply to the return value", func() {
				evaluate(`macro guard {result} {idem "guarded:$result"}`)
				evaluate("closure cmd1 {var} {idem $var}")
				evaluate("closure cmd2 {var} (guard {idem $var})")
				Expect(execute("cmd1 value")).To(Equal(OK(STR("value"))))
				Expect(execute("cmd2 value")).To(Equal(OK(STR("guarded:value"))))
			})
			It("should let body errors pass through", func() {
				evaluate("macro guard {result} {unreachable}")
				evaluate("closure cmd {var} (guard {error msg})")
				Expect(execute("cmd value")).To(Equal(ERROR("msg")))
			})
			It("should not access closure arguments", func() {
				evaluate("macro guard {result} {exists var}")
				evaluate("closure cmd {var} (guard {idem $var})")
				Expect(evaluate("cmd value")).To(Equal(FALSE))
			})
			It("should evaluate in the closure parent scope", func() {
				// evaluate("macro guard {result} {idem root}")
				// evaluate("closure cmd {} (guard {true})")
				// evaluate("scope scp {macro guard {result} {idem scp}}")
				// Expect(evaluate("scp eval {cmd}")).To(Equal(STR("root")))
			})

			Describe("Exceptions", func() {
				Specify("empty body specifier", func() {
					Expect(execute("closure a ()")).To(Equal(ERROR("empty body specifier")))
					Expect(execute("closure a b ()")).To(Equal(
						ERROR("empty body specifier"),
					))
				})
				Specify("invalid body specifier", func() {
					Expect(execute("closure a (b c d)")).To(Equal(
						ERROR("invalid body specifier"),
					))
					Expect(execute("closure a b (c d e)")).To(Equal(
						ERROR("invalid body specifier"),
					))
				})
				Specify("non-script body", func() {
					Expect(execute("closure a (b c)")).To(Equal(
						ERROR("body must be a script"),
					))
					Expect(execute("closure a b (c d)")).To(Equal(
						ERROR("body must be a script"),
					))
				})
			})
		})

		Describe("Control flow", func() {
			Describe("`return`", func() {
				It("should interrupt a closure with `RETURN` code", func() {
					evaluate("closure cmd {} {return val1; idem val2}")
					Expect(execute("cmd")).To(Equal(RETURN(STR("val1"))))
				})
			})
			Describe("`tailcall`", func() {
				It("should interrupt a closure with `RETURN` code", func() {
					evaluate("closure cmd {} {tailcall {idem val1}; idem val2}")
					Expect(execute("cmd")).To(Equal(RETURN(STR("val1"))))
				})
			})
			Describe("`yield`", func() {
				It("should interrupt a closure with `YIELD` code", func() {
					evaluate("closure cmd {} {yield val1; idem val2}")
					result := execute("cmd")
					Expect(result.Code).To(Equal(core.ResultCode_YIELD))
					Expect(result.Value).To(Equal(STR("val1")))
				})
				It("should provide a resumable state", func() {
					evaluate("closure cmd {} {idem _[yield val1]_}")
					process := rootScope.PrepareScript(*parse("cmd"))

					result := process.Run()
					Expect(result.Code).To(Equal(core.ResultCode_YIELD))
					Expect(result.Value).To(Equal(STR("val1")))

					process.YieldBack(STR("val2"))
					result = process.Run()
					Expect(result).To(Equal(OK(STR("_val2_"))))
				})
				It("should work recursively", func() {
					evaluate("closure cmd1 {} {yield [cmd2]; idem val5}")
					evaluate("closure cmd2 {} {yield [cmd3]; idem [cmd4]}")
					evaluate("closure cmd3 {} {yield val1}")
					evaluate("closure cmd4 {} {yield val3}")
					process := rootScope.PrepareScript(*parse("cmd1"))

					result := process.Run()
					Expect(result.Code).To(Equal(core.ResultCode_YIELD))
					Expect(result.Value).To(Equal(STR("val1")))

					process.YieldBack(STR("val2"))
					result = process.Run()
					Expect(result.Code).To(Equal(core.ResultCode_YIELD))
					Expect(result.Value).To(Equal(STR("val2")))

					result = process.Run()
					Expect(result.Code).To(Equal(core.ResultCode_YIELD))
					Expect(result.Value).To(Equal(STR("val3")))

					process.YieldBack(STR("val4"))
					result = process.Run()
					Expect(result.Code).To(Equal(core.ResultCode_YIELD))
					Expect(result.Value).To(Equal(STR("val4")))

					result = process.Run()
					Expect(result).To(Equal(OK(STR("val5"))))
				})
			})
			Describe("`error`", func() {
				It("should interrupt a closure with `ERROR` code", func() {
					evaluate("closure cmd {} {error msg; idem val}")
					Expect(execute("cmd")).To(Equal(ERROR("msg")))
				})
			})
			Describe("`break`", func() {
				It("should interrupt a closure with `BREAK` code", func() {
					evaluate("closure cmd {} {break; idem val}")
					Expect(execute("cmd")).To(Equal(BREAK(NIL)))
				})
			})
			Describe("`continue`", func() {
				It("should interrupt a closure with `CONTINUE` code", func() {
					evaluate("closure cmd {} {continue; idem val}")
					Expect(execute("cmd")).To(Equal(CONTINUE(NIL)))
				})
			})
		})
	})
})
