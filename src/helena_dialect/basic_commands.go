package helena_dialect

import "helena/core"

const IDEM_SIGNATURE = "idem value"

type idemCmd struct{}

func (idemCmd) Execute(args []core.Value, _ any) core.Result {
	if len(args) != 2 {
		return ARITY_ERROR(IDEM_SIGNATURE)
	}
	return core.OK(args[1])
}
func (idemCmd) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	if len(args) > 2 {
		return ARITY_ERROR(IDEM_SIGNATURE)
	}
	return core.OK(core.STR(IDEM_SIGNATURE))
}

const RETURN_SIGNATURE = "return ?result?"

type returnCmd struct{}

func (returnCmd) Execute(args []core.Value, _ any) core.Result {
	if len(args) > 2 {
		return ARITY_ERROR(RETURN_SIGNATURE)
	}
	if len(args) == 2 {
		return core.RETURN(args[1])
	} else {
		return core.RETURN(core.NIL)
	}
}
func (returnCmd) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	if len(args) > 2 {
		return ARITY_ERROR(RETURN_SIGNATURE)
	}
	return core.OK(core.STR(RETURN_SIGNATURE))
}

const YIELD_SIGNATURE = "yield ?result?"

type yieldCmd struct{}

func (yieldCmd) Execute(args []core.Value, _ any) core.Result {
	if len(args) > 2 {
		return ARITY_ERROR(YIELD_SIGNATURE)
	}
	if len(args) == 2 {
		return core.YIELD(args[1])
	} else {
		return core.YIELD(core.NIL)
	}
}
func (yieldCmd) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	if len(args) > 2 {
		return ARITY_ERROR(YIELD_SIGNATURE)
	}
	return core.OK(core.STR(YIELD_SIGNATURE))
}

const ERROR_SIGNATURE = "error message"

type errorCmd struct{}

func (errorCmd) Execute(args []core.Value, _ any) core.Result {
	if len(args) != 2 {
		return ARITY_ERROR(ERROR_SIGNATURE)
	}
	// TODO accept non-string messages?
	if result, _ := core.ValueToString(args[1]); result.Code != core.ResultCode_OK {
		return core.ERROR("invalid message")
	}
	return core.Result{
		Code:  core.ResultCode_ERROR,
		Value: args[1],
	}
}
func (errorCmd) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	if len(args) > 2 {
		return ARITY_ERROR(ERROR_SIGNATURE)
	}
	return core.OK(core.STR(ERROR_SIGNATURE))
}

const BREAK_SIGNATURE = "break"

type breakCmd struct{}

func (breakCmd) Execute(args []core.Value, _ any) core.Result {
	if len(args) != 1 {
		return ARITY_ERROR(BREAK_SIGNATURE)
	}
	return core.BREAK(core.NIL)
}
func (breakCmd) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	if len(args) > 1 {
		return ARITY_ERROR(BREAK_SIGNATURE)
	}
	return core.OK(core.STR(BREAK_SIGNATURE))
}

const CONTINUE_SIGNATURE = "continue"

type continueCmd struct{}

func (continueCmd) Execute(args []core.Value, _ any) core.Result {
	if len(args) != 1 {
		return ARITY_ERROR(CONTINUE_SIGNATURE)
	}
	return core.CONTINUE(core.NIL)
}
func (continueCmd) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	if len(args) > 1 {
		return ARITY_ERROR(CONTINUE_SIGNATURE)
	}
	return core.OK(core.STR(CONTINUE_SIGNATURE))
}

const EVAL_SIGNATURE = "eval body"

type evalCmd struct{}

func (evalCmd) Execute(args []core.Value, context any) core.Result {
	scope := context.(*Scope)
	if len(args) != 2 {
		return ARITY_ERROR(EVAL_SIGNATURE)
	}
	body := args[1]
	var program *core.Program
	switch body.Type() {
	case core.ValueType_SCRIPT:
		program = scope.CompileScriptValue(body.(core.ScriptValue))
	case core.ValueType_TUPLE:
		program = scope.CompileTupleValue(body.(core.TupleValue))
	default:
		return core.ERROR("body must be a script or tuple")
	}
	return CreateContinuationValue(scope, program)
}
func (evalCmd) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	if len(args) > 2 {
		return ARITY_ERROR(EVAL_SIGNATURE)
	}
	return core.OK(core.STR(EVAL_SIGNATURE))
}

const HELP_SIGNATURE = "help command ?arg ...?"

type helpCmd struct{}

func (helpCmd) Execute(args []core.Value, context any) core.Result {
	scope := context.(*Scope)
	if len(args) < 2 {
		return ARITY_ERROR(HELP_SIGNATURE)
	}
	command := scope.ResolveCommand(args[1])
	if command == nil {
		result, cmdname := core.ValueToString(args[1])
		if result.Code == core.ResultCode_OK {
			return core.ERROR(`unknown command "` + cmdname + `"`)
		} else {
			return core.ERROR("invalid command name")
		}
	}
	if c, ok := command.(core.CommandWithHelp); ok {
		return c.Help(args[1:], core.CommandHelpOptions{}, scope)
	} else {
		result, cmdname := core.ValueToString(args[1])
		if result.Code == core.ResultCode_OK {
			return core.ERROR(`no help for command "` + cmdname + `"`)
		} else {
			return core.ERROR("no help for command")
		}
	}
}
func (helpCmd) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	return core.OK(core.STR(HELP_SIGNATURE))
}

func registerBasicCommands(scope *Scope) {
	scope.RegisterNamedCommand("idem", idemCmd{})
	scope.RegisterNamedCommand("return", returnCmd{})
	scope.RegisterNamedCommand("yield", yieldCmd{})
	scope.RegisterNamedCommand("error", errorCmd{})
	scope.RegisterNamedCommand("break", breakCmd{})
	scope.RegisterNamedCommand("continue", continueCmd{})
	scope.RegisterNamedCommand("eval", evalCmd{})
	scope.RegisterNamedCommand("help", helpCmd{})
	scope.RegisterNamedCommand("^", core.LAST_RESULT)
	scope.RegisterNamedCommand("|>", core.SHIFT_LAST_FRAME_RESULT)
}
