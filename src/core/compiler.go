//
// Helena script compilation
//

package core

// Supported compiler opcodes
type OpCode int8

const (
	OpCode_PUSH_NIL OpCode = iota
	OpCode_PUSH_CONSTANT
	OpCode_OPEN_FRAME
	OpCode_CLOSE_FRAME
	OpCode_RESOLVE_VALUE
	OpCode_EXPAND_VALUE
	OpCode_SET_SOURCE
	OpCode_SELECT_INDEX
	OpCode_SELECT_KEYS
	OpCode_SELECT_RULES
	OpCode_EVALUATE_SENTENCE
	OpCode_PUSH_RESULT
	OpCode_JOIN_STRINGS
)

//
// Helena program
//
type Program struct {
	// Sequence of opcodes the program is made of
	OpCodes []OpCode

	// Constants the opcodes refer to
	Constants []Value
}

// Push a new opcode
func (program *Program) PushOpCode(opCode OpCode) {
	program.OpCodes = append(program.OpCodes, opCode)
}

// Push a new constant
func (program *Program) PushConstant(value Value) {
	program.Constants = append(program.Constants, value)
}

// Report whether the program is empty
func (program *Program) Empty() bool {
	return len(program.OpCodes) == 0
}

//
// Helena compiler
//
// This class transforms scripts, sentences and words into programs
//
type Compiler struct {
	// Syntax checker used during compilation
	syntaxChecker SyntaxChecker
}

//
// Scripts
//

// Compile the given script into a program
func (compiler Compiler) CompileScript(script Script) *Program {
	program := &Program{}
	if len(script.Sentences) == 0 {
		return program
	}
	compiler.emitScript(program, script)
	return program
}
func (compiler Compiler) emitScript(program *Program, script Script) {
	if len(script.Sentences) == 0 {
		program.PushOpCode(OpCode_PUSH_NIL)
		return
	}
	for _, sentence := range script.Sentences {
		program.PushOpCode(OpCode_OPEN_FRAME)
		compiler.emitSentence(program, sentence)
		program.PushOpCode(OpCode_CLOSE_FRAME)
		program.PushOpCode(OpCode_EVALUATE_SENTENCE)
	}
	program.PushOpCode(OpCode_PUSH_RESULT)
}

//
// Sentences
//

// Flatten and compile the given sentences into a program
func (compiler Compiler) CompileSentences(sentences []Sentence) *Program {
	program := &Program{}
	program.PushOpCode(OpCode_OPEN_FRAME)
	for _, sentence := range sentences {
		compiler.emitSentence(program, sentence)
	}
	program.PushOpCode(OpCode_CLOSE_FRAME)
	return program
}

// Compile the given sentence into a program
func (compiler Compiler) CompileSentence(sentence Sentence) *Program {
	program := &Program{}
	compiler.emitSentence(program, sentence)
	return program
}
func (compiler Compiler) emitSentence(program *Program, sentence Sentence) {
	for _, word := range sentence.Words {
		// if (word instanceof Word) {
		compiler.emitWord(program, word)
		// } else {
		//   compiler.emitConstant(program, word);
		// }
	}
}

//
// Words
//

// Compile the given word into a program
func (compiler Compiler) CompileWord(word Word) *Program {
	program := &Program{}
	compiler.emitWord(program, word)
	return program
}

