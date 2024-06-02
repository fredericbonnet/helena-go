package helena_dialect_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"helena/core"
	. "helena/helena_dialect"
)

var _ = Describe("Helena basic commands", func() {
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
		rootScope = NewScope(nil, false)
		InitCommands(rootScope)

		tokenizer = core.Tokenizer{}
		parser = core.NewParser(nil)
	}

	BeforeEach(init)

	Describe("idem", func() {

		Describe("Specifications", func() {
			Specify("usage", func() {
				Expect(evaluate("help idem")).To(Equal(STR("idem value")))
				Expect(evaluate("help idem val")).To(Equal(STR("idem value")))
			})

			It("should return its `value` argument", func() {
				Expect(evaluate("idem val")).To(Equal(STR("val")))
				Expect(evaluate("idem (a b c)")).To(Equal(
					TUPLE([]core.Value{STR("a"), STR("b"), STR("c")}),
				))
			})
		})

		Describe("Exceptions", func() {
			Specify("wrong arity", func() {
				Expect(execute("idem")).To(Equal(
					ERROR(`wrong # args: should be "idem value"`),
				))
				Expect(execute("idem a b")).To(Equal(
					ERROR(`wrong # args: should be "idem value"`),
				))
				Expect(execute("help idem a b")).To(Equal(
					ERROR(`wrong # args: should be "idem value"`),
				))
			})
		})
	})

	Describe("return", func() {
		Describe("Specifications", func() {
			Specify("usage", func() {
				Expect(evaluate("help return")).To(Equal(STR("return ?result?")))
				Expect(evaluate("help return val")).To(Equal(STR("return ?result?")))
			})

			Specify("result code should be `RETURN`", func() {
				Expect(execute("return").Code).To(Equal(core.ResultCode_RETURN))
			})
			It("should return nil by default", func() {
				Expect(evaluate("return")).To(Equal(NIL))
			})
			It("should return its optional `result` argument", func() {
				Expect(evaluate("return val")).To(Equal(STR("val")))
			})
			It("should interrupt the script", func() {
				Expect(execute("return val; unreachable")).To(Equal(RETURN(STR("val"))))
			})
		})

		Describe("Exceptions", func() {
			Specify("wrong arity", func() {
				Expect(execute("return a b")).To(Equal(
					ERROR(`wrong # args: should be "return ?result?"`),
				))
				Expect(execute("help return a b")).To(Equal(
					ERROR(`wrong # args: should be "return ?result?"`),
				))
			})
		})
	})

	Describe("tailcall", func() {
		Describe("Specifications", func() {
			Specify("usage", func() {
				Expect(evaluate("help tailcall")).To(Equal(STR("tailcall body")))
				Expect(evaluate("help tailcall {}")).To(Equal(STR("tailcall body")))
			})

			Specify("result code should be `RETURN`", func() {
				Expect(execute("tailcall {}").Code).To(Equal(core.ResultCode_RETURN))
			})
			It("should accept script values for its `body` argument", func() {
				Expect(execute("tailcall {}")).To(Equal(RETURN(NIL)))
			})
			It("should accept tuple values for its `body` argument", func() {
				Expect(execute("tailcall ()")).To(Equal(RETURN(NIL)))
			})
			It("should return the evaluation result of it `body` argument", func() {
				Expect(execute("tailcall {idem val}")).To(Equal(RETURN(STR("val"))))
				Expect(execute("tailcall {return val}")).To(Equal(RETURN(STR("val"))))
				Expect(execute("tailcall (idem val); unreachable")).To(Equal(
					RETURN(STR("val")),
				))
				Expect(execute("tailcall (return val); unreachable")).To(Equal(
					RETURN(STR("val")),
				))
			})
			It("should propagate `ERROR` code from `body`", func() {
				Expect(execute("tailcall {error msg}")).To(Equal(ERROR("msg")))
				Expect(execute("tailcall (error msg); unreachable")).To(Equal(
					ERROR("msg"),
				))
			})
			It("should propagate `BREAK` code from `body`", func() {
				Expect(execute("tailcall {break}")).To(Equal(BREAK(NIL)))
				Expect(execute("tailcall (break); unreachable")).To(Equal(BREAK(NIL)))
			})
			It("should propagate `CONTINUE` code from `body`", func() {
				Expect(execute("tailcall {continue}")).To(Equal(CONTINUE(NIL)))
				Expect(execute("tailcall (continue); unreachable")).To(Equal(CONTINUE(NIL)))
			})
			It("should interrupt the script", func() {
				Expect(execute("tailcall {idem val}; unreachable")).To(Equal(
					RETURN(STR("val")),
				))
				Expect(execute("tailcall (idem val); unreachable")).To(Equal(
					RETURN(STR("val")),
				))
			})
			It("should work recursively", func() {
				Expect(
					execute("tailcall {tailcall (idem val); unreachable}; unreachable"),
				).To(Equal(RETURN(STR("val"))))
			})
		})

		Describe("Exceptions", func() {
			Specify("wrong arity", func() {
				Expect(execute("tailcall")).To(Equal(
					ERROR(`wrong # args: should be "tailcall body"`),
				))
				Expect(execute("tailcall a b")).To(Equal(
					ERROR(`wrong # args: should be "tailcall body"`),
				))
				Expect(execute("help tailcall a b")).To(Equal(
					ERROR(`wrong # args: should be "tailcall body"`),
				))
			})
			Specify("invalid `body`", func() {
				Expect(execute("tailcall 1")).To(Equal(
					ERROR("body must be a script or tuple"),
				))
			})
		})
	})

	Describe("yield", func() {
		Describe("Specifications", func() {
			Specify("usage", func() {
				Expect(evaluate("help yield")).To(Equal(STR("yield ?result?")))
				Expect(evaluate("help yield val")).To(Equal(STR("yield ?result?")))
			})

			Specify("result code should be `YIELD`", func() {
				Expect(execute("yield").Code).To(Equal(core.ResultCode_YIELD))
			})
			It("should yield nil by default", func() {
				Expect(evaluate("yield")).To(Equal(NIL))
			})
			It("should yield its optional `result` argument", func() {
				Expect(evaluate("yield val")).To(Equal(STR("val")))
			})
		})

		Describe("Exceptions", func() {
			Specify("wrong arity", func() {
				Expect(execute("yield a b")).To(Equal(
					ERROR(`wrong # args: should be "yield ?result?"`),
				))
				Expect(execute("help yield a b")).To(Equal(
					ERROR(`wrong # args: should be "yield ?result?"`),
				))
			})
		})
	})

	Describe("error", func() {
		Describe("Specifications", func() {
			Specify("usage", func() {
				Expect(evaluate("help error")).To(Equal(STR("error message")))
				Expect(evaluate("help error val")).To(Equal(STR("error message")))
			})

			Specify("result code should be `ERROR`", func() {
				Expect(execute("error a").Code).To(Equal(core.ResultCode_ERROR))
			})
			Specify("result value should be its `message` argument", func() {
				Expect(evaluate("error val")).To(Equal(STR("val")))
			})
		})

		Describe("Exceptions", func() {
			Specify("wrong arity", func() {
				Expect(execute("error")).To(Equal(
					ERROR(`wrong # args: should be "error message"`),
				))
				Expect(execute("error a b")).To(Equal(
					ERROR(`wrong # args: should be "error message"`),
				))
				Expect(execute("help error a b")).To(Equal(
					ERROR(`wrong # args: should be "error message"`),
				))
			})
			Specify("non-string `message`", func() {
				Expect(execute("error ()")).To(Equal(ERROR("invalid message")))
			})
		})
	})

	Describe("break", func() {
		Describe("Specifications", func() {
			Specify("usage", func() {
				Expect(evaluate("help break")).To(Equal(STR("break")))
			})

			Specify("result code should be `BREAK`", func() {
				Expect(execute("break").Code).To(Equal(core.ResultCode_BREAK))
			})
		})

		Describe("Exceptions", func() {
			Specify("wrong arity", func() {
				Expect(execute("break a")).To(Equal(
					ERROR(`wrong # args: should be "break"`),
				))
				Expect(execute("help break a")).To(Equal(
					ERROR(`wrong # args: should be "break"`),
				))
			})
		})
	})

	Describe("continue", func() {
		Describe("Specifications", func() {
			Specify("usage", func() {
				Expect(evaluate("help continue")).To(Equal(STR("continue")))
			})

			Specify("result code should be `CONTINUE`", func() {
				Expect(execute("continue").Code).To(Equal(core.ResultCode_CONTINUE))
			})
		})

		Describe("Exceptions", func() {
			Specify("wrong arity", func() {
				Expect(execute("continue a")).To(Equal(
					ERROR(`wrong # args: should be "continue"`),
				))
				Expect(execute("help continue a")).To(Equal(
					ERROR(`wrong # args: should be "continue"`),
				))
			})
		})
	})

	Describe("eval", func() {
		Describe("Specifications", func() {
			Specify("usage", func() {
				Expect(evaluate("help eval")).To(Equal(STR("eval body")))
				Expect(evaluate("help eval body")).To(Equal(STR("eval body")))
			})

			It("should return nil for empty `body`", func() {
				Expect(evaluate("eval {}")).To(Equal(NIL))
			})
			It("should return the result of the last command evaluated in `body`", func() {
				Expect(execute("eval {idem val1; idem val2}")).To(Equal(OK(STR("val2"))))
			})
			It("should evaluate `body` in the current scope", func() {
				evaluate("eval {let var val}")
				Expect(evaluate("get var")).To(Equal(STR("val")))
			})
			It("should accept tuple `body` arguments", func() {
				Expect(evaluate("eval (idem val)")).To(Equal(STR("val")))
			})
			It("should work recursively", func() {
				process := prepareScript(
					"eval {eval {yield val1}; yield val2; eval {yield val3}}",
				)

				result := process.Run()
				Expect(result.Code).To(Equal(core.ResultCode_YIELD))
				Expect(result.Value).To(Equal(STR("val1")))

				result = process.Run()
				Expect(result.Code).To(Equal(core.ResultCode_YIELD))
				Expect(result.Value).To(Equal(STR("val2")))

				result = process.Run()
				Expect(result.Code).To(Equal(core.ResultCode_YIELD))
				Expect(result.Value).To(Equal(STR("val3")))

				process.YieldBack(STR("val4"))
				result = process.Run()
				Expect(result).To(Equal(OK(STR("val4"))))
			})
		})

		Describe("Exceptions", func() {
			Specify("wrong arity", func() {
				Expect(execute("eval")).To(Equal(
					ERROR(`wrong # args: should be "eval body"`),
				))
				Expect(execute("eval a b")).To(Equal(
					ERROR(`wrong # args: should be "eval body"`),
				))
				Expect(execute("help eval a b")).To(Equal(
					ERROR(`wrong # args: should be "eval body"`),
				))
			})
			Specify("invalid `body`", func() {
				Expect(execute("eval 1")).To(Equal(
					ERROR("body must be a script or tuple"),
				))
			})
		})

		Describe("Control flow", func() {
			Describe("`return`", func() {
				It("should interrupt the body with `RETURN` code", func() {
					Expect(
						execute("eval {set var val1; return; set var val2}").Code,
					).To(Equal(core.ResultCode_RETURN))
					Expect(evaluate("get var")).To(Equal(STR("val1")))
				})
				It("should return passed value", func() {
					Expect(execute("eval {return val}")).To(Equal(RETURN(STR("val"))))
				})
			})
			Describe("`tailcall`", func() {
				It("should interrupt the body with `RETURN` code", func() {
					Expect(
						execute("eval {set var val1; tailcall {}; set var val2}").Code,
					).To(Equal(core.ResultCode_RETURN))
					Expect(evaluate("get var")).To(Equal(STR("val1")))
				})
				It("should return tailcall result", func() {
					Expect(execute("eval {tailcall {idem val}}")).To(Equal(
						RETURN(STR("val")),
					))
				})
			})
			Describe("`yield`", func() {
				It("should interrupt the body with `YIELD` code", func() {
					Expect(
						execute("eval {set var val1; yield; set var val2}").Code,
					).To(Equal(core.ResultCode_YIELD))
					Expect(evaluate("get var")).To(Equal(STR("val1")))
				})
				It("should provide a resumable state", func() {
					process := prepareScript(
						"eval {set var val1; set var _[yield val2]_}",
					)

					result := process.Run()
					Expect(result.Code).To(Equal(core.ResultCode_YIELD))
					Expect(result.Value).To(Equal(STR("val2")))
					Expect(evaluate("get var")).To(Equal(STR("val1")))

					process.YieldBack(STR("val3"))
					result = process.Run()
					Expect(result).To(Equal(OK(STR("_val3_"))))
					Expect(evaluate("get var")).To(Equal(STR("_val3_")))
				})
			})
			Describe("`error`", func() {
				It("should interrupt the body with `ERROR` code", func() {
					Expect(
						execute("eval {set var val1; error msg; set var val2}"),
					).To(Equal(ERROR("msg")))
					Expect(evaluate("get var")).To(Equal(STR("val1")))
				})
			})
			Describe("`break`", func() {
				It("should interrupt the body with `BREAK` code", func() {
					Expect(execute("eval {set var val1; break; set var val2}")).To(Equal(
						BREAK(NIL),
					))
					Expect(evaluate("get var")).To(Equal(STR("val1")))
				})
			})
			Describe("`continue`", func() {
				It("should interrupt the body with `CONTINUE` code", func() {
					Expect(execute("eval {set var val1; continue; set var val2}")).To(Equal(
						CONTINUE(NIL),
					))
					Expect(evaluate("get var")).To(Equal(STR("val1")))
				})
			})
		})
	})

	Describe("help", func() {
		Describe("Specifications", func() {
			It("should give usage of itself", func() {
				Expect(evaluate("help help")).To(Equal(STR("help command ?arg ...?")))
			})
			It("should accept optional arguments", func() {
				Expect(evaluate("help help command")).To(Equal(
					STR("help command ?arg ...?"),
				))
			})
			It("should return the command help", func() {
				command := commandWithHelp{
					execute: func(_ []core.Value, _ any) core.Result {
						return OK(NIL)
					},
					help: func(_ []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
						return OK(STR("this is a help string"))
					},
				}
				rootScope.SetNamedConstant("cmd", core.NewCommandValue(command))
				rootScope.RegisterNamedCommand("cmd", command)
				Expect(evaluate("help cmd")).To(Equal(STR("this is a help string")))
				Expect(evaluate("help $cmd")).To(Equal(STR("this is a help string")))
			})
		})

		Describe("Exceptions", func() {
			Specify("wrong arity", func() {
				Expect(execute("help")).To(Equal(
					ERROR(`wrong # args: should be "help command ?arg ...?"`),
				))
			})
			Specify("invalid `command`", func() {
				Expect(execute("help []")).To(Equal(ERROR("invalid command name")))
			})
			Specify("unknown command", func() {
				Expect(execute("help unknownCommand")).To(Equal(
					ERROR(`unknown command "unknownCommand"`),
				))
			})
			Specify("command with no help", func() {
				command := simpleCommand{
					execute: func(_ []core.Value, _ any) core.Result {
						return OK(NIL)
					},
				}
				rootScope.SetNamedConstant("cmd", core.NewCommandValue(command))
				rootScope.RegisterNamedCommand("cmd", command)
				Expect(execute("help $cmd")).To(Equal(ERROR("no help for command")))
				Expect(execute("help cmd")).To(Equal(ERROR(`no help for command "cmd"`)))
			})
		})
	})
})
