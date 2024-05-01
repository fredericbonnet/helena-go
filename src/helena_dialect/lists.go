package helena_dialect

import "helena/core"

type listCommand struct {
	scope    *Scope
	ensemble *EnsembleCommand
}

func newListCommand(scope *Scope) *listCommand {
	list := &listCommand{}
	list.scope = NewScope(scope, false)
	argspec := ArgspecValueFromValue(core.LIST([]core.Value{core.STR("value")})).Data
	list.ensemble = NewEnsembleCommand(list.scope, argspec)
	return list
}
func (list *listCommand) Execute(args []core.Value, context any) core.Result {
	if len(args) == 2 {
		return ValueToList(args[1]).AsResult()
	}
	return list.ensemble.Execute(args, context)
}
func (list *listCommand) Help(args []core.Value, options core.CommandHelpOptions, context any) core.Result {
	return list.ensemble.Help(args, options, context)
}

const LIST_LENGTH_SIGNATURE = "list value length"

type listLengthCmd struct{}

func (listLengthCmd) Execute(args []core.Value, _ any) core.Result {
	if len(args) != 2 {
		return ARITY_ERROR(LIST_LENGTH_SIGNATURE)
	}
	result := ValueToArray(args[1])
	if result.Code != core.ResultCode_OK {
		return result.AsResult()
	}
	values := result.Data
	return core.OK(core.INT(int64(len(values))))
}
func (listLengthCmd) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	if len(args) > 2 {
		return ARITY_ERROR(LIST_LENGTH_SIGNATURE)
	}
	return core.OK(core.STR(LIST_LENGTH_SIGNATURE))
}

const LIST_AT_SIGNATURE = "list value at index ?default?"

type listAtCmd struct{}

func (listAtCmd) Execute(args []core.Value, _ any) core.Result {
	if len(args) != 3 && len(args) != 4 {
		return ARITY_ERROR(LIST_AT_SIGNATURE)
	}
	result := ValueToArray(args[1])
	if result.Code != core.ResultCode_OK {
		return result.AsResult()
	}
	values := result.Data
	if len(args) == 4 {
		return core.ListAtOrDefault(values, args[2], args[3])
	} else {
		return core.ListAt(values, args[2])
	}
}
func (listAtCmd) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	if len(args) > 4 {
		return ARITY_ERROR(LIST_AT_SIGNATURE)
	}
	return core.OK(core.STR(LIST_AT_SIGNATURE))
}

const LIST_RANGE_SIGNATURE = "list value range first ?last?"

type listRangeCmd struct{}

func (listRangeCmd) Execute(args []core.Value, _ any) core.Result {
	if len(args) != 3 && len(args) != 4 {
		return ARITY_ERROR(LIST_RANGE_SIGNATURE)
	}
	result := ValueToArray(args[1])
	if result.Code != core.ResultCode_OK {
		return result.AsResult()
	}
	values := result.Data
	length := int64(len(values))
	firstResult := core.ValueToInteger(args[2])
	if firstResult.Code != core.ResultCode_OK {
		return firstResult.AsResult()
	}
	first := max(0, firstResult.Data)
	if len(args) == 3 {
		if first >= length {
			return core.OK(core.LIST([]core.Value{}))
		}
		return core.OK(core.LIST(values[first:]))
	} else {
		lastResult := core.ValueToInteger(args[3])
		if lastResult.Code != core.ResultCode_OK {
			return lastResult.AsResult()
		}
		last := lastResult.Data
		if first >= length || last < first || last < 0 {
			return core.OK(core.LIST([]core.Value{}))
		}
		return core.OK(core.LIST(values[first:min(last+1, length)]))
	}
}
func (listRangeCmd) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	if len(args) > 4 {
		return ARITY_ERROR(LIST_RANGE_SIGNATURE)
	}
	return core.OK(core.STR(LIST_RANGE_SIGNATURE))
}

const LIST_APPEND_SIGNATURE = "list value append ?list ...?"

