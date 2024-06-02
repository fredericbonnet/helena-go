package helena_dialect

import "helena/core"

type dictCommand struct {
	scope    *Scope
	ensemble *EnsembleCommand
}

func newDictCommand(scope *Scope) *dictCommand {
	dict := &dictCommand{}
	dict.scope = scope.NewChildScope()
	argspec := ArgspecValueFromValue(core.LIST([]core.Value{core.STR("value")})).Data
	dict.ensemble = NewEnsembleCommand(dict.scope, argspec)
	return dict
}
func (dict *dictCommand) Execute(args []core.Value, context any) core.Result {
	if len(args) == 2 {
		return valueToDictionaryValue(args[1])
	}
	return dict.ensemble.Execute(args, context)
}
func (dict *dictCommand) Help(args []core.Value, options core.CommandHelpOptions, context any) core.Result {
	return dict.ensemble.Help(args, options, context)
}

const DICT_SIZE_SIGNATURE = "dict value size"

type dictSizeCmd struct{}

func (dictSizeCmd) Execute(args []core.Value, _ any) core.Result {
	if len(args) != 2 {
		return ARITY_ERROR(DICT_SIZE_SIGNATURE)
	}
	result := valueToMap(args[1])
	if result.Code != core.ResultCode_OK {
		return result.AsResult()
	}
	map_ := result.Data
	return core.OK(core.INT(int64(len(map_))))
}
func (dictSizeCmd) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	if len(args) > 2 {
		return ARITY_ERROR(DICT_SIZE_SIGNATURE)
	}
	return core.OK(core.STR(DICT_SIZE_SIGNATURE))
}

const DICT_HAS_SIGNATURE = "dict value has key"

type dictHasCmd struct{}

func (dictHasCmd) Execute(args []core.Value, _ any) core.Result {
	if len(args) != 3 {
		return ARITY_ERROR(DICT_HAS_SIGNATURE)
	}
	result := valueToMap(args[1])
	if result.Code != core.ResultCode_OK {
		return result.AsResult()
	}
	map_ := result.Data
	result2 := core.ValueToString(args[2])
	if result2.Code != core.ResultCode_OK {
		return core.ERROR("invalid key")
	}
	key := result2.Data
	return core.OK(core.BOOL(map_[key] != nil))
}
func (dictHasCmd) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	if len(args) > 3 {
		return ARITY_ERROR(DICT_HAS_SIGNATURE)
	}
	return core.OK(core.STR(DICT_HAS_SIGNATURE))
}

const DICT_GET_SIGNATURE = "dict value get key ?default?"

type dictGetCmd struct{}

func (dictGetCmd) Execute(args []core.Value, _ any) core.Result {
	if len(args) != 3 && len(args) != 4 {
		return ARITY_ERROR(DICT_GET_SIGNATURE)
	}
	result := valueToMap(args[1])
	if result.Code != core.ResultCode_OK {
		return result.AsResult()
	}
	map_ := result.Data
	switch args[2].Type() {
	case core.ValueType_TUPLE:
		{
			if len(args) == 4 {
				return core.ERROR("cannot use default with key tuples")
			}
			keys := args[2].(core.TupleValue).Values
			values := []core.Value{}
			for _, k := range keys {
				result := core.ValueToString(k)
				if result.Code != core.ResultCode_OK {
					return core.ERROR("invalid key")
				}
				key := result.Data
				if map_[key] == nil {
					return core.ERROR(`unknown key "` + key + `"`)
				}
				values = append(values, map_[key])
			}
			return core.OK(core.TUPLE(values))
		}
	default:
		{
			result := core.ValueToString(args[2])
			if result.Code != core.ResultCode_OK {
				return core.ERROR("invalid key")
			}
			key := result.Data
			if map_[key] == nil {
				if len(args) == 4 {
					return core.OK(args[3])
				} else {
					return core.ERROR(`unknown key "` + key + `"`)
				}
			}
			return core.OK(map_[key])
		}
	}
}
func (dictGetCmd) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	if len(args) > 4 {
		return ARITY_ERROR(DICT_GET_SIGNATURE)
	}
	return core.OK(core.STR(DICT_GET_SIGNATURE))
}

const DICT_ADD_SIGNATURE = "dict value add key value"

type dictAddCmd struct{}

func (dictAddCmd) Execute(args []core.Value, _ any) core.Result {
	if len(args) != 4 {
		return ARITY_ERROR(DICT_ADD_SIGNATURE)
	}
	result := valueToMap(args[1])
	if result.Code != core.ResultCode_OK {
		return result.AsResult()
	}
	map_ := result.Data
	result2 := core.ValueToString(args[2])
	if result2.Code != core.ResultCode_OK {
		return core.ERROR("invalid key")
	}
	key := result2.Data
	clone := map[string]core.Value{}
	for k, v := range map_ {
		clone[k] = v
	}
	clone[key] = args[3]
	return core.OK(core.DICT(clone))
}
func (dictAddCmd) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	if len(args) > 4 {
		return ARITY_ERROR(DICT_ADD_SIGNATURE)
	}
	return core.OK(core.STR(DICT_ADD_SIGNATURE))
}

