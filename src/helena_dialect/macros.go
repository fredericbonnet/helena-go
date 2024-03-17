package helena_dialect

import "helena/core"

type macroMetacommand struct {
	value core.Value
	macro *macroCommand
}

func newMacroMetacommand(macro *macroCommand) *macroMetacommand {
	metacommand := &macroMetacommand{}
	metacommand.value = core.NewCommandValue(metacommand)
	metacommand.macro = macro
	return metacommand
}

var macroMetacommandSubcommands = NewSubcommands([]string{"subcommands", "argspec"})

func (metacommand *macroMetacommand) Execute(args []core.Value, _ any) core.Result {
	if len(args) == 1 {
		return core.OK(metacommand.macro.value)
	}
	return macroMetacommandSubcommands.Dispatch(args[1], SubcommandHandlers{
		"subcommands": func() core.Result {
			if len(args) != 2 {
				return ARITY_ERROR("<macro> subcommands")
			}
			return core.OK(macroMetacommandSubcommands.List)
		},
		"argspec": func() core.Result {
			if len(args) != 2 {
				return ARITY_ERROR("<macro> argspec")
			}
			return core.OK(metacommand.macro.argspec)
		},
	})
}

func MACRO_COMMAND_SIGNATURE(name core.Value, help string) string {
	if help != "" {
		return core.ValueToStringOrDefault(name, "<macro>").Data + " " + help
	} else {
		return core.ValueToStringOrDefault(name, "<macro>").Data
	}
}

type macroCommand struct {
	value       core.Value
	metacommand *macroMetacommand
	argspec     ArgspecValue
	body        core.ScriptValue
	guard       core.Value
}

func newMacroCommand(argspec ArgspecValue, body core.ScriptValue, guard core.Value) *macroCommand {
	macro := &macroCommand{}
	macro.value = core.NewCommandValue(macro)
	macro.argspec = argspec
	macro.body = body
	macro.guard = guard
	macro.metacommand = newMacroMetacommand(macro)
	return macro
}

func (macro *macroCommand) Execute(args []core.Value, context any) core.Result {
	scope := context.(*Scope)
	if !macro.argspec.CheckArity(args, 1) {
		return ARITY_ERROR(
			MACRO_COMMAND_SIGNATURE(args[0], macro.argspec.Usage(0)),
		)
	}
	subscope := NewScope(scope, true)
	setarg := func(name string, value core.Value) core.Result {
		subscope.SetNamedLocal(name, value)
		return core.OK(value)
	}
	// TODO handle YIELD?
	result := macro.argspec.ApplyArguments(scope, args, 1, setarg)
	if result.Code != core.ResultCode_OK {
		return result
	}
	return CreateDeferredValue(core.ResultCode_YIELD, macro.body, subscope)
}
func (macro *macroCommand) Resume(result core.Result, context any) core.Result {
	scope := context.(*Scope)
	if macro.guard != nil {
		return CreateDeferredValue(
			core.ResultCode_OK,
			core.TUPLE([]core.Value{macro.guard, result.Value}),
			scope,
		)
	}
	return core.OK(result.Value)
}
func (macro *macroCommand) Help(args []core.Value, options core.CommandHelpOptions, _ any) core.Result {
	var usage string
	if options.Skip > 0 {
		usage = macro.argspec.Usage(options.Skip - 1)
	} else {
		usage = MACRO_COMMAND_SIGNATURE(args[0], macro.argspec.Usage(0))
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
	if !macro.argspec.CheckArity(args, 1) &&
		uint(len(args)) > macro.argspec.Argspec.NbRequired {
		return ARITY_ERROR(signature)
	}
	return core.OK(core.STR(signature))
}

const MACRO_SIGNATURE = "macro ?name? argspec body"

type macroCmd struct{}

func (macroCmd) Execute(args []core.Value, context any) core.Result {
	scope := context.(*Scope)
	var name, specs, body core.Value
	switch len(args) {
	case 3:
		specs, body = args[1], args[2]
	case 4:
		name, specs, body = args[1], args[2], args[3]
	default:
		return ARITY_ERROR(MACRO_SIGNATURE)
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
	macro := newMacroCommand(argspec, body.(core.ScriptValue), guard)
	if name != nil {
		result := scope.RegisterCommand(name, macro)
		if result.Code != core.ResultCode_OK {
			return result
		}
	}
	return core.OK(macro.metacommand.value)
}
func (macroCmd) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	if len(args) > 4 {
		return ARITY_ERROR(MACRO_SIGNATURE)
	}
	return core.OK(core.STR(MACRO_SIGNATURE))
}
