package helena_dialect

import "helena/core"

const SCOPE_SIGNATURE = "scope ?name? body"

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
	result := core.ValueToString(args[1])
	if result.Code != core.ResultCode_OK {
		return INVALID_SUBCOMMAND_ERROR()
	}
	subcommand := result.Data
	switch subcommand {
	case "subcommands":
		if len(args) != 2 {
			return ARITY_ERROR("<scope> subcommands")
		}
		return core.OK(scopeCommandSubcommands.List)

	case "eval":
		if len(args) != 3 {
			return ARITY_ERROR("<scope> eval body")
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
		return CreateContinuationValue(scope.scope, program, nil)

	case "call":
		if len(args) < 3 {
			return ARITY_ERROR("<scope> call cmdname ?arg ...?")
		}
		result := core.ValueToString(args[2])
		if result.Code != core.ResultCode_OK {
			return core.ERROR("invalid command name")
		}
		command := result.Data
		if !scope.scope.HasLocalCommand(command) {
			return core.ERROR(`unknown command "` + command + `"`)
		}
		cmdline := append([]core.Value{}, args[2:]...)
		program := scope.scope.CompileTupleValue(core.TUPLE(cmdline))
		return CreateContinuationValue(scope.scope, program, nil)

	default:
		return UNKNOWN_SUBCOMMAND_ERROR(subcommand)
	}
}

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

	subscope := NewScope(scope, false)
	program := subscope.CompileScriptValue(body.(core.ScriptValue))
	return CreateContinuationValue(subscope, program, func(result core.Result) core.Result {
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