// Compile the given constant value into a program
func (compiler Compiler) CompileConstant(value Value) *Program {
	program := &Program{}
	compiler.emitConstant(program, value)
	return program
}
func (compiler Compiler) emitWord(program *Program, word Word) {
	switch compiler.syntaxChecker.CheckWord(word) {
	case WordType_ROOT:
		compiler.emitRoot(program, word.Morphemes[0])
	case WordType_COMPOUND:
		compiler.emitCompound(program, word.Morphemes)
	case WordType_SUBSTITUTION:
		compiler.emitSubstitution(program, word.Morphemes)
	case WordType_QUALIFIED:
		compiler.emitQualified(program, word.Morphemes)
	case WordType_IGNORED:
	case WordType_INVALID:
		// throw new InvalidWordStructureError("invalid word structure");
		panic("TODO")
	default:
		panic("CANTHAPPEN")
	}
}
func (compiler Compiler) emitRoot(program *Program, root Morpheme) {
	switch root.Type() {
	case MorphemeType_LITERAL:
		{
			literal := root.(LiteralMorpheme)
			compiler.emitLiteral(program, literal)
		}

	case MorphemeType_TUPLE:
		{
			tuple := root.(TupleMorpheme)
			compiler.emitTuple(program, tuple)
		}

	case MorphemeType_BLOCK:
		{
			block := root.(BlockMorpheme)
			compiler.emitBlock(program, block)
		}

	case MorphemeType_EXPRESSION:
		{
			expression := root.(ExpressionMorpheme)
			compiler.emitExpression(program, expression)
		}

	case MorphemeType_STRING:
		{
			string := root.(StringMorpheme)
			compiler.emitString(program, string)
		}

	case MorphemeType_HERE_STRING:
		{
			string := root.(HereStringMorpheme)
			compiler.emitHereString(program, string)
		}

	case MorphemeType_TAGGED_STRING:
		{
			string := root.(TaggedStringMorpheme)
			compiler.emitTaggedString(program, string)
		}

	default:
		//         throw new UnexpectedMorphemeError("unexpected morpheme");
		panic("TODO")
	}
}
func (compiler Compiler) emitCompound(program *Program, morphemes []Morpheme) {
	program.PushOpCode(OpCode_OPEN_FRAME)
	compiler.emitStems(program, morphemes)
	program.PushOpCode(OpCode_CLOSE_FRAME)
	program.PushOpCode(OpCode_JOIN_STRINGS)
}
func (compiler Compiler) emitSubstitution(program *Program, morphemes []Morpheme) {
	substitute := morphemes[0].(SubstituteNextMorpheme)
	selectable := morphemes[1]
	switch selectable.Type() {
	case MorphemeType_LITERAL:
		{
			literal := selectable.(LiteralMorpheme)
			compiler.emitLiteralVarname(program, literal)
		}

	case MorphemeType_TUPLE:
		{
			tuple := selectable.(TupleMorpheme)
			compiler.emitTupleVarnames(program, tuple)
		}

	case MorphemeType_BLOCK:
		{
			block := selectable.(BlockMorpheme)
			compiler.emitBlockVarname(program, block)
		}

	case MorphemeType_EXPRESSION:
		{
			expression := selectable.(ExpressionMorpheme)
			compiler.emitExpression(program, expression)
		}

	default:
		// throw new UnexpectedMorphemeError("unexpected morpheme");
		panic("TODO")
	}
	for i := 2; i < len(morphemes); i++ {
		morpheme := morphemes[i]
		switch morpheme.Type() {
		case MorphemeType_TUPLE:
			{
				tuple := morpheme.(TupleMorpheme)
				compiler.emitKeyedSelector(program, tuple)
			}

		case MorphemeType_BLOCK:
			{
				block := morpheme.(BlockMorpheme)
				compiler.emitSelector(program, block)
			}

		case MorphemeType_EXPRESSION:
			{
				expression := morpheme.(ExpressionMorpheme)
				compiler.emitIndexedSelector(program, expression)
			}

		default:
			// throw new UnexpectedMorphemeError("unexpected morpheme");
			panic("TODO")
		}
	}
	for level := 1; level < int(substitute.Levels); level++ {
		program.PushOpCode(OpCode_RESOLVE_VALUE)
	}
	if substitute.Expansion {
		program.PushOpCode(OpCode_EXPAND_VALUE)
	}
}
func (compiler Compiler) emitQualified(program *Program, morphemes []Morpheme) {
	selectable := morphemes[0]
	switch selectable.Type() {
	case MorphemeType_LITERAL:
		{
			literal := selectable.(LiteralMorpheme)
			compiler.emitLiteralSource(program, literal)
		}

	case MorphemeType_TUPLE:
		{
			tuple := selectable.(TupleMorpheme)
			compiler.emitTupleSource(program, tuple)
		}

	case MorphemeType_BLOCK:
		{
			block := selectable.(BlockMorpheme)
			compiler.emitBlockSource(program, block)
		}

	default:
		// throw new UnexpectedMorphemeError("unexpected morpheme");
		panic("TODO")
	}
	for i := 1; i < len(morphemes); i++ {
		morpheme := morphemes[i]
		switch morpheme.Type() {
		case MorphemeType_TUPLE:
			{
				tuple := morpheme.(TupleMorpheme)
				compiler.emitKeyedSelector(program, tuple)
			}

		case MorphemeType_BLOCK:
			{
				block := morpheme.(BlockMorpheme)
				compiler.emitSelector(program, block)
			}

		case MorphemeType_EXPRESSION:
			{
				expression := morpheme.(ExpressionMorpheme)
				compiler.emitIndexedSelector(program, expression)
			}

		default:
			// throw new UnexpectedMorphemeError("unexpected morpheme");
			panic("TODO")
		}
	}
}
func (compiler Compiler) emitStems(program *Program, morphemes []Morpheme) {
	const (
		INITIAL = iota
		SUBSTITUTE
		SELECTABLE
	)
	mode := INITIAL
	var substitute SubstituteNextMorpheme
	for _, morpheme := range morphemes {
		if mode == SELECTABLE {
			switch morpheme.Type() {
			case MorphemeType_SUBSTITUTE_NEXT,
				MorphemeType_LITERAL:
				{
					// Terminate substitution sequence
					for level := 1; level < int(substitute.Levels); level++ {
						program.PushOpCode(OpCode_RESOLVE_VALUE)
					}
					// substitute = nil
					mode = INITIAL
				}
			}
		}

		switch mode {
		case SUBSTITUTE:
			{
				// Expecting a source (varname or expression)
				switch morpheme.Type() {
				case MorphemeType_LITERAL:
					{
						literal := morpheme.(LiteralMorpheme)
						compiler.emitLiteralVarname(program, literal)
					}

				case MorphemeType_TUPLE:
					{
						tuple := morpheme.(TupleMorpheme)
						compiler.emitTupleVarnames(program, tuple)
					}

				case MorphemeType_BLOCK:
					{
						block := morpheme.(BlockMorpheme)
						compiler.emitBlockVarname(program, block)
					}

				case MorphemeType_EXPRESSION:
					{
						expression := morpheme.(ExpressionMorpheme)
						compiler.emitExpression(program, expression)
					}

				default:
					// throw new UnexpectedMorphemeError("unexpected morpheme");
					panic("TODO")
				}
				mode = SELECTABLE
			}

		case SELECTABLE:
			{
				// Expecting a selector
				switch morpheme.Type() {
				case MorphemeType_TUPLE:
					{
						tuple := morpheme.(TupleMorpheme)
						compiler.emitKeyedSelector(program, tuple)
					}

				case MorphemeType_BLOCK:
					{
						block := morpheme.(BlockMorpheme)
						compiler.emitSelector(program, block)
					}

				case MorphemeType_EXPRESSION:
					{
						expression := morpheme.(ExpressionMorpheme)
						compiler.emitIndexedSelector(program, expression)
					}

				default:
					// throw new UnexpectedMorphemeError("unexpected morpheme");
					panic("TODO")
				}
			}

		default:
			{
				switch morpheme.Type() {
				case MorphemeType_SUBSTITUTE_NEXT:
					// Start substitution sequence
					substitute = morpheme.(SubstituteNextMorpheme)
					mode = SUBSTITUTE

				case MorphemeType_LITERAL:
					{
						literal := morpheme.(LiteralMorpheme)
						compiler.emitLiteral(program, literal)
					}

				case MorphemeType_EXPRESSION:
					{
						expression := morpheme.(ExpressionMorpheme)
						compiler.emitExpression(program, expression)
					}

				default:
					// throw new UnexpectedMorphemeError("unexpected morpheme");
					panic("TODO")
				}
			}
		}
	}

	if mode == SELECTABLE {
		// Terminate substitution sequence
		for level := 1; level < int(substitute.Levels); level++ {
			program.PushOpCode(OpCode_RESOLVE_VALUE)
		}
	}
}

