package helena_dialect_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"helena/core"
	. "helena/helena_dialect"
)

var _ = Describe("Helena argument handling", func() {
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
		parser = core.NewParser(nil)
	}

	example := specifyExample(func(spec exampleSpec) core.Result { return execute(spec.script) })

	BeforeEach(init)

	Describe("argspec", func() {
		Describe("Argspec creation and conversion", func() {
			It("should return argspec value", func() {
				Expect(func() { _ = evaluate("argspec ()").(ArgspecValue) }).NotTo(Panic())
			})
			It("should convert blocks to argspecs", func() {
				example(exampleSpec{
					script: "argspec {a b c}",
					result: evaluate("argspec (a b c)"),
				})
			})
			It("should convert tuples to argspecs", func() {
				example(exampleSpec{
					script: "argspec (a b c)",
				})
			})
			It("should convert lists to argspecs", func() {
				example(exampleSpec{
					script: "argspec [list {a b c}]",
					result: evaluate("argspec (a b c)"),
				})
			})

			Describe("Exceptions", func() {
				Specify("invalid values", func() {
					Expect(execute("argspec []")).To(Equal(ERROR("invalid argument list")))
					Expect(execute("argspec [1]")).To(Equal(ERROR("invalid argument list")))
					Expect(execute("argspec a")).To(Equal(ERROR("invalid argument list")))
				})
				Specify("blocks with side effects", func() {
					Expect(execute("argspec { $a }")).To(Equal(
						ERROR("invalid argument list"),
					))
					Expect(execute("argspec { [b] }")).To(Equal(
						ERROR("invalid argument list"),
					))
					Expect(execute("argspec { $[][a] }")).To(Equal(
						ERROR("invalid argument list"),
					))
					Expect(execute("argspec { $[](a) }")).To(Equal(
						ERROR("invalid argument list"),
					))
				})
			})
		})

		Describe("Argument specifications", func() {
			Describe("empty", func() {
				Specify("value", func() {
					value := evaluate("argspec ()").(ArgspecValue)
					Expect(evaluate("argspec {}")).To(Equal(value))
					Expect(value.Argspec).To(And(
						HaveField("NbRequired", uint(0)),
						HaveField("NbOptional", uint(0)),
						HaveField("HasRemainder", false),
					))
					Expect(value.Argspec.Args).To(BeEmpty())
				})
				Specify("usage", func() {
					Expect(evaluate("argspec () usage")).To(Equal(STR("")))
				})
				Specify("set", func() {
					evaluate("argspec () set ()")
					Expect(rootScope.Context.Variables).To(BeEmpty())
				})
			})

			Describe("one parameter", func() {
				Specify("value", func() {
					value := evaluate("argspec (a)").(ArgspecValue)
					Expect(evaluate("argspec {a}")).To(Equal(value))
					Expect(value.Argspec).To(And(
						HaveField("NbRequired", uint(1)),
						HaveField("NbOptional", uint(0)),
						HaveField("HasRemainder", false),
					))
					Expect(value.Argspec.Args).To(Equal([]Argument{{Name: "a", Type: ArgumentType_REQUIRED}}))
				})
				Specify("usage", func() {
					Expect(evaluate("argspec (a) usage")).To(Equal(STR("a")))
				})
				Specify("set", func() {
					evaluate("argspec (a) set (val1)")
					Expect(evaluate("get a")).To(Equal(STR("val1")))
					evaluate("argspec (a) set (val2)")
					Expect(evaluate("get a")).To(Equal(STR("val2")))
				})
			})

			Describe("two parameters", func() {
				Specify("value", func() {
					value := evaluate("argspec (a b)").(ArgspecValue)
					Expect(evaluate("argspec {a b}")).To(Equal(value))
					Expect(value.Argspec).To(And(
						HaveField("NbRequired", uint(2)),
						HaveField("NbOptional", uint(0)),
						HaveField("HasRemainder", false),
					))
					Expect(value.Argspec.Args).To(Equal([]Argument{
						{Name: "a", Type: ArgumentType_REQUIRED},
						{Name: "b", Type: ArgumentType_REQUIRED},
					}))
				})
				Specify("usage", func() {
					Expect(evaluate("argspec (a b) usage")).To(Equal(STR("a b")))
				})
				Specify("set", func() {
					evaluate("argspec {a b} set (val1 val2)")
					Expect(evaluate("get a")).To(Equal(STR("val1")))
					Expect(evaluate("get b")).To(Equal(STR("val2")))
				})
			})

			Describe("remainder", func() {
				Describe("anonymous", func() {
					Specify("value", func() {
						value := evaluate("argspec (*)").(ArgspecValue)
						Expect(evaluate("argspec {*}")).To(Equal(value))
						Expect(value.Argspec).To(And(
							HaveField("NbRequired", uint(0)),
							HaveField("NbOptional", uint(0)),
							HaveField("HasRemainder", true),
						))
						Expect(value.Argspec.Args).To(Equal([]Argument{
							{Name: "*", Type: ArgumentType_REMAINDER},
						}))
					})
					Specify("usage", func() {
						Expect(evaluate("argspec (*) usage")).To(Equal(STR("?arg ...?")))
					})
					Describe("set", func() {
						Specify("zero", func() {
							evaluate("argspec (*) set ()")
							Expect(evaluate("get *")).To(Equal(TUPLE([]core.Value{})))
						})
						Specify("one", func() {
							evaluate("argspec (*) set (val)")
							Expect(evaluate("get *")).To(Equal(TUPLE([]core.Value{STR("val")})))
						})
						Specify("two", func() {
							evaluate("argspec (*) set (val1 val2)")
							Expect(evaluate("get *")).To(Equal(
								TUPLE([]core.Value{STR("val1"), STR("val2")}),
							))
						})
					})
				})

				Describe("named", func() {
					Specify("value", func() {
						value := evaluate("argspec (*args)").(ArgspecValue)
						Expect(evaluate("argspec {*args}")).To(Equal(value))
						Expect(value.Argspec).To(And(
							HaveField("NbRequired", uint(0)),
							HaveField("NbOptional", uint(0)),
							HaveField("HasRemainder", true),
						))
						Expect(value.Argspec.Args).To(Equal([]Argument{
							{Name: "args", Type: ArgumentType_REMAINDER},
						}))
					})
					Specify("usage", func() {
						Expect(evaluate("argspec (*remainder) usage")).To(Equal(
							STR("?remainder ...?"),
						))
					})
					Describe("set", func() {
						Specify("zero", func() {
							evaluate("argspec (*args) set ()")
							Expect(evaluate("get args")).To(Equal(TUPLE([]core.Value{})))
						})
						Specify("one", func() {
							evaluate("argspec (*args) set (val)")
							Expect(evaluate("get args")).To(Equal(TUPLE([]core.Value{STR("val")})))
						})
						Specify("two", func() {
							evaluate("argspec (*args) set (val1 val2)")
							Expect(evaluate("get args")).To(Equal(
								TUPLE([]core.Value{STR("val1"), STR("val2")}),
							))
						})
					})
				})

				Describe("prefix", func() {
					Specify("one", func() {
						evaluate("argspec (* a) set (val)")
						Expect(evaluate("get *")).To(Equal(TUPLE([]core.Value{})))
						Expect(evaluate("get a")).To(Equal(STR("val")))
					})
					Specify("two", func() {
						evaluate("argspec (* a) set (val1 val2)")
						Expect(evaluate("get *")).To(Equal(TUPLE([]core.Value{STR("val1")})))
						Expect(evaluate("get a")).To(Equal(STR("val2")))
					})
					Specify("three", func() {
						evaluate("argspec (* a) set (val1 val2 val3)")
						Expect(evaluate("get *")).To(Equal(TUPLE([]core.Value{STR("val1"), STR("val2")})))
						Expect(evaluate("get a")).To(Equal(STR("val3")))
					})
				})
				Describe("infix", func() {
					Specify("two", func() {
						evaluate("argspec (a * b) set (val1 val2)")
						Expect(evaluate("get a")).To(Equal(STR("val1")))
						Expect(evaluate("get *")).To(Equal(TUPLE([]core.Value{})))
						Expect(evaluate("get b")).To(Equal(STR("val2")))
					})
					Specify("three", func() {
						evaluate("argspec (a * b) set (val1 val2 val3)")
						Expect(evaluate("get a")).To(Equal(STR("val1")))
						Expect(evaluate("get *")).To(Equal(TUPLE([]core.Value{STR("val2")})))
						Expect(evaluate("get b")).To(Equal(STR("val3")))
					})
					Specify("four", func() {
						evaluate("argspec (a * b) set (val1 val2 val3 val4)")
						Expect(evaluate("get a")).To(Equal(STR("val1")))
						Expect(evaluate("get *")).To(Equal(TUPLE([]core.Value{STR("val2"), STR("val3")})))
						Expect(evaluate("get b")).To(Equal(STR("val4")))
					})
				})
				Describe("suffix", func() {
					Specify("one", func() {
						evaluate("argspec (a *) set (val)")
						Expect(evaluate("get *")).To(Equal(TUPLE([]core.Value{})))
						Expect(evaluate("get a")).To(Equal(STR("val")))
						Expect(evaluate("get *")).To(Equal(TUPLE([]core.Value{})))
					})
					Specify("two", func() {
						evaluate("argspec (a *) set (val1 val2)")
						Expect(evaluate("get a")).To(Equal(STR("val1")))
						Expect(evaluate("get *")).To(Equal(TUPLE([]core.Value{STR("val2")})))
					})
					Specify("three", func() {
						evaluate("argspec (a *) set (val1 val2 val3)")
						Expect(evaluate("get a")).To(Equal(STR("val1")))
						Expect(evaluate("get *")).To(Equal(TUPLE([]core.Value{STR("val2"), STR("val3")})))
					})
				})

				It("cannot be used more than once", func() {
					Expect(execute("argspec (* *)")).To(Equal(
						ERROR("only one remainder argument is allowed"),
					))
					Expect(execute("argspec (*a *b)")).To(Equal(
						ERROR("only one remainder argument is allowed"),
					))
				})
			})

			Describe("optional parameter", func() {
				Describe("single", func() {
					Specify("value", func() {
						value := evaluate("argspec (?a)").(ArgspecValue)
						Expect(evaluate("argspec {?a}")).To(Equal(value))
						Expect(evaluate("argspec ((?a))")).To(Equal(value))
						Expect(evaluate("argspec {(?a)}")).To(Equal(value))
						Expect(evaluate("argspec ({?a})")).To(Equal(value))
						Expect(evaluate("argspec {{?a}}")).To(Equal(value))
						Expect(value.Argspec).To(And(
							HaveField("NbRequired", uint(0)),
							HaveField("NbOptional", uint(1)),
							HaveField("HasRemainder", false),
						))
						Expect(value.Argspec.Args).To(Equal([]Argument{
							{Name: "a", Type: ArgumentType_OPTIONAL},
						}))
					})
					Specify("usage", func() {
						Expect(evaluate("argspec (?a) usage")).To(Equal(STR("?a?")))
					})
					Describe("set", func() {
						Specify("zero", func() {
							evaluate("argspec (?a) set ()")
							Expect(execute("get a")).To(Equal(
								ERROR(`cannot get "a": no such variable`),
							))
						})
						Specify("one", func() {
							evaluate("argspec (?a) set (val)")
							Expect(evaluate("get a")).To(Equal(STR("val")))
						})
					})
				})
				Describe("multiple", func() {
					Specify("value", func() {
						value := evaluate("argspec {?a ?b}").(ArgspecValue)
						Expect(evaluate("argspec (?a ?b)")).To(Equal(value))
						Expect(value.Argspec).To(And(
							HaveField("NbRequired", uint(0)),
							HaveField("NbOptional", uint(2)),
							HaveField("HasRemainder", false),
						))
						Expect(value.Argspec.Args).To(Equal([]Argument{
							{Name: "a", Type: ArgumentType_OPTIONAL},
							{Name: "b", Type: ArgumentType_OPTIONAL},
						}))
					})
					Specify("usage", func() {
						Expect(evaluate("argspec (?a ?b) usage")).To(Equal(STR("?a? ?b?")))
					})
					Describe("set", func() {
						Specify("zero", func() {
							evaluate("argspec (?a ?b) set ()")
							Expect(execute("get a")).To(Equal(
								ERROR(`cannot get "a": no such variable`),
							))
							Expect(execute("get b")).To(Equal(
								ERROR(`cannot get "b": no such variable`),
							))
						})
						Specify("one", func() {
							evaluate("argspec (?a ?b) set (val)")
							Expect(evaluate("get a")).To(Equal(STR("val")))
							Expect(execute("get b")).To(Equal(
								ERROR(`cannot get "b": no such variable`),
							))
						})
						Specify("two", func() {
							evaluate("argspec (?a ?b) set (val1 val2)")
							Expect(evaluate("get a")).To(Equal(STR("val1")))
							Expect(evaluate("get b")).To(Equal(STR("val2")))
						})
					})
				})

				Describe("prefix", func() {
					Specify("one", func() {
						evaluate("argspec (?a b) set (val)")
						Expect(execute("get a")).To(Equal(
							ERROR(`cannot get "a": no such variable`),
						))
						Expect(evaluate("get b")).To(Equal(STR("val")))
					})
					Specify("two", func() {
						evaluate("argspec (?a b) set (val1 val2)")
						Expect(evaluate("get a")).To(Equal(STR("val1")))
						Expect(evaluate("get b")).To(Equal(STR("val2")))
					})
				})
				Describe("infix", func() {
					Specify("two", func() {
						evaluate("argspec (a ?b c) set (val1 val2)")
						Expect(evaluate("get a")).To(Equal(STR("val1")))
						Expect(execute("get b")).To(Equal(
							ERROR(`cannot get "b": no such variable`),
						))
						Expect(evaluate("get c")).To(Equal(STR("val2")))
					})
					Specify("three", func() {
						evaluate("argspec (a ?b c) set (val1 val2 val3)")
						Expect(evaluate("get a")).To(Equal(STR("val1")))
						Expect(evaluate("get b")).To(Equal(STR("val2")))
						Expect(evaluate("get c")).To(Equal(STR("val3")))
					})
				})
				Describe("suffix", func() {
					Specify("one", func() {
						evaluate("argspec (a ?b) set (val)")
						Expect(evaluate("get a")).To(Equal(STR("val")))
						Expect(execute("get b")).To(Equal(
							ERROR(`cannot get "b": no such variable`),
						))
					})
					Specify("two", func() {
						evaluate("argspec (a ?b) set (val1 val2)")
						Expect(evaluate("get a")).To(Equal(STR("val1")))
						Expect(evaluate("get b")).To(Equal(STR("val2")))
					})
				})
			})

			Describe("default parameter", func() {
				Specify("value", func() {
					value := evaluate("argspec ((?a val))").(ArgspecValue)
					Expect(evaluate("argspec {(?a val)}")).To(Equal(value))
					Expect(evaluate("argspec ({?a val})")).To(Equal(value))
					Expect(evaluate("argspec {{?a val}}")).To(Equal(value))
					Expect(value.Argspec).To(And(
						HaveField("NbRequired", uint(0)),
						HaveField("NbOptional", uint(1)),
						HaveField("HasRemainder", false),
					))
					Expect(value.Argspec.Args).To(Equal([]Argument{
						{Name: "a", Type: ArgumentType_OPTIONAL, Default: STR("val")},
					}))
				})
				Specify("usage", func() {
					Expect(evaluate("argspec ((?a def)) usage")).To(Equal(STR("?a?")))
				})
				Describe("set", func() {
					Describe("static", func() {
						Specify("zero", func() {
							evaluate("argspec ((?a def)) set ()")
							Expect(evaluate("get a")).To(Equal(STR("def")))
						})
						Specify("one", func() {
							evaluate("argspec ((?a def)) set (val)")
							Expect(evaluate("get a")).To(Equal(STR("val")))
						})
					})
					Describe("dynamic", func() {
						Specify("zero", func() {
							evaluate("argspec ((?a {+ 1 2})) set ()")
							Expect(evaluate("get a")).To(Equal(INT(3)))
						})
						Specify("one", func() {
							evaluate("argspec ((?a {unreachable})) set (val)")
							Expect(evaluate("get a")).To(Equal(STR("val")))
						})
						Specify("unexpected result", func() {
							Expect(execute("argspec ((?a {return})) set ()")).To(Equal(
								ERROR("unexpected return"),
							))
							Expect(execute("argspec ((?a {yield})) set ()")).To(Equal(
								ERROR("unexpected yield"),
							))
							Expect(execute("argspec ((?a {error msg})) set ()")).To(Equal(
								ERROR("msg"),
							))
							Expect(execute("argspec ((?a {break})) set ()")).To(Equal(
								ERROR("unexpected break"),
							))
							Expect(execute("argspec ((?a {continue})) set ()")).To(Equal(
								ERROR("unexpected continue"),
							))
						})
					})
				})
			})

			Describe("guard", func() {
				Specify("required parameter", func() {
					value := evaluate("argspec ((list a))").(ArgspecValue)
					Expect(value.Argspec).To(And(
						HaveField("NbRequired", uint(1)),
						HaveField("NbOptional", uint(0)),
						HaveField("HasRemainder", false),
					))
					Expect(value.Argspec.Args).To(Equal([]Argument{
						{Name: "a", Type: ArgumentType_REQUIRED, Guard: STR("list")},
					}))
				})
				Specify("optional parameter", func() {
					value := evaluate("argspec ((list ?a))").(ArgspecValue)
					Expect(value.Argspec).To(And(
						HaveField("NbRequired", uint(0)),
						HaveField("NbOptional", uint(1)),
						HaveField("HasRemainder", false),
					))
					Expect(value.Argspec.Args).To(Equal([]Argument{
						{Name: "a", Type: ArgumentType_OPTIONAL, Guard: STR("list")},
					}))
				})
				Specify("default parameter", func() {
					value := evaluate("argspec ((list ?a val))").(ArgspecValue)
					Expect(value.Argspec).To(And(
						HaveField("NbRequired", uint(0)),
						HaveField("NbOptional", uint(1)),
						HaveField("HasRemainder", false),
					))
					Expect(value.Argspec.Args).To(Equal([]Argument{
						{
							Name:    "a",
							Type:    ArgumentType_OPTIONAL,
							Guard:   STR("list"),
							Default: STR("val"),
						},
					}))
				})
				Specify("usage", func() {
					Expect(evaluate("argspec ((guard ?a def)) usage")).To(Equal(STR("?a?")))
				})
				Describe("set", func() {
					Describe("simple command", func() {
						Specify("required", func() {
							evaluate("set args [argspec ( (list a) )]")
							Expect(execute("argspec $args set ((1 2 3))")).To(Equal(OK(NIL)))
							Expect(evaluate("get a")).To(Equal(evaluate("list (1 2 3)")))
							Expect(execute("argspec $args set (value)")).To(Equal(
								ERROR("invalid list"),
							))
						})
						Specify("optional", func() {
							evaluate("set args [argspec ( (list ?a) )]")
							Expect(execute("argspec $args set ()")).To(Equal(OK(NIL)))
							Expect(execute("get a")).To(Equal(
								ERROR(`cannot get "a": no such variable`),
							))
							Expect(execute("argspec $args set ((1 2 3))")).To(Equal(OK(NIL)))
							Expect(evaluate("get a")).To(Equal(evaluate("list (1 2 3)")))
							Expect(execute("argspec $args set (value)")).To(Equal(
								ERROR("invalid list"),
							))
						})
						Specify("default", func() {
							evaluate("set args [argspec ( (list ?a ()) )]")
							Expect(execute("argspec $args set ()")).To(Equal(OK(NIL)))
							Expect(evaluate("get a")).To(Equal(evaluate("list ()")))
							Expect(execute("argspec $args set ((1 2 3))")).To(Equal(OK(NIL)))
							Expect(evaluate("get a")).To(Equal(evaluate("list (1 2 3)")))
							Expect(execute("argspec $args set (value)")).To(Equal(
								ERROR("invalid list"),
							))
						})
					})
					Describe("tuple prefix", func() {
						Specify("required", func() {
							evaluate("set args [argspec ( ((0 <) a) )]")
							Expect(execute("argspec $args set (1)")).To(Equal(OK(NIL)))
							Expect(evaluate("get a")).To(Equal(TRUE))
							Expect(execute("argspec $args set (-1)")).To(Equal(OK(NIL)))
							Expect(evaluate("get a")).To(Equal(FALSE))
							Expect(execute("argspec $args set (value)")).To(Equal(
								ERROR(`invalid number "value"`),
							))
						})
						Specify("optional", func() {
							evaluate("set args [argspec ( ((0 <) ?a) )]")
							Expect(execute("argspec $args set ()")).To(Equal(OK(NIL)))
							Expect(execute("get a")).To(Equal(
								ERROR(`cannot get "a": no such variable`),
							))
							Expect(execute("argspec $args set (1)")).To(Equal(OK(NIL)))
							Expect(evaluate("get a")).To(Equal(TRUE))
							Expect(execute("argspec $args set (-1)")).To(Equal(OK(NIL)))
							Expect(evaluate("get a")).To(Equal(FALSE))
							Expect(execute("argspec $args set (value)")).To(Equal(
								ERROR(`invalid number "value"`),
							))
						})
						Specify("default", func() {
							evaluate("set args [argspec ( ((0 <) ?a 1) )]")
							Expect(execute("argspec $args set ()")).To(Equal(OK(NIL)))
							Expect(evaluate("get a")).To(Equal(TRUE))
							Expect(execute("argspec $args set (1)")).To(Equal(OK(NIL)))
							Expect(evaluate("get a")).To(Equal(TRUE))
							Expect(execute("argspec $args set (-1)")).To(Equal(OK(NIL)))
							Expect(evaluate("get a")).To(Equal(FALSE))
							Expect(execute("argspec $args set (value)")).To(Equal(
								ERROR(`invalid number "value"`),
							))
						})
					})

					Describe("Exceptions", func() {
						Specify("unexpected result", func() {
							Expect(execute("argspec ( (eval a) ) set ({return})")).To(Equal(
								ERROR("unexpected return"),
							))
							Expect(execute("argspec ( (eval a) ) set ({yield})")).To(Equal(
								ERROR("unexpected yield"),
							))
							Expect(execute("argspec ( (eval a) ) set ({break})")).To(Equal(
								ERROR("unexpected break"),
							))
							Expect(execute("argspec ( (eval a) ) set ({continue})")).To(Equal(
								ERROR("unexpected continue"),
							))
						})
						Specify("wrong arity", func() {
							evaluate("macro guard0 {} {}")
							evaluate("macro guard2 {a b} {}")
							Expect(execute("argspec ( (guard0 a) ) set (val)")).To(Equal(
								ERROR(`wrong # args: should be "guard0"`),
							))
							Expect(execute("argspec ( (guard2 a) ) set (val)")).To(Equal(
								ERROR(`wrong # args: should be "guard2 a b"`),
							))
						})
					})
				})
			})

			Describe("Exceptions", func() {
				Specify("empty argument name", func() {
					Expect(execute(`argspec ("")`)).To(Equal(ERROR("empty argument name")))
					Expect(execute("argspec (?)")).To(Equal(ERROR("empty argument name")))
					Expect(execute(`argspec ((""))`)).To(Equal(
						ERROR("empty argument name"),
					))
					Expect(execute("argspec ((?))")).To(Equal(ERROR("empty argument name")))
				})
				Specify("invalid argument name", func() {
					Expect(execute("argspec ([])")).To(Equal(
						ERROR("invalid argument name"),
					))
					Expect(execute("argspec (([]))")).To(Equal(
						ERROR("invalid argument name"),
					))
				})
				Specify("duplicate arguments", func() {
					Expect(execute("argspec (a a)")).To(Equal(
						ERROR(`duplicate argument "a"`),
					))
					Expect(execute("argspec ((?a def) a)")).To(Equal(
						ERROR(`duplicate argument "a"`),
					))
					Expect(execute("argspec (a (?a def))")).To(Equal(
						ERROR(`duplicate argument "a"`),
					))
				})
				Specify("empty argument specifier", func() {
					Expect(execute("argspec (())")).To(Equal(
						ERROR("empty argument specifier"),
					))
					Expect(execute("argspec ({})")).To(Equal(
						ERROR("empty argument specifier"),
					))
				})
				Specify("too many specifiers", func() {
					Expect(execute("argspec ((a b c d))")).To(Equal(
						ERROR(`too many specifiers for argument "a"`),
					))
					Expect(execute("argspec ({a b c d})")).To(Equal(
						ERROR(`too many specifiers for argument "a"`),
					))
				})
				Specify("non-optional parameter with guard and default", func() {
					Expect(execute("argspec ((a b c))")).To(Equal(
						ERROR(`default argument "b" must be optional`),
					))
					Expect(execute("argspec ({a b c})")).To(Equal(
						ERROR(`default argument "b" must be optional`),
					))
				})
			})
		})

		Describe("Option specifications", func() {
			Describe("Required options", func() {
				Specify("value", func() {
					value := evaluate("argspec (-o a)").(ArgspecValue)
					Expect(evaluate("argspec {-o a}")).To(Equal(value))
					Expect(value.Argspec).To(And(
						HaveField("NbRequired", uint(2)),
						HaveField("NbOptional", uint(0)),
						HaveField("HasRemainder", false),
					))
					Expect(value.Argspec.Args).To(Equal([]Argument{
						{
							Name:   "a",
							Type:   ArgumentType_REQUIRED,
							Option: &Option{Names: []string{"-o"}, Type: OptionType_OPTION},
						},
					}))
				})
				Specify("usage", func() {
					Expect(evaluate("argspec (-o a) usage")).To(Equal(STR("-o a")))
					Expect(evaluate("argspec ((-o --opt) a) usage")).To(Equal(
						STR("-o|--opt a"),
					))
				})
				Describe("set", func() {
					Specify("one", func() {
						evaluate("argspec (-o a) set (-o val)")
						Expect(evaluate("get a")).To(Equal(STR("val")))
					})
					Specify("two", func() {
						evaluate("argspec (-o a -p b) set (-o val1 -p val2)")
						Expect(evaluate("get a")).To(Equal(STR("val1")))
						Expect(evaluate("get b")).To(Equal(STR("val2")))
					})
					Specify("out of order", func() {
						evaluate("argspec (-o a -p b) set (-p val1 -o val2)")
						Expect(evaluate("get a")).To(Equal(STR("val2")))
						Expect(evaluate("get b")).To(Equal(STR("val1")))
					})
					Specify("prefix", func() {
						evaluate("argspec (-o a -p b c) set (-o val1 -p val2 val3)")
						Expect(evaluate("get a")).To(Equal(STR("val1")))
						Expect(evaluate("get b")).To(Equal(STR("val2")))
						Expect(evaluate("get c")).To(Equal(STR("val3")))
					})
					Specify("suffix", func() {
						evaluate("argspec (a -o b -p c) set (val1 -o val2 -p val3)")
						Expect(evaluate("get a")).To(Equal(STR("val1")))
						Expect(evaluate("get b")).To(Equal(STR("val2")))
						Expect(evaluate("get c")).To(Equal(STR("val3")))
					})
					Specify("infix", func() {
						evaluate("argspec (a -o b -p c d) set (val1 -o val2 -p val3 val4)")
						Expect(evaluate("get a")).To(Equal(STR("val1")))
						Expect(evaluate("get b")).To(Equal(STR("val2")))
						Expect(evaluate("get c")).To(Equal(STR("val3")))
						Expect(evaluate("get d")).To(Equal(STR("val4")))
					})
					Specify("complex case", func() {
						evaluate(`argspec (
					  a 
					  -o1 b -o2 c 
					  d 
					  -o3 e
					  f g h 
					  -o4 i -o5 j -o6 k
					  l
					) set (
					  val1 
					  -o2 val2 -o1 val3
					  val4 
					  -o3 val5 
					  val6 val7 val8
					  -o6 val9 -o4 val10 -o5 val11
					  val12
					)`)
						Expect(evaluate("get a")).To(Equal(STR("val1")))
						Expect(evaluate("get b")).To(Equal(STR("val3")))
						Expect(evaluate("get c")).To(Equal(STR("val2")))
						Expect(evaluate("get d")).To(Equal(STR("val4")))
						Expect(evaluate("get e")).To(Equal(STR("val5")))
						Expect(evaluate("get f")).To(Equal(STR("val6")))
						Expect(evaluate("get g")).To(Equal(STR("val7")))
						Expect(evaluate("get h")).To(Equal(STR("val8")))
						Expect(evaluate("get i")).To(Equal(STR("val10")))
						Expect(evaluate("get j")).To(Equal(STR("val11")))
						Expect(evaluate("get k")).To(Equal(STR("val9")))
						Expect(evaluate("get l")).To(Equal(STR("val12")))
					})
				})
			})

			Describe("Optional options", func() {
				Specify("value", func() {
					value := evaluate("argspec (-o ?a)").(ArgspecValue)
					Expect(evaluate("argspec {-o ?a}")).To(Equal(value))
					Expect(value.Argspec).To(And(
						HaveField("NbRequired", uint(0)),
						HaveField("NbOptional", uint(0)),
						HaveField("HasRemainder", false),
					))
					Expect(value.Argspec.Args).To(Equal([]Argument{
						{
							Name:   "a",
							Type:   ArgumentType_OPTIONAL,
							Option: &Option{Names: []string{"-o"}, Type: OptionType_OPTION},
						},
					}))
				})
				Specify("usage", func() {
					Expect(evaluate("argspec (-o ?a) usage")).To(Equal(STR("?-o a?")))
					Expect(evaluate("argspec ((-o --opt) ?a) usage")).To(Equal(
						STR("?-o|--opt a?"),
					))
				})
				Describe("set", func() {
					Specify("zero", func() {
						evaluate("argspec (-o ?a) set ()")
						Expect(execute("get a")).To(Equal(
							ERROR(`cannot get "a": no such variable`),
						))
					})
					Specify("default", func() {
						evaluate("argspec (-o (?a def)) set ()")
						Expect(evaluate("get a")).To(Equal(STR("def")))
					})
					Specify("one", func() {
						evaluate("argspec (-o ?a) set (-o val)")
						Expect(evaluate("get a")).To(Equal(STR("val")))
					})
					Specify("two", func() {
						evaluate("argspec (-o ?a -p ?b) set (-o val)")
						Expect(evaluate("get a")).To(Equal(STR("val")))
						Expect(execute("get b")).To(Equal(
							ERROR(`cannot get "b": no such variable`),
						))
					})
				})
			})

			Describe("Flags", func() {
				Specify("value", func() {
					value := evaluate("argspec (?-o ?a)").(ArgspecValue)
					Expect(evaluate("argspec {?-o ?a}")).To(Equal(value))
					Expect(value.Argspec).To(And(
						HaveField("NbRequired", uint(0)),
						HaveField("NbOptional", uint(0)),
						HaveField("HasRemainder", false),
					))
					Expect(value.Argspec.Args).To(Equal([]Argument{
						{
							Name:   "a",
							Type:   ArgumentType_OPTIONAL,
							Option: &Option{Names: []string{"-o"}, Type: OptionType_FLAG},
						},
					}))
				})
				Specify("usage", func() {
					Expect(evaluate("argspec (?-o ?a) usage")).To(Equal(STR("?-o?")))
					Expect(evaluate("argspec ((?-o ?--opt) ?a) usage")).To(Equal(
						STR("?-o|--opt?"),
					))
				})
				Describe("set", func() {
					Specify("zero", func() {
						evaluate("argspec (?-o ?a) set ()")
						Expect(evaluate("get a")).To(Equal(FALSE))
					})
					Specify("one", func() {
						evaluate("argspec (?-o ?a) set (-o)")
						Expect(evaluate("get a")).To(Equal(TRUE))
					})
					Specify("two", func() {
						evaluate("argspec (?-o ?a ?-p ?b) set (-p)")
						Expect(evaluate("get a")).To(Equal(FALSE))
						Expect(evaluate("get b")).To(Equal(TRUE))
					})
				})

				Describe("Exceptions", func() {
					Specify("non-optional argument", func() {
						Expect(execute("argspec (?-o a)")).To(Equal(
							ERROR(`argument for flag "-o" must be optional`),
						))
					})
				})
			})

			Describe("Exceptions", func() {
				Specify("missing argument", func() {
					Expect(execute("argspec (-a)")).To(Equal(
						ERROR(`missing argument for option "-a"`),
					))
					Expect(execute("argspec ((-a --argument))")).To(Equal(
						ERROR(`missing argument for option "-a|--argument"`),
					))
					Expect(execute("argspec ({-a --arg --argument-name})")).To(Equal(
						ERROR(`missing argument for option "-a|--arg|--argument-name"`),
					))
					Expect(execute("argspec ([list (-a --argument)])")).To(Equal(
						ERROR(`missing argument for option "-a|--argument"`),
					))
				})
				Specify("incompatible aliases", func() {
					Expect(execute("argspec ((?-a --argument) a)")).To(Equal(
						ERROR(`incompatible aliases for option "-a"`),
					))
					Expect(execute("argspec ((-a ?--argument) a)")).To(Equal(
						ERROR(`incompatible aliases for option "-a"`),
					))
				})
				Specify("duplicate options", func() {
					Expect(execute("argspec (-o a -o b)")).To(Equal(
						ERROR(`duplicate option "-o"`),
					))
					Expect(execute("argspec ((-o --opt) a --opt b)")).To(Equal(
						ERROR(`duplicate option "--opt"`),
					))
					Expect(execute("argspec ((-o -o) a --opt b)")).To(Equal(
						ERROR(`duplicate option "-o"`),
					))
				})
				Specify("remainder before options", func() {
					Expect(execute("argspec (*args -o o)")).To(Equal(
						ERROR("cannot use remainder argument before options"),
					))
					Expect(execute("argspec (*args -o ?o)")).To(Equal(
						ERROR("cannot use remainder argument before options"),
					))
					Expect(execute("argspec (*args ?-o ?o)")).To(Equal(
						ERROR("cannot use remainder argument before options"),
					))
				})
				Specify("option terminator", func() {
					Expect(execute("argspec (--)")).To(Equal(
						ERROR(`cannot use option terminator as option name`),
					))
					Expect(execute("argspec (?--)")).To(Equal(
						ERROR(`cannot use option terminator as option name`),
					))
					Expect(execute("argspec ((-a --))")).To(Equal(
						ERROR(`cannot use option terminator as option name`),
					))
				})
			})
		})

		Describe("Evaluation order", func() {
			Specify("required positionals only", func() {
				evaluate("set s [argspec (a b c)]")
				Expect(execute("argspec $s set ()")).To(Equal(
					ERROR(`wrong # values: should be "a b c"`),
				))
				Expect(execute("argspec $s set (1 2 3 4)")).To(Equal(
					ERROR(`wrong # values: should be "a b c"`),
				))
				Expect(evaluate("argspec $s set (1 2 3); get (a b c)")).To(Equal(
					evaluate("idem (1 2 3)"),
				))
			})
			Specify("required options only", func() {
				evaluate("set s [argspec (-a a -b b -c c)]")
				Expect(execute("argspec $s set ()")).To(Equal(
					ERROR(`wrong # values: should be "-a a -b b -c c"`),
				))
				Expect(execute("argspec $s set (-a 1)")).To(Equal(
					ERROR(`wrong # values: should be "-a a -b b -c c"`),
				))
				Expect(execute("argspec $s set (-a 1 -b 2 -c 3 4)")).To(Equal(
					ERROR("extra values after arguments"),
				))
				Expect(evaluate("argspec $s set (-a 1 -b 2 -c 3); get (a b c)")).To(Equal(
					evaluate("idem (1 2 3)"),
				))
				Expect(evaluate("argspec $s set (-b 1 -c 2 -a 3); get (a b c)")).To(Equal(
					evaluate("idem (3 1 2)"),
				))
			})
			Specify("required and optional options", func() {
				evaluate("set s [argspec (-a a -b ?b ?-c ?c -d ?d -e e)]")
				evaluate(
					"macro cleanup {} {list (a b c d e) foreach v {catch {unset $v}}}",
				)
				Expect(execute("argspec $s set ()")).To(Equal(
					ERROR(`wrong # values: should be "-a a ?-b b? ?-c? ?-d d? -e e"`),
				))
				Expect(execute("argspec $s set (-a 1)")).To(Equal(
					ERROR(`wrong # values: should be "-a a ?-b b? ?-c? ?-d d? -e e"`),
				))
				Expect(execute("argspec $s set (-a 1 -b 2 -c -d 3 -e 4 5)")).To(Equal(
					ERROR("extra values after arguments"),
				))
				Expect(
					evaluate("argspec $s set (-a 1 -b 2 -c -d 3 -e 4); get (a b c d e)"),
				).To(Equal(evaluate("idem (1 2 [true] 3 4)")))
				Expect(
					evaluate(
						"cleanup; argspec $s set (-a 1 -e 2); list ($a [exists b] $c [exists d] $e)",
					),
				).To(Equal(evaluate("list (1 [false] [false] [false] 2)")))
				Expect(execute("argspec $s set (-b 1 -d 2)")).To(Equal(
					ERROR(`missing value for option "-a"`),
				))
				Expect(execute("argspec $s set (-b 1 -a 2 -d 3)")).To(Equal(
					ERROR(`missing value for option "-e"`),
				))
				Expect(execute("argspec $s set (-c -a 1 -d 2)")).To(Equal(
					ERROR(`missing value for option "-e"`),
				))
			})
			Specify("required positional and option groups", func() {
				evaluate("set s [argspec (a b -c c d -e e -f f -g g h i j -k k)]")
				Expect(execute("argspec $s set ()")).To(Equal(
					ERROR(
						`wrong # values: should be "a b -c c d -e e -f f -g g h i j -k k"`,
					),
				))
				Expect(execute("argspec $s set (a b -c)")).To(Equal(
					ERROR(
						`wrong # values: should be "a b -c c d -e e -f f -g g h i j -k k"`,
					),
				))
				Expect(
					evaluate(
						"argspec $s set (1 2 -c 3 4 -e 5 -f 6 -g 7 8 9 10 -k 11); get (a b c d e f g h i j k)",
					),
				).To(Equal(evaluate("idem (1 2 3 4 5 6 7 8 9 10 11)")))
				Expect(
					evaluate(
						"argspec $s set (1 2 -c 3 4 -e 5 -g 6 -f 7 8 9 10 -k 11); get (a b c d e f g h i j k)",
					),
				).To(Equal(evaluate("idem (1 2 3 4 5 7 6 8 9 10 11)")))
				Expect(
					execute("argspec $s set (1 2 -e 3 4 -c 5 -f 6 -g 7 8 9 10 -k 11)"),
				).To(Equal(ERROR(`unexpected option "-e"`)))
				Expect(
					execute("argspec $s set (1 2 -c 3 4 -e 5 -f 6 -f 7 8 9 10 -k 11)"),
				).To(Equal(ERROR(`duplicate values for option "-f"`)))
			})
			Specify("optional arguments", func() {
				evaluate("set s [argspec (?a ?b ?-c ?c -d ?d ?e ?f)]")
				evaluate(
					"macro cleanup {} {list (a b c d e f) foreach v {catch {unset $v}}}",
				)
				Expect(
					evaluate(
						"argspec $s set (); list ([exists a] [exists b] $c [exists d] [exists e] [exists f])",
					),
				).To(Equal(
					evaluate("list ([false] [false] [false] [false] [false] [false])"),
				))
				Expect(
					evaluate(
						"cleanup; argspec $s set (1); list ($a [exists b] $c [exists d] [exists e] [exists f])",
					),
				).To(Equal(evaluate("list (1 [false] [false] [false] [false] [false])")))
				Expect(
					evaluate(
						"cleanup; argspec $s set (1 2); list ($a $b $c [exists d] [exists e] [exists f])",
					),
				).To(Equal(evaluate("list (1 2 [false] [false] [false] [false])")))
				Expect(
					evaluate(
						"cleanup; argspec $s set (1 2 3); list ($a $b $c [exists d] $e [exists f])",
					),
				).To(Equal(evaluate("list (1 2 [false] [false] 3 [false])")))
				Expect(
					evaluate(
						"cleanup; argspec $s set (1 2 3 4); list ($a $b $c [exists d] $e $f)",
					),
				).To(Equal(evaluate("list (1 2 [false] [false] 3 4)")))
				Expect(execute("argspec $s set (1 2 3 4 5)")).To(Equal(
					ERROR("extra values after arguments"),
				))
				Expect(
					evaluate(
						"cleanup; argspec $s set (1 2 -c 3 4); list ($a $b $c [exists d] $e $f)",
					),
				).To(Equal(evaluate("list (1 2 [true] [false] 3 4)")))
				Expect(
					evaluate(
						"cleanup; argspec $s set (1 2 -d 3 4); list ($a $b $c $d $e [exists f])",
					),
				).To(Equal(evaluate("list (1 2 [false] 3 4 [false])")))
				Expect(execute("argspec $s set (1 -d 2 3 4)")).To(Equal(
					ERROR("extra values after arguments"),
				))
				Expect(
					evaluate(
						"cleanup; argspec $s set (1 2 -c -d 3 4); list ($a $b $c $d $e [exists f])",
					),
				).To(Equal(evaluate("list (1 2 [true] 3 4 [false])")))
				Expect(
					evaluate(
						"cleanup; argspec $s set (1 2 -c -d 3 4 5); list ($a $b $c $d $e $f)",
					),
				).To(Equal(evaluate("list (1 2 [true] 3 4 5)")))
			})
			Specify("remainder argument", func() {
				evaluate("set s [argspec (?a ?-b ?b -c ?c -d d *args e)]")
				evaluate(
					"macro cleanup {} {list (a b c d e args) foreach v {catch {unset $v}}}",
				)
				Expect(execute("argspec $s set ()")).To(Equal(
					ERROR(`wrong # values: should be "?a? ?-b? ?-c c? -d d ?args ...? e"`),
				))
				Expect(
					evaluate(
						"cleanup; argspec $s set (-d 1 2); list ([exists a] $b [exists c] $d $e $args)",
					),
				).To(Equal(evaluate("list ([false] [false] [false] 1 2 ())")))
				Expect(execute("argspec $s set (-d 1 2 3)")).To(Equal(
					ERROR(`unknown option "1"`),
				))
				Expect(
					evaluate(
						"cleanup; argspec $s set (1 -d 2 3 4); list ($a $b [exists c] $d $e $args)",
					),
				).To(Equal(evaluate("list (1 [false] [false] 2 4 (3))")))
				Expect(
					evaluate(
						"cleanup; argspec $s set (1 -b -d 2 3); list ($a $b [exists c] $d $e $args)",
					),
				).To(Equal(evaluate("list (1 [true] [false] 2 3 ())")))
				Expect(
					evaluate(
						"cleanup; argspec $s set (1 -b -d 2 3 4); list ($a $b [exists c] $d $e $args)",
					),
				).To(Equal(evaluate("list (1 [true] [false] 2 4 (3))")))
				Expect(
					evaluate(
						"cleanup; argspec $s set (1 -b -c 2 -d 3 4); list ($a $b $c $d $e $args)",
					),
				).To(Equal(evaluate("list (1 [true] 2 3 4 ())")))
				Expect(
					evaluate(
						"cleanup; argspec $s set (1 -b -c 2 -d 3 4 5 6); list ($a $b $c $d $e $args)",
					),
				).To(Equal(evaluate("list (1 [true] 2 3 6 (4 5))")))
			})
			Specify("option terminator", func() {
				evaluate("set s [argspec (-a a -b b c -d ?d ?-e ?e *args)]")
				evaluate(
					"macro cleanup {} {list (a b c d e args) foreach v {catch {unset $v}}}",
				)
				Expect(execute("argspec $s set ()")).To(Equal(
					ERROR(
						`wrong # values: should be "-a a -b b c ?-d d? ?-e? ?args ...?"`,
					),
				))
				Expect(
					evaluate(
						"cleanup; argspec $s set (-a 1 -b 2 3 -d 4 -e); get (a b c d e args)",
					),
				).To(Equal(evaluate("idem (1 2 3 4 [true] ())")))
				Expect(execute("argspec $s set (-- -a 1 -b 2 3 -d 4 -e)")).To(Equal(
					ERROR("unexpected option terminator"),
				))
				Expect(execute("argspec $s set (-a 1 -- -b 2 3 -d 4 -e)")).To(Equal(
					ERROR("unexpected option terminator"),
				))
				Expect(
					evaluate(
						"cleanup; argspec $s set (-a 1 -b 2 -- 3 -d 4 -e); get (a b c d e args)",
					),
				).To(Equal(evaluate("idem (1 2 3 4 [true] ())")))
				Expect(
					evaluate(
						"cleanup; argspec $s set (-a 1 -b 2 3 -- 4 5 6); list ($a $b $c [exists d] $e $args)",
					),
				).To(Equal(evaluate("list (1 2 3 [false] [false] (4 5 6))")))
				Expect(
					evaluate(
						"cleanup; argspec $s set (-a 1 -b 2 3 -d 4 -e --); get (a b c d e args)",
					),
				).To(Equal(evaluate("idem (1 2 3 4 [true] ())")))
			})
			Specify("complex case", func() {
				evaluate(
					"set s [argspec (a ?b ?-c ?c -d ?d -e e -f ?f -g g ?h *args ?i j k)]",
				)
				evaluate(
					"macro cleanup {} {list (a b c d e f g h i j k args) foreach v {catch {unset $v}}}",
				)
				Expect(execute("argspec $s set ()")).To(Equal(
					ERROR(
						`wrong # values: should be "a ?b? ?-c? ?-d d? -e e ?-f f? -g g ?h? ?args ...? ?i? j k"`,
					),
				))
				Expect(
					evaluate(
						"cleanup; argspec $s set (1 2 -c -d 3 -e 4 -f 5 -g 6 7 8 9 10 11 12 13); get (a b c d e f g h i j k args)",
					),
				).To(Equal(evaluate("idem (1 2 [true] 3 4 5 6 7 11 12 13 (8 9 10))")))
				Expect(
					evaluate(
						"cleanup; argspec $s set (1 -e 2 -g 3 4 5); get (a c e g j k args)",
					),
				).To(Equal(evaluate("idem (1 [false] 2 3 4 5 ())")))
				Expect(
					evaluate(
						"cleanup; argspec $s set (1 2 -e 3 -g 4 5 6); get (a b c e g j k args)",
					),
				).To(Equal(evaluate("idem (1 2 [false] 3 4 5 6 ())")))
				Expect(
					evaluate(
						"cleanup; argspec $s set (1 2 -e 3 -g 4 5 6); get (a b c e g j k args)",
					),
				).To(Equal(evaluate("idem (1 2 [false] 3 4 5 6 ())")))
				Expect(
					evaluate(
						"cleanup; argspec $s set (1 2 -c -e 3 -g 4 5 6); get (a b c e g j k args)",
					),
				).To(Equal(evaluate("idem (1 2 [true] 3 4 5 6 ())")))
				Expect(
					evaluate(
						"cleanup; argspec $s set (1 2 -d 3 -e 4 -g 5 6 7); get (a b c d e g j k args)",
					),
				).To(Equal(evaluate("idem (1 2 [false] 3 4 5 6 7 ())")))
				Expect(
					evaluate(
						"cleanup; argspec $s set (1 2 -d 3 -e 4 -g 5 6 7 8); get (a b c d e g h j k args)",
					),
				).To(Equal(evaluate("idem (1 2 [false] 3 4 5 6 7 8 ())")))
				Expect(
					evaluate(
						"cleanup; argspec $s set (1 2 -d 3 -e 4 -g 5 6 7 8 9); get (a b c d e g h i j k args)",
					),
				).To(Equal(evaluate("idem (1 2 [false] 3 4 5 6 7 8 9 ())")))
				Expect(
					evaluate(
						"argspec $s set (1 2 -d 3 -e 4 -g 5 6 7 8 9 10); get (a b c d e g h i j k args)",
					),
				).To(Equal(evaluate("idem (1 2 [false] 3 4 5 6 8 9 10 (7))")))
			})
		})

		Describe("Subcommands", func() {
			Describe("`subcommands`", func() {
				Specify("usage", func() {
					Expect(evaluate("help argspec () subcommands")).To(Equal(
						STR("argspec value subcommands"),
					))
				})

				It("should return list of subcommands", func() {
					// Expect(evaluate("argspec {} subcommands")).To(Equal(
					// 	TODO specify order?
					// 	evaluate("list (subcommands usage set)"),
					// ))
				})

				Describe("Exceptions", func() {
					Specify("wrong arity", func() {
						Expect(execute("argspec {} subcommands a")).To(Equal(
							ERROR(`wrong # args: should be "argspec value subcommands"`),
						))
						Expect(execute("help argspec {} subcommands a")).To(Equal(
							ERROR(`wrong # args: should be "argspec value subcommands"`),
						))
					})
				})
			})

			Describe("`usage`", func() {
				It("should return a usage string with argument names", func() {
					Expect(evaluate("argspec {a b ?c *} usage")).To(Equal(
						STR("a b ?c? ?arg ...?"),
					))
				})

				Describe("Exceptions", func() {
					Specify("wrong arity", func() {
						Expect(execute("argspec {} usage a")).To(Equal(
							ERROR(`wrong # args: should be "argspec value usage"`),
						))
					})
				})
			})

			Describe("`set`", func() {
				It("should return nil", func() {
					Expect(evaluate("argspec {} set ()")).To(Equal(NIL))
				})
				It("should set argument variables in the caller scope", func() {
					evaluate("argspec {a} set (val)")
					Expect(evaluate("get a")).To(Equal(STR("val")))
				})
				It("should enforce minimum number of arguments", func() {
					Expect(execute("argspec {a} set ()")).To(Equal(
						ERROR(`wrong # values: should be "a"`),
					))
					Expect(execute("argspec {a ?b} set ()")).To(Equal(
						ERROR(`wrong # values: should be "a ?b?"`),
					))
					Expect(execute("argspec {?a b c} set (val)")).To(Equal(
						ERROR(`wrong # values: should be "?a? b c"`),
					))
					Expect(execute("argspec {a *b c} set (val)")).To(Equal(
						ERROR(`wrong # values: should be "a ?b ...? c"`),
					))
				})
				It("should enforce maximum number of arguments", func() {
					Expect(execute("argspec {} set (val1)")).To(Equal(
						ERROR(`wrong # values: should be ""`),
					))
					Expect(execute("argspec {a} set (val1 val2)")).To(Equal(
						ERROR(`wrong # values: should be "a"`),
					))
					Expect(execute("argspec {a ?b} set (val1 val2 val3)")).To(Equal(
						ERROR(`wrong # values: should be "a ?b?"`),
					))
				})
				It("should set required attributes first", func() {
					evaluate("argspec {?a b ?c} set (val)")
					Expect(evaluate("get b")).To(Equal(STR("val")))
				})
				It("should skip missing optional attributes", func() {
					evaluate("argspec {?a b (?c def)} set (val)")
					Expect(execute("get a")).To(Equal(
						ERROR(`cannot get "a": no such variable`),
					))
					Expect(evaluate("get b")).To(Equal(STR("val")))
					Expect(evaluate("get c")).To(Equal(STR("def")))
				})
				It("should set optional attributes in order", func() {
					evaluate("argspec {(?a def) b ?c} set (val1 val2)")
					Expect(evaluate("get a")).To(Equal(STR("val1")))
					Expect(evaluate("get b")).To(Equal(STR("val2")))
					Expect(execute("get c")).To(Equal(
						ERROR(`cannot get "c": no such variable`),
					))
				})
				It("should set remainder after optional attributes", func() {
					evaluate("argspec {?a *b c} set (val1 val2)")
					Expect(evaluate("get a")).To(Equal(STR("val1")))
					Expect(evaluate("get b")).To(Equal(TUPLE([]core.Value{})))
					Expect(evaluate("get c")).To(Equal(STR("val2")))
				})
				It("should set all present attributes in order", func() {
					evaluate("argspec {?a *b c} set (val1 val2 val3 val4)")
					Expect(evaluate("get a")).To(Equal(STR("val1")))
					Expect(evaluate("get b")).To(Equal(TUPLE([]core.Value{STR("val2"), STR("val3")})))
					Expect(evaluate("get c")).To(Equal(STR("val4")))
				})

				Describe("Exceptions", func() {
					Specify("wrong arity", func() {
						Expect(execute("argspec {} set")).To(Equal(
							ERROR(`wrong # args: should be "argspec value set values"`),
						))
						Expect(execute("argspec {} set a b")).To(Equal(
							ERROR(`wrong # args: should be "argspec value set values"`),
						))
						Expect(execute("help argspec {} set a b")).To(Equal(
							ERROR(`wrong # args: should be "argspec value set values"`),
						))
					})
				})
			})

			Describe("Exceptions", func() {
				Specify("unknown subcommand", func() {
					Expect(execute("argspec () unknownSubcommand")).To(Equal(
						ERROR(`unknown subcommand "unknownSubcommand"`),
					))
				})
				Specify("invalid subcommand name", func() {
					Expect(execute("argspec () []")).To(Equal(
						ERROR("invalid subcommand name"),
					))
				})
			})
		})

		Describe("Examples", func() {
			Specify("Currying and encapsulation", func() {
				example([]exampleSpec{
					{
						script: "set l (argspec {a b ?c *})",
					},
					{
						script: "$l",
						result: evaluate("argspec {a b ?c *}"),
					},
					{
						script: "$l usage",
						result: STR("a b ?c? ?arg ...?"),
					},
					{
						script: "$l set (val1 val2 val3); get (a b c)",
						result: TUPLE([]core.Value{STR("val1"), STR("val2"), STR("val3")}),
					},
				})
			})
			Specify("Argument type guard", func() {
				example([]exampleSpec{
					{
						script: "macro usage ( (argspec a) ) {argspec $a usage}",
					},
					{
						script: "usage {a b ?c *}",
						result: STR("a b ?c? ?arg ...?"),
					},
					{
						script: "usage invalidValue",
						result: ERROR("invalid argument list"),
					},
				})
			})
		})

		Describe("Ensemble command", func() {
			It("should return its ensemble metacommand when called with no argument", func() {
				Expect(evaluate("argspec").Type()).To(Equal(core.ValueType_COMMAND))
			})
			It("should be extensible", func() {
				evaluate(`
					[argspec] eval {
						macro foo {value} {idem bar}
					}
				`)
				Expect(evaluate("argspec (a b c) foo")).To(Equal(STR("bar")))
			})
			It("should support help for custom subcommands", func() {
				evaluate(`
					[argspec] eval {
						macro foo {value a b} {idem bar}
					}
				`)
				Expect(evaluate("help argspec (a b c) foo")).To(Equal(
					STR("argspec value foo a b"),
				))
				Expect(execute("help argspec (a b c) foo 1 2 3")).To(Equal(
					ERROR(`wrong # args: should be "argspec value foo a b"`),
				))
			})

			Describe("Examples", func() {
				Specify("Adding a `help` subcommand", func() {
					example([]exampleSpec{
						{
							script: `
								[argspec] eval {
									macro help {value prefix} {
										idem "$prefix [argspec $value usage]"
									}
								}
							`,
						},
						{
							script: "argspec {a b ?c *} help foo",
							result: STR("foo a b ?c? ?arg ...?"),
						},
					})
				})
			})
		})
	})
})
