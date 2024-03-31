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
	return namespaceMetacommandSubcommands.Dispatch(args[1], SubcommandHandlers{
		"subcommands": func() core.Result {
			if len(args) != 2 {
				return ARITY_ERROR("<namespace> subcommands")
			}
			return core.OK(namespaceMetacommandSubcommands.List)
		},
		"eval": func() core.Result {
			if len(args) != 3 {
				return ARITY_ERROR("<namespace> eval body")
			}
			return CreateDeferredValue(
				core.ResultCode_YIELD,
				args[2],
				metacommand.namespace.scope,
			)
		},
		"call": func() core.Result {
			if len(args) < 3 {
				return ARITY_ERROR("<namespace> call cmdname ?arg ...?")
			}
			result := core.ValueToString(args[2])
			if result.Code != core.ResultCode_OK {
				return core.ERROR("invalid command name")
			}
			subcommand := result.Data
			if !metacommand.namespace.scope.HasLocalCommand(subcommand) {
				return core.ERROR(`unknown command "` + subcommand + `"`)
			}
			command := metacommand.namespace.scope.ResolveNamedCommand(subcommand)
			cmdline := append([]core.Value{core.NewCommandValue(command)}, args[3:]...)
			return CreateDeferredValue(
				core.ResultCode_YIELD,
				core.TUPLE(cmdline),
				metacommand.namespace.scope,
			)
		},
		"import": func() core.Result {
			if len(args) != 3 && len(args) != 4 {
				return ARITY_ERROR("<namespace> import name ?alias?")
			}
			result := core.ValueToString(args[2])
			if result.Code != core.ResultCode_OK {
				return core.ERROR("invalid import name")
			}
			name := result.Data
			var alias string
			if len(args) == 4 {
				result := core.ValueToString(args[3])
				if result.Code != core.ResultCode_OK {
					return core.ERROR("invalid alias name")
				}
				alias = result.Data
			} else {
				alias = name
			}
			command := metacommand.namespace.scope.ResolveNamedCommand(name)
			if command == nil {
				return core.ERROR(`cannot resolve imported command "` + name + `"`)
			}
			scope.RegisterNamedCommand(alias, command)
			return core.OK(core.NIL)
		},
	})
}

func NAMESPACE_COMMAND_PREFIX(name core.Value) string {
	return core.ValueToStringOrDefault(name, "<namespace>").Data
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
	result := core.ValueToString(args[1])
	if result.Code != core.ResultCode_OK {
		return core.ERROR("invalid subcommand name")
	}
	subcommand := result.Data
	if subcommand == "subcommands" {
		if len(args) != 2 {
			return ARITY_ERROR(NAMESPACE_COMMAND_PREFIX(args[0]) + " subcommands")
		}
		localCommands := namespace.scope.GetLocalCommands()
		list := make([]core.Value, len(localCommands)+1)
		list[0] = args[1]
		for i, name := range localCommands {
			list[i+1] = core.STR(name)
		}
		return core.OK(core.LIST(list))
	}
	if !namespace.scope.HasLocalCommand(subcommand) {
		return core.ERROR(`unknown subcommand "` + subcommand + `"`)
	}
	command := namespace.scope.ResolveNamedCommand(subcommand)
	cmdline := append(
		[]core.Value{core.NewCommandValue(command)},
		args[2:]...,
	)
	return CreateDeferredValue(core.ResultCode_YIELD, core.TUPLE(cmdline), namespace.scope)
}
func (namespace *namespaceCommand) Help(args []core.Value, options core.CommandHelpOptions, context any) core.Result {
	var usage string
	if options.Skip > 0 {
		usage = ""
	} else {
		usage = NAMESPACE_COMMAND_PREFIX(args[0])
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
	if len(args) <= 1 {
		return core.OK(core.STR(signature + " ?subcommand? ?arg ...?"))
	}
	result := core.ValueToString(args[1])
	if result.Code != core.ResultCode_OK {
		return core.ERROR("invalid subcommand name")
	}
	subcommand := result.Data
	if subcommand == "subcommands" {
		if len(args) > 2 {
			return ARITY_ERROR(signature + " subcommands")
		}
		return core.OK(core.STR(signature + " subcommands"))
	}
	if !namespace.scope.HasLocalCommand(subcommand) {
		return core.ERROR(`unknown subcommand "` + subcommand + `"`)
	}
	command := namespace.scope.ResolveNamedCommand(subcommand)
	if c, ok := command.(core.CommandWithHelp); ok {
		return c.Help(args[1:], core.CommandHelpOptions{
			Prefix: signature + " " + subcommand,
			Skip:   1,
		}, context,
		)
	}
	return core.ERROR(`no help for subcommand "` + subcommand + `"`)
}

const NAMESPACE_SIGNATURE = "namespace ?name? body"

type namespaceBodyState struct {
	scope    *Scope
	subscope *Scope
	process  *Process
	name     core.Value
}

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

	subscope := NewScope(scope, false)
	process := subscope.PrepareScriptValue(body.(core.ScriptValue))
	return executeNamespaceBody(&namespaceBodyState{scope, subscope, process, name})
}
func (namespaceCmd) Resume(result core.Result, _ any) core.Result {
	state := result.Data.(*namespaceBodyState)
	state.process.YieldBack(result.Value)
	return executeNamespaceBody(state)
}
func (namespaceCmd) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	if len(args) > 3 {
		return ARITY_ERROR(NAMESPACE_SIGNATURE)
	}
	return core.OK(core.STR(NAMESPACE_SIGNATURE))
}

func executeNamespaceBody(state *namespaceBodyState) core.Result {
	result := state.process.Run()
	switch result.Code {
	case core.ResultCode_OK,
		core.ResultCode_RETURN:
		{
			namespace := newNamespaceCommand(state.subscope)
			if state.name != nil {
				result := state.scope.RegisterCommand(state.name, namespace)
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
	case core.ResultCode_YIELD:
		return core.YIELD_STATE(result.Value, state)
	case core.ResultCode_ERROR:
		return result
	default:
		return core.ERROR("unexpected " + core.RESULT_CODE_NAME(result))
	}
}
