package helena_dialect

import (
	"fmt"
	"helena/core"
	"os"
	"path/filepath"
)

type Exports = map[string]core.Value

const EXPORT_SIGNATURE = "export name"

type exportCommand struct {
	exports *Exports
}

func newExportCommand(exports *Exports) exportCommand {
	return exportCommand{exports}
}

func (cmd exportCommand) Execute(args []core.Value, _ any) core.Result {
	if len(args) != 2 {
		return ARITY_ERROR(EXPORT_SIGNATURE)
	}
	result := core.ValueToString(args[1])
	if result.Code != core.ResultCode_OK {
		return core.ERROR("invalid export name")
	}
	name := result.Data
	(*cmd.exports)[name] = core.STR(name)
	return core.OK(core.NIL)
}
func (exportCommand) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	if len(args) > 2 {
		return ARITY_ERROR(EXPORT_SIGNATURE)
	}
	return core.OK(core.STR(EXPORT_SIGNATURE))
}

type Module struct {
	value   core.Value
	scope   *Scope
	exports *Exports
}

func NewModule(scope *Scope, exports *Exports) *Module {
	module := &Module{}
	module.value = core.NewCommandValue(module)
	module.scope = scope
	module.exports = exports
	return module
}

var moduleSubcommands = NewSubcommands([]string{
	"subcommands",
	"exports",
	"import",
})

func (module *Module) Execute(args []core.Value, context any) core.Result {
	scope := context.(*Scope)
	if len(args) == 1 {
		return core.OK(module.value)
	}
	result := core.ValueToString(args[1])
	if result.Code != core.ResultCode_OK {
		return INVALID_SUBCOMMAND_ERROR()
	}
	subcommand := result.Data
	switch subcommand {
	case "subcommands":
		if len(args) != 2 {
			return ARITY_ERROR("<module> subcommands")
		}
		return core.OK(moduleSubcommands.List)

	case "exports":
		if len(args) != 2 {
			return ARITY_ERROR("<module> exports")
		}
		values := make([]core.Value, len(*module.exports))
		i := 0
		for _, value := range *module.exports {
			values[i] = value
			i++
		}
		return core.OK(core.LIST(values))

	case "import":
		if len(args) != 3 && len(args) != 4 {
			return ARITY_ERROR("<module> import name ?alias?")
		}
		var aliasName core.Value
		if len(args) == 4 {
			aliasName = args[3]
		} else {
			aliasName = args[2]
		}
		return importCommand(
			args[2],
			aliasName,
			module.exports,
			module.scope,
			scope,
		)

	default:
		return UNKNOWN_SUBCOMMAND_ERROR(subcommand)
	}
}

func importCommand(
	importName core.Value,
	aliasName core.Value,
	exports *Exports,
	source *Scope,
	destination *Scope,
) core.Result {
	result := core.ValueToString(importName)
	if result.Code != core.ResultCode_OK {
		return core.ERROR("invalid import name")
	}
	name := result.Data
	result2 := core.ValueToString(aliasName)
	if result2.Code != core.ResultCode_OK {
		return core.ERROR("invalid alias name")
	}
	alias := result2.Data
	if (*exports)[name] == nil {
		return core.ERROR(`unknown export "` + name + `"`)
	}
	command := source.ResolveNamedCommand(name)
	if command == nil {
		return core.ERROR(`cannot resolve export "` + name + `"`)
	}
	destination.RegisterNamedCommand(alias, command)
	return core.OK(core.NIL)
}

type ModuleRegistry struct {
	modules       map[string]*Module
	reservedNames map[string]struct{}
}

func NewModuleRegistry() *ModuleRegistry {
	return &ModuleRegistry{
		map[string]*Module{},
		map[string]struct{}{},
	}
}

func (registry *ModuleRegistry) IsReserved(name string) bool {
	_, ok := registry.reservedNames[name]
	return ok
}
func (registry *ModuleRegistry) Reserve(name string) {
	registry.reservedNames[name] = struct{}{}
}
func (registry *ModuleRegistry) Release(name string) {
	delete(registry.reservedNames, name)
}
func (registry *ModuleRegistry) IsRegistered(name string) bool {
	_, ok := registry.modules[name]
	return ok
}
func (registry *ModuleRegistry) Register(name string, module *Module) {
	registry.modules[name] = module
}
func (registry *ModuleRegistry) Get(name string) *Module {
	return registry.modules[name]
}

const MODULE_SIGNATURE = "module ?name? body"

type moduleCommand struct {
	moduleRegistry *ModuleRegistry
	rootDir        string
}

func newModuleCommand(moduleRegistry *ModuleRegistry, rootDir string) *moduleCommand {
	return &moduleCommand{moduleRegistry, rootDir}
}

func (cmd *moduleCommand) Execute(args []core.Value, context any) core.Result {
	scope := context.(*Scope)
	var name, body core.Value
	switch len(args) {
	case 2:
		body = args[1]
	case 3:
		name, body = args[1], args[2]
	default:
		return ARITY_ERROR(MODULE_SIGNATURE)
	}
	if body.Type() != core.ValueType_SCRIPT {
		return core.ERROR("body must be a script")
	}

	result := createModule(
		cmd.moduleRegistry,
		cmd.rootDir,
		body.(core.ScriptValue).Script,
	)
	if result.Code != core.ResultCode_OK {
		return result.AsResult()
	}
	module := result.Data
	if name != nil {
		result := scope.RegisterCommand(name, module)
		if result.Code != core.ResultCode_OK {
			return result
		}
	}
	return core.OK(module.value)
}
func (*moduleCommand) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	if len(args) > 3 {
		return ARITY_ERROR(MODULE_SIGNATURE)
	}
	return core.OK(core.STR(MODULE_SIGNATURE))
}

