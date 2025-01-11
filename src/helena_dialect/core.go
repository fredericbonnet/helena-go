package helena_dialect

import (
	"helena/core"
	"sync"
)

type ContinuationValue struct {
	Scope    *Scope
	Program  *core.Program
	Callback func(result core.Result) core.Result
}

func NewContinuationValue(scope *Scope, program *core.Program, callback func(result core.Result) core.Result) ContinuationValue {
	return ContinuationValue{scope, program, callback}
}
func (ContinuationValue) Type() core.ValueType {
	return core.ValueType_CUSTOM
}
func (ContinuationValue) CustomType() core.CustomValueType {
	return core.CustomValueType{Name: "continuation"}
}
func CreateContinuationValue(scope *Scope, program *core.Program, callback func(result core.Result) core.Result) core.Result {
	return core.YIELD(NewContinuationValue(scope, program, callback))
}

type ProcessContext struct {
	scope    *Scope
	program  *core.Program
	state    *core.ProgramState
	callback func(result core.Result) core.Result
}
type ProcessStack struct {
	stack []ProcessContext
}

var programStatePool = sync.Pool{
	New: func() any {
		return core.NewProgramState()
	},
}

func NewProcessStack() ProcessStack {
	return ProcessStack{[]ProcessContext{}}
}
func (processStack *ProcessStack) Depth() uint {
	return uint(len(processStack.stack))
}
func (processStack *ProcessStack) CurrentContext() ProcessContext {
	return processStack.stack[len(processStack.stack)-1]
}
func (processStack *ProcessStack) PushProgram(scope *Scope, program *core.Program) ProcessContext {
	context := ProcessContext{
		scope,
		program,
		programStatePool.Get().(*core.ProgramState),
		nil,
	}
	processStack.stack = append(processStack.stack, context)
	return context
}
func (processStack *ProcessStack) PushContinuation(continuation ContinuationValue) ProcessContext {
	context := ProcessContext{
		continuation.Scope,
		continuation.Program,
		programStatePool.Get().(*core.ProgramState),
		continuation.Callback,
	}
	processStack.stack = append(processStack.stack, context)
	return context
}
func (processStack *ProcessStack) Pop() {
	last := processStack.stack[len(processStack.stack)-1]
	last.state.Reset()
	programStatePool.Put(last.state)
	processStack.stack = processStack.stack[:len(processStack.stack)-1]
}

type ProcessOptions struct {
	CaptureErrorStack bool
}
type Process struct {
	options ProcessOptions
	stack   ProcessStack
}

func NewProcess(scope *Scope, program *core.Program, options *ProcessOptions) *Process {
	process := &Process{}
	if options == nil {
		process.options = ProcessOptions{false}
	} else {
		process.options = *options
	}
	process.stack = NewProcessStack()
	process.stack.PushProgram(scope, program)
	return process
}

func (process *Process) Run() core.Result {
	context := process.stack.CurrentContext()
	result := context.scope.Execute(context.program, context.state)
	for process.stack.Depth() > 0 {
		if continuation, ok := result.Value.(ContinuationValue); ok {
			if result.Code != core.ResultCode_YIELD && context.callback == nil {
				// End and replace current context
				process.stack.Pop()
			}

			// Push and execute result continuation context
			context = process.stack.PushContinuation(continuation)
			result = context.scope.Execute(context.program, context.state)
			continue
		}

		if result.Code == core.ResultCode_YIELD {
			// Yield result to caller
			break
		}
		if context.callback != nil {
			// Process result with callback
			result = context.callback(result)
		}

		if result.Code == core.ResultCode_ERROR {
			if process.options.CaptureErrorStack {
				// Push to error stack
				if result.Data == nil {
					result = core.Result{
						Code:  result.Code,
						Value: result.Value,
						Data:  core.NewErrorStack(),
					}
				}
				errorStack := result.Data.(*core.ErrorStack)
				var level core.ErrorStackLevel
				var frame = append([]core.Value{}, context.state.LastFrame...)
				if context.program.OpCodePositions != nil {
					level = core.ErrorStackLevel{
						Frame:    &frame,
						Source:   context.program.Source,
						Position: context.program.OpCodePositions[context.state.PC-1],
					}
				} else {
					level = core.ErrorStackLevel{
						Frame: &frame,
					}
				}
				errorStack.Push(level)
			} else if result.Data != nil {
				// Erase error stack from result
				result = core.Result{
					Code:  result.Code,
					Value: result.Value,
				}
			}
		}

		if process.stack.Depth() == 1 {
			// Reached bottom of stack, stop there
			break
		}
		process.stack.Pop()

		context = process.stack.CurrentContext()
		if _, ok := result.Value.(ContinuationValue); ok {
			// Process continuation above
			continue
		}
		if result.Code != core.ResultCode_OK {
			// Pass result down to previous context
			continue
		}

		// Yield back and resume current context
		context.state.SetResult(result)
		result = context.scope.Execute(context.program, context.state)
	}
	return result
}
func (process *Process) SetResult(result core.Result) {
	context := process.stack.CurrentContext()
	context.state.SetResult(result)
}
func (process *Process) YieldBack(value core.Value) {
	context := process.stack.CurrentContext()
	result := context.state.Result
	result.Value = value
	context.state.SetResult(result)
}

