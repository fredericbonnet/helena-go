package helena_dialect

import (
	"helena/core"
	"sync"
)

var booleanSubcommands = NewSubcommands([]string{"subcommands", "?", "!?"})

type trueCmd struct{}

func (trueCmd) Execute(args []core.Value, _ any) core.Result {
	if len(args) == 1 {
		return core.OK(core.TRUE)
	}
	result, subcommand := core.ValueToString(args[1])
	if result.Code != core.ResultCode_OK {
		return INVALID_SUBCOMMAND_ERROR()
	}
	switch subcommand {
	case "subcommands":
		if len(args) != 2 {
			return ARITY_ERROR("true subcommands")
		}
		return core.OK(booleanSubcommands.List)

	case "?":
		if len(args) < 3 || len(args) > 4 {
			return ARITY_ERROR("true ? arg ?arg?")
		}
		return core.OK(args[2])

	case "!?":
		if len(args) < 3 || len(args) > 4 {
			return ARITY_ERROR("true !? arg ?arg?")
		}
		if len(args) == 4 {
			return core.OK(args[3])
		} else {
			return core.OK(core.NIL)
		}

	default:
		return UNKNOWN_SUBCOMMAND_ERROR(subcommand)
	}
}
func (trueCmd) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	if len(args) == 1 {
		return core.OK(core.STR("true ?subcommand?"))
	}
	result, subcommand := core.ValueToString(args[1])
	if result.Code != core.ResultCode_OK {
		return INVALID_SUBCOMMAND_ERROR()
	}
	switch subcommand {
	case "subcommands":
		if len(args) > 2 {
			return ARITY_ERROR("true subcommands")
		}
		return core.OK(core.STR("true subcommands"))

	case "?":
		if len(args) > 4 {
			return ARITY_ERROR("true ? arg ?arg?")
		}
		return core.OK(core.STR("true ? arg ?arg?"))

	case "!?":
		if len(args) > 4 {
			return ARITY_ERROR("true !? arg ?arg?")
		}
		return core.OK(core.STR("true !? arg ?arg?"))

	default:
		return UNKNOWN_SUBCOMMAND_ERROR(subcommand)
	}
}

type falseCmd struct{}

func (falseCmd) Execute(args []core.Value, _ any) core.Result {
	if len(args) == 1 {
		return core.OK(core.FALSE)
	}
	result, subcommand := core.ValueToString(args[1])
	if result.Code != core.ResultCode_OK {
		return INVALID_SUBCOMMAND_ERROR()
	}
	switch subcommand {
	case "subcommands":
		if len(args) != 2 {
			return ARITY_ERROR("false subcommands")
		}
		return core.OK(booleanSubcommands.List)

	case "?":
		if len(args) < 3 || len(args) > 4 {
			return ARITY_ERROR("false ? arg ?arg?")
		}
		if len(args) == 4 {
			return core.OK(args[3])
		} else {
			return core.OK(core.NIL)
		}

	case "!?":
		if len(args) < 3 || len(args) > 4 {
			return ARITY_ERROR("false !? arg ?arg?")
		}
		return core.OK(args[2])

	default:
		return UNKNOWN_SUBCOMMAND_ERROR(subcommand)
	}
}
func (falseCmd) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	if len(args) == 1 {
		return core.OK(core.STR("false ?subcommand?"))
	}
	result, subcommand := core.ValueToString(args[1])
	if result.Code != core.ResultCode_OK {
		return INVALID_SUBCOMMAND_ERROR()
	}
	switch subcommand {
	case "subcommands":
		if len(args) > 2 {
			return ARITY_ERROR("false subcommands")
		}
		return core.OK(core.STR("false subcommands"))

	case "?":
		if len(args) > 4 {
			return ARITY_ERROR("false ? arg ?arg?")
		}
		return core.OK(core.STR("false ? arg ?arg?"))

	case "!?":
		if len(args) > 4 {
			return ARITY_ERROR("false !? arg ?arg?")
		}
		return core.OK(core.STR("false !? arg ?arg?"))

	default:
		return UNKNOWN_SUBCOMMAND_ERROR(subcommand)
	}
}

type boolCommand struct {
	scope    *Scope
	ensemble *EnsembleCommand
}

func newBoolCommand(scope *Scope) *boolCommand {
	cmd := &boolCommand{}
	cmd.scope = scope.NewChildScope()
	_, argspec := ArgspecValueFromValue(core.LIST([]core.Value{core.STR("value")}))
	cmd.ensemble = NewEnsembleCommand(cmd.scope, argspec)
	return cmd
}
func (cmd *boolCommand) Execute(args []core.Value, context any) core.Result {
	if len(args) == 2 {
		result, _ := core.BooleanValueFromValue(args[1])
		return result
	}
	return cmd.ensemble.Execute(args, context)
}
func (cmd *boolCommand) Help(args []core.Value, options core.CommandHelpOptions, context any) core.Result {
	return cmd.ensemble.Help(args, options, context)
}

const NOT_SIGNATURE = "! arg"

type notCmd struct{}