//
// Morphemes
//

func (compiler Compiler) emitLiteral(program *Program, literal LiteralMorpheme) {
	value := NewStringValue(literal.Value)
	compiler.emitConstant(program, value)
}
func (compiler Compiler) emitTuple(program *Program, tuple TupleMorpheme) {
	program.PushOpCode(OpCode_OPEN_FRAME)
	for _, sentence := range tuple.Subscript.Sentences {
		compiler.emitSentence(program, sentence)
	}
	program.PushOpCode(OpCode_CLOSE_FRAME)
}
func (compiler Compiler) emitBlock(program *Program, block BlockMorpheme) {
	value := NewScriptValue(block.Subscript, block.Value)
	compiler.emitConstant(program, value)
}
func (compiler Compiler) emitExpression(program *Program, expression ExpressionMorpheme) {
	compiler.emitScript(program, expression.Subscript)
}
func (compiler Compiler) emitString(program *Program, string StringMorpheme) {
	program.PushOpCode(OpCode_OPEN_FRAME)
	compiler.emitStems(program, string.Morphemes)
	program.PushOpCode(OpCode_CLOSE_FRAME)
	program.PushOpCode(OpCode_JOIN_STRINGS)
}
func (compiler Compiler) emitHereString(program *Program, string HereStringMorpheme) {
	value := NewStringValue(string.Value)
	compiler.emitConstant(program, value)
}
func (compiler Compiler) emitTaggedString(program *Program, string TaggedStringMorpheme) {
	value := NewStringValue(string.Value)
	compiler.emitConstant(program, value)
}
func (compiler Compiler) emitLiteralVarname(program *Program, literal LiteralMorpheme) {
	compiler.emitLiteral(program, literal)
	program.PushOpCode(OpCode_RESOLVE_VALUE)
}
func (compiler Compiler) emitTupleVarnames(program *Program, tuple TupleMorpheme) {
	compiler.emitTuple(program, tuple)
	program.PushOpCode(OpCode_RESOLVE_VALUE)
}
func (compiler Compiler) emitBlockVarname(program *Program, block BlockMorpheme) {
	value := NewStringValue(block.Value)
	compiler.emitConstant(program, value)
	program.PushOpCode(OpCode_RESOLVE_VALUE)
}
func (compiler Compiler) emitLiteralSource(program *Program, literal LiteralMorpheme) {
	compiler.emitLiteral(program, literal)
	program.PushOpCode(OpCode_SET_SOURCE)
}
func (compiler Compiler) emitTupleSource(program *Program, tuple TupleMorpheme) {
	compiler.emitTuple(program, tuple)
	program.PushOpCode(OpCode_SET_SOURCE)
}
func (compiler Compiler) emitBlockSource(program *Program, block BlockMorpheme) {
	value := NewStringValue(block.Value)
	compiler.emitConstant(program, value)
	program.PushOpCode(OpCode_SET_SOURCE)
}
func (compiler Compiler) emitKeyedSelector(program *Program, tuple TupleMorpheme) {
	compiler.emitTuple(program, tuple)
	program.PushOpCode(OpCode_SELECT_KEYS)
}
func (compiler Compiler) emitIndexedSelector(program *Program, expression ExpressionMorpheme) {
	compiler.emitScript(program, expression.Subscript)
	program.PushOpCode(OpCode_SELECT_INDEX)
}
func (compiler Compiler) emitSelector(program *Program, block BlockMorpheme) {
	program.PushOpCode(OpCode_OPEN_FRAME)
	for _, sentence := range block.Subscript.Sentences {
		program.PushOpCode(OpCode_OPEN_FRAME)
		compiler.emitSentence(program, sentence)
		program.PushOpCode(OpCode_CLOSE_FRAME)
	}
	program.PushOpCode(OpCode_CLOSE_FRAME)
	program.PushOpCode(OpCode_SELECT_RULES)
}
func (compiler Compiler) emitConstant(program *Program, value Value) {
	program.PushOpCode(OpCode_PUSH_CONSTANT)
	program.PushConstant(value)
}

