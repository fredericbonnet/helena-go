package helena_dialect_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"helena/core"
	. "helena/helena_dialect"
)

var _ = Describe("Helena macros", func() {
	var rootScope *Scope

	var tokenizer core.Tokenizer
	var parser *core.Parser

	parse := func(script string) *core.Script {
		return parser.Parse(tokenizer.Tokenize(script)).Script
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
		rootScope = NewRootScope()
		InitCommands(rootScope)

		tokenizer = core.Tokenizer{}
		parser = core.NewParser(nil)
	}

	example := specifyExample(func(spec exampleSpec) core.Result { return execute(spec.script) })

	BeforeEach(init)

	Describe("macro", func() {
		Describe("Specifications", func() {
			Specify("usage", func() {
				Expect(evaluate("help macro")).To(Equal(STR("macro ?name? argspec body")))
				Expect(evaluate("help macro args")).To(Equal(
					STR("macro ?name? argspec body"),
				))
				Expect(evaluate("help macro args {}")).To(Equal(
					STR("macro ?name? argspec body"),
				))
				Expect(evaluate("help macro cmd args {}")).To(Equal(
					STR("macro ?name? argspec body"),
				))
			})

			It("should define a new command", func() {
				evaluate("macro cmd {} {}")
				Expect(rootScope.Context.Commands["cmd"]).NotTo(BeNil())
			})
			It("should replace existing commands", func() {
				evaluate("macro cmd {} {}")
				Expect(execute("macro cmd {} {}").Code).To(Equal(core.ResultCode_OK))
			})
		})

		Describe("Exceptions", func() {
			Specify("wrong arity", func() {
				Expect(execute("macro")).To(Equal(
					ERROR(`wrong # args: should be "macro ?name? argspec body"`),
				))
				Expect(execute("macro a")).To(Equal(
					ERROR(`wrong # args: should be "macro ?name? argspec body"`),
				))
				Expect(execute("macro a b c d")).To(Equal(
					ERROR(`wrong # args: should be "macro ?name? argspec body"`),
				))
				Expect(execute("help macro a b c d")).To(Equal(
					ERROR(`wrong # args: should be "macro ?name? argspec body"`),
				))
			})
			Specify("invalid `argspec`", func() {
				Expect(execute("macro a {}")).To(Equal(ERROR("invalid argument list")))
				Expect(execute("macro cmd a {}")).To(Equal(
					ERROR("invalid argument list"),
				))
			})
			Specify("invalid `name`", func() {
				Expect(execute("macro [] {} {}")).To(Equal(ERROR("invalid command name")))
			})
			Specify("non-script body", func() {
				Expect(execute("macro a b")).To(Equal(ERROR("body must be a script")))
				Expect(execute("macro a b c")).To(Equal(ERROR("body must be a script")))
			})
		})

		Describe("Metacommand", func() {
			It("should return a metacommand", func() {
				Expect(evaluate("macro {} {}").Type()).To(Equal(core.ValueType_COMMAND))
				Expect(evaluate("macro cmd {} {}").Type()).To(Equal(core.ValueType_COMMAND))
			})
			Specify("the metacommand should return the macro", func() {
				value := evaluate("set cmd [macro {val} {idem _${val}_}]")
				Expect(evaluate("$cmd").Type()).To(Equal(core.ValueType_COMMAND))
				Expect(evaluate("$cmd")).NotTo(Equal(value))
				Expect(evaluate("[$cmd] arg")).To(Equal(STR("_arg_")))
			})

			Describe("Examples", func() {
				Specify("Calling macro through its wrapped metacommand", func() {
					example([]exampleSpec{
						{
							script: `
								set cmd [macro double {val} {* 2 $val}]
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
						Expect(evaluate("[macro {} {}] subcommands")).To(Equal(
							evaluate("list (subcommands argspec)"),
						))
					})

					Describe("Exceptions", func() {
						Specify("wrong arity", func() {
							Expect(execute("[macro {} {}] subcommands a")).To(Equal(
								ERROR(`wrong # args: should be "<macro> subcommands"`),
							))
						})
					})
				})

				Describe("`argspec`", func() {
					It("should return the macro's argspec", func() {
						example([]exampleSpec{
							{
								script: `
									[macro {a b} {}] argspec
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
							Expect(execute("[macro {} {}] argspec a")).To(Equal(
								ERROR(`wrong # args: should be "<macro> argspec"`),
							))
						})
					})
				})

				Describe("Exceptions", func() {
					Specify("unknown subcommand", func() {
						Expect(execute("[macro {} {}] unknownSubcommand")).To(Equal(
							ERROR(`unknown subcommand "unknownSubcommand"`),
						))
					})
					Specify("invalid subcommand name", func() {
						Expect(execute("[macro {} {}] []")).To(Equal(
							ERROR("invalid subcommand name"),
						))
					})
				})
			})
		})
	})

	Describe("Macro commands", func() {
		Describe("Help", func() {
			Specify("zero", func() {
				evaluate("macro cmd {} {}")
				Expect(evaluate("help cmd")).To(Equal(STR("cmd")))
				Expect(execute("help cmd foo")).To(Equal(
					ERROR(`wrong # args: should be "cmd"`),
				))
			})
			Specify("one", func() {
				evaluate("macro cmd {a} {}")
				Expect(evaluate("help cmd")).To(Equal(STR("cmd a")))
				Expect(evaluate("help cmd foo")).To(Equal(STR("cmd a")))
				Expect(execute("help cmd foo bar")).To(Equal(
					ERROR(`wrong # args: should be "cmd a"`),
				))
			})
			Specify("two", func() {
				evaluate("macro cmd {a b} {}")
				Expect(evaluate("help cmd")).To(Equal(STR("cmd a b")))
				Expect(evaluate("help cmd foo")).To(Equal(STR("cmd a b")))
				Expect(evaluate("help cmd foo bar")).To(Equal(STR("cmd a b")))
				Expect(execute("help cmd foo bar baz")).To(Equal(
					ERROR(`wrong # args: should be "cmd a b"`),
				))
			})
			Specify("optional", func() {
				evaluate("macro cmd {?a} {}")
				Expect(evaluate("help cmd")).To(Equal(STR("cmd ?a?")))
				Expect(evaluate("help cmd foo")).To(Equal(STR("cmd ?a?")))
				Expect(execute("help cmd foo bar")).To(Equal(
					ERROR(`wrong # args: should be "cmd ?a?"`),
				))
			})
			Specify("remainder", func() {
				evaluate("macro cmd {a *} {}")
				Expect(evaluate("help cmd")).To(Equal(STR("cmd a ?arg ...?")))
				Expect(evaluate("help cmd foo")).To(Equal(STR("cmd a ?arg ...?")))
				Expect(evaluate("help cmd foo bar")).To(Equal(STR("cmd a ?arg ...?")))
				Expect(evaluate("help cmd foo bar baz")).To(Equal(STR("cmd a ?arg ...?")))
			})
			Specify("anonymous", func() {
				evaluate("set cmd [macro {a ?b} {}]")
				Expect(evaluate("help [$cmd]")).To(Equal(STR("<macro> a ?b?")))
				Expect(evaluate("help [$cmd] foo")).To(Equal(STR("<macro> a ?b?")))
				Expect(evaluate("help [$cmd] foo bar")).To(Equal(STR("<macro> a ?b?")))
				Expect(execute("help [$cmd] foo bar baz")).To(Equal(
					ERROR(`wrong # args: should be "<macro> a ?b?"`),
				))
			})
		})

		Describe("Arguments", func() {
			It("should shadow scope variables", func() {
				evaluate("set var val")
				evaluate("macro cmd {var} {idem $var}")
				Expect(evaluate("cmd val2")).To(Equal(STR("val2")))
			})
			It("should be macro-local", func() {
				evaluate("set var val")
				evaluate("macro cmd {var} {[[macro {} {idem $var}]]}")
				Expect(evaluate("cmd val2")).To(Equal(STR("val")))
			})

			Describe("Exceptions", func() {
				Specify("wrong arity", func() {
					evaluate("macro cmd {a} {}")
					Expect(execute("cmd")).To(Equal(
						ERROR(`wrong # args: should be "cmd a"`),
					))
					Expect(execute("cmd 1 2")).To(Equal(
						ERROR(`wrong # args: should be "cmd a"`),
					))
					Expect(execute("[[macro {a} {}]]")).To(Equal(
						ERROR(`wrong # args: should be "<macro> a"`),
					))
					Expect(execute("[[macro cmd {a} {}]]")).To(Equal(
						ERROR(`wrong # args: should be "<macro> a"`),
					))
				})
			})
		})

		Describe("Command calls", func() {
			It("should return nil for empty body", func() {
				evaluate("macro cmd {} {}")
				Expect(evaluate("cmd")).To(Equal(NIL))
			})
			It("should return the result of the last command", func() {
				evaluate("macro cmd {} {idem val1; idem val2}")
				Expect(execute("cmd")).To(Equal(OK(STR("val2"))))
			})
			Describe("should evaluate in the caller scope", func() {
				Specify("global scope", func() {
					evaluate(
						"macro cmd {} {let cst val1; set var val2; macro cmd2 {} {idem val3}}",
					)
					evaluate("cmd")
					Expect(rootScope.Context.Constants["cst"]).To(Equal(STR("val1")))
					Expect(rootScope.Context.Variables["var"]).To(Equal(STR("val2")))
					Expect(rootScope.Context.Commands["cmd2"]).NotTo(BeNil())
				})
				Specify("child scope", func() {
					evaluate(
						"macro cmd {} {let cst val1; set var val2; macro cmd2 {} {idem val3}}",
					)
					evaluate("scope scp {cmd}")
					Expect(rootScope.Context.Constants["cst"]).To(BeNil())
					Expect(rootScope.Context.Variables["var"]).To(BeNil())
					Expect(rootScope.Context.Commands["cmd2"]).To(BeNil())
					Expect(evaluate("scp eval {get cst}")).To(Equal(STR("val1")))
					Expect(evaluate("scp eval {get var}")).To(Equal(STR("val2")))
					Expect(evaluate("scp eval {cmd2}")).To(Equal(STR("val3")))
				})
				Specify("scoped macro", func() {
					evaluate(
						"scope scp1 {set cmd [macro {} {let cst val1; set var val2; macro cmd2 {} {idem val3}}]}",
					)
					evaluate("scope scp2 {[[scp1 eval {get cmd}]]}")
					Expect(execute("scp1 eval {get cst}").Code).To(Equal(core.ResultCode_ERROR))
					Expect(execute("scp1 eval {get var}").Code).To(Equal(core.ResultCode_ERROR))
					Expect(execute("scp1 eval {cmd2}").Code).To(Equal(core.ResultCode_ERROR))
					Expect(evaluate("scp2 eval {get cst}")).To(Equal(STR("val1")))
					Expect(evaluate("scp2 eval {get var}")).To(Equal(STR("val2")))
					Expect(evaluate("scp2 eval {cmd2}")).To(Equal(STR("val3")))
				})
			})
			It("should access scope variables", func() {
				evaluate("set var val")
				evaluate("macro cmd {} {get var}")
				Expect(evaluate("cmd")).To(Equal(STR("val")))
			})
			It("should set scope variables", func() {
				evaluate("set var old")
				evaluate("macro cmd {} {set var val; set var2 val2}")
				evaluate("cmd")
				Expect(evaluate("get var")).To(Equal(STR("val")))
				Expect(evaluate("get var2")).To(Equal(STR("val2")))
			})
			It("should access scope commands", func() {
				evaluate("macro cmd2 {} {set var val}")
				evaluate("macro cmd {} {cmd2}")
				evaluate("cmd")
				Expect(evaluate("get var")).To(Equal(STR("val")))
			})
		})

		Describe("Return guards", func() {
			It("should apply to the return value", func() {
				evaluate(`macro guard {result} {idem "guarded:$result"}`)
				evaluate("macro cmd1 {var} {idem $var}")
				evaluate("macro cmd2 {var} (guard {idem $var})")
				Expect(execute("cmd1 value")).To(Equal(OK(STR("value"))))
				Expect(execute("cmd2 value")).To(Equal(OK(STR("guarded:value"))))
			})
			It("should let body errors pass through", func() {
				evaluate("macro guard {result} {unreachable}")
				evaluate("macro cmd {var} (guard {error msg})")
				Expect(execute("cmd value")).To(Equal(ERROR("msg")))
			})
			It("should not access macro arguments", func() {
				evaluate("macro guard {result} {exists var}")
				evaluate("macro cmd {var} (guard {idem $var})")
				Expect(evaluate("cmd value")).To(Equal(FALSE))
			})
			It("should evaluate in the caller scope", func() {
				evaluate("macro guard {result} {idem root}")
				evaluate("macro cmd {} (guard {true})")
				evaluate("scope scp {macro guard {result} {idem scp}}")
				Expect(evaluate("scp eval {cmd}")).To(Equal(STR("scp")))
			})

			Describe("Exceptions", func() {
				Specify("empty body specifier", func() {
					Expect(execute("macro a ()")).To(Equal(ERROR("empty body specifier")))
					Expect(execute("macro a b ()")).To(Equal(ERROR("empty body specifier")))
				})
				Specify("invalid body specifier", func() {
					Expect(execute("macro a (b c d)")).To(Equal(
						ERROR("invalid body specifier"),
					))
					Expect(execute("macro a b (c d e)")).To(Equal(
						ERROR("invalid body specifier"),
					))
				})
				Specify("non-script body", func() {
					Expect(execute("macro a (b c)")).To(Equal(
						ERROR("body must be a script"),
					))
					Expect(execute("macro a b (c d)")).To(Equal(
						ERROR("body must be a script"),
					))
				})
			})
		})

		Describe("Control flow", func() {
			Describe("`return`", func() {
				It("should interrupt a macro with `RETURN` code", func() {
					evaluate("macro cmd {} {return val1; idem val2}")
					Expect(execute("cmd")).To(Equal(RETURN(STR("val1"))))
				})
			})
			Describe("`tailcall`", func() {
				It("should interrupt a macro with `RETURN` code", func() {
					evaluate("macro cmd {} {tailcall {idem val1}; idem val2}")
					Expect(execute("cmd")).To(Equal(RETURN(STR("val1"))))
				})
			})
			Describe("`yield`", func() {
				It("should interrupt a macro with `YIELD` code", func() {
					evaluate("macro cmd {} {yield val1; idem val2}")
					result := execute("cmd")
					Expect(result.Code).To(Equal(core.ResultCode_YIELD))
					Expect(result.Value).To(Equal(STR("val1")))
				})
				It("should provide a resumable state", func() {
					evaluate("macro cmd {} {idem _[yield val1]_}")
					process := prepareScript("cmd")

					result := process.Run()
					Expect(result.Code).To(Equal(core.ResultCode_YIELD))
					Expect(result.Value).To(Equal(STR("val1")))

					process.YieldBack(STR("val2"))
					result = process.Run()
					Expect(result).To(Equal(OK(STR("_val2_"))))
				})
				It("should work recursively", func() {
					evaluate("macro cmd1 {} {yield [cmd2]; idem val5}")
					evaluate("macro cmd2 {} {yield [cmd3]; idem [cmd4]}")
					evaluate("macro cmd3 {} {yield val1}")
					evaluate("macro cmd4 {} {yield val3}")
					process := prepareScript("cmd1")

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
				It("should interrupt a macro with `ERROR` code", func() {
					evaluate("macro cmd {} {error msg; idem val}")
					Expect(execute("cmd")).To(Equal(ERROR("msg")))
				})
			})
			Describe("`break`", func() {
				It("should interrupt a macro with `BREAK` code", func() {
					evaluate("macro cmd {} {break; unreachable}")
					Expect(execute("cmd")).To(Equal(BREAK(NIL)))
				})
			})
			Describe("`continue`", func() {
				It("should interrupt a macro with `CONTINUE` code", func() {
					evaluate("macro cmd {} {continue; unreachable}")
					Expect(execute("cmd")).To(Equal(CONTINUE(NIL)))
				})
			})
		})
	})
})
