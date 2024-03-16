package helena_dialect

import "helena/core"

// import { ERROR, OK, Result, ResultCode, RETURN, YIELD } from "../core/results";
// import { Command } from "../core/command";
// import {
//   Compiler,
//   Executor,
//   OpCode,
//   Program,
//   ProgramState,
// } from "../core/compiler";
// import { VariableResolver, CommandResolver } from "../core/resolvers";
// import { Script } from "../core/syntax";
// import {
//   Value,
//   ValueType,
//   CustomValueType,
//   ScriptValue,
//   RealValue,
//   TupleValue,
//   NIL,
//   StringValue,
//   CommandValue,
// } from "../core/values";
// import { numberCmd } from "./numbers";

// const deferredValueType: CustomValueType = { name: "deferred" };
type DeferredValue struct {
	// type = deferredValueType;
	Scope   *Scope
	Program *core.Program
}

func (DeferredValue) Type() core.ValueType {
	return -1
}
func CreateDeferredValue(code core.ResultCode, value core.Value, scope *Scope) core.Result {
	var program *core.Program
	switch value.Type() {
	case core.ValueType_SCRIPT:
		program = scope.CompileScriptValue(value.(core.ScriptValue))
	case core.ValueType_TUPLE:
		program = scope.CompileTupleValue(value.(core.TupleValue))
	default:
		return core.ERROR("body must be a script or tuple")
	}
	return core.Result{
		Code:  code,
		Value: DeferredValue{scope, program},
	}
}

type ProcessContext struct {
	scope   *Scope
	program *core.Program
	state   *core.ProgramState
}
type Process struct {
	contextStack []ProcessContext
}

func NewProcess(scope *Scope, program *core.Program) *Process {
	process := &Process{[]ProcessContext{}}
	process.pushContext(scope, program)
	return process
}