type listAppendCmd struct{}

func (listAppendCmd) Execute(args []core.Value, _ any) core.Result {
	result := ValueToArray(args[1])
	if result.Code != core.ResultCode_OK {
		return result.AsResult()
	}
	values := result.Data
	values2 := append([]core.Value{}, values...)
	for i := 2; i < len(args); i++ {
		result := ValueToArray(args[i])
		if result.Code != core.ResultCode_OK {
			return result.AsResult()
		}
		values := result.Data
		values2 = append(values2, values...)
	}
	return core.OK(core.LIST(values2))
}
func (listAppendCmd) Help(_ []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	return core.OK(core.STR(LIST_APPEND_SIGNATURE))
}

const LIST_REMOVE_SIGNATURE = "list value remove first last"

type listRemoveCmd struct{}

func (listRemoveCmd) Execute(args []core.Value, _ any) core.Result {
	if len(args) != 4 && len(args) != 5 {
		return ARITY_ERROR(LIST_REMOVE_SIGNATURE)
	}
	result := ValueToArray(args[1])
	if result.Code != core.ResultCode_OK {
		return result.AsResult()
	}
	values := result.Data
	length := int64(len(values))
	firstResult := core.ValueToInteger(args[2])
	if firstResult.Code != core.ResultCode_OK {
		return firstResult.AsResult()
	}
	first := max(0, firstResult.Data)
	lastResult := core.ValueToInteger(args[3])
	if lastResult.Code != core.ResultCode_OK {
		return lastResult.AsResult()
	}
	last := lastResult.Data
	head := values[0:min(first, length)]
	tail := values[min(max(first, last+1), length):]
	return core.OK(core.LIST(append(append([]core.Value{}, head...), tail...)))
}
func (listRemoveCmd) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	if len(args) > 4 {
		return ARITY_ERROR(LIST_REMOVE_SIGNATURE)
	}
	return core.OK(core.STR(LIST_REMOVE_SIGNATURE))
}

const LIST_INSERT_SIGNATURE = "list value insert index value2"

type listInsertCmd struct{}

func (listInsertCmd) Execute(args []core.Value, _ any) core.Result {
	if len(args) != 4 {
		return ARITY_ERROR(LIST_INSERT_SIGNATURE)
	}
	result := ValueToArray(args[1])
	if result.Code != core.ResultCode_OK {
		return result.AsResult()
	}
	values := result.Data
	length := int64(len(values))
	indexResult := core.ValueToInteger(args[2])
	if indexResult.Code != core.ResultCode_OK {
		return indexResult.AsResult()
	}
	index := max(0, indexResult.Data)
	result2 := ValueToArray(args[3])
	if result2.Code != core.ResultCode_OK {
		return result2.AsResult()
	}
	insert := result2.Data
	head := values[0:min(index, length)]
	tail := values[min(index, length):]
	return core.OK(core.LIST(append(append(append([]core.Value{}, head...), insert...), tail...)))
}
func (listInsertCmd) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	if len(args) > 4 {
		return ARITY_ERROR(LIST_INSERT_SIGNATURE)
	}
	return core.OK(core.STR(LIST_INSERT_SIGNATURE))
}

const LIST_REPLACE_SIGNATURE = "list value replace first last value2"

type listReplaceCmd struct{}

