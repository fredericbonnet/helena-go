package helena_dialect

import "helena/core"

const IDEM_SIGNATURE = "idem value"

type IdemCommand struct{}

func (IdemCommand) Execute(args []core.Value, _ any) core.Result {
	if len(args) != 2 {
		return ARITY_ERROR(IDEM_SIGNATURE)
	}
	return core.OK(args[1])
}
func (IdemCommand) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	if len(args) > 2 {
		return ARITY_ERROR(IDEM_SIGNATURE)
	}
	return core.OK(core.STR(IDEM_SIGNATURE))
}

const RETURN_SIGNATURE = "return ?result?"

type ReturnCommand struct{}

func (ReturnCommand) Execute(args []core.Value, _ any) core.Result {
	if len(args) > 2 {
		return ARITY_ERROR(RETURN_SIGNATURE)
	}
	if len(args) == 2 {
		return core.RETURN(args[1])
	} else {
		return core.RETURN(core.NIL)
	}
}
func (ReturnCommand) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	if len(args) > 2 {
		return ARITY_ERROR(RETURN_SIGNATURE)
	}
	return core.OK(core.STR(RETURN_SIGNATURE))
}

const YIELD_SIGNATURE = "yield ?result?"

type YieldCommand struct{}

func (YieldCommand) Execute(args []core.Value, _ any) core.Result {
	if len(args) > 2 {
		return ARITY_ERROR(YIELD_SIGNATURE)
	}
	if len(args) == 2 {
		return core.YIELD(args[1])
	} else {
		return core.YIELD(core.NIL)
	}
}
func (YieldCommand) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	if len(args) > 2 {
		return ARITY_ERROR(YIELD_SIGNATURE)
	}
	return core.OK(core.STR(YIELD_SIGNATURE))
}

const TAILCALL_SIGNATURE = "tailcall body"

type TailcallCommand struct{}

func (TailcallCommand) Execute(args []core.Value, context any) core.Result {
	scope := context.(*Scope)
	if len(args) != 2 {
		return ARITY_ERROR(TAILCALL_SIGNATURE)
	}
	return CreateDeferredValue(core.ResultCode_RETURN, args[1], scope)
}
func (TailcallCommand) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	if len(args) > 2 {
		return ARITY_ERROR(TAILCALL_SIGNATURE)
	}
	return core.OK(core.STR(TAILCALL_SIGNATURE))
}

const ERROR_SIGNATURE = "error message"

type ErrorCommand struct{}

func (ErrorCommand) Execute(args []core.Value, _ any) core.Result {
	if len(args) != 2 {
		return ARITY_ERROR(ERROR_SIGNATURE)
	}
	// TODO accept non-string messages?
	if core.ValueToString(args[1]).Code != core.ResultCode_OK {
		return core.ERROR("invalid message")
	}
	return core.Result{
		Code:  core.ResultCode_ERROR,
		Value: args[1],
	}
}
func (ErrorCommand) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	if len(args) > 2 {
		return ARITY_ERROR(ERROR_SIGNATURE)
	}
	return core.OK(core.STR(ERROR_SIGNATURE))
}

const BREAK_SIGNATURE = "break"

type BreakCommand struct{}

func (BreakCommand) Execute(args []core.Value, _ any) core.Result {
	if len(args) != 1 {
		return ARITY_ERROR(BREAK_SIGNATURE)
	}
	return core.BREAK(core.NIL)
}
func (BreakCommand) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	if len(args) > 1 {
		return ARITY_ERROR(BREAK_SIGNATURE)
	}
	return core.OK(core.STR(BREAK_SIGNATURE))
}

const CONTINUE_SIGNATURE = "continue"

type ContinueCommand struct{}

func (ContinueCommand) Execute(args []core.Value, _ any) core.Result {
	if len(args) != 1 {
		return ARITY_ERROR(CONTINUE_SIGNATURE)
	}
	return core.CONTINUE(core.NIL)
}
func (ContinueCommand) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	if len(args) > 1 {
		return ARITY_ERROR(CONTINUE_SIGNATURE)
	}
	return core.OK(core.STR(CONTINUE_SIGNATURE))
}

const EVAL_SIGNATURE = "eval body"

type EvalCommand struct{}

func (EvalCommand) Execute(args []core.Value, context any) core.Result {
	scope := context.(*Scope)
	if len(args) != 2 {
		return ARITY_ERROR(EVAL_SIGNATURE)
	}
	return CreateDeferredValue(core.ResultCode_YIELD, args[1], scope)
}
func (EvalCommand) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	if len(args) > 2 {
		return ARITY_ERROR(EVAL_SIGNATURE)
	}
	return core.OK(core.STR(EVAL_SIGNATURE))
}

const HELP_SIGNATURE = "help command ?arg ...?"

type HelpCommand struct{}

func (HelpCommand) Execute(args []core.Value, context any) core.Result {
	scope := context.(*Scope)
	if len(args) < 2 {
		return ARITY_ERROR(HELP_SIGNATURE)
	}
	command := scope.ResolveCommand(args[1])
	if command == nil {
		result := core.ValueToString(args[1])
		if result.Code == core.ResultCode_OK {
			cmdname := result.Data
			return core.ERROR(`unknown command "` + cmdname + `"`)
		} else {
			return core.ERROR("invalid command name")
		}
	}
	if c, ok := command.(core.CommandWithHelp); ok {
		return c.Help(args[1:], core.CommandHelpOptions{}, scope)
	} else {
		result := core.ValueToString(args[1])
		if result.Code == core.ResultCode_OK {
			cmdname := result.Data
			return core.ERROR(`no help for command "` + cmdname + `"`)
		} else {
			return core.ERROR("no help for command")
		}
	}
}
func (HelpCommand) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	return core.OK(core.STR(HELP_SIGNATURE))
}

func registerBasicCommands(scope *Scope) {
	scope.RegisterNamedCommand("idem", IdemCommand{})
	scope.RegisterNamedCommand("return", ReturnCommand{})
	scope.RegisterNamedCommand("tailcall", TailcallCommand{})
	scope.RegisterNamedCommand("yield", YieldCommand{})
	scope.RegisterNamedCommand("error", ErrorCommand{})
	scope.RegisterNamedCommand("break", BreakCommand{})
	scope.RegisterNamedCommand("continue", ContinueCommand{})
	scope.RegisterNamedCommand("eval", EvalCommand{})
	scope.RegisterNamedCommand("help", HelpCommand{})
}
