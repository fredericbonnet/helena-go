package helena_dialect

import "helena/core"

type ensembleMetacommand struct {
	value    core.Value
	ensemble *EnsembleCommand
}

func newEnsembleMetacommand(ensemble *EnsembleCommand) *ensembleMetacommand {
	metacommand := &ensembleMetacommand{}
	metacommand.value = core.NewCommandValue(metacommand)
	metacommand.ensemble = ensemble
	return metacommand
}

var ensembleMetacommandSubcommands = NewSubcommands([]string{
	"subcommands",
	"eval",
	"call",
	"argspec",
})

func (metacommand *ensembleMetacommand) Execute(args []core.Value, context any) core.Result {
	scope := context.(*Scope)
	if len(args) == 1 {
		return core.OK(metacommand.value)
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
		return core.OK(ensembleMetacommandSubcommands.List)

	case "eval":
		if len(args) != 3 {
			return ARITY_ERROR("<metacommand> eval body")
		}
		body := args[2]
		var program *core.Program
		switch body.Type() {
		case core.ValueType_SCRIPT:
			program = metacommand.ensemble.scope.CompileScriptValue(
				body.(core.ScriptValue),
			)
		case core.ValueType_TUPLE:
			program = metacommand.ensemble.scope.CompileTupleValue(
				body.(core.TupleValue),
			)
		default:
			return core.ERROR("body must be a script or tuple")
		}
		return CreateContinuationValue(metacommand.ensemble.scope, program)

	case "call":
		if len(args) < 3 {
			return ARITY_ERROR("<metacommand> call cmdname ?arg ...?")
		}
		result, subcommand := core.ValueToString(args[2])
		if result.Code != core.ResultCode_OK {
			return core.ERROR("invalid command name")
		}
		command := metacommand.ensemble.scope.ResolveLocalCommand(subcommand)
		if command == nil {
			return core.ERROR(`unknown command "` + subcommand + `"`)
		}
		cmdline := make([]core.Value, 1, len(args)-2)
		cmdline[0] = command
		cmdline = append(cmdline, args[3:]...)
		program := scope.CompileArgs(cmdline)
		return CreateContinuationValue(scope, program)

	case "argspec":
		if len(args) != 2 {
			return ARITY_ERROR("<metacommand> argspec")
		}
		return core.OK(metacommand.ensemble.argspec)

	default:
		return UNKNOWN_SUBCOMMAND_ERROR(subcommand)
	}
}
func (*ensembleMetacommand) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
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

	case "eval":
		if len(args) > 3 {
			return ARITY_ERROR("<metacommand> eval body")
		}
		return core.OK(core.STR("<metacommand> eval body"))

	case "call":
		return core.OK(core.STR("<metacommand> call cmdname ?arg ...?"))

	case "argspec":
		if len(args) > 2 {
			return ARITY_ERROR("<metacommand> argspec")
		}
		return core.OK(core.STR("<metacommand> argspec"))

	default:
		return UNKNOWN_SUBCOMMAND_ERROR(subcommand)
	}
}

func ENSEMBLE_COMMAND_PREFIX(name core.Value, argspec ArgspecValue) string {
	return USAGE_ARGSPEC(name, "<ensemble>", argspec, core.CommandHelpOptions{})
}
func ENSEMBLE_HELP_PREFIX(name core.Value, argspec ArgspecValue, options core.CommandHelpOptions) string {
	return USAGE_ARGSPEC(name, "<ensemble>", argspec, options)
}

type EnsembleCommand struct {
	metacommand *ensembleMetacommand
	scope       *Scope
	argspec     ArgspecValue
}
type ensembleSubcommandState struct {
	subcommand core.Command
	result     core.Result
}

