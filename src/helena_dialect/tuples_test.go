package helena_dialect_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"helena/core"
	. "helena/helena_dialect"
)

var _ = Describe("Helena tuples", func() {

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

	Describe("tuple", func() {
		Describe("Tuple creation and conversion", func() {
			It("should return tuple value", func() {
				Expect(evaluate("tuple ()")).To(Equal(TUPLE([]core.Value{})))
			})
			It("should convert lists to tuple", func() {
				Expect(evaluate("tuple [list (a b c)]")).To(Equal(
					TUPLE([]core.Value{STR("a"), STR("b"), STR("c")}),
				))
			})
			It("should convert blocks to tuples", func() {
				example(exampleSpec{
					script: "tuple {a b c}",
					result: TUPLE([]core.Value{STR("a"), STR("b"), STR("c")}),
				})
			})

			Describe("Exceptions", func() {
				Specify("invalid values", func() {
					Expect(execute("tuple []")).To(Equal(ERROR("invalid tuple")))
					Expect(execute("tuple [1]")).To(Equal(ERROR("invalid tuple")))
					Expect(execute("tuple a")).To(Equal(ERROR("invalid tuple")))
				})
				Specify("blocks with side effects", func() {
					Expect(execute("tuple { $a }")).To(Equal(ERROR("invalid list")))
					Expect(execute("tuple { [b] }")).To(Equal(ERROR("invalid list")))
					Expect(execute("tuple { $[][a] }")).To(Equal(ERROR("invalid list")))
					Expect(execute("tuple { $[](a) }")).To(Equal(ERROR("invalid list")))
				})
			})
		})

		Describe("Subcommands", func() {
			Describe("Introspection", func() {
				Describe("`subcommands`", func() {
					Specify("usage", func() {
						Expect(evaluate("help tuple () subcommands")).To(Equal(
							STR("tuple value subcommands"),
						))
					})

					It("should return list of subcommands", func() {
						// Expect(evaluate("tuple () subcommands")).To(Equal(
						// 	TODO specify order?
						// 	evaluate("list (subcommands length at)"),
						// ))
					})

					Describe("Exceptions", func() {
						Specify("wrong arity", func() {
							Expect(execute("tuple () subcommands a")).To(Equal(
								ERROR(`wrong # args: should be "tuple value subcommands"`),
							))
							Expect(execute("help tuple () subcommands a")).To(Equal(
								ERROR(`wrong # args: should be "tuple value subcommands"`),
							))
						})
					})
				})
			})

			Describe("Accessors", func() {
				Describe("`length`", func() {
					Specify("usage", func() {
						Expect(evaluate("help tuple () length")).To(Equal(
							STR("tuple value length"),
						))
					})

					It("should return the tuple length", func() {
						Expect(evaluate("tuple () length")).To(Equal(INT(0)))
						Expect(evaluate("tuple (a b c) length")).To(Equal(INT(3)))
					})

					Describe("Exceptions", func() {
						Specify("wrong arity", func() {
							Expect(execute("tuple () length a")).To(Equal(
								ERROR(`wrong # args: should be "tuple value length"`),
							))
							Expect(execute("help tuple () length a")).To(Equal(
								ERROR(`wrong # args: should be "tuple value length"`),
							))
						})
					})
				})

				Describe("`at`", func() {
					Specify("usage", func() {
						Expect(evaluate("help tuple () at")).To(Equal(
							STR("tuple value at index ?default?"),
						))
					})

					It("should return the element at `index`", func() {
						Expect(evaluate("tuple (a b c) at 1")).To(Equal(STR("b")))
					})
					It("should return the default value for an out-of-range `index`", func() {
						Expect(evaluate("tuple (a b c) at 10 default")).To(Equal(
							STR("default"),
						))
					})

					Describe("Exceptions", func() {
						Specify("wrong arity", func() {
							Expect(execute("tuple (a b c) at")).To(Equal(
								ERROR(
									`wrong # args: should be "tuple value at index ?default?"`,
								),
							))
							Expect(execute("tuple (a b c) at a b c")).To(Equal(
								ERROR(
									`wrong # args: should be "tuple value at index ?default?"`,
								),
							))
							Expect(execute("help tuple (a b c) at a b c")).To(Equal(
								ERROR(
									`wrong # args: should be "tuple value at index ?default?"`,
								),
							))
						})
						Specify("invalid `index`", func() {
							Expect(execute("tuple (a b c) at a")).To(Equal(
								ERROR(`invalid integer "a"`),
							))
						})
						Specify("`index` out of range", func() {
							Expect(execute("tuple (a b c) at -1")).To(Equal(
								ERROR(`index out of range "-1"`),
							))
							Expect(execute("tuple (a b c) at 10")).To(Equal(
								ERROR(`index out of range "10"`),
							))
						})
					})
				})
			})

			Describe("Exceptions", func() {
				Specify("unknown subcommand", func() {
					Expect(execute("tuple () unknownSubcommand")).To(Equal(
						ERROR(`unknown subcommand "unknownSubcommand"`),
					))
				})
				Specify("invalid subcommand name", func() {
					Expect(execute("tuple () []")).To(Equal(
						ERROR("invalid subcommand name"),
					))
				})
			})
		})

		Describe("Examples", func() {
			Specify("Currying and encapsulation", func() {
				example([]exampleSpec{
					{
						script: "set t (tuple (a b c d e f g))",
					},
					{
						script: "$t",
						result: evaluate("tuple (a b c d e f g)"),
					},
					{
						script: "$t length",
						result: INT(7),
					},
					{
						script: "$t at 2",
						result: STR("c"),
					},
				})
			})
			Specify("Argument type guard", func() {
				example([]exampleSpec{
					{
						script: "macro len ( (tuple t) ) {tuple $t length}",
					},
					{
						script: "len (1 2 3 4)",
						result: INT(4),
					},
					{
						script: "len [list {1 2 3}]",
						result: INT(3),
					},
					{
						script: "len invalidValue",
						result: ERROR("invalid tuple"),
					},
				})
			})
		})

		Describe("Ensemble command", func() {
			It("should return its ensemble metacommand when called with no argument", func() {
				Expect(evaluate("tuple").Type()).To(Equal(core.ValueType_COMMAND))
			})
			It("should be extensible", func() {
				evaluate(`
	          [tuple] eval {
	            macro foo {value} {idem bar}
	          }
	        `)
				Expect(evaluate("tuple (a b c) foo")).To(Equal(STR("bar")))
			})
			It("should support help for custom subcommands", func() {
				evaluate(`
	          [tuple] eval {
	            macro foo {value a b} {idem bar}
	          }
	        `)
				Expect(evaluate("help tuple (a b c) foo")).To(Equal(
					STR("tuple value foo a b"),
				))
				Expect(execute("help tuple (a b c) foo 1 2 3")).To(Equal(
					ERROR(`wrong # args: should be "tuple value foo a b"`),
				))
			})

			Describe("Examples", func() {
				Specify("Adding a `last` subcommand", func() {
					example([]exampleSpec{
						{
							script: `
	              [tuple] eval {
	                macro last {value} {
	                  tuple $value at [- [tuple $value length] 1]
	                }
	              }
	            `,
						},
						{
							script: "tuple (a b c) last",
							result: STR("c"),
						},
					})
				})
			})
		})
	})
})
