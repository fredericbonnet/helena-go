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
	OpCode_MAKE_TUPLE
)

//
// Helena program
//
type Program struct {
	// Sequence of opcodes the program is made of
	OpCodes []OpCode

	// Constants the opcodes refer to
	Constants []Value

	// Program source
	Source *Source

	// Opcode positions
	OpCodePositions []*SourcePosition
}

func NewProgram(capturePositions bool, source *Source) *Program {
	if capturePositions {
		return &Program{OpCodePositions: []*SourcePosition{}, Source: source}
	}
	return &Program{}
}

// Push a new opcode
func (program *Program) PushOpCode(opCode OpCode, position *SourcePosition) {
	program.OpCodes = append(program.OpCodes, opCode)
	if program.OpCodePositions != nil {
		program.OpCodePositions = append(program.OpCodePositions, position)
	}
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
// Compiler options
//
type CompilerOptions struct {
	// Whether to capture opcode and constant positions
	CapturePositions bool
}

//
// Helena compiler
//
// This class transforms scripts, sentences and words into programs
//
type Compiler struct {
	// Syntax checker used during compilation
	syntaxChecker SyntaxChecker

	// Compiler options
	options CompilerOptions
}

func NewCompiler(options *CompilerOptions) Compiler {
	if options == nil {
		return Compiler{options: CompilerOptions{false}}
	} else {
		return Compiler{options: *options}
	}
}

//
// Scripts
//

// Compile the given script into a program
func (compiler Compiler) CompileScript(script Script) *Program {
	program := NewProgram(
		compiler.options.CapturePositions,
		script.Source,
	)
	if len(script.Sentences) == 0 {
		return program
	}
	compiler.emitScript(program, script)
	return program
}
func (compiler Compiler) emitScript(program *Program, script Script) {
	if len(script.Sentences) == 0 {
		program.PushOpCode(OpCode_PUSH_NIL, script.Position)
		return
	}
	for _, sentence := range script.Sentences {
		program.PushOpCode(OpCode_OPEN_FRAME, sentence.Position)
		compiler.emitSentence(program, sentence)
		program.PushOpCode(OpCode_CLOSE_FRAME, sentence.Position)
		program.PushOpCode(OpCode_EVALUATE_SENTENCE, sentence.Position)
	}
	program.PushOpCode(OpCode_PUSH_RESULT, script.Position)
}

//
// Sentences
//

// Flatten and compile the given sentences into a program
func (compiler Compiler) CompileSentences(sentences []Sentence) *Program {
	program := NewProgram(compiler.options.CapturePositions, nil)
	compiler.emitSentences(program, sentences, nil)
	program.PushOpCode(OpCode_MAKE_TUPLE, nil)
	return program
}
func (compiler Compiler) emitSentences(
	program *Program,
	sentences []Sentence,
	position *SourcePosition,
) {
	program.PushOpCode(OpCode_OPEN_FRAME, position)
	for _, sentence := range sentences {
		compiler.emitSentence(program, sentence)
	}
	program.PushOpCode(OpCode_CLOSE_FRAME, position)
}

// Compile the given sentence into a program
func (compiler Compiler) CompileSentence(sentence Sentence) *Program {
	program := NewProgram(compiler.options.CapturePositions, nil)
	compiler.emitSentence(program, sentence)
	return program
}
func (compiler Compiler) emitSentence(program *Program, sentence Sentence) {
	for _, word := range sentence.Words {
		if word.Value == nil {
			compiler.emitWord(program, word.Word)
		} else {
			compiler.emitConstant(program, word.Value, nil)
		}
	}
}

//
// Words
//

// Compile the given word into a program
func (compiler Compiler) CompileWord(word Word) *Program {
	program := NewProgram(compiler.options.CapturePositions, nil)
	compiler.emitWord(program, word)
	return program
}

// Compile the given constant value into a program
func (compiler Compiler) CompileConstant(value Value) *Program {
	program := NewProgram(compiler.options.CapturePositions, nil)
	compiler.emitConstant(program, value, nil)
	return program
}
func (compiler Compiler) emitWord(program *Program, word Word) {
	switch compiler.syntaxChecker.CheckWord(word) {
	case WordType_ROOT:
		compiler.emitRoot(program, word)
	case WordType_COMPOUND:
		compiler.emitCompound(program, word)
	case WordType_SUBSTITUTION:
		compiler.emitSubstitution(program, word)
	case WordType_QUALIFIED:
		compiler.emitQualified(program, word)
	case WordType_IGNORED:
	case WordType_INVALID:
		panic(InvalidWordStructureError)
	default:
		panic("CANTHAPPEN")
	}
}
func (compiler Compiler) emitRoot(program *Program, word Word) {
	root := word.Morphemes[0]
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
		panic(UnexpectedMorphemeError)
	}
}
func (compiler Compiler) emitCompound(program *Program, word Word) {
	program.PushOpCode(OpCode_OPEN_FRAME, word.Position)
	compiler.emitStems(program, word.Morphemes)
	program.PushOpCode(OpCode_CLOSE_FRAME, word.Position)
	program.PushOpCode(OpCode_JOIN_STRINGS, word.Position)
}
func (compiler Compiler) emitSubstitution(program *Program, word Word) {
	firstSubstitute := 0
	expand := (word.Morphemes[0].(SubstituteNextMorpheme)).Expansion
	levels := 1
	i := 1
	for word.Morphemes[i].Type() == MorphemeType_SUBSTITUTE_NEXT {
		i++
		levels++
	}
	selectable := word.Morphemes[i]
	i++
	switch selectable.Type() {
	case MorphemeType_LITERAL:
		{
			literal := selectable.(LiteralMorpheme)
			substitute := word.Morphemes[firstSubstitute+levels-1]
			compiler.emitLiteral(program, literal)
			program.PushOpCode(OpCode_RESOLVE_VALUE, substitute.Position())
		}

	case MorphemeType_TUPLE:
		{
			tuple := selectable.(TupleMorpheme)
			substitute := word.Morphemes[firstSubstitute+levels-1]
			compiler.emitTuple(program, tuple)
			program.PushOpCode(OpCode_RESOLVE_VALUE, substitute.Position())
		}

	case MorphemeType_BLOCK:
		{
			block := selectable.(BlockMorpheme)
			substitute := word.Morphemes[firstSubstitute+levels-1]
			compiler.emitBlockString(program, block)
			program.PushOpCode(OpCode_RESOLVE_VALUE, substitute.Position())
		}

	case MorphemeType_EXPRESSION:
		{
			expression := selectable.(ExpressionMorpheme)
			compiler.emitExpression(program, expression)
		}

	default:
		panic(UnexpectedMorphemeError)
	}
	for i < len(word.Morphemes) {
		morpheme := word.Morphemes[i]
		i++
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
			panic(UnexpectedMorphemeError)
		}
	}
	compiler.terminateSubstitutionSequence(
		program,
		word.Morphemes,
		firstSubstitute,
		levels,
	)
	if expand {
		program.PushOpCode(OpCode_EXPAND_VALUE, word.Position)
	}
}
func (compiler Compiler) emitQualified(program *Program, word Word) {
	selectable := word.Morphemes[0]
	switch selectable.Type() {
	case MorphemeType_LITERAL:
		{
			literal := selectable.(LiteralMorpheme)
			compiler.emitLiteral(program, literal)
		}

	case MorphemeType_TUPLE:
		{
			tuple := selectable.(TupleMorpheme)
			compiler.emitTuple(program, tuple)
		}

	case MorphemeType_BLOCK:
		{
			block := selectable.(BlockMorpheme)
			compiler.emitBlockString(program, block)
		}

	default:
		panic(UnexpectedMorphemeError)
	}
	program.PushOpCode(OpCode_SET_SOURCE, selectable.Position())
	for i := 1; i < len(word.Morphemes); i++ {
		morpheme := word.Morphemes[i]
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
			panic(UnexpectedMorphemeError)
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
	var firstSubstitute int
	var levels int
	for i, morpheme := range morphemes {
		if mode == SELECTABLE {
			switch morpheme.Type() {
			case MorphemeType_SUBSTITUTE_NEXT,
				MorphemeType_LITERAL:
				compiler.terminateSubstitutionSequence(
					program,
					morphemes,
					firstSubstitute,
					levels,
				)
				mode = INITIAL
			}
		}

		switch mode {
		case SUBSTITUTE:
			{
				switch morpheme.Type() {
				case MorphemeType_SUBSTITUTE_NEXT:
					// Continue substitution sequence
					levels++
					continue

					// Expecting a source (varname or expression)
				case MorphemeType_LITERAL:
					{
						literal := morpheme.(LiteralMorpheme)
						substitute := morphemes[firstSubstitute+levels-1]
						compiler.emitLiteral(program, literal)
						program.PushOpCode(OpCode_RESOLVE_VALUE, substitute.Position())
					}

				case MorphemeType_TUPLE:
					{
						tuple := morpheme.(TupleMorpheme)
						substitute := morphemes[firstSubstitute+levels-1]
						compiler.emitTuple(program, tuple)
						program.PushOpCode(OpCode_RESOLVE_VALUE, substitute.Position())
					}

				case MorphemeType_BLOCK:
					{
						block := morpheme.(BlockMorpheme)
						substitute := morphemes[firstSubstitute+levels-1]
						compiler.emitBlockString(program, block)
						program.PushOpCode(OpCode_RESOLVE_VALUE, substitute.Position())
					}

				case MorphemeType_EXPRESSION:
					{
						expression := morpheme.(ExpressionMorpheme)
						compiler.emitExpression(program, expression)
					}

				default:
					panic(UnexpectedMorphemeError)
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
					panic(UnexpectedMorphemeError)
				}
			}

		default:
			{
				switch morpheme.Type() {
				case MorphemeType_SUBSTITUTE_NEXT:
					// Start substitution sequence
					mode = SUBSTITUTE
					firstSubstitute = i
					levels = 1

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
					panic(UnexpectedMorphemeError)
				}
			}
		}
	}
	if mode == SELECTABLE {
		compiler.terminateSubstitutionSequence(
			program,
			morphemes,
			firstSubstitute,
			levels,
		)
		mode = INITIAL
	}
}
func (Compiler) terminateSubstitutionSequence(
	program *Program,
	morphemes []Morpheme,
	first int,
	levels int,
) {
	for level := levels - 1; level >= 1; level-- {
		substitute := morphemes[first+level-1]
		program.PushOpCode(OpCode_RESOLVE_VALUE, substitute.Position())
	}
}

//
// Morphemes
//

func (compiler Compiler) emitLiteral(program *Program, literal LiteralMorpheme) {
	value := NewStringValue(literal.Value)
	compiler.emitConstant(program, value, literal.position)
}
func (compiler Compiler) emitTuple(program *Program, tuple TupleMorpheme) {
	compiler.emitSentences(program, tuple.Subscript.Sentences, tuple.position)
	program.PushOpCode(OpCode_MAKE_TUPLE, tuple.position)
}
func (compiler Compiler) emitBlock(program *Program, block BlockMorpheme) {
	value := NewScriptValue(block.Subscript, block.Value)
	compiler.emitConstant(program, value, block.position)
}
func (compiler Compiler) emitBlockString(program *Program, block BlockMorpheme) {
	value := NewStringValue(block.Value)
	compiler.emitConstant(program, value, block.position)
}
func (compiler Compiler) emitExpression(program *Program, expression ExpressionMorpheme) {
	compiler.emitScript(program, expression.Subscript)
}
func (compiler Compiler) emitString(program *Program, string StringMorpheme) {
	program.PushOpCode(OpCode_OPEN_FRAME, string.position)
	compiler.emitStems(program, string.Morphemes)
	program.PushOpCode(OpCode_CLOSE_FRAME, string.position)
	program.PushOpCode(OpCode_JOIN_STRINGS, string.position)
}
func (compiler Compiler) emitHereString(program *Program, string HereStringMorpheme) {
	value := NewStringValue(string.Value)
	compiler.emitConstant(program, value, string.position)
}
func (compiler Compiler) emitTaggedString(program *Program, string TaggedStringMorpheme) {
	value := NewStringValue(string.Value)
	compiler.emitConstant(program, value, string.position)
}
func (compiler Compiler) emitKeyedSelector(program *Program, tuple TupleMorpheme) {
	compiler.emitSentences(program, tuple.Subscript.Sentences, tuple.position)
	program.PushOpCode(OpCode_SELECT_KEYS, tuple.position)
}
func (compiler Compiler) emitIndexedSelector(program *Program, expression ExpressionMorpheme) {
	compiler.emitScript(program, expression.Subscript)
	program.PushOpCode(OpCode_SELECT_INDEX, expression.position)
}
func (compiler Compiler) emitSelector(program *Program, block BlockMorpheme) {
	program.PushOpCode(OpCode_OPEN_FRAME, block.position)
	for _, sentence := range block.Subscript.Sentences {
		program.PushOpCode(OpCode_OPEN_FRAME, sentence.Position)
		compiler.emitSentence(program, sentence)
		program.PushOpCode(OpCode_CLOSE_FRAME, sentence.Position)
		program.PushOpCode(OpCode_MAKE_TUPLE, sentence.Position)
	}
	program.PushOpCode(OpCode_CLOSE_FRAME, block.position)
	program.PushOpCode(OpCode_SELECT_RULES, block.position)
}
func (compiler Compiler) emitConstant(
	program *Program,
	value Value,
	position *SourcePosition,
) {
	program.PushOpCode(OpCode_PUSH_CONSTANT, position)
	program.PushConstant(value)
}

//
// Helena program state
//
// This class encapsulates the state of a program being executed, allowing
// reentrancy and parallelism of executors
//
type ProgramState struct {
	// Execution stack
	stack []Value

	// Execution frame start indexes; each frame is a slice of the stack
	frames []int

	// Execution results for each frame
	frameResults []Result

	// Last closed frame
	LastFrame []Value

	// Program counter
	PC uint

	// Constant counter
	CC uint

	// Last executed command
	Command Command

	// Last execution result
	Result Result
}

func NewProgramState() *ProgramState {
	return &ProgramState{
		stack:        make([]Value, 0, 64),
		frames:       make([]int, 1, 16),
		frameResults: append(make([]Result, 0, 16), OK(NIL)),
		PC:           0,
		CC:           0,
		Result:       OK(NIL),
	}
}

// Reset program state
func (state *ProgramState) Reset() {
	state.stack = state.stack[:0]
	state.frames = state.frames[:1]
	state.frames[0] = 0
	state.frameResults = state.frameResults[:1]
	state.frameResults[0] = OK(NIL)
	state.LastFrame = nil
	state.PC = 0
	state.CC = 0
	state.Command = nil
	state.Result = OK(NIL)
}

// Set result for the current frame
func (state *ProgramState) SetResult(result Result) {
	state.Result = result
	state.frameResults[len(state.frameResults)-1] = result
}

// Return result of the last frame
func (state *ProgramState) lastFrameResult() Result {
	return state.frameResults[len(state.frameResults)-1]
}

// Open a new frame
func (state *ProgramState) OpenFrame() {
	state.frames = append(state.frames, len(state.stack))
	state.frameResults = append(state.frameResults, OK(NIL))
	state.LastFrame = nil
}

// Close the current frame
func (state *ProgramState) CloseFrame() {
	length := state.frames[len(state.frames)-1]
	state.frames = state.frames[:len(state.frames)-1]
	state.frameResults = state.frameResults[:len(state.frameResults)-1]
	state.LastFrame = state.stack[length:]
	state.stack = state.stack[:length]
}

// Report whether current frame is empty
func (state *ProgramState) Empty() bool {
	return len(state.frames) == 0 || state.frames[len(state.frames)-1] == len(state.stack)
}

// Push value on current frame
func (state *ProgramState) Push(value Value) {
	state.stack = append(state.stack, value)
}

// Pop and return last value on current frame
func (state *ProgramState) Pop() Value {
	last := state.stack[len(state.stack)-1]
	state.stack = state.stack[:len(state.stack)-1]
	return last
}

// Return last value on current frame
func (state *ProgramState) lastValue() Value {
	if len(state.stack) == 0 {
		return nil
	}
	return state.stack[len(state.stack)-1]
}

// Expand last value in current frame
func (state *ProgramState) Expand() {
	last := state.lastValue()
	if last != nil && last.Type() == ValueType_TUPLE {
		state.stack = state.stack[:len(state.stack)-1]
		state.stack = append(state.stack, last.(TupleValue).Values...)
	}
}

//
// Helena program executor
//
// This class executes compiled programs in an isolated state
//
type Executor struct {
	// Variable resolver used during execution
	VariableResolver VariableResolver

	// Command resolver used during execution
	CommandResolver CommandResolver

	// Selector resolver used during execution
	SelectorResolver SelectorResolver

	// Opaque context passed to commands
	Context any
}

// Execute the given program and return last executed result
//
// Runs a flat loop over the program opcodes
//
// By default a new state is created at each call. Passing a state object
// can be used to implement resumability, context switching, trampolines,
// coroutines, etc.
func (executor *Executor) Execute(program *Program, state *ProgramState) Result {
	if state == nil {
		state = NewProgramState()
	}
	result := executor.ExecuteUntil(program, state, uint(len(program.OpCodes)))
	if result.Code != ResultCode_OK {
		return result
	}
	if !state.Empty() {
		state.SetResult(OK(state.Pop()))
	}
	return state.Result
}

// Execute the given program until the provided stop point
//
// Runs a flat loop over the program opcodes
//
// Return OK(NIL) upon success, else last result
func (executor *Executor) ExecuteUntil(program *Program, state *ProgramState, stop uint) Result {
	if stop > uint(len(program.OpCodes)) {
		stop = uint(len(program.OpCodes))
	}
	if state.PC >= stop {
		return OK(NIL)
	}
	if state.Result.Code == ResultCode_YIELD {
		if resumable, ok := state.Command.(ResumableCommand); ok {
			result := resumable.Resume(state.Result, executor.Context)
			state.SetResult(result)
			if result.Code != ResultCode_OK {
				return result
			}
		}
	}
	for state.PC < stop {
		opcode := program.OpCodes[state.PC]
		state.PC++
		switch opcode {
		case OpCode_PUSH_NIL:
			state.Push(NIL)

		case OpCode_PUSH_CONSTANT:
			{
				state.Push(program.Constants[state.CC])
				state.CC++
			}

		case OpCode_OPEN_FRAME:
			state.OpenFrame()

		case OpCode_CLOSE_FRAME:
			state.CloseFrame()

		case OpCode_RESOLVE_VALUE:
			{
				source := state.Pop()
				result := executor.resolveValue(source)
				if result.Code != ResultCode_OK {
					return result
				}
				state.Push(result.Value)
			}

		case OpCode_EXPAND_VALUE:
			state.Expand()

		case OpCode_SET_SOURCE:
			{
				source := state.Pop()
				state.Push(NewQualifiedValue(source, []Selector{}))
			}

		case OpCode_SELECT_INDEX:
			{
				index := state.Pop()
				value := state.Pop()
				result2, selector := CreateIndexedSelector(index)
				if result2.Code != ResultCode_OK {
					return result2
				}
				result := selector.Apply(value)
				if result.Code != ResultCode_OK {
					return result
				}
				state.Push(result.Value)
			}

		case OpCode_SELECT_KEYS:
			{
				keys := state.LastFrame
				value := state.Pop()
				result2, selector := CreateKeyedSelector(append([]Value{}, keys...))
				if result2.Code != ResultCode_OK {
					return result2
				}
				result := selector.Apply(value)
				if result.Code != ResultCode_OK {
					return result
				}
				state.Push(result.Value)
			}

		case OpCode_SELECT_RULES:
			{
				rules := state.LastFrame
				value := state.Pop()
				result, selector := executor.resolveSelector(rules)
				if result.Code != ResultCode_OK {
					return result
				}
				result2 := ApplySelector(value, selector)
				if result2.Code != ResultCode_OK {
					return result2
				}
				state.Push(result2.Value)
			}

		case OpCode_EVALUATE_SENTENCE:
			{
				args := state.LastFrame
				for len(args) > 0 {
					// Loop for successive command resolution
					cmdname := args[0]
					result, command := executor.resolveCommand(cmdname)
					if result.Code != ResultCode_OK {
						return result
					}
					lastCommand := state.Command
					state.Command = command
					if command == LAST_RESULT {
						// Intrinsic command: return last result = no-op
					} else if command == SHIFT_LAST_FRAME_RESULT {
						// // Intrinsic command: swap last frame result with argument 1
						if len(args) < 2 || lastCommand == command {
							// No-op
						} else {
							args[0] = args[1]
							args[1] = state.lastFrameResult().Value
							continue
						}
					} else {
						// Execute regular command
						state.SetResult(state.Command.Execute(args, executor.Context))
					}
					if state.Result.Code != ResultCode_OK {
						return state.Result
					}
					break
				}
			}

		case OpCode_PUSH_RESULT:
			state.Push(state.Result.Value)

		case OpCode_JOIN_STRINGS:
			{
				values := state.LastFrame
				s := ""
				for _, value := range values {
					result, s2 := ValueToString(value)
					if result.Code != ResultCode_OK {
						return result
					}
					s += s2
				}
				state.Push(NewStringValue(s))
			}

		case OpCode_MAKE_TUPLE:
			{
				values := state.LastFrame
				state.Push(NewTupleValue(append([]Value{}, values...)))
			}

		default:
			panic("CANTHAPPEN")
		}
	}
	return OK(NIL)
}

// Resolve value
//
// - If source value is a tuple, resolve each of its elements recursively
// - If source value is a qualified word, resolve source and apply selectors
// - Else, resolve variable from the source string value
func (executor *Executor) resolveValue(source Value) Result {
	switch source.Type() {
	case ValueType_TUPLE:
		return executor.resolveTuple(source.(TupleValue))
	case ValueType_QUALIFIED:
		return executor.resolveQualified(source.(QualifiedValue))
	default:
		{
			result, varname := ValueToString(source)
			if result.Code != ResultCode_OK {
				return ERROR("invalid variable name")
			}
			return executor.resolveVariable(varname)
		}
	}
}

func (executor *Executor) resolveQualified(qualified QualifiedValue) Result {
	result := executor.resolveValue(qualified.Source)
	if result.Code != ResultCode_OK {
		return result
	}
	for _, selector := range qualified.Selectors {
		result = selector.Apply(result.Value)
		if result.Code != ResultCode_OK {
			return result
		}
	}
	return result
}

// Resolve tuple values recursively
func (executor *Executor) resolveTuple(tuple TupleValue) Result {
	values := make([]Value, len(tuple.Values))
	for i, value := range tuple.Values {
		var result Result
		switch value.Type() {
		case ValueType_TUPLE:
			result = executor.resolveTuple(value.(TupleValue))
		default:
			result = executor.resolveValue(value)
		}
		if result.Code != ResultCode_OK {
			return result
		}
		values[i] = result.Value
	}
	return OK(NewTupleValue(values))
}

func (executor *Executor) resolveVariable(varname string) Result {
	if executor.VariableResolver == nil {
		return ERROR("no variable resolver")
	}
	value := executor.VariableResolver.Resolve(varname)
	if value == nil {
		return ERROR(`cannot resolve variable "` + varname + `"`)
	}
	return OK(value)
}

func (executor *Executor) resolveCommand(cmdname Value) (Result, Command) {
	if executor.CommandResolver == nil {
		return ERROR("no command resolver"), nil
	}
	command := executor.CommandResolver.Resolve(cmdname)
	if command == nil {
		result, name := ValueToString(cmdname)
		if result.Code != ResultCode_OK {
			return ERROR("invalid command name"), nil
		}
		return ERROR(`cannot resolve command "` + name + `"`), nil
	}
	return OK(NIL), command
}

func (executor *Executor) resolveSelector(rules []Value) (Result, Selector) {
	if executor.SelectorResolver == nil {
		return ERROR("no selector resolver"), nil
	}
	result, s := executor.SelectorResolver.Resolve(rules)
	if result.Code != ResultCode_OK {
		return result, nil
	}
	if s == nil {
		return ERROR(`cannot resolve selector {` + DisplayList(rules, nil) + `}`), nil
	}
	return result, s
}
