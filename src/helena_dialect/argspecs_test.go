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
		parser = &core.Parser{}
	}

	type exampleSpec struct {
		script string
		result core.Value
	}
	example := func(spec exampleSpec) {
		value := evaluate(spec.script)
		if spec.result != nil {
			Expect(value).To(Equal(spec.result))
		}
	}

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
			// It("should convert lists to argspecs", func() {
			// 	example(exampleSpec{
			// 		script: "argspec [list {a b c}]",
			// 		result: evaluate("argspec (a b c)"),
			// 	})
			// })

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
				//         Specify("usage", func() {
				//           Expect(evaluate("argspec () usage")).To(Equal(STR(""));
				//         });
				//         Specify("set", func() {
				//           evaluate("argspec () set ()");
				//           Expect(rootScope.context.variables).to.be.empty;
				//         });
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
					// Expect(value.Argspec.Args).To(Equal([]Argument{{Name: "a", Type: ArgumentType_REQUIRED}}))
				})
				// Specify("usage", func() {
				// 	Expect(evaluate("argspec (a) usage")).To(Equal(STR("a")))
				// })
				//         Specify("set", func() {
				//           evaluate("argspec (a) set (val1)");
				//           Expect(evaluate("get a")).To(Equal(STR("val1"));
				//           evaluate("argspec (a) set (val2)");
				//           Expect(evaluate("get a")).To(Equal(STR("val2"));
				//         });
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
				//         Specify("usage", func() {
				//           Expect(evaluate("argspec (a b) usage")).To(Equal(STR("a b"));
				//         });
				//         Specify("set", func() {
				//           evaluate("argspec {a b} set (val1 val2)");
				//           Expect(evaluate("get a")).To(Equal(STR("val1"));
				//           Expect(evaluate("get b")).To(Equal(STR("val2"));
				//         });
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
					//           Specify("usage", func() {
					//             Expect(evaluate("argspec (*) usage")).To(Equal(STR("?arg ...?"));
					//           });
					//           Describe("set", func() {
					//             Specify("zero", func() {
					//               evaluate("argspec (*) set ()");
					//               Expect(evaluate("get *")).To(Equal(TUPLE([]));
					//             });
					//             Specify("one", func() {
					//               evaluate("argspec (*) set (val)");
					//               Expect(evaluate("get *")).To(Equal(TUPLE([STR("val")]));
					//             });
					//             Specify("two", func() {
					//               evaluate("argspec (*) set (val1 val2)");
					//               Expect(evaluate("get *")).To(Equal(
					//                 TUPLE([STR("val1"), STR("val2")])
					//               );
					//             });
					//           });
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
					//           Specify("usage", func() {
					//             Expect(evaluate("argspec (*remainder) usage")).To(Equal(
					//               STR("?remainder ...?")
					//             );
					//           });
					//           Describe("set", func() {
					//             Specify("zero", func() {
					//               evaluate("argspec (*args) set ()");
					//               Expect(evaluate("get args")).To(Equal(TUPLE([]));
					//             });
					//             Specify("one", func() {
					//               evaluate("argspec (*args) set (val)");
					//               Expect(evaluate("get args")).To(Equal(TUPLE([STR("val")]));
					//             });
					//             Specify("two", func() {
					//               evaluate("argspec (*args) set (val1 val2)");
					//               Expect(evaluate("get args")).To(Equal(
					//                 TUPLE([STR("val1"), STR("val2")])
					//               );
					//             });
					//           });
				})

				// Describe("prefix", func() {
				// 	Specify("one", func() {
				// 		evaluate("argspec (* a) set (val)")
				// 		Expect(evaluate("get *")).To(Equal(TUPLE([]core.Value{})))
				// 		Expect(evaluate("get a")).To(Equal(STR("val")))
				// 	})
				// 	Specify("two", func() {
				// 		evaluate("argspec (* a) set (val1 val2)")
				// 		Expect(evaluate("get *")).To(Equal(TUPLE([]core.Value{STR("val1")})))
				// 		Expect(evaluate("get a")).To(Equal(STR("val2")))
				// 	})
				// 	Specify("three", func() {
				// 		evaluate("argspec (* a) set (val1 val2 val3)")
				// 		Expect(evaluate("get *")).To(Equal(TUPLE([]core.Value{STR("val1"), STR("val2")})))
				// 		Expect(evaluate("get a")).To(Equal(STR("val3")))
				// 	})
				// })
				//         Describe("infix", func() {
				//           Specify("two", func() {
				//             evaluate("argspec (a * b) set (val1 val2)");
				//             Expect(evaluate("get a")).To(Equal(STR("val1"));
				//             Expect(evaluate("get *")).To(Equal(TUPLE([]));
				//             Expect(evaluate("get b")).To(Equal(STR("val2"));
				//           });
				//           Specify("three", func() {
				//             evaluate("argspec (a * b) set (val1 val2 val3)");
				//             Expect(evaluate("get a")).To(Equal(STR("val1"));
				//             Expect(evaluate("get *")).To(Equal(TUPLE([STR("val2")]));
				//             Expect(evaluate("get b")).To(Equal(STR("val3"));
				//           });
				//           Specify("four", func() {
				//             evaluate("argspec (a * b) set (val1 val2 val3 val4)");
				//             Expect(evaluate("get a")).To(Equal(STR("val1"));
				//             Expect(evaluate("get *")).To(Equal(TUPLE([STR("val2"), STR("val3")]));
				//             Expect(evaluate("get b")).To(Equal(STR("val4"));
				//           });
				//         });
				//         Describe("suffix", func() {
				//           Specify("one", func() {
				//             evaluate("argspec (a *) set (val)");
				//             Expect(evaluate("get *")).To(Equal(TUPLE([]));
				//             Expect(evaluate("get a")).To(Equal(STR("val"));
				//             Expect(evaluate("get *")).To(Equal(TUPLE([]));
				//           });
				//           Specify("two", func() {
				//             evaluate("argspec (a *) set (val1 val2)");
				//             Expect(evaluate("get a")).To(Equal(STR("val1"));
				//             Expect(evaluate("get *")).To(Equal(TUPLE([STR("val2")]));
				//           });
				//           Specify("three", func() {
				//             evaluate("argspec (a *) set (val1 val2 val3)");
				//             Expect(evaluate("get a")).To(Equal(STR("val1"));
				//             Expect(evaluate("get *")).To(Equal(TUPLE([STR("val2"), STR("val3")]));
				//           });
				//         });

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
					//           Specify("usage", func() {
					//             Expect(evaluate("argspec (?a) usage")).To(Equal(STR("?a?"));
					//           });
					//           Describe("set", func() {
					//             Specify("zero", func() {
					//               evaluate("argspec ?a set ()");
					//               Expect(execute("get a")).To(Equal(
					//                 ERROR(`cannot get "a": no such variable`)
					//               );
					//             });
					//             Specify("one", func() {
					//               evaluate("argspec (?a) set (val)");
					//               Expect(evaluate("get a")).To(Equal(STR("val"));
					//             });
					//           });
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
					//           Specify("usage", func() {
					//             Expect(evaluate("argspec (?a ?b) usage")).To(Equal(STR("?a? ?b?"));
					//           });
					//           Describe("set", func() {
					//             Specify("zero", func() {
					//               evaluate("argspec (?a ?b) set ()");
					//               Expect(execute("get a")).To(Equal(
					//                 ERROR(`cannot get "a": no such variable`)
					//               );
					//               Expect(execute("get b")).To(Equal(
					//                 ERROR(`cannot get "b": no such variable`)
					//               );
					//             });
					//             Specify("one", func() {
					//               evaluate("argspec (?a ?b) set (val)");
					//               Expect(evaluate("get a")).To(Equal(STR("val"));
					//               Expect(execute("get b")).To(Equal(
					//                 ERROR(`cannot get "b": no such variable`)
					//               );
					//             });
					//             Specify("two", func() {
					//               evaluate("argspec (?a ?b) set (val1 val2)");
					//               Expect(evaluate("get a")).To(Equal(STR("val1"));
					//               Expect(evaluate("get b")).To(Equal(STR("val2"));
					//             });
					//           });
				})

				//         Describe("prefix", func() {
				//           Specify("one", func() {
				//             evaluate("argspec (?a b) set (val)");
				//             Expect(execute("get a")).To(Equal(
				//               ERROR(`cannot get "a": no such variable`)
				//             );
				//             Expect(evaluate("get b")).To(Equal(STR("val"));
				//           });
				//           Specify("two", func() {
				//             evaluate("argspec (?a b) set (val1 val2)");
				//             Expect(evaluate("get a")).To(Equal(STR("val1"));
				//             Expect(evaluate("get b")).To(Equal(STR("val2"));
				//           });
				//         });
				//         Describe("infix", func() {
				//           Specify("two", func() {
				//             evaluate("argspec (a ?b c) set (val1 val2)");
				//             Expect(evaluate("get a")).To(Equal(STR("val1"));
				//             Expect(execute("get b")).To(Equal(
				//               ERROR(`cannot get "b": no such variable`)
				//             );
				//             Expect(evaluate("get c")).To(Equal(STR("val2"));
				//           });
				//           Specify("three", func() {
				//             evaluate("argspec (a ?b c) set (val1 val2 val3)");
				//             Expect(evaluate("get a")).To(Equal(STR("val1"));
				//             Expect(evaluate("get b")).To(Equal(STR("val2"));
				//             Expect(evaluate("get c")).To(Equal(STR("val3"));
				//           });
				//         });
				//         Describe("suffix", func() {
				//           Specify("one", func() {
				//             evaluate("argspec (a ?b) set (val)");
				//             Expect(evaluate("get a")).To(Equal(STR("val"));
				//             Expect(execute("get b")).To(Equal(
				//               ERROR(`cannot get "b": no such variable`)
				//             );
				//           });
				//           Specify("two", func() {
				//             evaluate("argspec (a ?b) set (val1 val2)");
				//             Expect(evaluate("get a")).To(Equal(STR("val1"));
				//             Expect(evaluate("get b")).To(Equal(STR("val2"));
				//           });
				//         });
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
				//         Specify("usage", func() {
				//           Expect(evaluate("argspec ((?a def)) usage")).To(Equal(STR("?a?"));
				//         });
				//         Describe("set", func() {
				//           Describe("static", func() {
				//             Specify("zero", func() {
				//               evaluate("argspec ((?a def)) set ()");
				//               Expect(evaluate("get a")).To(Equal(STR("def"));
				//             });
				//             Specify("one", func() {
				//               evaluate("argspec ((?a def)) set (val)");
				//               Expect(evaluate("get a")).To(Equal(STR("val"));
				//             });
				//           });
				//           Describe("dynamic", func() {
				//             Specify("zero", func() {
				//               evaluate("argspec ((?a {+ 1 2})) set ()");
				//               Expect(evaluate("get a")).To(Equal(INT(3));
				//             });
				//             Specify("one", func() {
				//               evaluate("argspec ((?a def)) set (val)");
				//               Expect(evaluate("get a")).To(Equal(STR("val"));
				//             });
				//           });
				//         });
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
				//         Specify("usage", func() {
				//           Expect(evaluate("argspec ((guard ?a def)) usage")).To(Equal(STR("?a?"));
				//         });
				//         Describe("set", func() {
				//           Describe("simple command", func() {
				//             Specify("required", func() {
				//               evaluate("set args [argspec ( (list a) )]");
				//               Expect(execute("argspec $args set ((1 2 3))")).To(Equal(OK(NIL));
				//               Expect(evaluate("get a")).To(Equal(evaluate("list (1 2 3)"));
				//               Expect(execute("argspec $args set (value)")).To(Equal(
				//                 ERROR("invalid list")
				//               );
				//             });
				//             Specify("optional", func() {
				//               evaluate("set args [argspec ( (list ?a) )]");
				//               Expect(execute("argspec $args set ()")).To(Equal(OK(NIL));
				//               Expect(execute("get a")).To(Equal(
				//                 ERROR(`cannot get "a": no such variable`)
				//               );
				//               Expect(execute("argspec $args set ((1 2 3))")).To(Equal(OK(NIL));
				//               Expect(evaluate("get a")).To(Equal(evaluate("list (1 2 3)"));
				//               Expect(execute("argspec $args set (value)")).To(Equal(
				//                 ERROR("invalid list")
				//               );
				//             });
				//             Specify("default", func() {
				//               evaluate("set args [argspec ( (list ?a ()) )]");
				//               Expect(execute("argspec $args set ()")).To(Equal(OK(NIL));
				//               Expect(evaluate("get a")).To(Equal(evaluate("list ()"));
				//               Expect(execute("argspec $args set ((1 2 3))")).To(Equal(OK(NIL));
				//               Expect(evaluate("get a")).To(Equal(evaluate("list (1 2 3)"));
				//               Expect(execute("argspec $args set (value)")).To(Equal(
				//                 ERROR("invalid list")
				//               );
				//             });
				//           });
				//           Describe("tuple prefix", func() {
				//             Specify("required", func() {
				//               evaluate("set args [argspec ( ((0 <) a) )]");
				//               Expect(execute("argspec $args set (1)")).To(Equal(OK(NIL));
				//               Expect(evaluate("get a")).To(Equal(TRUE);
				//               Expect(execute("argspec $args set (-1)")).To(Equal(OK(NIL));
				//               Expect(evaluate("get a")).To(Equal(FALSE);
				//               Expect(execute("argspec $args set (value)")).To(Equal(
				//                 ERROR('invalid number "value"')
				//               );
				//             });
				//             Specify("optional", func() {
				//               evaluate("set args [argspec ( ((0 <) ?a) )]");
				//               Expect(execute("argspec $args set ()")).To(Equal(OK(NIL));
				//               Expect(execute("get a")).To(Equal(
				//                 ERROR(`cannot get "a": no such variable`)
				//               );
				//               Expect(execute("argspec $args set (1)")).To(Equal(OK(NIL));
				//               Expect(evaluate("get a")).To(Equal(TRUE);
				//               Expect(execute("argspec $args set (-1)")).To(Equal(OK(NIL));
				//               Expect(evaluate("get a")).To(Equal(FALSE);
				//               Expect(execute("argspec $args set (value)")).To(Equal(
				//                 ERROR('invalid number "value"')
				//               );
				//             });
				//             Specify("default", func() {
				//               evaluate("set args [argspec ( ((0 <) ?a 1) )]");
				//               Expect(execute("argspec $args set ()")).To(Equal(OK(NIL));
				//               Expect(evaluate("get a")).To(Equal(TRUE);
				//               Expect(execute("argspec $args set (1)")).To(Equal(OK(NIL));
				//               Expect(evaluate("get a")).To(Equal(TRUE);
				//               Expect(execute("argspec $args set (-1)")).To(Equal(OK(NIL));
				//               Expect(evaluate("get a")).To(Equal(FALSE);
				//               Expect(execute("argspec $args set (value)")).To(Equal(
				//                 ERROR('invalid number "value"')
				//               );
				//             });
				//           });
				//         });
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

		//     Describe("Subcommands", func() {
		//       mochadoc.description(func() {
		//         /**
		//          * The `argspec` ensemble comes with a number of predefined subcommands
		//          * listed here.
		//          */
		//       });

		//       Describe("`subcommands`", func() {
		//         mochadoc.description(usage("argspec () subcommands"));
		//         mochadoc.description(func() {
		//           /**
		//            * This subcommand is useful for introspection and interactive
		//            * calls.
		//            */
		//         });

		//         Specify("usage", func() {
		//           Expect(evaluate("help argspec () subcommands")).To(Equal(
		//             STR("argspec value subcommands")
		//           );
		//         });

		//         It("should return list of subcommands", func() {
		//           Expect(evaluate("argspec {} subcommands")).To(Equal(
		//             evaluate("list (subcommands usage set)")
		//           );
		//         });

		//         Describe("Exceptions", func() {
		//           Specify("wrong arity", func() {
		//             /**
		//              * The subcommand will return an error message with usage when
		//              * given the wrong number of arguments.
		//              */
		//             Expect(execute("argspec {} subcommands a")).To(Equal(
		//               ERROR('wrong # args: should be "argspec value subcommands"')
		//             );
		//             Expect(execute("help argspec {} subcommands a")).To(Equal(
		//               ERROR('wrong # args: should be "argspec value subcommands"')
		//             );
		//           });
		//         });
		//       });

		//       Describe("`usage`", func() {
		//         mochadoc.description(usage("argspec () usage"));
		//         mochadoc.description(func() {
		//           /**
		//            * Get a help string
		//            */
		//         });

		//         It("should return a usage string with argument names", func() {
		//           /**
		//            * This subcommand returns a help string for the argspec command.
		//            *
		//            */
		//           Expect(evaluate("argspec {a b ?c *} usage")).To(Equal(
		//             STR("a b ?c? ?arg ...?")
		//           );
		//         });

		//         Describe("Exceptions", func() {
		//           Specify("wrong arity", func() {
		//             /**
		//              * The subcommand will return an error message with usage when given
		//              * the wrong number of arguments.
		//              */
		//             Expect(execute("argspec {} usage a")).To(Equal(
		//               ERROR('wrong # args: should be "argspec value usage"')
		//             );
		//           });
		//         });
		//       });

		//       Describe("`set`", func() {
		//         mochadoc.description(usage("argspec () usage"));
		//         mochadoc.description(func() {
		//           /**
		//            * Set parameter variables from a list of argument values
		//            */
		//         });

		//         It("should return nil", func() {
		//           Expect(evaluate("argspec {} set ()")).To(Equal(NIL);
		//         });
		//         It("should set argument variables in the caller scope", func() {
		//           evaluate("argspec {a} set (val)");
		//           Expect(evaluate("get a")).To(Equal(STR("val"));
		//         });
		//         It("should enforce minimum number of arguments", func() {
		//           Expect(execute("argspec {a} set ()")).To(Equal(
		//             ERROR(`wrong # values: should be "a"`)
		//           );
		//           Expect(execute("argspec {a ?b} set ()")).To(Equal(
		//             ERROR(`wrong # values: should be "a ?b?"`)
		//           );
		//           Expect(execute("argspec {?a b c} set (val)")).To(Equal(
		//             ERROR(`wrong # values: should be "?a? b c"`)
		//           );
		//           Expect(execute("argspec {a *b c} set (val)")).To(Equal(
		//             ERROR(`wrong # values: should be "a ?b ...? c"`)
		//           );
		//         });
		//         It("should enforce maximum number of arguments", func() {
		//           Expect(execute("argspec {} set (val1)")).To(Equal(
		//             ERROR(`wrong # values: should be ""`)
		//           );
		//           Expect(execute("argspec {a} set (val1 val2)")).To(Equal(
		//             ERROR(`wrong # values: should be "a"`)
		//           );
		//           Expect(execute("argspec {a ?b} set (val1 val2 val3)")).To(Equal(
		//             ERROR(`wrong # values: should be "a ?b?"`)
		//           );
		//         });
		//         It("should set required attributes first", func() {
		//           evaluate("argspec {?a b ?c} set (val)");
		//           Expect(evaluate("get b")).To(Equal(STR("val"));
		//         });
		//         It("should skip missing optional attributes", func() {
		//           evaluate("argspec {?a b (?c def)} set (val)");
		//           Expect(execute("get a")).To(Equal(
		//             ERROR(`cannot get "a": no such variable`)
		//           );
		//           Expect(evaluate("get b")).To(Equal(STR("val"));
		//           Expect(evaluate("get c")).To(Equal(STR("def"));
		//         });
		//         It("should set optional attributes in order", func() {
		//           evaluate("argspec {(?a def) b ?c} set (val1 val2)");
		//           Expect(evaluate("get a")).To(Equal(STR("val1"));
		//           Expect(evaluate("get b")).To(Equal(STR("val2"));
		//           Expect(execute("get c")).To(Equal(
		//             ERROR(`cannot get "c": no such variable`)
		//           );
		//         });
		//         It("should set remainder after optional attributes", func() {
		//           evaluate("argspec {?a *b c} set (val1 val2)");
		//           Expect(evaluate("get a")).To(Equal(STR("val1"));
		//           Expect(evaluate("get b")).To(Equal(TUPLE([]));
		//           Expect(evaluate("get c")).To(Equal(STR("val2"));
		//         });
		//         It("should set all present attributes in order", func() {
		//           evaluate("argspec {?a *b c} set (val1 val2 val3 val4)");
		//           Expect(evaluate("get a")).To(Equal(STR("val1"));
		//           Expect(evaluate("get b")).To(Equal(TUPLE([STR("val2"), STR("val3")]));
		//           Expect(evaluate("get c")).To(Equal(STR("val4"));
		//         });

		//         Describe("Exceptions", func() {
		//           Specify("wrong arity", func() {
		//             /**
		//              * The subcommand will return an error message with usage when given
		//              * the wrong number of arguments.
		//              */
		//             Expect(execute("argspec {} set")).To(Equal(
		//               ERROR('wrong # args: should be "argspec value set values"')
		//             );
		//             Expect(execute("argspec {} set a b")).To(Equal(
		//               ERROR('wrong # args: should be "argspec value set values"')
		//             );
		//           });
		//         });
		//       });

		//       Describe("Exceptions", func() {
		//         Specify("unknown subcommand", func() {
		//           Expect(execute("argspec () unknownSubcommand")).To(Equal(
		//             ERROR('unknown subcommand "unknownSubcommand"')
		//           );
		//         });
		//         Specify("invalid subcommand name", func() {
		//           Expect(execute("argspec () []")).To(Equal(
		//             ERROR("invalid subcommand name")
		//           );
		//         });
		//       });
		//     });

		//     Describe("Examples", func() {
		//       It("Currying and encapsulation", func(){example(exampleSpec[
		//         {
		//           doc: func() {
		//             /**
		//              * Thanks to leading tuple auto-expansion, it is very simple to
		//              * bundle the `argspec` command and a value into a tuple to create a
		//              * pseudo-object value. For example:
		//              */
		//           },
		//           script: "set l (argspec {a b ?c *})",
		//         },
		//         {
		//           doc: func() {
		//             /**
		//              * We can then use this variable like a regular command. Calling it
		//              * without argument will return the wrapped value:
		//              */
		//           },
		//           script: "$l",
		//           result: evaluate("argspec {a b ?c *}"),
		//         },
		//         {
		//           doc: func() {
		//             /**
		//              * Subcommands then behave like object methods:
		//              */
		//           },
		//           script: "$l usage",
		//           result: STR("a b ?c? ?arg ...?"),
		//         },
		//         {
		//           script: "$l set (val1 val2 val3); get (a b c)",
		//           result: TUPLE([STR("val1"), STR("val2"), STR("val3")]),
		//         },
		//       ]);
		//       It("Argument type guard", func(){example(exampleSpec[
		//         {
		//           doc: func() {
		//             /**
		//              * Calling `argspec` with a single argument returns its value as a
		//              * argspec. This property allows `argspec` to be used as a type
		//              * guard for argspecs.
		//              *
		//              * Here we create a macro `usage` that returns the usage of the
		//              * provided argspec. Using `argspec` as guard has three effects:
		//              *
		//              * - it validates the argument on the caller side
		//              * - it converts the value at most once
		//              * - it ensures type safety within the body
		//              *
		//              * Note how using `argspec` as a guard for argument `a` makes it
		//              * look like a static type declaration:
		//              */
		//           },
		//           script: "macro usage ( (argspec a) ) {argspec $a usage}",
		//         },
		//         {
		//           doc: func() {
		//             /**
		//              * Passing a valid value will give the Expected result:
		//              */
		//           },
		//           script: "usage {a b ?c *}",
		//           result: STR("a b ?c? ?arg ...?"),
		//         },
		//         {
		//           doc: func() {
		//             /**
		//              * Passing an invalid value will produce an error:
		//              */
		//           },
		//           script: "usage invalidValue",
		//           result: ERROR("invalid argument list"),
		//         },
		//       ]);
		//     });

		//     Describe("Ensemble command", func() {
		//       mochadoc.description(func() {
		//         /**
		//          * `argspec` is an ensemble command, which means that it is a collection
		//          * of subcommands defined in an ensemble scope.
		//          */
		//       });

		//       It("should return its ensemble metacommand when called with no argument", func() {
		//         /**
		//          * The typical application of this property is to access the ensemble
		//          * metacommand by wrapping the command within brackets, i.e.
		//          * `[argspec]`.
		//          */
		//         Expect(evaluate("argspec").type).To(Equal(ValueType.COMMAND);
		//       });
		//       It("should be extensible", func() {
		//         /**
		//          * Creating a command in the `argspec` ensemble scope will add it to its
		//          * subcommands.
		//          */
		//         evaluate(`
		//           [argspec] eval {
		//             macro foo {value} {idem bar}
		//           }
		//         `);
		//         Expect(evaluate("argspec (a b c) foo")).To(Equal(STR("bar"));
		//       });
		//       It("should support help for custom subcommands", func() {
		//         /**
		//          * Like all ensemble commands, `argspec` have built-in support for
		//          * `help` on all subcommands that support it.
		//          */
		//         evaluate(`
		//           [argspec] eval {
		//             macro foo {value a b} {idem bar}
		//           }
		//         `);
		//         Expect(evaluate("help argspec (a b c) foo")).To(Equal(
		//           STR("argspec value foo a b")
		//         );
		//         Expect(execute("help argspec (a b c) foo 1 2 3")).To(Equal(
		//           ERROR('wrong # args: should be "argspec value foo a b"')
		//         );
		//       });

		//       Describe("Examples", func() {
		//         It("Adding a `help` subcommand", func(){example(exampleSpec[
		//           {
		//             doc: func() {
		//               /**
		//                * Here we create a `help` alias to the existing `usage` within
		//                * the `argspec` ensemble, returning the `usage` with a prefix
		//                * string:
		//                */
		//             },
		//             script: `
		//               [argspec] eval {
		//                 macro help {value prefix} {
		//                   idem "$prefix [argspec $value usage]"
		//                 }
		//               }
		//             `,
		//           },
		//           {
		//             doc: func() {
		//               /**
		//                * We can then use `help` just like the predefined `argspec`
		//                * subcommands:
		//                */
		//             },
		//             script: "argspec {a b ?c *} help foo",
		//             result: STR("foo a b ?c? ?arg ...?"),
		//           },
		//         ]);
		//       });
		//     });
	})
})
