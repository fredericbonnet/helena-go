package go_regexp

import (
	"helena/core"
	"helena/helena_dialect"
)

/**
 * Main static module entry point.
 */
func Initmodule() *helena_dialect.Module {
	scope := helena_dialect.NewRootScope(nil)
	exports := &helena_dialect.Exports{}
	module := helena_dialect.NewModule(scope, exports)
	exportCommand(module, "regexp", RegexpCmd{})
	return module
}

func exportCommand(module *helena_dialect.Module, name string, cmd core.Command) {
	module.Scope.RegisterNamedCommand(name, cmd)
	(*module.Exports)[name] = core.STR(name)
}
