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
	compiler.emitSentences(program, sentences)
	program.PushOpCode(OpCode_MAKE_TUPLE)
	return program
}
func (compiler Compiler) emitSentences(program *Program, sentences []Sentence) {
	program.PushOpCode(OpCode_OPEN_FRAME)
	for _, sentence := range sentences {
		compiler.emitSentence(program, sentence)
	}
	program.PushOpCode(OpCode_CLOSE_FRAME)
}

// Compile the given sentence into a program
func (compiler Compiler) CompileSentence(sentence Sentence) *Program {
	program := &Program{}
	compiler.emitSentence(program, sentence)
	return program
}
func (compiler Compiler) emitSentence(program *Program, sentence Sentence) {
	for _, word := range sentence.Words {
		if word.Value == nil {
			compiler.emitWord(program, word.Word)
		} else {
			compiler.emitConstant(program, word.Value)
		}
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
		panic(InvalidWordStructureError)
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
		panic(UnexpectedMorphemeError)
	}
}
func (compiler Compiler) emitCompound(program *Program, morphemes []Morpheme) {
	program.PushOpCode(OpCode_OPEN_FRAME)
	compiler.emitStems(program, morphemes)
	program.PushOpCode(OpCode_CLOSE_FRAME)
	program.PushOpCode(OpCode_JOIN_STRINGS)
}
func (compiler Compiler) emitSubstitution(program *Program, morphemes []Morpheme) {
	expand := (morphemes[0].(SubstituteNextMorpheme)).Expansion
	levels := 1
	i := 1
	for morphemes[i].Type() == MorphemeType_SUBSTITUTE_NEXT {
		i++
		levels++
	}
	selectable := morphemes[i]
	i++
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
		panic(UnexpectedMorphemeError)
	}
	for i < len(morphemes) {
		morpheme := morphemes[i]
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
	for level := 1; level < levels; level++ {
		program.PushOpCode(OpCode_RESOLVE_VALUE)
	}
	if expand {
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
		panic(UnexpectedMorphemeError)
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
	var levels int
	for _, morpheme := range morphemes {
		if mode == SELECTABLE {
			switch morpheme.Type() {
			case MorphemeType_SUBSTITUTE_NEXT,
				MorphemeType_LITERAL:
				{
					// Terminate substitution sequence
					for level := 1; level < levels; level++ {
						program.PushOpCode(OpCode_RESOLVE_VALUE)
					}
					mode = INITIAL
				}
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
}

//
// Morphemes
//

func (compiler Compiler) emitLiteral(program *Program, literal LiteralMorpheme) {
	value := NewStringValue(literal.Value)
	compiler.emitConstant(program, value)
}
func (compiler Compiler) emitTuple(program *Program, tuple TupleMorpheme) {
	compiler.emitSentences(program, tuple.Subscript.Sentences)
	program.PushOpCode(OpCode_MAKE_TUPLE)
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
	compiler.emitSentences(program, tuple.Subscript.Sentences)
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
		program.PushOpCode(OpCode_MAKE_TUPLE)
	}
	program.PushOpCode(OpCode_CLOSE_FRAME)
	program.PushOpCode(OpCode_SELECT_RULES)
}
func (compiler Compiler) emitConstant(program *Program, value Value) {
	program.PushOpCode(OpCode_PUSH_CONSTANT)
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

	// Last closed frame
	LastFrame []Value

	// Program counter
	PC uint

	// Constant counter
	CC uint

	// Last executed command
	Command Command

	// Last executed result value
	Result Result
}

func NewProgramState() *ProgramState {
	return &ProgramState{
		stack:  []Value{},
		frames: []int{0},
		PC:     0,
		CC:     0,
		Result: OK(NIL),
	}
}

// Reset program state
func (state *ProgramState) Reset() {
	state.stack = state.stack[:0]
	state.frames = state.frames[:1]
	state.frames[0] = 0
	state.LastFrame = nil
	state.PC = 0
	state.CC = 0
	state.Command = nil
	state.Result = OK(NIL)
}

// Open a new frame
func (state *ProgramState) OpenFrame() {
	state.frames = append(state.frames, len(state.stack))
}

// Close the current frame
func (state *ProgramState) CloseFrame() {
	length := state.frames[len(state.frames)-1]
	state.frames = state.frames[:len(state.frames)-1]
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
func (state *ProgramState) last() Value {
	if len(state.stack) == 0 {
		return nil
	}
	return state.stack[len(state.stack)-1]
}

// Expand last value in current frame
func (state *ProgramState) Expand() {
	last := state.last()
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
	if state.Result.Code == ResultCode_YIELD {
		if resumable, ok := state.Command.(ResumableCommand); ok {
			state.Result = resumable.Resume(state.Result, executor.Context)
			if state.Result.Code != ResultCode_OK {
				return state.Result
			}
		}
	}
	for state.PC < uint(len(program.OpCodes)) {
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
				result2 := CreateIndexedSelector(index)
				if result2.Code != ResultCode_OK {
					return result2.AsResult()
				}
				selector := result2.Data
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
				result2 := CreateKeyedSelector(append([]Value{}, keys...))
				if result2.Code != ResultCode_OK {
					return result2.AsResult()
				}
				selector := result2.Data
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
				result := executor.resolveSelector(rules)
				if result.Code != ResultCode_OK {
					return result.AsResult()
				}
				selector := result.Data
				result2 := ApplySelector(value, selector)
				if result2.Code != ResultCode_OK {
					return result2
				}
				state.Push(result2.Value)
			}

		case OpCode_EVALUATE_SENTENCE:
			{
				args := state.LastFrame
				if len(args) > 0 {
					cmdname := args[0]
					result := executor.resolveCommand(cmdname)
					if result.Code != ResultCode_OK {
						return result.AsResult()
					}
					command := result.Data
					state.Command = command
					state.Result = state.Command.Execute(args, executor.Context)
					if state.Result.Code != ResultCode_OK {
						return state.Result
					}
				}
			}

		case OpCode_PUSH_RESULT:
			state.Push(state.Result.Value)

		case OpCode_JOIN_STRINGS:
			{
				values := state.LastFrame
				s := ""
				for _, value := range values {
					result := ValueToString(value)
					if result.Code != ResultCode_OK {
						return result.AsResult()
					}
					s += result.Data
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
	if !state.Empty() {
		state.Result = OK(state.Pop())
	}
	return state.Result
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
			result := ValueToString(source)
			if result.Code != ResultCode_OK {
				return ERROR("invalid variable name")
			}
			varname := result.Data
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

func (executor *Executor) resolveCommand(cmdname Value) TypedResult[Command] {
	if executor.CommandResolver == nil {
		return ERROR_T[Command]("no command resolver")
	}
	command := executor.CommandResolver.Resolve(cmdname)
	if command == nil {
		result := ValueToString(cmdname)
		if result.Code != ResultCode_OK {
			return ERROR_T[Command]("invalid command name")
		}
		name := result.Data
		return ERROR_T[Command](`cannot resolve command "` + name + `"`)
	}
	return OK_T(NIL, command)
}

func (executor *Executor) resolveSelector(rules []Value) TypedResult[Selector] {
	if executor.SelectorResolver == nil {
		return ERROR_T[Selector]("no selector resolver")
	}
	result := executor.SelectorResolver.Resolve(rules)
	if result.Code != ResultCode_OK {
		return result
	}
	if result.Data == nil {
		return ERROR_T[Selector](`cannot resolve selector {` + DisplayList(rules, nil) + `}`)
	}
	return result
}
