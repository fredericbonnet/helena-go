package helena_dialect

import "helena/core"

type procMetacommand struct {
	value core.Value
	proc  *procCommand
}

func newProcMetacommand(proc *procCommand) *procMetacommand {
	metacommand := &procMetacommand{}
	metacommand.value = core.NewCommandValue(metacommand)
	metacommand.proc = proc
	return metacommand
}

var procMetacommandSubcommands = NewSubcommands([]string{"subcommands", "argspec"})

func (metacommand *procMetacommand) Execute(args []core.Value, _ any) core.Result {
	if len(args) == 1 {
		return core.OK(metacommand.proc.value)
	}
	result := core.ValueToString(args[1])
	if result.Code != core.ResultCode_OK {
		return INVALID_SUBCOMMAND_ERROR()
	}
	subcommand := result.Data
	switch subcommand {
	case "subcommands":
		if len(args) != 2 {
			return ARITY_ERROR("<proc> subcommands")
		}
		return core.OK(procMetacommandSubcommands.List)

	case "argspec":
		if len(args) != 2 {
			return ARITY_ERROR("<proc> argspec")
		}
		return core.OK(metacommand.proc.argspec)

	default:
		return UNKNOWN_SUBCOMMAND_ERROR(subcommand)
	}
}

func PROC_COMMAND_SIGNATURE(name core.Value, help string) string {
	if help != "" {
		return core.ValueToStringOrDefault(name, "<proc>").Data + " " + help
	} else {
		return core.ValueToStringOrDefault(name, "<proc>").Data
	}
}

type procCommand struct {
	value       core.Value
	metacommand *procMetacommand
	scope       *Scope
	argspec     ArgspecValue
	body        core.ScriptValue
	guard       core.Value
	program     *core.Program
}

func newProcCommand(
	scope *Scope,
	argspec ArgspecValue,
	body core.ScriptValue,
	guard core.Value,
	program *core.Program,
) *procCommand {
	proc := &procCommand{}
	proc.value = core.NewCommandValue(proc)
	proc.scope = scope
	proc.argspec = argspec
	proc.body = body
	proc.guard = guard
	proc.program = program
	proc.metacommand = newProcMetacommand(proc)
	return proc
}

func (proc *procCommand) Execute(args []core.Value, _ any) core.Result {
	if !proc.argspec.CheckArity(args, 1) {
		return ARITY_ERROR(PROC_COMMAND_SIGNATURE(args[0], proc.argspec.Usage(0)))
	}
	subscope := proc.scope.NewChildScope()
	setarg := func(name string, value core.Value) core.Result {
		return subscope.SetNamedVariable(name, value)
	}
	result := proc.argspec.ApplyArguments(proc.scope, args, 1, setarg)
	if result.Code != core.ResultCode_OK {
		return result
	}
	if proc.guard != nil {
		return CreateContinuationValue(subscope, proc.program, func(result core.Result) core.Result {
			switch result.Code {
			case core.ResultCode_OK,
				core.ResultCode_RETURN:
				{
					program := proc.scope.CompileArgs(proc.guard, result.Value)
					return CreateContinuationValue(proc.scope, program, nil)
				}
			case core.ResultCode_ERROR:
				return result
			default:
				return core.ERROR("unexpected " + core.RESULT_CODE_NAME(result))
			}
		})
	} else {
		return CreateContinuationValue(subscope, proc.program, func(result core.Result) core.Result {
			switch result.Code {
			case core.ResultCode_OK,
				core.ResultCode_RETURN:
				return core.OK(result.Value)
			case core.ResultCode_ERROR:
				return result
			default:
				return core.ERROR("unexpected " + core.RESULT_CODE_NAME(result))
			}
		})
	}
}
func (proc *procCommand) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	if !proc.argspec.CheckArity(args, 1) &&
		uint(len(args)) > proc.argspec.Argspec.NbRequired {
		return ARITY_ERROR(PROC_COMMAND_SIGNATURE(args[0], proc.argspec.Usage(0)))
	}
	return core.OK(core.STR(PROC_COMMAND_SIGNATURE(args[0], proc.argspec.Usage(0))))
}

const PROC_SIGNATURE = "proc ?name? argspec body"

type procCmd struct{}

func (procCmd) Execute(args []core.Value, context any) core.Result {
	scope := context.(*Scope)
	var name, specs, body core.Value
	switch len(args) {
	case 3:
		specs, body = args[1], args[2]
	case 4:
		name, specs, body = args[1], args[2], args[3]
	default:
		return ARITY_ERROR(PROC_SIGNATURE)
	}
	var guard core.Value
	if body.Type() == core.ValueType_TUPLE {
		bodySpec := body.(core.TupleValue).Values
		switch len(bodySpec) {
		case 0:
			return core.ERROR("empty body specifier")
		case 2:
			guard, body = bodySpec[0], bodySpec[1]
		default:
			return core.ERROR(`invalid body specifier`)
		}
	}
	if body.Type() != core.ValueType_SCRIPT {
		return core.ERROR("body must be a script")
	}

	result := ArgspecValueFromValue(specs)
	if result.Code != core.ResultCode_OK {
		return result.AsResult()
	}
	argspec := result.Data
	program := scope.CompileScriptValue(body.(core.ScriptValue))
	proc := newProcCommand(
		scope.NewLocalScope(),
		argspec,
		body.(core.ScriptValue),
		guard,
		program,
	)
	if name != nil {
		result := scope.RegisterCommand(name, proc)
		if result.Code != core.ResultCode_OK {
			return result
		}
	}
	return core.OK(proc.metacommand.value)
}
func (procCmd) Help(args []core.Value, options core.CommandHelpOptions, _ any) core.Result {
	if len(args) > 4 {
		return ARITY_ERROR(PROC_SIGNATURE)
	}
	return core.OK(core.STR(PROC_SIGNATURE))
}
