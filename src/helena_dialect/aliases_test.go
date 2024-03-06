package helena_dialect_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"helena/core"
	. "helena/helena_dialect"
)

var _ = Describe("Helena aliases", func() {
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

	//   const example = specifyExample(({ script }) => execute(script));

	BeforeEach(init)

	Describe("alias", func() {
		Describe("Specifications", func() {
			Specify("usage", func() {
				Expect(evaluate("help alias")).To(Equal(STR("alias name command")))
				Expect(evaluate("help alias cmd")).To(Equal(STR("alias name command")))
				Expect(evaluate("help alias cmd cmd2")).To(Equal(
					STR("alias name command"),
				))
			})

			It("should define a new command", func() {
				evaluate("alias cmd idem")
				Expect(rootScope.Context.Commands["cmd"]).ToNot(BeNil())
			})
			It("should replace existing commands", func() {
				evaluate("alias cmd set")
				Expect(execute("alias cmd idem").Code).To(Equal(core.ResultCode_OK))
			})
		})

		Describe("Exceptions", func() {
			Specify("wrong arity", func() {
				Expect(execute("alias a")).To(Equal(
					ERROR(`wrong # args: should be "alias name command"`),
				))
				Expect(execute("alias a b c")).To(Equal(
					ERROR(`wrong # args: should be "alias name command"`),
				))
				Expect(execute("help alias a b c")).To(Equal(
					ERROR(`wrong # args: should be "alias name command"`),
				))
			})
			Specify("invalid `name`", func() {
				Expect(execute("alias [] set")).To(Equal(ERROR("invalid command name")))
			})
		})

		Describe("Command calls", func() {
			// It("should call the aliased command", func() {
			// 	evaluate("macro mac {} {set var val}")
			// 	evaluate("alias cmd mac")
			// 	evaluate("cmd")
			// 	Expect(evaluate("get var")).To(Equal(STR("val")))
			// })
			It("should pass arguments to aliased commands", func() {
				evaluate("alias cmd (set var)")
				Expect(execute("cmd val")).To(Equal(OK(STR("val"))))
				Expect(evaluate("get var")).To(Equal(STR("val")))
			})

			Describe("Exceptions", func() {
				Specify("wrong arity", func() {
					evaluate("alias cmd (set var)")
					Expect(execute("cmd")).To(Equal(
						ERROR(`wrong # args: should be "set varname value"`),
					))
					Expect(execute("cmd 1 2")).To(Equal(
						ERROR(`wrong # args: should be "set varname value"`),
					))
				})
			})

			Describe("Command tuples", func() {
				Specify("zero", func() {
					evaluate("alias cmd ()")
					Expect(execute("cmd")).To(Equal(OK(NIL)))
					Expect(execute("cmd idem val")).To(Equal(OK(STR("val"))))
				})
				Specify("one", func() {
					evaluate("alias cmd return")
					Expect(execute("cmd")).To(Equal(RETURN(NIL)))
					Expect(execute("cmd val")).To(Equal(RETURN(STR("val"))))
				})
				Specify("two", func() {
					evaluate("alias cmd (idem val)")
					Expect(execute("cmd")).To(Equal(OK(STR("val"))))
				})
				Specify("three", func() {
					evaluate("alias cmd (set var val)")
					Expect(execute("cmd")).To(Equal(OK(STR("val"))))
					Expect(evaluate("get var")).To(Equal(STR("val")))
				})

				//   Describe("Examples", func() {
				//     example("Currying", {
				//       doc: func() {
				//         /**
				//          * Thanks to leading tuple auto-expansion, it is very simple to
				//          * create curried commands by bundling a command name and its
				//          * arguments into a tuple. Here we create a new command `double`
				//          * by currying the prefix multiplication operator `*` with 2:
				//          */
				//       },
				//       script: `
				//         alias double (* 2)
				//         double 3
				//       `,
				//       result: INT(6),
				//     });
				//     example("Encapsulation", [
				//       {
				//         doc: func() {
				//           /**
				//            * Here we create a new command `mylist` by encapsulating a
				//            * list value passed to the `list` command; we then can call
				//            * `list` subcommands without having to provide the value:
				//            */
				//         },
				//         script: `
				//           alias mylist (list (1 2 3))
				//           mylist length
				//         `,
				//         result: INT(3),
				//       },
				//       {
				//         doc: func() {
				//           /**
				//            * A nice side effect of how `list` works is that calling the
				//            * alias with no argument will return the encapsulated value:
				//            */
				//         },
				//         script: `
				//           mylist
				//         `,
				//         result: LIST([STR("1"), STR("2"), STR("3")]),
				//       },
				//     ]);
				//   });
			})

			Describe("Control flow", func() {
				Describe("`return`", func() {
					// It("should interrupt a macro alias with `RETURN` code", func() {
					// 	evaluate("macro mac {} {return val1; idem val2}")
					// 	evaluate("alias cmd mac")
					// 	Expect(execute("cmd")).To(Equal(RETURN(STR("val1"))))
					// })
					It("should interrupt a tuple alias with `RETURN` code", func() {
						evaluate("alias cmd (return val)")
						Expect(execute("cmd")).To(Equal(RETURN(STR("val"))))
					})
				})
				Describe("`tailcall`", func() {
					// It("should interrupt a macro alias with `RETURN` code", func() {
					// 	evaluate("macro mac {} {tailcall {idem val1}; idem val2}")
					// 	evaluate("alias cmd mac")
					// 	Expect(execute("cmd")).To(Equal(RETURN(STR("val1"))))
					// })
					It("should interrupt a tuple alias with `RETURN` code", func() {
						evaluate("alias cmd (tailcall {idem val})")
						Expect(execute("cmd")).To(Equal(RETURN(STR("val"))))
					})
				})
				Describe("`yield`", func() {
					// It("should interrupt a macro alias with `YIELD` code", func() {
					// 	evaluate("macro mac {} {yield val1; idem val2}")
					// 	evaluate("alias cmd mac")
					// 	result := execute("cmd")
					// 	Expect(result.Code).To(Equal(core.ResultCode_YIELD))
					// 	Expect(result.Value).To(Equal(STR("val1")))
					// })
					It("should interrupt a tuple alias with `YIELD` code", func() {
						evaluate("alias cmd (yield val1)")
						result := execute("cmd")
						Expect(result.Code).To(Equal(core.ResultCode_YIELD))
						Expect(result.Value).To(Equal(STR("val1")))
					})
					// It("should provide a resumable state for macro alias", func() {
					// 	evaluate("macro mac {} {idem _[yield val1]_}")
					// 	evaluate("alias cmd mac")
					// 	process := rootScope.PrepareScript(*parse("cmd"))

					// 	result := process.Run()
					// 	Expect(result.Code).To(Equal(core.ResultCode_YIELD))
					// 	Expect(result.Value).To(Equal(STR("val1")))

					// 	process.YieldBack(STR("val2"))
					// 	result = process.Run()
					// 	Expect(result).To(Equal(OK(STR("_val2_"))))
					// })
					It("should provide a resumable state for tuple alias", func() {
						evaluate("alias cmd (yield val1)")
						process := rootScope.PrepareScript(*parse("cmd"))

						result := process.Run()
						Expect(result.Code).To(Equal(core.ResultCode_YIELD))
						Expect(result.Value).To(Equal(STR("val1")))
						Expect(result.Data).ToNot(BeNil())

						process.YieldBack(STR("val2"))
						result = process.Run()
						Expect(result).To(Equal(OK(STR("val2"))))
					})
				})
				Describe("`error`", func() {
					// It("should interrupt a macro alias with `ERROR` code", func() {
					// 	evaluate("macro mac {} {error msg; idem val}")
					// 	evaluate("alias cmd mac")
					// 	Expect(execute("cmd")).To(Equal(ERROR("msg")))
					// })
					It("should interrupt a tuple alias with `ERROR` code", func() {
						evaluate("alias cmd (error msg)")
						Expect(execute("cmd")).To(Equal(ERROR("msg")))
					})
				})
				Describe("`break`", func() {
					// It("should interrupt a macro alias with `BREAK` code", func() {
					// 	evaluate("macro mac {} {break; idem val}")
					// 	evaluate("alias cmd mac")
					// 	Expect(execute("cmd")).To(Equal(BREAK(NIL)))
					// })
					It("should interrupt a tuple alias with `BREAK` code", func() {
						evaluate("alias cmd (break)")
						Expect(execute("cmd")).To(Equal(BREAK(NIL)))
					})
				})
				Describe("`continue`", func() {
					// It("should interrupt a macro alias with `CONTINUE` code", func() {
					// 	evaluate("macro mac {} {continue; idem val}")
					// 	evaluate("alias cmd mac")
					// 	Expect(execute("cmd")).To(Equal(CONTINUE(NIL)))
					// })
					It("should interrupt a tuple alias with `CONTINUE` code", func() {
						evaluate("alias cmd (continue)")
						Expect(execute("cmd")).To(Equal(CONTINUE(NIL)))
					})
				})
			})
		})

		Describe("Metacommand", func() {
			It("should return a metacommand", func() {
				Expect(evaluate("alias cmd idem").Type()).To(Equal(core.ValueType_COMMAND))
			})
			Specify("the metacommand should return the aliased command", func() {
				value := evaluate("set cmd [alias cmd set]")
				Expect(evaluate("$cmd").Type()).To(Equal(core.ValueType_COMMAND))
				Expect(evaluate("$cmd")).ToNot(Equal(value))
				Expect(evaluate("[$cmd] var val")).To(Equal(STR("val")))
				Expect(evaluate("get var")).To(Equal(STR("val")))
			})

			// Describe("Examples", func() {
			//   example("Calling alias through its wrapped metacommand", [
			//     {
			//       doc: func() {
			//         /**
			//          * Here we alias the command `list` and call it through the
			//          * alias metacommand:
			//          */
			//       },
			//       script: `
			//         set cmd [alias foo list]
			//         [$cmd] (1 2 3)
			//       `,
			//       result: LIST([STR("1"), STR("2"), STR("3")]),
			//     },
			//     {
			//       doc: func() {
			//         /**
			//          * This behaves the same as calling the alias directly:
			//          */
			//       },
			//       script: `
			//         foo (1 2 3)
			//       `,
			//       result: LIST([STR("1"), STR("2"), STR("3")]),
			//     },
			//   ]);
			// });

			Describe("Subcommands", func() {
				Describe("`subcommands`", func() {
					// It("should return list of subcommands", func() {
					// 	Expect(evaluate("[alias cmd idem] subcommands")).To(Equal(
					// 		evaluate("list (subcommands command)"),
					// 	))
					// })

					Describe("Exceptions", func() {
						Specify("wrong arity", func() {
							Expect(execute("[alias cmd idem] subcommands a")).To(Equal(
								ERROR(`wrong # args: should be "<alias> subcommands"`),
							))
						})
					})
				})

				Describe("`command`", func() {
					// example("should return the aliased command", {
					//   doc: func() {
					//     /**
					//      * This will return the value of the `command` argument at alias
					//      * creation time.
					//      */
					//   },
					//   script: `
					//     set cmd [alias cmd (idem val)]
					//     $cmd command
					//   `,
					//   result: TUPLE([STR("idem"), STR("val")]),
					// });

					Describe("Exceptions", func() {
						Specify("wrong arity", func() {
							Expect(execute("[alias cmd set] command a")).To(Equal(
								ERROR(`wrong # args: should be "<alias> command"`),
							))
						})
					})
				})

				Describe("Exceptions", func() {
					Specify("unknown subcommand", func() {
						Expect(execute("[alias cmd idem] unknownSubcommand")).To(Equal(
							ERROR(`unknown subcommand "unknownSubcommand"`),
						))
					})
					Specify("invalid subcommand name", func() {
						Expect(execute("[alias cmd idem] []")).To(Equal(
							ERROR("invalid subcommand name"),
						))
					})
				})
			})
		})
	})
})
