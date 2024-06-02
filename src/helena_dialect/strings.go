package helena_dialect

import "helena/core"

type stringCommand struct {
	scope    *Scope
	ensemble *EnsembleCommand
}

func newStringCommand(scope *Scope) *stringCommand {
	cmd := &stringCommand{}
	cmd.scope = scope.NewChildScope()
	argspec := ArgspecValueFromValue(core.LIST([]core.Value{core.STR("value")})).Data
	cmd.ensemble = NewEnsembleCommand(cmd.scope, argspec)
	return cmd
}
func (cmd *stringCommand) Execute(args []core.Value, context any) core.Result {
	if len(args) == 2 {
		return core.StringValueFromValue(args[1]).AsResult()
	}
	return cmd.ensemble.Execute(args, context)
}
func (cmd *stringCommand) Help(args []core.Value, options core.CommandHelpOptions, context any) core.Result {
	return cmd.ensemble.Help(args, options, context)
}

const STRING_LENGTH_SIGNATURE = "string value length"

type stringLengthCmd struct{}

func (stringLengthCmd) Execute(args []core.Value, _ any) core.Result {
	if len(args) != 2 {
		return ARITY_ERROR(STRING_LENGTH_SIGNATURE)
	}
	result := core.ValueToString(args[1])
	if result.Code != core.ResultCode_OK {
		return result.AsResult()
	}
	str := result.Data
	return core.OK(core.INT(int64(len(str))))
}
func (stringLengthCmd) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	if len(args) > 2 {
		return ARITY_ERROR(STRING_LENGTH_SIGNATURE)
	}
	return core.OK(core.STR(STRING_LENGTH_SIGNATURE))
}

const STRING_AT_SIGNATURE = "string value at index ?default?"

type stringAtCmd struct{}

func (stringAtCmd) Execute(args []core.Value, _ any) core.Result {
	if len(args) != 3 && len(args) != 4 {
		return ARITY_ERROR(STRING_AT_SIGNATURE)
	}
	result := core.ValueToString(args[1])
	if result.Code != core.ResultCode_OK {
		return result.AsResult()
	}
	str := result.Data
	if len(args) == 4 {
		return core.StringAtOrDefault(str, args[2], args[3])
	} else {
		return core.StringAt(str, args[2])
	}
}
func (stringAtCmd) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	if len(args) > 4 {
		return ARITY_ERROR(STRING_AT_SIGNATURE)
	}
	return core.OK(core.STR(STRING_AT_SIGNATURE))
}

const STRING_RANGE_SIGNATURE = "string value range first ?last?"

type stringRangeCmd struct{}

func (stringRangeCmd) Execute(args []core.Value, _ any) core.Result {
	if len(args) != 3 && len(args) != 4 {
		return ARITY_ERROR(STRING_RANGE_SIGNATURE)
	}
	result := core.ValueToString(args[1])
	if result.Code != core.ResultCode_OK {
		return result.AsResult()
	}
	str := result.Data
	length := int64(len(str))
	firstResult := core.ValueToInteger(args[2])
	if firstResult.Code != core.ResultCode_OK {
		return firstResult.AsResult()
	}
	first := max(0, firstResult.Data)
	if len(args) == 3 {
		if first >= length {
			return core.OK(core.STR(""))
		}
		return core.OK(core.STR(str[first:]))
	} else {
		lastResult := core.ValueToInteger(args[3])
		if lastResult.Code != core.ResultCode_OK {
			return lastResult.AsResult()
		}
		last := lastResult.Data
		if first >= length || last < first || last < 0 {
			return core.OK(core.STR(""))
		}
		return core.OK(core.STR(str[first:min(last+1, length)]))
	}
}
func (stringRangeCmd) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	if len(args) > 4 {
		return ARITY_ERROR(STRING_RANGE_SIGNATURE)
	}
	return core.OK(core.STR(STRING_RANGE_SIGNATURE))
}

const STRING_APPEND_SIGNATURE = "string value append ?string ...?"

type stringAppendCmd struct{}

func (stringAppendCmd) Execute(args []core.Value, _ any) core.Result {
	result := core.ValueToString(args[1])
	if result.Code != core.ResultCode_OK {
		return result.AsResult()
	}
	str := result.Data
	str2 := str
	for i := 2; i < len(args); i++ {
		result := core.ValueToString(args[i])
		if result.Code != core.ResultCode_OK {
			return result.AsResult()
		}
		append := result.Data
		str2 += append
	}
	return core.OK(core.STR(str2))
}
func (stringAppendCmd) Help(_ []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	return core.OK(core.STR(STRING_APPEND_SIGNATURE))
}

const STRING_REMOVE_SIGNATURE = "string value remove first last"

type stringRemoveCmd struct{}

func (stringRemoveCmd) Execute(args []core.Value, _ any) core.Result {
	if len(args) != 4 && len(args) != 5 {
		return ARITY_ERROR(STRING_REMOVE_SIGNATURE)
	}
	result := core.ValueToString(args[1])
	if result.Code != core.ResultCode_OK {
		return result.AsResult()
	}
	str := result.Data
	length := int64(len(str))
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
	head := str[0:min(first, length)]
	tail := str[min(max(first, last+1), length):]
	return core.OK(core.STR(head + tail))
}
func (stringRemoveCmd) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	if len(args) > 5 {
		return ARITY_ERROR(STRING_REMOVE_SIGNATURE)
	}
	return core.OK(core.STR(STRING_REMOVE_SIGNATURE))
}

