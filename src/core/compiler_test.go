package core_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "helena/core"
)

// const asString = (value) => StringValue.toString(value).data;

type mockVariableResolver struct {
	variables map[string]Value
}

func newMockVariableResolver() *mockVariableResolver {
	return &mockVariableResolver{
		variables: map[string]Value{},
	}
}
func (resolver *mockVariableResolver) Resolve(name string) (value Value, ok bool) {
	v, ok := resolver.variables[name]
	return v, ok
}
func (resolver *mockVariableResolver) register(name string, value Value) {
	resolver.variables[name] = value
}

// class IntCommand implements Command {
//   execute(args: Value[]): Result {
//     return OK(args[0]);
//   }
// }
// const INT_CMD = new IntCommand();
// class MockCommandResolver implements CommandResolver {
//   resolve(name: Value): Command {
//     if (name.type == ValueType.INTEGER || !isNaN(parseInt(asString(name))))
//       return INT_CMD;
//     return this.commands.get(asString(name));
//   }

//   commands: Map<string, Command> = new Map();
//   register(name: string, command: Command) {
//     this.commands.set(name, command);
//   }
// }
// class FunctionCommand implements Command {
//   fn: (args: Value[]) => Value;
//   constructor(fn: (args: Value[]) => Value) {
//     this.fn = fn;
//   }

//   execute(args: Value[]): Result {
//     return OK(this.fn(args));
//   }
// }

// class MockSelectorResolver implements SelectorResolver {
//   resolve(rules: Value[]): Result<Selector> {
//     return this.builder(rules);
//   }
//   builder: (rules) => Result<Selector> = () => OK(NIL, undefined);
//   register(builder: (rules) => Result<Selector>) {
//     this.builder = builder;
//   }
// }