// /**
//  * Helena program state
//  *
//  * This class encapsulates the state of a program being executed, allowing
//  * reentrancy and parallelism of executors
//  */
// export class ProgramState {
//   /** Execution frames; each frame is a stack of values */
//   private readonly frames: Value[][] = [[]];

//   /** Program counter */
//   pc = 0;

//   /** Constant counter */
//   cc = 0;

//   /** Last executed command */
//   command: Command;

//   /** Last executed result value */
//   result: Result = OK(NIL);

//   /** Open a new frame */
//   openFrame() {
//     this.frames.push([]);
//   }

//   /**
//    * Close the current frame
//    *
//    * @returns The closed frame
//    */
//   closeFrame() {
//     return this.frames.pop();
//   }

//   /** @returns Current frame */
//   frame() {
//     return this.frames[this.frames.length - 1];
//   }

//   /**
//    * Push value on current frame
//    *
//    * @param value - Value to push
//    */
//   push(value: Value) {
//     this.frame().push(value);
//   }

//   /**
//    * Pop last value on current frame
//    *
//    * @returns Popped value
//    */
//   pop() {
//     return this.frame().pop();
//   }

//   /** @returns Last value on current frame */
//   private last() {
//     return this.frame()[this.frame().length - 1];
//   }

//   /** Expand last value in current frame */
//   expand() {
//     const last = this.last();
//     if (last && last.type == ValueType.TUPLE) {
//       this.frame().pop();
//       this.frame().push(...(last as TupleValue).values);
//     }
//   }
// }

