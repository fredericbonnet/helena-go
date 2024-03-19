package helena_dialect_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"helena/core"
	. "helena/helena_dialect"
)

var _ = Describe("Helena procedures", func() {
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

	Describe("proc", func() {
		Describe("Specifications", func() {
			Specify("usage", func() {
				Expect(evaluate("help proc")).To(Equal(STR("proc ?name? argspec body")))
				Expect(evaluate("help proc args")).To(Equal(
					STR("proc ?name? argspec body"),
				))
				Expect(evaluate("help proc args {}")).To(Equal(
					STR("proc ?name? argspec body"),
				))
				Expect(evaluate("help proc cmd args {}")).To(Equal(
					STR("proc ?name? argspec body"),
				))
			})

			It("should define a new command", func() {
				evaluate("proc cmd {} {}")
				Expect(rootScope.Context.Commands["cmd"]).NotTo(BeNil())
			})
			It("should replace existing commands", func() {
				evaluate("proc cmd {} {}")
				Expect(execute("proc cmd {} {}").Code).To(Equal(core.ResultCode_OK))
			})
		})

		Describe("Exceptions", func() {
			Specify("wrong arity", func() {
				Expect(execute("proc")).To(Equal(
					ERROR(`wrong # args: should be "proc ?name? argspec body"`),
				))
				Expect(execute("proc a")).To(Equal(
					ERROR(`wrong # args: should be "proc ?name? argspec body"`),
				))
				Expect(execute("proc a b c d")).To(Equal(
					ERROR(`wrong # args: should be "proc ?name? argspec body"`),
				))
				Expect(execute("help proc a b c d")).To(Equal(
					ERROR(`wrong # args: should be "proc ?name? argspec body"`),
				))
			})
			Specify("invalid `argspec`", func() {
				Expect(execute("proc a {}")).To(Equal(ERROR("invalid argument list")))
			})
			Specify("invalid `name`", func() {
				Expect(execute("proc [] {} {}")).To(Equal(ERROR("invalid command name")))
			})
			Specify("non-script body", func() {
				Expect(execute("proc a b")).To(Equal(ERROR("body must be a script")))
				Expect(execute("proc a b c")).To(Equal(ERROR("body must be a script")))
			})
		})

		Describe("Metacommand", func() {
			It("should return a metacommand", func() {
				Expect(evaluate("proc {} {}").Type()).To(Equal(core.ValueType_COMMAND))
				Expect(evaluate("proc cmd {} {}").Type()).To(Equal(core.ValueType_COMMAND))
			})
			Specify("the metacommand should return the procedure", func() {
				value := evaluate("set cmd [proc {val} {idem _${val}_}]")
				Expect(evaluate("$cmd").Type()).To(Equal(core.ValueType_COMMAND))
				Expect(evaluate("$cmd")).NotTo(Equal(value))
				Expect(evaluate("[$cmd] arg")).To(Equal(STR("_arg_")))
			})

			Describe("Examples", func() {
				Specify("Calling procedure through its wrapped metacommand", func() {
					example([]exampleSpec{
						{
							script: `
								set cmd [proc double {val} {* 2 $val}]
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
						Expect(evaluate("[proc {} {}] subcommands")).To(Equal(
							evaluate("list (subcommands argspec)"),
						))
					})

					Describe("Exceptions", func() {
						Specify("wrong arity", func() {
							Expect(execute("[proc {} {}] subcommands a")).To(Equal(
								ERROR(`wrong # args: should be "<proc> subcommands"`),
							))
						})
					})
				})

				Describe("`argspec`", func() {
					It("should return the procedure's argspec", func() {
						example([]exampleSpec{
							{
								script: `
									[proc {a b} {}] argspec
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
							Expect(execute("[proc {} {}] argspec a")).To(Equal(
								ERROR(`wrong # args: should be "<proc> argspec"`),
							))
						})
					})
				})

				Describe("Exceptions", func() {
					Specify("unknown subcommand", func() {
						Expect(execute("[proc {} {}] unknownSubcommand")).To(Equal(
							ERROR(`unknown subcommand "unknownSubcommand"`),
						))
					})
					Specify("invalid subcommand name", func() {
						Expect(execute("[proc {} {}] []")).To(Equal(
							ERROR("invalid subcommand name"),
						))
					})
				})
			})
		})
	})

	Describe("Procedure commands", func() {
		Describe("Help", func() {
			Specify("zero", func() {
				evaluate("proc cmd {} {}")
				Expect(evaluate("help cmd")).To(Equal(STR("cmd")))
				Expect(execute("help cmd foo")).To(Equal(
					ERROR(`wrong # args: should be "cmd"`),
				))
			})
			Specify("one", func() {
				evaluate("proc cmd {a} {}")
				Expect(evaluate("help cmd")).To(Equal(STR("cmd a")))
				Expect(evaluate("help cmd foo")).To(Equal(STR("cmd a")))
				Expect(execute("help cmd foo bar")).To(Equal(
					ERROR(`wrong # args: should be "cmd a"`),
				))
			})
			Specify("two", func() {
				evaluate("proc cmd {a b} {}")
				Expect(evaluate("help cmd")).To(Equal(STR("cmd a b")))
				Expect(evaluate("help cmd foo")).To(Equal(STR("cmd a b")))
				Expect(evaluate("help cmd foo bar")).To(Equal(STR("cmd a b")))
				Expect(execute("help cmd foo bar baz")).To(Equal(
					ERROR(`wrong # args: should be "cmd a b"`),
				))
			})
			Specify("optional", func() {
				evaluate("proc cmd {?a} {}")
				Expect(evaluate("help cmd")).To(Equal(STR("cmd ?a?")))
				Expect(evaluate("help cmd foo")).To(Equal(STR("cmd ?a?")))
				Expect(execute("help cmd foo bar")).To(Equal(
					ERROR(`wrong # args: should be "cmd ?a?"`),
				))
			})
			Specify("remainder", func() {
				evaluate("proc cmd {a *} {}")
				Expect(evaluate("help cmd")).To(Equal(STR("cmd a ?arg ...?")))
				Expect(evaluate("help cmd foo")).To(Equal(STR("cmd a ?arg ...?")))
				Expect(evaluate("help cmd foo bar")).To(Equal(STR("cmd a ?arg ...?")))
				Expect(evaluate("help cmd foo bar baz")).To(Equal(STR("cmd a ?arg ...?")))
			})
			Specify("anonymous", func() {
				evaluate("set cmd [proc {a ?b} {}]")
				Expect(evaluate("help [$cmd]")).To(Equal(STR("<proc> a ?b?")))
				Expect(evaluate("help [$cmd] foo")).To(Equal(STR("<proc> a ?b?")))
				Expect(evaluate("help [$cmd] foo bar")).To(Equal(STR("<proc> a ?b?")))
				Expect(execute("help [$cmd] foo bar baz")).To(Equal(
					ERROR(`wrong # args: should be "<proc> a ?b?"`),
				))
			})
		})

		Describe("Arguments", func() {
			It("should be scope variables", func() {
				evaluate("set var val")
				evaluate("proc cmd {var} {macro cmd2 {} {set var _$var}; cmd2}")
				Expect(evaluate("cmd val2")).To(Equal(STR("_val2")))
			})

			Describe("Exceptions", func() {
				Specify("wrong arity", func() {
					evaluate("proc cmd {a} {}")
					Expect(execute("cmd")).To(Equal(
						ERROR(`wrong # args: should be "cmd a"`),
					))
					Expect(execute("cmd 1 2")).To(Equal(
						ERROR(`wrong # args: should be "cmd a"`),
					))
					Expect(execute("[[proc {a} {}]]")).To(Equal(
						ERROR(`wrong # args: should be "<proc> a"`),
					))
					Expect(execute("[[proc cmd {a} {}]]")).To(Equal(
						ERROR(`wrong # args: should be "<proc> a"`),
					))
				})
			})
		})

		Describe("Command calls", func() {
			It("should return nil for empty body", func() {
				evaluate("proc cmd {} {}")
				Expect(evaluate("cmd")).To(Equal(NIL))
			})
			It("should return the result of the last command", func() {
				evaluate("proc cmd {} {idem val1; idem val2}")
				Expect(execute("cmd")).To(Equal(OK(STR("val2"))))
			})
			It("should evaluate in their own scope", func() {
				evaluate(
					"proc cmd {} {let cst val1; set var val2; macro cmd2 {} {idem val3}; set var [cmd2]}",
				)
				Expect(execute("cmd")).To(Equal(OK(STR("val3"))))
				Expect(rootScope.Context.Constants["cst"]).To(BeNil())
				Expect(rootScope.Context.Variables["var"]).To(BeNil())
				Expect(rootScope.Context.Commands["cmd2"]).To(BeNil())
			})
			It("should evaluate from their parent scope", func() {
				evaluate("closure cls {} {set var val}")
				evaluate("proc cmd {} {cls}")
				Expect(
					evaluate("[scope {closure cls {} {set var val2}}] eval {cmd}"),
				).To(Equal(STR("val")))
				Expect(evaluate("get var")).To(Equal(STR("val")))
			})
			It("should access external commands", func() {
				evaluate("proc cmd {} {idem val}")
				Expect(evaluate("cmd")).To(Equal(STR("val")))
			})
			It("should not access external variables", func() {
				evaluate("set var val")
				evaluate("proc cmd {} {get var}")
				Expect(execute("cmd").Code).To(Equal(core.ResultCode_ERROR))
			})
			It("should not set external variables", func() {
				evaluate("set var val")
				evaluate("proc cmd {} {set var val2; let cst val3}")
				evaluate("cmd")
				Expect(rootScope.Context.Variables["var"]).To(Equal(STR("val")))
				Expect(rootScope.Context.Constants["cst"]).To(BeNil())
			})
			Specify("local commands should shadow external commands", func() {
				evaluate("macro mac {} {idem val}")
				evaluate("proc cmd {} {macro mac {} {idem val2}; mac}")
				Expect(evaluate("cmd")).To(Equal(STR("val2")))
			})
		})

		Describe("Return guards", func() {
			It("should apply to the return value", func() {
				evaluate(`macro guard {result} {idem "guarded:$result"}`)
				evaluate("proc cmd1 {var} {return $var}")
				evaluate("proc cmd2 {var} (guard {return $var})")
				Expect(execute("cmd1 value")).To(Equal(OK(STR("value"))))
				Expect(execute("cmd2 value")).To(Equal(OK(STR("guarded:value"))))
			})
			It("should let body errors pass through", func() {
				evaluate("macro guard {result} {unreachable}")
				evaluate("proc cmd {var} (guard {error msg})")
				Expect(execute("cmd value")).To(Equal(ERROR("msg")))
			})
			It("should not access proc arguments", func() {
				evaluate("macro guard {result} {exists var}")
				evaluate("proc cmd {var} (guard {return $var})")
				Expect(evaluate("cmd value")).To(Equal(FALSE))
			})
			It("should evaluate in the proc parent scope", func() {
				// evaluate("macro guard {result} {idem root}")
				// evaluate("proc cmd {} (guard {true})")
				// evaluate("scope scp {macro guard {result} {idem scp}}")
				// Expect(evaluate("scp eval {cmd}")).To(Equal(STR("root")))
			})
			Describe("Exceptions", func() {
				Specify("empty body specifier", func() {
					Expect(execute("proc a ()")).To(Equal(ERROR("empty body specifier")))
					Expect(execute("proc a b ()")).To(Equal(ERROR("empty body specifier")))
				})
				Specify("invalid body specifier", func() {
					Expect(execute("proc a (b c d)")).To(Equal(
						ERROR("invalid body specifier"),
					))
					Expect(execute("proc a b (c d e)")).To(Equal(
						ERROR("invalid body specifier"),
					))
				})
				Specify("non-script body", func() {
					Expect(execute("proc a (b c)")).To(Equal(
						ERROR("body must be a script"),
					))
					Expect(execute("proc a b (c d)")).To(Equal(
						ERROR("body must be a script"),
					))
				})
			})
		})

		Describe("Control flow", func() {
			Describe("`return`", func() {
				It("should interrupt a proc with `OK` code", func() {
					evaluate("proc cmd {} {return val1; idem val2}")
					Expect(execute("cmd")).To(Equal(OK(STR("val1"))))
				})
			})
			Describe("`tailcall`", func() {
				It("should interrupt a proc with `OK` code", func() {
					evaluate("proc cmd {} {tailcall (idem val1); idem val2}")
					Expect(execute("cmd")).To(Equal(OK(STR("val1"))))
				})
			})
			Describe("`yield`", func() {
				It("should interrupt a proc with `YIELD` code", func() {
					evaluate("proc cmd {} {yield val1; idem val2}")
					result := execute("cmd")
					Expect(result.Code).To(Equal(core.ResultCode_YIELD))
					Expect(result.Value).To(Equal(STR("val1")))
				})
				It("should provide a resumable state", func() {
					evaluate("proc cmd {} {idem _[yield val1]_}")
					process := rootScope.PrepareScript(*parse("cmd"))

					result := process.Run()
					Expect(result.Code).To(Equal(core.ResultCode_YIELD))
					Expect(result.Value).To(Equal(STR("val1")))
					Expect(result.Data).NotTo(BeNil())

					process.YieldBack(STR("val2"))
					result = process.Run()
					Expect(result).To(Equal(OK(STR("_val2_"))))
				})
				It("should work recursively", func() {
					evaluate("proc cmd1 {} {yield [cmd2]; idem val5}")
					evaluate("proc cmd2 {} {yield [cmd3]; idem [cmd4]}")
					evaluate("proc cmd3 {} {yield val1}")
					evaluate("proc cmd4 {} {yield val3}")
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
				It("should interrupt a proc with `ERROR` code", func() {
					evaluate("proc cmd {} {error msg; idem val}")
					Expect(execute("cmd")).To(Equal(ERROR("msg")))
				})
			})
			Describe("`break`", func() {
				It("should interrupt a proc with `ERROR` code", func() {
					evaluate("proc cmd {} {break; idem val}")
					Expect(execute("cmd")).To(Equal(ERROR("unexpected break")))
				})
			})
			Describe("`continue`", func() {
				It("should interrupt a proc with `ERROR` code", func() {
					evaluate("proc cmd {} {continue; idem val}")
					Expect(execute("cmd")).To(Equal(ERROR("unexpected continue")))
				})
			})
		})
	})
})
