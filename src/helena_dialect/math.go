package helena_dialect

import "helena/core"

const ADD_SIGNATURE = "+ number ?number ...?"

type AddCommand struct{}

func (AddCommand) Execute(args []core.Value, _ any) core.Result {
	if len(args) < 2 {
		return ARITY_ERROR(ADD_SIGNATURE)
	}
	total := 0.0
	for i := 1; i < len(args); i++ {
		result := core.ValueToFloat(args[i])
		if result.Code != core.ResultCode_OK {
			return result.AsResult()
		}
		operand := result.Data
		total += operand
	}
	return core.OK(floatToValue(total))
}
func (AddCommand) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	return core.OK(core.STR(ADD_SIGNATURE))
}

const SUBTRACT_SIGNATURE = "- number ?number ...?"

type SubtractCommand struct{}

func (SubtractCommand) Execute(args []core.Value, _ any) core.Result {
	if len(args) < 2 {
		return ARITY_ERROR(SUBTRACT_SIGNATURE)
	}
	result := core.ValueToFloat(args[1])
	if result.Code != core.ResultCode_OK {
		return result.AsResult()
	}
	first := result.Data
	if len(args) == 2 {
		return core.OK(floatToValue(-first))
	}
	total := first
	for i := 2; i < len(args); i++ {
		result := core.ValueToFloat(args[i])
		if result.Code != core.ResultCode_OK {
			return result.AsResult()
		}
		total -= result.Data
	}
	return core.OK(floatToValue(total))
}
func (SubtractCommand) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	return core.OK(core.STR(SUBTRACT_SIGNATURE))
}

const MULTIPLY_SIGNATURE = "* number ?number ...?"

type MultiplyCommand struct{}

func (MultiplyCommand) Execute(args []core.Value, _ any) core.Result {
	if len(args) < 2 {
		return ARITY_ERROR(MULTIPLY_SIGNATURE)
	}
	result := core.ValueToFloat(args[1])
	if result.Code != core.ResultCode_OK {
		return result.AsResult()
	}
	first := result.Data
	if len(args) == 2 {
		return core.OK(floatToValue(first))
	}
	total := first
	for i := 2; i < len(args); i++ {
		result := core.ValueToFloat(args[i])
		if result.Code != core.ResultCode_OK {
			return result.AsResult()
		}
		total *= result.Data
	}
	return core.OK(floatToValue(total))
}
func (MultiplyCommand) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	return core.OK(core.STR(MULTIPLY_SIGNATURE))
}

const DIVIDE_SIGNATURE = "/ number number ?number ...?"

type DivideCommand struct{}

func (DivideCommand) Execute(args []core.Value, _ any) core.Result {
	if len(args) < 3 {
		return ARITY_ERROR(DIVIDE_SIGNATURE)
	}
	result := core.ValueToFloat(args[1])
	if result.Code != core.ResultCode_OK {
		return result.AsResult()
	}
	first := result.Data
	total := first
	for i := 2; i < len(args); i++ {
		result := core.ValueToFloat(args[i])
		if result.Code != core.ResultCode_OK {
			return result.AsResult()
		}
		total /= result.Data
	}
	return core.OK(floatToValue(total))
}
func (DivideCommand) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	return core.OK(core.STR(DIVIDE_SIGNATURE))
}
func registerMathCommands(scope *Scope) {
	scope.RegisterNamedCommand("+", AddCommand{})
	scope.RegisterNamedCommand("-", SubtractCommand{})
	scope.RegisterNamedCommand("*", MultiplyCommand{})
	scope.RegisterNamedCommand("/", DivideCommand{})
}
