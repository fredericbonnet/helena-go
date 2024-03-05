package helena_dialect

import "helena/core"

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
		return core.ERROR("invalid subcommand name")
	}
	name := result.Data
	if handlers[name] == nil {
		return core.ERROR(`unknown subcommand "` + name + `"`)
	}
	return handlers[name]()
}
