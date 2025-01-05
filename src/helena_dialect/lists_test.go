package helena_dialect_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"helena/core"
	. "helena/helena_dialect"
)

var _ = Describe("Helena lists", func() {
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

	Describe("list", func() {
		Describe("List creation and conversion", func() {
			It("should return list value", func() {
				Expect(evaluate("list ()")).To(Equal(LIST([]core.Value{})))
			})
			It("should convert tuples to lists", func() {
				example(exampleSpec{
					script: "list (a b c)",
					result: LIST([]core.Value{STR("a"), STR("b"), STR("c")}),
				})
			})
			It("should convert blocks to lists", func() {
				example(exampleSpec{
					script: "list {a b c}",
					result: LIST([]core.Value{STR("a"), STR("b"), STR("c")}),
				})
			})

			Describe("Exceptions", func() {
				Specify("invalid values", func() {
					Expect(execute("list []")).To(Equal(ERROR("invalid list")))
					Expect(execute("list [1]")).To(Equal(ERROR("invalid list")))
					Expect(execute("list a")).To(Equal(ERROR("invalid list")))
				})
				Specify("blocks with side effects", func() {
					Expect(execute("list { $a }")).To(Equal(ERROR("invalid list")))
					Expect(execute("list { [b] }")).To(Equal(ERROR("invalid list")))
					Expect(execute("list { $[][a] }")).To(Equal(ERROR("invalid list")))
					Expect(execute("list { $[](a) }")).To(Equal(ERROR("invalid list")))
				})
			})
		})

		Describe("Subcommands", func() {
			Describe("Introspection", func() {
				Describe("`subcommands`", func() {
					Specify("usage", func() {
						Expect(evaluate("help list () subcommands")).To(Equal(
							STR("list value subcommands"),
						))
					})

					It("should return list of subcommands", func() {
						// Expect(evaluate("list {} subcommands")).To(Equal(
						// 	evaluate(
						// 		TODO specify order?
						// 		"list (subcommands length at range append remove insert replace foreach)",
						// 	),
						// ))
					})
				})

				Describe("Exceptions", func() {
					Specify("wrong arity", func() {
						Expect(execute("list {} subcommands a")).To(Equal(
							ERROR(`wrong # args: should be "list value subcommands"`),
						))
						Expect(execute("help list {} subcommands a")).To(Equal(
							ERROR(`wrong # args: should be "list value subcommands"`),
						))
					})
				})
			})

			Describe("Accessors", func() {
				Describe("`length`", func() {
					Specify("usage", func() {
						Expect(evaluate("help list () length")).To(Equal(
							STR("list value length"),
						))
					})

					It("should return the list length", func() {
						Expect(evaluate("list () length")).To(Equal(INT(0)))
						Expect(evaluate("list (a b c) length")).To(Equal(INT(3)))
					})
					Describe("Exceptions", func() {
						Specify("wrong arity", func() {
							Expect(execute("list () length a")).To(Equal(
								ERROR(`wrong # args: should be "list value length"`),
							))
							Expect(execute("help list () length a")).To(Equal(
								ERROR(`wrong # args: should be "list value length"`),
							))
						})
					})
				})

				Describe("`at`", func() {
					Specify("usage", func() {
						Expect(evaluate("help list () at")).To(Equal(
							STR("list value at index ?default?"),
						))
					})

					It("should return the element at `index`", func() {
						Expect(evaluate("list (a b c) at 1")).To(Equal(STR("b")))
					})
					It("should return the default value for an out-of-range `index`", func() {
						Expect(evaluate("list (a b c) at 10 default")).To(Equal(
							STR("default"),
						))
					})
					Specify("`at` <-> indexed selector equivalence", func() {
						rootScope.SetNamedVariable(
							"v",
							LIST([]core.Value{STR("a"), STR("b"), STR("c")}),
						)
						evaluate("set l (list $v)")

						Expect(execute("list $v at 2")).To(Equal(execute("idem $v[2]")))
						Expect(execute("$l at 2")).To(Equal(execute("idem $v[2]")))
						Expect(execute("idem $[$l][2]")).To(Equal(execute("idem $v[2]")))

						Expect(execute("list $l at -1")).To(Equal(execute("idem $v[-1]")))
						Expect(execute("$l at -1")).To(Equal(execute("idem $v[-1]")))
						Expect(execute("idem $[$l][-1]")).To(Equal(execute("idem $v[-1]")))
					})

					Describe("Exceptions", func() {
						Specify("wrong arity", func() {
							Expect(execute("list (a b c) at")).To(Equal(
								ERROR(`wrong # args: should be "list value at index ?default?"`),
							))
							Expect(execute("list (a b c) at a b c")).To(Equal(
								ERROR(`wrong # args: should be "list value at index ?default?"`),
							))
							Expect(execute("help list (a b c) at a b c")).To(Equal(
								ERROR(`wrong # args: should be "list value at index ?default?"`),
							))
						})
						Specify("invalid `index`", func() {
							Expect(execute("list (a b c) at a")).To(Equal(
								ERROR(`invalid integer "a"`),
							))
						})
						Specify("`index` out of range", func() {
							Expect(execute("list (a b c) at -1")).To(Equal(
								ERROR(`index out of range "-1"`),
							))
							Expect(execute("list (a b c) at 10")).To(Equal(
								ERROR(`index out of range "10"`),
							))
						})
					})
				})
			})

			Describe("Operations", func() {
				Describe("`range`", func() {
					Specify("usage", func() {
						Expect(evaluate("help list () range")).To(Equal(
							STR("list value range first ?last?"),
						))
					})

					It("should return the list included within [`first`, `last`]", func() {
						Expect(evaluate("list (a b c d e f) range 1 3")).To(Equal(
							evaluate("list (b c d)"),
						))
					})
					It("should return the remainder of the list when given `first` only", func() {
						Expect(evaluate("list (a b c) range 2")).To(Equal(
							evaluate("list (c)"),
						))
					})
					It("should truncate out of range boundaries", func() {
						Expect(evaluate("list (a b c) range -1")).To(Equal(
							evaluate("list (a b c)"),
						))
						Expect(evaluate("list (a b c) range -10 1")).To(Equal(
							evaluate("list (a b)"),
						))
						Expect(evaluate("list (a b c) range 2 10")).To(Equal(
							evaluate("list (c)"),
						))
						Expect(evaluate("list (a b c) range -2 10")).To(Equal(
							evaluate("list (a b c)"),
						))
					})
					It("should return an empty list when `last` is before `first`", func() {
						Expect(evaluate("list (a b c) range 2 0")).To(Equal(LIST([]core.Value{})))
					})
					It("should return an empty list when `first` is past the list length", func() {
						Expect(evaluate("list (a b c) range 10 12")).To(Equal(
							evaluate("list ()"),
						))
					})
					It("should return an empty list when `last` is negative", func() {
						Expect(evaluate("list (a b c) range -3 -1")).To(Equal(
							evaluate("list ()"),
						))
					})

					Describe("Exceptions", func() {
						Specify("wrong arity", func() {
							Expect(execute("list (a b c) range")).To(Equal(
								ERROR(`wrong # args: should be "list value range first ?last?"`),
							))
							Expect(execute("list (a b c) range a b c")).To(Equal(
								ERROR(`wrong # args: should be "list value range first ?last?"`),
							))
							Expect(execute("help list (a b c) range a b c")).To(Equal(
								ERROR(`wrong # args: should be "list value range first ?last?"`),
							))
						})
						Specify("invalid index", func() {
							Expect(execute("list (a b c) range a")).To(Equal(
								ERROR(`invalid integer "a"`),
							))
							Expect(execute("list (a b c) range 1 b")).To(Equal(
								ERROR(`invalid integer "b"`),
							))
						})
					})
				})

				Describe("`remove`", func() {
					Specify("usage", func() {
						Expect(evaluate("help list () remove")).To(Equal(
							STR("list value remove first last"),
						))
					})

					It("should remove the range included within [`first`, `last`]", func() {
						Expect(evaluate("list (a b c d e f) remove 1 3")).To(Equal(
							evaluate("list (a e f)"),
						))
					})
					It("should truncate out of range boundaries", func() {
						Expect(evaluate("list (a b c) remove -10 1")).To(Equal(
							evaluate("list (c)"),
						))
						Expect(evaluate("list (a b c) remove 2 10")).To(Equal(
							evaluate("list (a b)"),
						))
						Expect(evaluate("list (a b c) remove -2 10")).To(Equal(
							evaluate("list ()"),
						))
					})
					It("should do nothing when `last` is before `first`", func() {
						Expect(evaluate("list (a b c) remove 2 0")).To(Equal(
							evaluate("list (a b c)"),
						))
					})
					It("should do nothing when `last` is negative", func() {
						Expect(evaluate("list (a b c) remove -3 -1")).To(Equal(
							evaluate("list (a b c)"),
						))
					})
					It("should do nothing when `first` is past the list length", func() {
						Expect(evaluate("list (a b c) remove 10 12")).To(Equal(
							evaluate("list (a b c)"),
						))
					})

					Describe("Exceptions", func() {
						Specify("wrong arity", func() {
							Expect(execute("list (a b c) remove a")).To(Equal(
								ERROR(`wrong # args: should be "list value remove first last"`),
							))
							Expect(execute("list (a b c) remove a b c d")).To(Equal(
								ERROR(`wrong # args: should be "list value remove first last"`),
							))
							Expect(execute("help list (a b c) remove a b c d")).To(Equal(
								ERROR(`wrong # args: should be "list value remove first last"`),
							))
						})
						Specify("invalid index", func() {
							Expect(execute("list (a b c) remove a b")).To(Equal(
								ERROR(`invalid integer "a"`),
							))
							Expect(execute("list (a b c) remove 1 b")).To(Equal(
								ERROR(`invalid integer "b"`),
							))
						})
					})
				})

				Describe("`append`", func() {
					Specify("usage", func() {
						Expect(evaluate("help list () append")).To(Equal(
							STR("list value append ?list ...?"),
						))
					})

					It("should append two lists", func() {
						Expect(evaluate("list (a b c) append (foo bar)")).To(Equal(
							evaluate("list (a b c foo bar)"),
						))
					})
					It("should accept several lists", func() {
						Expect(
							evaluate("list (a b c) append (foo bar) (baz) (sprong yada)"),
						).To(Equal(evaluate("list (a b c foo bar baz sprong yada)")))
					})
					It("should accept zero list", func() {
						Expect(evaluate("list (a b c) append")).To(Equal(
							evaluate("list (a b c)"),
						))
					})

					Describe("Exceptions", func() {
						Specify("invalid list values", func() {
							Expect(execute("list (a b c) append []")).To(Equal(
								ERROR("invalid list"),
							))
							Expect(execute("list (a b c) append [1]")).To(Equal(
								ERROR("invalid list"),
							))
							Expect(execute("list (a b c) append a")).To(Equal(
								ERROR("invalid list"),
							))
						})
					})
				})

				Describe("`insert`", func() {
					Specify("usage", func() {
						Expect(evaluate("help list () insert")).To(Equal(
							STR("list value insert index value2"),
						))
					})

					It("should insert `list` at `index`", func() {
						Expect(evaluate("list (a b c) insert 1 (foo bar)")).To(Equal(
							evaluate("list (a foo bar b c)"),
						))
					})
					It("should prepend `list` when `index` is negative", func() {
						Expect(evaluate("list (a b c) insert -10 (foo bar)")).To(Equal(
							evaluate("list (foo bar a b c)"),
						))
					})
					It("should append `list` when `index` is past the target list length", func() {
						Expect(evaluate("list (a b c) insert 10 (foo bar)")).To(Equal(
							evaluate("list (a b c foo bar)"),
						))
					})

					Describe("Exceptions", func() {
						Specify("wrong arity", func() {
							Expect(execute("list (a b c) insert a")).To(Equal(
								ERROR(
									`wrong # args: should be "list value insert index value2"`,
								),
							))
							Expect(execute("list (a b c) insert a b c")).To(Equal(
								ERROR(
									`wrong # args: should be "list value insert index value2"`,
								),
							))
							Expect(execute("help list (a b c) insert a b c")).To(Equal(
								ERROR(
									`wrong # args: should be "list value insert index value2"`,
								),
							))
						})
						Specify("invalid `index`", func() {
							Expect(execute("list (a b c) insert a b")).To(Equal(
								ERROR(`invalid integer "a"`),
							))
						})
						Specify("invalid `list`", func() {
							Expect(execute("list (a b c) insert 1 []")).To(Equal(
								ERROR("invalid list"),
							))
							Expect(execute("list (a b c) append [1]")).To(Equal(
								ERROR("invalid list"),
							))
							Expect(execute("list (a b c) insert 1 a")).To(Equal(
								ERROR("invalid list"),
							))
						})
					})
				})

				Describe("`replace`", func() {
					Specify("usage", func() {
						Expect(evaluate("help list () replace")).To(Equal(
							STR("list value replace first last value2"),
						))
					})

					It("should replace the range included within [`first`, `last`] with `list`", func() {
						Expect(evaluate("list (a b c d e) replace 1 3 (foo bar)")).To(Equal(
							evaluate("list (a foo bar e)"),
						))
					})
					It("should truncate out of range boundaries", func() {
						Expect(evaluate("list (a b c) replace -10 1 (foo bar)")).To(Equal(
							evaluate("list (foo bar c)"),
						))
						Expect(evaluate("list (a b c) replace 2 10 (foo bar)")).To(Equal(
							evaluate("list (a b foo bar)"),
						))
						Expect(evaluate("list (a b c) replace -2 10 (foo bar)")).To(Equal(
							evaluate("list (foo bar)"),
						))
					})
					It("should insert `list` at `first` when `last` is before `first`", func() {
						Expect(evaluate("list (a b c) replace 2 0 (foo bar)")).To(Equal(
							evaluate("list (a b foo bar c)"),
						))
					})
					It("should prepend `list` when `last` is negative", func() {
						Expect(evaluate("list (a b c) replace -3 -1 (foo bar)")).To(Equal(
							evaluate("list (foo bar a b c)"),
						))
					})
					It("should append `list` when `first` is past the target list length", func() {
						Expect(evaluate("list (a b c) replace 10 12 (foo bar)")).To(Equal(
							evaluate("list (a b c foo bar)"),
						))
					})

					Describe("Exceptions", func() {
						Specify("wrong arity", func() {
							Expect(execute("list (a b c) replace a b")).To(Equal(
								ERROR(
									`wrong # args: should be "list value replace first last value2"`,
								),
							))
							Expect(execute("list (a b c) replace a b c d")).To(Equal(
								ERROR(
									`wrong # args: should be "list value replace first last value2"`,
								),
							))
							Expect(execute("help list (a b c) replace a b c d")).To(Equal(
								ERROR(
									`wrong # args: should be "list value replace first last value2"`,
								),
							))
						})
						Specify("invalid index", func() {
							Expect(execute("list (a b c) replace a b c")).To(Equal(
								ERROR(`invalid integer "a"`),
							))
							Expect(execute("list (a b c) replace 1 b c")).To(Equal(
								ERROR(`invalid integer "b"`),
							))
						})
						Specify("invalid `list`", func() {
							Expect(execute("list (a b c) replace 1 1 []")).To(Equal(
								ERROR("invalid list"),
							))
							Expect(execute("list (a b c) replace 1 1 [1]")).To(Equal(
								ERROR("invalid list"),
							))
							Expect(execute("list (a b c) replace 1 1 a")).To(Equal(
								ERROR("invalid list"),
							))
						})
					})
				})

				Describe("`sort`", func() {
					Specify("usage", func() {
						Expect(evaluate("help list () sort")).To(Equal(
							STR("list value sort"),
						))
					})

					It("should sort elements as strings in lexical order", func() {
						Expect(evaluate("list (c a d b) sort")).To(Equal(
							evaluate("list (a b c d)"),
						))
					})

					Describe("Exceptions", func() {
						Specify("wrong arity", func() {
							Expect(execute("list (a b c) sort a")).To(Equal(
								ERROR(`wrong # args: should be "list value sort"`),
							))
							Expect(execute("help list (a b c) sort a")).To(Equal(
								ERROR(`wrong # args: should be "list value sort"`),
							))
						})
						Specify("values with no string representation", func() {
							Expect(execute("list ([] ()) sort")).To(Equal(
								ERROR("value has no string representation"),
							))
						})
					})
				})
			})

			Describe("Iteration", func() {
				Describe("`foreach`", func() {
					Specify("usage", func() {
						Expect(evaluate("help list () foreach")).To(Equal(
							STR("list value foreach ?index? element body"),
						))
					})

					It("should iterate over elements", func() {
						evaluate(`
							set elements [list ()]
							set l [list (a b c)]
							list $l foreach element {
								set elements [list $elements append ($element)]
							}
						`)
						Expect(evaluate("get elements")).To(Equal(evaluate("get l")))
					})
					Describe("parameter tuples", func() {
						It("should be supported", func() {
							evaluate(`
								set elements [list ()]
								set l [list ((a b) (c d))]
								list $l foreach (i j) {
									set elements [list $elements append (($i $j))]
								}
							`)
							Expect(evaluate("get elements")).To(Equal(evaluate("get l")))
						})
						It("should accept empty tuple", func() {
							evaluate(`
								set i 0
								list ((a b) (c d) (e f)) foreach () {
									set i [+ $i 1]
								}
							`)
							Expect(evaluate("get i")).To(Equal(INT(3)))
						})
					})
					It("should return the result of the last command", func() {
						Expect(execute("list () foreach element {}")).To(Equal(OK(NIL)))
						Expect(execute("list (a b c) foreach element {}")).To(Equal(OK(NIL)))
						Expect(
							evaluate("set i 0; list (a b c) foreach element {set i [+ $i 1]}"),
						).To(Equal(INT(3)))
					})
					It("should increment `index` at each iteration", func() {
						Expect(
							evaluate(
								`set s ""; list (a b c) foreach index element {set s $s$index$element}`,
							),
						).To(Equal(STR("0a1b2c")))
					})

					Describe("Control flow", func() {
						Describe("`return`", func() {
							It("should interrupt the loop with `RETURN` code", func() {
								Expect(
									execute(
										"set i 0; list (a b c) foreach element {set i [+ $i 1]; return $element; unreachable}",
									),
								).To(Equal(execute("return a")))
								Expect(evaluate("get i")).To(Equal(INT(1)))
							})
						})
						Describe("`tailcall`", func() {
							It("should interrupt the loop with `RETURN` code", func() {
								Expect(
									execute(
										"set i 0; list (a b c) foreach element {set i [+ $i 1]; tailcall {idem $element}; unreachable}",
									),
								).To(Equal(execute("return a")))
								Expect(evaluate("get i")).To(Equal(INT(1)))
							})
						})
						Describe("`yield`", func() {
							It("should interrupt the body with `YIELD` code", func() {
								Expect(
									execute("list (a b c) foreach element {yield; unreachable}").Code,
								).To(Equal(core.ResultCode_YIELD))
							})
							It("should provide a resumable state", func() {
								process := prepareScript(
									"list (a b c) foreach element {idem _$[yield $element]_}",
								)

								result := process.Run()
								Expect(result.Code).To(Equal(core.ResultCode_YIELD))
								Expect(result.Value).To(Equal(STR("a")))

								process.YieldBack(STR("step 1"))
								result = process.Run()
								Expect(result.Code).To(Equal(core.ResultCode_YIELD))
								Expect(result.Value).To(Equal(STR("b")))

								process.YieldBack(STR("step 2"))
								result = process.Run()
								Expect(result.Code).To(Equal(core.ResultCode_YIELD))
								Expect(result.Value).To(Equal(STR("c")))

								process.YieldBack(STR("step 3"))
								result = process.Run()
								Expect(result).To(Equal(OK(STR("_step 3_"))))
							})
						})
						Describe("`error`", func() {
							It("should interrupt the loop with `ERROR` code", func() {
								Expect(
									execute(
										"set i 0; list (a b c) foreach element {set i [+ $i 1]; error msg; unreachable}",
									),
								).To(Equal(ERROR("msg")))
								Expect(evaluate("get i")).To(Equal(INT(1)))
							})
						})
						Describe("`break`", func() {
							It("should interrupt the body with nil result", func() {
								Expect(
									execute(
										"set i 0; list (a b c) foreach element {set i [+ $i 1]; break; unreachable}",
									),
								).To(Equal(OK(NIL)))
								Expect(evaluate("get i")).To(Equal(INT(1)))
							})
						})
						Describe("`continue`", func() {
							It("should interrupt the body iteration", func() {
								Expect(
									execute(
										"set i 0; list (a b c) foreach element {set i [+ $i 1]; continue; unreachable}",
									),
								).To(Equal(OK(NIL)))
								Expect(evaluate("get i")).To(Equal(INT(3)))
							})
						})
					})

					Describe("Exceptions", func() {
						Specify("wrong arity", func() {
							Expect(execute("list (a b c) foreach a")).To(Equal(
								ERROR(
									`wrong # args: should be "list value foreach ?index? element body"`,
								),
							))
							Expect(execute("list (a b c) foreach a b c d")).To(Equal(
								ERROR(
									`wrong # args: should be "list value foreach ?index? element body"`,
								),
							))
							Expect(execute("help list (a b c) foreach a b c d")).To(Equal(
								ERROR(
									`wrong # args: should be "list value foreach ?index? element body"`,
								),
							))
						})
						Specify("non-script body", func() {
							Expect(execute("list (a b c) foreach a b")).To(Equal(
								ERROR("body must be a script"),
							))
						})
						Specify("invalid `index` name", func() {
							Expect(execute("list (a b c) foreach [] a {}")).To(Equal(
								ERROR("invalid index name"),
							))
						})
						Specify("bad value shape", func() {
							Expect(execute("list (a b c) foreach () {}")).To(Equal(
								ERROR("bad value shape"),
							))
							Expect(
								execute("list ((a b) (c d) (e f)) foreach (i j k) {}"),
							).To(Equal(ERROR("bad value shape")))
						})
					})
				})
			})

			Describe("Exceptions", func() {
				Specify("unknown subcommand", func() {
					Expect(execute("list () unknownSubcommand")).To(Equal(
						ERROR(`unknown subcommand "unknownSubcommand"`),
					))
				})
				Specify("invalid subcommand name", func() {
					Expect(execute("list () []")).To(Equal(
						ERROR("invalid subcommand name"),
					))
				})
			})
		})

		Describe("Examples", func() {
			Specify("Currying and encapsulation", func() {
				example([]exampleSpec{
					{
						script: "set l (list (a b c d e f g))",
					},
					{
						script: "$l",
						result: evaluate("list (a b c d e f g)"),
					},
					{
						script: "$l length",
						result: INT(7),
					},
					{
						script: "$l at 2",
						result: STR("c"),
					},
					{
						script: "$l range 3 5",
						result: evaluate("list (d e f)"),
					},
				})
			})
			Specify("Argument type guard", func() {
				example([]exampleSpec{
					{
						script: "macro len ( (list l) ) {list $l length}",
					},
					{
						script: "len (1 2 3 4)",
						result: INT(4),
					},
					{
						script: "len invalidValue",
						result: ERROR("invalid list"),
					},
				})
			})
		})

		Describe("Ensemble command", func() {
			It("should return its ensemble metacommand when called with no argument", func() {
				Expect(evaluate("list").Type()).To(Equal(core.ValueType_COMMAND))
			})
			It("should be extensible", func() {
				evaluate(`
					[list] eval {
						macro foo {value} {idem bar}
					}
				`)
				Expect(evaluate("list (a b c) foo")).To(Equal(STR("bar")))
			})
			It("should support help for custom subcommands", func() {
				evaluate(`
					[list] eval {
						macro foo {value a b} {idem bar}
					}
				`)
				Expect(evaluate("help list (a b c) foo")).To(Equal(
					STR("list value foo a b"),
				))
				Expect(execute("help list (a b c) foo 1 2 3")).To(Equal(
					ERROR(`wrong # args: should be "list value foo a b"`),
				))
			})

			Describe("Examples", func() {
				Specify("Adding a `last` subcommand", func() {
					example([]exampleSpec{
						{
							script: `
								[list] eval {
									macro last {value} {
										list $value at [- [list $value length] 1]
									}
								}
							`,
						},
						{
							script: "list (a b c) last",
							result: STR("c"),
						},
					})
				})
				Specify("Using `foreach` to implement a `includes` subcommand", func() {
					example([]exampleSpec{
						{
							script: `
								[list] eval {
									proc includes {haystack needle} {
										list $haystack foreach element {
											if [string $needle == $element] {return [true]}
										}
										return [false]
									}
								}
							`,
						},
						{
							script: "list (a b c) includes b",
							result: TRUE,
						},
						{
							script: "list (a b c) includes d",
							result: FALSE,
						},
					})
				})
			})
		})
	})

	Describe("`DisplayListValue`", func() {
		It("should display lists as `list` command + tuple values", func() {
			list := LIST([]core.Value{STR("a"), STR("b"), STR("c")})
			Expect(DisplayListValue(list, nil)).To(Equal("[list (a b c)]"))
		})
		It("should produce an isomorphic string", func() {
			list := LIST([]core.Value{STR("a"), STR("b"), STR("c")})
			Expect(evaluate(`idem ` + DisplayListValue(list, nil))).To(Equal(list))
		})
	})
})