var _ = Describe("Compilation and execution", func() {
	var tokenizer Tokenizer
	var parser *Parser
	var variableResolver *mockVariableResolver
	// var commandResolver MockCommandResolver
	// var selectorResolver MockSelectorResolver
	var compiler Compiler
	var executor *Executor

	parse := func(script string) *Script {
		return parser.Parse(tokenizer.Tokenize(script)).Script
	}
	compileFirstWord := func(script *Script) *Program {
		word := script.Sentences[0].Words[0]
		//     if (word instanceof Word) {
		return compiler.CompileWord(word)
		//     } else {
		// return compiler.CompileConstant(word)
		//     }
	}

	//   const executionModes = [
	//     {
	//       label: "with Executor.execute()",
	//       execute: (program: Program) => executor.execute(program),
	//     },
	//     {
	//       label: "with Executor.functionify()",
	//       execute: (program: Program) => executor.functionify(program)(),
	//     },
	//   ];
	execute := func(program *Program) Result {
		return executor.Execute(program, nil)
	}

	BeforeEach(func() {
		tokenizer = Tokenizer{}
		parser = &Parser{}
		compiler = Compiler{}
		variableResolver = newMockVariableResolver()
		//     commandResolver = new MockCommandResolver();
		//     selectorResolver = new MockSelectorResolver();
		executor = &Executor{
			variableResolver,
			//       commandResolver,
			//       selectorResolver
		}
	})

	//   for (const { label, execute } of executionModes) {
	//     const evaluate = (program: Program) => execute(program).value;
	//     describe(label, func() {
	evaluate := func(program *Program) Value {
		return execute(program).Value
	}
	Describe("Compiler", func() {
		Describe("words", func() {
			Describe("roots", func() {
				Specify("literal", func() {
					script := parse("word")
					program := compileFirstWord(script)
					Expect(program.OpCodes).To(Equal([]OpCode{OpCode_PUSH_CONSTANT}))
					Expect(program.Constants).To(Equal([]Value{STR("word")}))

					Expect(evaluate(program)).To(Equal(STR("word")))
				})

				Describe("tuples", func() {
					Specify("empty tuple", func() {
						script := parse("()")
						program := compileFirstWord(script)
						Expect(program.OpCodes).To(Equal([]OpCode{
							OpCode_OPEN_FRAME,
							OpCode_CLOSE_FRAME,
						}))

						Expect(evaluate(program)).To(Equal(TUPLE([]Value{})))
					})
					Specify("tuple with literals", func() {
						script := parse("( lit1 lit2 )")
						program := compileFirstWord(script)
						Expect(program.OpCodes).To(Equal([]OpCode{
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_PUSH_CONSTANT,
							OpCode_CLOSE_FRAME,
						}))
						Expect(program.Constants).To(Equal([]Value{STR("lit1"), STR("lit2")}))

						Expect(evaluate(program)).To(Equal(
							TUPLE([]Value{STR("lit1"), STR("lit2")}),
						))
					})
					Specify("complex case", func() {
						script := parse(
							`( this [cmd] $var1 "complex" ${var2}(key) )`,
						)
						program := compileFirstWord(script)
						Expect(program.OpCodes).To(Equal([]OpCode{
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_CLOSE_FRAME,
							OpCode_EVALUATE_SENTENCE,
							OpCode_PUSH_RESULT,
							OpCode_PUSH_CONSTANT,
							OpCode_RESOLVE_VALUE,
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_CLOSE_FRAME,
							OpCode_JOIN_STRINGS,
							OpCode_PUSH_CONSTANT,
							OpCode_RESOLVE_VALUE,
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_CLOSE_FRAME,
							OpCode_SELECT_KEYS,
							OpCode_CLOSE_FRAME,
						}))
						Expect(program.Constants).To(Equal([]Value{
							STR("this"),
							STR("cmd"),
							STR("var1"),
							STR("complex"),
							STR("var2"),
							STR("key"),
						}))

						// commandResolver.register(
						//   "cmd",
						//   new FunctionCommand(() => STR("is"))
						// );
						// variableResolver.register("var1", STR("a"));
						// variableResolver.register("var2", DICT(map[string]Value{ key: STR("tuple") }));
						// Expect(evaluate(program)).To(Equal(
						//   TUPLE([
						//     STR("this"),
						//     STR("is"),
						//     STR("a"),
						//     STR("complex"),
						//     STR("tuple"),
						//   ])
						// );
					})
				})

				Describe("blocks", func() {
					Specify("empty block", func() {
						const source = ""
						script := parse(`{` + source + `}`)
						value := NewScriptValue(*parse(source), source)
						program := compileFirstWord(script)
						Expect(program.OpCodes).To(Equal([]OpCode{OpCode_PUSH_CONSTANT}))
						Expect(program.Constants).To(Equal([]Value{value}))

						Expect(evaluate(program)).To(Equal(value))
					})
					Specify("block with literals", func() {
						const source = " lit1 lit2 "
						script := parse(`{` + source + `}`)
						value := NewScriptValue(*parse(source), source)
						program := compileFirstWord(script)
						Expect(program.OpCodes).To(Equal([]OpCode{OpCode_PUSH_CONSTANT}))
						Expect(program.Constants).To(Equal([]Value{value}))

						Expect(evaluate(program)).To(Equal(value))
					})
					Specify("complex case", func() {
						source := ` this [cmd] $var1 "complex" ${var2}(key) `
						script := parse(`{` + source + `}`)
						// const block = (script.sentences[0].words[0] as Word)
						//   .morphemes[0] as BlockMorpheme;
						block := script.Sentences[0].Words[0].Morphemes[0].(BlockMorpheme)
						value := NewScriptValue(block.Subscript, source)
						program := compileFirstWord(script)
						Expect(program.OpCodes).To(Equal([]OpCode{OpCode_PUSH_CONSTANT}))
						Expect(program.Constants).To(Equal([]Value{value}))

						Expect(evaluate(program)).To(Equal(value))
					})
				})

				Describe("expressions", func() {
					Specify("empty expression", func() {
						script := parse("[]")
						program := compileFirstWord(script)
						Expect(program.OpCodes).To(Equal([]OpCode{OpCode_PUSH_NIL}))

						Expect(evaluate(program)).To(Equal(NIL))
					})
					Specify("expression with literals", func() {
						script := parse("[ cmd arg ]")
						program := compileFirstWord(script)
						Expect(program.OpCodes).To(Equal([]OpCode{
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_PUSH_CONSTANT,
							OpCode_CLOSE_FRAME,
							OpCode_EVALUATE_SENTENCE,
							OpCode_PUSH_RESULT,
						}))
						Expect(program.Constants).To(Equal([]Value{STR("cmd"), STR("arg")}))

						// commandResolver.register(
						//   "cmd",
						//   new FunctionCommand((args) => TUPLE([...args, STR("foo")]))
						// );
						// Expect(evaluate(program)).To(Equal(
						//   TUPLE([STR("cmd"), STR("arg"), STR("foo")])
						// );
					})
					Specify("complex case", func() {
						script := parse(
							`[ this [cmd] $var1 "complex" ${var2}(key) ]`,
						)
						program := compileFirstWord(script)
						Expect(program.OpCodes).To(Equal([]OpCode{
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_CLOSE_FRAME,
							OpCode_EVALUATE_SENTENCE,
							OpCode_PUSH_RESULT,
							OpCode_PUSH_CONSTANT,
							OpCode_RESOLVE_VALUE,
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_CLOSE_FRAME,
							OpCode_JOIN_STRINGS,
							OpCode_PUSH_CONSTANT,
							OpCode_RESOLVE_VALUE,
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_CLOSE_FRAME,
							OpCode_SELECT_KEYS,
							OpCode_CLOSE_FRAME,
							OpCode_EVALUATE_SENTENCE,
							OpCode_PUSH_RESULT,
						}))
						Expect(program.Constants).To(Equal([]Value{
							STR("this"),
							STR("cmd"),
							STR("var1"),
							STR("complex"),
							STR("var2"),
							STR("key"),
						}))

						// commandResolver.register(
						//   "cmd",
						//   new FunctionCommand(() => STR("is"))
						// );
						// variableResolver.register("var1", STR("a"));
						// variableResolver.register(
						//   "var2",
						//   DICT(map[string]Value{ key: STR("expression") })
						// );
						// commandResolver.register(
						//   "this",
						//   new FunctionCommand((args) => TUPLE(args))
						// );
						// Expect(evaluate(program)).To(Equal(
						//   TUPLE([
						//     STR("this"),
						//     STR("is"),
						//     STR("a"),
						//     STR("complex"),
						//     STR("expression"),
						//   ])
						// );
					})
					//               Describe("exceptions", func() {
					//                 Specify("unresolved command", func() {
					//                    script := parse("[ cmd arg ]");
					//                    program := compileFirstWord(script);
					//                   Expect(program.OpCodes).To(Equal([]OpCode{
					//                     OpCode_OPEN_FRAME,
					//                     OpCode_PUSH_CONSTANT,
					//                     OpCode_PUSH_CONSTANT,
					//                     OpCode_CLOSE_FRAME,
					//                     OpCode_EVALUATE_SENTENCE,
					//                     OpCode_PUSH_RESULT,
					//                   }));
					//                   Expect(program.Constants).To(Equal([]Value{STR("cmd"), STR("arg")}))

					//                   Expect(execute(program)).To(Equal(
					//                     ERROR('cannot resolve command "cmd"')
					//                   );
					//                 });
					//               });
				})

				Describe("strings", func() {
					Specify("empty string", func() {
						script := parse(`""`)
						program := compileFirstWord(script)
						Expect(program.OpCodes).To(Equal([]OpCode{
							OpCode_OPEN_FRAME,
							OpCode_CLOSE_FRAME,
							OpCode_JOIN_STRINGS,
						}))

						Expect(evaluate(program)).To(Equal(STR("")))
					})
					Specify("simple string", func() {
						script := parse(`"this is a string"`)
						program := compileFirstWord(script)
						Expect(program.OpCodes).To(Equal([]OpCode{
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_CLOSE_FRAME,
							OpCode_JOIN_STRINGS,
						}))
						Expect(program.Constants).To(Equal([]Value{STR("this is a string")}))

						Expect(evaluate(program)).To(Equal(STR("this is a string")))
					})

					Describe("expressions", func() {
						Specify("simple command", func() {
							script := parse(`"this [cmd] a string"`)
							program := compileFirstWord(script)
							Expect(program.OpCodes).To(Equal([]OpCode{
								OpCode_OPEN_FRAME,
								OpCode_PUSH_CONSTANT,
								OpCode_OPEN_FRAME,
								OpCode_PUSH_CONSTANT,
								OpCode_CLOSE_FRAME,
								OpCode_EVALUATE_SENTENCE,
								OpCode_PUSH_RESULT,
								OpCode_PUSH_CONSTANT,
								OpCode_CLOSE_FRAME,
								OpCode_JOIN_STRINGS,
							}))
							Expect(program.Constants).To(Equal([]Value{
								STR("this "),
								STR("cmd"),
								STR(" a string"),
							}))

							// commandResolver.register(
							//   "cmd",
							//   new FunctionCommand(() => STR("is"))
							// );
							// Expect(evaluate(program)).To(Equal(STR("this is a string"));
						})
						Specify("multiple commands", func() {
							script := parse(`"this [cmd1][cmd2] a string"`)
							program := compileFirstWord(script)
							Expect(program.OpCodes).To(Equal([]OpCode{
								OpCode_OPEN_FRAME,
								OpCode_PUSH_CONSTANT,
								OpCode_OPEN_FRAME,
								OpCode_PUSH_CONSTANT,
								OpCode_CLOSE_FRAME,
								OpCode_EVALUATE_SENTENCE,
								OpCode_PUSH_RESULT,
								OpCode_OPEN_FRAME,
								OpCode_PUSH_CONSTANT,
								OpCode_CLOSE_FRAME,
								OpCode_EVALUATE_SENTENCE,
								OpCode_PUSH_RESULT,
								OpCode_PUSH_CONSTANT,
								OpCode_CLOSE_FRAME,
								OpCode_JOIN_STRINGS,
							}))
							Expect(program.Constants).To(Equal([]Value{
								STR("this "),
								STR("cmd1"),
								STR("cmd2"),
								STR(" a string"),
							}))

							// commandResolver.register(
							//   "cmd1",
							//   new FunctionCommand(() => STR("i"))
							// );
							// commandResolver.register(
							//   "cmd2",
							//   new FunctionCommand(() => STR("s"))
							// );
							// Expect(evaluate(program)).To(Equal(STR("this is a string"));
						})
					})

					Describe("substitutions", func() {
						Describe("scalars", func() {
							Specify("simple substitution", func() {
								script := parse(`"this $var a string"`)
								program := compileFirstWord(script)
								Expect(program.OpCodes).To(Equal([]OpCode{
									OpCode_OPEN_FRAME,
									OpCode_PUSH_CONSTANT,
									OpCode_PUSH_CONSTANT,
									OpCode_RESOLVE_VALUE,
									OpCode_PUSH_CONSTANT,
									OpCode_CLOSE_FRAME,
									OpCode_JOIN_STRINGS,
								}))
								Expect(program.Constants).To(Equal([]Value{
									STR("this "),
									STR("var"),
									STR(" a string"),
								}))

								variableResolver.register("var", STR("is"))
								Expect(evaluate(program)).To(Equal(STR("this is a string")))
							})
							Specify("double substitution", func() {
								script := parse(`"this $$var1 a string"`)
								program := compileFirstWord(script)
								Expect(program.OpCodes).To(Equal([]OpCode{
									OpCode_OPEN_FRAME,
									OpCode_PUSH_CONSTANT,
									OpCode_PUSH_CONSTANT,
									OpCode_RESOLVE_VALUE,
									OpCode_RESOLVE_VALUE,
									OpCode_PUSH_CONSTANT,
									OpCode_CLOSE_FRAME,
									OpCode_JOIN_STRINGS,
								}))
								Expect(program.Constants).To(Equal([]Value{
									STR("this "),
									STR("var1"),
									STR(" a string"),
								}))

								variableResolver.register("var1", STR("var2"))
								variableResolver.register("var2", STR("is"))
								Expect(evaluate(program)).To(Equal(STR("this is a string")))
							})
							Specify("triple substitution", func() {
								script := parse(`"this $$$var1 a string"`)
								program := compileFirstWord(script)
								Expect(program.OpCodes).To(Equal([]OpCode{
									OpCode_OPEN_FRAME,
									OpCode_PUSH_CONSTANT,
									OpCode_PUSH_CONSTANT,
									OpCode_RESOLVE_VALUE,
									OpCode_RESOLVE_VALUE,
									OpCode_RESOLVE_VALUE,
									OpCode_PUSH_CONSTANT,
									OpCode_CLOSE_FRAME,
									OpCode_JOIN_STRINGS,
								}))
								Expect(program.Constants).To(Equal([]Value{
									STR("this "),
									STR("var1"),
									STR(" a string"),
								}))

								variableResolver.register("var1", STR("var2"))
								variableResolver.register("var2", STR("var3"))
								variableResolver.register("var3", STR("is"))
								Expect(evaluate(program)).To(Equal(STR("this is a string")))
							})
						})

						Describe("blocks", func() {
							Specify("varname with spaces", func() {
								script := parse(`"this ${variable name} a string"`)
								program := compileFirstWord(script)
								Expect(program.OpCodes).To(Equal([]OpCode{
									OpCode_OPEN_FRAME,
									OpCode_PUSH_CONSTANT,
									OpCode_PUSH_CONSTANT,
									OpCode_RESOLVE_VALUE,
									OpCode_PUSH_CONSTANT,
									OpCode_CLOSE_FRAME,
									OpCode_JOIN_STRINGS,
								}))
								Expect(program.Constants).To(Equal([]Value{
									STR("this "),
									STR("variable name"),
									STR(" a string"),
								}))

								variableResolver.register("variable name", STR("is"))
								Expect(evaluate(program)).To(Equal(STR("this is a string")))
							})
							Specify("varname with special characters", func() {
								script := parse(
									`"this ${variable " " name} a string"`,
								)
								program := compileFirstWord(script)
								Expect(program.OpCodes).To(Equal([]OpCode{
									OpCode_OPEN_FRAME,
									OpCode_PUSH_CONSTANT,
									OpCode_PUSH_CONSTANT,
									OpCode_RESOLVE_VALUE,
									OpCode_PUSH_CONSTANT,
									OpCode_CLOSE_FRAME,
									OpCode_JOIN_STRINGS,
								}))
								Expect(program.Constants).To(Equal([]Value{
									STR("this "),
									STR(`variable " " name`),
									STR(" a string"),
								}))

								variableResolver.register(`variable " " name`, STR("is"))
								Expect(evaluate(program)).To(Equal(STR("this is a string")))
							})
							Specify("double substitution", func() {
								script := parse(`"this $${variable name} a string"`)
								program := compileFirstWord(script)
								Expect(program.OpCodes).To(Equal([]OpCode{
									OpCode_OPEN_FRAME,
									OpCode_PUSH_CONSTANT,
									OpCode_PUSH_CONSTANT,
									OpCode_RESOLVE_VALUE,
									OpCode_RESOLVE_VALUE,
									OpCode_PUSH_CONSTANT,
									OpCode_CLOSE_FRAME,
									OpCode_JOIN_STRINGS,
								}))
								Expect(program.Constants).To(Equal([]Value{
									STR("this "),
									STR("variable name"),
									STR(" a string"),
								}))

								variableResolver.register("variable name", STR("var2"))
								variableResolver.register("var2", STR("is"))
								Expect(evaluate(program)).To(Equal(STR("this is a string")))
							})
						})

						Describe("expressions", func() {
							Specify("simple substitution", func() {
								script := parse(`"this $[cmd] a string"`)
								program := compileFirstWord(script)
								Expect(program.OpCodes).To(Equal([]OpCode{
									OpCode_OPEN_FRAME,
									OpCode_PUSH_CONSTANT,
									OpCode_OPEN_FRAME,
									OpCode_PUSH_CONSTANT,
									OpCode_CLOSE_FRAME,
									OpCode_EVALUATE_SENTENCE,
									OpCode_PUSH_RESULT,
									OpCode_PUSH_CONSTANT,
									OpCode_CLOSE_FRAME,
									OpCode_JOIN_STRINGS,
								}))
								Expect(program.Constants).To(Equal([]Value{
									STR("this "),
									STR("cmd"),
									STR(" a string"),
								}))

								// commandResolver.register(
								//   "cmd",
								//   new FunctionCommand(() => STR("is"))
								// );
								// Expect(evaluate(program)).To(Equal(STR("this is a string"));
							})
							Specify("double substitution", func() {
								script := parse(`"this $$[cmd] a string"`)
								program := compileFirstWord(script)
								Expect(program.OpCodes).To(Equal([]OpCode{
									OpCode_OPEN_FRAME,
									OpCode_PUSH_CONSTANT,
									OpCode_OPEN_FRAME,
									OpCode_PUSH_CONSTANT,
									OpCode_CLOSE_FRAME,
									OpCode_EVALUATE_SENTENCE,
									OpCode_PUSH_RESULT,
									OpCode_RESOLVE_VALUE,
									OpCode_PUSH_CONSTANT,
									OpCode_CLOSE_FRAME,
									OpCode_JOIN_STRINGS,
								}))
								Expect(program.Constants).To(Equal([]Value{
									STR("this "),
									STR("cmd"),
									STR(" a string"),
								}))

								// commandResolver.register(
								//   "cmd",
								//   new FunctionCommand(() => STR("var"))
								// );
								// variableResolver.register("var", STR("is"));
								// Expect(evaluate(program)).To(Equal(STR("this is a string"));
							})
						})
					})

					Describe("indexed selectors", func() {
						Specify("simple substitution", func() {
							script := parse(`"this $varname[1] a string"`)
							program := compileFirstWord(script)
							Expect(program.OpCodes).To(Equal([]OpCode{
								OpCode_OPEN_FRAME,
								OpCode_PUSH_CONSTANT,
								OpCode_PUSH_CONSTANT,
								OpCode_RESOLVE_VALUE,
								OpCode_OPEN_FRAME,
								OpCode_PUSH_CONSTANT,
								OpCode_CLOSE_FRAME,
								OpCode_EVALUATE_SENTENCE,
								OpCode_PUSH_RESULT,
								OpCode_SELECT_INDEX,
								OpCode_PUSH_CONSTANT,
								OpCode_CLOSE_FRAME,
								OpCode_JOIN_STRINGS,
							}))
							Expect(program.Constants).To(Equal([]Value{
								STR("this "),
								STR("varname"),
								STR("1"),
								STR(" a string"),
							}))

							// variableResolver.register(
							// 	"varname",
							// 	LIST([]Value{STR("value"), STR("is")}),
							// )
							// Expect(evaluate(program)).To(Equal(STR("this is a string")))
						})
						Specify("double substitution", func() {
							script := parse(`"this $$var1[0] a string"`)
							program := compileFirstWord(script)
							Expect(program.OpCodes).To(Equal([]OpCode{
								OpCode_OPEN_FRAME,
								OpCode_PUSH_CONSTANT,
								OpCode_PUSH_CONSTANT,
								OpCode_RESOLVE_VALUE,
								OpCode_OPEN_FRAME,
								OpCode_PUSH_CONSTANT,
								OpCode_CLOSE_FRAME,
								OpCode_EVALUATE_SENTENCE,
								OpCode_PUSH_RESULT,
								OpCode_SELECT_INDEX,
								OpCode_RESOLVE_VALUE,
								OpCode_PUSH_CONSTANT,
								OpCode_CLOSE_FRAME,
								OpCode_JOIN_STRINGS,
							}))
							Expect(program.Constants).To(Equal([]Value{
								STR("this "),
								STR("var1"),
								STR("0"),
								STR(" a string"),
							}))

							// variableResolver.register("var1", LIST([]Value{STR("var2")}))
							// variableResolver.register("var2", STR("is"))
							// Expect(evaluate(program)).To(Equal(STR("this is a string")))
						})
						Specify("successive indexes", func() {
							script := parse(`"this $varname[1][0] a string"`)
							program := compileFirstWord(script)
							Expect(program.OpCodes).To(Equal([]OpCode{
								OpCode_OPEN_FRAME,
								OpCode_PUSH_CONSTANT,
								OpCode_PUSH_CONSTANT,
								OpCode_RESOLVE_VALUE,
								OpCode_OPEN_FRAME,
								OpCode_PUSH_CONSTANT,
								OpCode_CLOSE_FRAME,
								OpCode_EVALUATE_SENTENCE,
								OpCode_PUSH_RESULT,
								OpCode_SELECT_INDEX,
								OpCode_OPEN_FRAME,
								OpCode_PUSH_CONSTANT,
								OpCode_CLOSE_FRAME,
								OpCode_EVALUATE_SENTENCE,
								OpCode_PUSH_RESULT,
								OpCode_SELECT_INDEX,
								OpCode_PUSH_CONSTANT,
								OpCode_CLOSE_FRAME,
								OpCode_JOIN_STRINGS,
							}))
							Expect(program.Constants).To(Equal([]Value{
								STR("this "),
								STR("varname"),
								STR("1"),
								STR("0"),
								STR(" a string"),
							}))

							// variableResolver.register(
							//   "varname",
							//   LIST([STR("value1"), LIST([STR("is"), STR("value2")])])
							// );
							// Expect(evaluate(program)).To(Equal(STR("this is a string"));
						})
					})

					Describe("keyed selectors", func() {
						Specify("simple substitution", func() {
							script := parse(`"this $varname(key) a string"`)
							program := compileFirstWord(script)
							Expect(program.OpCodes).To(Equal([]OpCode{
								OpCode_OPEN_FRAME,
								OpCode_PUSH_CONSTANT,
								OpCode_PUSH_CONSTANT,
								OpCode_RESOLVE_VALUE,
								OpCode_OPEN_FRAME,
								OpCode_PUSH_CONSTANT,
								OpCode_CLOSE_FRAME,
								OpCode_SELECT_KEYS,
								OpCode_PUSH_CONSTANT,
								OpCode_CLOSE_FRAME,
								OpCode_JOIN_STRINGS,
							}))
							Expect(program.Constants).To(Equal([]Value{
								STR("this "),
								STR("varname"),
								STR("key"),
								STR(" a string"),
							}))

							variableResolver.register(
								"varname",
								DICT(map[string]Value{
									"key": STR("is"),
								}),
							)
							Expect(evaluate(program)).To(Equal(STR("this is a string")))
						})
						Specify("double substitution", func() {
							script := parse(`"this $$var1(key) a string"`)
							program := compileFirstWord(script)
							Expect(program.OpCodes).To(Equal([]OpCode{
								OpCode_OPEN_FRAME,
								OpCode_PUSH_CONSTANT,
								OpCode_PUSH_CONSTANT,
								OpCode_RESOLVE_VALUE,
								OpCode_OPEN_FRAME,
								OpCode_PUSH_CONSTANT,
								OpCode_CLOSE_FRAME,
								OpCode_SELECT_KEYS,
								OpCode_RESOLVE_VALUE,
								OpCode_PUSH_CONSTANT,
								OpCode_CLOSE_FRAME,
								OpCode_JOIN_STRINGS,
							}))
							Expect(program.Constants).To(Equal([]Value{
								STR("this "),
								STR("var1"),
								STR("key"),
								STR(" a string"),
							}))

							variableResolver.register("var1", DICT(map[string]Value{"key": STR("var2")}))
							variableResolver.register("var2", STR("is"))
							Expect(evaluate(program)).To(Equal(STR("this is a string")))
						})
						Specify("successive keys", func() {
							script := parse(`"this $varname(key1)(key2) a string"`)
							program := compileFirstWord(script)
							Expect(program.OpCodes).To(Equal([]OpCode{
								OpCode_OPEN_FRAME,
								OpCode_PUSH_CONSTANT,
								OpCode_PUSH_CONSTANT,
								OpCode_RESOLVE_VALUE,
								OpCode_OPEN_FRAME,
								OpCode_PUSH_CONSTANT,
								OpCode_CLOSE_FRAME,
								OpCode_SELECT_KEYS,
								OpCode_OPEN_FRAME,
								OpCode_PUSH_CONSTANT,
								OpCode_CLOSE_FRAME,
								OpCode_SELECT_KEYS,
								OpCode_PUSH_CONSTANT,
								OpCode_CLOSE_FRAME,
								OpCode_JOIN_STRINGS,
							}))
							Expect(program.Constants).To(Equal([]Value{
								STR("this "),
								STR("varname"),
								STR("key1"),
								STR("key2"),
								STR(" a string"),
							}))

							variableResolver.register(
								"varname",
								DICT(map[string]Value{
									"key1": DICT(map[string]Value{"key2": STR("is")}),
								}),
							)
							Expect(evaluate(program)).To(Equal(STR("this is a string")))
						})
					})

					Describe("custom selectors", func() {
						//                 const lastSelector = {
						//                   apply(value: Value): Result {
						//                     const list = value as ListValue;
						//                     return OK(list.values[list.values.length - 1]);
						//                   },
						//                 };
						//                 Specify("simple substitution", func() {
						//                    script := parse('"this $varname{last} a string"');
						//                    program := compileFirstWord(script);
						//                   Expect(program.OpCodes).To(Equal([]OpCode{
						//                     OpCode_OPEN_FRAME,
						//                     OpCode_PUSH_CONSTANT,
						//                     OpCode_PUSH_CONSTANT,
						//                     OpCode_RESOLVE_VALUE,
						//                     OpCode_OPEN_FRAME,
						//                     OpCode_OPEN_FRAME,
						//                     OpCode_PUSH_CONSTANT,
						//                     OpCode_CLOSE_FRAME,
						//                     OpCode_CLOSE_FRAME,
						//                     OpCode_SELECT_RULES,
						//                     OpCode_PUSH_CONSTANT,
						//                     OpCode_CLOSE_FRAME,
						//                     OpCode_JOIN_STRINGS,
						//                   }));
						//                   Expect(program.Constants).To(Equal([]Value{
						//                     STR("this "),
						//                     STR("varname"),
						//                     STR("last"),
						//                     STR(" a string"),
						//                   }));

						//                   variableResolver.register(
						//                     "varname",
						//                     LIST([STR("value1"), STR("value2"), STR("is")])
						//                   );
						//                   selectorResolver.register(() => OK(NIL, lastSelector));
						//                   Expect(evaluate(program)).To(Equal(STR("this is a string"));
						//                 });
						//                 Specify("double substitution", func() {
						//                    script := parse('"this $$var1{last} a string"');
						//                    program := compileFirstWord(script);
						//                   Expect(program.OpCodes).To(Equal([]OpCode{
						//                     OpCode_OPEN_FRAME,
						//                     OpCode_PUSH_CONSTANT,
						//                     OpCode_PUSH_CONSTANT,
						//                     OpCode_RESOLVE_VALUE,
						//                     OpCode_OPEN_FRAME,
						//                     OpCode_OPEN_FRAME,
						//                     OpCode_PUSH_CONSTANT,
						//                     OpCode_CLOSE_FRAME,
						//                     OpCode_CLOSE_FRAME,
						//                     OpCode_SELECT_RULES,
						//                     OpCode_RESOLVE_VALUE,
						//                     OpCode_PUSH_CONSTANT,
						//                     OpCode_CLOSE_FRAME,
						//                     OpCode_JOIN_STRINGS,
						//                   }));
						//                   Expect(program.Constants).To(Equal([]Value{
						//                     STR("this "),
						//                     STR("var1"),
						//                     STR("last"),
						//                     STR(" a string"),
						//                   }));

						//                   variableResolver.register(
						//                     "var1",
						//                     LIST([STR("var2"), STR("var3")])
						//                   );
						//                   variableResolver.register("var3", STR("is"));
						//                   selectorResolver.register(() => OK(NIL, lastSelector));
						//                   Expect(evaluate(program)).To(Equal(STR("this is a string"));
						//                 });
						//                 Specify("successive selectors", func() {
						//                    script := parse('"this $var{last}{last} a string"');
						//                    program := compileFirstWord(script);
						//                   Expect(program.OpCodes).To(Equal([]OpCode{
						//                     OpCode_OPEN_FRAME,
						//                     OpCode_PUSH_CONSTANT,
						//                     OpCode_PUSH_CONSTANT,
						//                     OpCode_RESOLVE_VALUE,
						//                     OpCode_OPEN_FRAME,
						//                     OpCode_OPEN_FRAME,
						//                     OpCode_PUSH_CONSTANT,
						//                     OpCode_CLOSE_FRAME,
						//                     OpCode_CLOSE_FRAME,
						//                     OpCode_SELECT_RULES,
						//                     OpCode_OPEN_FRAME,
						//                     OpCode_OPEN_FRAME,
						//                     OpCode_PUSH_CONSTANT,
						//                     OpCode_CLOSE_FRAME,
						//                     OpCode_CLOSE_FRAME,
						//                     OpCode_SELECT_RULES,
						//                     OpCode_PUSH_CONSTANT,
						//                     OpCode_CLOSE_FRAME,
						//                     OpCode_JOIN_STRINGS,
						//                   }));
						//                   Expect(program.Constants).To(Equal([]Value{
						//                     STR("this "),
						//                     STR("var"),
						//                     STR("last"),
						//                     STR("last"),
						//                     STR(" a string"),
						//                   }));

						//                   variableResolver.register(
						//                     "var",
						//                     LIST([STR("value1"), LIST([STR("value2"), STR("is")])])
						//                   );
						//                   selectorResolver.register(() => OK(NIL, lastSelector));
						//                   Expect(evaluate(program)).To(Equal(STR("this is a string"));
						//                 });
						//                 Describe("exceptions", func() {
						//                   Specify("unresolved selector", func() {
						//                      script := parse('"this $varname{last} a string"');
						//                      program := compileFirstWord(script);
						//                     Expect(program.OpCodes).To(Equal([]OpCode{
						//                       OpCode_OPEN_FRAME,
						//                       OpCode_PUSH_CONSTANT,
						//                       OpCode_PUSH_CONSTANT,
						//                       OpCode_RESOLVE_VALUE,
						//                       OpCode_OPEN_FRAME,
						//                       OpCode_OPEN_FRAME,
						//                       OpCode_PUSH_CONSTANT,
						//                       OpCode_CLOSE_FRAME,
						//                       OpCode_CLOSE_FRAME,
						//                       OpCode_SELECT_RULES,
						//                       OpCode_PUSH_CONSTANT,
						//                       OpCode_CLOSE_FRAME,
						//                       OpCode_JOIN_STRINGS,
						//                     }));
						//                     Expect(program.Constants).To(Equal([]Value{
						//                       STR("this "),
						//                       STR("varname"),
						//                       STR("last"),
						//                       STR(" a string"),
						//                     }));

						//                     variableResolver.register(
						//                       "varname",
						//                       LIST([STR("value1"), STR("value2"), STR("is")])
						//                     );
						//                     Expect(execute(program)).To(Equal(
						//                       ERROR("cannot resolve selector {(last)}")
						//                     );
						//                   });
						//                 });
					})

					Specify("string with multiple substitutions", func() {
						script := parse(
							`"this $$var1$${variable 2} [cmd1] with subst[cmd2]${var3}[cmd3]$var4"`,
						)
						program := compileFirstWord(script)

						Expect(program.OpCodes).To(Equal([]OpCode{
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_PUSH_CONSTANT,
							OpCode_RESOLVE_VALUE,
							OpCode_RESOLVE_VALUE,
							OpCode_PUSH_CONSTANT,
							OpCode_RESOLVE_VALUE,
							OpCode_RESOLVE_VALUE,
							OpCode_PUSH_CONSTANT,
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_CLOSE_FRAME,
							OpCode_EVALUATE_SENTENCE,
							OpCode_PUSH_RESULT,
							OpCode_PUSH_CONSTANT,
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_CLOSE_FRAME,
							OpCode_EVALUATE_SENTENCE,
							OpCode_PUSH_RESULT,
							OpCode_PUSH_CONSTANT,
							OpCode_RESOLVE_VALUE,
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_CLOSE_FRAME,
							OpCode_EVALUATE_SENTENCE,
							OpCode_PUSH_RESULT,
							OpCode_SELECT_INDEX,
							OpCode_PUSH_CONSTANT,
							OpCode_RESOLVE_VALUE,
							OpCode_CLOSE_FRAME,
							OpCode_JOIN_STRINGS,
						}))
						Expect(program.Constants).To(Equal([]Value{
							STR("this "),
							STR("var1"),
							STR("variable 2"),
							STR(" "),
							STR("cmd1"),
							STR(" with subst"),
							STR("cmd2"),
							STR("var3"),
							STR("cmd3"),
							STR("var4"),
						}))

						//                 variableResolver.register("var1", STR("var5"));
						//                 variableResolver.register("var5", STR("is"));
						//                 variableResolver.register("variable 2", STR("var6"));
						//                 variableResolver.register("var6", STR(" a"));
						//                 commandResolver.register(
						//                   "cmd1",
						//                   new FunctionCommand(() => STR("string"))
						//                 );
						//                 commandResolver.register(
						//                   "cmd2",
						//                   new FunctionCommand(() => STR("it"))
						//                 );
						//                 variableResolver.register(
						//                   "var3",
						//                   LIST([STR("foo"), STR("ut")])
						//                 );
						//                 commandResolver.register(
						//                   "cmd3",
						//                   new FunctionCommand(() => INT(1))
						//                 );
						//                 variableResolver.register("var4", STR("ions"));
						//                 Expect(evaluate(program)).To(Equal(
						//                   STR("this is a string with substitutions")
						//                 );
					})
				})

				Specify("here-strings", func() {
					script := parse("\"\"\"this is a \"'\\ $ \nhere-string\"\"\"")
					program := compileFirstWord(script)
					Expect(program.OpCodes).To(Equal([]OpCode{OpCode_PUSH_CONSTANT}))
					Expect(program.Constants).To(Equal([]Value{
						STR("this is a \"'\\ $ \nhere-string"),
					}))

					Expect(evaluate(program)).To(Equal(
						STR("this is a \"'\\ $ \nhere-string"),
					))
				})

				Specify("tagged strings", func() {
					script := parse(
						"\"\"SOME_TAG\nthis is \n a \n \"'\\ $ tagged string\nSOME_TAG\"\"",
					)
					program := compileFirstWord(script)
					Expect(program.OpCodes).To(Equal([]OpCode{OpCode_PUSH_CONSTANT}))
					Expect(program.Constants).To(Equal([]Value{
						STR("this is \n a \n \"'\\ $ tagged string\n"),
					}))

					Expect(evaluate(program)).To(Equal(
						STR("this is \n a \n \"'\\ $ tagged string\n"),
					))
				})
			})

			Describe("compounds", func() {
				Specify("literal prefix", func() {
					script := parse("this_${var}(key)_a_[cmd a b]_compound")
					program := compileFirstWord(script)
					Expect(program.OpCodes).To(Equal([]OpCode{
						OpCode_OPEN_FRAME,
						OpCode_PUSH_CONSTANT,
						OpCode_PUSH_CONSTANT,
						OpCode_RESOLVE_VALUE,
						OpCode_OPEN_FRAME,
						OpCode_PUSH_CONSTANT,
						OpCode_CLOSE_FRAME,
						OpCode_SELECT_KEYS,
						OpCode_PUSH_CONSTANT,
						OpCode_OPEN_FRAME,
						OpCode_PUSH_CONSTANT,
						OpCode_PUSH_CONSTANT,
						OpCode_PUSH_CONSTANT,
						OpCode_CLOSE_FRAME,
						OpCode_EVALUATE_SENTENCE,
						OpCode_PUSH_RESULT,
						OpCode_PUSH_CONSTANT,
						OpCode_CLOSE_FRAME,
						OpCode_JOIN_STRINGS,
					}))
					Expect(program.Constants).To(Equal([]Value{
						STR("this_"),
						STR("var"),
						STR("key"),
						STR("_a_"),
						STR("cmd"),
						STR("a"),
						STR("b"),
						STR("_compound"),
					}))

					// variableResolver.register("var", DICT(map[string]Value{ key: STR("is") }));
					// commandResolver.register(
					//   "cmd",
					//   new FunctionCommand(() => STR("literal-prefixed"))
					// );
					// Expect(evaluate(program)).To(Equal(
					//   STR("this_is_a_literal-prefixed_compound")
					// );
				})
				Specify("expression prefix", func() {
					script := parse("[cmd a b]_is_an_${var}(key)_compound")
					program := compileFirstWord(script)
					Expect(program.OpCodes).To(Equal([]OpCode{
						OpCode_OPEN_FRAME,
						OpCode_OPEN_FRAME,
						OpCode_PUSH_CONSTANT,
						OpCode_PUSH_CONSTANT,
						OpCode_PUSH_CONSTANT,
						OpCode_CLOSE_FRAME,
						OpCode_EVALUATE_SENTENCE,
						OpCode_PUSH_RESULT,
						OpCode_PUSH_CONSTANT,
						OpCode_PUSH_CONSTANT,
						OpCode_RESOLVE_VALUE,
						OpCode_OPEN_FRAME,
						OpCode_PUSH_CONSTANT,
						OpCode_CLOSE_FRAME,
						OpCode_SELECT_KEYS,
						OpCode_PUSH_CONSTANT,
						OpCode_CLOSE_FRAME,
						OpCode_JOIN_STRINGS,
					}))
					Expect(program.Constants).To(Equal([]Value{
						STR("cmd"),
						STR("a"),
						STR("b"),
						STR("_is_an_"),
						STR("var"),
						STR("key"),
						STR("_compound"),
					}))

					// commandResolver.register(
					//   "cmd",
					//   new FunctionCommand(() => STR("this"))
					// );
					// variableResolver.register(
					//   "var",
					//   DICT(map[string]Value{ key: STR("expression-prefixed") })
					// );
					// Expect(evaluate(program)).To(Equal(
					//   STR("this_is_an_expression-prefixed_compound")
					// );
				})
				Specify("substitution prefix", func() {
					script := parse("${var}(key)_is_a_[cmd a b]_compound")
					program := compileFirstWord(script)
					Expect(program.OpCodes).To(Equal([]OpCode{
						OpCode_OPEN_FRAME,
						OpCode_PUSH_CONSTANT,
						OpCode_RESOLVE_VALUE,
						OpCode_OPEN_FRAME,
						OpCode_PUSH_CONSTANT,
						OpCode_CLOSE_FRAME,
						OpCode_SELECT_KEYS,
						OpCode_PUSH_CONSTANT,
						OpCode_OPEN_FRAME,
						OpCode_PUSH_CONSTANT,
						OpCode_PUSH_CONSTANT,
						OpCode_PUSH_CONSTANT,
						OpCode_CLOSE_FRAME,
						OpCode_EVALUATE_SENTENCE,
						OpCode_PUSH_RESULT,
						OpCode_PUSH_CONSTANT,
						OpCode_CLOSE_FRAME,
						OpCode_JOIN_STRINGS,
					}))
					Expect(program.Constants).To(Equal([]Value{
						STR("var"),
						STR("key"),
						STR("_is_a_"),
						STR("cmd"),
						STR("a"),
						STR("b"),
						STR("_compound"),
					}))

					// variableResolver.register("var", DICT(map[string]Value{ key: STR("this") }));
					// commandResolver.register(
					//   "cmd",
					//   new FunctionCommand(() => STR("substitution-prefixed"))
					// );
					// Expect(evaluate(program)).To(Equal(
					//   STR("this_is_a_substitution-prefixed_compound")
					// );
				})
			})

			Describe("substitutions", func() {
				Describe("scalars", func() {
					Specify("simple substitution", func() {
						script := parse("$varname")
						program := compileFirstWord(script)
						Expect(program.OpCodes).To(Equal([]OpCode{
							OpCode_PUSH_CONSTANT,
							OpCode_RESOLVE_VALUE,
						}))
						Expect(program.Constants).To(Equal([]Value{STR("varname")}))

						variableResolver.register("varname", STR("value"))
						Expect(evaluate(program)).To(Equal(STR("value")))
					})
					Specify("double substitution", func() {
						script := parse("$$var1")
						program := compileFirstWord(script)
						Expect(program.OpCodes).To(Equal([]OpCode{
							OpCode_PUSH_CONSTANT,
							OpCode_RESOLVE_VALUE,
							OpCode_RESOLVE_VALUE,
						}))
						Expect(program.Constants).To(Equal([]Value{STR("var1")}))

						variableResolver.register("var1", STR("var2"))
						variableResolver.register("var2", STR("value"))
						Expect(evaluate(program)).To(Equal(STR("value")))
					})
					Specify("triple substitution", func() {
						script := parse("$$$var1")
						program := compileFirstWord(script)
						Expect(program.OpCodes).To(Equal([]OpCode{
							OpCode_PUSH_CONSTANT,
							OpCode_RESOLVE_VALUE,
							OpCode_RESOLVE_VALUE,
							OpCode_RESOLVE_VALUE,
						}))
						Expect(program.Constants).To(Equal([]Value{STR("var1")}))

						variableResolver.register("var1", STR("var2"))
						variableResolver.register("var2", STR("var3"))
						variableResolver.register("var3", STR("value"))
						Expect(evaluate(program)).To(Equal(STR("value")))
					})
				})

				Describe("tuples", func() {
					Specify("single variable", func() {
						script := parse("$(varname)")
						program := compileFirstWord(script)
						Expect(program.OpCodes).To(Equal([]OpCode{
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_CLOSE_FRAME,
							OpCode_RESOLVE_VALUE,
						}))
						Expect(program.Constants).To(Equal([]Value{STR("varname")}))

						variableResolver.register("varname", STR("value"))
						Expect(evaluate(program)).To(Equal(TUPLE([]Value{STR("value")})))
					})
					Specify("multiple variables", func() {
						script := parse("$(var1 var2)")
						program := compileFirstWord(script)
						Expect(program.OpCodes).To(Equal([]OpCode{
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_PUSH_CONSTANT,
							OpCode_CLOSE_FRAME,
							OpCode_RESOLVE_VALUE,
						}))
						Expect(program.Constants).To(Equal([]Value{STR("var1"), STR("var2")}))

						variableResolver.register("var1", STR("value1"))
						variableResolver.register("var2", STR("value2"))
						Expect(evaluate(program)).To(Equal(
							TUPLE([]Value{STR("value1"), STR("value2")}),
						))
					})
					Specify("double substitution", func() {
						script := parse("$$(var1)")
						program := compileFirstWord(script)
						Expect(program.OpCodes).To(Equal([]OpCode{
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_CLOSE_FRAME,
							OpCode_RESOLVE_VALUE,
							OpCode_RESOLVE_VALUE,
						}))
						Expect(program.Constants).To(Equal([]Value{STR("var1")}))

						variableResolver.register("var1", STR("var2"))
						variableResolver.register("var2", STR("value"))
						Expect(evaluate(program)).To(Equal(TUPLE([]Value{STR("value")})))
					})
					Specify("nested tuples", func() {
						script := parse("$(var1 (var2))")
						program := compileFirstWord(script)
						Expect(program.OpCodes).To(Equal([]OpCode{
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_CLOSE_FRAME,
							OpCode_CLOSE_FRAME,
							OpCode_RESOLVE_VALUE,
						}))
						Expect(program.Constants).To(Equal([]Value{STR("var1"), STR("var2")}))

						variableResolver.register("var1", STR("value1"))
						variableResolver.register("var2", STR("value2"))
						Expect(evaluate(program)).To(Equal(
							TUPLE([]Value{STR("value1"), TUPLE([]Value{STR("value2")})}),
						))
					})
					Specify("nested double substitution", func() {
						script := parse("$$((var1))")
						program := compileFirstWord(script)
						Expect(program.OpCodes).To(Equal([]OpCode{
							OpCode_OPEN_FRAME,
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_CLOSE_FRAME,
							OpCode_CLOSE_FRAME,
							OpCode_RESOLVE_VALUE,
							OpCode_RESOLVE_VALUE,
						}))
						Expect(program.Constants).To(Equal([]Value{STR("var1")}))

						variableResolver.register("var1", STR("var2"))
						variableResolver.register("var2", STR("value"))
						Expect(evaluate(program)).To(Equal(
							TUPLE([]Value{TUPLE([]Value{STR("value")})}),
						))
					})
				})

				Describe("blocks", func() {
					Specify("varname with spaces", func() {
						script := parse("${variable name}")
						program := compileFirstWord(script)
						Expect(program.OpCodes).To(Equal([]OpCode{
							OpCode_PUSH_CONSTANT,
							OpCode_RESOLVE_VALUE,
						}))
						Expect(program.Constants).To(Equal([]Value{STR("variable name")}))

						variableResolver.register("variable name", STR("value"))
						Expect(evaluate(program)).To(Equal(STR("value")))
					})
					Specify("varname with special characters", func() {
						script := parse(`${variable " " name}`)
						program := compileFirstWord(script)
						Expect(program.OpCodes).To(Equal([]OpCode{
							OpCode_PUSH_CONSTANT,
							OpCode_RESOLVE_VALUE,
						}))
						Expect(program.Constants).To(Equal([]Value{STR(`variable " " name`)}))

						variableResolver.register(`variable " " name`, STR("value"))
						Expect(evaluate(program)).To(Equal(STR("value")))
					})
					Specify("varname with continuations", func() {
						script := parse("${variable\\\n \t\r     name}")
						program := compileFirstWord(script)
						Expect(program.OpCodes).To(Equal([]OpCode{
							OpCode_PUSH_CONSTANT,
							OpCode_RESOLVE_VALUE,
						}))
						Expect(program.Constants).To(Equal([]Value{STR("variable name")}))

						variableResolver.register("variable name", STR("value"))
						variableResolver.register(`variable " " name`, STR("value"))
						Expect(evaluate(program)).To(Equal(STR("value")))
					})
				})

				Describe("expressions", func() {
					Specify("simple substitution", func() {
						script := parse("$[cmd]")
						program := compileFirstWord(script)
						Expect(program.OpCodes).To(Equal([]OpCode{
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_CLOSE_FRAME,
							OpCode_EVALUATE_SENTENCE,
							OpCode_PUSH_RESULT,
						}))
						Expect(program.Constants).To(Equal([]Value{STR("cmd")}))

						//                 commandResolver.register(
						//                   "cmd",
						//                   new FunctionCommand(() =>
						//                     LIST([STR("value1"), STR("value2")])
						//                   )
						//                 );
						//                 Expect(evaluate(program)).To(Equal(
						//                   LIST([STR("value1"), STR("value2")])
						//                 );
					})
					Specify("double substitution, scalar", func() {
						script := parse("$$[cmd]")
						program := compileFirstWord(script)
						Expect(program.OpCodes).To(Equal([]OpCode{
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_CLOSE_FRAME,
							OpCode_EVALUATE_SENTENCE,
							OpCode_PUSH_RESULT,
							OpCode_RESOLVE_VALUE,
						}))
						Expect(program.Constants).To(Equal([]Value{STR("cmd")}))

						//                 commandResolver.register(
						//                   "cmd",
						//                   new FunctionCommand(() => STR("var"))
						//                 );
						//                 variableResolver.register("var", STR("value"));
						//                 Expect(evaluate(program)).To(Equal(STR("value"));
					})
					Specify("double substitution, tuple", func() {
						script := parse("$$[cmd]")
						program := compileFirstWord(script)
						Expect(program.OpCodes).To(Equal([]OpCode{
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_CLOSE_FRAME,
							OpCode_EVALUATE_SENTENCE,
							OpCode_PUSH_RESULT,
							OpCode_RESOLVE_VALUE,
						}))
						Expect(program.Constants).To(Equal([]Value{STR("cmd")}))

						//                 commandResolver.register(
						//                   "cmd",
						//                   new FunctionCommand(() => TUPLE([STR("var1"), STR("var2")]))
						//                 );
						//                 variableResolver.register("var1", STR("value1"));
						//                 variableResolver.register("var2", STR("value2"));
						//                 Expect(evaluate(program)).To(Equal(
						//                   TUPLE([STR("value1"), STR("value2")])
						//                 );
					})
					Specify("two sentences", func() {
						script := parse("[cmd1 result1; cmd2 result2]")
						program := compileFirstWord(script)
						Expect(program.OpCodes).To(Equal([]OpCode{
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_PUSH_CONSTANT,
							OpCode_CLOSE_FRAME,
							OpCode_EVALUATE_SENTENCE,
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_PUSH_CONSTANT,
							OpCode_CLOSE_FRAME,
							OpCode_EVALUATE_SENTENCE,
							OpCode_PUSH_RESULT,
						}))
						Expect(program.Constants).To(Equal([]Value{
							STR("cmd1"),
							STR("result1"),
							STR("cmd2"),
							STR("result2"),
						}))

						//                 const called = {};
						//                 const fn: Command = new FunctionCommand((args) => {
						//                   const cmd = asString(args[0]);
						//                   called[cmd] = called[cmd] ?? 0 + 1;
						//                   return args[1];
						//                 });
						//                 commandResolver.register("cmd1", fn);
						//                 commandResolver.register("cmd2", fn);
						//                 Expect(evaluate(program)).To(Equal(STR("result2"));
						//                 Expect(called).To(Equal({ cmd1: 1, cmd2: 1 });
					})
					Specify("indirect command", func() {
						script := parse("[$cmdname]")
						program := compileFirstWord(script)
						Expect(program.OpCodes).To(Equal([]OpCode{
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_RESOLVE_VALUE,
							OpCode_CLOSE_FRAME,
							OpCode_EVALUATE_SENTENCE,
							OpCode_PUSH_RESULT,
						}))
						Expect(program.Constants).To(Equal([]Value{STR("cmdname")}))

						// variableResolver.register("cmdname", STR("cmd"));
						// commandResolver.register(
						//   "cmd",
						//   new FunctionCommand(() => STR("value"))
						// );
						// Expect(evaluate(program)).To(Equal(STR("value"));
					})
				})

				Describe("indexed selectors", func() {
					Specify("simple substitution", func() {
						script := parse("$varname[1]")
						program := compileFirstWord(script)
						Expect(program.OpCodes).To(Equal([]OpCode{
							OpCode_PUSH_CONSTANT,
							OpCode_RESOLVE_VALUE,
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_CLOSE_FRAME,
							OpCode_EVALUATE_SENTENCE,
							OpCode_PUSH_RESULT,
							OpCode_SELECT_INDEX,
						}))
						Expect(program.Constants).To(Equal([]Value{STR("varname"), STR("1")}))

						// variableResolver.register(
						//   "varname",
						//   LIST([STR("value1"), STR("value2")])
						// );
						// Expect(evaluate(program)).To(Equal(STR("value2"));
					})
					Specify("double substitution", func() {
						script := parse("$$var1[0]")
						program := compileFirstWord(script)
						Expect(program.OpCodes).To(Equal([]OpCode{
							OpCode_PUSH_CONSTANT,
							OpCode_RESOLVE_VALUE,
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_CLOSE_FRAME,
							OpCode_EVALUATE_SENTENCE,
							OpCode_PUSH_RESULT,
							OpCode_SELECT_INDEX,
							OpCode_RESOLVE_VALUE,
						}))
						Expect(program.Constants).To(Equal([]Value{STR("var1"), STR("0")}))

						// variableResolver.register("var1", LIST([STR("var2")]));
						// variableResolver.register("var2", STR("value"));
						// Expect(evaluate(program)).To(Equal(STR("value"));
					})
					Specify("successive indexes", func() {
						script := parse("$varname[1][0]")
						program := compileFirstWord(script)
						Expect(program.OpCodes).To(Equal([]OpCode{
							OpCode_PUSH_CONSTANT,
							OpCode_RESOLVE_VALUE,
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_CLOSE_FRAME,
							OpCode_EVALUATE_SENTENCE,
							OpCode_PUSH_RESULT,
							OpCode_SELECT_INDEX,
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_CLOSE_FRAME,
							OpCode_EVALUATE_SENTENCE,
							OpCode_PUSH_RESULT,
							OpCode_SELECT_INDEX,
						}))
						Expect(program.Constants).To(Equal([]Value{
							STR("varname"),
							STR("1"),
							STR("0"),
						}))

						// variableResolver.register(
						//   "varname",
						//   LIST([
						//     STR("value1"),
						//     LIST([STR("value2_1"), STR("value2_2")]),
						//   ])
						// );
						// Expect(evaluate(program)).To(Equal(STR("value2_1"));
					})
					Specify("indirect index", func() {
						script := parse("$var1[$var2]")
						program := compileFirstWord(script)
						Expect(program.OpCodes).To(Equal([]OpCode{
							OpCode_PUSH_CONSTANT,
							OpCode_RESOLVE_VALUE,
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_RESOLVE_VALUE,
							OpCode_CLOSE_FRAME,
							OpCode_EVALUATE_SENTENCE,
							OpCode_PUSH_RESULT,
							OpCode_SELECT_INDEX,
						}))
						Expect(program.Constants).To(Equal([]Value{STR("var1"), STR("var2")}))

						// variableResolver.register(
						//   "var1",
						//   LIST([STR("value1"), STR("value2"), STR("value3")])
						// );
						// variableResolver.register("var2", STR("1"));
						// Expect(evaluate(program)).To(Equal(STR("value2"));
					})
					Specify("command index", func() {
						script := parse("$varname[cmd]")
						program := compileFirstWord(script)
						Expect(program.OpCodes).To(Equal([]OpCode{
							OpCode_PUSH_CONSTANT,
							OpCode_RESOLVE_VALUE,
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_CLOSE_FRAME,
							OpCode_EVALUATE_SENTENCE,
							OpCode_PUSH_RESULT,
							OpCode_SELECT_INDEX,
						}))
						Expect(program.Constants).To(Equal([]Value{STR("varname"), STR("cmd")}))

						// commandResolver.register(
						//   "cmd",
						//   new FunctionCommand(() => STR("1"))
						// );
						// variableResolver.register(
						//   "varname",
						//   LIST([STR("value1"), STR("value2"), STR("value3")])
						// );
						// Expect(evaluate(program)).To(Equal(STR("value2"));
					})
					Specify("scalar expression", func() {
						script := parse("$[cmd][0]")
						program := compileFirstWord(script)
						Expect(program.OpCodes).To(Equal([]OpCode{
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_CLOSE_FRAME,
							OpCode_EVALUATE_SENTENCE,
							OpCode_PUSH_RESULT,
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_CLOSE_FRAME,
							OpCode_EVALUATE_SENTENCE,
							OpCode_PUSH_RESULT,
							OpCode_SELECT_INDEX,
						}))
						Expect(program.Constants).To(Equal([]Value{STR("cmd"), STR("0")}))

						// commandResolver.register(
						//   "cmd",
						//   new FunctionCommand(() => LIST([STR("value")]))
						// );
						// Expect(evaluate(program)).To(Equal(STR("value"));
					})
					Specify("tuple expression", func() {
						script := parse("$[cmd][0]")
						program := compileFirstWord(script)
						Expect(program.OpCodes).To(Equal([]OpCode{
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_CLOSE_FRAME,
							OpCode_EVALUATE_SENTENCE,
							OpCode_PUSH_RESULT,
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_CLOSE_FRAME,
							OpCode_EVALUATE_SENTENCE,
							OpCode_PUSH_RESULT,
							OpCode_SELECT_INDEX,
						}))
						Expect(program.Constants).To(Equal([]Value{STR("cmd"), STR("0")}))

						// commandResolver.register(
						//   "cmd",
						//   new FunctionCommand(() =>
						//     TUPLE([LIST([STR("value1")]), LIST([STR("value2")])])
						//   )
						// );
						// Expect(evaluate(program)).To(Equal(
						//   TUPLE([STR("value1"), STR("value2")])
						// );
					})
				})

				Describe("keyed selectors", func() {
					Specify("simple substitution", func() {
						script := parse("$varname(key)")
						program := compileFirstWord(script)
						Expect(program.OpCodes).To(Equal([]OpCode{
							OpCode_PUSH_CONSTANT,
							OpCode_RESOLVE_VALUE,
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_CLOSE_FRAME,
							OpCode_SELECT_KEYS,
						}))
						Expect(program.Constants).To(Equal([]Value{STR("varname"), STR("key")}))

						variableResolver.register(
							"varname",
							DICT(map[string]Value{
								"key": STR("value"),
							}),
						)
						Expect(evaluate(program)).To(Equal(STR("value")))
					})
					Specify("double substitution", func() {
						script := parse("$$var1(key)")
						program := compileFirstWord(script)
						Expect(program.OpCodes).To(Equal([]OpCode{
							OpCode_PUSH_CONSTANT,
							OpCode_RESOLVE_VALUE,
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_CLOSE_FRAME,
							OpCode_SELECT_KEYS,
							OpCode_RESOLVE_VALUE,
						}))
						Expect(program.Constants).To(Equal([]Value{STR("var1"), STR("key")}))

						variableResolver.register("var1", DICT(map[string]Value{"key": STR("var2")}))
						variableResolver.register("var2", STR("value"))
						Expect(evaluate(program)).To(Equal(STR("value")))
					})
					Specify("recursive keys", func() {
						script := parse("$varname(key1 key2)")
						program := compileFirstWord(script)
						Expect(program.OpCodes).To(Equal([]OpCode{
							OpCode_PUSH_CONSTANT,
							OpCode_RESOLVE_VALUE,
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_PUSH_CONSTANT,
							OpCode_CLOSE_FRAME,
							OpCode_SELECT_KEYS,
						}))
						Expect(program.Constants).To(Equal([]Value{
							STR("varname"),
							STR("key1"),
							STR("key2"),
						}))

						variableResolver.register(
							"varname",
							DICT(map[string]Value{
								"key1": DICT(map[string]Value{"key2": STR("value")}),
							}),
						)
						Expect(evaluate(program)).To(Equal(STR("value")))
					})
					Specify("successive keys", func() {
						script := parse("$varname(key1)(key2)")
						program := compileFirstWord(script)
						Expect(program.OpCodes).To(Equal([]OpCode{
							OpCode_PUSH_CONSTANT,
							OpCode_RESOLVE_VALUE,
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_CLOSE_FRAME,
							OpCode_SELECT_KEYS,
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_CLOSE_FRAME,
							OpCode_SELECT_KEYS,
						}))
						Expect(program.Constants).To(Equal([]Value{
							STR("varname"),
							STR("key1"),
							STR("key2"),
						}))

						variableResolver.register(
							"varname",
							DICT(map[string]Value{
								"key1": DICT(map[string]Value{"key2": STR("value")}),
							}),
						)
						Expect(evaluate(program)).To(Equal(STR("value")))
					})
					Specify("indirect key", func() {
						script := parse("$var1($var2)")
						program := compileFirstWord(script)
						Expect(program.OpCodes).To(Equal([]OpCode{
							OpCode_PUSH_CONSTANT,
							OpCode_RESOLVE_VALUE,
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_RESOLVE_VALUE,
							OpCode_CLOSE_FRAME,
							OpCode_SELECT_KEYS,
						}))
						Expect(program.Constants).To(Equal([]Value{STR("var1"), STR("var2")}))

						variableResolver.register(
							"var1",
							DICT(map[string]Value{
								"key": STR("value"),
							}),
						)
						variableResolver.register("var2", STR("key"))
						Expect(evaluate(program)).To(Equal(STR("value")))
					})
					Specify("string key", func() {
						script := parse(`$varname("arbitrary key")`)
						program := compileFirstWord(script)
						Expect(program.OpCodes).To(Equal([]OpCode{
							OpCode_PUSH_CONSTANT,
							OpCode_RESOLVE_VALUE,
							OpCode_OPEN_FRAME,
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_CLOSE_FRAME,
							OpCode_JOIN_STRINGS,
							OpCode_CLOSE_FRAME,
							OpCode_SELECT_KEYS,
						}))
						Expect(program.Constants).To(Equal([]Value{
							STR("varname"),
							STR("arbitrary key"),
						}))

						variableResolver.register(
							"varname",
							DICT(map[string]Value{
								"arbitrary key": STR("value"),
							}),
						)
						Expect(evaluate(program)).To(Equal(STR("value")))
					})
					Specify("block key", func() {
						script := parse("$varname({arbitrary key})")
						program := compileFirstWord(script)
						Expect(program.OpCodes).To(Equal([]OpCode{
							OpCode_PUSH_CONSTANT,
							OpCode_RESOLVE_VALUE,
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_CLOSE_FRAME,
							OpCode_SELECT_KEYS,
						}))
						Expect(program.Constants).To(Equal([]Value{
							STR("varname"),
							NewScriptValue(*parse("arbitrary key"), "arbitrary key"),
						}))

						variableResolver.register(
							"varname",
							DICT(map[string]Value{
								"arbitrary key": STR("value"),
							}),
						)
						Expect(evaluate(program)).To(Equal(STR("value")))
					})
					Specify("tuple", func() {
						script := parse("$(var1 var2)(key)")
						program := compileFirstWord(script)
						Expect(program.OpCodes).To(Equal([]OpCode{
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_PUSH_CONSTANT,
							OpCode_CLOSE_FRAME,
							OpCode_RESOLVE_VALUE,
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_CLOSE_FRAME,
							OpCode_SELECT_KEYS,
						}))
						Expect(program.Constants).To(Equal([]Value{
							STR("var1"),
							STR("var2"),
							STR("key"),
						}))

						variableResolver.register("var1", DICT(map[string]Value{"key": STR("value1")}))
						variableResolver.register("var2", DICT(map[string]Value{"key": STR("value2")}))
						Expect(evaluate(program)).To(Equal(
							TUPLE([]Value{STR("value1"), STR("value2")}),
						))
					})
					Specify("recursive tuple", func() {
						script := parse("$(var1 (var2))(key)")
						program := compileFirstWord(script)
						Expect(program.OpCodes).To(Equal([]OpCode{
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_CLOSE_FRAME,
							OpCode_CLOSE_FRAME,
							OpCode_RESOLVE_VALUE,
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_CLOSE_FRAME,
							OpCode_SELECT_KEYS,
						}))
						Expect(program.Constants).To(Equal([]Value{
							STR("var1"),
							STR("var2"),
							STR("key"),
						}))

						variableResolver.register("var1", DICT(map[string]Value{"key": STR("value1")}))
						variableResolver.register("var2", DICT(map[string]Value{"key": STR("value2")}))
						Expect(evaluate(program)).To(Equal(
							TUPLE([]Value{STR("value1"), TUPLE([]Value{STR("value2")})}),
						))
					})
					Specify("tuple with double substitution", func() {
						script := parse("$$(var1 var2)(key)")
						program := compileFirstWord(script)
						Expect(program.OpCodes).To(Equal([]OpCode{
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_PUSH_CONSTANT,
							OpCode_CLOSE_FRAME,
							OpCode_RESOLVE_VALUE,
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_CLOSE_FRAME,
							OpCode_SELECT_KEYS,
							OpCode_RESOLVE_VALUE,
						}))
						Expect(program.Constants).To(Equal([]Value{
							STR("var1"),
							STR("var2"),
							STR("key"),
						}))

						variableResolver.register("var1", DICT(map[string]Value{"key": STR("var3")}))
						variableResolver.register("var2", DICT(map[string]Value{"key": STR("var4")}))
						variableResolver.register("var3", STR("value3"))
						variableResolver.register("var4", STR("value4"))
						Expect(evaluate(program)).To(Equal(
							TUPLE([]Value{STR("value3"), STR("value4")}),
						))
					})
					Specify("scalar expression", func() {
						script := parse("$[cmd](key)")
						program := compileFirstWord(script)
						Expect(program.OpCodes).To(Equal([]OpCode{
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_CLOSE_FRAME,
							OpCode_EVALUATE_SENTENCE,
							OpCode_PUSH_RESULT,
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_CLOSE_FRAME,
							OpCode_SELECT_KEYS,
						}))
						Expect(program.Constants).To(Equal([]Value{STR("cmd"), STR("key")}))

						// commandResolver.register(
						//   "cmd",
						//   new FunctionCommand(() => DICT(map[string]Value{ key: STR("value") }))
						// );
						// Expect(evaluate(program)).To(Equal(STR("value"));
					})
					Specify("tuple expression", func() {
						script := parse("$[cmd](key)")
						program := compileFirstWord(script)
						Expect(program.OpCodes).To(Equal([]OpCode{
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_CLOSE_FRAME,
							OpCode_EVALUATE_SENTENCE,
							OpCode_PUSH_RESULT,
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_CLOSE_FRAME,
							OpCode_SELECT_KEYS,
						}))
						Expect(program.Constants).To(Equal([]Value{STR("cmd"), STR("key")}))

						// commandResolver.register(
						//   "cmd",
						//   new FunctionCommand(() =>
						//     TUPLE([
						//       DICT(map[string]Value{ key: STR("value1") }),
						//       DICT(map[string]Value{ key: STR("value2") }),
						//     ])
						//   )
						// );
						// Expect(evaluate(program)).To(Equal(
						//   TUPLE([STR("value1"), STR("value2")])
						// );
					})
				})

				Describe("custom selectors", func() {
					//               const lastSelector = {
					//                 apply(value: Value): Result {
					//                   const list = value as ListValue;
					//                   return OK(list.values[list.values.length - 1]);
					//                 },
					//               };

					Specify("simple substitution", func() {
						script := parse("$varname{last}")
						program := compileFirstWord(script)
						Expect(program.OpCodes).To(Equal([]OpCode{
							OpCode_PUSH_CONSTANT,
							OpCode_RESOLVE_VALUE,
							OpCode_OPEN_FRAME,
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_CLOSE_FRAME,
							OpCode_CLOSE_FRAME,
							OpCode_SELECT_RULES,
						}))
						Expect(program.Constants).To(Equal([]Value{STR("varname"), STR("last")}))

						// variableResolver.register(
						//   "varname",
						//   LIST([STR("value1"), STR("value2"), STR("value3")])
						// );
						// selectorResolver.register(() => OK(NIL, lastSelector));
						// Expect(evaluate(program)).To(Equal(STR("value3"));
					})
					Specify("double substitution", func() {
						script := parse("$$var1{last}")
						program := compileFirstWord(script)
						Expect(program.OpCodes).To(Equal([]OpCode{
							OpCode_PUSH_CONSTANT,
							OpCode_RESOLVE_VALUE,
							OpCode_OPEN_FRAME,
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_CLOSE_FRAME,
							OpCode_CLOSE_FRAME,
							OpCode_SELECT_RULES,
							OpCode_RESOLVE_VALUE,
						}))
						Expect(program.Constants).To(Equal([]Value{STR("var1"), STR("last")}))

						//                 variableResolver.register(
						//                   "var1",
						//                   LIST([STR("var2"), STR("var3")])
						//                 );
						//                 variableResolver.register("var3", STR("value"));
						//                 selectorResolver.register(() => OK(NIL, lastSelector));
						//                 Expect(evaluate(program)).To(Equal(STR("value"));
					})
					Specify("successive selectors", func() {
						script := parse("$var{last}{last}")
						program := compileFirstWord(script)
						Expect(program.OpCodes).To(Equal([]OpCode{
							OpCode_PUSH_CONSTANT,
							OpCode_RESOLVE_VALUE,
							OpCode_OPEN_FRAME,
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_CLOSE_FRAME,
							OpCode_CLOSE_FRAME,
							OpCode_SELECT_RULES,
							OpCode_OPEN_FRAME,
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_CLOSE_FRAME,
							OpCode_CLOSE_FRAME,
							OpCode_SELECT_RULES,
						}))
						Expect(program.Constants).To(Equal([]Value{
							STR("var"),
							STR("last"),
							STR("last"),
						}))

						//                 variableResolver.register(
						//                   "var",
						//                   LIST([
						//                     STR("value1"),
						//                     LIST([STR("value2_1"), STR("value2_2")]),
						//                   ])
						//                 );
						//                 selectorResolver.register(() => OK(NIL, lastSelector));
						//                 Expect(evaluate(program)).To(Equal(STR("value2_2"));
					})
					Specify("indirect selector", func() {
						script := parse("$var1{$var2}")
						program := compileFirstWord(script)
						Expect(program.OpCodes).To(Equal([]OpCode{
							OpCode_PUSH_CONSTANT,
							OpCode_RESOLVE_VALUE,
							OpCode_OPEN_FRAME,
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_RESOLVE_VALUE,
							OpCode_CLOSE_FRAME,
							OpCode_CLOSE_FRAME,
							OpCode_SELECT_RULES,
						}))
						Expect(program.Constants).To(Equal([]Value{STR("var1"), STR("var2")}))

						//                 variableResolver.register(
						//                   "var1",
						//                   LIST([STR("value1"), STR("value2"), STR("value3")])
						//                 );
						//                 variableResolver.register("var2", STR("last"));
						//                 selectorResolver.register(() => OK(NIL, lastSelector));
						//                 Expect(evaluate(program)).To(Equal(STR("value3"));
					})
					Specify("expression", func() {
						script := parse("$[cmd]{last}")
						program := compileFirstWord(script)
						Expect(program.OpCodes).To(Equal([]OpCode{
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_CLOSE_FRAME,
							OpCode_EVALUATE_SENTENCE,
							OpCode_PUSH_RESULT,
							OpCode_OPEN_FRAME,
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_CLOSE_FRAME,
							OpCode_CLOSE_FRAME,
							OpCode_SELECT_RULES,
						}))
						Expect(program.Constants).To(Equal([]Value{STR("cmd"), STR("last")}))

						//                 commandResolver.register(
						//                   "cmd",
						//                   new FunctionCommand(() =>
						//                     LIST([STR("value1"), STR("value2")])
						//                   )
						//                 );
						//                 selectorResolver.register(() => OK(NIL, lastSelector));
						//                 Expect(evaluate(program)).To(Equal(STR("value2"));
					})
					//               Describe("exceptions", func() {
					//                 Specify("unresolved selector", func() {
					//                    script := parse("$varname{last}");
					//                    program := compileFirstWord(script);
					//                   Expect(program.OpCodes).To(Equal([]OpCode{
					//                     OpCode_PUSH_CONSTANT,
					//                     OpCode_RESOLVE_VALUE,
					//                     OpCode_OPEN_FRAME,
					//                     OpCode_OPEN_FRAME,
					//                     OpCode_PUSH_CONSTANT,
					//                     OpCode_CLOSE_FRAME,
					//                     OpCode_CLOSE_FRAME,
					//                     OpCode_SELECT_RULES,
					//                   }));
					//                   Expect(program.Constants).To(Equal([]Value{
					//                     STR("varname"),
					//                     STR("last"),
					//                   }));

					//                   variableResolver.register(
					//                     "varname",
					//                     LIST([STR("value1"), STR("value2"), STR("value3")])
					//                   );
					//                   Expect(execute(program)).To(Equal(
					//                     ERROR("cannot resolve selector {(last)}")
					//                   );
					//                 });
					//               });
				})

				Describe("exceptions", func() {
					Specify("unresolved variable", func() {
						script := parse("$varname")
						program := compileFirstWord(script)
						Expect(program.OpCodes).To(Equal([]OpCode{
							OpCode_PUSH_CONSTANT,
							OpCode_RESOLVE_VALUE,
						}))
						Expect(program.Constants).To(Equal([]Value{STR("varname")}))

						Expect(execute(program)).To(Equal(
							ERROR(`cannot resolve variable "varname"`),
						))
					})
				})
			})

			Describe("qualified words", func() {
				Describe("literal prefix", func() {
					Specify("indexed selector", func() {
						script := parse("varname[cmd]")
						program := compileFirstWord(script)
						Expect(program.OpCodes).To(Equal([]OpCode{
							OpCode_PUSH_CONSTANT,
							OpCode_SET_SOURCE,
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_CLOSE_FRAME,
							OpCode_EVALUATE_SENTENCE,
							OpCode_PUSH_RESULT,
							OpCode_SELECT_INDEX,
						}))
						Expect(program.Constants).To(Equal([]Value{STR("varname"), STR("cmd")}))

						// commandResolver.register(
						//   "cmd",
						//   new FunctionCommand(() => STR("index"))
						// );
						// Expect(evaluate(program)).To(Equal(
						//   new QualifiedValue(STR("varname"), [
						//     new IndexedSelector(STR("index")),
						//   ])
						// );
					})
					Specify("keyed selector", func() {
						script := parse("varname(key1 key2)")
						program := compileFirstWord(script)
						Expect(program.OpCodes).To(Equal([]OpCode{
							OpCode_PUSH_CONSTANT,
							OpCode_SET_SOURCE,
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_PUSH_CONSTANT,
							OpCode_CLOSE_FRAME,
							OpCode_SELECT_KEYS,
						}))
						Expect(program.Constants).To(Equal([]Value{
							STR("varname"),
							STR("key1"),
							STR("key2"),
						}))

						Expect(evaluate(program)).To(Equal(
							NewQualifiedValue(STR("varname"), []Selector{
								NewKeyedSelector([]Value{STR("key1"), STR("key2")}),
							}),
						))
					})
					Specify("generic selector", func() {
						script := parse("varname{rule1 arg1; rule2 arg2}")
						program := compileFirstWord(script)
						Expect(program.OpCodes).To(Equal([]OpCode{
							OpCode_PUSH_CONSTANT,
							OpCode_SET_SOURCE,
							OpCode_OPEN_FRAME,
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_PUSH_CONSTANT,
							OpCode_CLOSE_FRAME,
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_PUSH_CONSTANT,
							OpCode_CLOSE_FRAME,
							OpCode_CLOSE_FRAME,
							OpCode_SELECT_RULES,
						}))
						Expect(program.Constants).To(Equal([]Value{
							STR("varname"),
							STR("rule1"),
							STR("arg1"),
							STR("rule2"),
							STR("arg2"),
						}))

						// selectorResolver.register((rules) =>
						//   GenericSelector.create(rules)
						// );
						// Expect(evaluate(program)).To(Equal(
						//   new QualifiedValue(STR("varname"), [
						//     new GenericSelector([
						//       TUPLE([STR("rule1"), STR("arg1")]),
						//       TUPLE([STR("rule2"), STR("arg2")]),
						//     ]),
						//   ])
						// );
					})
					Specify("complex case", func() {
						script := parse(
							"varname(key1 $var1){$var2; [cmd1]}[cmd2]([$var3])(key4)",
						)
						program := compileFirstWord(script)
						Expect(program.OpCodes).To(Equal([]OpCode{
							OpCode_PUSH_CONSTANT,
							OpCode_SET_SOURCE,
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_PUSH_CONSTANT,
							OpCode_RESOLVE_VALUE,
							OpCode_CLOSE_FRAME,
							OpCode_SELECT_KEYS,
							OpCode_OPEN_FRAME,
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_RESOLVE_VALUE,
							OpCode_CLOSE_FRAME,
							OpCode_OPEN_FRAME,
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_CLOSE_FRAME,
							OpCode_EVALUATE_SENTENCE,
							OpCode_PUSH_RESULT,
							OpCode_CLOSE_FRAME,
							OpCode_CLOSE_FRAME,
							OpCode_SELECT_RULES,
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_CLOSE_FRAME,
							OpCode_EVALUATE_SENTENCE,
							OpCode_PUSH_RESULT,
							OpCode_SELECT_INDEX,
							OpCode_OPEN_FRAME,
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_RESOLVE_VALUE,
							OpCode_CLOSE_FRAME,
							OpCode_EVALUATE_SENTENCE,
							OpCode_PUSH_RESULT,
							OpCode_CLOSE_FRAME,
							OpCode_SELECT_KEYS,
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_CLOSE_FRAME,
							OpCode_SELECT_KEYS,
						}))
						Expect(program.Constants).To(Equal([]Value{
							STR("varname"),
							STR("key1"),
							STR("var1"),
							STR("var2"),
							STR("cmd1"),
							STR("cmd2"),
							STR("var3"),
							STR("key4"),
						}))

						//                 variableResolver.register("var1", STR("key2"));
						//                 variableResolver.register("var2", STR("rule1"));
						//                 variableResolver.register("var3", STR("cmd3"));
						//                 commandResolver.register(
						//                   "cmd1",
						//                   new FunctionCommand(() => STR("rule2"))
						//                 );
						//                 commandResolver.register(
						//                   "cmd2",
						//                   new FunctionCommand(() => STR("index1"))
						//                 );
						//                 commandResolver.register(
						//                   "cmd3",
						//                   new FunctionCommand(() => STR("key3"))
						//                 );
						//                 selectorResolver.register((rules) =>
						//                   GenericSelector.create(rules)
						//                 );
						//                 Expect(evaluate(program)).To(Equal(
						//                   new QualifiedValue(STR("varname"), [
						//                     new KeyedSelector([STR("key1"), STR("key2")]),
						//                     new GenericSelector([
						//                       TUPLE([STR("rule1")]),
						//                       TUPLE([STR("rule2")]),
						//                     ]),
						//                     new IndexedSelector(STR("index1")),
						//                     new KeyedSelector([STR("key3"), STR("key4")]),
						//                   ])
						//                 );
					})
				})
				Describe("tuple prefix", func() {
					Specify("indexed selector", func() {
						script := parse("(varname1 varname2)[cmd]")
						program := compileFirstWord(script)
						Expect(program.OpCodes).To(Equal([]OpCode{
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_PUSH_CONSTANT,
							OpCode_CLOSE_FRAME,
							OpCode_SET_SOURCE,
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_CLOSE_FRAME,
							OpCode_EVALUATE_SENTENCE,
							OpCode_PUSH_RESULT,
							OpCode_SELECT_INDEX,
						}))
						Expect(program.Constants).To(Equal([]Value{
							STR("varname1"),
							STR("varname2"),
							STR("cmd"),
						}))

						// commandResolver.register(
						//   "cmd",
						//   new FunctionCommand(() => STR("index"))
						// );
						// Expect(evaluate(program)).To(Equal(
						//   new QualifiedValue(
						//     TUPLE([STR("varname1"), STR("varname2")]),
						//     [new IndexedSelector(STR("index"))]
						//   )
						// );
					})
					Specify("keyed selector", func() {
						script := parse("(varname1 varname2)(key1 key2)")
						program := compileFirstWord(script)
						Expect(program.OpCodes).To(Equal([]OpCode{
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_PUSH_CONSTANT,
							OpCode_CLOSE_FRAME,
							OpCode_SET_SOURCE,
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_PUSH_CONSTANT,
							OpCode_CLOSE_FRAME,
							OpCode_SELECT_KEYS,
						}))
						Expect(program.Constants).To(Equal([]Value{
							STR("varname1"),
							STR("varname2"),
							STR("key1"),
							STR("key2"),
						}))

						Expect(evaluate(program)).To(Equal(
							NewQualifiedValue(
								TUPLE([]Value{STR("varname1"), STR("varname2")}),
								[]Selector{NewKeyedSelector([]Value{STR("key1"), STR("key2")})},
							),
						))
					})
					Specify("generic selector", func() {
						script := parse(
							"(varname1 varname2){rule1 arg1; rule2 arg2}",
						)
						program := compileFirstWord(script)
						Expect(program.OpCodes).To(Equal([]OpCode{
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_PUSH_CONSTANT,
							OpCode_CLOSE_FRAME,
							OpCode_SET_SOURCE,
							OpCode_OPEN_FRAME,
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_PUSH_CONSTANT,
							OpCode_CLOSE_FRAME,
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_PUSH_CONSTANT,
							OpCode_CLOSE_FRAME,
							OpCode_CLOSE_FRAME,
							OpCode_SELECT_RULES,
						}))
						Expect(program.Constants).To(Equal([]Value{
							STR("varname1"),
							STR("varname2"),
							STR("rule1"),
							STR("arg1"),
							STR("rule2"),
							STR("arg2"),
						}))

						// selectorResolver.register((rules) =>
						//   GenericSelector.create(rules)
						// );
						// Expect(evaluate(program)).To(Equal(
						//   new QualifiedValue(
						//     TUPLE([STR("varname1"), STR("varname2")]),
						//     [
						//       new GenericSelector([
						//         TUPLE([STR("rule1"), STR("arg1")]),
						//         TUPLE([STR("rule2"), STR("arg2")]),
						//       ]),
						//     ]
						//   )
						// );
					})
					Specify("complex case", func() {
						script := parse(
							"(varname1 $var1)[cmd1](key1 $var2)([$var3]){$var4; [cmd2]}[cmd4]",
						)
						program := compileFirstWord(script)
						Expect(program.OpCodes).To(Equal([]OpCode{
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_PUSH_CONSTANT,
							OpCode_RESOLVE_VALUE,
							OpCode_CLOSE_FRAME,
							OpCode_SET_SOURCE,
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_CLOSE_FRAME,
							OpCode_EVALUATE_SENTENCE,
							OpCode_PUSH_RESULT,
							OpCode_SELECT_INDEX,
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_PUSH_CONSTANT,
							OpCode_RESOLVE_VALUE,
							OpCode_CLOSE_FRAME,
							OpCode_SELECT_KEYS,
							OpCode_OPEN_FRAME,
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_RESOLVE_VALUE,
							OpCode_CLOSE_FRAME,
							OpCode_EVALUATE_SENTENCE,
							OpCode_PUSH_RESULT,
							OpCode_CLOSE_FRAME,
							OpCode_SELECT_KEYS,
							OpCode_OPEN_FRAME,
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_RESOLVE_VALUE,
							OpCode_CLOSE_FRAME,
							OpCode_OPEN_FRAME,
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_CLOSE_FRAME,
							OpCode_EVALUATE_SENTENCE,
							OpCode_PUSH_RESULT,
							OpCode_CLOSE_FRAME,
							OpCode_CLOSE_FRAME,
							OpCode_SELECT_RULES,
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_CLOSE_FRAME,
							OpCode_EVALUATE_SENTENCE,
							OpCode_PUSH_RESULT,
							OpCode_SELECT_INDEX,
						}))
						Expect(program.Constants).To(Equal([]Value{
							STR("varname1"),
							STR("var1"),
							STR("cmd1"),
							STR("key1"),
							STR("var2"),
							STR("var3"),
							STR("var4"),
							STR("cmd2"),
							STR("cmd4"),
						}))

						//                 variableResolver.register("var1", STR("varname2"));
						//                 variableResolver.register("var2", STR("key2"));
						//                 variableResolver.register("var3", STR("cmd3"));
						//                 variableResolver.register("var4", STR("rule1"));
						//                 commandResolver.register(
						//                   "cmd1",
						//                   new FunctionCommand(() => STR("index1"))
						//                 );
						//                 commandResolver.register(
						//                   "cmd2",
						//                   new FunctionCommand(() => STR("rule2"))
						//                 );
						//                 commandResolver.register(
						//                   "cmd3",
						//                   new FunctionCommand(() => STR("key3"))
						//                 );
						//                 commandResolver.register(
						//                   "cmd4",
						//                   new FunctionCommand(() => STR("index2"))
						//                 );
						//                 selectorResolver.register((rules) =>
						//                   GenericSelector.create(rules)
						//                 );
						//                 Expect(evaluate(program)).To(Equal(
						//                   new QualifiedValue(
						//                     TUPLE([STR("varname1"), STR("varname2")]),
						//                     [
						//                       new IndexedSelector(STR("index1")),
						//                       new KeyedSelector([
						//                         STR("key1"),
						//                         STR("key2"),
						//                         STR("key3"),
						//                       ]),
						//                       new GenericSelector([
						//                         TUPLE([STR("rule1")]),
						//                         TUPLE([STR("rule2")]),
						//                       ]),
						//                       new IndexedSelector(STR("index2")),
						//                     ]
						//                   )
						//                 );
					})
				})
				Describe("block prefix", func() {
					Specify("indexed selector", func() {
						script := parse("{source name}[cmd]")
						program := compileFirstWord(script)
						Expect(program.OpCodes).To(Equal([]OpCode{
							OpCode_PUSH_CONSTANT,
							OpCode_SET_SOURCE,
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_CLOSE_FRAME,
							OpCode_EVALUATE_SENTENCE,
							OpCode_PUSH_RESULT,
							OpCode_SELECT_INDEX,
						}))
						Expect(program.Constants).To(Equal([]Value{
							STR("source name"),
							STR("cmd"),
						}))

						//                 commandResolver.register(
						//                   "cmd",
						//                   new FunctionCommand(() => STR("index"))
						//                 );
						//                 Expect(evaluate(program)).To(Equal(
						//                   new QualifiedValue(STR("source name"), [
						//                     new IndexedSelector(STR("index")),
						//                   ])
						//                 );
					})
					Specify("keyed selector", func() {
						script := parse("{source name}(key1 key2)")
						program := compileFirstWord(script)
						Expect(program.OpCodes).To(Equal([]OpCode{
							OpCode_PUSH_CONSTANT,
							OpCode_SET_SOURCE,
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_PUSH_CONSTANT,
							OpCode_CLOSE_FRAME,
							OpCode_SELECT_KEYS,
						}))
						Expect(program.Constants).To(Equal([]Value{
							STR("source name"),
							STR("key1"),
							STR("key2"),
						}))

						Expect(evaluate(program)).To(Equal(
							NewQualifiedValue(STR("source name"), []Selector{
								NewKeyedSelector([]Value{STR("key1"), STR("key2")}),
							}),
						))
					})
					Specify("generic selector", func() {
						script := parse("{source name}{rule1 arg1; rule2 arg2}")
						program := compileFirstWord(script)
						Expect(program.OpCodes).To(Equal([]OpCode{
							OpCode_PUSH_CONSTANT,
							OpCode_SET_SOURCE,
							OpCode_OPEN_FRAME,
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_PUSH_CONSTANT,
							OpCode_CLOSE_FRAME,
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_PUSH_CONSTANT,
							OpCode_CLOSE_FRAME,
							OpCode_CLOSE_FRAME,
							OpCode_SELECT_RULES,
						}))
						Expect(program.Constants).To(Equal([]Value{
							STR("source name"),
							STR("rule1"),
							STR("arg1"),
							STR("rule2"),
							STR("arg2"),
						}))

						//                 selectorResolver.register((rules) =>
						//                   GenericSelector.create(rules)
						//                 );
						//                 Expect(evaluate(program)).To(Equal(
						//                   new QualifiedValue(STR("source name"), [
						//                     new GenericSelector([
						//                       TUPLE([STR("rule1"), STR("arg1")]),
						//                       TUPLE([STR("rule2"), STR("arg2")]),
						//                     ]),
						//                   ])
						//                 );
					})
					Specify("complex case", func() {
						script := parse(
							"{source name}(key1 $var1){$var2; [cmd1]}[cmd2]([$var3])(key4)",
						)
						program := compileFirstWord(script)
						Expect(program.OpCodes).To(Equal([]OpCode{
							OpCode_PUSH_CONSTANT,
							OpCode_SET_SOURCE,
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_PUSH_CONSTANT,
							OpCode_RESOLVE_VALUE,
							OpCode_CLOSE_FRAME,
							OpCode_SELECT_KEYS,
							OpCode_OPEN_FRAME,
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_RESOLVE_VALUE,
							OpCode_CLOSE_FRAME,
							OpCode_OPEN_FRAME,
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_CLOSE_FRAME,
							OpCode_EVALUATE_SENTENCE,
							OpCode_PUSH_RESULT,
							OpCode_CLOSE_FRAME,
							OpCode_CLOSE_FRAME,
							OpCode_SELECT_RULES,
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_CLOSE_FRAME,
							OpCode_EVALUATE_SENTENCE,
							OpCode_PUSH_RESULT,
							OpCode_SELECT_INDEX,
							OpCode_OPEN_FRAME,
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_RESOLVE_VALUE,
							OpCode_CLOSE_FRAME,
							OpCode_EVALUATE_SENTENCE,
							OpCode_PUSH_RESULT,
							OpCode_CLOSE_FRAME,
							OpCode_SELECT_KEYS,
							OpCode_OPEN_FRAME,
							OpCode_PUSH_CONSTANT,
							OpCode_CLOSE_FRAME,
							OpCode_SELECT_KEYS,
						}))
						Expect(program.Constants).To(Equal([]Value{
							STR("source name"),
							STR("key1"),
							STR("var1"),
							STR("var2"),
							STR("cmd1"),
							STR("cmd2"),
							STR("var3"),
							STR("key4"),
						}))

						//                 variableResolver.register("var1", STR("key2"));
						//                 variableResolver.register("var2", STR("rule1"));
						//                 variableResolver.register("var3", STR("cmd3"));
						//                 commandResolver.register(
						//                   "cmd1",
						//                   new FunctionCommand(() => STR("rule2"))
						//                 );
						//                 commandResolver.register(
						//                   "cmd2",
						//                   new FunctionCommand(() => STR("index1"))
						//                 );
						//                 commandResolver.register(
						//                   "cmd3",
						//                   new FunctionCommand(() => STR("key3"))
						//                 );
						//                 selectorResolver.register((rules) =>
						//                   GenericSelector.create(rules)
						//                 );
						//                 Expect(evaluate(program)).To(Equal(
						//                   new QualifiedValue(STR("source name"), [
						//                     new KeyedSelector([STR("key1"), STR("key2")]),
						//                     new GenericSelector([
						//                       TUPLE([STR("rule1")]),
						//                       TUPLE([STR("rule2")]),
						//                     ]),
						//                     new IndexedSelector(STR("index1")),
						//                     new KeyedSelector([STR("key3"), STR("key4")]),
						//                   ])
						//                 );
					})
				})
			})

			Describe("ignored words", func() {
				Specify("line comments", func() {
					script := parse("# this ; is$ (\\\na [comment{")
					program := compileFirstWord(script)
					Expect(program.OpCodes).To(HaveLen(0))
				})
				Specify("block comments", func() {
					script := parse("##{ this \n ; is$ (a  \n#{comment{[( }##")
					program := compileFirstWord(script)
					Expect(program.OpCodes).To(HaveLen(0))
				})
			})
		})

		Specify("constants", func() {
			value := STR("value")
			program := compiler.CompileConstant(value)
			Expect(program.OpCodes).To(Equal([]OpCode{OpCode_PUSH_CONSTANT}))
			Expect(program.Constants).To(Equal([]Value{STR("value")}))

			Expect(evaluate(program)).To(Equal(STR("value")))
		})

		Describe("word expansion", func() {
			Specify("tuples", func() {
				script := parse("(prefix $*var suffix)")
				program := compileFirstWord(script)
				Expect(program.OpCodes).To(Equal([]OpCode{
					OpCode_OPEN_FRAME,
					OpCode_PUSH_CONSTANT,
					OpCode_PUSH_CONSTANT,
					OpCode_RESOLVE_VALUE,
					OpCode_EXPAND_VALUE,
					OpCode_PUSH_CONSTANT,
					OpCode_CLOSE_FRAME,
				}))
				Expect(program.Constants).To(Equal([]Value{
					STR("prefix"),
					STR("var"),
					STR("suffix"),
				}))

				variableResolver.register(
					"var",
					TUPLE([]Value{STR("value1"), STR("value2")}),
				)
				Expect(evaluate(program)).To(Equal(
					TUPLE([]Value{
						STR("prefix"),
						STR("value1"),
						STR("value2"),
						STR("suffix"),
					}),
				))
			})
			Specify("expressions", func() {
				script := parse("(prefix $*[cmd] suffix)")
				program := compileFirstWord(script)
				Expect(program.OpCodes).To(Equal([]OpCode{
					OpCode_OPEN_FRAME,
					OpCode_PUSH_CONSTANT,
					OpCode_OPEN_FRAME,
					OpCode_PUSH_CONSTANT,
					OpCode_CLOSE_FRAME,
					OpCode_EVALUATE_SENTENCE,
					OpCode_PUSH_RESULT,
					OpCode_EXPAND_VALUE,
					OpCode_PUSH_CONSTANT,
					OpCode_CLOSE_FRAME,
				}))
				Expect(program.Constants).To(Equal([]Value{
					STR("prefix"),
					STR("cmd"),
					STR("suffix"),
				}))

				//             commandResolver.register(
				//               "cmd",
				//               new FunctionCommand(() => TUPLE([STR("value1"), STR("value2")]))
				//             );
				//             Expect(evaluate(program)).To(Equal(
				//               TUPLE([
				//                 STR("prefix"),
				//                 STR("value1"),
				//                 STR("value2"),
				//                 STR("suffix"),
				//               ])
				//             );
			})
			Describe("scripts", func() {
				//             beforeEach(func() {
				//               commandResolver.register(
				//                 "cmd",
				//                 new FunctionCommand((args) => TUPLE(args))
				//               );
				//             });
				Specify("single variable", func() {
					script := parse("cmd $*var arg")
					program := compiler.CompileScript(*script)
					Expect(program.OpCodes).To(Equal([]OpCode{
						OpCode_OPEN_FRAME,
						OpCode_PUSH_CONSTANT,
						OpCode_PUSH_CONSTANT,
						OpCode_RESOLVE_VALUE,
						OpCode_EXPAND_VALUE,
						OpCode_PUSH_CONSTANT,
						OpCode_CLOSE_FRAME,
						OpCode_EVALUATE_SENTENCE,
						OpCode_PUSH_RESULT,
					}))
					Expect(program.Constants).To(Equal([]Value{
						STR("cmd"),
						STR("var"),
						STR("arg"),
					}))

					//               variableResolver.register(
					//                 "var",
					//                 TUPLE([STR("value1"), STR("value2")])
					//               );
					//               Expect(evaluate(program)).To(Equal(
					//                 TUPLE([STR("cmd"), STR("value1"), STR("value2"), STR("arg")])
					//               );
				})
				Specify("multiple variables", func() {
					script := parse("cmd $*(var1 var2) arg")
					program := compiler.CompileScript(*script)
					Expect(program.OpCodes).To(Equal([]OpCode{
						OpCode_OPEN_FRAME,
						OpCode_PUSH_CONSTANT,
						OpCode_OPEN_FRAME,
						OpCode_PUSH_CONSTANT,
						OpCode_PUSH_CONSTANT,
						OpCode_CLOSE_FRAME,
						OpCode_RESOLVE_VALUE,
						OpCode_EXPAND_VALUE,
						OpCode_PUSH_CONSTANT,
						OpCode_CLOSE_FRAME,
						OpCode_EVALUATE_SENTENCE,
						OpCode_PUSH_RESULT,
					}))
					Expect(program.Constants).To(Equal([]Value{
						STR("cmd"),
						STR("var1"),
						STR("var2"),
						STR("arg"),
					}))

					//               variableResolver.register("var1", STR("value1"));
					//               variableResolver.register("var2", STR("value2"));
					//               Expect(evaluate(program)).To(Equal(
					//                 TUPLE([STR("cmd"), STR("value1"), STR("value2"), STR("arg")])
					//               );
				})
				Specify("expressions", func() {
					script := parse("cmd $*[cmd2] arg")
					program := compiler.CompileScript(*script)
					Expect(program.OpCodes).To(Equal([]OpCode{
						OpCode_OPEN_FRAME,
						OpCode_PUSH_CONSTANT,
						OpCode_OPEN_FRAME,
						OpCode_PUSH_CONSTANT,
						OpCode_CLOSE_FRAME,
						OpCode_EVALUATE_SENTENCE,
						OpCode_PUSH_RESULT,
						OpCode_EXPAND_VALUE,
						OpCode_PUSH_CONSTANT,
						OpCode_CLOSE_FRAME,
						OpCode_EVALUATE_SENTENCE,
						OpCode_PUSH_RESULT,
					}))
					Expect(program.Constants).To(Equal([]Value{
						STR("cmd"),
						STR("cmd2"),
						STR("arg"),
					}))

					//               commandResolver.register(
					//                 "cmd2",
					//                 new FunctionCommand(() => TUPLE([STR("value1"), STR("value2")]))
					//               );
					//               Expect(evaluate(program)).To(Equal(
					//                 TUPLE([STR("cmd"), STR("value1"), STR("value2"), STR("arg")])
					//               );
				})
			})
		})

		Describe("scripts", func() {
			Specify("empty", func() {
				script := parse("")
				program := compiler.CompileScript(*script)
				Expect(program.OpCodes).To(HaveLen(0))
				Expect(evaluate(program)).To(Equal(NIL))
			})

			//           Specify("conditional evaluation", func() {
			//             const script1 = parse("if true {cmd1 a} else {cmd2 b}");
			//             const program1 = compiler.CompileScript(script1);
			//             Expect(program1.OpCodes).To(Equal([]OpCode{
			//               OpCode_OPEN_FRAME,
			//               OpCode_PUSH_CONSTANT,
			//               OpCode_PUSH_CONSTANT,
			//               OpCode_PUSH_CONSTANT,
			//               OpCode_PUSH_CONSTANT,
			//               OpCode_PUSH_CONSTANT,
			//               OpCode_CLOSE_FRAME,
			//               OpCode_EVALUATE_SENTENCE,
			//               OpCode_PUSH_RESULT,
			//             }));
			//             Expect(program1.Constants).To(Equal([]Value{
			//               STR("if"),
			//               STR("true"),
			//               new ScriptValue(parse("cmd1 a"), "cmd1 a"),
			//               STR("else"),
			//               new ScriptValue(parse("cmd2 b"), "cmd2 b"),
			//             }));

			//             const script2 = parse("if false {cmd1 a} else {cmd2 b}");
			//             const program2 = compiler.CompileScript(script2);
			//             Expect(program2.OpCodes).To(Equal([]OpCode{
			//               OpCode_OPEN_FRAME,
			//               OpCode_PUSH_CONSTANT,
			//               OpCode_PUSH_CONSTANT,
			//               OpCode_PUSH_CONSTANT,
			//               OpCode_PUSH_CONSTANT,
			//               OpCode_PUSH_CONSTANT,
			//               OpCode_CLOSE_FRAME,
			//               OpCode_EVALUATE_SENTENCE,
			//               OpCode_PUSH_RESULT,
			//             }));
			//             Expect(program2.Constants).To(Equal([]Value{
			//               STR("if"),
			//               STR("false"),
			//               new ScriptValue(parse("cmd1 a"), "cmd1 a"),
			//               STR("else"),
			//               new ScriptValue(parse("cmd2 b"), "cmd2 b"),
			//             }));

			//             commandResolver.register(
			//               "if",
			//               new FunctionCommand((args) => {
			//                 const condition = args[1];
			//                 const block = asString(condition) == "true" ? args[2] : args[4];
			//                 const script =
			//                   block.type == ValueType.SCRIPT
			//                     ? (block as ScriptValue).script
			//                     : parse(asString(block));
			//                  program := compiler.CompileScript(script);
			//                 return evaluate(program);
			//               })
			//             );
			//             commandResolver.register(
			//               "cmd1",
			//               new FunctionCommand((args) => args[1])
			//             );
			//             commandResolver.register(
			//               "cmd2",
			//               new FunctionCommand((args) => args[1])
			//             );

			//             Expect(evaluate(program1)).To(Equal(STR("a"));
			//             Expect(evaluate(program2)).To(Equal(STR("b"));
			//           });

			//           Specify("loop", func() {
			//              script := parse("repeat 10 {cmd foo}");
			//              program := compiler.CompileScript(script);
			//             Expect(program.OpCodes).To(Equal([]OpCode{
			//               OpCode_OPEN_FRAME,
			//               OpCode_PUSH_CONSTANT,
			//               OpCode_PUSH_CONSTANT,
			//               OpCode_PUSH_CONSTANT,
			//               OpCode_CLOSE_FRAME,
			//               OpCode_EVALUATE_SENTENCE,
			//               OpCode_PUSH_RESULT,
			//             }));
			//             Expect(program.Constants).To(Equal([]Value{
			//               STR("repeat"),
			//               STR("10"),
			//               new ScriptValue(parse("cmd foo"), "cmd foo"),
			//             }));

			//             commandResolver.register(
			//               "repeat",
			//               new FunctionCommand((args) => {
			//                 const nb = IntegerValue.toInteger(args[1]).data;
			//                 const block = args[2];
			//                 const script =
			//                   block.type == ValueType.SCRIPT
			//                     ? (block as ScriptValue).script
			//                     : parse(asString(block));
			//                  program := compiler.CompileScript(script);
			//                 let value: Value = NIL;
			//                 for (let i = 0; i < nb; i++) {
			//                   value = evaluate(program);
			//                 }
			//                 return value;
			//               })
			//             );
			//             let counter = 0;
			//             let acc = "";
			//             commandResolver.register(
			//               "cmd",
			//               new FunctionCommand((args) => {
			//                 const value = asString(args[1]);
			//                 acc += value;
			//                 return INT(counter++);
			//               })
			//             );
			//             Expect(evaluate(program)).To(Equal(INT(9));
			//             Expect(counter).To(Equal(10);
			//             Expect(acc).To(Equal("foo".repeat(10));
			//           });
		})

		Describe("sentences", func() {
			Specify("single sentence", func() {
				script := parse("cmd $*[cmd2] arg")
				program := compiler.CompileSentence(script.Sentences[0])
				Expect(program.OpCodes).To(Equal([]OpCode{
					OpCode_PUSH_CONSTANT,
					OpCode_OPEN_FRAME,
					OpCode_PUSH_CONSTANT,
					OpCode_CLOSE_FRAME,
					OpCode_EVALUATE_SENTENCE,
					OpCode_PUSH_RESULT,
					OpCode_EXPAND_VALUE,
					OpCode_PUSH_CONSTANT,
				}))
				Expect(program.Constants).To(Equal([]Value{
					STR("cmd"),
					STR("cmd2"),
					STR("arg"),
				}))
			})
			Specify("multiple sentences", func() {
				script := parse(
					"cmd1 $arg1 arg2; $*[cmd2] arg3; cmd3 $$arg4 arg5",
				)
				program := compiler.CompileSentences(script.Sentences)
				Expect(program.OpCodes).To(Equal([]OpCode{
					OpCode_OPEN_FRAME,
					OpCode_PUSH_CONSTANT,
					OpCode_PUSH_CONSTANT,
					OpCode_RESOLVE_VALUE,
					OpCode_PUSH_CONSTANT,
					OpCode_OPEN_FRAME,
					OpCode_PUSH_CONSTANT,
					OpCode_CLOSE_FRAME,
					OpCode_EVALUATE_SENTENCE,
					OpCode_PUSH_RESULT,
					OpCode_EXPAND_VALUE,
					OpCode_PUSH_CONSTANT,
					OpCode_PUSH_CONSTANT,
					OpCode_PUSH_CONSTANT,
					OpCode_RESOLVE_VALUE,
					OpCode_RESOLVE_VALUE,
					OpCode_PUSH_CONSTANT,
					OpCode_CLOSE_FRAME,
				}))
				Expect(program.Constants).To(Equal([]Value{
					STR("cmd1"),
					STR("arg1"),
					STR("arg2"),
					STR("cmd2"),
					STR("arg3"),
					STR("cmd3"),
					STR("arg4"),
					STR("arg5"),
				}))
			})
		})
	})

	Describe("Executor", func() {
		//         It("should pass opaque context to commands", func() {
		//            script := parse("cmd");
		//            program := compiler.CompileScript(script);

		//           let commandContext;
		//           commandResolver.register("cmd", {
		//             execute: (_args, context) => {
		//               commandContext = context;
		//               return OK(NIL);
		//             },
		//           });

		//           const context = Symbol();
		//           executor = new Executor(
		//             variableResolver,
		//             commandResolver,
		//             selectorResolver,
		//             context
		//           );
		//           execute(program);
		//           Expect(commandContext).to.equal(context);
		//         });

		Describe("exceptions", func() {
			//           Specify("invalid command name", func() {
			//              script := parse("[]");
			//              program := compiler.CompileScript(script);

			//             Expect(execute(program)).To(Equal(ERROR("invalid command name"));
			//           });
			Specify("invalid variable name", func() {
				script := parse("$([])")
				program := compiler.CompileScript(*script)

				Expect(execute(program)).To(Equal(ERROR("invalid variable name")))
			})
			Specify("variable substitution with no string representation", func() {
				script := parse(`"$var"`)
				program := compiler.CompileScript(*script)

				variableResolver.register("var", NIL)

				Expect(execute(program)).To(Equal(
					ERROR("value has no string representation"),
				))
			})
			//           Specify("command substitution with no string representation", func() {
			//              script := parse('"[]"');
			//              program := compiler.CompileScript(script);

			//             Expect(execute(program)).To(Equal(
			//               ERROR("value has no string representation")
			//             );
			//           });
		})
	})

	Describe("Program", func() {
		//         It("should be resumable", func() {
		//            script := parse(
		//             "break 1; ok 2; break 3; break 4; ok 5; break 6"
		//           );
		//            program := compiler.CompileScript(script);

		//           commandResolver.register("break", {
		//             execute(args) {
		//               return BREAK(args[1]);
		//             },
		//           });
		//           commandResolver.register("ok", {
		//             execute(args) {
		//               return OK(args[1]);
		//             },
		//           });
		//           const state = new ProgramState();
		//           Expect(executor.execute(program, state)).To(Equal(BREAK(STR("1")));
		//           Expect(executor.execute(program, state)).To(Equal(BREAK(STR("3")));
		//           Expect(executor.execute(program, state)).To(Equal(BREAK(STR("4")));
		//           Expect(executor.execute(program, state)).To(Equal(BREAK(STR("6")));
		//           Expect(executor.execute(program, state)).To(Equal(OK(STR("6")));
		//         });
		//         It("result should be settable", func() {
		//            script := parse("ok [break 1]");
		//            program := compiler.CompileScript(script);

		//           commandResolver.register("break", {
		//             execute(args) {
		//               return BREAK(args[1]);
		//             },
		//           });
		//           commandResolver.register("ok", {
		//             execute(args) {
		//               return OK(args[1]);
		//             },
		//           });
		//           const state = new ProgramState();
		//           Expect(executor.execute(program, state)).To(Equal(BREAK(STR("1")));
		//           state.result = OK(STR("2"));
		//           Expect(executor.execute(program, state)).To(Equal(OK(STR("2")));
		//         });
		//         It("should support resumable commands", func() {
		//            script := parse("ok [cmd]");
		//            program := compiler.CompileScript(script);

		//           commandResolver.register("cmd", {
		//             execute(_args) {
		//               return YIELD(INT(1));
		//             },
		//             resume(result) {
		//               const i = (result.value as IntegerValue).value;
		//               if (i == 5) return OK(STR("done"));
		//               return YIELD(INT(i + 1));
		//             },
		//           });
		//           commandResolver.register("ok", {
		//             execute(args) {
		//               return OK(args[1]);
		//             },
		//           });

		//           const state = new ProgramState();
		//           Expect(executor.execute(program, state)).To(Equal(YIELD(INT(1)));
		//           Expect(executor.execute(program, state)).To(Equal(YIELD(INT(2)));
		//           Expect(executor.execute(program, state)).To(Equal(YIELD(INT(3)));
		//           Expect(executor.execute(program, state)).To(Equal(YIELD(INT(4)));
		//           Expect(executor.execute(program, state)).To(Equal(YIELD(INT(5)));
		//           Expect(executor.execute(program, state)).To(Equal(OK(STR("done")));
		//         });
		//         It("should support resumable command state", func() {
		//            script := parse("ok [cmd]");
		//            program := compiler.CompileScript(script);

		//           commandResolver.register("cmd", {
		//             execute(_args) {
		//               return YIELD(STR("begin"), 1);
		//             },
		//             resume(result) {
		//               const step = result.data as number;
		//               switch (step) {
		//                 case 1:
		//                   return YIELD(STR(`step one`), step + 1);
		//                 case 2:
		//                   return YIELD(STR(`step two`), step + 1);
		//                 case 3:
		//                   return YIELD(STR(`step three`), step + 1);
		//                 case 4:
		//                   return OK(STR("end"));
		//               }
		//             },
		//           });
		//           commandResolver.register("ok", {
		//             execute(args) {
		//               return OK(args[1]);
		//             },
		//           });

		//           const state = new ProgramState();
		//           Expect(executor.execute(program, state)).To(Equal(
		//             YIELD(STR("begin"), 1)
		//           );
		//           Expect(executor.execute(program, state)).To(Equal(
		//             YIELD(STR("step one"), 2)
		//           );
		//           Expect(executor.execute(program, state)).To(Equal(
		//             YIELD(STR("step two"), 3)
		//           );
		//           Expect(executor.execute(program, state)).To(Equal(
		//             YIELD(STR("step three"), 4)
		//           );
		//           Expect(executor.execute(program, state)).To(Equal(OK(STR("end")));
		//         });
	})
	//     });
	//   }
})
