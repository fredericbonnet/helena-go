package helena_dialect_test

import (
	"math"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"helena/core"
	. "helena/helena_dialect"
)

// import { expect } from "chai";
// import * as mochadoc from "../../mochadoc";
// import { ERROR } from "../core/results";
// import { Parser } from "../core/parser";
// import { Tokenizer } from "../core/tokenizer";
// import {
//   FALSE,
//   INT,
//   REAL,
//   STR,
//   StringValue,
//   TRUE,
//   ValueType,
// } from "../core/values";
// import { Scope } from "./core";
// import { initCommands } from "./helena-dialect";
// import { codeBlock, describeCommand, specifyExample } from "./test-helpers";

var _ = Describe("Helena numbers", func() {
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

	Describe("Number commands", func() {
		Describe("Integer numbers", func() {
			It("are valid commands", func() {
				Expect(evaluate("1")).To(Equal(INT(1)))
			})
			It("are idempotent", func() {
				Expect(evaluate("[1]")).To(Equal(INT(1)))
			})
			It("can be expressed as strings", func() {
				Expect(evaluate(`"123"`)).To(Equal(INT(123)))
			})

			Describe("Exceptions", func() {
				Specify("unknown subcommand", func() {
					Expect(execute("1 unknownSubcommand")).To(Equal(
						ERROR(`unknown subcommand "unknownSubcommand"`),
					))
				})
				Specify("invalid subcommand name", func() {
					Expect(execute("1 []")).To(Equal(ERROR("invalid subcommand name")))
				})
			})
		})

		Describe("Real numbers", func() {
			It("are valid commands", func() {
				Expect(evaluate("1.25")).To(Equal(REAL(1.25)))
			})
			It("are idempotent", func() {
				Expect(evaluate("[1.25]")).To(Equal(REAL(1.25)))
			})
			It("can be expressed as strings", func() {
				Expect(evaluate(`"0.5"`)).To(Equal(REAL(0.5)))
			})

			Describe("Exceptions", func() {
				Specify("unknown subcommand", func() {
					Expect(execute("1.23 unknownSubcommand")).To(Equal(
						ERROR(`unknown subcommand "unknownSubcommand"`),
					))
				})
				Specify("invalid subcommand name", func() {
					Expect(execute("1.23 []")).To(Equal(ERROR("invalid subcommand name")))
				})
			})
		})

		Describe("Infix operators", func() {
			Describe("Arithmetic", func() {
				Specify("`+`", func() {
					Expect(evaluate("1 + 2")).To(Equal(INT(3)))
					Expect(evaluate("1 + 2 + 3 + 4")).To(Equal(INT(10)))
				})

				Specify("`-`", func() {
					Expect(evaluate("1 - 2")).To(Equal(INT(-1)))
					Expect(evaluate("1 - 2 - 3 - 4")).To(Equal(INT(-8)))
				})

				Specify("`*`", func() {
					Expect(evaluate("1 * 2")).To(Equal(INT(2)))
					Expect(evaluate("1 * 2 * 3 * 4")).To(Equal(INT(24)))
				})

				Specify("`/`", func() {
					Expect(evaluate("1 / 2")).To(Equal(REAL(0.5)))
					Expect(evaluate("1 / 2 / 4 / 8")).To(Equal(REAL(0.015625)))
					Expect(evaluate("1 / 0")).To(Equal(REAL(math.Inf(1))))
					Expect(evaluate("-1 / 0")).To(Equal(REAL(math.Inf(-1))))
					Expect(math.IsNaN(evaluate("0 / 0").(core.RealValue).Value))
				})

				Specify("Precedence rules", func() {
					Expect(evaluate("1 + 2 * 3 * 4 + 5")).To(Equal(INT(30)))
					Expect(evaluate("1 * 2 + 3 * 4 + 5 + 6 * 7")).To(Equal(INT(61)))
					Expect(evaluate("1 - 2 * 3 * 4 + 5")).To(Equal(INT(-18)))
					Expect(evaluate("1 - 2 * 3 / 4 + 5 * 6 / 10")).To(Equal(REAL(2.5)))
					Expect(evaluate("10 / 2 / 5")).To(Equal(INT(1)))
				})

				Specify("Conversions", func() {
					Expect(evaluate("1 + 2.3")).To(Equal(REAL(3.3)))
					Expect(evaluate("1.5 + 2.5")).To(Equal(INT(4)))
				})

				Describe("Exceptions", func() {
					Specify("wrong arity", func() {
						Expect(execute("1 +")).To(Equal(
							ERROR(
								`wrong # operands: should be "operand ?operator operand? ?...?"`,
							),
						))
					})
					Specify("invalid value", func() {
						Expect(execute("1 + a")).To(Equal(ERROR(`invalid number "a"`)))
					})
					Specify("unknown operator", func() {
						Expect(execute("1 + 2 a 3")).To(Equal(ERROR(`invalid operator "a"`)))
					})
					Specify("invalid operator", func() {
						Expect(execute("1 + 2 [] 3")).To(Equal(ERROR("invalid operator")))
					})
				})
			})

			Describe("Comparisons", func() {
				Describe("`==`", func() {
					It("should compare two numbers", func() {
						Expect(evaluate(`"123" == -34`)).To(Equal(FALSE))
						Expect(evaluate(`56 == "56.0"`)).To(Equal(TRUE))
						Expect(evaluate("set var 1; $var == $var")).To(Equal(TRUE))
					})

					Describe("Exceptions", func() {
						Specify("wrong arity", func() {
							Expect(execute("1 ==")).To(Equal(
								ERROR(`wrong # operands: should be "operand1 == operand2"`),
							))
							Expect(execute("1 == 2 3")).To(Equal(
								ERROR(`wrong # operands: should be "operand1 == operand2"`),
							))
						})
						Specify("invalid value", func() {
							Expect(execute("1 == a")).To(Equal(ERROR(`invalid number "a"`)))
						})
					})
				})

				Describe("`!=`", func() {
					It("should compare two numbers", func() {
						Expect(evaluate(`"123" != -34`)).To(Equal(TRUE))
						Expect(evaluate(`56 != "56.0"`)).To(Equal(FALSE))
						Expect(evaluate("set var 1; $var != $var")).To(Equal(FALSE))
					})

					Describe("Exceptions", func() {
						Specify("wrong arity", func() {
							Expect(execute("1 !=")).To(Equal(
								ERROR(`wrong # operands: should be "operand1 != operand2"`),
							))
							Expect(execute("1 != 2 3")).To(Equal(
								ERROR(`wrong # operands: should be "operand1 != operand2"`),
							))
						})
						Specify("invalid value", func() {
							Expect(execute("1 != a")).To(Equal(ERROR(`invalid number "a"`)))
						})
					})
				})

				Describe("`>`", func() {
					It("should compare two numbers", func() {
						Expect(evaluate("12 > -34")).To(Equal(TRUE))
						Expect(evaluate(`56 > "56.0"`)).To(Equal(FALSE))
						Expect(evaluate("set var 1; $var > $var")).To(Equal(FALSE))
					})

					Describe("Exceptions", func() {
						Specify("wrong arity", func() {
							Expect(execute("1 >")).To(Equal(
								ERROR(`wrong # operands: should be "operand1 > operand2"`),
							))
							Expect(execute("1 > 2 3")).To(Equal(
								ERROR(`wrong # operands: should be "operand1 > operand2"`),
							))
						})
						Specify("invalid value", func() {
							Expect(execute("1 > a")).To(Equal(ERROR(`invalid number "a"`)))
						})
					})
				})

				Describe("`>=`", func() {
					It("should compare two numbers", func() {
						Expect(evaluate("12 >= -34")).To(Equal(TRUE))
						Expect(evaluate(`56 >= "56.0"`)).To(Equal(TRUE))
						Expect(evaluate("set var 1; $var >= $var")).To(Equal(TRUE))
					})

					Describe("Exceptions", func() {
						Specify("wrong arity", func() {
							Expect(execute("1 >=")).To(Equal(
								ERROR(`wrong # operands: should be "operand1 >= operand2"`),
							))
							Expect(execute("1 >= 2 3")).To(Equal(
								ERROR(`wrong # operands: should be "operand1 >= operand2"`),
							))
						})
						Specify("invalid value", func() {
							Expect(execute("1 >= a")).To(Equal(ERROR(`invalid number "a"`)))
						})
					})
				})

				Describe("`<`", func() {
					It("should compare two numbers", func() {
						Expect(evaluate("12 < -34")).To(Equal(FALSE))
						Expect(evaluate(`56 < "56.0"`)).To(Equal(FALSE))
						Expect(evaluate("set var 1; $var < $var")).To(Equal(FALSE))
					})

					Describe("Exceptions", func() {
						Specify("wrong arity", func() {
							Expect(execute("1 <")).To(Equal(
								ERROR(`wrong # operands: should be "operand1 < operand2"`),
							))
							Expect(execute("1 < 2 3")).To(Equal(
								ERROR(`wrong # operands: should be "operand1 < operand2"`),
							))
						})
						Specify("invalid value", func() {
							Expect(execute("1 < a")).To(Equal(ERROR(`invalid number "a"`)))
						})
					})
				})

				Describe("`<=`", func() {
					It("should compare two numbers", func() {
						Expect(evaluate("12 <= -34")).To(Equal(FALSE))
						Expect(evaluate(`56 <= "56.0"`)).To(Equal(TRUE))
						Expect(evaluate("set var 1; $var <= $var")).To(Equal(TRUE))
					})

					Describe("Exceptions", func() {
						Specify("wrong arity", func() {
							Expect(execute("1 <=")).To(Equal(
								ERROR(`wrong # operands: should be "operand1 <= operand2"`),
							))
							Expect(execute("1 <= 2 3")).To(Equal(
								ERROR(`wrong # operands: should be "operand1 <= operand2"`),
							))
						})
						Specify("invalid value", func() {
							Expect(execute("1 <= a")).To(Equal(ERROR(`invalid number "a"`)))
						})
					})
				})
			})
		})

		Describe("Subcommands", func() {
			Describe("Introspection", func() {
				Describe("`subcommands`", func() {
					It("should return list of subcommands", func() {
						Expect(evaluate("1 subcommands")).To(Equal(
							evaluate("list (subcommands + - * / == != > >= < <=)"),
						))
						Expect(evaluate("1.2 subcommands")).To(Equal(
							evaluate("list (subcommands + - * / == != > >= < <=)"),
						))
					})

					Describe("Exceptions", func() {
						Specify("wrong arity", func() {
							Expect(execute("1 subcommands a")).To(Equal(
								ERROR(`wrong # args: should be "<number> subcommands"`),
							))
							Expect(execute("1.2 subcommands a")).To(Equal(
								ERROR(`wrong # args: should be "<number> subcommands"`),
							))
						})
					})
				})
			})
		})
	})

	//   describeCommand("int", func() {
	//     mochadoc.summary("Integer number handling");
	//     mochadoc.usage(usage("int"));
	//     mochadoc.description(func() {
	//       /**
	//        * The `int` command is a type command dedicated to integer values.
	//        *
	//        * Integer values are Helena values whose internal type is `INTEGER`. The
	//        * name `int` was preferred over `integer` because it is shorter and is
	//        * already used in many other languages like Python and C.
	//        */
	//     });

	//     Describe("Integer conversion", func() {
	//       mochadoc.description(func() {
	//         /**
	//          * Like with most type commands, passing a single argument to `int` will
	//          * ensure an integer value in return. This property means that `int` can
	//          * be used for creation and conversion, but also as a type guard in
	//          * argspecs.
	//          */
	//       });

	//       It("should return integer value", func() {
	//         Expect(evaluate("int 0")).To(Equal(INT(0));
	//       });

	//       Describe("Exceptions", func() {
	//         Specify("values with no string representation", func() {
	//           Expect(execute("int []")).To(Equal(
	//             ERROR("value has no string representation")
	//           );
	//           Expect(execute("int ()")).To(Equal(
	//             ERROR("value has no string representation")
	//           );
	//         });
	//         Specify("invalid values", func() {
	//           Expect(execute("int a")).To(Equal(ERROR('invalid integer "a"'));
	//         });
	//         Specify("real values", func() {
	//           /**
	//            * Non-integer real values are not accepted.
	//            */
	//           Expect(execute("int 1.1")).To(Equal(ERROR('invalid integer "1.1"'));
	//         });
	//       });
	//     });

	//     Describe("Subcommands", func() {
	//       mochadoc.description(func() {
	//         /**
	//          * The `int` ensemble comes with a number of predefined subcommands
	//          * listed here.
	//          */
	//       });

	//       Describe("Introspection", func() {
	//         Describe("`subcommands`", func() {
	//           mochadoc.description(usage("int 0 subcommands"));
	//           mochadoc.description(func() {
	//             /**
	//              * This subcommand is useful for introspection and interactive
	//              * calls.
	//              */
	//           });

	//           Specify("usage", func() {
	//             Expect(evaluate("help int 0 subcommands")).To(Equal(
	//               STR("int value subcommands")
	//             );
	//           });

	//           It("should return list of subcommands", func() {
	//             Expect(evaluate("int 0 subcommands")).To(Equal(
	//               evaluate("list (subcommands)")
	//             );
	//           });

	//           Describe("Exceptions", func() {
	//             Specify("wrong arity", func() {
	//               /**
	//                * The subcommand will return an error message with usage when
	//                * given the wrong number of arguments.
	//                */
	//               Expect(execute("int 0 subcommands a")).To(Equal(
	//                 ERROR(`wrong # args: should be "int value subcommands"')
	//               );
	//               Expect(execute("help int 0 subcommands a")).To(Equal(
	//                 ERROR(`wrong # args: should be "int value subcommands"')
	//               );
	//             });
	//           });
	//         });
	//       });

	//       Describe("Exceptions", func() {
	//         Specify("unknown subcommand", func() {
	//           Expect(execute("int 0 unknownSubcommand")).To(Equal(
	//             ERROR('unknown subcommand "unknownSubcommand"')
	//           );
	//         });
	//         Specify("invalid subcommand name", func() {
	//           Expect(execute("int 0 []")).To(Equal(ERROR("invalid subcommand name"));
	//         });
	//       });
	//     });

	//     Describe("Ensemble command", func() {
	//       mochadoc.description(func() {
	//         /**
	//          * `int` is an ensemble command, which means that it is a collection
	//          * of subcommands defined in an ensemble scope.
	//          */
	//       });

	//       It("should return its ensemble metacommand when called with no argument", func() {
	//         /**
	//          * The typical application of this property is to access the ensemble
	//          * metacommand by wrapping the command within brackets, i.e. `[int]`.
	//          */
	//         Expect(evaluate("int").type).To(Equal(ValueType.COMMAND);
	//       });
	//       It("should be extensible", func() {
	//         /**
	//          * Creating a command in the `int` ensemble scope will add it to its
	//          * subcommands.
	//          */
	//         evaluate(`
	//           [int] eval {
	//             macro foo {value} {idem bar}
	//           }
	//         `);
	//         Expect(evaluate("int example foo")).To(Equal(STR("bar"));
	//       });
	//       It("should support help for custom subcommands", func() {
	//         /**
	//          * Like all ensemble commands, `int` have built-in support for `help`
	//          * on all subcommands that support it.
	//          */
	//         evaluate(`
	//           [int] eval {
	//             macro foo {value a b} {idem bar}
	//           }
	//         `);
	//         Expect(evaluate("help int 0 foo")).To(Equal(STR("int value foo a b"));
	//         Expect(execute("help int 0 foo 1 2 3")).To(Equal(
	//           ERROR(`wrong # args: should be "int value foo a b"')
	//         );
	//       });

	//       Describe("Examples", func() {
	//         example("Adding a `positive` subcommand", [
	//           {
	//             doc: func() {
	//               /**
	//                * Here we create a `positive` macro within the `int` ensemble
	//                * scope, returning whether the value is strictly positive:
	//                */
	//             },
	//             script: `
	//               [int] eval {
	//                 macro positive {value} {
	//                   $value > 0
	//                 }
	//               }
	//             `,
	//           },
	//           {
	//             doc: func() {
	//               /**
	//                * We can then use `positive` just like the predefined `int`
	//                * subcommands:
	//                */
	//             },
	//             script: "int 1 positive",
	//             result: TRUE,
	//           },
	//           {
	//             script: "int 0 positive",
	//             result: FALSE,
	//           },
	//           {
	//             script: "int -1 positive",
	//             result: FALSE,
	//           },
	//         ]);
	//       });
	//     });
	//   });

	//   describeCommand("real", func() {
	//     mochadoc.summary("Real number handling");
	//     mochadoc.usage(usage("real"));
	//     mochadoc.description(func() {
	//       /**
	//        * The `real` command is a type command dedicated to real values.
	//        *
	//        * Real values are Helena values whose internal type is `REAL`.
	//        */
	//     });

	//     Describe("Real conversion", func() {
	//       mochadoc.description(func() {
	//         /**
	//          * Like with most type commands, passing a single argument to `real`
	//          * will ensure a real value in return. This property means that `real`
	//          * can be used for creation and conversion, but also as a type guard in
	//          * argspecs.
	//          */
	//       });

	//       It("should return real value", func() {
	//         Expect(evaluate("real 0")).To(Equal(REAL(0));
	//       });

	//       Describe("Exceptions", func() {
	//         Specify("values with no string representation", func() {
	//           Expect(execute("real []")).To(Equal(
	//             ERROR("value has no string representation")
	//           );
	//           Expect(execute("real ()")).To(Equal(
	//             ERROR("value has no string representation")
	//           );
	//         });
	//         Specify("invalid values", func() {
	//           Expect(execute("real a")).To(Equal(ERROR('invalid number "a"'));
	//         });
	//       });
	//     });

	//     Describe("Subcommands", func() {
	//       mochadoc.description(func() {
	//         /**
	//          * The `real` ensemble comes with a number of predefined subcommands
	//          * listed here.
	//          */
	//       });

	//       Describe("Introspection", func() {
	//         Describe("`subcommands`", func() {
	//           mochadoc.description(usage("real 0 subcommands"));
	//           mochadoc.description(func() {
	//             /**
	//              * This subcommand is useful for introspection and interactive
	//              * calls.
	//              */
	//           });

	//           Specify("usage", func() {
	//             Expect(evaluate("help real 0 subcommands")).To(Equal(
	//               STR("real value subcommands")
	//             );
	//           });

	//           It("should return list of subcommands", func() {
	//             Expect(evaluate("real 0 subcommands")).To(Equal(
	//               evaluate("list (subcommands)")
	//             );
	//           });

	//           Describe("Exceptions", func() {
	//             Specify("wrong arity", func() {
	//               /**
	//                * The subcommand will return an error message with usage when
	//                * given the wrong number of arguments.
	//                */
	//               Expect(execute("real 0 subcommands a")).To(Equal(
	//                 ERROR(`wrong # args: should be "real value subcommands"')
	//               );
	//               Expect(execute("help real 0 subcommands a")).To(Equal(
	//                 ERROR(`wrong # args: should be "real value subcommands"')
	//               );
	//             });
	//           });
	//         });
	//       });

	//       Describe("Exceptions", func() {
	//         Specify("unknown subcommand", func() {
	//           Expect(execute("real 0 unknownSubcommand")).To(Equal(
	//             ERROR('unknown subcommand "unknownSubcommand"')
	//           );
	//         });
	//         Specify("invalid subcommand name", func() {
	//           Expect(execute("real 0 []")).To(Equal(ERROR("invalid subcommand name"));
	//         });
	//       });
	//     });

	//     Describe("Ensemble command", func() {
	//       mochadoc.description(func() {
	//         /**
	//          * `real` is an ensemble command, which means that it is a collection of
	//          * subcommands defined in an ensemble scope.
	//          */
	//       });

	//       It("should return its ensemble metacommand when called with no argument", func() {
	//         /**
	//          * The typical application of this property is to access the ensemble
	//          * metacommand by wrapping the command within brackets, i.e.
	//          * `[real]`.
	//          */
	//         Expect(evaluate("real").type).To(Equal(ValueType.COMMAND);
	//       });
	//       It("should be extensible", func() {
	//         /**
	//          * Creating a command in the `real` ensemble scope will add it to its
	//          * subcommands.
	//          */
	//         evaluate(`
	//           [real] eval {
	//             macro foo {value} {idem bar}
	//           }
	//         `);
	//         Expect(evaluate("real 0 foo")).To(Equal(STR("bar"));
	//       });
	//       It("should support help for custom subcommands", func() {
	//         /**
	//          * Like all ensemble commands, `real` have built-in support for `help`
	//          * on all subcommands that support it.
	//          */
	//         evaluate(`
	//           [real] eval {
	//             macro foo {value a b} {idem bar}
	//           }
	//         `);
	//         Expect(evaluate("help real 0 foo")).To(Equal(STR("real value foo a b"));
	//         Expect(execute("help real 0 foo 1 2 3")).To(Equal(
	//           ERROR(`wrong # args: should be "real value foo a b"')
	//         );
	//       });

	//       Describe("Examples", func() {
	//         example("Adding a `positive` subcommand", [
	//           {
	//             doc: func() {
	//               /**
	//                * Here we create a `positive` macro within the `real` ensemble
	//                * scope, returning whether the value is strictly positive:
	//                */
	//             },
	//             script: `
	//               [real] eval {
	//                 macro positive {value} {
	//                   $value > 0
	//                 }
	//               }
	//             `,
	//           },
	//           {
	//             doc: func() {
	//               /**
	//                * We can then use `positive` just like the predefined `real`
	//                * subcommands:
	//                */
	//             },
	//             script: "real 0.1 positive",
	//             result: TRUE,
	//           },
	//           {
	//             script: "real 0 positive",
	//             result: FALSE,
	//           },
	//           {
	//             script: "real -1 positive",
	//             result: FALSE,
	//           },
	//         ]);
	//       });
	//     });
	// })
})