func (listReplaceCmd) Execute(args []core.Value, _ any) core.Result {
	if len(args) != 5 {
		return ARITY_ERROR(LIST_REPLACE_SIGNATURE)
	}
	result := ValueToArray(args[1])
	if result.Code != core.ResultCode_OK {
		return result.AsResult()
	}
	values := result.Data
	length := int64(len(values))
	firstResult := core.ValueToInteger(args[2])
	if firstResult.Code != core.ResultCode_OK {
		return firstResult.AsResult()
	}
	first := max(0, firstResult.Data)
	lastResult := core.ValueToInteger(args[3])
	if lastResult.Code != core.ResultCode_OK {
		return lastResult.AsResult()
	}
	last := lastResult.Data
	head := values[0:min(first, length)]
	tail := values[min(max(first, last+1), length):]
	result2 := ValueToArray(args[4])
	if result2.Code != core.ResultCode_OK {
		return result2.AsResult()
	}
	insert := result2.Data
	return core.OK(core.LIST(append(append(append([]core.Value{}, head...), insert...), tail...)))
}
func (listReplaceCmd) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	if len(args) > 5 {
		return ARITY_ERROR(LIST_REPLACE_SIGNATURE)
	}
	return core.OK(core.STR(LIST_REPLACE_SIGNATURE))
}

const LIST_FOREACH_SIGNATURE = "list value foreach element body"

type listForeachCmd struct{}

func (listForeachCmd) Execute(args []core.Value, context any) core.Result {
	scope := context.(*Scope)
	if len(args) != 4 {
		return ARITY_ERROR(LIST_FOREACH_SIGNATURE)
	}
	result := ValueToList(args[1])
	if result.Code != core.ResultCode_OK {
		return result.AsResult()
	}
	list := result.Data
	varname := args[2]
	body := args[3]
	if body.Type() != core.ValueType_SCRIPT {
		return core.ERROR("body must be a script")
	}
	program := scope.Compile(body.(core.ScriptValue).Script)
	subscope := NewScope(scope, true)
	i := 0
	lastResult := core.OK(core.NIL)
	var next func() core.Result
	next = func() core.Result {
		if i >= len(list.Values) {
			return lastResult
		}
		value := list.Values[i]
		i++
		result := DestructureValue(
			func(name core.Value, value core.Value, check bool) core.Result {
				return subscope.DestructureLocal(name, value, check)
			},
			varname,
			value,
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
func (listForeachCmd) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	if len(args) > 4 {
		return ARITY_ERROR(LIST_FOREACH_SIGNATURE)
	}
	return core.OK(core.STR(LIST_FOREACH_SIGNATURE))
}

func ValueToList(value core.Value) core.TypedResult[core.ListValue] {
	if value.Type() == core.ValueType_SCRIPT {
		result := ValueToArray(value)
		if result.Code != core.ResultCode_OK {
			return core.ResultAs[core.ListValue](result.AsResult())
		}
		data := result.Data
		list := core.LIST(data)
		return core.OK_T(list, list)
	}
	return core.ListValueFromValue(value)
}

func ValueToArray(value core.Value) core.TypedResult[[]core.Value] {
	if value.Type() == core.ValueType_SCRIPT {
		program := core.Compiler{}.CompileSentences(
			value.(core.ScriptValue).Script.Sentences,
		)
		listExecutor := core.Executor{}
		result := listExecutor.Execute(program, nil)
		if result.Code != core.ResultCode_OK {
			return core.ERROR_T[[]core.Value]("invalid list")
		}
		return core.OK_T(core.NIL, result.Value.(core.TupleValue).Values)
	}
	return core.ValueToValues(value)
}

func DisplayListValue(
	list core.ListValue,
	fn core.DisplayFunction,
) string {
	return `[list (` + core.DisplayList(list.Values, fn) + `)]`
}

func registerListCommands(scope *Scope) {
	command := newListCommand(scope)
	scope.RegisterNamedCommand("list", command)
	command.scope.RegisterNamedCommand("length", listLengthCmd{})
	command.scope.RegisterNamedCommand("at", listAtCmd{})
	command.scope.RegisterNamedCommand("range", listRangeCmd{})
	command.scope.RegisterNamedCommand("append", listAppendCmd{})
	command.scope.RegisterNamedCommand("remove", listRemoveCmd{})
	command.scope.RegisterNamedCommand("insert", listInsertCmd{})
	command.scope.RegisterNamedCommand("replace", listReplaceCmd{})
	command.scope.RegisterNamedCommand("foreach", listForeachCmd{})
}
