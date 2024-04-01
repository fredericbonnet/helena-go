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
