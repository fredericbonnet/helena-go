package helena_dialect

import "helena/core"

type namespaceMetacommand struct {
	value     core.Value
	namespace *namespaceCommand
}
type namespaceMetacommandValue struct {
	metacommand *namespaceMetacommand
}

func newNamespaceMetacommand(namespace *namespaceCommand) *namespaceMetacommand {
	metacommand := &namespaceMetacommand{}
	metacommand.value = newNamespaceMetacommandValue(metacommand)
	metacommand.namespace = namespace
	return metacommand
}
func newNamespaceMetacommandValue(metacommand *namespaceMetacommand) core.Value {
	return namespaceMetacommandValue{metacommand}
}
func (value namespaceMetacommandValue) Type() core.ValueType {
	return core.ValueType_COMMAND
}
func (value namespaceMetacommandValue) Command() core.Command {
	return value.metacommand
}
func (value namespaceMetacommandValue) SelectKey(key core.Value) core.Result {
	return value.metacommand.namespace.scope.GetVariable(key, nil)
}

var namespaceMetacommandSubcommands = NewSubcommands([]string{
	"subcommands",
	"eval",
	"call",
	"import",
})

func (metacommand *namespaceMetacommand) Execute(args []core.Value, context any) core.Result {
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
		return core.OK(namespaceMetacommandSubcommands.List)

	case "eval":
		if len(args) != 3 {
			return ARITY_ERROR("<metacommand> eval body")
		}
		body := args[2]
		var program *core.Program
		switch body.Type() {
		case core.ValueType_SCRIPT:
			program = metacommand.namespace.scope.CompileScriptValue(
				body.(core.ScriptValue),
			)
		case core.ValueType_TUPLE:
			program = metacommand.namespace.scope.CompileTupleValue(
				body.(core.TupleValue),
			)
		default:
			return core.ERROR("body must be a script or tuple")
		}
		return CreateContinuationValue(metacommand.namespace.scope, program)

	case "call":
		if len(args) < 3 {
			return ARITY_ERROR("<metacommand> call cmdname ?arg ...?")
		}
		result, subcommand := core.ValueToString(args[2])
		if result.Code != core.ResultCode_OK {
			return core.ERROR("invalid command name")
		}
		command := metacommand.namespace.scope.ResolveLocalCommand(subcommand)
		if command == nil {
			return core.ERROR(`unknown command "` + subcommand + `"`)
		}
		cmdline := make([]core.Value, 1, len(args)-2)
		cmdline[0] = command
		cmdline = append(cmdline, args[3:]...)
		program := metacommand.namespace.scope.CompileArgs(cmdline)
		return CreateContinuationValue(metacommand.namespace.scope, program)

	case "import":
		if len(args) != 3 && len(args) != 4 {
			return ARITY_ERROR("<metacommand> import name ?alias?")
		}
		result, name := core.ValueToString(args[2])
		if result.Code != core.ResultCode_OK {
			return core.ERROR("invalid import name")
		}
		var alias string
		if len(args) == 4 {
			result, s := core.ValueToString(args[3])
			if result.Code != core.ResultCode_OK {
				return core.ERROR("invalid alias name")
			}
			alias = s
		} else {
			alias = name
		}
		command := metacommand.namespace.scope.ResolveNamedCommand(name)
		if command == nil {
			return core.ERROR(`cannot resolve imported command "` + name + `"`)
		}
		scope.RegisterNamedCommand(alias, command)
		return core.OK(core.NIL)

	default:
		return UNKNOWN_SUBCOMMAND_ERROR(subcommand)
	}
}
func (*namespaceMetacommand) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
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

	case "import":
		if len(args) > 4 {
			return ARITY_ERROR("<metacommand> import name ?alias?")
		}
		return core.OK(core.STR("<metacommand> import name ?alias?"))

	default:
		return UNKNOWN_SUBCOMMAND_ERROR(subcommand)
	}
}

