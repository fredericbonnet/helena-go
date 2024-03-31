package helena_dialect

import "helena/core"

type coroutineState uint8

const (
	coroutineState_inactive coroutineState = iota
	coroutineState_active
	coroutineState_done
)

type coroutineCommand struct {
	value   core.Value
	scope   *Scope
	body    core.ScriptValue
	state   coroutineState
	process *Process
}

func newCoroutineCommand(scope *Scope, body core.ScriptValue) *coroutineCommand {
	cmd := &coroutineCommand{}
	cmd.value = core.NewCommandValue(cmd)
	cmd.scope = scope
	cmd.body = body
	cmd.state = coroutineState_inactive
	return cmd
}

var coroutineSubcommands = NewSubcommands([]string{
	"subcommands",
	"wait",
	"active",
	"done",
	"yield",
})

func (cmd *coroutineCommand) Execute(args []core.Value, _ any) core.Result {
	if len(args) == 1 {
		return core.OK(cmd.value)
	}
	return coroutineSubcommands.Dispatch(args[1], SubcommandHandlers{
		"subcommands": func() core.Result {
			if len(args) != 2 {
				return ARITY_ERROR("<coroutine> subcommands")
			}
			return core.OK(coroutineSubcommands.List)
		},
		"wait": func() core.Result {
			if len(args) != 2 {
				return ARITY_ERROR("<coroutine> wait")
			}
			if cmd.state == coroutineState_inactive {
				cmd.state = coroutineState_active
				cmd.process = cmd.scope.PrepareScriptValue(cmd.body)
			}
			return cmd.run()
		},
		"active": func() core.Result {
			if len(args) != 2 {
				return ARITY_ERROR("<coroutine> active")
			}
			return core.OK(core.BOOL(cmd.state == coroutineState_active))
		},
		"done": func() core.Result {
			if len(args) != 2 {
				return ARITY_ERROR("<coroutine> done")
			}
			return core.OK(core.BOOL(cmd.state == coroutineState_done))
		},
		"yield": func() core.Result {
			if len(args) != 2 && len(args) != 3 {
				return ARITY_ERROR("<coroutine> yield ?value?")
			}
			if cmd.state == coroutineState_inactive {
				return core.ERROR("coroutine is inactive")
			}
			if cmd.state == coroutineState_done {
				return core.ERROR("coroutine is done")
			}
			if len(args) == 3 {
				cmd.process.YieldBack(args[2])
			}
			return cmd.run()
		},
	})
}

func (cmd *coroutineCommand) run() core.Result {
	result := cmd.process.Run()
	switch result.Code {
	case core.ResultCode_OK,
		core.ResultCode_RETURN:
		cmd.state = coroutineState_done
		return core.OK(result.Value)
	case core.ResultCode_YIELD:
		return core.OK(result.Value)
	case core.ResultCode_ERROR:
		return result
	default:
		return core.ERROR("unexpected " + core.RESULT_CODE_NAME(result))
	}
}

const COROUTINE_SIGNATURE = "coroutine body"

type coroutineCmd struct{}

func (coroutineCmd) Execute(args []core.Value, context any) core.Result {
	scope := context.(*Scope)
	var body core.Value
	switch len(args) {
	case 2:
		body = args[1]
	default:
		return ARITY_ERROR(COROUTINE_SIGNATURE)
	}
	if body.Type() != core.ValueType_SCRIPT {
		return core.ERROR("body must be a script")
	}

	value := newCoroutineCommand(
		NewScope(scope, true),
		body.(core.ScriptValue),
	)
	return core.OK(value.value)
}
func (coroutineCmd) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	if len(args) > 2 {
		return ARITY_ERROR(COROUTINE_SIGNATURE)
	}
	return core.OK(core.STR(COROUTINE_SIGNATURE))
}