type scopeContext struct {
	parent    *scopeContext
	Constants map[string]core.Value
	Variables map[string]core.Value
	Commands  map[string]core.Command
}

func newScopeContext(parent *scopeContext) *scopeContext {
	return &scopeContext{
		parent:    parent,
		Constants: map[string]core.Value{},
		Variables: map[string]core.Value{},
		Commands:  map[string]core.Command{},
	}
}

type ScopeOptions struct {
	CapturePositions  bool
	CaptureErrorStack bool
}
type Scope struct {
	options  ScopeOptions
	Context  *scopeContext
	locals   map[string]core.Value
	compiler core.Compiler
	executor core.Executor
}

type variableResolver struct{ scope *Scope }

func (resolver variableResolver) Resolve(name string) core.Value {
	return resolver.scope.ResolveVariable(name)
}

type commandResolver struct{ scope *Scope }

func (resolver commandResolver) Resolve(name core.Value) core.Command {
	return resolver.scope.ResolveCommand(name)
}

func newScope(
	context *scopeContext,
	options *ScopeOptions,
) *Scope {
	scope := &Scope{}
	if options == nil {
		scope.options = ScopeOptions{
			CaptureErrorStack: false,
			CapturePositions:  false,
		}
	} else {
		scope.options = *options
	}
	scope.Context = context
	scope.locals = map[string]core.Value{}
	scope.compiler = core.NewCompiler(&core.CompilerOptions{
		CapturePositions: scope.options.CapturePositions,
	})
	scope.executor = core.Executor{
		VariableResolver: variableResolver{scope},
		CommandResolver:  commandResolver{scope},
		SelectorResolver: nil,
		Context:          scope,
	}
	return scope
}
func NewRootScope(options *ScopeOptions) *Scope {
	return newScope(newScopeContext(nil), options)
}
func (scope *Scope) NewChildScope() *Scope {
	return newScope(newScopeContext(scope.Context), &scope.options)
}
func (scope *Scope) NewLocalScope() *Scope {
	return newScope(scope.Context, &scope.options)
}

func (scope *Scope) Compile(script core.Script) *core.Program {
	return scope.compiler.CompileScript(script)
}
func (scope *Scope) Execute(program *core.Program, state *core.ProgramState) core.Result {
	return scope.executor.Execute(program, state)
}

func (scope *Scope) CompileScriptValue(script core.ScriptValue) *core.Program {
	return scope.Compile(script.Script)
}
func (scope *Scope) CompileTupleValue(tuple core.TupleValue) *core.Program {
	program := &core.Program{}
	program.PushOpCode(core.OpCode_OPEN_FRAME, nil)
	program.PushOpCode(core.OpCode_PUSH_CONSTANT, nil)
	program.PushOpCode(core.OpCode_EXPAND_VALUE, nil)
	program.PushOpCode(core.OpCode_CLOSE_FRAME, nil)
	program.PushOpCode(core.OpCode_EVALUATE_SENTENCE, nil)
	program.PushOpCode(core.OpCode_PUSH_RESULT, nil)
	program.PushConstant(tuple)
	return program
}
func (scope *Scope) CompileArgs(args ...core.Value) *core.Program {
	program := &core.Program{}
	program.PushOpCode(core.OpCode_OPEN_FRAME, nil)
	for _, arg := range args {
		program.PushOpCode(core.OpCode_PUSH_CONSTANT, nil)
		program.PushConstant(arg)
	}
	program.PushOpCode(core.OpCode_CLOSE_FRAME, nil)
	program.PushOpCode(core.OpCode_EVALUATE_SENTENCE, nil)
	program.PushOpCode(core.OpCode_PUSH_RESULT, nil)
	return program
}

func (scope *Scope) PrepareProcess(program *core.Program) *Process {
	return NewProcess(scope, program, &ProcessOptions{
		CaptureErrorStack: scope.options.CaptureErrorStack,
	})
}

