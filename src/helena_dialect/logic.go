package helena_dialect

import "helena/core"

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
	return ExecuteCondition(scope, args[1], nil, func(result core.Result, b bool, data any) core.Result {
		if result.Code != core.ResultCode_OK {
			return result
		}
		if b {
			return core.OK(core.FALSE)
		} else {
			return core.OK(core.TRUE)
		}
	})
}
func (notCmd) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	if len(args) > 2 {
		return ARITY_ERROR(NOT_SIGNATURE)
	}
	return core.OK(core.STR(NOT_SIGNATURE))
}

const AND_SIGNATURE = "&& arg ?arg ...?"

type andCmd struct{}

func (andCmd) Execute(args []core.Value, context any) core.Result {
	scope := context.(*Scope)
	if len(args) < 2 {
		return ARITY_ERROR(AND_SIGNATURE)
	}
	i := 1
	var callback func(result core.Result, b bool, data any) core.Result
	callback = func(result core.Result, b bool, d any) core.Result {
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
		return ExecuteCondition(scope, args[i], nil, callback)
	}
	return ExecuteCondition(scope, args[i], nil, callback)
}
func (andCmd) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	return core.OK(core.STR(AND_SIGNATURE))
}

const OR_SIGNATURE = "|| arg ?arg ...?"

type orCmd struct{}

func (orCmd) Execute(args []core.Value, context any) core.Result {
	scope := context.(*Scope)
	if len(args) < 2 {
		return ARITY_ERROR(OR_SIGNATURE)
	}
	i := 1
	var callback func(result core.Result, b bool, data any) core.Result
	callback = func(result core.Result, b bool, data any) core.Result {
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
		return ExecuteCondition(scope, args[i], nil, callback)
	}
	return ExecuteCondition(scope, args[i], nil, callback)
}
func (orCmd) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	return core.OK(core.STR(OR_SIGNATURE))
}

func ExecuteCondition(
	scope *Scope,
	value core.Value,
	data any,
	callback func(result core.Result, b bool, data any) core.Result,
) core.Result {
	if value.Type() == core.ValueType_SCRIPT {
		program := scope.CompileScriptValue(value.(core.ScriptValue))
		return CreateContinuationValueWithCallback(scope, program, data, func(result core.Result, data any) core.Result {
			if result.Code != core.ResultCode_OK {
				return result
			}
			r, b := core.ValueToBoolean(result.Value)
			return callback(r, b, data)
		})
	}
	// TODO ensure tail call in trampoline, or unroll in caller
	r, b := core.ValueToBoolean(value)
	return callback(r, b, data)
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
