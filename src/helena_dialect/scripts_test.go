package helena_dialect_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"helena/core"
	. "helena/helena_dialect"
)

var _ = Describe("Helena scripts", func() {
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

	Describe("parse", func() {
		Describe("Specifications", func() {
			It("should return a script value", func() {
				Expect(func() { _ = evaluate(`parse ""`).(core.ScriptValue) }).NotTo(Panic())
			})
			It("should return parsed script and source", func() {
				source := "cmd arg1 arg2"
				script := *parse(source)
				Expect(evaluate(`parse "` + source + `"`)).To(Equal(
					core.NewScriptValue(script, source),
				))
			})
			It("should parse blocks as string values", func() {
				evaluate("set script {cmd arg1 arg2}")
				Expect(evaluate(`parse $script`)).To(Equal(evaluate("get script")))
			})
		})

		Describe("Exceptions", func() {
			Specify("wrong arity", func() {
				Expect(execute("parse")).To(Equal(
					ERROR(`wrong # args: should be "parse source"`),
				))
				Expect(execute("parse a b")).To(Equal(
					ERROR(`wrong # args: should be "parse source"`),
				))
				Expect(execute("help parse a b")).To(Equal(
					ERROR(`wrong # args: should be "parse source"`),
				))
			})
			Specify("parsing error", func() {
				Expect(execute(`parse "{"`)).To(Equal(ERROR("unmatched left brace")))
				Expect(execute(`parse ")"`)).To(Equal(
					ERROR("unmatched right parenthesis"),
				))
				Expect(execute(`parse "#{"`)).To(Equal(
					ERROR("unmatched block comment delimiter"),
				))
			})
			Specify("values with no string representation", func() {
				Expect(execute("parse []")).To(Equal(
					ERROR("value has no string representation"),
				))
				Expect(execute("parse ()")).To(Equal(
					ERROR("value has no string representation"),
				))
			})
		})
	})

	Describe("script", func() {
		Describe("Script creation and conversion", func() {
			It("should return script value", func() {
				Expect(func() { _ = evaluate("script {}").(core.ScriptValue) }).NotTo(Panic())
			})
			It("should accept blocks", func() {
				Expect(evaluate("script {}")).To(Equal(core.NewScriptValue(*parse(""), "")))
				Expect(evaluate("script {a b c; d e}")).To(Equal(
					core.NewScriptValue(*parse("a b c; d e"), "a b c; d e"),
				))
			})
			Describe("tuples", func() {
				It("should be converted to scripts", func() {
					Expect(func() { _ = evaluate("script ()").(core.ScriptValue) }).NotTo(Panic())
				})
				Specify("string value should be undefined", func() {
					Expect(evaluate("script ()").(core.ScriptValue).Source).To(BeNil())
					Expect(evaluate("script (a b)").(core.ScriptValue).Source).To(BeNil())
				})
				Specify("empty tuples should return empty scripts", func() {
					script := evaluate("script ()").(core.ScriptValue)
					Expect(script.Script.Sentences).To(BeEmpty())
				})
				It("non-empty tuples should return single-sentence scripts", func() {
					script := evaluate(
						"script (cmd (a) ; ; #{comment}# [1])",
					).(core.ScriptValue)
					Expect(script.Script.Sentences).To(HaveLen(1))
					Expect(script.Script.Sentences[0].Words).To(Equal([]core.WordOrValue{
						{Value: STR("cmd")},
						{Value: TUPLE([]core.Value{STR("a")})},
						{Value: INT(1)},
					}))
				})
			})
		})

		Describe("Subcommands", func() {
			Describe("Introspection", func() {
				var _ = Describe("`subcommands`", func() {
					Specify("usage", func() {
						Expect(evaluate("help script {} subcommands")).To(Equal(
							STR("script value subcommands"),
						))
					})

					It("should return list of subcommands", func() {
						// Expect(evaluate("script {} subcommands")).To(Equal(
						// 	TODO specify order?
						// 	evaluate("list (subcommands length append split)"),
						// ))
					})

					var _ = Describe("Exceptions", func() {
						Specify("wrong arity", func() {
							Expect(execute("script {} subcommands a")).To(Equal(
								ERROR(`wrong # args: should be "script value subcommands"`),
							))
							Expect(execute("help script {} subcommands a")).To(Equal(
								ERROR(`wrong # args: should be "script value subcommands"`),
							))
						})
					})
				})
			})

			Describe("Accessors", func() {
				var _ = Describe("`length`", func() {
					It("should return the number of sentences", func() {
						Expect(evaluate("script {} length")).To(Equal(INT(0)))
						Expect(evaluate("script {a b; c d;; ;} length")).To(Equal(INT(2)))
						Expect(evaluate("script () length")).To(Equal(INT(0)))
						Expect(evaluate("script (a b; c d;; ;) length")).To(Equal(INT(1)))
					})

					var _ = Describe("Exceptions", func() {
						Specify("wrong arity", func() {
							Expect(execute("script () length a")).To(Equal(
								ERROR(`wrong # args: should be "script value length"`),
							))
							Expect(execute("help script () length a")).To(Equal(
								ERROR(`wrong # args: should be "script value length"`),
							))
						})
					})
				})
			})

			Describe("Operations", func() {
				var _ = Describe("`append`", func() {
					Specify("usage", func() {
						Expect(evaluate("help script {} append")).To(Equal(
							STR("script value append ?script ...?"),
						))
					})

					It("should append two scripts", func() {
						Expect(evaluate("script {a b c} append {foo bar}")).To(Equal(
							core.NewScriptValueWithNoSource(*parse("a b c; foo bar")),
						))
					})
					It("should accept several scripts", func() {
						Expect(
							evaluate(
								"script {a b; c ; d e} append {f g} {h i; j k l} {m n; o}",
							),
						).To(Equal(
							core.NewScriptValueWithNoSource(
								*parse("a b; c; d e; f g; h i; j k l; m n; o"),
							),
						))
					})
					It("should accept both scripts and tuples scripts", func() {
						Expect(
							(evaluate(
								"script {a b; c ; d e} append (f g) {h i; j k l} (m n; o)",
							).(core.ScriptValue)).Script.Sentences,
						).To(HaveLen(7))
					})
					It("should accept zero scripts", func() {
						Expect(evaluate("script {a b c} append")).To(Equal(
							evaluate("script {a b c}"),
						))
					})

					var _ = Describe("Exceptions", func() {
						Specify("invalid values", func() {
							Expect(execute("script {} append []")).To(Equal(
								ERROR("value must be a script or tuple"),
							))
							Expect(execute("script {} append a")).To(Equal(
								ERROR("value must be a script or tuple"),
							))
							Expect(execute("script {} append a [1]")).To(Equal(
								ERROR("value must be a script or tuple"),
							))
						})
					})
				})

				var _ = Describe("`split`", func() {
					Specify("usage", func() {
						Expect(evaluate("help script {} split")).To(Equal(
							STR("script value split"),
						))
					})

					It("should split script sentences into list of scripts", func() {
						Expect(evaluate("script {} split")).To(Equal(evaluate("list {}")))
						Expect(evaluate("script {a b; c d;; ;} split")).To(Equal(
							LIST([]core.Value{
								core.NewScriptValueWithNoSource(*parse("a b")),
								core.NewScriptValueWithNoSource(*parse("c d")),
							}),
						))
						Expect(evaluate("script () split")).To(Equal(LIST([]core.Value{})))
						Expect(evaluate("script (a b; c d;; ;) split")).To(Equal(
							evaluate("list ([script (a b c d)])"),
						))
					})

					var _ = Describe("Exceptions", func() {
						Specify("wrong arity", func() {
							Expect(execute("script () split a")).To(Equal(
								ERROR(`wrong # args: should be "script value split"`),
							))
							Expect(execute("help script () split a")).To(Equal(
								ERROR(`wrong # args: should be "script value split"`),
							))
						})
					})
				})
			})

			Describe("Exceptions", func() {
				Specify("unknown subcommand", func() {
					Expect(execute("script {} unknownSubcommand")).To(Equal(
						ERROR(`unknown subcommand "unknownSubcommand"`),
					))
				})
				Specify("invalid subcommand name", func() {
					Expect(execute("script {} []")).To(Equal(
						ERROR("invalid subcommand name"),
					))
				})
			})
		})

		Describe("Examples", func() {
			Specify("Currying and encapsulation", func() {
				example([]exampleSpec{
					{
						script: "set s (script {a b c; d e; f})",
					},
					{
						script: "$s",
						result: evaluate("script {a b c; d e; f}"),
					},
					{
						script: "$s length",
						result: INT(3),
					},
				})
			})
			Specify("Argument type guard", func() {
				example([]exampleSpec{
					{
						script: "macro len ( (script s) ) {script $s length}",
					},
					{
						script: "len {a b c; d e; f}",
						result: INT(3),
					},
					{
						script: "len invalidValue",
						result: ERROR("value must be a script or tuple"),
					},
				})
			})
		})

		Describe("Ensemble command", func() {
			It("should return its ensemble metacommand when called with no argument", func() {
				Expect(evaluate("script").Type()).To(Equal(core.ValueType_COMMAND))
			})
			It("should be extensible", func() {
				evaluate(`
          [script] eval {
            macro foo {value} {idem bar}
          }
        `)
				Expect(evaluate("script {} foo")).To(Equal(STR("bar")))
			})
			It("should support help for custom subcommands", func() {
				evaluate(`
          [script] eval {
            macro foo {value a b} {idem bar}
          }
        `)
				Expect(evaluate("help script {} foo")).To(Equal(
					STR("script value foo a b"),
				))
				Expect(execute("help script {} foo 1 2 3")).To(Equal(
					ERROR(`wrong # args: should be "script value foo a b"`),
				))
			})

			Describe("Examples", func() {
				Specify("Adding a `last` subcommand", func() {
					example([]exampleSpec{
						{
							script: `
              [script] eval {
                macro last {value} {
                  list [script $value split] at [- [script $value length] 1]
                }
              }
            `,
						},
						{
							script: `
              set s [script {error a; return b; idem c} last]
              eval $s
            `,
							result: STR("c"),
						},
					})
				})
			})
		})
	})
})