func (scope *Scope) ResolveVariable(name string) core.Value {
	v := scope.locals[name]
	if v != nil {
		return v
	}
	v = scope.Context.Constants[name]
	if v != nil {
		return v
	}
	v = scope.Context.Variables[name]
	if v != nil {
		return v
	}
	return nil
}
func (scope *Scope) ResolveCommand(value core.Value) core.Command {
	switch value.Type() {
	case core.ValueType_TUPLE:
		return expandPrefixCmd
	case core.ValueType_COMMAND:
		return value.(core.CommandValue).Command()
	case core.ValueType_INTEGER,
		core.ValueType_REAL:
		return numberCmd
	}
	result, cmdname := core.ValueToString(value)
	if result.Code != core.ResultCode_OK {
		return nil
	}
	command := scope.ResolveNamedCommand(cmdname)
	if command != nil {
		return command
	}
	if core.StringIsNumber(cmdname) {
		return numberCmd
	}
	return nil
}
func (scope *Scope) ResolveNamedCommand(name string) core.Command {
	context := scope.Context
	for context != nil {
		command := context.Commands[name]
		if command != nil {
			return command
		}
		context = context.parent
	}
	return nil
}

func (scope *Scope) ClearLocals() {
	clear(scope.locals)
}
func (scope *Scope) SetNamedLocal(name string, value core.Value) {
	scope.locals[name] = value
}
func (scope *Scope) DestructureLocal(constant core.Value, value core.Value, check bool) core.Result {
	result, name := core.ValueToString(constant)
	if result.Code != core.ResultCode_OK {
		return core.ERROR("invalid local name")
	}
	if check {
		core.OK(core.NIL)
	}
	scope.SetNamedLocal(name, value)
	return core.OK(core.NIL)
}
func (scope *Scope) SetNamedConstant(name string, value core.Value) core.Result {
	result := scope.checkNamedConstant(name)
	if result.Code != core.ResultCode_OK {
		return result
	}
	scope.Context.Constants[name] = value
	return core.OK(value)
}
func (scope *Scope) DestructureConstant(constant core.Value, value core.Value, check bool) core.Result {
	result, name := core.ValueToString(constant)
	if result.Code != core.ResultCode_OK {
		return core.ERROR("invalid constant name")
	}
	if check {
		return scope.checkNamedConstant(name)
	}
	scope.Context.Constants[name] = value
	return core.OK(core.NIL)
}
func (scope *Scope) checkNamedConstant(name string) core.Result {
	if scope.locals[name] != nil {
		return core.ERROR(`cannot define constant "` + name + `": local already exists`)
	}
	if scope.Context.Constants[name] != nil {
		return core.ERROR(`cannot redefine constant "` + name + `"`)
	}
	if scope.Context.Variables[name] != nil {
		return core.ERROR(`cannot define constant "` + name + `": variable already exists`)
	}
	return core.OK(core.NIL)
}
func (scope *Scope) SetNamedVariable(name string, value core.Value) core.Result {
	result := scope.checkNamedVariable(name)
	if result.Code != core.ResultCode_OK {
		return result
	}
	scope.Context.Variables[name] = value
	return core.OK(value)
}
func (scope *Scope) DestructureVariable(variable core.Value, value core.Value, check bool) core.Result {
	result, name := core.ValueToString(variable)
	if result.Code != core.ResultCode_OK {
		return core.ERROR("invalid variable name")
	}
	if check {
		return scope.checkNamedVariable(name)
	}
	scope.Context.Variables[name] = value
	return core.OK(core.NIL)
}
func (scope *Scope) checkNamedVariable(name string) core.Result {
	if scope.locals[name] != nil {
		return core.ERROR(`cannot redefine local "` + name + `"`)
	}
	if scope.Context.Constants[name] != nil {
		return core.ERROR(`cannot redefine constant "` + name + `"`)
	}
	return core.OK(core.NIL)
}
func (scope *Scope) UnsetVariable(variable core.Value, check bool) core.Result {
	result, name := core.ValueToString(variable)
	if result.Code != core.ResultCode_OK {
		return core.ERROR("invalid variable name")
	}
	if scope.locals[name] != nil {
		return core.ERROR(`cannot unset local "` + name + `"`)
	}
	if scope.Context.Constants[name] != nil {
		return core.ERROR(`cannot unset constant "` + name + `"`)
	}
	if scope.Context.Variables[name] == nil {
		return core.ERROR(`cannot unset "` + name + `": no such variable`)
	}
	if check {
		return core.OK(core.NIL)
	}
	delete(scope.Context.Variables, name)
	return core.OK(core.NIL)
}
func (scope *Scope) GetVariable(variable core.Value, def core.Value) core.Result {
	result, name := core.ValueToString(variable)
	if result.Code != core.ResultCode_OK {
		return core.ERROR("invalid variable name")
	}
	value := scope.ResolveVariable(name)
	if value != nil {
		return core.OK(value)
	}
	if def != nil {
		return core.OK(def)
	}
	return core.ERROR(`cannot get "` + name + `": no such variable`)
}
func (scope *Scope) ResolveValue(value core.Value) core.Result {
	program := &core.Program{}
	program.PushOpCode(core.OpCode_PUSH_CONSTANT, nil)
	program.PushOpCode(core.OpCode_RESOLVE_VALUE, nil)
	program.PushConstant(value)
	return scope.Execute(program, nil)
}

