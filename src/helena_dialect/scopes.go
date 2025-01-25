package helena_dialect

import "helena/core"

func SCOPE_COMMAND_PREFIX(name core.Value) string {
	return USAGE_PREFIX(name, "<scope>", core.CommandHelpOptions{})
}
func SCOPE_HELP_PREFIX(name core.Value, options core.CommandHelpOptions) string {
	return USAGE_PREFIX(name, "<scope>", options)
}

type scopeCommand struct {
	value core.Value
	scope *Scope
}

func newScopeCommand(scope *Scope) *scopeCommand {
	command := &scopeCommand{}
	command.value = core.NewCommandValue(command)
	command.scope = scope
	return command
}

var scopeCommandSubcommands = NewSubcommands([]string{
	"subcommands",
	"eval",
	"call",
})

func (scope *scopeCommand) Execute(args []core.Value, _ any) core.Result {
	if len(args) == 1 {
		return core.OK(scope.value)
	}
	result, subcommand := core.ValueToString(args[1])
	if result.Code != core.ResultCode_OK {
		return INVALID_SUBCOMMAND_ERROR()
	}
	switch subcommand {
	case "subcommands":
		if len(args) != 2 {
			return ARITY_ERROR(SCOPE_COMMAND_PREFIX(args[0]) + " subcommands")
		}
		return core.OK(scopeCommandSubcommands.List)

	case "eval":
		if len(args) != 3 {
			return ARITY_ERROR(SCOPE_COMMAND_PREFIX(args[0]) + " eval body")
		}
		body := args[2]
		var program *core.Program
		switch body.Type() {
		case core.ValueType_SCRIPT:
			program = scope.scope.CompileScriptValue(body.(core.ScriptValue))
		case core.ValueType_TUPLE:
			program = scope.scope.CompileTupleValue(body.(core.TupleValue))
		default:
			return core.ERROR("body must be a script or tuple")
		}
		return CreateContinuationValue(scope.scope, program)

	case "call":
		if len(args) < 3 {
			return ARITY_ERROR(
				SCOPE_COMMAND_PREFIX(args[0]) + " call cmdname ?arg ...?",
			)
		}
		result, command := core.ValueToString(args[2])
		if result.Code != core.ResultCode_OK {
			return core.ERROR("invalid command name")
		}
		if !scope.scope.HasLocalCommand(command) {
			return core.ERROR(`unknown command "` + command + `"`)
		}
		program := scope.scope.CompileArgs(args[2:]...)
		return CreateContinuationValue(scope.scope, program)

	default:
		return UNKNOWN_SUBCOMMAND_ERROR(subcommand)
	}
}
func (scope *scopeCommand) Help(args []core.Value, options core.CommandHelpOptions, _ any) core.Result {
	signature := SCOPE_HELP_PREFIX(args[0], options)
	if len(args) == 1 {
		return core.OK(
			core.STR(signature + " ?subcommand? ?arg ...?"),
		)
	}
	result, subcommand := core.ValueToString(args[1])
	if result.Code != core.ResultCode_OK {
		return INVALID_SUBCOMMAND_ERROR()
	}
	switch subcommand {
	case "subcommands":
		if len(args) > 2 {
			return ARITY_ERROR(signature + " subcommands")
		}
		return core.OK(core.STR(signature + " subcommands"))

	case "eval":
		if len(args) > 3 {
			return ARITY_ERROR(signature + " eval body")
		}
		return core.OK(core.STR(signature + " eval body"))

	case "call":
		return core.OK(
			core.STR(signature + " call cmdname ?arg ...?"),
		)

	default:
		return UNKNOWN_SUBCOMMAND_ERROR(subcommand)
	}
}

const SCOPE_SIGNATURE = "scope ?name? body"

type scopeCmd struct{}

func (scopeCmd) Execute(args []core.Value, context any) core.Result {
	scope := context.(*Scope)
	var name, body core.Value
	switch len(args) {
	case 2:
		body = args[1]
	case 3:
		name, body = args[1], args[2]
	default:
		return ARITY_ERROR(SCOPE_SIGNATURE)
	}
	if body.Type() != core.ValueType_SCRIPT {
		return core.ERROR("body must be a script")
	}

	subscope := scope.NewChildScope()
	program := subscope.CompileScriptValue(body.(core.ScriptValue))
	return CreateContinuationValueWithCallback(subscope, program, nil, func(result core.Result, data any) core.Result {
		switch result.Code {
		case core.ResultCode_OK,
			core.ResultCode_RETURN:
			{
				command := newScopeCommand(subscope)
				if name != nil {
					result := scope.RegisterCommand(name, command)
					if result.Code != core.ResultCode_OK {
						return result
					}
				}
				if result.Code == core.ResultCode_RETURN {
					return core.OK(result.Value)
				} else {
					return core.OK(command.value)
				}
			}
		case core.ResultCode_ERROR:
			return result
		default:
			return core.ERROR("unexpected " + core.RESULT_CODE_NAME(result))
		}
	})
}
func (scopeCmd) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	if len(args) > 3 {
		return ARITY_ERROR(SCOPE_SIGNATURE)
	}
	return core.OK(core.STR(SCOPE_SIGNATURE))
}
