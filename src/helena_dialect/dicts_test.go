package helena_dialect_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"helena/core"
	. "helena/helena_dialect"
)

var _ = Describe("Helena dictionaries", func() {
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

	Describe("dict", func() {
		Describe("Dictionary creation and conversion", func() {
			It("should return dictionary value", func() {
				Expect(evaluate("dict ()")).To(Equal(DICT(map[string]core.Value{})))
			})
			It("should convert key-value tuples to dictionaries", func() {
				example(exampleSpec{
					script: "dict (a b c d)",
					result: DICT(map[string]core.Value{
						"a": STR("b"),
						"c": STR("d"),
					}),
				})
			})
			It("should convert key-value blocks to dictionaries", func() {
				example(exampleSpec{
					script: "dict {a b c d}",
					result: DICT(map[string]core.Value{
						"a": STR("b"),
						"c": STR("d"),
					}),
				})
			})
			It("should convert key-value lists to dictionaries", func() {
				example(exampleSpec{
					script: "dict [list (a b c d)]",
					result: DICT(map[string]core.Value{
						"a": STR("b"),
						"c": STR("d"),
					}),
				})
			})
			It("should convert non-string keys to strings", func() {
				Expect(evaluate("dict ([1] a [2.5] b [true] c {block} d)")).To(Equal(
					DICT(map[string]core.Value{
						"1":     STR("a"),
						"2.5":   STR("b"),
						"true":  STR("c"),
						"block": STR("d"),
					}),
				))
			})
			It("should preserve values", func() {
				Expect(evaluate("dict (a [1] b () c [])")).To(Equal(
					DICT(map[string]core.Value{
						"a": INT(1),
						"b": TUPLE([]core.Value{}),
						"c": NIL,
					}),
				))
			})

			Describe("Exceptions", func() {
				Specify("invalid lists", func() {
					Expect(execute("dict []")).To(Equal(ERROR("invalid dictionary")))
					Expect(execute("dict [1]")).To(Equal(ERROR("invalid dictionary")))
					Expect(execute("dict a")).To(Equal(ERROR("invalid dictionary")))
				})
				Specify("invalid keys", func() {
					Expect(execute("dict ([] a)")).To(Equal(ERROR("invalid key")))
					Expect(execute("dict (() a)")).To(Equal(ERROR("invalid key")))
				})
				Specify("odd lists", func() {
					Expect(execute("dict (a)")).To(Equal(ERROR("invalid key-value list")))
					Expect(execute("dict {a b c}")).To(Equal(
						ERROR("invalid key-value list"),
					))
				})
				Specify("blocks with side effects", func() {
					Expect(execute("dict { $a b}")).To(Equal(ERROR("invalid list")))
					Expect(execute("dict { a [b] }")).To(Equal(ERROR("invalid list")))
					Expect(execute("dict { $[][a] b}")).To(Equal(ERROR("invalid list")))
					Expect(execute("dict { a $[](b) }")).To(Equal(ERROR("invalid list")))
				})
			})
		})

		Describe("Subcommands", func() {
			Describe("Introspection", func() {
				Describe("`subcommands`", func() {
					Specify("usage", func() {
						Expect(evaluate("help dict () subcommands")).To(Equal(
							STR("dict value subcommands"),
						))
					})

					It("should return list of subcommands", func() {
						// Expect(evaluate("dict () subcommands")).To(Equal(
						// 	evaluate(
						// 		TODO specify order?
						// 		"list (subcommands size has get add remove merge keys values entries foreach)",
						// 	),
						// ))
					})

					Describe("Exceptions", func() {
						Specify("wrong arity", func() {
							Expect(execute("dict () subcommands a")).To(Equal(
								ERROR(`wrong # args: should be "dict value subcommands"`),
							))
							Expect(execute("help dict () subcommands a")).To(Equal(
								ERROR(`wrong # args: should be "dict value subcommands"`),
							))
						})
					})
				})
			})

			Describe("Accessors", func() {
				Describe("`size`", func() {
					Specify("usage", func() {
						Expect(evaluate("help dict () size")).To(Equal(
							STR("dict value size"),
						))
					})

					It("should return the dictionary size", func() {
						Expect(evaluate("dict () size")).To(Equal(INT(0)))
						Expect(evaluate("dict (a b c d) size")).To(Equal(INT(2)))
					})

					Describe("Exceptions", func() {
						Specify("wrong arity", func() {
							Expect(execute("dict () size a")).To(Equal(
								ERROR(`wrong # args: should be "dict value size"`),
							))
							Expect(execute("help dict () size a")).To(Equal(
								ERROR(`wrong # args: should be "dict value size"`),
							))
						})
					})
				})

				Describe("`has`", func() {
					Specify("usage", func() {
						Expect(evaluate("help dict () has")).To(Equal(
							STR("dict value has key"),
						))
					})

					It("should test for `key` existence", func() {
						Expect(evaluate("dict (a b c d) has a")).To(Equal(TRUE))
						Expect(evaluate("dict (a b c d) has e")).To(Equal(FALSE))
					})

					Describe("Exceptions", func() {
						Specify("wrong arity", func() {
							Expect(execute("dict (a b c d) has")).To(Equal(
								ERROR(`wrong # args: should be "dict value has key"`),
							))
							Expect(execute("dict (a b c d) has a b")).To(Equal(
								ERROR(`wrong # args: should be "dict value has key"`),
							))
							Expect(execute("help dict (a b c d) has a b")).To(Equal(
								ERROR(`wrong # args: should be "dict value has key"`),
							))
						})
						Specify("invalid `key`", func() {
							Expect(execute("dict (a b c d) has []")).To(Equal(
								ERROR("invalid key"),
							))
							Expect(execute("dict (a b c d) has ()")).To(Equal(
								ERROR("invalid key"),
							))
						})
					})
				})

				Describe("`get`", func() {
					Specify("usage", func() {
						Expect(evaluate("help dict () get")).To(Equal(
							STR("dict value get key ?default?"),
						))
					})

					It("should return the value at `key`", func() {
						Expect(evaluate("dict (a b c d) get a")).To(Equal(STR("b")))
					})
					It("should return the default value for a non-existing key", func() {
						Expect(evaluate("dict (a b c d) get e default")).To(Equal(
							STR("default"),
						))
					})
					It("should support key tuples", func() {
						Expect(evaluate("dict (a b c d e f) get (a e)")).To(Equal(
							evaluate("idem (b f)"),
						))
					})
					Specify("`get` <-> keyed selector equivalence", func() {
						rootScope.SetNamedVariable(
							"v",
							DICT(map[string]core.Value{
								"a": STR("b"),
								"c": STR("d"),
							}),
						)
						evaluate("set d (dict $v)")

						Expect(execute("dict $v get a")).To(Equal(execute("idem $v(a)")))
						Expect(execute("$d get a")).To(Equal(execute("idem $v(a)")))
						Expect(execute("idem $[$d](a)")).To(Equal(execute("idem $v(a)")))

						Expect(execute("dict $v get c")).To(Equal(execute("idem $v(c)")))
						Expect(execute("$d get c")).To(Equal(execute("idem $v(c)")))
						Expect(execute("idem $[$d](c)")).To(Equal(execute("idem $v(c)")))
					})

					Describe("Exceptions", func() {
						Specify("wrong arity", func() {
							Expect(execute("dict (a b c d) get")).To(Equal(
								ERROR(`wrong # args: should be "dict value get key ?default?"`),
							))
							Expect(execute("dict (a b c d) get a b c")).To(Equal(
								ERROR(`wrong # args: should be "dict value get key ?default?"`),
							))
							Expect(execute("help dict (a b c d) get a b c")).To(Equal(
								ERROR(`wrong # args: should be "dict value get key ?default?"`),
							))
						})
						Specify("unknow key", func() {
							Expect(execute("dict (a b c d) get e")).To(Equal(
								ERROR(`unknown key "e"`),
							))
							Expect(execute("dict (a b c d) get (a e)")).To(Equal(
								ERROR(`unknown key "e"`),
							))
						})
						Specify("invalid key", func() {
							Expect(execute("dict (a b c d) get ([])")).To(Equal(
								ERROR("invalid key"),
							))
							Expect(execute("dict (a b c d) get []")).To(Equal(
								ERROR("invalid key"),
							))
							Expect(execute("dict (a b c d) get [list ()]")).To(Equal(
								ERROR("invalid key"),
							))
						})
						Specify("key tuples with default", func() {
							Expect(execute("dict (a b c d) get (a) default")).To(Equal(
								ERROR("cannot use default with key tuples"),
							))
						})
					})
				})

				Describe("`keys`", func() {
					Specify("usage", func() {
						Expect(evaluate("help dict () keys")).To(Equal(
							STR("dict value keys"),
						))
					})

					It("should return the list of keys", func() {
						// Expect(evaluate("dict (a b c d) keys")).To(Equal(
						// TODO specify order?
						// evaluate("list (a c)"),
						// ))
					})

					Describe("Exceptions", func() {
						Specify("wrong arity", func() {
							Expect(execute("dict (a b c d) keys a")).To(Equal(
								ERROR(`wrong # args: should be "dict value keys"`),
							))
							Expect(execute("help dict (a b c d) keys a")).To(Equal(
								ERROR(`wrong # args: should be "dict value keys"`),
							))
						})
					})
				})

				Describe("`values`", func() {
					Specify("usage", func() {
						Expect(evaluate("help dict () values")).To(Equal(
							STR("dict value values"),
						))
					})

					It("should return the list of values", func() {
						// Expect(evaluate("dict (a b c d) values")).To(Equal(
						// 	TODO preserve order?
						// 	evaluate("list (b d)"),
						// ))
					})

					Describe("Exceptions", func() {
						Specify("wrong arity", func() {
							Expect(execute("dict (a b c d) values a")).To(Equal(
								ERROR(`wrong # args: should be "dict value values"`),
							))
							Expect(execute("help dict (a b c d) values a")).To(Equal(
								ERROR(`wrong # args: should be "dict value values"`),
							))
						})
					})
				})

				Describe("`entries`", func() {
					Specify("usage", func() {
						Expect(evaluate("help dict () entries")).To(Equal(
							STR("dict value entries"),
						))
					})

					It("should return the list of key-value tuples", func() {
						// Expect(evaluate("dict (a b c d) entries")).To(Equal(
						// 	TODO preserve order?
						// 	evaluate("list ((a b) (c d))"),
						// ))
					})

					Describe("Exceptions", func() {
						Specify("wrong arity", func() {
							Expect(execute("dict (a b c d) entries a")).To(Equal(
								ERROR(`wrong # args: should be "dict value entries"`),
							))
							Expect(execute("help dict (a b c d) entries a")).To(Equal(
								ERROR(`wrong # args: should be "dict value entries"`),
							))
						})
					})
				})
			})

			Describe("Operations", func() {
				Describe("`add`", func() {
					Specify("usage", func() {
						Expect(evaluate("help dict () add")).To(Equal(
							STR("dict value add key value"),
						))
					})

					It("should add `value` for a new `key`", func() {
						Expect(evaluate("dict (a b c d) add e f")).To(Equal(
							evaluate("dict (a b c d e f)"),
						))
					})
					It("should replace the value for an existing `key`", func() {
						Expect(evaluate("dict (a b c d) add a e")).To(Equal(
							evaluate("dict (a e c d)"),
						))
					})

					Describe("Exceptions", func() {
						Specify("wrong arity", func() {
							Expect(execute("dict (a b c d) add a")).To(Equal(
								ERROR(`wrong # args: should be "dict value add key value"`),
							))
							Expect(execute("dict (a b c d) add a b c")).To(Equal(
								ERROR(`wrong # args: should be "dict value add key value"`),
							))
							Expect(execute("help dict (a b c d) add a b c")).To(Equal(
								ERROR(`wrong # args: should be "dict value add key value"`),
							))
						})
						Specify("invalid key", func() {
							Expect(execute("dict (a b c d) add [] b")).To(Equal(
								ERROR("invalid key"),
							))
							Expect(execute("dict (a b c d) add () b")).To(Equal(
								ERROR("invalid key"),
							))
						})
					})
				})

				Describe("`remove`", func() {
					Specify("usage", func() {
						Expect(evaluate("help dict () remove")).To(Equal(
							STR("dict value remove ?key ...?"),
						))
					})

					It("should remove the provided `key`", func() {
						Expect(evaluate("dict (a b c d) remove a")).To(Equal(
							evaluate("dict (c d)"),
						))
					})
					It("should accept several keys to remove", func() {
						Expect(evaluate("dict (a b c d e f) remove a e")).To(Equal(
							evaluate("dict (c d)"),
						))
					})
					It("should ignore unknown keys", func() {
						Expect(evaluate("dict (a b c d e f) remove g")).To(Equal(
							evaluate("dict (a b c d e f)"),
						))
					})
					It("should accept zero key", func() {
						Expect(evaluate("dict (a b c d e f) remove")).To(Equal(
							evaluate("dict (a b c d e f)"),
						))
					})

					Describe("Exceptions", func() {
						Specify("invalid `key`", func() {
							Expect(execute("dict (a b c d) remove []")).To(Equal(
								ERROR("invalid key"),
							))
							Expect(execute("dict (a b c d) remove ()")).To(Equal(
								ERROR("invalid key"),
							))
						})
					})
				})

				Describe("`merge`", func() {
					Specify("usage", func() {
						Expect(evaluate("help dict () merge")).To(Equal(
							STR("dict value merge ?dict ...?"),
						))
					})

					It("should merge two dictionaries", func() {
						Expect(evaluate("dict (a b c d) merge (foo bar)")).To(Equal(
							evaluate("dict (a b c d foo bar)"),
						))
					})
					It("should accept several dictionaries", func() {
						Expect(
							evaluate("dict (a b c d) merge (foo bar) (baz sprong)"),
						).To(Equal(evaluate("dict (a b c d foo bar baz sprong)")))
					})
					It("should accept zero dictionary", func() {
						Expect(evaluate("dict (a b c d) merge")).To(Equal(
							evaluate("dict (a b c d)"),
						))
					})

					Describe("Exceptions", func() {
						Specify("invalid dictionary values", func() {
							Expect(execute("dict (a b c d) merge []")).To(Equal(
								ERROR("invalid list"),
							))
							Expect(execute("dict (a b c d) merge [1]")).To(Equal(
								ERROR("invalid list"),
							))
							Expect(execute("dict (a b c d) merge e")).To(Equal(
								ERROR("invalid list"),
							))
							Expect(execute("dict (a b c d) merge (e)")).To(Equal(
								ERROR("invalid key-value list"),
							))
						})
					})
				})
			})

			Describe("Iteration", func() {
				Describe("`foreach`", func() {
					Specify("usage", func() {
						Expect(evaluate("help dict () foreach")).To(Equal(
							STR("dict value foreach entry body"),
						))
					})

					It("should iterate over entries", func() {
						evaluate(`
							set entries [list ()]
							set d [dict (a b c d e f)]
							dict $d foreach entry {
								set entries [list $entries append ($entry)]
							}
						`)
						// Expect(evaluate("get entries")).To(Equal(evaluate("dict $d entries"))) // TODO specify order?
					})
					Describe("entry parameter tuples", func() {
						It("should be supported", func() {
							evaluate(`
								set keys [list ()]
								set values [list ()]
								set d [dict (a b c d e f)]
								dict $d foreach (key value) {
									set keys [list $keys append ($key)]
									set values [list $values append ($value)]
								}
							`)
							// Expect(evaluate("get keys")).To(Equal(evaluate("dict $d keys"))) // TODO specify order?
							// Expect(evaluate("get values")).To(Equal(evaluate("dict $d values"))) // TODO specify order?
						})
						It("should accept empty tuple", func() {
							evaluate(`
								set i 0
								dict (a b c d e f) foreach () {
									set i [+ $i 1]
								}
							`)
							Expect(evaluate("get i")).To(Equal(INT(3)))
						})
						It("should accept `(key)` tuple", func() {
							evaluate(`
								set keys [list ()]
								set d [dict (a b c d e f)]
								dict (a b c d e f) foreach (key) {
									set keys [list $keys append ($key)]
								}
							`)
							// Expect(evaluate("get keys")).To(Equal(evaluate("dict $d keys"))) // TODO specify order?
						})
					})
					It("should return the result of the last command", func() {
						Expect(execute("dict () foreach entry {}")).To(Equal(OK(NIL)))
						Expect(execute("dict (a b) foreach entry {}")).To(Equal(OK(NIL)))
						Expect(
							evaluate(
								"set i 0; dict (a b c d e f) foreach entry {set i [+ $i 1]}",
							),
						).To(Equal(INT(3)))
					})

					Describe("Control flow", func() {
						Describe("`return`", func() {
							It("should interrupt the loop with `RETURN` code", func() {
								Expect(
									execute(
										"set i 0; dict (a b c d e f) foreach entry {set i [+ $i 1]; return $entry; unreachable}",
									),
								// TODO specify order?
								// ).To(Equal(execute("return (a b)")))
								).NotTo(BeNil())
								Expect(evaluate("get i")).To(Equal(INT(1)))
							})
						})
						Describe("`tailcall`", func() {
							It("should interrupt the loop with `RETURN` code", func() {
								Expect(
									execute(
										"set i 0; dict (a b c d e f) foreach entry {set i [+ $i 1]; tailcall {idem $entry}; unreachable}",
									),
								// TODO specify order?
								// ).To(Equal(execute("return (a b)")))
								).NotTo(BeNil())
								Expect(evaluate("get i")).To(Equal(INT(1)))
							})
						})
						Describe("`yield`", func() {
							It("should interrupt the body with `YIELD` code", func() {
								Expect(
									execute(
										"dict (a b c d e f) foreach entry {yield; unreachable}",
									).Code,
								).To(Equal(core.ResultCode_YIELD))
							})
							It("should provide a resumable state", func() {
								process := rootScope.PrepareScript(
									*parse(
										"dict (a b c d e f) foreach (key value) {idem _$[yield $key]_}",
									),
								)

								result := process.Run()
								Expect(result.Code).To(Equal(core.ResultCode_YIELD))
								//  Expect(result.Value).To(Equal(STR("a"))) // TODO specify order?

								process.YieldBack(STR("step 1"))
								result = process.Run()
								Expect(result.Code).To(Equal(core.ResultCode_YIELD))
								// Expect(result.Value).To(Equal(STR("c"))) // TODO specify order?

								process.YieldBack(STR("step 2"))
								result = process.Run()
								Expect(result.Code).To(Equal(core.ResultCode_YIELD))
								// Expect(result.Value).To(Equal(STR("e"))) // TODO specify order?

								process.YieldBack(STR("step 3"))
								result = process.Run()
								Expect(result).To(Equal(OK(STR("_step 3_"))))
							})
						})
						Describe("`error`", func() {
							It("should interrupt the loop with `ERROR` code", func() {
								Expect(
									execute(
										"set i 0; dict (a b c d e f) foreach entry {set i [+ $i 1]; error msg; unreachable}",
									),
								).To(Equal(ERROR("msg")))
								Expect(evaluate("get i")).To(Equal(INT(1)))
							})
						})
						Describe("`break`", func() {
							It("should interrupt the body with nil result", func() {
								Expect(
									execute(
										"set i 0; dict (a b c d e f) foreach entry {set i [+ $i 1]; break; unreachable}",
									),
								).To(Equal(OK(NIL)))
								Expect(evaluate("get i")).To(Equal(INT(1)))
							})
						})
						Describe("`continue`", func() {
							It("should interrupt the body iteration", func() {
								Expect(
									execute(
										"set i 0; dict (a b c d e f) foreach entry {set i [+ $i 1]; continue; unreachable}",
									),
								).To(Equal(OK(NIL)))
								Expect(evaluate("get i")).To(Equal(INT(3)))
							})
						})
					})

					Describe("Exceptions", func() {
						Specify("wrong arity", func() {
							Expect(execute("dict (a b c d) foreach a")).To(Equal(
								ERROR(`wrong # args: should be "dict value foreach entry body"`),
							))
							Expect(execute("dict (a b c d) foreach a b c")).To(Equal(
								ERROR(`wrong # args: should be "dict value foreach entry body"`),
							))
							Expect(execute("help dict (a b c d) foreach a b c")).To(Equal(
								ERROR(`wrong # args: should be "dict value foreach entry body"`),
							))
						})
						Specify("non-script body", func() {
							Expect(execute("dict (a b c d) foreach a b")).To(Equal(
								ERROR("body must be a script"),
							))
						})
						Specify("bad value shape", func() {
							Expect(
								execute("dict (a b c d e f) foreach ((key) value) {}"),
							).To(Equal(ERROR("bad value shape")))
							Expect(
								execute("dict (a b c d e f) foreach (key (value)) {}"),
							).To(Equal(ERROR("bad value shape")))
							Expect(
								execute("dict (a b c d e f) foreach (key value foo) {}"),
							).To(Equal(ERROR("bad value shape")))
						})
					})
				})
			})

			Describe("Exceptions", func() {
				Specify("unknown subcommand", func() {
					Expect(execute("dict () unknownSubcommand")).To(Equal(
						ERROR(`unknown subcommand "unknownSubcommand"`),
					))
				})
				Specify("invalid subcommand name", func() {
					Expect(execute("dict () []")).To(Equal(
						ERROR("invalid subcommand name"),
					))
				})
			})
		})

		Describe("Examples", func() {
			Specify("Currying and encapsulation", func() {
				example([]exampleSpec{
					{
						script: "set d (dict (a b c d))",
					},
					{
						script: "$d",
						result: evaluate("dict (a b c d)"),
					},
					{
						script: "$d size",
						result: INT(2),
					},
					{
						script: "$d get a",
						result: STR("b"),
					},
					// TODO preserve order?
					// {
					// 	script: "$d entries",
					// 	result: evaluate("dict (a b c d) entries"),
					// },
				})
			})
			Specify("Argument type guard", func() {
				example([]exampleSpec{
					{
						script: "macro len ( (dict d) ) {[dict $d size] * 2}",
					},
					{
						script: "len (1 2 3 4)",
						result: INT(4),
					},
					{
						script: "len invalidValue",
						result: ERROR("invalid dictionary"),
					},
				})
			})
		})

		Describe("Ensemble command", func() {
			It("should return its ensemble metacommand when called with no argument", func() {
				Expect(evaluate("dict").Type()).To(Equal(core.ValueType_COMMAND))
			})
			It("should be extensible", func() {
				evaluate(`
					[dict] eval {
						macro foo {value} {idem bar}
					}
				`)
				Expect(evaluate("dict (a b c d) foo")).To(Equal(STR("bar")))
			})
			It("should support help for custom subcommands", func() {
				evaluate(`
					[dict] eval {
						macro foo {value a b} {idem bar}
					}
				`)
				Expect(evaluate("help dict (a b c d) foo")).To(Equal(
					STR("dict value foo a b"),
				))
				Expect(execute("help dict (a b c d) foo 1 2 3")).To(Equal(
					ERROR(`wrong # args: should be "dict value foo a b"`),
				))
			})

			Describe("Examples", func() {
				Specify("Adding a `+` operator", func() {
					example([]exampleSpec{
						{
							script: `
								[dict] eval {
									macro + {d1 d2} {dict $d1 merge $d2}
								}
							`,
						},
						{
							script: "dict (a b c d) + (a e f g)",
							result: evaluate("dict (a e c d f g)"),
						},
					})
				})
			})
		})
	})

	Describe("`DisplayDictionaryValue`", func() {
		It("should display dictionaries as `dict` command + key-value tuple", func() {
			// dict := DICT(map[string]core.Value{
			// 	"a": STR("b"),
			// 	"c": STR("d"),
			// })

			// TODO preserve order?
			// Expect(DisplayDictionaryValue(dict, nil)).To(Equal("[dict (a b c d)]"))
		})
		It("should produce an isomorphic string", func() {
			dict := DICT(map[string]core.Value{
				"a": STR("b"),
				"c": STR("d"),
			})
			Expect(evaluate(`idem ` + DisplayDictionaryValue(dict, nil))).To(Equal(dict))
		})
	})
})