func (scope *Scope) RegisterCommand(name core.Value, command core.Command) core.Result {
	result, cmdname := core.ValueToString(name)
	if result.Code != core.ResultCode_OK {
		return core.ERROR("invalid command name")
	}
	scope.RegisterNamedCommand(cmdname, command)
	return core.OK(core.NIL)
}
func (scope *Scope) RegisterNamedCommand(name string, command core.Command) {
	scope.Context.Commands[name] = command
}
func (scope *Scope) HasLocalCommand(name string) bool {
	return scope.Context.Commands[name] != nil
}
func (scope *Scope) GetLocalCommands() []string {
	names := make([]string, 0, len(scope.Context.Commands))
	for name := range scope.Context.Commands {
		names = append(names, name)
	}
	return names
}

type ExpandPrefixState struct {
	command core.Command
	result  core.Result
}
type ExpandPrefixCommand struct{}

func (ExpandPrefixCommand) Execute(args []core.Value, context any) core.Result {
	scope := context.(*Scope)
	command, args2 := resolveLeadingTuple(args, scope)
	if command == nil {
		if len(args2) == 0 {
			return core.OK(core.NIL)
		}
		result, cmdname := core.ValueToString(args2[0])
		if result.Code != core.ResultCode_OK {
			return core.ERROR(`invalid command name`)
		} else {
			return core.ERROR(`cannot resolve command "` + cmdname + `"`)

		}
	}
	result := command.Execute(args2, scope)
	if result.Code == core.ResultCode_YIELD {

		state := ExpandPrefixState{command, result}
		return core.YIELD_STATE(state.result.Value, state)
	}
	return result
}
func (ExpandPrefixCommand) Resume(result core.Result, context any) core.Result {
	scope := context.(*Scope)
	state := result.Data.(ExpandPrefixState)
	command := state.command
	commandResult := state.result
	resumable, ok := command.(core.ResumableCommand)
	if !ok {
		return core.OK(result.Value)
	}
	result2 := resumable.Resume(
		core.Result{
			Code:  commandResult.Code,
			Value: result.Value,
			Data:  commandResult.Data,
		},
		scope,
	)
	if result2.Code == core.ResultCode_YIELD {
		return core.YIELD_STATE(result2.Value, ExpandPrefixState{command, result2})
	}
	return result2
}

var expandPrefixCmd = ExpandPrefixCommand{}

func resolveLeadingTuple(args []core.Value, scope *Scope) (core.Command, []core.Value) {
	if len(args) == 0 {
		return nil, nil
	}
	lead, rest := args[0], args[1:]
	if lead.Type() != core.ValueType_TUPLE {
		command := scope.ResolveCommand(lead)
		return command, args
	}
	tuple := lead.(core.TupleValue)
	return resolveLeadingTuple(append(append([]core.Value{}, tuple.Values...), rest...), scope)
}

func DestructureValue(
	apply func(name core.Value, value core.Value, check bool) core.Result,
	shape core.Value,
	value core.Value,
) core.Result {
	result := checkValues(apply, shape, value)
	if result.Code != core.ResultCode_OK {
		return result
	}
	applyValues(apply, shape, value)
	return core.OK(value)
}
func checkValues(
	apply func(name core.Value, value core.Value, check bool) core.Result,
	shape core.Value,
	value core.Value,
) core.Result {
	if shape.Type() != core.ValueType_TUPLE {
		return apply(shape, value, true)
	}
	if value.Type() != core.ValueType_TUPLE {
		return core.ERROR("bad value shape")
	}
	variables := shape.(core.TupleValue).Values
	values := value.(core.TupleValue).Values
	if len(values) < len(variables) {
		return core.ERROR("bad value shape")
	}
	for i := 0; i < len(variables); i++ {
		result := checkValues(apply, variables[i], values[i])
		if result.Code != core.ResultCode_OK {
			return result
		}
	}
	return core.OK(core.NIL)
}
func applyValues(
	apply func(name core.Value, value core.Value, check bool) core.Result,
	shape core.Value,
	value core.Value,
) {
	if shape.Type() != core.ValueType_TUPLE {
		apply(shape, value, false)
		return
	}
	variables := shape.(core.TupleValue).Values
	values := value.(core.TupleValue).Values
	for i := 0; i < len(variables); i++ {
		applyValues(apply, variables[i], values[i])
	}
}