func (notCmd) Execute(args []core.Value, context any) core.Result {
	scope := context.(*Scope)
	if len(args) != 2 {
		return ARITY_ERROR(NOT_SIGNATURE)
	}
	return notCmdCondition(scope, args[1])
}
func (notCmd) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	if len(args) > 2 {
		return ARITY_ERROR(NOT_SIGNATURE)
	}
	return core.OK(core.STR(NOT_SIGNATURE))
}
func notCmdCondition(
	scope *Scope,
	value core.Value,
) core.Result {
	if value.Type() == core.ValueType_SCRIPT {
		program := scope.CompileScriptValue(value.(core.ScriptValue))
		return CreateContinuationValueWithCallback(scope, program, nil, func(result core.Result, data any) core.Result {
			if result.Code != core.ResultCode_OK {
				return result
			}
			return notCmdResult(result.Value)
		})
	}
	// TODO ensure tail call in trampoline, or unroll in caller
	return notCmdResult(value)
}
func notCmdResult(value core.Value) core.Result {
	result, b := core.ValueToBoolean(value)
	if result.Code != core.ResultCode_OK {
		return result
	}
	if b {
		return core.OK(core.FALSE)
	} else {
		return core.OK(core.TRUE)
	}
}

const AND_SIGNATURE = "&& arg ?arg ...?"

type andCmd struct{}
type andCmdState struct {
	scope *Scope
	i     int
	args  []core.Value
}

var andCmdStatePool = sync.Pool{
	New: func() any {
		return &andCmdState{}
	},
}

func (andCmd) Execute(args []core.Value, context any) core.Result {
	scope := context.(*Scope)
	if len(args) < 2 {
		return ARITY_ERROR(AND_SIGNATURE)
	}
	return andCmdCondition(scope, args, 1)
}
func (andCmd) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	return core.OK(core.STR(AND_SIGNATURE))
}
func andCmdCondition(scope *Scope, args []core.Value, i int) core.Result {
	value := args[i]
	if value.Type() == core.ValueType_SCRIPT {
		program := scope.CompileScriptValue(value.(core.ScriptValue))
		state := andCmdStatePool.Get().(*andCmdState)
		state.scope = scope
		state.i = i
		state.args = args
		return CreateContinuationValueWithCallback(scope, program, state, func(result core.Result, data any) core.Result {
			state := data.(*andCmdState)
			andCmdStatePool.Put(state)
			if result.Code != core.ResultCode_OK {
				return result
			}
			return andCmdNext(result.Value, state.scope, state.args, state.i)
		})
	}
	// TODO ensure tail call in trampoline, or unroll in caller
	return andCmdNext(value, scope, args, i)
}
func andCmdNext(value core.Value, scope *Scope, args []core.Value, i int) core.Result {
	result, b := core.ValueToBoolean(value)
	if result.Code != core.ResultCode_OK {
		return result
	}
	if !b {
		return core.OK(core.FALSE)
	}
	i++
	if i >= len(args) {
		return core.OK(core.TRUE)
	}
	return andCmdCondition(scope, args, i)
}

const OR_SIGNATURE = "|| arg ?arg ...?"

type orCmd struct{}

func (orCmd) Execute(args []core.Value, context any) core.Result {
	scope := context.(*Scope)
	if len(args) < 2 {
		return ARITY_ERROR(OR_SIGNATURE)
	}
	return orCmdCondition(scope, args, 1)
}
func (orCmd) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	return core.OK(core.STR(OR_SIGNATURE))
}
func orCmdCondition(scope *Scope, args []core.Value, i int) core.Result {
	value := args[i]
	if value.Type() == core.ValueType_SCRIPT {
		program := scope.CompileScriptValue(value.(core.ScriptValue))
		state := andCmdStatePool.Get().(*andCmdState)
		state.scope = scope
		state.i = i
		state.args = args
		return CreateContinuationValueWithCallback(scope, program, state, func(result core.Result, data any) core.Result {
			state := data.(*andCmdState)
			andCmdStatePool.Put(state)
			if result.Code != core.ResultCode_OK {
				return result
			}
			return orCmdNext(result.Value, state.scope, state.args, state.i)
		})
	}
	// TODO ensure tail call in trampoline, or unroll in caller
	return orCmdNext(value, scope, args, i)
}
func orCmdNext(value core.Value, scope *Scope, args []core.Value, i int) core.Result {
	result, b := core.ValueToBoolean(value)
	if result.Code != core.ResultCode_OK {
		return result
	}
	if b {
		return core.OK(core.TRUE)
	}
	i++
	if i >= len(args) {
		return core.OK(core.FALSE)
	}
	return orCmdCondition(scope, args, i)
}

func registerLogicCommands(scope *Scope) {
	scope.RegisterNamedCommand("true", trueCmd{})
	scope.RegisterNamedCommand("false", falseCmd{})

	boolCommand := newBoolCommand(scope)
	scope.RegisterNamedCommand("bool", boolCommand)

	scope.RegisterNamedCommand("!", notCmd{})
	scope.RegisterNamedCommand("&&", andCmd{})
	scope.RegisterNamedCommand("||", orCmd{})
}