func NAMESPACE_COMMAND_PREFIX(name core.Value) string {
	return USAGE_PREFIX(name, "<namespace>", core.CommandHelpOptions{})
}
func NAMESPACE_HELP_PREFIX(name core.Value, options core.CommandHelpOptions) string {
	return USAGE_PREFIX(name, "<namespace>", options)
}

type namespaceCommand struct {
	metacommand *namespaceMetacommand
	scope       *Scope
}

func newNamespaceCommand(scope *Scope) *namespaceCommand {
	namespace := &namespaceCommand{}
	namespace.scope = scope
	namespace.metacommand = newNamespaceMetacommand(namespace)
	return namespace
}
func (namespace *namespaceCommand) Execute(args []core.Value, _ any) core.Result {
	if len(args) == 1 {
		return core.OK(namespace.metacommand.value)
	}
	result, subcommand := core.ValueToString(args[1])
	if result.Code != core.ResultCode_OK {
		return INVALID_SUBCOMMAND_ERROR()
	}
	if subcommand == "subcommands" {
		if len(args) != 2 {
			return ARITY_ERROR(NAMESPACE_COMMAND_PREFIX(args[0]) + " subcommands")
		}
		localCommands := namespace.scope.GetLocalCommandNames()
		list := make([]core.Value, len(localCommands)+1)
		list[0] = args[1]
		for i, name := range localCommands {
			list[i+1] = core.STR(name)
		}
		return core.OK(core.LIST(list))
	}
	command := namespace.scope.ResolveLocalCommand(subcommand)
	if command == nil {
		return UNKNOWN_SUBCOMMAND_ERROR(subcommand)
	}
	cmdline := make([]core.Value, 1, len(args)-1)
	cmdline[0] = command
	cmdline = append(cmdline, args[2:]...)
	program := namespace.scope.CompileArgs(cmdline)
	return CreateContinuationValue(namespace.scope, program)
}
func (namespace *namespaceCommand) Help(args []core.Value, options core.CommandHelpOptions, context any) core.Result {
	signature := NAMESPACE_HELP_PREFIX(args[0], options)
	if len(args) <= 1 {
		return core.OK(core.STR(signature + " ?subcommand? ?arg ...?"))
	}
	result, subcommand := core.ValueToString(args[1])
	if result.Code != core.ResultCode_OK {
		return INVALID_SUBCOMMAND_ERROR()
	}
	if subcommand == "subcommands" {
		if len(args) > 2 {
			return ARITY_ERROR(signature + " subcommands")
		}
		return core.OK(core.STR(signature + " subcommands"))
	}
	command := namespace.scope.ResolveLocalCommand(subcommand)
	if command == nil {
		return UNKNOWN_SUBCOMMAND_ERROR(subcommand)
	}
	if c, ok := command.Command().(core.CommandWithHelp); ok {
		return c.Help(args[1:], core.CommandHelpOptions{
			Prefix: signature + " " + subcommand,
			Skip:   1,
		}, context,
		)
	}
	return core.ERROR(`no help for subcommand "` + subcommand + `"`)
}

const NAMESPACE_SIGNATURE = "namespace ?name? body"

type namespaceCmd struct{}

func (namespaceCmd) Execute(args []core.Value, context any) core.Result {
	scope := context.(*Scope)
	var name, body core.Value
	switch len(args) {
	case 2:
		body = args[1]
	case 3:
		name, body = args[1], args[2]
	default:
		return ARITY_ERROR(NAMESPACE_SIGNATURE)
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
				namespace := newNamespaceCommand(subscope)
				if name != nil {
					result := scope.RegisterCommand(name, namespace)
					if result.Code != core.ResultCode_OK {
						return result
					}
				}
				if result.Code == core.ResultCode_RETURN {
					return core.OK(result.Value)
				} else {
					return core.OK(namespace.metacommand.value)
				}
			}
		case core.ResultCode_ERROR:
			return result
		default:
			return core.ERROR("unexpected " + core.RESULT_CODE_NAME(result))
		}
	})
}
func (namespaceCmd) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	if len(args) > 3 {
		return ARITY_ERROR(NAMESPACE_SIGNATURE)
	}
	return core.OK(core.STR(NAMESPACE_SIGNATURE))
}
