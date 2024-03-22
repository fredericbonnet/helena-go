package helena_dialect

import "helena/core"

var booleanSubcommands = NewSubcommands([]string{"subcommands", "?", "!?"})

type trueCmd struct{}

func (trueCmd) Execute(args []core.Value, _ any) core.Result {
	if len(args) == 1 {
		return core.OK(core.TRUE)
	}
	return booleanSubcommands.Dispatch(args[1], SubcommandHandlers{
		"subcommands": func() core.Result {
			if len(args) != 2 {
				return ARITY_ERROR("true subcommands")
			}
			return core.OK(booleanSubcommands.List)
		},
		"?": func() core.Result {
			if len(args) < 3 || len(args) > 4 {
				return ARITY_ERROR("true ? arg ?arg?")
			}
			return core.OK(args[2])
		},
		"!?": func() core.Result {
			if len(args) < 3 || len(args) > 4 {
				return ARITY_ERROR("true !? arg ?arg?")
			}
			if len(args) == 4 {
				return core.OK(args[3])
			} else {
				return core.OK(core.NIL)
			}
		},
	})
}
func (trueCmd) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	if len(args) == 1 {
		return core.OK(core.STR("true ?subcommand?"))
	}
	return booleanSubcommands.Dispatch(args[1], SubcommandHandlers{
		"subcommands": func() core.Result {
			if len(args) > 2 {
				return ARITY_ERROR("true subcommands")
			}
			return core.OK(core.STR("true subcommands"))
		},
		"?": func() core.Result {
			if len(args) > 4 {
				return ARITY_ERROR("true ? arg ?arg?")
			}
			return core.OK(core.STR("true ? arg ?arg?"))
		},
		"!?": func() core.Result {
			if len(args) > 4 {
				return ARITY_ERROR("true !? arg ?arg?")
			}
			return core.OK(core.STR("true !? arg ?arg?"))
		},
	})
}

type falseCmd struct{}

func (falseCmd) Execute(args []core.Value, _ any) core.Result {
	if len(args) == 1 {
		return core.OK(core.FALSE)
	}
	return booleanSubcommands.Dispatch(args[1], SubcommandHandlers{
		"subcommands": func() core.Result {
			if len(args) != 2 {
				return ARITY_ERROR("false subcommands")
			}
			return core.OK(booleanSubcommands.List)
		},
		"?": func() core.Result {
			if len(args) < 3 || len(args) > 4 {
				return ARITY_ERROR("false ? arg ?arg?")
			}
			if len(args) == 4 {
				return core.OK(args[3])
			} else {
				return core.OK(core.NIL)
			}
		},
		"!?": func() core.Result {
			if len(args) < 3 || len(args) > 4 {
				return ARITY_ERROR("false !? arg ?arg?")
			}
			return core.OK(args[2])
		},
	})
}
func (falseCmd) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	if len(args) == 1 {
		return core.OK(core.STR("false ?subcommand?"))
	}
	return booleanSubcommands.Dispatch(args[1], SubcommandHandlers{
		"subcommands": func() core.Result {
			if len(args) > 2 {
				return ARITY_ERROR("false subcommands")
			}
			return core.OK(core.STR("false subcommands"))
		},
		"?": func() core.Result {
			if len(args) > 4 {
				return ARITY_ERROR("false ? arg ?arg?")
			}
			return core.OK(core.STR("false ? arg ?arg?"))
		},
		"!?": func() core.Result {
			if len(args) > 4 {
				return ARITY_ERROR("false !? arg ?arg?")
			}
			return core.OK(core.STR("false !? arg ?arg?"))
		},
	})
}

type boolCommand struct {
	scope    *Scope
	ensemble *EnsembleCommand
}

func newBoolCommand(scope *Scope) *boolCommand {
	cmd := &boolCommand{}
	cmd.scope = NewScope(scope, false)
	argspec := ArgspecValueFromValue(core.LIST([]core.Value{core.STR("value")})).Data
	cmd.ensemble = NewEnsembleCommand(cmd.scope, argspec)
	return cmd
}
func (cmd *boolCommand) Execute(args []core.Value, context any) core.Result {
	if len(args) == 2 {
		return core.BooleanValueFromValue(args[1]).AsResult()
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
	result := ExecuteCondition(scope, args[1])
	if result.Code != core.ResultCode_OK {
		return result
	}
	if result.Value.(core.BooleanValue).Value {
		return core.OK(core.FALSE)
	} else {
		return core.OK(core.TRUE)
	}
}
func (notCmd) Resume(result core.Result, _ any) core.Result {
	result = ResumeCondition(result)
	if result.Code != core.ResultCode_OK {
		return result
	}
	if result.Value.(core.BooleanValue).Value {
		return core.OK(core.FALSE)
	} else {
		return core.OK(core.TRUE)
	}
}
func (notCmd) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	if len(args) > 2 {
		return ARITY_ERROR(NOT_SIGNATURE)
	}
	return core.OK(core.STR(NOT_SIGNATURE))
}