// /**
//  * Helena program executor
//  *
//  * This class executes compiled programs in an isolated state
//  */
// export class Executor {
//   /** Variable resolver used during execution */
//   private readonly variableResolver: VariableResolver;

//   /** Command resolver used during execution */
//   private readonly commandResolver: CommandResolver;

//   /** Selector resolver used during execution */
//   private readonly selectorResolver: SelectorResolver;

//   /** Opaque context passed to commands */
//   private readonly context: unknown;

//   /**
//    * @param variableResolver - Variable resolver
//    * @param commandResolver  - Command resolver
//    * @param selectorResolver - Selector resolver
//    * @param [context]        - Opaque context
//    */
//   constructor(
//     variableResolver: VariableResolver,
//     commandResolver: CommandResolver,
//     selectorResolver: SelectorResolver,
//     context?: unknown
//   ) {
//     this.variableResolver = variableResolver;
//     this.commandResolver = commandResolver;
//     this.selectorResolver = selectorResolver;
//     this.context = context;
//   }

//   /**
//    * Execute the given program
//    *
//    * Runs a flat loop over the program opcodes
//    *
//    * By default a new state is created at each call. Passing a state object
//    * can be used to implement resumability, context switching, trampolines,
//    * coroutines, etc.
//    *
//    * @param program - Program to execute
//    * @param [state] - Program state (defaults to new)
//    *
//    * @returns         Last executed result
//    */
//   execute(program *Program, state = new ProgramState()): Result {
//     if (state.result.code == ResultCode.YIELD && state.command?.resume) {
//       state.result = state.command.resume(state.result, this.context);
//       if (state.result.code != ResultCode.OK) return state.result;
//     }
//     while (state.pc < program.opCodes.length) {
//       const opcode = program.opCodes[state.pc++];
//       switch (opcode) {
//         case OpCode_PUSH_NIL:
//           state.push(NIL);
//           break;

//         case OpCode_PUSH_CONSTANT:
//           state.push(program.constants[state.cc++]);
//           break;

//         case OpCode_OPEN_FRAME:
//           state.openFrame();
//           break;

//         case OpCode_CLOSE_FRAME:
//           {
//             const values = state.closeFrame();
//             state.push(new TupleValue(values));
//           }
//           break;

