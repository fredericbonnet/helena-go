package helena_dialect_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"helena/core"
	. "helena/helena_dialect"
)

var _ = Describe("Helena ensembles", func() {
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

	Describe("ensemble", func() {
		Describe("Specifications", func() {
			Specify("usage", func() {
				Expect(evaluate("help ensemble")).To(Equal(
					STR("ensemble ?name? argspec body"),
				))
				Expect(evaluate("help ensemble {}")).To(Equal(
					STR("ensemble ?name? argspec body"),
				))
				Expect(evaluate("help ensemble cmd {}")).To(Equal(
					STR("ensemble ?name? argspec body"),
				))
			})

			It("should define a new command", func() {
				evaluate("ensemble cmd {} {}")
				Expect(rootScope.Context.Commands["cmd"]).NotTo(BeNil())
			})
			It("should replace existing commands", func() {
				evaluate("ensemble cmd {} {}")
				Expect(execute("ensemble cmd {} {}").Code).To(Equal(core.ResultCode_OK))
			})
		})

		Describe("Exceptions", func() {
			Specify("wrong arity", func() {
				Expect(execute("ensemble a")).To(Equal(
					ERROR(`wrong # args: should be "ensemble ?name? argspec body"`),
				))
				Expect(execute("ensemble a b c d")).To(Equal(
					ERROR(`wrong # args: should be "ensemble ?name? argspec body"`),
				))
				Expect(execute("help ensemble a b c d")).To(Equal(
					ERROR(`wrong # args: should be "ensemble ?name? argspec body"`),
				))
			})
			Specify("invalid `argspec`", func() {
				Expect(execute("ensemble a {}")).To(Equal(ERROR("invalid argument list")))
			})
			Specify("variadic arguments", func() {
				Expect(execute("ensemble {?a} {}")).To(Equal(
					ERROR("ensemble arguments cannot be variadic"),
				))
				Expect(execute("ensemble {*a} {}")).To(Equal(
					ERROR("ensemble arguments cannot be variadic"),
				))
			})
			Specify("invalid `name`", func() {
				Expect(execute("ensemble [] {} {}")).To(Equal(
					ERROR("invalid command name"),
				))
			})
			Specify("non-script body", func() {
				Expect(execute("ensemble {} a")).To(Equal(ERROR("body must be a script")))
				Expect(execute("ensemble a {} b")).To(Equal(
					ERROR("body must be a script"),
				))
			})
		})

		Describe("`body`", func() {
			It("should be executed", func() {
				evaluate("closure cmd {} {let var val}")
				Expect(rootScope.Context.Constants["var"]).To(BeNil())
				evaluate("ensemble {} {cmd}")
				Expect(rootScope.Context.Constants["var"]).To(Equal(STR("val")))
			})
			It("should access global commands", func() {
				Expect(execute("ensemble {} {idem val}").Code).To(Equal(core.ResultCode_OK))
			})
			It("should not access global variables", func() {
				evaluate("set var val")
				Expect(execute("ensemble {} {get var}").Code).To(Equal(core.ResultCode_ERROR))
			})
			It("should not set global variables", func() {
				evaluate("set var val")
				evaluate("ensemble {} {set var val2; let cst val3}")
				Expect(rootScope.Context.Variables["var"]).To(Equal(STR("val")))
				Expect(rootScope.Context.Constants["cst"]).To(BeNil())
			})
			It("should set ensemble variables", func() {
				evaluate("set var val")
				evaluate("ensemble cmd {} {set var val2; let cst val3}")
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
						Expect(execute("ensemble {} {cmd1; return; cmd2}").Code).To(Equal(
							core.ResultCode_OK,
						))
						Expect(evaluate("get var")).To(Equal(STR("val1")))
					})
					It("should still define the named command", func() {
						evaluate("ensemble cmd {} {return}")
						Expect(rootScope.Context.Commands["cmd"]).NotTo(BeNil())
					})
					It("should return passed value instead of the command object", func() {
						Expect(execute("ensemble {} {return val}")).To(Equal(OK(STR("val"))))
					})
				})
				Describe("`tailcall`", func() {
					It("should interrupt the body with `OK` code", func() {
						evaluate("closure cmd1 {} {set var val1}")
						evaluate("closure cmd2 {} {set var val2}")
						Expect(
							execute("ensemble {} {cmd1; tailcall {}; cmd2}").Code,
						).To(Equal(core.ResultCode_OK))
						Expect(evaluate("get var")).To(Equal(STR("val1")))
					})
					It("should still define the named command", func() {
						evaluate("ensemble cmd {} {tailcall {}}")
						Expect(rootScope.Context.Commands["cmd"]).NotTo(BeNil())
					})
					It("should return passed value instead of the command object", func() {
						Expect(execute("ensemble {} {tailcall {idem val}}")).To(Equal(
							OK(STR("val")),
						))
					})
				})
				Describe("`yield`", func() {
					It("should interrupt the body with `YIELD` code", func() {
						evaluate("closure cmd1 {} {set var val1}")
						evaluate("closure cmd2 {} {set var val2}")
						Expect(execute("ensemble cmd {} {cmd1; yield; cmd2}").Code).To(Equal(
							core.ResultCode_YIELD,
						))
						Expect(evaluate("get var")).To(Equal(STR("val1")))
					})
					It("should provide a resumable state", func() {
						evaluate("closure cmd1 {} {set var val1}")
						evaluate("closure cmd2 {val} {set var $val}")
						process := rootScope.PrepareScript(
							*parse("ensemble cmd {} {cmd1; cmd2 _[yield val2]_}"),
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
					It("should delay the definition of ensemble command until resumed", func() {
						process := rootScope.PrepareScript(
							*parse("ensemble cmd {} {yield}"),
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
						Expect(execute("ensemble {} {cmd1; error msg; cmd2}")).To(Equal(
							ERROR("msg"),
						))
						Expect(evaluate("get var")).To(Equal(STR("val1")))
					})
					It("should not define the ensemble command", func() {
						evaluate("ensemble cmd {} {error msg}")
						Expect(rootScope.Context.Commands["cmd"]).To(BeNil())
					})
				})
				Describe("`break`", func() {
					It("should interrupt the body with `ERROR` code", func() {
						evaluate("closure cmd1 {} {set var val1}")
						evaluate("closure cmd2 {} {set var val2}")
						Expect(execute("ensemble {} {cmd1; break; cmd2}")).To(Equal(
							ERROR("unexpected break"),
						))
						Expect(evaluate("get var")).To(Equal(STR("val1")))
					})
					It("should not define the ensemble command", func() {
						evaluate("ensemble cmd {} {break}")
						Expect(rootScope.Context.Commands["cmd"]).To(BeNil())
					})
				})
				Describe("`continue`", func() {
					It("should interrupt the body with `ERROR` code", func() {
						evaluate("closure cmd1 {} {set var val1}")
						evaluate("closure cmd2 {} {set var val2}")
						Expect(execute("ensemble {} {cmd1; continue; cmd2}")).To(Equal(
							ERROR("unexpected continue"),
						))
						Expect(evaluate("get var")).To(Equal(STR("val1")))
					})
					It("should not define the ensemble command", func() {
						evaluate("ensemble cmd {} {continue}")
						Expect(rootScope.Context.Commands["cmd"]).To(BeNil())
					})
				})
			})
		})

		Describe("Metacommand", func() {
			It("should return a metacommand", func() {
				Expect(evaluate("ensemble {} {}").Type()).To(Equal(core.ValueType_COMMAND))
				Expect(evaluate("ensemble cmd {} {}").Type()).To(Equal(core.ValueType_COMMAND))
			})
			Specify("the metacommand should return itself", func() {
				value := evaluate("set cmd [ensemble {} {}]")
				Expect(evaluate("$cmd")).To(Equal(value))
			})

			Describe("Subcommands", func() {
				Describe("`subcommands`", func() {
					It("should return list of subcommands", func() {
						Expect(evaluate("[ensemble {} {}] subcommands")).To(Equal(
							evaluate("list (subcommands eval call argspec)"),
						))
					})

					Describe("Exceptions", func() {
						Specify("wrong arity", func() {
							Expect(execute("[ensemble {} {}] subcommands a")).To(Equal(
								ERROR(`wrong # args: should be "<ensemble> subcommands"`),
							))
						})
					})
				})

				Describe("`eval`", func() {
					It("should evaluate body in ensemble scope", func() {
						evaluate("ensemble cmd {} {let cst val}")
						Expect(evaluate("[cmd] eval {get cst}")).To(Equal(STR("val")))
					})
					It("should accept tuple bodies", func() {
						evaluate("ensemble cmd {} {let cst val}")
						Expect(evaluate("[cmd] eval (get cst)")).To(Equal(STR("val")))
					})
					It("should evaluate macros in ensemble scope", func() {
						evaluate("ensemble cmd {} {macro mac {} {let cst val}}")
						evaluate("[cmd] eval {mac}")
						Expect(rootScope.Context.Constants["cst"]).To(BeNil())
						Expect(evaluate("[cmd] eval {get cst}")).To(Equal(STR("val")))
					})
					It("should evaluate closures in their scope", func() {
						evaluate("closure cls {} {let cst val}")
						evaluate("ensemble cmd {} {}")
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
								evaluate("ensemble cmd {} {}")
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
								evaluate("ensemble cmd {} {}")
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
								evaluate("ensemble cmd {} {}")
								Expect(execute("[cmd] eval {cmd1; yield; cmd2}").Code).To(Equal(
									core.ResultCode_YIELD,
								))
								Expect(evaluate("get var")).To(Equal(STR("val1")))
							})
							It("should provide a resumable state", func() {
								evaluate("closure cmd1 {} {set var val1}")
								evaluate("closure cmd2 {val} {set var $val}")
								evaluate("ensemble cmd {} {}")
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
								evaluate("ensemble cmd {} {}")
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
								evaluate("ensemble cmd {} {}")
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
								evaluate("ensemble cmd {} {}")
								Expect(execute("[cmd] eval {cmd1; continue; cmd2}")).To(Equal(
									CONTINUE(NIL),
								))
								Expect(evaluate("get var")).To(Equal(STR("val1")))
							})
						})
					})

					Describe("Exceptions", func() {
						Specify("wrong arity", func() {
							Expect(execute("[ensemble {} {}] eval")).To(Equal(
								ERROR(`wrong # args: should be "<ensemble> eval body"`),
							))
							Expect(execute("[ensemble {} {}] eval a b")).To(Equal(
								ERROR(`wrong # args: should be "<ensemble> eval body"`),
							))
						})
						Specify("invalid body", func() {
							Expect(execute("[ensemble {} {}] eval 1")).To(Equal(
								ERROR("body must be a script or tuple"),
							))
						})
					})
				})

				Describe("`call`", func() {
					It("should call ensemble commands", func() {
						evaluate("ensemble cmd {} {macro mac {} {idem val}}")
						Expect(evaluate("[cmd] call mac")).To(Equal(STR("val")))
					})
					It("should evaluate macros in the caller scope", func() {
						evaluate("ensemble cmd {} {macro mac {} {let cst val}}")
						evaluate("[cmd] call mac")
						Expect(rootScope.Context.Constants["cst"]).To(Equal(STR("val")))
						evaluate("scope scp {[cmd] call mac}")
						Expect(evaluate("[scp] eval {get cst}")).To(Equal(STR("val")))
					})
					It("should evaluate ensemble closures in ensemble scope", func() {
						evaluate("ensemble cmd {} {closure cls {} {let cst val}}")
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
									"ensemble cmd {} {macro mac {} {cmd1; return val3; cmd2}}",
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
									"ensemble cmd {} {macro mac {} {cmd1; tailcall {idem val3}; cmd2}}",
								)
								Expect(execute("[cmd] call mac")).To(Equal(RETURN(STR("val3"))))
								Expect(evaluate("get var")).To(Equal(STR("val1")))
							})
						})
						Describe("`yield`", func() {
							It("should interrupt the call with `YIELD` code", func() {
								evaluate("closure cmd1 {} {set var val1}")
								evaluate("closure cmd2 {} {set var val2}")
								evaluate("ensemble cmd {} {macro mac {} {cmd1; yield; cmd2}}")
								Expect(execute("[cmd] call mac").Code).To(Equal(core.ResultCode_YIELD))
								Expect(evaluate("get var")).To(Equal(STR("val1")))
							})
							It("should provide a resumable state", func() {
								evaluate("closure cmd1 {} {set var val1}")
								evaluate("closure cmd2 {val} {set var $val}")
								evaluate(
									"ensemble cmd {} {proc p {} {cmd1; cmd2 _[yield val2]_}}",
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
									"ensemble cmd {} {macro mac {} {cmd1; error msg; cmd2}}",
								)
								Expect(execute("[cmd] call mac")).To(Equal(ERROR("msg")))
								Expect(evaluate("get var")).To(Equal(STR("val1")))
							})
						})
						Describe("`break`", func() {
							It("should interrupt the body with `BREAK` code", func() {
								evaluate("closure cmd1 {} {set var val1}")
								evaluate("closure cmd2 {} {set var val2}")
								evaluate("ensemble cmd {} {macro mac {} {cmd1; break; cmd2}}")
								Expect(execute("[cmd] call mac")).To(Equal(BREAK(NIL)))
								Expect(evaluate("get var")).To(Equal(STR("val1")))
							})
						})
						Describe("`continue`", func() {
							It("should interrupt the body with `CONTINUE` code", func() {
								evaluate("closure cmd1 {} {set var val1}")
								evaluate("closure cmd2 {} {set var val2}")
								evaluate(
									"ensemble cmd {} {macro mac {} {cmd1; continue; cmd2}}",
								)
								Expect(execute("[cmd] call mac")).To(Equal(CONTINUE(NIL)))
								Expect(evaluate("get var")).To(Equal(STR("val1")))
							})
						})
					})

					Describe("Exceptions", func() {
						Specify("wrong arity", func() {
							Expect(execute("[ensemble {} {}] call")).To(Equal(
								ERROR(
									`wrong # args: should be "<ensemble> call cmdname ?arg ...?"`,
								),
							))
						})
						Specify("unknown command", func() {
							Expect(execute("[ensemble {} {}] call unknownCommand")).To(Equal(
								ERROR(`unknown command "unknownCommand"`),
							))
						})
						Specify("out-of-scope command", func() {
							Expect(
								execute("macro cmd {} {}; [ensemble {} {}] call cmd"),
							).To(Equal(ERROR(`unknown command "cmd"`)))
						})
						Specify("invalid command name", func() {
							Expect(execute("[ensemble {} {}] call []")).To(Equal(
								ERROR("invalid command name"),
							))
						})
					})
				})

				Describe("`argspec`", func() {
					It("should return the ensemble's argspec", func() {
						example(exampleSpec{
							script: `
								[ensemble {a b} {}] argspec
							`,
							result: evaluate("argspec {a b}"),
						})
					})

					Describe("Exceptions", func() {
						Specify("wrong arity", func() {
							Expect(execute("[ensemble {} {}] argspec a")).To(Equal(
								ERROR(`wrong # args: should be "<ensemble> argspec"`),
							))
						})
					})
				})

				Describe("Exceptions", func() {
					Specify("unknown subcommand", func() {
						Expect(execute("[ensemble {} {}] unknownSubcommand")).To(Equal(
							ERROR(`unknown subcommand "unknownSubcommand"`),
						))
					})
					Specify("invalid subcommand name", func() {
						Expect(execute("[ensemble {} {}] []")).To(Equal(
							ERROR("invalid subcommand name"),
						))
					})
				})
			})
		})
	})

	Describe("Ensemble commands", func() {
		Describe("Specifications", func() {
			It("should return its ensemble metacommand when called with no argument", func() {
				value := evaluate("ensemble cmd {} {}")
				Expect(evaluate("cmd")).To(Equal(value))
			})
			It("should return the provided arguments tuple when called with no subcommand", func() {
				evaluate("ensemble cmd {a b} {macro opt {a b} {idem val}}")
				Expect(evaluate("cmd foo bar")).To(Equal(TUPLE([]core.Value{STR("foo"), STR("bar")})))
			})
			It("should evaluate argument guards", func() {
				evaluate("ensemble cmd {(int a) (list b)} {}")
				Expect(evaluate("cmd 1 (foo bar)")).To(Equal(
					TUPLE([]core.Value{INT(1), LIST([]core.Value{STR("foo"), STR("bar")})}),
				))
			})
		})

		Describe("Exceptions", func() {
			Specify("wrong arity", func() {
				evaluate("ensemble cmd {a b} {}")
				Expect(execute("cmd a")).To(Equal(
					ERROR(`wrong # args: should be "cmd a b ?subcommand? ?arg ...?"`),
				))
			})
			Specify("failed guards", func() {
				evaluate("ensemble cmd {(int a) (list b)} {}")
				Expect(execute("cmd a ()")).To(Equal(ERROR(`invalid integer "a"`)))
				Expect(execute("cmd 1 a")).To(Equal(ERROR("invalid list")))
			})
		})

		Describe("Ensemble subcommands", func() {
			Specify("first argument after ensemble arguments should be ensemble subcommand name", func() {
				evaluate("ensemble cmd {a b} {macro opt {a b} {idem val}}")
				Expect(evaluate("cmd foo bar opt")).To(Equal(STR("val")))
			})
			It("should pass ensemble arguments to ensemble subcommand", func() {
				evaluate("ensemble cmd {a b} {macro opt {a b} {idem $a$b}}")
				Expect(evaluate("cmd foo bar opt")).To(Equal(STR("foobar")))
			})
			It("should apply guards to passed ensemble arguments", func() {
				evaluate(
					"ensemble cmd {(int a) (list b)} {macro opt {a b} {idem ($a $b)}}",
				)
				Expect(evaluate("cmd 1 (foo bar) opt")).To(Equal(
					TUPLE([]core.Value{INT(1), LIST([]core.Value{STR("foo"), STR("bar")})}),
				))
			})
			It("should pass remaining arguments to ensemble subcommand", func() {
				evaluate("ensemble cmd {a b} {macro opt {a b c d} {idem $a$b$c$d}}")
				Expect(evaluate("cmd foo bar opt baz sprong")).To(Equal(
					STR("foobarbazsprong"),
				))
			})
			It("should evaluate subcommand in the caller scope", func() {
				evaluate("ensemble cmd {} {macro mac {} {let cst val}}")
				evaluate("cmd mac")
				Expect(rootScope.Context.Constants["cst"]).To(Equal(STR("val")))
				evaluate("scope scp {cmd mac}")
				Expect(evaluate("[scp] eval {get cst}")).To(Equal(STR("val")))
			})
			It("should work recursively", func() {
				evaluate(
					"ensemble en1 {a b} {ensemble en2 {a b c d} {macro opt {a b c d e f} {idem $a$b$c$d$e$f}}}",
				)
				Expect(evaluate("en1 foo bar en2 baz sprong opt val1 val2")).To(Equal(
					STR("foobarbazsprongval1val2"),
				))
			})

			Describe("Introspection", func() {
				Describe("`subcommands`", func() {
					BeforeEach(func() {
						evaluate("ensemble cmd1 {} {}")
						evaluate("ensemble cmd2 {a b} {}")
					})
					It("should return list of subcommands", func() {
						Expect(evaluate("cmd1 subcommands")).To(Equal(
							evaluate("list (subcommands)"),
						))
						evaluate("[cmd1] eval {macro mac1 {} {}}")
						Expect(evaluate("cmd1 subcommands")).To(Equal(
							evaluate("list (subcommands mac1)"),
						))

						Expect(evaluate("cmd2 a b subcommands")).To(Equal(
							evaluate("list (subcommands)"),
						))
						evaluate("[cmd2] eval {macro mac2 {} {}}")
						Expect(evaluate("cmd2 a b subcommands")).To(Equal(
							evaluate("list (subcommands mac2)"),
						))
					})

					Describe("Exceptions", func() {
						Specify("wrong arity", func() {
							Expect(execute("cmd1 subcommands a")).To(Equal(
								ERROR(`wrong # args: should be "cmd1 subcommands"`),
							))
							Expect(execute("help cmd1 subcommands a")).To(Equal(
								ERROR(`wrong # args: should be "cmd1 subcommands"`),
							))
							Expect(execute("cmd2 a b subcommands c")).To(Equal(
								ERROR(`wrong # args: should be "cmd2 a b subcommands"`),
							))
							Expect(execute("help cmd2 a b subcommands c")).To(Equal(
								ERROR(`wrong # args: should be "cmd2 a b subcommands"`),
							))
						})
					})
				})
			})

			Describe("Help", func() {
				It("should provide subcommand help", func() {
					evaluate(`
						ensemble cmd {a} {
							macro opt1 {a b} {}
							closure opt2 {c d} {}
						}
					`)
					Expect(evaluate("help cmd")).To(Equal(
						STR("cmd a ?subcommand? ?arg ...?"),
					))
					Expect(evaluate("help cmd 1")).To(Equal(
						STR("cmd a ?subcommand? ?arg ...?"),
					))
					Expect(evaluate("help cmd 1 subcommands")).To(Equal(
						STR("cmd a subcommands"),
					))
					Expect(evaluate("help cmd 1 opt1")).To(Equal(STR("cmd a opt1 b")))
					Expect(evaluate("help cmd 2 opt1 3")).To(Equal(STR("cmd a opt1 b")))
					Expect(evaluate("help cmd 4 opt2")).To(Equal(STR("cmd a opt2 d")))
					Expect(evaluate("help cmd 5 opt2 6")).To(Equal(STR("cmd a opt2 d")))
				})
				It("should work recursively", func() {
					evaluate(`
						ensemble cmd {a} {
							ensemble sub {a b} {
								macro opt {a b c} {}
							}
						}
					`)
					Expect(evaluate("help cmd 1 sub")).To(Equal(
						STR("cmd a sub b ?subcommand? ?arg ...?"),
					))
					Expect(evaluate("help cmd 1 sub 2")).To(Equal(
						STR("cmd a sub b ?subcommand? ?arg ...?"),
					))
					Expect(evaluate("help cmd 1 sub 2 subcommands")).To(Equal(
						STR("cmd a sub b subcommands"),
					))
					Expect(evaluate("help cmd 1 sub 2 opt")).To(Equal(
						STR("cmd a sub b opt c"),
					))
					Expect(evaluate("help cmd 1 sub 2 opt 3")).To(Equal(
						STR("cmd a sub b opt c"),
					))
				})

				Describe("Exceptions", func() {
					Specify("wrong arity", func() {
						evaluate(`
							ensemble cmd {a} {
								macro opt {a b} {}
								ensemble sub {a b} {
									macro opt {a b c} {}
								}
							}
						`)
						Expect(execute("help cmd 1 subcommands 2")).To(Equal(
							ERROR(`wrong # args: should be "cmd a subcommands"`),
						))
						Expect(execute("help cmd 1 opt 2 3")).To(Equal(
							ERROR(`wrong # args: should be "cmd a opt b"`),
						))
						Expect(execute("help cmd 1 sub 2 subcommands 3")).To(Equal(
							ERROR(`wrong # args: should be "cmd a sub b subcommands"`),
						))
						Expect(execute("help cmd 1 sub 2 opt 3 4")).To(Equal(
							ERROR(`wrong # args: should be "cmd a sub b opt c"`),
						))
					})
					Specify("invalid `subcommand`", func() {
						evaluate("ensemble cmd {a} {}")
						Expect(execute("help cmd 1 []")).To(Equal(
							ERROR("invalid subcommand name"),
						))
					})
					Specify("unknown subcommand", func() {
						evaluate("ensemble cmd {a} {}")
						Expect(execute("help cmd 1 unknownSubcommand")).To(Equal(
							ERROR(`unknown subcommand "unknownSubcommand"`),
						))
					})
					Specify("subcommand with no help", func() {
						rootScope.RegisterNamedCommand("foo", simpleCommand{
							func(_ []core.Value, _ any) core.Result { return OK(NIL) },
						},
						)
						evaluate("ensemble cmd {a} {alias opt foo}")
						Expect(execute("help cmd 1 opt")).To(Equal(
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
						evaluate(
							"ensemble cmd {} {macro mac {} {cmd1; return val3; cmd2}}",
						)
						Expect(execute("cmd mac")).To(Equal(RETURN(STR("val3"))))
						Expect(evaluate("get var")).To(Equal(STR("val1")))
					})
				})
				Describe("`tailcall`", func() {
					It("should interrupt the call with `RETURN` code", func() {
						evaluate("closure cmd1 {} {set var val1}")
						evaluate("closure cmd2 {} {set var val2}")
						evaluate(
							"ensemble cmd {} {macro mac {} {cmd1; tailcall {idem val3}; cmd2}}",
						)
						Expect(execute("cmd mac")).To(Equal(RETURN(STR("val3"))))
						Expect(evaluate("get var")).To(Equal(STR("val1")))
					})
				})
				Describe("`yield`", func() {
					It("should interrupt the call with `YIELD` code", func() {
						evaluate("closure cmd1 {} {set var val1}")
						evaluate("closure cmd2 {} {set var val2}")
						evaluate("ensemble cmd {} {macro mac {} {cmd1; yield; cmd2}}")
						Expect(execute("cmd mac").Code).To(Equal(core.ResultCode_YIELD))
						Expect(evaluate("get var")).To(Equal(STR("val1")))
					})
					It("should provide a resumable state", func() {
						evaluate("closure cmd1 {} {set var val1}")
						evaluate("closure cmd2 {val} {set var $val}")
						evaluate("ensemble cmd {} {proc p {} {cmd1; cmd2 _[yield val2]_}}")
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
						evaluate("ensemble cmd {} {macro mac {} {cmd1; error msg; cmd2}}")
						Expect(execute("cmd mac")).To(Equal(ERROR("msg")))
						Expect(evaluate("get var")).To(Equal(STR("val1")))
					})
				})
				Describe("`break`", func() {
					It("should interrupt the call with `BREAK` code", func() {
						evaluate("closure cmd1 {} {set var val1}")
						evaluate("closure cmd2 {} {set var val2}")
						evaluate("ensemble cmd {} {macro mac {} {cmd1; break; cmd2}}")
						Expect(execute("cmd mac")).To(Equal(BREAK(NIL)))
						Expect(evaluate("get var")).To(Equal(STR("val1")))
					})
				})
				Describe("`continue`", func() {
					It("should interrupt the call with `CONTINUE` code", func() {
						evaluate("closure cmd1 {} {set var val1}")
						evaluate("closure cmd2 {} {set var val2}")
						evaluate("ensemble cmd {} {macro mac {} {cmd1; continue; cmd2}}")
						Expect(execute("cmd mac")).To(Equal(CONTINUE(NIL)))
						Expect(evaluate("get var")).To(Equal(STR("val1")))
					})
				})
			})

			Describe("Exceptions", func() {
				Specify("unknown subcommand", func() {
					evaluate("ensemble cmd {} {}")
					Expect(execute("cmd unknownCommand")).To(Equal(
						ERROR(`unknown subcommand "unknownCommand"`),
					))
				})
				Specify("out-of-scope subcommand", func() {
					evaluate("macro mac {} {}; ensemble cmd {} {}")
					Expect(execute("cmd mac")).To(Equal(ERROR(`unknown subcommand "mac"`)))
				})
				Specify("invalid subcommand name", func() {
					evaluate("ensemble cmd {} {}")
					Expect(execute("cmd []")).To(Equal(ERROR("invalid subcommand name")))
				})
			})
		})
	})
})