const AND_SIGNATURE = "&& arg ?arg ...?"

type andCommandState struct {
	args   []core.Value
	i      int
	result *core.Result
}
type andCmd struct{}

func (cmd andCmd) Execute(args []core.Value, context any) core.Result {
	scope := context.(*Scope)
	if len(args) < 2 {
		return ARITY_ERROR(AND_SIGNATURE)
	}
	return cmd.run(&andCommandState{args: args, i: 1}, scope)
}
func (cmd andCmd) Resume(result core.Result, context any) core.Result {
	scope := context.(*Scope)
	state := result.Data.(*andCommandState)
	state.result = &core.Result{Code: state.result.Code, Value: result.Value, Data: state.result.Data}
	return cmd.run(result.Data.(*andCommandState), scope)
}
func (andCmd) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	return core.OK(core.STR(AND_SIGNATURE))
}
func (cmd andCmd) run(state *andCommandState, scope *Scope) core.Result {
	r := core.TRUE
	for state.i < len(state.args) {
		var result core.Result
		if state.result != nil {
			result = ResumeCondition(*state.result)
		} else {
			result = ExecuteCondition(scope, state.args[state.i])
		}
		state.result = &result
		if state.result.Code == core.ResultCode_YIELD {
			return core.YIELD_STATE(state.result.Value, state)
		}
		if state.result.Code != core.ResultCode_OK {
			return *state.result
		}
		if !(state.result.Value.(core.BooleanValue).Value) {
			r = core.FALSE
			break
		}
		state.result = nil
		state.i++
	}

	return core.OK(r)
}

const OR_SIGNATURE = "|| arg ?arg ...?"

type orCommandState struct {
	args   []core.Value
	i      int
	result *core.Result
}
type orCmd struct{}

func (cmd orCmd) Execute(args []core.Value, context any) core.Result {
	scope := context.(*Scope)
	if len(args) < 2 {
		return ARITY_ERROR(OR_SIGNATURE)
	}
	return cmd.run(&orCommandState{args: args, i: 1}, scope)
}
func (cmd orCmd) Resume(result core.Result, context any) core.Result {
	scope := context.(*Scope)
	state := result.Data.(*orCommandState)
	state.result = &core.Result{Code: state.result.Code, Value: result.Value, Data: state.result.Data}
	return cmd.run(result.Data.(*orCommandState), scope)
}
func (orCmd) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	return core.OK(core.STR(OR_SIGNATURE))
}
func (orCmd) run(state *orCommandState, scope *Scope) core.Result {
	r := core.FALSE
	for state.i < len(state.args) {
		var result core.Result
		if state.result != nil {
			result = ResumeCondition(*state.result)
		} else {
			result = ExecuteCondition(scope, state.args[state.i])
		}
		state.result = &result
		if state.result.Code == core.ResultCode_YIELD {
			return core.YIELD_STATE(state.result.Value, state)
		}
		if state.result.Code != core.ResultCode_OK {
			return *state.result
		}
		if state.result.Value.(core.BooleanValue).Value {
			r = core.TRUE
			break
		}
		state.result = nil
		state.i++
	}

	return core.OK(r)
}

func ExecuteCondition(scope *Scope, value core.Value) core.Result {
	if value.Type() == core.ValueType_SCRIPT {
		process := scope.PrepareScriptValue(value.(core.ScriptValue))
		return runCondition(process)
	}
	return core.BooleanValueFromValue(value).AsResult()
}
func ResumeCondition(result core.Result) core.Result {
	process := result.Data.(*Process)
	process.YieldBack(result.Value)
	return runCondition(process)
}
func runCondition(process *Process) core.Result {
	result := process.Run()
	if result.Code == core.ResultCode_YIELD {
		return core.YIELD_STATE(result.Value, process)
	}
	if result.Code != core.ResultCode_OK {
		return result
	}
	return core.BooleanValueFromValue(result.Value).AsResult()
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
