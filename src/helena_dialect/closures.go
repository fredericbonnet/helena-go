package helena_dialect

import "helena/core"

type closureMetacommand struct {
	value   core.Value
	closure *closureCommand
}

func newClosureMetacommand(closure *closureCommand) *closureMetacommand {
	metacommand := &closureMetacommand{}
	metacommand.value = core.NewCommandValue(metacommand)
	metacommand.closure = closure
	return metacommand
}

var closureMetacommandSubcommands = NewSubcommands([]string{"subcommands", "argspec"})

func (metacommand *closureMetacommand) Execute(args []core.Value, _ any) core.Result {
	if len(args) == 1 {
		return core.OK(metacommand.closure.value)
	}
	result, subcommand := core.ValueToString(args[1])
	if result.Code != core.ResultCode_OK {
		return INVALID_SUBCOMMAND_ERROR()
	}
	switch subcommand {
	case "subcommands":
		if len(args) != 2 {
			return ARITY_ERROR("<metacommand> subcommands")
		}
		return core.OK(closureMetacommandSubcommands.List)

	case "argspec":
		if len(args) != 2 {
			return ARITY_ERROR("<metacommand> argspec")
		}
		return core.OK(metacommand.closure.argspec)

	default:
		return UNKNOWN_SUBCOMMAND_ERROR(subcommand)
	}
}
func (*closureMetacommand) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	if len(args) == 1 {
		return core.OK(core.STR("<metacommand> ?subcommand? ?arg ...?"))
	}
	result, subcommand := core.ValueToString(args[1])
	if result.Code != core.ResultCode_OK {
		return INVALID_SUBCOMMAND_ERROR()
	}
	switch subcommand {
	case "subcommands":
		if len(args) > 2 {
			return ARITY_ERROR("<metacommand> subcommands")
		}
		return core.OK(core.STR("<metacommand> subcommands"))

	case "argspec":
		if len(args) > 2 {
			return ARITY_ERROR("<metacommand> argspec")
		}
		return core.OK(core.STR("<metacommand> argspec"))

	default:
		return UNKNOWN_SUBCOMMAND_ERROR(subcommand)
	}
}

func CLOSURE_COMMAND_SIGNATURE(name core.Value, help string) string {
	_, s := core.ValueToStringOrDefault(name, "<closure>")
	if help != "" {
		return s + " " + help
	} else {
		return s
	}
}

type closureCommand struct {
	value       core.Value
	metacommand *closureMetacommand
	scope       *Scope
	argspec     ArgspecValue
	body        core.ScriptValue
	guard       core.Value
}

func newClosureCommand(
	scope *Scope,
	argspec ArgspecValue,
	body core.ScriptValue,
	guard core.Value,
) *closureCommand {
	closure := &closureCommand{}
	closure.value = core.NewCommandValue(closure)
	closure.scope = scope
	closure.argspec = argspec
	closure.body = body
	closure.guard = guard
	closure.metacommand = newClosureMetacommand(closure)
	return closure
}

func (closure *closureCommand) Execute(args []core.Value, _ any) core.Result {
	if !closure.argspec.CheckArity(args, 1) {
		return ARITY_ERROR(
			CLOSURE_COMMAND_SIGNATURE(args[0], closure.argspec.Usage(0)),
		)
	}
	subscope := closure.scope.NewLocalScope()
	setarg := func(name string, value core.Value) core.Result {
		subscope.SetNamedLocal(name, value)
		return core.OK(value)
	}
	result := closure.argspec.ApplyArguments(closure.scope, args, 1, setarg)
	if result.Code != core.ResultCode_OK {
		return result
	}
	program := subscope.CompileScriptValue(closure.body)
	if closure.guard != nil {
		return CreateContinuationValue(subscope, program, func(result core.Result) core.Result {
			if result.Code != core.ResultCode_OK {
				return result
			}
			program := closure.scope.CompileArgs(closure.guard, result.Value)
			return CreateContinuationValue(closure.scope, program, nil)
		})
	} else {
		return CreateContinuationValue(subscope, program, nil)
	}
}
func (closure *closureCommand) Help(args []core.Value, options core.CommandHelpOptions, _ any) core.Result {
	var usage string
	if options.Skip > 0 {
		usage = closure.argspec.Usage(options.Skip - 1)
	} else {
		usage = CLOSURE_COMMAND_SIGNATURE(args[0], closure.argspec.Usage(0))
	}
	signature := ""
	if len(options.Prefix) > 0 {
		signature += options.Prefix
	}
	if len(usage) > 0 {
		if len(signature) > 0 {
			signature += " "
		}
		signature += usage
	}
	if !closure.argspec.CheckArity(args, 1) &&
		uint(len(args)) > closure.argspec.Argspec.NbRequired {
		return ARITY_ERROR(signature)
	}
	return core.OK(core.STR(signature))
}

const CLOSURE_SIGNATURE = "closure ?name? argspec body"

type closureCmd struct{}

func (closureCmd) Execute(args []core.Value, context any) core.Result {
	scope := context.(*Scope)
	var name, specs, body core.Value
	switch len(args) {
	case 3:
		specs, body = args[1], args[2]
	case 4:
		name, specs, body = args[1], args[2], args[3]
	default:
		return ARITY_ERROR(CLOSURE_SIGNATURE)
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

	result, argspec := ArgspecValueFromValue(specs)
	if result.Code != core.ResultCode_OK {
		return result
	}
	closure := newClosureCommand(
		scope.NewLocalScope(),
		argspec,
		body.(core.ScriptValue),
		guard,
	)
	if name != nil {
		result := scope.RegisterCommand(name, closure)
		if result.Code != core.ResultCode_OK {
			return result
		}
	}
	return core.OK(closure.metacommand.value)
}
func (closureCmd) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	if len(args) > 4 {
		return ARITY_ERROR(CLOSURE_SIGNATURE)
	}
	return core.OK(core.STR(CLOSURE_SIGNATURE))
}