const DICT_REMOVE_SIGNATURE = "dict value remove ?key ...?"

type dictRemoveCmd struct{}

func (dictRemoveCmd) Execute(args []core.Value, _ any) core.Result {
	if len(args) == 2 {
		return valueToDictionaryValue(args[1])
	}
	result := valueToMap(args[1])
	if result.Code != core.ResultCode_OK {
		return result.AsResult()
	}
	map_ := result.Data
	clone := map[string]core.Value{}
	for k, v := range map_ {
		clone[k] = v
	}
	for i := 2; i < len(args); i++ {
		result := core.ValueToString(args[i])
		if result.Code != core.ResultCode_OK {
			return core.ERROR("invalid key")
		}
		key := result.Data
		delete(clone, key)
	}
	return core.OK(core.DICT(clone))
}
func (dictRemoveCmd) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	return core.OK(core.STR(DICT_REMOVE_SIGNATURE))
}

const DICT_MERGE_SIGNATURE = "dict value merge ?dict ...?"

type dictMergeCmd struct{}

func (dictMergeCmd) Execute(args []core.Value, _ any) core.Result {
	if len(args) == 2 {
		return valueToDictionaryValue(args[1])
	}
	result := valueToMap(args[1])
	if result.Code != core.ResultCode_OK {
		return result.AsResult()
	}
	map_ := result.Data
	clone := map[string]core.Value{}
	for k, v := range map_ {
		clone[k] = v
	}
	for i := 2; i < len(args); i++ {
		result2 := valueToMap(args[i])
		if result2.Code != core.ResultCode_OK {
			return result2.AsResult()
		}
		map2 := result2.Data
		for key, value := range map2 {
			clone[key] = value
		}
	}
	return core.OK(core.DICT(clone))
}
func (dictMergeCmd) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	return core.OK(core.STR(DICT_MERGE_SIGNATURE))
}

const DICT_KEYS_SIGNATURE = "dict value keys"

type dictKeysCmd struct{}

func (dictKeysCmd) Execute(args []core.Value, _ any) core.Result {
	if len(args) != 2 {
		return ARITY_ERROR(DICT_KEYS_SIGNATURE)
	}
	result := valueToMap(args[1])
	if result.Code != core.ResultCode_OK {
		return result.AsResult()
	}
	map_ := result.Data
	values := make([]core.Value, 0, len(map_))
	for key := range map_ {
		values = append(values, core.STR(key))
	}
	return core.OK(core.LIST(values))
}
func (dictKeysCmd) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	if len(args) > 2 {
		return ARITY_ERROR(DICT_KEYS_SIGNATURE)
	}
	return core.OK(core.STR(DICT_KEYS_SIGNATURE))
}

const DICT_VALUES_SIGNATURE = "dict value values"

type dictValuesCmd struct{}

func (dictValuesCmd) Execute(args []core.Value, _ any) core.Result {
	if len(args) != 2 {
		return ARITY_ERROR(DICT_VALUES_SIGNATURE)
	}
	result := valueToMap(args[1])
	if result.Code != core.ResultCode_OK {
		return result.AsResult()
	}
	map_ := result.Data
	values := make([]core.Value, 0, len(map_))
	for _, value := range map_ {
		values = append(values, value)
	}
	return core.OK(core.LIST(values))
}
func (dictValuesCmd) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	if len(args) > 2 {
		return ARITY_ERROR(DICT_VALUES_SIGNATURE)
	}
	return core.OK(core.STR(DICT_VALUES_SIGNATURE))
}

const DICT_ENTRIES_SIGNATURE = "dict value entries"

type dictEntriesCmd struct{}

func (dictEntriesCmd) Execute(args []core.Value, _ any) core.Result {
	if len(args) != 2 {
		return ARITY_ERROR(DICT_ENTRIES_SIGNATURE)
	}
	result := valueToMap(args[1])
	if result.Code != core.ResultCode_OK {
		return result.AsResult()
	}
	map_ := result.Data
	values := make([]core.Value, 0, len(map_))
	for key, value := range map_ {
		values = append(values, core.TUPLE([]core.Value{core.STR(key), value}))
	}
	return core.OK(core.LIST(values))
}
func (dictEntriesCmd) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	if len(args) > 2 {
		return ARITY_ERROR(DICT_ENTRIES_SIGNATURE)
	}
	return core.OK(core.STR(DICT_ENTRIES_SIGNATURE))
}

