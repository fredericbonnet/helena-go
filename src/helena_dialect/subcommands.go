package helena_dialect

import "helena/core"

func INVALID_SUBCOMMAND_ERROR() core.Result { return core.ERROR("invalid subcommand name") }
func UNKNOWN_SUBCOMMAND_ERROR(name string) core.Result {
	return core.ERROR(`unknown subcommand "` + name + `"`)
}

type Subcommands struct {
	List core.Value
}
type SubcommandHandlers map[string](func() core.Result)

func NewSubcommands(names []string) Subcommands {
	values := make([]core.Value, len(names))
	for i, name := range names {
		values[i] = core.STR(name)
	}
	return Subcommands{
		List: core.LIST(values),
	}
}

func (subcommands Subcommands) Dispatch(
	subcommand core.Value,
	handlers SubcommandHandlers,
) core.Result {
	result := core.ValueToString(subcommand)
	if result.Code != core.ResultCode_OK {
		return INVALID_SUBCOMMAND_ERROR()
	}
	name := result.Data
	if handlers[name] == nil {
		return UNKNOWN_SUBCOMMAND_ERROR(name)
	}
	return handlers[name]()
}
