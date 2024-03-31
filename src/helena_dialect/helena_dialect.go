package helena_dialect

import "os"

var globalModuleRegistry = NewModuleRegistry()

func InitCommands(scope *Scope) {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	InitCommandsForModule(scope, globalModuleRegistry, cwd)
}

func InitCommandsForModule(
	scope *Scope,
	moduleRegistry *ModuleRegistry,
	rootDir string,
) {
	registerBasicCommands(scope)
	registerVariableCommands(scope)

	registerMathCommands(scope)
	registerLogicCommands(scope)
	registerControlCommands(scope)

	registerNumberCommands(scope)
	registerStringCommands(scope)
	registerListCommands(scope)
	registerDictCommands(scope)
	registerTupleCommands(scope)
	registerScriptCommands(scope)
	registerArgspecCommands(scope)

	scope.RegisterNamedCommand("scope", scopeCmd{})
	scope.RegisterNamedCommand("namespace", namespaceCmd{})
	scope.RegisterNamedCommand("ensemble", ensembleCmd{})

	registerModuleCommands(
		scope,
		moduleRegistry,
		rootDir,
	)

	scope.RegisterNamedCommand("macro", macroCmd{})
	scope.RegisterNamedCommand("closure", closureCmd{})
	scope.RegisterNamedCommand("proc", procCmd{})
	//   scope.RegisterNamedCommand("coroutine", coroutineCmd);
	scope.RegisterNamedCommand("alias", aliasCmd{})
}
