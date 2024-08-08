package helena_dialect

import "helena/core"

type aliasMetacommand struct {
	value core.Value
	alias *aliasCommand
}

func newAliasMetacommand(alias *aliasCommand) *aliasMetacommand {
	metacommand := &aliasMetacommand{}
	metacommand.value = core.NewCommandValue(metacommand)
	metacommand.alias = alias
	return metacommand
}

var aliasMetacommandSubcommands = NewSubcommands([]string{"subcommands", "command"})

func (metacommand *aliasMetacommand) Execute(args []core.Value, _ any) core.Result {
	if len(args) == 1 {
		return core.OK(metacommand.alias.value)
	}
	result, subcommand := core.ValueToString(args[1])
	if result.Code != core.ResultCode_OK {
		return INVALID_SUBCOMMAND_ERROR()
	}
	switch subcommand {
	case "subcommands":
		if len(args) != 2 {
			return ARITY_ERROR("<alias> subcommands")
		}
		return core.OK(aliasMetacommandSubcommands.List)

	case "command":
		if len(args) != 2 {
			return ARITY_ERROR("<alias> command")
		}
		return core.OK(metacommand.alias.cmd)

	default:
		return UNKNOWN_SUBCOMMAND_ERROR(subcommand)
	}
}
func (metacommand *aliasMetacommand) Resume(result core.Result, context any) core.Result {
	return metacommand.alias.Resume(result, context)
}

type aliasCommand struct {
	value       core.Value
	cmd         core.Value
	metacommand *aliasMetacommand
}

func newAliasCommand(cmd core.Value) *aliasCommand {
	alias := &aliasCommand{}
	alias.value = core.NewCommandValue(alias)
	alias.cmd = cmd
	alias.metacommand = newAliasMetacommand(alias)
	return alias
}

func (command *aliasCommand) Execute(args []core.Value, context any) core.Result {
	cmdline := append([]core.Value{command.cmd}, args[1:]...)
	return expandPrefixCmd.Execute(cmdline, context)
}
func (command *aliasCommand) Resume(result core.Result, context any) core.Result {
	return expandPrefixCmd.Resume(result, context)
}

const ALIAS_SIGNATURE = "alias name command"

type aliasCmd struct{}

func (aliasCmd) Execute(args []core.Value, context any) core.Result {
	scope := context.(*Scope)
	if len(args) != 3 {
		return ARITY_ERROR(ALIAS_SIGNATURE)
	}
	name, cmd := args[1], args[2]

	alias := newAliasCommand(cmd)
	result := scope.RegisterCommand(name, alias)
	if result.Code != core.ResultCode_OK {
		return result
	}
	return core.OK(alias.metacommand.value)
}
func (aliasCmd) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	if len(args) > 3 {
		return ARITY_ERROR(ALIAS_SIGNATURE)
	}
	return core.OK(core.STR(ALIAS_SIGNATURE))
}