const DICT_FOREACH_SIGNATURE = "dict value foreach entry body"

type dictForeachCmd struct{}

func (dictForeachCmd) Execute(args []core.Value, context any) core.Result {
	scope := context.(*Scope)
	if len(args) != 4 {
		return ARITY_ERROR(DICT_FOREACH_SIGNATURE)
	}
	result := valueToMap(args[1])
	if result.Code != core.ResultCode_OK {
		return result.AsResult()
	}
	map_ := result.Data
	varname := args[2]
	body := args[3]
	if body.Type() != core.ValueType_SCRIPT {
		return core.ERROR("body must be a script")
	}
	program := scope.CompileScriptValue(body.(core.ScriptValue))
	subscope := scope.NewLocalScope()
	entries := make([][2]core.Value, len(map_))
	i := 0
	for key, value := range map_ {
		entries[i] = [2]core.Value{core.STR(key), value}
		i++
	}
	i = 0
	lastResult := core.OK(core.NIL)
	var next func() core.Result
	next = func() core.Result {
		if i >= len(entries) {
			return lastResult
		}
		entry := entries[i]
		i++
		result := DestructureValue(
			func(name core.Value, value core.Value, check bool) core.Result {
				return subscope.DestructureLocal(name, value, check)
			},
			varname,
			core.TUPLE(entry[:]),
		)
		if result.Code != core.ResultCode_OK {
			return result
		}
		return CreateContinuationValue(subscope, program, func(result core.Result) core.Result {
			switch result.Code {
			case core.ResultCode_BREAK:
				return lastResult
			case core.ResultCode_CONTINUE:
			case core.ResultCode_OK:
				lastResult = result
			default:
				return result
			}
			return next()
		})
	}
	return next()
}
func (dictForeachCmd) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	if len(args) > 4 {
		return ARITY_ERROR(DICT_FOREACH_SIGNATURE)
	}
	return core.OK(core.STR(DICT_FOREACH_SIGNATURE))
}

func valueToDictionaryValue(value core.Value) core.Result {
	switch value.Type() {
	case core.ValueType_DICTIONARY:
		return core.OK(value)
	case core.ValueType_SCRIPT,
		core.ValueType_LIST,
		core.ValueType_TUPLE:
		{
			result := valueToMap(value)
			if result.Code != core.ResultCode_OK {
				return result.AsResult()
			}
			data := result.Data
			return core.OK(core.DICT(data))
		}
	default:
		return core.ERROR("invalid dictionary")
	}
}
func valueToMap(value core.Value) core.TypedResult[map[string]core.Value] {
	if value.Type() == core.ValueType_DICTIONARY {
		return core.OK_T(core.NIL, value.(core.DictionaryValue).Map)
	}
	result := ValueToArray(value)
	if result.Code != core.ResultCode_OK {
		return core.ResultAs[map[string]core.Value](result.AsResult())
	}
	values := result.Data
	if len(values)%2 != 0 {
		return core.ERROR_T[map[string]core.Value]("invalid key-value list")
	}
	map_ := map[string]core.Value{}
	for i := 0; i < len(values); i += 2 {
		result := core.ValueToString(values[i])
		if result.Code != core.ResultCode_OK {
			return core.ERROR_T[map[string]core.Value]("invalid key")
		}
		key := result.Data
		value := values[i+1]
		map_[key] = value
	}
	return core.OK_T(core.NIL, map_)
}

func DisplayDictionaryValue(
	dictionary core.DictionaryValue,
	fn core.DisplayFunction,
) string {
	values := make([]core.Value, 0, len(dictionary.Map)*2)
	for key, value := range dictionary.Map {
		values = append(values, core.STR(key), value)
	}
	return `[dict (` + core.DisplayList(values, fn) + `)]`
}

func registerDictCommands(scope *Scope) {
	command := newDictCommand(scope)
	scope.RegisterNamedCommand("dict", command)
	command.scope.RegisterNamedCommand("size", dictSizeCmd{})
	command.scope.RegisterNamedCommand("has", dictHasCmd{})
	command.scope.RegisterNamedCommand("get", dictGetCmd{})
	command.scope.RegisterNamedCommand("add", dictAddCmd{})
	command.scope.RegisterNamedCommand("remove", dictRemoveCmd{})
	command.scope.RegisterNamedCommand("merge", dictMergeCmd{})
	command.scope.RegisterNamedCommand("keys", dictKeysCmd{})
	command.scope.RegisterNamedCommand("values", dictValuesCmd{})
	command.scope.RegisterNamedCommand("entries", dictEntriesCmd{})
	command.scope.RegisterNamedCommand("foreach", dictForeachCmd{})
}
