package helena_dialect_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"helena/core"
	. "helena/helena_dialect"
)

var _ = Describe("Helena coroutines", func() {
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

	BeforeEach(init)

	Describe("coroutine", func() {
		Describe("Specifications", func() {
			Specify("usage", func() {
				Expect(evaluate("help coroutine")).To(Equal(STR("coroutine body")))
				Expect(evaluate("help coroutine body")).To(Equal(STR("coroutine body")))
			})

			It("should return a coroutine object", func() {
				Expect(evaluate("coroutine {}").Type()).To(Equal(core.ValueType_COMMAND))
			})
		})

		Describe("Exceptions", func() {
			Specify("wrong arity", func() {
				Expect(execute("coroutine")).To(Equal(
					ERROR(`wrong # args: should be "coroutine body"`),
				))
				Expect(execute("coroutine a b")).To(Equal(
					ERROR(`wrong # args: should be "coroutine body"`),
				))
				Expect(execute("help coroutine a b")).To(Equal(
					ERROR(`wrong # args: should be "coroutine body"`),
				))
			})
			Specify("non-script body", func() {
				Expect(execute("coroutine a")).To(Equal(ERROR("body must be a script")))
			})
		})

		Describe("`body`", func() {
			It("should access scope variables", func() {
				evaluate("set var val")
				Expect(evaluate("[coroutine {get var}] wait")).To(Equal(STR("val")))
			})
			It("should set scope variables", func() {
				evaluate("set var old")
				evaluate("[coroutine {set var val; set var2 val2}] wait")
				Expect(evaluate("get var")).To(Equal(STR("val")))
				Expect(evaluate("get var2")).To(Equal(STR("val2")))
			})
			It("should access scope commands", func() {
				evaluate("macro cmd2 {} {set var val}")
				evaluate("macro cmd {} {cmd2}")
				evaluate("[coroutine {cmd}] wait")
				Expect(evaluate("get var")).To(Equal(STR("val")))
			})

			Describe("Control flow", func() {
				Describe("`return`", func() {
					It("should interrupt the body with `OK` code", func() {
						Expect(
							execute("[coroutine {set var val1; return; set var val2}] wait").Code,
						).To(Equal(core.ResultCode_OK))
						Expect(evaluate("get var")).To(Equal(STR("val1")))
					})
					It("should return passed value", func() {
						Expect(execute("[coroutine {return val}] wait")).To(Equal(
							OK(STR("val")),
						))
					})
				})
				Describe("`tailcall`", func() {
					It("should interrupt the body with `OK` code", func() {
						Expect(
							execute(
								"[coroutine {set var val1; tailcall {}; set var val2}] wait",
							).Code,
						).To(Equal(core.ResultCode_OK))
						Expect(evaluate("get var")).To(Equal(STR("val1")))
					})
					It("should return passed value", func() {
						Expect(execute("[coroutine {tailcall {idem val}}] wait")).To(Equal(
							OK(STR("val")),
						))
					})
				})
				Describe("`yield`", func() {
					It("should interrupt the body with `OK` code", func() {
						Expect(
							execute("[coroutine {set var val1; return; set var val2}] wait").Code,
						).To(Equal(core.ResultCode_OK))
						Expect(evaluate("get var")).To(Equal(STR("val1")))
					})
					It("should return yielded value", func() {
						Expect(execute("[coroutine {yield val}] wait")).To(Equal(
							OK(STR("val")),
						))
					})
					It("should work recursively", func() {
						evaluate("macro cmd1 {} {yield [cmd2]; idem val5}")
						evaluate("macro cmd2 {} {yield [cmd3]; idem [cmd4]}")
						evaluate("macro cmd3 {} {yield val1; idem val2}")
						evaluate("macro cmd4 {} {yield val3; idem val4}")
						evaluate("set cr [coroutine {cmd1}]")
						Expect(execute("$cr wait")).To(Equal(OK(STR("val1"))))
						Expect(execute("$cr done")).To(Equal(OK(FALSE)))
						Expect(execute("$cr wait")).To(Equal(OK(STR("val2"))))
						Expect(execute("$cr done")).To(Equal(OK(FALSE)))
						Expect(execute("$cr wait")).To(Equal(OK(STR("val3"))))
						Expect(execute("$cr done")).To(Equal(OK(FALSE)))
						Expect(execute("$cr wait")).To(Equal(OK(STR("val4"))))
						Expect(execute("$cr done")).To(Equal(OK(FALSE)))
						Expect(execute("$cr wait")).To(Equal(OK(STR("val5"))))
						Expect(execute("$cr done")).To(Equal(OK(TRUE)))
					})
				})
				Describe("`error`", func() {
					It("should interrupt the body with `ERROR` code", func() {
						Expect(
							execute(
								"[coroutine {set var val1; error msg; set var val2}] wait",
							),
						).To(Equal(ERROR("msg")))
						Expect(evaluate("get var")).To(Equal(STR("val1")))
					})
				})
				Describe("`break`", func() {
					It("should interrupt the body with `ERROR` code", func() {
						Expect(
							execute("[coroutine {set var val1; break; set var val2}] wait"),
						).To(Equal(ERROR("unexpected break")))
						Expect(evaluate("get var")).To(Equal(STR("val1")))
					})
				})
				Describe("`continue`", func() {
					It("should interrupt the body with `ERROR` code", func() {
						Expect(
							execute("[coroutine {set var val1; continue; set var val2}] wait"),
						).To(Equal(ERROR("unexpected continue")))
						Expect(evaluate("get var")).To(Equal(STR("val1")))
					})
				})
			})
		})

		Describe("Coroutine object", func() {
			Specify("the coroutine object should return itself", func() {
				value := evaluate("set cr [coroutine {}]")
				Expect(evaluate("$cr")).To(Equal(value))
			})

			Describe("Subcommands", func() {
				Describe("`subcommands`", func() {
					It("should return list of subcommands", func() {
						Expect(evaluate("[coroutine {}] subcommands")).To(Equal(
							evaluate("list (subcommands wait active done yield)"),
						))
					})

					Describe("Exceptions", func() {
						Specify("wrong arity", func() {
							Expect(execute("[coroutine {}] subcommands a")).To(Equal(
								ERROR(`wrong # args: should be "<coroutine> subcommands"`),
							))
						})
					})
				})
				Describe("`wait`", func() {
					It("should evaluate body", func() {
						evaluate("set cr [coroutine {idem val}]")
						Expect(evaluate("$cr wait")).To(Equal(STR("val")))
					})
					It("should resume yielded body", func() {
						evaluate("set cr [coroutine {yield val1; idem val2}]")
						Expect(evaluate("$cr wait")).To(Equal(STR("val1")))
						Expect(evaluate("$cr wait")).To(Equal(STR("val2")))
					})
					It("should return result of completed coroutines", func() {
						evaluate("set cr [coroutine {idem val}]; $cr wait")
						Expect(evaluate("$cr wait")).To(Equal(STR("val")))
					})

					Describe("Exceptions", func() {
						Specify("wrong arity", func() {
							Expect(execute("[coroutine {}] wait a")).To(Equal(
								ERROR(`wrong # args: should be "<coroutine> wait"`),
							))
						})
					})
				})
				Describe("`active`", func() {
					It("should return `false` on new coroutines", func() {
						evaluate("set cr [coroutine {}]")
						Expect(evaluate("$cr active")).To(Equal(FALSE))
					})
					It("should return `false` on completed coroutines", func() {
						evaluate("set cr [coroutine {}]")
						evaluate("$cr wait")
						Expect(evaluate("$cr active")).To(Equal(FALSE))
					})
					It("should return `true` on yielded coroutines", func() {
						evaluate("set cr [coroutine {yield}]")
						evaluate("$cr wait")
						Expect(evaluate("$cr active")).To(Equal(TRUE))
					})
					It("should return `false` on yielded coroutines ran to completion", func() {
						evaluate("set cr [coroutine {yield}]")
						evaluate("$cr wait")
						evaluate("$cr wait")
						Expect(evaluate("$cr active")).To(Equal(FALSE))
					})

					Describe("Exceptions", func() {
						Specify("wrong arity", func() {
							Expect(execute("[coroutine {}] active a")).To(Equal(
								ERROR(`wrong # args: should be "<coroutine> active"`),
							))
						})
					})
				})
				Describe("`done`", func() {
					It("should return `false` on new coroutines", func() {
						evaluate("set cr [coroutine {}]")
						Expect(evaluate("$cr done")).To(Equal(FALSE))
					})
					It("should return `true` on completed coroutines", func() {
						evaluate("set cr [coroutine {}]")
						evaluate("$cr wait")
						Expect(evaluate("$cr done")).To(Equal(TRUE))
					})
					It("should return `false` on yielded coroutines", func() {
						evaluate("set cr [coroutine {yield}]")
						evaluate("$cr wait")
						Expect(evaluate("$cr done")).To(Equal(FALSE))
					})
					It("should return `true` on yielded coroutines ran to completion", func() {
						evaluate("set cr [coroutine {yield}]")
						evaluate("$cr wait")
						evaluate("$cr wait")
						Expect(evaluate("$cr done")).To(Equal(TRUE))
					})

					Describe("Exceptions", func() {
						Specify("wrong arity", func() {
							Expect(execute("[coroutine {}] done a")).To(Equal(
								ERROR(`wrong # args: should be "<coroutine> done"`),
							))
						})
					})
				})
				Describe("`yield`", func() {
					It("should resume yielded body", func() {
						evaluate("set cr [coroutine {set var val1; yield; set var val2}]")
						evaluate("$cr wait")
						Expect(evaluate("get var")).To(Equal(STR("val1")))
						Expect(evaluate("$cr yield")).To(Equal(STR("val2")))
						Expect(evaluate("get var")).To(Equal(STR("val2")))
					})
					It("should yield back value to coroutine", func() {
						evaluate("set cr [coroutine {set var [yield]}]")
						evaluate("$cr wait; $cr yield val")
						Expect(evaluate("get var")).To(Equal(STR("val")))
					})

					Describe("Exceptions", func() {
						Specify("wrong arity", func() {
							Expect(execute("[coroutine {}] yield a b")).To(Equal(
								ERROR(`wrong # args: should be "<coroutine> yield ?value?"`),
							))
						})
						Specify("inactive coroutine", func() {
							evaluate("set cr [coroutine {}]")
							Expect(execute("[coroutine {}] yield")).To(Equal(
								ERROR("coroutine is inactive"),
							))
						})
						Specify("completed coroutine", func() {
							evaluate("set cr [coroutine {}]; $cr wait")
							Expect(execute("$cr yield")).To(Equal(ERROR("coroutine is done")))
						})
					})
				})

				Describe("Exceptions", func() {
					Specify("unknown subcommand", func() {
						Expect(execute("[coroutine {}] unknownSubcommand")).To(Equal(
							ERROR(`unknown subcommand "unknownSubcommand"`),
						))
					})
					Specify("invalid subcommand name", func() {
						Expect(execute("[coroutine {}] []")).To(Equal(
							ERROR("invalid subcommand name"),
						))
					})
				})
			})
		})
	})
})