func NewEnsembleCommand(scope *Scope, argspec ArgspecValue) *EnsembleCommand {
	ensemble := &EnsembleCommand{}
	ensemble.scope = scope
	ensemble.argspec = argspec
	ensemble.metacommand = newEnsembleMetacommand(ensemble)
	return ensemble
}
func (ensemble *EnsembleCommand) Execute(args []core.Value, context any) core.Result {
	scope := context.(*Scope)
	if len(args) == 1 {
		return core.OK(ensemble.metacommand.value)
	}
	minArgs := ensemble.argspec.Argspec.NbRequired + 1
	if uint(len(args)) < minArgs {
		return ARITY_ERROR(
			ENSEMBLE_COMMAND_PREFIX(args[0], ensemble.argspec) +
				" ?subcommand? ?arg ...?",
		)
	}
	if uint(len(args)) == minArgs {
		result, ensembleArgs := ensemble.argspec.CollectArguments(
			scope,
			args,
			1,
		)
		if result.Code != core.ResultCode_OK {
			return result
		}
		return core.OK(core.TUPLE(ensembleArgs))
	}
	result2, subcommand := core.ValueToString(args[minArgs])
	if result2.Code != core.ResultCode_OK {
		return INVALID_SUBCOMMAND_ERROR()
	}
	if subcommand == "subcommands" {
		if uint(len(args)) != minArgs+1 {
			return ARITY_ERROR(
				ENSEMBLE_COMMAND_PREFIX(args[0], ensemble.argspec) + " subcommands",
			)
		}
		localCommands := ensemble.scope.GetLocalCommandNames()
		list := make([]core.Value, len(localCommands)+1)
		list[0] = args[minArgs]
		for i, name := range localCommands {
			list[i+1] = core.STR(name)
		}
		return core.OK(core.LIST(list))
	}
	command := ensemble.scope.ResolveLocalCommand(subcommand)
	if command == nil {
		return UNKNOWN_SUBCOMMAND_ERROR(subcommand)
	}
	var result core.Result
	if ensemble.argspec.Argspec.IsFast() {
		// Here we can just modify the args array in place and avoid the
		// overhead of copying it to a new array. This provides a huge speedup
		// for most subcommand dispatches.
		old := args[minArgs]
		copy(args[2:minArgs+1], args[1:minArgs])
		args[1] = command
		result = command.Command().Execute(args[1:], scope)
		copy(args[1:minArgs], args[2:minArgs+1])
		args[minArgs] = old
	} else {
		cmdline := make([]core.Value, 1, len(args)-1)
		cmdline[0] = command
		// Note: this will only collect the required args and ignore the remainder so we can just pass the whole array
		result2, values := ensemble.argspec.CollectArguments(scope, args[1:minArgs], 0)
		if result2.Code != core.ResultCode_OK {
			return result2
		}
		cmdline = append(cmdline, values...)
		cmdline = append(cmdline, args[minArgs+1:]...)
		result = command.Command().Execute(cmdline, scope)
	}
	if result.Code == core.ResultCode_YIELD {
		state := ensembleSubcommandState{command.Command(), result}
		return core.YIELD_STATE(result.Value, state)
	}
	return result
}
func (ensemble *EnsembleCommand) Resume(result core.Result, context any) core.Result {
	scope := context.(*Scope)
	state := result.Data.(ensembleSubcommandState)
	subcommand := state.subcommand
	commandResult := state.result
	resumable, ok := subcommand.(core.ResumableCommand)
	if !ok {
		return core.OK(result.Value)
	}
	result2 := resumable.Resume(
		core.Result{
			Code:  commandResult.Code,
			Value: result.Value,
			Data:  commandResult.Data,
		},
		scope,
	)
	if result2.Code == core.ResultCode_YIELD {
		return core.YIELD_STATE(result2.Value, ensembleSubcommandState{subcommand, result2})
	}
	return result2
}
func (ensemble *EnsembleCommand) Help(args []core.Value, options core.CommandHelpOptions, context any) core.Result {
	signature := ENSEMBLE_HELP_PREFIX(args[0], ensemble.argspec, options)
	minArgs := ensemble.argspec.Argspec.NbRequired + 1
	if uint(len(args)) <= minArgs {
		return core.OK(core.STR(signature + " ?subcommand? ?arg ...?"))
	}
	result, subcommand := core.ValueToString(args[minArgs])
	if result.Code != core.ResultCode_OK {
		return INVALID_SUBCOMMAND_ERROR()
	}
	if subcommand == "subcommands" {
		if uint(len(args)) > minArgs+1 {
			return ARITY_ERROR(signature + " subcommands")
		}
		return core.OK(core.STR(signature + " subcommands"))
	}
	command := ensemble.scope.ResolveLocalCommand(subcommand)
	if command == nil {
		return UNKNOWN_SUBCOMMAND_ERROR(subcommand)
	}
	if c, ok := command.Command().(core.CommandWithHelp); ok {
		return c.Help(
			append(append([]core.Value{args[minArgs]}, args[1:minArgs]...), args[minArgs+1:]...),
			core.CommandHelpOptions{
				Prefix: signature + " " + subcommand,
				Skip:   minArgs,
			},
			context,
		)
	}
	return core.ERROR(`no help for subcommand "` + subcommand + `"`)
}

const ENSEMBLE_SIGNATURE = "ensemble ?name? argspec body"

type ensembleCmd struct{}

func (ensembleCmd) Execute(args []core.Value, context any) core.Result {
	scope := context.(*Scope)
	var name, specs, body core.Value
	switch len(args) {
	case 3:
		specs, body = args[1], args[2]
	case 4:
		name, specs, body = args[1], args[2], args[3]
	default:
		return ARITY_ERROR(ENSEMBLE_SIGNATURE)
	}
	if body.Type() != core.ValueType_SCRIPT {
		return core.ERROR("body must be a script")
	}

	result, argspec := ArgspecValueFromValue(specs)
	if result.Code != core.ResultCode_OK {
		return result
	}
	if argspec.Argspec.IsVariadic() {
		return core.ERROR("ensemble arguments cannot be variadic")
	}
	if argspec.Argspec.HasOptions {
		return core.ERROR("ensemble arguments cannot have options")
	}

	subscope := scope.NewChildScope()
	program := subscope.CompileScriptValue(body.(core.ScriptValue))
	return CreateContinuationValueWithCallback(subscope, program, nil, func(result core.Result, data any) core.Result {
		switch result.Code {
		case core.ResultCode_OK,
			core.ResultCode_RETURN:
			{
				ensemble := NewEnsembleCommand(subscope, argspec)
				if name != nil {
					result := scope.RegisterCommand(name, ensemble)
					if result.Code != core.ResultCode_OK {
						return result
					}
				}
				if result.Code == core.ResultCode_RETURN {
					return core.OK(result.Value)
				} else {
					return core.OK(ensemble.metacommand.value)

				}
			}
		case core.ResultCode_ERROR:
			return result
		default:
			return core.ERROR("unexpected " + core.RESULT_CODE_NAME(result))
		}
	})
}
func (ensembleCmd) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	if len(args) > 4 {
		return ARITY_ERROR(ENSEMBLE_SIGNATURE)
	}
	return core.OK(core.STR(ENSEMBLE_SIGNATURE))
}
