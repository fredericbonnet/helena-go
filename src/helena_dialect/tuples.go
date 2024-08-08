package helena_dialect

import "helena/core"

type tupleCommand struct {
	scope    *Scope
	ensemble *EnsembleCommand
}

func newTupleCommand(scope *Scope) *tupleCommand {
	tuple := &tupleCommand{}
	tuple.scope = scope.NewChildScope()
	_, argspec := ArgspecValueFromValue(core.LIST([]core.Value{core.STR("value")}))
	tuple.ensemble = NewEnsembleCommand(tuple.scope, argspec)
	return tuple
}
func (tuple *tupleCommand) Execute(args []core.Value, context any) core.Result {
	if len(args) == 2 {
		return ValueToTuple(args[1])
	}
	return tuple.ensemble.Execute(args, context)
}
func (tuple *tupleCommand) Help(args []core.Value, options core.CommandHelpOptions, context any) core.Result {
	return tuple.ensemble.Help(args, options, context)
}

const TUPLE_LENGTH_SIGNATURE = "tuple value length"

type tupleLengthCmd struct{}

func (tupleLengthCmd) Execute(args []core.Value, _ any) core.Result {
	if len(args) != 2 {
		return ARITY_ERROR(TUPLE_LENGTH_SIGNATURE)
	}
	result, values := ValueToArray(args[1])
	if result.Code != core.ResultCode_OK {
		return result
	}
	return core.OK(core.INT(int64(len(values))))
}
func (tupleLengthCmd) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	if len(args) > 2 {
		return ARITY_ERROR(TUPLE_LENGTH_SIGNATURE)
	}
	return core.OK(core.STR(TUPLE_LENGTH_SIGNATURE))
}

const TUPLE_AT_SIGNATURE = "tuple value at index ?default?"

type tupleAtCmd struct{}

func (tupleAtCmd) Execute(args []core.Value, _ any) core.Result {
	if len(args) != 3 && len(args) != 4 {
		return ARITY_ERROR(TUPLE_AT_SIGNATURE)
	}
	result, values := ValueToArray(args[1])
	if result.Code != core.ResultCode_OK {
		return result
	}
	if len(args) == 4 {
		return core.ListAtOrDefault(values, args[2], args[3])
	} else {
		return core.ListAt(values, args[2])
	}
}
func (tupleAtCmd) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	if len(args) > 4 {
		return ARITY_ERROR(TUPLE_AT_SIGNATURE)
	}
	return core.OK(core.STR(TUPLE_AT_SIGNATURE))
}

func ValueToTuple(value core.Value) core.Result {
	switch value.Type() {
	case core.ValueType_TUPLE:
		return core.OK(value)
	case core.ValueType_LIST,
		core.ValueType_SCRIPT:
		{
			result, values := ValueToArray(value)
			if result.Code != core.ResultCode_OK {
				return result
			}
			return core.OK(core.TUPLE(values))
		}
	default:
		return core.ERROR("invalid tuple")
	}
}

func registerTupleCommands(scope *Scope) {
	command := newTupleCommand(scope)
	scope.RegisterNamedCommand("tuple", command)
	command.scope.RegisterNamedCommand("length", tupleLengthCmd{})
	command.scope.RegisterNamedCommand("at", tupleAtCmd{})
}
