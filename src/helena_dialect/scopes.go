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
	return scopeCommandSubcommands.Dispatch(args[1], SubcommandHandlers{
		"subcommands": func() core.Result {
			if len(args) != 2 {
				return ARITY_ERROR("<scope> subcommands")
			}
			return core.OK(scopeCommandSubcommands.List)
		},
		"eval": func() core.Result {
			if len(args) != 3 {
				return ARITY_ERROR("<scope> eval body")
			}
			return CreateDeferredValue(core.ResultCode_YIELD, args[2], scope.scope)
		},
		"call": func() core.Result {
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
			cmdline := args[2:]
			return CreateDeferredValue(
				core.ResultCode_YIELD,
				core.TUPLE(cmdline),
				scope.scope,
			)
		},
	})
}

type scopeBodyState struct {
	scope    *Scope
	subscope *Scope
	process  *Process
	name     core.Value
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
	process := subscope.PrepareScriptValue(body.(core.ScriptValue))
	return executeScopeBody(&scopeBodyState{scope, subscope, process, name})
}
func (scopeCmd) Resume(result core.Result, _ any) core.Result {
	state := result.Data.(*scopeBodyState)
	state.process.YieldBack(result.Value)
	return executeScopeBody(state)
}
func (scopeCmd) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	if len(args) > 3 {
		return ARITY_ERROR(SCOPE_SIGNATURE)
	}
	return core.OK(core.STR(SCOPE_SIGNATURE))
}
func executeScopeBody(state *scopeBodyState) core.Result {
	result := state.process.Run()
	switch result.Code {
	case core.ResultCode_OK,
		core.ResultCode_RETURN:
		{
			command := newScopeCommand(state.subscope)
			if state.name != nil {
				result := state.scope.RegisterCommand(state.name, command)
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
	case core.ResultCode_YIELD:
		return core.YIELD_STATE(result.Value, state)
	case core.ResultCode_ERROR:
		return result
	default:
		return core.ERROR("unexpected " + core.RESULT_CODE_NAME(result.Code))
	}
}
