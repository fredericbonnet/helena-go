package helena_dialect

// /* eslint-disable jsdoc/require-jsdoc */ // TODO
// import { macroCmd } from "./macros";
// import { Scope } from "./core";
// import { scopeCmd } from "./scopes";
// import { closureCmd } from "./closures";
// import { coroutineCmd } from "./coroutines";
// import { registerArgspecCommands } from "./argspecs";
// import { aliasCmd } from "./aliases";
// import { procCmd } from "./procs";
// import { namespaceCmd } from "./namespaces";
// import { registerBasicCommands } from "./basic-commands";
// import { registerVariableCommands } from "./variables";
// import { registerMathCommands } from "./math";
// import { registerLogicCommands } from "./logic";
// import { registerControlCommands } from "./controls";
// import { registerStringCommands } from "./strings";
// import { registerListCommands } from "./lists";
// import { registerDictCommands } from "./dicts";
// import { registerTupleCommands } from "./tuples";
// import { registerScriptCommands } from "./scripts";
// import { ensembleCmd } from "./ensembles";
// import { ModuleRegistry, registerModuleCommands } from "./modules";
// import { registerNumberCommands } from "./numbers";

// export { Scope } from "./core";

// const globalModuleRegistry = new ModuleRegistry();

func InitCommands(
	scope *Scope,
	//   moduleRegistry?: ModuleRegistry,
	//   rootDir?: string
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
	//   scope.RegisterNamedCommand("namespace", namespaceCmd);
	scope.RegisterNamedCommand("ensemble", ensembleCmd{})

	//   registerModuleCommands(
	//     scope,
	//     moduleRegistry ?? globalModuleRegistry,
	//     rootDir ?? process.cwd()
	//   );

	scope.RegisterNamedCommand("macro", macroCmd{})
	scope.RegisterNamedCommand("closure", closureCmd{})
	scope.RegisterNamedCommand("proc", procCmd{})
	//   scope.RegisterNamedCommand("coroutine", coroutineCmd);
	scope.RegisterNamedCommand("alias", aliasCmd{})
}