func resolveModule(
	moduleRegistry *ModuleRegistry,
	rootDir string,
	nameOrPath string,
) core.TypedResult[*Module] {
	if moduleRegistry.IsRegistered(nameOrPath) {
		module := moduleRegistry.Get(nameOrPath)
		return core.OK_T(module.value, module)
	}
	return resolveFileBasedModule(moduleRegistry, rootDir, nameOrPath)
}

func resolveFileBasedModule(
	moduleRegistry *ModuleRegistry,
	rootDir string,
	filePath string,
) core.TypedResult[*Module] {
	modulePath := filepath.Join(rootDir, filePath)
	if moduleRegistry.IsRegistered(modulePath) {
		module := moduleRegistry.Get(modulePath)
		return core.OK_T(module.value, module)
	}

	result := loadFileBasedModule(
		moduleRegistry,
		modulePath,
	)
	if result.Code != core.ResultCode_OK {
		return result
	}
	module := result.Data
	moduleRegistry.Register(modulePath, module)
	return core.OK_T(module.value, module)
}

func loadFileBasedModule(
	moduleRegistry *ModuleRegistry,
	modulePath string,
) core.TypedResult[*Module] {
	if moduleRegistry.IsReserved(modulePath) {
		return core.ERROR_T[*Module]("circular imports are forbidden")
	}
	moduleRegistry.Reserve(modulePath)

	data, err := os.ReadFile(modulePath)
	if err != nil {
		moduleRegistry.Release(modulePath)
		return core.ERROR_T[*Module]("error reading module: " + fmt.Sprint(err))
	}
	tokens := core.Tokenizer{}.Tokenize(string(data))
	parser := core.NewParser(nil)
	parseResult := parser.Parse(tokens)
	if !parseResult.Success {
		moduleRegistry.Release(modulePath)
		return core.ERROR_T[*Module](parseResult.Message)
	}

	result := createModule(moduleRegistry, filepath.Dir(modulePath), *parseResult.Script)
	moduleRegistry.Release(modulePath)
	return result
}

func createModule(
	moduleRegistry *ModuleRegistry,
	rootDir string,
	script core.Script,
) core.TypedResult[*Module] {
	rootScope := NewScope(nil, false)
	InitCommandsForModule(rootScope, moduleRegistry, rootDir)

	exports := &Exports{}
	rootScope.RegisterNamedCommand("export", newExportCommand(exports))

	program := rootScope.Compile(script)
	process := rootScope.PrepareProcess(program)
	result := process.Run()
	if result.Code == core.ResultCode_ERROR {
		return core.ResultAs[*Module](result)
	}
	if result.Code != core.ResultCode_OK {
		return core.ERROR_T[*Module]("unexpected " + core.RESULT_CODE_NAME(result))
	}

	module := NewModule(rootScope, exports)
	return core.OK_T(module.value, module)
}

const IMPORT_SIGNATURE = "import path ?name|imports?"

type importCmd struct {
	moduleRegistry *ModuleRegistry
	rootDir        string
}

func newImportCommand(moduleRegistry *ModuleRegistry, rootDir string) *importCmd {
	return &importCmd{moduleRegistry, rootDir}
}

func (cmd *importCmd) Execute(args []core.Value, context any) core.Result {
	scope := context.(*Scope)
	if len(args) != 2 && len(args) != 3 {
		return ARITY_ERROR(IMPORT_SIGNATURE)
	}

	result := core.ValueToString(args[1])
	if result.Code != core.ResultCode_OK {
		return core.ERROR("invalid path")
	}
	path := result.Data

	result2 := resolveModule(
		cmd.moduleRegistry,
		cmd.rootDir,
		path,
	)
	if result2.Code != core.ResultCode_OK {
		return result2.AsResult()
	}
	module := result2.Data

	if len(args) >= 3 {
		switch args[2].Type() {
		case core.ValueType_LIST,
			core.ValueType_TUPLE,
			core.ValueType_SCRIPT:
			{
				// Import names
				result := ValueToArray(args[2])
				if result.Code != core.ResultCode_OK {
					return result.AsResult()
				}
				names := result.Data
				for _, name := range names {
					if name.Type() == core.ValueType_TUPLE {
						values := name.(core.TupleValue).Values
						if len(values) != 2 {
							return core.ERROR("invalid (name alias) tuple")
						}
						result := importCommand(
							values[0],
							values[1],
							module.exports,
							module.scope,
							scope,
						)
						if result.Code != core.ResultCode_OK {
							return result
						}
					} else {
						result := importCommand(
							name,
							name,
							module.exports,
							module.scope,
							scope,
						)
						if result.Code != core.ResultCode_OK {
							return result
						}
					}
				}
			}
		default:
			{
				// Module command name
				result := scope.RegisterCommand(args[2], module)
				if result.Code != core.ResultCode_OK {
					return result
				}
			}
		}
	}
	return core.OK(module.value)
}
func (importCmd) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	if len(args) > 3 {
		return ARITY_ERROR(IMPORT_SIGNATURE)
	}
	return core.OK(core.STR(IMPORT_SIGNATURE))
}

func registerModuleCommands(
	scope *Scope,
	moduleRegistry *ModuleRegistry,
	rootDir string,
) {
	scope.RegisterNamedCommand(
		"module",
		newModuleCommand(moduleRegistry, rootDir),
	)
	scope.RegisterNamedCommand(
		"import",
		newImportCommand(moduleRegistry, rootDir),
	)
}