const STRING_INSERT_SIGNATURE = "string value insert index value2"

type stringInsertCmd struct{}

func (stringInsertCmd) Execute(args []core.Value, _ any) core.Result {
	if len(args) != 4 {
		return ARITY_ERROR(STRING_INSERT_SIGNATURE)
	}
	result := core.ValueToString(args[1])
	if result.Code != core.ResultCode_OK {
		return result.AsResult()
	}
	str := result.Data
	length := int64(len(str))
	indexResult := core.ValueToInteger(args[2])
	if indexResult.Code != core.ResultCode_OK {
		return indexResult.AsResult()
	}
	index := max(0, indexResult.Data)
	result2 := core.ValueToString(args[3])
	if result2.Code != core.ResultCode_OK {
		return result2.AsResult()
	}
	insert := result2.Data
	head := str[0:min(index, length)]
	tail := str[min(index, length):]
	return core.OK(core.STR(head + insert + tail))
}
func (stringInsertCmd) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	if len(args) > 4 {
		return ARITY_ERROR(STRING_INSERT_SIGNATURE)
	}
	return core.OK(core.STR(STRING_INSERT_SIGNATURE))
}

const STRING_REPLACE_SIGNATURE = "string value replace first last value2"

type stringReplaceCmd struct{}

func (stringReplaceCmd) Execute(args []core.Value, _ any) core.Result {
	if len(args) != 5 {
		return ARITY_ERROR(STRING_REPLACE_SIGNATURE)
	}
	result := core.ValueToString(args[1])
	if result.Code != core.ResultCode_OK {
		return result.AsResult()
	}
	str := result.Data
	length := int64(len(str))
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
	head := str[0:min(first, length)]
	tail := str[min(max(first, last+1), length):]
	result2 := core.ValueToString(args[4])
	if result2.Code != core.ResultCode_OK {
		return result2.AsResult()
	}
	insert := result2.Data
	return core.OK(core.STR(head + insert + tail))
}

func (stringReplaceCmd) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	if len(args) > 5 {
		return ARITY_ERROR(STRING_REPLACE_SIGNATURE)
	}
	return core.OK(core.STR(STRING_REPLACE_SIGNATURE))
}

func STRING_OPERATOR_SIGNATURE(operator string) string {
	return `string value1 ` + operator + ` value2`
}

func STRING_OPERATOR_ARITY_ERROR(operator string) core.Result {
	return core.ERROR(`wrong # operands: should be "` + STRING_OPERATOR_SIGNATURE(operator) + `"`)
}

type binaryCmd struct {
	name      string
	whenEqual bool
	fn        func(op1 string, op2 string) bool
}

func (cmd binaryCmd) Execute(args []core.Value, context any) core.Result {
	if len(args) != 3 {
		return STRING_OPERATOR_ARITY_ERROR(cmd.name)
	}
	if args[1] == args[2] {
		return core.OK(core.BOOL(cmd.whenEqual))
	}
	result1 := core.ValueToString(args[1])
	if result1.Code != core.ResultCode_OK {
		return result1.AsResult()
	}
	operand1 := result1.Data
	result2 := core.ValueToString(args[2])
	if result2.Code != core.ResultCode_OK {
		return result2.AsResult()
	}
	operand2 := result2.Data
	return core.OK(core.BOOL(cmd.fn(operand1, operand2)))
}
func (cmd binaryCmd) Help(args []core.Value, options core.CommandHelpOptions, context any) core.Result {
	if len(args) > 3 {
		return STRING_OPERATOR_ARITY_ERROR(cmd.name)
	}
	return core.OK(core.STR(STRING_OPERATOR_SIGNATURE(cmd.name)))
}

var eqCmd = binaryCmd{"==", true, func(op1 string, op2 string) bool { return op1 == op2 }}
var neCmd = binaryCmd{"!=", false, func(op1 string, op2 string) bool { return op1 != op2 }}
var gtCmd = binaryCmd{">", false, func(op1 string, op2 string) bool { return op1 > op2 }}
var geCmd = binaryCmd{">=", true, func(op1 string, op2 string) bool { return op1 >= op2 }}
var ltCmd = binaryCmd{"<", false, func(op1 string, op2 string) bool { return op1 < op2 }}
var leCmd = binaryCmd{"<=", true, func(op1 string, op2 string) bool { return op1 <= op2 }}

func registerStringCommands(scope *Scope) {
	command := newStringCommand(scope)
	scope.RegisterNamedCommand("string", command)
	command.scope.RegisterNamedCommand("length", stringLengthCmd{})
	command.scope.RegisterNamedCommand("at", stringAtCmd{})
	command.scope.RegisterNamedCommand("range", stringRangeCmd{})
	command.scope.RegisterNamedCommand("append", stringAppendCmd{})
	command.scope.RegisterNamedCommand("remove", stringRemoveCmd{})
	command.scope.RegisterNamedCommand("insert", stringInsertCmd{})
	command.scope.RegisterNamedCommand("replace", stringReplaceCmd{})
	command.scope.RegisterNamedCommand("==", eqCmd)
	command.scope.RegisterNamedCommand("!=", neCmd)
	command.scope.RegisterNamedCommand(">", gtCmd)
	command.scope.RegisterNamedCommand(">=", geCmd)
	command.scope.RegisterNamedCommand("<", ltCmd)
	command.scope.RegisterNamedCommand("<=", leCmd)
}