//         case OpCode_RESOLVE_VALUE:
//           {
//             const source = state.pop();
//             const result = this.resolveValue(source);
//             if (result.code != ResultCode.OK) return result;
//             state.push(result.value);
//           }
//           break;

//         case OpCode_EXPAND_VALUE:
//           state.expand();
//           break;

//         case OpCode_SET_SOURCE:
//           {
//             const source = state.pop();
//             state.push(new QualifiedValue(source, []));
//           }
//           break;

//         case OpCode_SELECT_INDEX:
//           {
//             const index = state.pop();
//             const value = state.pop();
//             const { data: selector, ...result2 } =
//               IndexedSelector.create(index);
//             if (result2.code != ResultCode.OK) return result2;
//             const result = selector.apply(value);
//             if (result.code != ResultCode.OK) return result;
//             state.push(result.value);
//           }
//           break;

//         case OpCode_SELECT_KEYS:
//           {
//             const keys = state.pop() as TupleValue;
//             const value = state.pop();
//             const { data: selector, ...result2 } = KeyedSelector.create(
//               keys.values
//             );
//             if (result2.code != ResultCode.OK) return result2;
//             const result = selector.apply(value);
//             if (result.code != ResultCode.OK) return result;
//             state.push(result.value);
//           }
//           break;

//         case OpCode_SELECT_RULES:
//           {
//             const rules = state.pop() as TupleValue;
//             const value = state.pop();
//             const { data: selector, ...result } = this.resolveSelector(
//               rules.values
//             );
//             if (result.code != ResultCode.OK) return result;
//             const result2 = applySelector(value, selector);
//             if (result2.code != ResultCode.OK) return result2;
//             state.push(result2.value);
//           }
//           break;

//         case OpCode_EVALUATE_SENTENCE:
//           {
//             const args = state.pop() as TupleValue;
//             if (args.values.length) {
//               const cmdname = args.values[0];
//               const { data: command, ...result } = this.resolveCommand(cmdname);
//               if (result.code != ResultCode.OK) return result;
//               state.command = command;
//               state.result = state.command.execute(args.values, this.context);
//               if (state.result.code != ResultCode.OK) return state.result;
//             }
//           }
//           break;

//         case OpCode_PUSH_RESULT:
//           state.push(state.result.value);
//           break;

//         case OpCode_JOIN_STRINGS:
//           {
//             const tuple = state.pop() as TupleValue;
//             let s = "";
//             for (const value of tuple.values) {
//               const { data, ...result } = StringValue.toString(value);
//               if (result.code != ResultCode.OK) return result;
//               s += data;
//             }
//             state.push(new StringValue(s));
//           }
//           break;

//         default:
//           throw new Error("CANTHAPPEN");
//       }
//     }
//     if (state.frame().length) state.result = OK(state.pop());
//     return state.result;
//   }

//   /**
//    * Transform the given program into a callable function
//    *
//    * The program is first translated into JS code then wrapped into a function
//    * with all the dependencies injected as parameters. The resulting function
//    * is itself curried with the current executor context.
//    *
//    * @param program - Program to translate
//    * @returns         Resulting function
//    */
//   functionify(program *Program): (state?: ProgramState) => Result {
//     const translator = new Translator();
//     const source = translator.translate(program);
//     const imports = {
//       ResultCode,
//       OK,
//       ERROR,
//       NIL,
//       StringValue,
//       TupleValue,
//       QualifiedValue,
//       IndexedSelector,
//       KeyedSelector,
//       applySelector,
//     };
//     const importsCode = `
//     const {
//       ResultCode,
//       OK,
//       ERROR,
//       NIL,
//       StringValue,
//       TupleValue,
//       QualifiedValue,
//       IndexedSelector,
//       KeyedSelector,
//       applySelector,
//     } = imports;
//     `;

//     const f = new Function(
//       "state",
//       "resolver",
//       "context",
//       "constants",
//       "imports",
//       importsCode + source
//     );
//     return (state = new ProgramState()) =>
//       f(state, this, this.context, program.constants, imports);
//   }