func (process *Process) currentContext() ProcessContext {
	return process.contextStack[len(process.contextStack)-1]
}
func (process *Process) pushContext(scope *Scope, program *core.Program) {
	process.contextStack = append(process.contextStack, ProcessContext{scope, program, core.NewProgramState()})
}
func (process *Process) popContext() {
	process.contextStack = process.contextStack[:len(process.contextStack)-1]
}
func (process *Process) Run() core.Result {
	for {
		context := process.currentContext()
		result := context.scope.Execute(context.program, context.state)
		if deferred, ok := result.Value.(DeferredValue); ok {
			process.pushContext(deferred.Scope, deferred.Program)
			continue
		}
		if result.Code == core.ResultCode_OK && len(process.contextStack) > 1 {
			process.popContext()
			previousResult := process.currentContext()
			switch previousResult.state.Result.Code {
			case core.ResultCode_OK:
				return core.OK(result.Value)
			case core.ResultCode_RETURN:
				return core.RETURN(result.Value)
			case core.ResultCode_YIELD:
				process.YieldBack(result.Value)
				continue
			default:
				return core.ERROR("unexpected deferred result")
			}
		}
		return result
	}
}
func (process *Process) YieldBack(value core.Value) {
	context := process.currentContext()
	context.state.Result.Value = value
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

type Scope struct {
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

func NewScope(
	parent *Scope,
	shared bool,
) *Scope {
	scope := &Scope{}
	if shared {
		scope.Context = parent.Context
	} else if parent != nil && parent.Context != nil {
		scope.Context = newScopeContext(parent.Context)
	} else {
		scope.Context = newScopeContext(nil)
	}
	scope.compiler = core.Compiler{}
	scope.executor = core.Executor{
		VariableResolver: variableResolver{scope},
		CommandResolver:  commandResolver{scope},
		SelectorResolver: nil,
		Context:          scope,
	}
	return scope
}

func (scope *Scope) ExecuteScriptValue(script core.ScriptValue) core.Result {
	return scope.ExecuteScript(script.Script)
}
func (scope *Scope) ExecuteScript(script core.Script) core.Result {
	return scope.PrepareScript(script).Run()
}
func (scope *Scope) CompileScriptValue(script core.ScriptValue) *core.Program {
	return scope.Compile(script.Script)
}
func (scope *Scope) CompileTupleValue(tuple core.TupleValue) *core.Program {
	program := &core.Program{}
	program.PushOpCode(core.OpCode_PUSH_CONSTANT)
	program.PushOpCode(core.OpCode_EVALUATE_SENTENCE)
	program.PushOpCode(core.OpCode_PUSH_RESULT)
	program.PushConstant(tuple)
	return program
}
func (scope *Scope) Compile(script core.Script) *core.Program {
	return scope.compiler.CompileScript(script)
}
func (scope *Scope) Execute(program *core.Program, state *core.ProgramState) core.Result {
	return scope.executor.Execute(program, state)
}

func (scope *Scope) PrepareScriptValue(script core.ScriptValue) *Process {
	return scope.PrepareProcess(scope.CompileScriptValue(script))
}
func (scope *Scope) PrepareTupleValue(tuple core.TupleValue) *Process {
	return scope.PrepareProcess(scope.CompileTupleValue(tuple))
}
func (scope *Scope) PrepareScript(script core.Script) *Process {
	return scope.PrepareProcess(scope.Compile(script))
}
func (scope *Scope) PrepareProcess(program *core.Program) *Process {
	return NewProcess(scope, program)
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
	if value.Type() == core.ValueType_TUPLE {
		return expandPrefixCmd
	}
	if value.Type() == core.ValueType_COMMAND {
		return value.(core.CommandValue).Command
	}
	if core.ValueIsNumber(value) {
		return numberCmd
	}
	result := core.ValueToString(value)
	if result.Code != core.ResultCode_OK {
		return nil
	}
	cmdname := result.Data
	return scope.ResolveNamedCommand(cmdname)
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

//   setNamedLocal(name: string, value: Value) {
//     this.locals.set(name, value);
//   }
//   destructureLocal(constant: Value, value: Value, check: boolean) core.Result {
//     const { data: name, code } = core.ValueToString(constant);
//     if (code != core.ResultCode_OK) return core.ERROR("invalid local name");
//     if (check) return core.OK(NIL);
//     this.setNamedLocal(name, value);
//     return core.OK(NIL);
//   }
func (scope *Scope) SetNamedConstant(name string, value core.Value) core.Result {
	result := scope.checkNamedConstant(name)
	if result.Code != core.ResultCode_OK {
		return result
	}
	scope.Context.Constants[name] = value
	return core.OK(value)
}
func (scope *Scope) DestructureConstant(constant core.Value, value core.Value, check bool) core.Result {
	result := core.ValueToString(constant)
	if result.Code != core.ResultCode_OK {
		return core.ERROR("invalid constant name")
	}
	name := result.Data
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
	result := core.ValueToString(variable)
	if result.Code != core.ResultCode_OK {
		return core.ERROR("invalid variable name")
	}
	name := result.Data
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
	result := core.ValueToString(variable)
	if result.Code != core.ResultCode_OK {
		return core.ERROR("invalid variable name")
	}
	name := result.Data
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
	result := core.ValueToString(variable)
	if result.Code != core.ResultCode_OK {
		return core.ERROR("invalid variable name")
	}
	name := result.Data
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
	program.PushOpCode(core.OpCode_PUSH_CONSTANT)
	program.PushOpCode(core.OpCode_RESOLVE_VALUE)
	program.PushConstant(value)
	return scope.Execute(program, nil)
}

func (scope *Scope) RegisterCommand(name core.Value, command core.Command) core.Result {
	result := core.ValueToString(name)
	if result.Code != core.ResultCode_OK {
		return core.ERROR("invalid command name")
	}
	cmdname := result.Data
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
		result := core.ValueToString(args2[0])
		if result.Code != core.ResultCode_OK {
			return core.ERROR(`invalid command name`)
		} else {
			cmdname := result.Data
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
