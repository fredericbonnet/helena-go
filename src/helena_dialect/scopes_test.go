package helena_dialect_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"helena/core"
	. "helena/helena_dialect"
)

var _ = Describe("Helena scopes", func() {
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

	BeforeEach(init)

	Describe("scope", func() {
		Describe("Specifications", func() {
			Specify("usage", func() {
				Expect(evaluate("help scope")).To(Equal(STR("scope ?name? body")))
				Expect(evaluate("help scope {}")).To(Equal(STR("scope ?name? body")))
				Expect(evaluate("help scope cmd {}")).To(Equal(STR("scope ?name? body")))
			})

			It("should define a new command", func() {
				evaluate("scope cmd {}")
				Expect(rootScope.Context.Commands["cmd"]).NotTo(BeNil())
			})
			It("should replace existing commands", func() {
				evaluate("scope cmd {}")
				Expect(execute("scope cmd {}").Code).To(Equal(core.ResultCode_OK))
			})
			It("should return a command object", func() {
				Expect(evaluate("scope {}").Type()).To(Equal(core.ValueType_COMMAND))
				Expect(evaluate("scope cmd  {}").Type()).To(Equal(core.ValueType_COMMAND))
			})
			Specify("the named command should return its command object", func() {
				value := evaluate("scope cmd {}")
				Expect(evaluate("cmd")).To(Equal(value))
			})
			Specify("the command object should return itself", func() {
				value := evaluate("set cmd [scope {}]")
				Expect(evaluate("$cmd")).To(Equal(value))
			})
		})

		Describe("Exceptions", func() {
			Specify("wrong arity", func() {
				Expect(execute("scope")).To(Equal(
					ERROR(`wrong # args: should be "scope ?name? body"`),
				))
				Expect(execute("scope a b c")).To(Equal(
					ERROR(`wrong # args: should be "scope ?name? body"`),
				))
				Expect(execute("help scope a b c")).To(Equal(
					ERROR(`wrong # args: should be "scope ?name? body"`),
				))
			})
			Specify("invalid `name`", func() {
				Expect(execute("scope [] {}")).To(Equal(ERROR("invalid command name")))
			})
			Specify("non-script body", func() {
				Expect(execute("scope a")).To(Equal(ERROR("body must be a script")))
				Expect(execute("scope a b")).To(Equal(ERROR("body must be a script")))
			})
		})

		Describe("`body`", func() {
			It("should be executed", func() {
				evaluate("closure cmd {} {let var val}")
				Expect(rootScope.Context.Constants["var"]).To(BeNil())
				evaluate("scope {cmd}")
				Expect(rootScope.Context.Constants["var"]).To(Equal(STR("val")))
			})
			It("should access global commands", func() {
				Expect(execute("scope {idem val}").Code).To(Equal(core.ResultCode_OK))
			})
			It("should not access global variables", func() {
				evaluate("set var val")
				Expect(execute("scope {get var}").Code).To(Equal(core.ResultCode_ERROR))
			})
			It("should not set global variables", func() {
				evaluate("set var val")
				evaluate("scope {set var val2; let cst val3}")
				Expect(rootScope.Context.Variables["var"]).To(Equal(STR("val")))
				Expect(rootScope.Context.Constants["cst"]).To(BeNil())
			})
			It("should set scope variables", func() {
				evaluate("set var val")
				evaluate("scope cmd {set var val2; let cst val3}")
				Expect(rootScope.Context.Variables["var"]).To(Equal(STR("val")))
				Expect(rootScope.Context.Constants["cst"]).To(BeNil())
				Expect(evaluate("cmd eval {get var}")).To(Equal(STR("val2")))
				Expect(evaluate("cmd eval {get cst}")).To(Equal(STR("val3")))
			})

			Describe("Control flow", func() {
				Describe("`return`", func() {
					It("should interrupt the body with `OK` code", func() {
						evaluate("closure cmd1 {} {set var val1}")
						evaluate("closure cmd2 {} {set var val2}")
						Expect(execute("scope {cmd1; return; cmd2}").Code).To(Equal(
							core.ResultCode_OK,
						))
						Expect(evaluate("get var")).To(Equal(STR("val1")))
					})
					It("should still define the named command", func() {
						evaluate("scope cmd {return}")
						Expect(rootScope.Context.Commands["cmd"]).NotTo(BeNil())
					})
					It("should return passed value instead of the command object", func() {
						Expect(execute("scope {return val}")).To(Equal(OK(STR("val"))))
					})
				})
				Describe("`tailcall`", func() {
					It("should interrupt the body with `OK` code", func() {
						evaluate("closure cmd1 {} {set var val1}")
						evaluate("closure cmd2 {} {set var val2}")
						Expect(execute("scope {cmd1; tailcall {}; cmd2}").Code).To(Equal(
							core.ResultCode_OK,
						))
						Expect(evaluate("get var")).To(Equal(STR("val1")))
					})
					It("should still define the named command", func() {
						evaluate("scope cmd {tailcall {}}")
						Expect(rootScope.Context.Commands["cmd"]).NotTo(BeNil())
					})
					It("should return passed value instead of the command object", func() {
						Expect(execute("scope {tailcall {idem val}}")).To(Equal(
							OK(STR("val")),
						))
					})
				})
				Describe("`yield`", func() {
					It("should interrupt the body with `YIELD` code", func() {
						evaluate("closure cmd1 {} {set var val1}")
						evaluate("closure cmd2 {} {set var val2}")
						Expect(execute("scope cmd {cmd1; yield; cmd2}").Code).To(Equal(
							core.ResultCode_YIELD,
						))
						Expect(evaluate("get var")).To(Equal(STR("val1")))
					})
					It("should provide a resumable state", func() {
						evaluate("closure cmd1 {} {set var val1}")
						evaluate("closure cmd2 {val} {set var $val}")
						process := prepareScript(
							"scope cmd {cmd1; cmd2 _[yield val2]_}",
						)

						result := process.Run()
						Expect(result.Code).To(Equal(core.ResultCode_YIELD))
						Expect(result.Value).To(Equal(STR("val2")))

						process.YieldBack(STR("val3"))
						result = process.Run()
						Expect(result.Code).To(Equal(core.ResultCode_OK))
						Expect(result.Value.Type()).To(Equal(core.ValueType_COMMAND))
						Expect(evaluate("get var")).To(Equal(STR("_val3_")))
					})
					It("should delay the definition of scope command until resumed", func() {
						process := prepareScript("scope cmd {yield}")

						result := process.Run()
						Expect(result.Code).To(Equal(core.ResultCode_YIELD))
						Expect(rootScope.Context.Commands["cmd"]).To(BeNil())

						result = process.Run()
						Expect(result.Code).To(Equal(core.ResultCode_OK))
						Expect(rootScope.Context.Commands["cmd"]).NotTo(BeNil())
					})
				})
				Describe("`error`", func() {
					It("should interrupt the body with `ERROR` code", func() {
						evaluate("closure cmd1 {} {set var val1}")
						evaluate("closure cmd2 {} {set var val2}")
						Expect(execute("scope {cmd1; error msg; cmd2}")).To(Equal(
							ERROR("msg"),
						))
						Expect(evaluate("get var")).To(Equal(STR("val1")))
					})
					It("should not define the scope command", func() {
						evaluate("scope cmd {error msg}")
						Expect(rootScope.Context.Commands["cmd"]).To(BeNil())
					})
				})
				Describe("`break`", func() {
					It("should interrupt the body with `ERROR` code", func() {
						evaluate("closure cmd1 {} {set var val1}")
						evaluate("closure cmd2 {} {set var val2}")
						Expect(execute("scope {cmd1; break; cmd2}")).To(Equal(
							ERROR("unexpected break"),
						))
						Expect(evaluate("get var")).To(Equal(STR("val1")))
					})
					It("should not define the scope command", func() {
						evaluate("scope cmd {break}")
						Expect(rootScope.Context.Commands["cmd"]).To(BeNil())
					})
				})
				Describe("`continue`", func() {
					It("should interrupt the body with `ERROR` code", func() {
						evaluate("closure cmd1 {} {set var val1}")
						evaluate("closure cmd2 {} {set var val2}")
						Expect(execute("scope {cmd1; continue; cmd2}")).To(Equal(
							ERROR("unexpected continue"),
						))
						Expect(evaluate("get var")).To(Equal(STR("val1")))
					})
					It("should not define the scope command", func() {
						evaluate("scope cmd {continue}")
						Expect(rootScope.Context.Commands["cmd"]).To(BeNil())
					})
				})
			})
		})

		Describe("Subcommands", func() {
			Describe("`subcommands`", func() {
				It("should return list of subcommands", func() {
					Expect(evaluate("[scope {}] subcommands")).To(Equal(
						evaluate("list (subcommands eval call)"),
					))
				})

				Describe("Exceptions", func() {
					Specify("wrong arity", func() {
						Expect(execute("[scope {}] subcommands a")).To(Equal(
							ERROR(`wrong # args: should be "<scope> subcommands"`),
						))
					})
				})
			})

			Describe("`eval`", func() {
				It("should evaluate body", func() {
					evaluate("scope cmd {let cst val}")
					Expect(evaluate("cmd eval {get cst}")).To(Equal(STR("val")))
				})
				It("should accept tuple bodies", func() {
					evaluate("scope cmd {let cst val}")
					Expect(evaluate("cmd eval (get cst)")).To(Equal(STR("val")))
				})
				It("should evaluate macros in scope", func() {
					evaluate("scope cmd {macro mac {} {let cst val}}")
					evaluate("cmd eval {mac}")
					Expect(rootScope.Context.Constants["cst"]).To(BeNil())
					Expect(evaluate("cmd eval {get cst}")).To(Equal(STR("val")))
				})
				It("should evaluate closures in their scope", func() {
					evaluate("closure cls {} {let cst val}")
					evaluate("scope cmd {}")
					evaluate("cmd eval {cls}")
					Expect(rootScope.Context.Constants["cst"]).To(Equal(STR("val")))
					Expect(execute("cmd eval {get cst}").Code).To(Equal(core.ResultCode_ERROR))
				})

				Describe("Control flow", func() {
					Describe("`return`", func() {
						It("should interrupt the body with `RETURN` code", func() {
							evaluate("closure cmd1 {} {set var val1}")
							evaluate("closure cmd2 {} {set var val2}")
							evaluate("scope cmd {}")
							Expect(execute("cmd eval {cmd1; return val3; cmd2}")).To(Equal(
								RETURN(STR("val3")),
							))
							Expect(evaluate("get var")).To(Equal(STR("val1")))
						})
					})
					Describe("`tailcall`", func() {
						It("should interrupt the body with `RETURN` code", func() {
							evaluate("closure cmd1 {} {set var val1}")
							evaluate("closure cmd2 {} {set var val2}")
							evaluate("scope cmd {}")
							Expect(
								execute("cmd eval {cmd1; tailcall {idem val3}; cmd2}"),
							).To(Equal(RETURN(STR("val3"))))
							Expect(evaluate("get var")).To(Equal(STR("val1")))
						})
					})
					Describe("`yield`", func() {
						It("should interrupt the body with `YIELD` code", func() {
							evaluate("closure cmd1 {} {set var val1}")
							evaluate("closure cmd2 {} {set var val2}")
							evaluate("scope cmd {}")
							Expect(execute("cmd eval {cmd1; yield; cmd2}").Code).To(Equal(
								core.ResultCode_YIELD,
							))
							Expect(evaluate("get var")).To(Equal(STR("val1")))
						})
						It("should provide a resumable state", func() {
							evaluate("closure cmd1 {} {set var val1}")
							evaluate("closure cmd2 {val} {set var $val}")
							evaluate("scope cmd {}")
							process := prepareScript(
								"cmd eval {cmd1; cmd2 _[yield val2]_}",
							)

							result := process.Run()
							Expect(result.Code).To(Equal(core.ResultCode_YIELD))
							Expect(result.Value).To(Equal(STR("val2")))

							process.YieldBack(STR("val3"))
							result = process.Run()
							Expect(result).To(Equal(OK(STR("_val3_"))))
							Expect(evaluate("get var")).To(Equal(STR("_val3_")))
						})
					})
					Describe("`error`", func() {
						It("should interrupt the body with `ERROR` code", func() {
							evaluate("closure cmd1 {} {set var val1}")
							evaluate("closure cmd2 {} {set var val2}")
							evaluate("scope cmd {}")
							Expect(execute("cmd eval {cmd1; error msg; cmd2}")).To(Equal(
								ERROR("msg"),
							))
							Expect(evaluate("get var")).To(Equal(STR("val1")))
						})
					})
					Describe("`break`", func() {
						It("should interrupt the body with `BREAK` code", func() {
							evaluate("closure cmd1 {} {set var val1}")
							evaluate("closure cmd2 {} {set var val2}")
							evaluate("scope cmd {}")
							Expect(execute("cmd eval {cmd1; break; cmd2}")).To(Equal(BREAK(NIL)))
							Expect(evaluate("get var")).To(Equal(STR("val1")))
						})
					})
					Describe("`continue`", func() {
						It("should interrupt the body with `CONTINUE` code", func() {
							evaluate("closure cmd1 {} {set var val1}")
							evaluate("closure cmd2 {} {set var val2}")
							evaluate("scope cmd {}")
							Expect(execute("cmd eval {cmd1; continue; cmd2}")).To(Equal(
								CONTINUE(NIL),
							))
							Expect(evaluate("get var")).To(Equal(STR("val1")))
						})
					})
				})

				Describe("Exceptions", func() {
					Specify("wrong arity", func() {
						Expect(execute("[scope {}] eval")).To(Equal(
							ERROR(`wrong # args: should be "<scope> eval body"`),
						))
						Expect(execute("[scope {}] eval a b")).To(Equal(
							ERROR(`wrong # args: should be "<scope> eval body"`),
						))
					})
					Specify("invalid body", func() {
						Expect(execute("[scope {}] eval 1")).To(Equal(
							ERROR("body must be a script or tuple"),
						))
					})
				})
			})

			Describe("`call`", func() {
				It("should call scope commands", func() {
					evaluate("scope cmd {macro mac {} {idem val}}")
					Expect(evaluate("cmd call mac")).To(Equal(STR("val")))
				})
				It("should evaluate macros in scope", func() {
					evaluate("scope cmd {macro mac {} {let cst val}}")
					evaluate("cmd call mac")
					Expect(rootScope.Context.Constants["cst"]).To(BeNil())
					Expect(evaluate("cmd eval {get cst}")).To(Equal(STR("val")))
				})
				It("should evaluate closures in scope", func() {
					evaluate("scope cmd {closure cls {} {let cst val}}")
					evaluate("cmd call cls")
					Expect(rootScope.Context.Constants["cst"]).To(BeNil())
					Expect(evaluate("cmd eval {get cst}")).To(Equal(STR("val")))
				})

				Describe("Control flow", func() {
					Describe("`return`", func() {
						It("should interrupt the body with `RETURN` code", func() {
							evaluate("closure cmd1 {} {set var val1}")
							evaluate("closure cmd2 {} {set var val2}")
							evaluate("scope cmd {macro mac {} {cmd1; return val3; cmd2}}")
							Expect(execute("cmd call mac")).To(Equal(RETURN(STR("val3"))))
							Expect(evaluate("get var")).To(Equal(STR("val1")))
						})
					})
					Describe("`tailcall`", func() {
						It("should interrupt the body with `RETURN` code", func() {
							evaluate("closure cmd1 {} {set var val1}")
							evaluate("closure cmd2 {} {set var val2}")
							evaluate(
								"scope cmd {macro mac {} {cmd1; tailcall {idem val3}; cmd2}}",
							)
							Expect(execute("cmd call mac")).To(Equal(RETURN(STR("val3"))))
							Expect(evaluate("get var")).To(Equal(STR("val1")))
						})
					})
					Describe("`yield`", func() {
						It("should interrupt the body with `YIELD` code", func() {
							evaluate("closure cmd1 {} {set var val1}")
							evaluate("closure cmd2 {} {set var val2}")
							evaluate("scope cmd {macro mac {} {cmd1; yield; cmd2}}")
							Expect(execute("cmd call mac").Code).To(Equal(core.ResultCode_YIELD))
							Expect(evaluate("get var")).To(Equal(STR("val1")))
						})
						It("should provide a resumable state", func() {
							evaluate("closure cmd1 {} {set var val1}")
							evaluate("closure cmd2 {val} {set var $val}")
							evaluate("scope cmd {macro mac {} {cmd1; cmd2 _[yield val2]_}}")
							process := prepareScript("cmd call mac")

							result := process.Run()
							Expect(result.Code).To(Equal(core.ResultCode_YIELD))
							Expect(result.Value).To(Equal(STR("val2")))

							process.YieldBack(STR("val3"))
							result = process.Run()
							Expect(result).To(Equal(OK(STR("_val3_"))))
							Expect(evaluate("get var")).To(Equal(STR("_val3_")))
						})
					})
					Describe("`error`", func() {
						It("should interrupt the body with `ERROR` code", func() {
							evaluate("closure cmd1 {} {set var val1}")
							evaluate("closure cmd2 {} {set var val2}")
							evaluate("scope cmd {macro mac {} {cmd1; error msg; cmd2}}")
							Expect(execute("cmd call mac")).To(Equal(ERROR("msg")))
							Expect(evaluate("get var")).To(Equal(STR("val1")))
						})
					})
					Describe("`break`", func() {
						It("should interrupt the body with `BREAK` code", func() {
							evaluate("closure cmd1 {} {set var val1}")
							evaluate("closure cmd2 {} {set var val2}")
							evaluate("scope cmd {macro mac {} {cmd1; break; cmd2}}")
							Expect(execute("cmd call mac")).To(Equal(BREAK(NIL)))
							Expect(evaluate("get var")).To(Equal(STR("val1")))
						})
					})
					Describe("`continue`", func() {
						It("should interrupt the body with `CONTINUE` code", func() {
							evaluate("closure cmd1 {} {set var val1}")
							evaluate("closure cmd2 {} {set var val2}")
							evaluate("scope cmd {macro mac {} {cmd1; continue; cmd2}}")
							Expect(execute("cmd call mac")).To(Equal(CONTINUE(NIL)))
							Expect(evaluate("get var")).To(Equal(STR("val1")))
						})
					})
				})

				Describe("Exceptions", func() {
					Specify("wrong arity", func() {
						Expect(execute("[scope {}] call")).To(Equal(
							ERROR(`wrong # args: should be "<scope> call cmdname ?arg ...?"`),
						))
					})
					Specify("unknown command", func() {
						Expect(execute("[scope {}] call unknownCommand")).To(Equal(
							ERROR(`unknown command "unknownCommand"`),
						))
					})
					Specify("out-of-scope command", func() {
						Expect(execute("macro cmd {} {}; [scope {}] call cmd")).To(Equal(
							ERROR(`unknown command "cmd"`),
						))
					})
					Specify("invalid command name", func() {
						Expect(execute("[scope {}] call []")).To(Equal(
							ERROR("invalid command name"),
						))
					})
				})
			})

			Describe("Exceptions", func() {
				Specify("unknown subcommand", func() {
					Expect(execute("[scope {}] unknownSubcommand")).To(Equal(
						ERROR(`unknown subcommand "unknownSubcommand"`),
					))
				})
				Specify("invalid subcommand name", func() {
					Expect(execute("[scope {}] []")).To(Equal(
						ERROR("invalid subcommand name"),
					))
				})
			})
		})
	})
})
