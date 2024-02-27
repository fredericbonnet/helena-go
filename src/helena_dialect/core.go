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

func (value DeferredValue) Type() core.ValueType {
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
	constants map[string]core.Value
	variables map[string]core.Value
	commands  map[string]core.Command
}

func newScopeContext(parent *scopeContext) *scopeContext {
	return &scopeContext{
		parent:    parent,
		constants: map[string]core.Value{},
		variables: map[string]core.Value{},
		commands:  map[string]core.Command{},
	}
}

type Scope struct {
	context  *scopeContext
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
		scope.context = parent.context
	} else if parent != nil && parent.context != nil {
		scope.context = newScopeContext(parent.context)
	} else {
		scope.context = newScopeContext(nil)
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
	v = scope.context.constants[name]
	if v != nil {
		return v
	}
	v = scope.context.variables[name]
	if v != nil {
		return v
	}
	return nil
}
func (scope *Scope) ResolveCommand(value core.Value) core.Command {
	// if (value.type == ValueType.TUPLE) return expandPrefixCmd;
	// if (value.type == ValueType.COMMAND) return (value as CommandValue).command;
	// if (RealValue.isNumber(value)) return numberCmd;
	result := core.ValueToString(value)
	if result.Code != core.ResultCode_OK {
		return nil
	}
	cmdname := result.Data
	return scope.ResolveNamedCommand(cmdname)
}
func (scope *Scope) ResolveNamedCommand(name string) core.Command {
	context := scope.context
	for context != nil {
		command := context.commands[name]
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
//   destructureLocal(constant: Value, value: Value, check: boolean): Result {
//     const { data: name, code } = StringValue.toString(constant);
//     if (code != ResultCode_OK) return ERROR("invalid local name");
//     if (check) return OK(NIL);
//     this.setNamedLocal(name, value);
//     return OK(NIL);
//   }
//   setNamedConstant(name: string, value: Value): Result {
//     const result = this.checkNamedConstant(name);
//     if (result.code != ResultCode_OK) return result;
//     this.context.constants.set(name, value);
//     return OK(value);
//   }
//   destructureConstant(constant: Value, value: Value, check: boolean): Result {
//     const { data: name, code } = StringValue.toString(constant);
//     if (code != ResultCode_OK) return ERROR("invalid constant name");
//     if (check) return this.checkNamedConstant(name);
//     this.context.constants.set(name, value);
//     return OK(NIL);
//   }
//   private checkNamedConstant(name: string): Result {
//     if (this.locals.has(name)) {
//       return ERROR(`cannot define constant "${name}": local already exists`);
//     }
//     if (this.context.constants.has(name)) {
//       return ERROR(`cannot redefine constant "${name}"`);
//     }
//     if (this.context.variables.has(name)) {
//       return ERROR(`cannot define constant "${name}": variable already exists`);
//     }
//     return OK(NIL);
//   }
//   setNamedVariable(name: string, value: Value): Result {
//     const result = this.checkNamedVariable(name);
//     if (result.code != ResultCode_OK) return result;
//     this.context.variables.set(name, value);
//     return OK(value);
//   }
//   destructureVariable(variable: Value, value: Value, check: boolean): Result {
//     const { data: name, code } = StringValue.toString(variable);
//     if (code != ResultCode_OK) return ERROR("invalid variable name");
//     if (check) return this.checkNamedVariable(name);
//     this.context.variables.set(name, value);
//     return OK(NIL);
//   }
//   private checkNamedVariable(name: string): Result {
//     if (this.locals.has(name)) {
//       return ERROR(`cannot redefine local "${name}"`);
//     }
//     if (this.context.constants.has(name)) {
//       return ERROR(`cannot redefine constant "${name}"`);
//     }
//     return OK(NIL);
//   }
//   unsetVariable(variable: Value, check = false): Result {
//     const { data: name, code } = StringValue.toString(variable);
//     if (code != ResultCode_OK) return ERROR("invalid variable name");
//     if (this.locals.has(name)) {
//       return ERROR(`cannot unset local "${name}"`);
//     }
//     if (this.context.constants.has(name)) {
//       return ERROR(`cannot unset constant "${name}"`);
//     }
//     if (!this.context.variables.has(name)) {
//       return ERROR(`cannot unset "${name}": no such variable`);
//     }
//     if (check) return OK(NIL);
//     this.context.variables.delete(name);
//     return OK(NIL);
//   }
//   getVariable(variable: Value, def?: Value): Result {
//     const { data: name, code } = StringValue.toString(variable);
//     if (code != ResultCode_OK) return ERROR("invalid variable name");
//     const value = this.resolveVariable(name);
//     if (value) return OK(value);
//     if (def) return OK(def);
//     return ERROR(`cannot get "${name}": no such variable`);
//   }
//   resolveValue(value: Value): Result {
//     const program = new Program();
//     program.pushOpCode(OpCode.PUSH_CONSTANT);
//     program.pushOpCode(OpCode.RESOLVE_VALUE);
//     program.pushConstant(value);
//     return this.execute(program);
//   }

//   registerCommand(name: Value, command: Command): Result {
//     const { data: cmdname, code } = StringValue.toString(name);
//     if (code != ResultCode_OK) return ERROR("invalid command name");
//     this.registerNamedCommand(cmdname, command);
//     return OK(NIL);
//   }
func (scope *Scope) RegisterNamedCommand(name string, command core.Command) {
	scope.context.commands[name] = command
}

//   hasLocalCommand(name: string): boolean {
//     return this.context.commands.has(name);
//   }
//   getLocalCommands(): string[] {
//     return [...this.context.commands.keys()];
//   }
// }

// type ExpandPrefixState = {
//   command: Command;
//   result: Result;
// };
// export const expandPrefixCmd: Command = {
//   execute(args: Value[], scope: Scope): Result {
//     const [command, args2] = resolveLeadingTuple(args, scope);
//     if (!command) {
//       if (!args2 || args2.length == 0) return OK(NIL);
//       const { data: cmdname, code } = StringValue.toString(args2[0]);
//       return ERROR(
//         code != ResultCode_OK
//           ? `invalid command name`
//           : `cannot resolve command "${cmdname}"`
//       );
//     }
//     const result = command.execute(args2, scope);
//     if (result.code == ResultCode.YIELD) {
//       const state = { command, result } as ExpandPrefixState;
//       return YIELD(state.result.value, state);
//     }
//     return result;
//   },
//   resume(result: Result, scope: Scope): Result {
//     const { command, result: commandResult } = result.data as ExpandPrefixState;
//     if (!command.resume) return OK(result.value);
//     const result2 = command.resume(
//       { ...commandResult, value: result.value },
//       scope
//     );
//     if (result2.code == ResultCode.YIELD)
//       return YIELD(result2.value, { command, result: result2 });
//     return result2;
//   },
// };

// function resolveLeadingTuple(args: Value[], scope: Scope): [Command, Value[]] {
//   if (args.length == 0) return [null, null];
//   const [lead, ...rest] = args;
//   if (lead.type != ValueType.TUPLE) {
//     const command = scope.resolveCommand(lead);
//     return [command, args];
//   }
//   const tuple = lead as TupleValue;
//   return resolveLeadingTuple([...tuple.values, ...rest], scope);
// }

// export function destructureValue(
//   apply: (name: Value, value: Value, check: boolean) => Result,
//   shape: Value,
//   value: Value
// ): Result {
//   const result = checkValues(apply, shape, value);
//   if (result.code != ResultCode_OK) return result;
//   applyValues(apply, shape, value);
//   return OK(value);
// }
// function checkValues(
//   apply: (name: Value, value: Value, check: boolean) => Result,
//   shape: Value,
//   value: Value
// ): Result {
//   if (shape.type != ValueType.TUPLE) return apply(shape, value, true);
//   if (value.type != ValueType.TUPLE) return ERROR("bad value shape");
//   const variables = (shape as TupleValue).values;
//   const values = (value as TupleValue).values;
//   if (values.length < variables.length) return ERROR("bad value shape");
//   for (let i = 0; i < variables.length; i++) {
//     const result = checkValues(apply, variables[i], values[i]);
//     if (result.code != ResultCode_OK) return result;
//   }
//   return OK(NIL);
// }
// function applyValues(
//   apply: (name: Value, value: Value, check: boolean) => Result,
//   shape: Value,
//   value: Value
// ) {
//   if (shape.type != ValueType.TUPLE) {
//     apply(shape, value, false);
//     return;
//   }
//   const variables = (shape as TupleValue).values;
//   const values = (value as TupleValue).values;
//   for (let i = 0; i < variables.length; i++) {
//     applyValues(apply, variables[i], values[i]);
//   }
// }
