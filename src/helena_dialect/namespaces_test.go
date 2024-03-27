package helena_dialect_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"helena/core"
	. "helena/helena_dialect"
)

var _ = Describe("Helena namespaces", func() {
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

	BeforeEach(init)

	Describe("namespace", func() {
		Describe("Specifications", func() {
			Specify("usage", func() {
				Expect(evaluate("help namespace")).To(Equal(STR("namespace ?name? body")))
				Expect(evaluate("help namespace {}")).To(Equal(
					STR("namespace ?name? body"),
				))
				Expect(evaluate("help namespace cmd {}")).To(Equal(
					STR("namespace ?name? body"),
				))
			})

			It("should define a new command", func() {
				evaluate("namespace cmd {}")
				Expect(rootScope.Context.Commands["cmd"]).NotTo(BeNil())
			})
			It("should replace existing commands", func() {
				evaluate("namespace cmd {}")
				Expect(execute("namespace cmd {}").Code).To(Equal(core.ResultCode_OK))
			})
		})

		Describe("Exceptions", func() {
			Specify("wrong arity", func() {
				Expect(execute("namespace")).To(Equal(
					ERROR(`wrong # args: should be "namespace ?name? body"`),
				))
				Expect(execute("namespace a b c")).To(Equal(
					ERROR(`wrong # args: should be "namespace ?name? body"`),
				))
				Expect(execute("help namespace a b c")).To(Equal(
					ERROR(`wrong # args: should be "namespace ?name? body"`),
				))
			})
			Specify("invalid `name`", func() {
				Expect(execute("namespace [] {}")).To(Equal(
					ERROR("invalid command name"),
				))
			})
			Specify("non-script body", func() {
				Expect(execute("namespace a")).To(Equal(ERROR("body must be a script")))
				Expect(execute("namespace a b")).To(Equal(ERROR("body must be a script")))
			})
		})

		Describe("`body`", func() {
			It("should be executed", func() {
				evaluate("closure cmd {} {let var val}")
				Expect(rootScope.Context.Constants["var"]).To(BeNil())
				evaluate("namespace {cmd}")
				Expect(rootScope.Context.Constants["var"]).To(Equal(STR("val")))
			})
			It("should access global commands", func() {
				Expect(execute("namespace {idem val}").Code).To(Equal(core.ResultCode_OK))
			})
			It("should not access global variables", func() {
				evaluate("set var val")
				Expect(execute("namespace {get var}").Code).To(Equal(core.ResultCode_ERROR))
			})
			It("should not set global variables", func() {
				evaluate("set var val")
				evaluate("namespace {set var val2; let cst val3}")
				Expect(rootScope.Context.Variables["var"]).To(Equal(STR("val")))
				Expect(rootScope.Context.Constants["cst"]).To(BeNil())
			})
			It("should set namespace variables", func() {
				evaluate("set var val")
				evaluate("namespace cmd {set var val2; let cst val3}")
				Expect(rootScope.Context.Variables["var"]).To(Equal(STR("val")))
				Expect(rootScope.Context.Constants["cst"]).To(BeNil())
				Expect(evaluate("[cmd] eval {get var}")).To(Equal(STR("val2")))
				Expect(evaluate("[cmd] eval {get cst}")).To(Equal(STR("val3")))
			})

			Describe("Control flow", func() {
				Describe("`return`", func() {
					It("should interrupt the body with `OK` code", func() {
						evaluate("closure cmd1 {} {set var val1}")
						evaluate("closure cmd2 {} {set var val2}")
						Expect(execute("namespace {cmd1; return; cmd2}").Code).To(Equal(
							core.ResultCode_OK,
						))
						Expect(evaluate("get var")).To(Equal(STR("val1")))
					})
					It("should still define the named command", func() {
						evaluate("namespace cmd {return}")
						Expect(rootScope.Context.Commands["cmd"]).NotTo(BeNil())
					})
					It("should return passed value instead of the command object", func() {
						Expect(execute("namespace {return val}")).To(Equal(OK(STR("val"))))
					})
				})
				Describe("`tailcall`", func() {
					It("should interrupt the body with `OK` code", func() {
						evaluate("closure cmd1 {} {set var val1}")
						evaluate("closure cmd2 {} {set var val2}")
						Expect(execute("namespace {cmd1; tailcall {}; cmd2}").Code).To(Equal(
							core.ResultCode_OK,
						))
						Expect(evaluate("get var")).To(Equal(STR("val1")))
					})
					It("should still define the named command", func() {
						evaluate("namespace cmd {tailcall {}}")
						Expect(rootScope.Context.Commands["cmd"]).NotTo(BeNil())
					})
					It("should return passed value instead of the command object", func() {
						Expect(execute("namespace {tailcall {idem val}}")).To(Equal(
							OK(STR("val")),
						))
					})
				})
				Describe("`yield`", func() {
					It("should interrupt the body with `YIELD` code", func() {
						evaluate("closure cmd1 {} {set var val1}")
						evaluate("closure cmd2 {} {set var val2}")
						Expect(execute("namespace cmd {cmd1; yield; cmd2}").Code).To(Equal(
							core.ResultCode_YIELD,
						))
						Expect(evaluate("get var")).To(Equal(STR("val1")))
					})
					It("should provide a resumable state", func() {
						evaluate("closure cmd1 {} {set var val1}")
						evaluate("closure cmd2 {val} {set var $val}")
						process := rootScope.PrepareScript(
							*parse("namespace cmd {cmd1; cmd2 _[yield val2]_}"),
						)

						result := process.Run()
						Expect(result.Code).To(Equal(core.ResultCode_YIELD))
						Expect(result.Value).To(Equal(STR("val2")))
						Expect(result.Data).NotTo(BeNil())

						process.YieldBack(STR("val3"))
						result = process.Run()
						Expect(result.Code).To(Equal(core.ResultCode_OK))
						Expect(result.Value.Type()).To(Equal(core.ValueType_COMMAND))
						Expect(evaluate("get var")).To(Equal(STR("_val3_")))
					})
					It("should delay the definition of namespace command until resumed", func() {
						process := rootScope.PrepareScript(
							*parse("namespace cmd {yield}"),
						)

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
						Expect(execute("namespace {cmd1; error msg; cmd2}")).To(Equal(
							ERROR("msg"),
						))
						Expect(evaluate("get var")).To(Equal(STR("val1")))
					})
					It("should not define the namespace command", func() {
						evaluate("namespace cmd {error msg}")
						Expect(rootScope.Context.Commands["cmd"]).To(BeNil())
					})
				})
				Describe("`break`", func() {
					It("should interrupt the body with `ERROR` code", func() {
						evaluate("closure cmd1 {} {set var val1}")
						evaluate("closure cmd2 {} {set var val2}")
						Expect(execute("namespace {cmd1; break; cmd2}")).To(Equal(
							ERROR("unexpected break"),
						))
						Expect(evaluate("get var")).To(Equal(STR("val1")))
					})
					It("should not define the namespace command", func() {
						evaluate("namespace cmd {break}")
						Expect(rootScope.Context.Commands["cmd"]).To(BeNil())
					})
				})
				Describe("`continue`", func() {
					It("should interrupt the body with `ERROR` code", func() {
						evaluate("closure cmd1 {} {set var val1}")
						evaluate("closure cmd2 {} {set var val2}")
						Expect(execute("namespace {cmd1; continue; cmd2}")).To(Equal(
							ERROR("unexpected continue"),
						))
						Expect(evaluate("get var")).To(Equal(STR("val1")))
					})
					It("should not define the namespace command", func() {
						evaluate("namespace cmd {continue}")
						Expect(rootScope.Context.Commands["cmd"]).To(BeNil())
					})
				})
			})
		})

		Describe("Metacommand", func() {
			It("should return a metacommand", func() {
				Expect(evaluate("namespace {}").Type()).To(Equal(core.ValueType_COMMAND))
				Expect(evaluate("namespace cmd {}").Type()).To(Equal(core.ValueType_COMMAND))
			})
			Specify("the metacommand should return itself", func() {
				value := evaluate("set cmd [namespace {}]")
				Expect(evaluate("$cmd")).To(Equal(value))
			})

			Describe("Subcommands", func() {
				Describe("`subcommands`", func() {
					It("should return list of subcommands", func() {
						Expect(evaluate("[namespace {}] subcommands")).To(Equal(
							evaluate("list (subcommands eval call import)"),
						))
					})

					Describe("Exceptions", func() {
						Specify("wrong arity", func() {
							Expect(execute("[namespace {}] subcommands a")).To(Equal(
								ERROR(`wrong # args: should be "<namespace> subcommands"`),
							))
						})
					})
				})

				Describe("`eval`", func() {
					It("should evaluate body in namespace scope", func() {
						evaluate("namespace cmd {let cst val}")
						Expect(evaluate("[cmd] eval {get cst}")).To(Equal(STR("val")))
					})
					It("should accept tuple bodies", func() {
						evaluate("namespace cmd {let cst val}")
						Expect(evaluate("[cmd] eval (get cst)")).To(Equal(STR("val")))
					})
					It("should evaluate macros in namespace scope", func() {
						evaluate("namespace cmd {macro mac {} {let cst val}}")
						evaluate("[cmd] eval {mac}")
						Expect(rootScope.Context.Constants["cst"]).To(BeNil())
						Expect(evaluate("[cmd] eval {get cst}")).To(Equal(STR("val")))
					})
					It("should evaluate closures in their scope", func() {
						evaluate("closure cls {} {let cst val}")
						evaluate("namespace cmd {}")
						evaluate("[cmd] eval {cls}")
						Expect(rootScope.Context.Constants["cst"]).To(Equal(STR("val")))
						Expect(execute("[cmd] eval {get cst}").Code).To(Equal(
							core.ResultCode_ERROR,
						))
					})

					Describe("Control flow", func() {
						Describe("`return`", func() {
							It("should interrupt the body with `RETURN` code", func() {
								evaluate("closure cmd1 {} {set var val1}")
								evaluate("closure cmd2 {} {set var val2}")
								evaluate("namespace cmd {}")
								Expect(execute("[cmd] eval {cmd1; return val3; cmd2}")).To(Equal(
									RETURN(STR("val3")),
								))
								Expect(evaluate("get var")).To(Equal(STR("val1")))
							})
						})
						Describe("`tailcall`", func() {
							It("should interrupt the body with `RETURN` code", func() {
								evaluate("closure cmd1 {} {set var val1}")
								evaluate("closure cmd2 {} {set var val2}")
								evaluate("namespace cmd {}")
								Expect(
									execute("[cmd] eval {cmd1; tailcall {idem val3}; cmd2}"),
								).To(Equal(RETURN(STR("val3"))))
								Expect(evaluate("get var")).To(Equal(STR("val1")))
							})
						})
						Describe("`yield`", func() {
							It("should interrupt the body with `YIELD` code", func() {
								evaluate("closure cmd1 {} {set var val1}")
								evaluate("closure cmd2 {} {set var val2}")
								evaluate("namespace cmd {}")
								Expect(execute("[cmd] eval {cmd1; yield; cmd2}").Code).To(Equal(
									core.ResultCode_YIELD,
								))
								Expect(evaluate("get var")).To(Equal(STR("val1")))
							})
							It("should provide a resumable state", func() {
								evaluate("closure cmd1 {} {set var val1}")
								evaluate("closure cmd2 {val} {set var $val}")
								evaluate("namespace cmd {}")
								process := rootScope.PrepareScript(
									*parse("[cmd] eval {cmd1; cmd2 _[yield val2]_}"),
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
								evaluate("namespace cmd {}")
								Expect(execute("[cmd] eval {cmd1; error msg; cmd2}")).To(Equal(
									ERROR("msg"),
								))
								Expect(evaluate("get var")).To(Equal(STR("val1")))
							})
						})
						Describe("`break`", func() {
							It("should interrupt the body with `BREAK` code", func() {
								evaluate("closure cmd1 {} {set var val1}")
								evaluate("closure cmd2 {} {set var val2}")
								evaluate("namespace cmd {}")
								Expect(execute("[cmd] eval {cmd1; break; cmd2}")).To(Equal(
									BREAK(NIL),
								))
								Expect(evaluate("get var")).To(Equal(STR("val1")))
							})
						})
						Describe("`continue`", func() {
							It("should interrupt the body with `CONTINUE` code", func() {
								evaluate("closure cmd1 {} {set var val1}")
								evaluate("closure cmd2 {} {set var val2}")
								evaluate("namespace cmd {}")
								Expect(execute("[cmd] eval {cmd1; continue; cmd2}")).To(Equal(
									CONTINUE(NIL),
								))
								Expect(evaluate("get var")).To(Equal(STR("val1")))
							})
						})
					})

					Describe("Exceptions", func() {
						Specify("wrong arity", func() {
							Expect(execute("[namespace {}] eval")).To(Equal(
								ERROR(`wrong # args: should be "<namespace> eval body"`),
							))
							Expect(execute("[namespace {}] eval a b")).To(Equal(
								ERROR(`wrong # args: should be "<namespace> eval body"`),
							))
						})
						Specify("invalid body", func() {
							Expect(execute("[namespace {}] eval 1")).To(Equal(
								ERROR("body must be a script or tuple"),
							))
						})
					})
				})

				Describe("`call`", func() {
					It("should call namespace commands", func() {
						evaluate("namespace cmd {macro mac {} {idem val}}")
						Expect(evaluate("[cmd] call mac")).To(Equal(STR("val")))
					})
					It("should evaluate macros in namespace", func() {
						evaluate("namespace cmd {macro mac {} {let cst val}}")
						evaluate("[cmd] call mac")
						Expect(rootScope.Context.Constants["cst"]).To(BeNil())
						Expect(evaluate("[cmd] eval {get cst}")).To(Equal(STR("val")))
					})
					It("should evaluate namespace closures in namespace", func() {
						evaluate("namespace cmd {closure cls {} {let cst val}}")
						evaluate("[cmd] call cls")
						Expect(rootScope.Context.Constants["cst"]).To(BeNil())
						Expect(evaluate("[cmd] eval {get cst}")).To(Equal(STR("val")))
					})

					Describe("Control flow", func() {
						Describe("`return`", func() {
							It("should interrupt the body with `RETURN` code", func() {
								evaluate("closure cmd1 {} {set var val1}")
								evaluate("closure cmd2 {} {set var val2}")
								evaluate(
									"namespace cmd {macro mac {} {cmd1; return val3; cmd2}}",
								)
								Expect(execute("[cmd] call mac")).To(Equal(RETURN(STR("val3"))))
								Expect(evaluate("get var")).To(Equal(STR("val1")))
							})
						})
						Describe("`tailcall`", func() {
							It("should interrupt the body with `RETURN` code", func() {
								evaluate("closure cmd1 {} {set var val1}")
								evaluate("closure cmd2 {} {set var val2}")
								evaluate(
									"namespace cmd {macro mac {} {cmd1; tailcall {idem val3}; cmd2}}",
								)
								Expect(execute("[cmd] call mac")).To(Equal(RETURN(STR("val3"))))
								Expect(evaluate("get var")).To(Equal(STR("val1")))
							})
						})
						Describe("`yield`", func() {
							It("should interrupt the call with `YIELD` code", func() {
								evaluate("closure cmd1 {} {set var val1}")
								evaluate("closure cmd2 {} {set var val2}")
								evaluate("namespace cmd {macro mac {} {cmd1; yield; cmd2}}")
								Expect(execute("[cmd] call mac").Code).To(Equal(core.ResultCode_YIELD))
								Expect(evaluate("get var")).To(Equal(STR("val1")))
							})
							It("should provide a resumable state", func() {
								evaluate("closure cmd1 {} {set var val1}")
								evaluate("closure cmd2 {val} {set var $val}")
								evaluate(
									"namespace cmd {proc p {} {cmd1; cmd2 _[yield val2]_}}",
								)
								process := rootScope.PrepareScript(*parse("[cmd] call p"))

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
								evaluate(
									"namespace cmd {macro mac {} {cmd1; error msg; cmd2}}",
								)
								Expect(execute("[cmd] call mac")).To(Equal(ERROR("msg")))
								Expect(evaluate("get var")).To(Equal(STR("val1")))
							})
						})
						Describe("`break`", func() {
							It("should interrupt the body with `BREAK` code", func() {
								evaluate("closure cmd1 {} {set var val1}")
								evaluate("closure cmd2 {} {set var val2}")
								evaluate("namespace cmd {macro mac {} {cmd1; break; cmd2}}")
								Expect(execute("[cmd] call mac")).To(Equal(BREAK(NIL)))
								Expect(evaluate("get var")).To(Equal(STR("val1")))
							})
						})
						Describe("`continue`", func() {
							It("should interrupt the body with `CONTINUE` code", func() {
								evaluate("closure cmd1 {} {set var val1}")
								evaluate("closure cmd2 {} {set var val2}")
								evaluate("namespace cmd {macro mac {} {cmd1; continue; cmd2}}")
								Expect(execute("[cmd] call mac")).To(Equal(CONTINUE(NIL)))
								Expect(evaluate("get var")).To(Equal(STR("val1")))
							})
						})
					})

					Describe("Exceptions", func() {
						Specify("wrong arity", func() {
							Expect(execute("[namespace {}] call")).To(Equal(
								ERROR(
									`wrong # args: should be "<namespace> call cmdname ?arg ...?"`,
								),
							))
						})
						Specify("unknown command", func() {
							Expect(execute("[namespace {}] call unknownCommand")).To(Equal(
								ERROR(`unknown command "unknownCommand"`),
							))
						})
						Specify("out-of-scope command", func() {
							Expect(
								execute("macro cmd {} {}; [namespace {}] call cmd"),
							).To(Equal(ERROR(`unknown command "cmd"`)))
						})
						Specify("invalid command name", func() {
							Expect(execute("[namespace {}] call []")).To(Equal(
								ERROR("invalid command name"),
							))
						})
					})
				})

				Describe("`import`", func() {
					It("should declare imported commands in the calling scope", func() {
						evaluate(`namespace ns {macro cmd {} {idem value}}`)
						evaluate("[ns] import cmd")
						Expect(evaluate("cmd")).To(Equal(STR("value")))
					})
					It("should return nil", func() {
						evaluate(`namespace ns {macro cmd {} {idem value}}`)
						Expect(execute("[ns] import cmd")).To(Equal(OK(NIL)))
					})
					It("should replace existing commands", func() {
						evaluate("closure cmd {} {idem val1} ")
						Expect(evaluate("cmd")).To(Equal(STR("val1")))
						evaluate(`namespace ns {macro cmd {} {idem val2}}`)
						evaluate("[ns] import cmd")
						Expect(evaluate("cmd")).To(Equal(STR("val2")))
					})
					It("should evaluate macros in the caller scope", func() {
						evaluate(`namespace ns {macro cmd {} {set var val}}`)
						evaluate("[ns] import cmd")
						evaluate("cmd")
						Expect(evaluate("get var")).To(Equal(STR("val")))
					})
					It("should evaluate closures in their scope", func() {
						evaluate(`namespace ns {set var val; closure cmd {} {get var}}`)
						evaluate("[ns] import cmd")
						Expect(evaluate("cmd")).To(Equal(STR("val")))
						Expect(execute("get var").Code).To(Equal(core.ResultCode_ERROR))
					})
					It("should resolve imported commands at call time", func() {
						evaluate(`
							namespace ns {
								closure cmd {} {idem val1}
								closure redefine {} {
									closure cmd {} {idem val2}
								}
							}
						`)
						Expect(evaluate("[ns] import cmd; cmd")).To(Equal(STR("val1")))
						evaluate("ns redefine")
						Expect(evaluate("cmd")).To(Equal(STR("val1")))
						Expect(evaluate("[ns] import cmd; cmd")).To(Equal(STR("val2")))
					})
					It("should accept an optional alias name", func() {
						evaluate("macro cmd {} {idem original}")
						evaluate(`namespace ns {macro cmd {} {idem imported}}`)
						evaluate("[ns] import cmd cmd2")
						Expect(evaluate("cmd")).To(Equal(STR("original")))
						Expect(evaluate("cmd2")).To(Equal(STR("imported")))
					})

					Describe("Exceptions", func() {
						Specify("wrong arity", func() {
							Expect(execute("[namespace {}] import")).To(Equal(
								ERROR(
									`wrong # args: should be "<namespace> import name ?alias?"`,
								),
							))
							Expect(execute("[namespace {}] import a b c")).To(Equal(
								ERROR(
									`wrong # args: should be "<namespace> import name ?alias?"`,
								),
							))
						})
						Specify("unresolved command", func() {
							Expect(execute("[namespace {}] import a")).To(Equal(
								ERROR(`cannot resolve imported command "a"`),
							))
						})
						Specify("invalid import name", func() {
							Expect(execute("[namespace {}] import []")).To(Equal(
								ERROR("invalid import name"),
							))
						})
						Specify("invalid alias name", func() {
							Expect(execute("[namespace {}] import a []")).To(Equal(
								ERROR("invalid alias name"),
							))
						})
					})
				})

				Describe("Exceptions", func() {
					Specify("unknown subcommand", func() {
						Expect(execute("[namespace {}] unknownSubcommand")).To(Equal(
							ERROR(`unknown subcommand "unknownSubcommand"`),
						))
					})
					Specify("invalid subcommand name", func() {
						Expect(execute("[namespace {}] []")).To(Equal(
							ERROR("invalid subcommand name"),
						))
					})
				})
			})
		})
	})

	Describe("Namespace commands", func() {
		Describe("Specifications", func() {
			It("should return its namespace metacommand when called with no argument", func() {
				value := evaluate("namespace cmd {}")
				Expect(evaluate("cmd")).To(Equal(value))
			})
		})

		Describe("Namespace subcommands", func() {
			Specify("first argument should be namespace subcommand name", func() {
				evaluate("namespace cmd {macro opt {} {idem val}}")
				Expect(evaluate("cmd opt")).To(Equal(STR("val")))
			})
			It("should pass remaining arguments to namespace subcommand", func() {
				evaluate("namespace cmd {macro opt {arg} {idem $arg}}")
				Expect(evaluate("cmd opt val")).To(Equal(STR("val")))
			})
			It("should evaluate subcommand in namespace scope", func() {
				evaluate("namespace cmd {macro mac {} {let cst val}}")
				evaluate("cmd mac")
				Expect(rootScope.Context.Constants["cst"]).To(BeNil())
				Expect(evaluate("[cmd] eval {get cst}")).To(Equal(STR("val")))
			})
			It("should work recursively", func() {
				evaluate("namespace ns1 {namespace ns2 {macro opt {} {idem val}}}")
				Expect(evaluate("ns1 ns2 opt")).To(Equal(STR("val")))
			})

			Describe("Introspection", func() {
				Describe("`subcommands`", func() {
					It("should return list of subcommands", func() {
						evaluate("namespace cmd {}")
						Expect(evaluate("cmd subcommands")).To(Equal(
							evaluate("list (subcommands)"),
						))
						evaluate("[cmd] eval {macro mac {} {}}")
						Expect(evaluate("cmd subcommands")).To(Equal(
							evaluate("list (subcommands mac)"),
						))
					})

					Describe("Exceptions", func() {
						Specify("wrong arity", func() {
							evaluate("namespace cmd {}")
							Expect(execute("cmd subcommands a")).To(Equal(
								ERROR(`wrong # args: should be "cmd subcommands"`),
							))
							Expect(execute("help cmd subcommands a")).To(Equal(
								ERROR(`wrong # args: should be "cmd subcommands"`),
							))
						})
					})
				})
			})

			Describe("Help", func() {
				It("should provide subcommand help", func() {
					evaluate(`
						namespace cmd {
							macro opt1 {a} {}
							closure opt2 {b} {}
						}
					`)
					Expect(evaluate("help cmd")).To(Equal(
						STR("cmd ?subcommand? ?arg ...?"),
					))
					Expect(evaluate("help cmd subcommands")).To(Equal(
						STR("cmd subcommands"),
					))
					Expect(evaluate("help cmd opt1")).To(Equal(STR("cmd opt1 a")))
					Expect(evaluate("help cmd opt1 1")).To(Equal(STR("cmd opt1 a")))
					Expect(evaluate("help cmd opt2")).To(Equal(STR("cmd opt2 b")))
					Expect(evaluate("help cmd opt2 2")).To(Equal(STR("cmd opt2 b")))
				})
				It("should work recursively", func() {
					evaluate(`
						namespace cmd {
							namespace sub {
								macro opt {a} {}
							}
						}
					`)
					Expect(evaluate("help cmd sub")).To(Equal(
						STR("cmd sub ?subcommand? ?arg ...?"),
					))
					Expect(evaluate("help cmd sub subcommands")).To(Equal(
						STR("cmd sub subcommands"),
					))
					Expect(evaluate("help cmd sub opt")).To(Equal(STR("cmd sub opt a")))
					Expect(evaluate("help cmd sub opt 1")).To(Equal(STR("cmd sub opt a")))
				})

				Describe("Exceptions", func() {
					Specify("wrong arity", func() {
						evaluate(`
						namespace cmd {
							macro opt {a} {}
							namespace sub {
								macro opt {b} {}
							}
						}
						`)
						Expect(execute("help cmd subcommands 1")).To(Equal(
							ERROR(`wrong # args: should be "cmd subcommands"`),
						))
						Expect(execute("help cmd opt 1 2")).To(Equal(
							ERROR(`wrong # args: should be "cmd opt a"`),
						))
						Expect(execute("help cmd sub subcommands 1")).To(Equal(
							ERROR(`wrong # args: should be "cmd sub subcommands"`),
						))
						Expect(execute("help cmd sub opt 1 2")).To(Equal(
							ERROR(`wrong # args: should be "cmd sub opt b"`),
						))
					})
					Specify("invalid `subcommand`", func() {
						evaluate("namespace cmd {}")
						Expect(execute("help cmd []")).To(Equal(
							ERROR("invalid subcommand name"),
						))
					})
					Specify("unknown subcommand", func() {
						evaluate("namespace cmd {}")
						Expect(execute("help cmd unknownSubcommand")).To(Equal(
							ERROR(`unknown subcommand "unknownSubcommand"`),
						))
					})
					Specify("subcommand with no help", func() {
						rootScope.RegisterNamedCommand("foo", simpleCommand{
							func(_ []core.Value, _ any) core.Result { return OK(NIL) },
						})
						evaluate("namespace cmd {alias opt foo}")
						Expect(execute("help cmd opt")).To(Equal(
							ERROR(`no help for subcommand "opt"`),
						))
					})
				})
			})

			Describe("Control flow", func() {
				Describe("`return`", func() {
					It("should interrupt the call with `RETURN` code", func() {
						evaluate("closure cmd1 {} {set var val1}")
						evaluate("closure cmd2 {} {set var val2}")
						evaluate("namespace cmd {macro mac {} {cmd1; return val3; cmd2}}")
						Expect(execute("cmd mac")).To(Equal(RETURN(STR("val3"))))
						Expect(evaluate("get var")).To(Equal(STR("val1")))
					})
				})
				Describe("`tailcall`", func() {
					It("should interrupt the call with `RETURN` code", func() {
						evaluate("closure cmd1 {} {set var val1}")
						evaluate("closure cmd2 {} {set var val2}")
						evaluate(
							"namespace cmd {macro mac {} {cmd1; tailcall {idem val3}; cmd2}}",
						)
						Expect(execute("cmd mac")).To(Equal(RETURN(STR("val3"))))
						Expect(evaluate("get var")).To(Equal(STR("val1")))
					})
				})
				Describe("`yield`", func() {
					It("should interrupt the call with `YIELD` code", func() {
						evaluate("closure cmd1 {} {set var val1}")
						evaluate("closure cmd2 {} {set var val2}")
						evaluate("namespace cmd {macro mac {} {cmd1; yield; cmd2}}")
						Expect(execute("cmd mac").Code).To(Equal(core.ResultCode_YIELD))
						Expect(evaluate("get var")).To(Equal(STR("val1")))
					})
					It("should provide a resumable state", func() {
						evaluate("closure cmd1 {} {set var val1}")
						evaluate("closure cmd2 {val} {set var $val}")
						evaluate("namespace cmd {proc p {} {cmd1; cmd2 _[yield val2]_}}")
						process := rootScope.PrepareScript(*parse("cmd p"))

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
					It("should interrupt the call with `ERROR` code", func() {
						evaluate("closure cmd1 {} {set var val1}")
						evaluate("closure cmd2 {} {set var val2}")
						evaluate("namespace cmd {macro mac {} {cmd1; error msg; cmd2}}")
						Expect(execute("cmd mac")).To(Equal(ERROR("msg")))
						Expect(evaluate("get var")).To(Equal(STR("val1")))
					})
				})
				Describe("`break`", func() {
					It("should interrupt the call with `BREAK` code", func() {
						evaluate("closure cmd1 {} {set var val1}")
						evaluate("closure cmd2 {} {set var val2}")
						evaluate("namespace cmd {macro mac {} {cmd1; break; cmd2}}")
						Expect(execute("cmd mac")).To(Equal(BREAK(NIL)))
						Expect(evaluate("get var")).To(Equal(STR("val1")))
					})
				})
				Describe("`continue`", func() {
					It("should interrupt the call with `CONTINUE` code", func() {
						evaluate("closure cmd1 {} {set var val1}")
						evaluate("closure cmd2 {} {set var val2}")
						evaluate("namespace cmd {macro mac {} {cmd1; continue; cmd2}}")
						Expect(execute("cmd mac")).To(Equal(CONTINUE(NIL)))
						Expect(evaluate("get var")).To(Equal(STR("val1")))
					})
				})
			})

			Describe("Exceptions", func() {
				Specify("unknown subcommand", func() {
					evaluate("namespace cmd {}")
					Expect(execute("cmd unknownCommand")).To(Equal(
						ERROR(`unknown subcommand "unknownCommand"`),
					))
				})
				Specify("out-of-scope subcommand", func() {
					evaluate("macro mac {} {}; namespace cmd {}")
					Expect(execute("cmd mac")).To(Equal(ERROR(`unknown subcommand "mac"`)))
				})
				Specify("invalid subcommand name", func() {
					evaluate("namespace cmd {}")
					Expect(execute("cmd []")).To(Equal(ERROR("invalid subcommand name")))
				})
			})
		})

		Describe("Namespace variables", func() {
			It("should map to value keys", func() {
				evaluate("set ns [namespace cmd {let cst val1; set var val2}]")
				Expect(evaluate("idem $[cmd](cst)")).To(Equal(STR("val1")))
				Expect(evaluate("idem $[cmd](var)")).To(Equal(STR("val2")))
				Expect(evaluate("idem $ns(cst)")).To(Equal(STR("val1")))
				Expect(evaluate("idem $ns(var)")).To(Equal(STR("val2")))
				evaluate("$ns eval {set var2 val3}")
				Expect(evaluate("idem $ns(var2)")).To(Equal(STR("val3")))
			})
			It("should work recursively", func() {
				evaluate(
					"set ns1 [namespace {set ns2 [namespace {let cst val1; set var val2}]}]",
				)
				Expect(evaluate("idem $ns1(ns2)(cst)")).To(Equal(STR("val1")))
				Expect(evaluate("idem $ns1(ns2)(var)")).To(Equal(STR("val2")))
				Expect(evaluate("idem $ns1(ns2 cst)")).To(Equal(STR("val1")))
				Expect(evaluate("idem $ns1(ns2 var)")).To(Equal(STR("val2")))
			})

			Describe("Exceptions", func() {
				Specify("unknown variables", func() {
					evaluate("namespace cmd {}")
					Expect(execute("$[cmd](unknownVariable)")).To(Equal(
						ERROR(`cannot get "unknownVariable": no such variable`),
					))
				})
				Specify("out-of-scope variable", func() {
					evaluate("let cst var; namespace cmd {}")
					Expect(execute("$[cmd](cst)")).To(Equal(
						ERROR(`cannot get "cst": no such variable`),
					))
				})
				Specify("invalid variable name", func() {
					evaluate("namespace cmd {}")
					Expect(execute("$[cmd]([])")).To(Equal(ERROR("invalid variable name")))
				})
			})
		})
	})
})