//   /**
//    * Resolve value
//    *
//    * - If source value is a tuple, resolve each of its elements recursively
//    * - If source value is a qualified word, resolve source and apply selectors
//    * - Else, resolve variable from the source string value
//    *
//    * @param source - Value(s) to resolve
//    * @returns        Resolved value(s)
//    */
//   private resolveValue(source: Value): Result {
//     switch (source.type) {
//       case ValueType.TUPLE:
//         return this.resolveTuple(source as TupleValue);
//       case ValueType.QUALIFIED:
//         return this.resolveQualified(source as QualifiedValue);
//       default: {
//         const { data: varname, code } = StringValue.toString(source);
//         if (code != ResultCode.OK) return ERROR("invalid variable name");
//         return this.resolveVariable(varname);
//       }
//     }
//   }
//   private resolveQualified(qualified: QualifiedValue): Result {
//     let result = this.resolveValue(qualified.source);
//     if (result.code != ResultCode.OK) return result;
//     for (const selector of qualified.selectors) {
//       result = selector.apply(result.value);
//       if (result.code != ResultCode.OK) return result;
//     }
//     return result;
//   }

//   /**
//    * Resolve tuple values recursively
//    *
//    * @param tuple - Tuple to resolve
//    *
//    * @returns       Resolved tuple
//    */
//   private resolveTuple(tuple: TupleValue): Result {
//     const values: Value[] = [];
//     for (const value of tuple.values) {
//       let result: Result;
//       switch (value.type) {
//         case ValueType.TUPLE:
//           result = this.resolveTuple(value as TupleValue);
//           break;
//         default:
//           result = this.resolveValue(value);
//       }
//       if (result.code != ResultCode.OK) return result;
//       values.push(result.value);
//     }
//     return OK(new TupleValue(values));
//   }

//   private resolveVariable(varname: string): Result {
//     const value = this.variableResolver.resolve(varname);
//     if (!value) return ERROR(`cannot resolve variable "${varname}"`);
//     return OK(value);
//   }
//   private resolveCommand(cmdname: Value): Result<Command> {
//     const command = this.commandResolver.resolve(cmdname);
//     if (!command) {
//       const { data: name, code } = StringValue.toString(cmdname);
//       if (code != ResultCode.OK) return ERROR("invalid command name");
//       return ERROR(`cannot resolve command "${name}"`);
//     }
//     return OK(NIL, command);
//   }
//   private resolveSelector(rules: Value[]): Result<Selector> {
//     const result = this.selectorResolver.resolve(rules);
//     if (result.code != ResultCode.OK) return result;
//     if (!result.data)
//       return ERROR(`cannot resolve selector {${displayList(rules)}}`);
//     return result;
//   }
// }

// /**
//  * Helena program translator
//  *
//  * This class translates compiled programs into JavaScript code
//  */
// export class Translator {
//   /**
//    * Translate the given program
//    *
//    * Runs a flat loop over the program opcodes and generates JS code of each
//    * opcode in sequence; constants are inlined in the order they are encountered
//    *
//    * Resumability is implemented using `switch` as a jump table (similar to
//    * `Duff's device technique):
//    *
//    * - translated opcodes are wrapped into a `switch` statement whose control
//    * variable is the current {@link ProgramState.pc}
//    * - each opcode is behind a case statement whose value is the opcode position
//    * - case statements fall through (there is no break statement)
//    *
//    * This allows a resumed program to jump directly to the current
//    * {@link ProgramState.pc} and continue execution from there until the next
//    * `return`.
//    *
//    * @see Executor.execute(): The generated code must be kept in sync with the
//    * execution loop
//    *
//    * @param program - Program to execute
//    *
//    * @returns         Translated code
//    */
//   translate(program *Program) {
//     const sections: string[] = [];

//     sections.push(`
//     if (state.result.code == ResultCode.YIELD && state.command?.resume) {
//       state.result = state.command.resume(state.result, context);
//       if (state.result.code != ResultCode.OK) return state.result;
//     }
//     `);

//     sections.push(`
//     switch (state.pc) {
//     `);
//     let pc = 0;
//     let cc = 0;
//     while (pc < program.opCodes.length) {
//       sections.push(`
//       case ${pc}: state.pc++;
//       `);
//       const opcode = program.opCodes[pc++];
//       switch (opcode) {
//         case OpCode_PUSH_NIL:
//           sections.push(`
//           state.push(NIL);
//           `);
//           break;

//         case OpCode_PUSH_CONSTANT:
//           sections.push(`
//           state.push(constants[${cc++}]);
//           `);
//           break;

//         case OpCode_OPEN_FRAME:
//           sections.push(`
//           state.openFrame();
//           `);
//           break;

//         case OpCode_CLOSE_FRAME:
//           sections.push(`
//           {
//             const values = state.closeFrame();
//             state.push(new TupleValue(values));
//           }
//           `);
//           break;

//         case OpCode_RESOLVE_VALUE:
//           sections.push(`
//           {
//             const source = state.pop();
//             const result = resolver.resolveValue(source);
//             if (result.code != ResultCode.OK) return result;
//             state.push(result.value);
//           }
//           `);
//           break;

//         case OpCode_EXPAND_VALUE:
//           sections.push(`
//           state.expand();
//           `);
//           break;

//         case OpCode_SET_SOURCE:
//           sections.push(`
//           {
//             const source = state.pop();
//             state.push(new QualifiedValue(source, []));
//           }
//           `);
//           break;

//         case OpCode_SELECT_INDEX:
//           sections.push(`
//           {
//             const index = state.pop();
//             const value = state.pop();
//             const { data: selector, ...result2 } =
//               IndexedSelector.create(index);
//             const result = selector.apply(value);
//             if (result.code != ResultCode.OK) return result;
//             state.push(result.value);
//           }
//           `);
//           break;

//         case OpCode_SELECT_KEYS:
//           sections.push(`
//           {
//             const keys = state.pop();
//             const value = state.pop();
//             const { data: selector, ...result2 } = KeyedSelector.create(
//               keys.values
//             );
//             if (result2.code != ResultCode.OK) return result2;
//             const result = selector.apply(value);
//             if (result.code != ResultCode.OK) return result;
//             state.push(result.value);
//           }
//           `);
//           break;

//         case OpCode_SELECT_RULES:
//           sections.push(`
//           {
//             const rules = state.pop();
//             const value = state.pop();
//             const { data: selector, ...result } = resolver.resolveSelector(
//               rules.values
//             );
//             if (result.code != ResultCode.OK) return result;
//             const result2 = applySelector(value, selector);
//             if (result2.code != ResultCode.OK) return result2;
//             state.push(result2.value);
//           }
//           `);
//           break;

//         case OpCode_EVALUATE_SENTENCE:
//           sections.push(`
//           {
//             const args = state.pop();
//             if (args.values.length) {
//               const cmdname = args.values[0];
//               const { data: command, ...result } = resolver.resolveCommand(cmdname);
//               if (result.code != ResultCode.OK) return result;
//               state.command = command;
//               state.result = state.command.execute(args.values, context);
//               if (state.result.code != ResultCode.OK) return state.result;
//             }
//           }
//           `);
//           break;

//         case OpCode_PUSH_RESULT:
//           sections.push(`
//           state.push(state.result.value);
//           `);
//           break;

//         case OpCode_JOIN_STRINGS:
//           sections.push(`
//           {
//             const tuple = state.pop();
//             let s = "";
//             for (const value of tuple.values) {
//               const { data, ...result } = StringValue.toString(value);
//               if (result.code != ResultCode.OK) return result;
//               s += data;
//             }
//             state.push(new StringValue(s));
//           }
//           `);
//           break;

//         default:
//           throw new Error("CANTHAPPEN");
//       }
//     }
//     sections.push(`
//     }
//     if (state.frame().length) state.result = OK(state.pop());
//     return state.result;
//     `);
//     return sections.join("\n");
//   }
// }
