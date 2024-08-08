package helena_dialect

import "helena/core"

const ADD_SIGNATURE = "+ number ?number ...?"

type addCmd struct{}

func (addCmd) Execute(args []core.Value, _ any) core.Result {
	if len(args) < 2 {
		return ARITY_ERROR(ADD_SIGNATURE)
	}
	total := 0.0
	for i := 1; i < len(args); i++ {
		result, operand := core.ValueToFloat(args[i])
		if result.Code != core.ResultCode_OK {
			return result
		}
		total += operand
	}
	return core.OK(floatToValue(total))
}
func (addCmd) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	return core.OK(core.STR(ADD_SIGNATURE))
}

const SUBTRACT_SIGNATURE = "- number ?number ...?"

type subtractCmd struct{}

func (subtractCmd) Execute(args []core.Value, _ any) core.Result {
	if len(args) < 2 {
		return ARITY_ERROR(SUBTRACT_SIGNATURE)
	}
	result, first := core.ValueToFloat(args[1])
	if result.Code != core.ResultCode_OK {
		return result
	}
	if len(args) == 2 {
		return core.OK(floatToValue(-first))
	}
	total := first
	for i := 2; i < len(args); i++ {
		result, f := core.ValueToFloat(args[i])
		if result.Code != core.ResultCode_OK {
			return result
		}
		total -= f
	}
	return core.OK(floatToValue(total))
}
func (subtractCmd) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	return core.OK(core.STR(SUBTRACT_SIGNATURE))
}

const MULTIPLY_SIGNATURE = "* number ?number ...?"

type multiplyCmd struct{}

func (multiplyCmd) Execute(args []core.Value, _ any) core.Result {
	if len(args) < 2 {
		return ARITY_ERROR(MULTIPLY_SIGNATURE)
	}
	result, first := core.ValueToFloat(args[1])
	if result.Code != core.ResultCode_OK {
		return result
	}
	if len(args) == 2 {
		return core.OK(floatToValue(first))
	}
	total := first
	for i := 2; i < len(args); i++ {
		result, f := core.ValueToFloat(args[i])
		if result.Code != core.ResultCode_OK {
			return result
		}
		total *= f
	}
	return core.OK(floatToValue(total))
}
func (multiplyCmd) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	return core.OK(core.STR(MULTIPLY_SIGNATURE))
}

const DIVIDE_SIGNATURE = "/ number number ?number ...?"

type divideCmd struct{}

func (divideCmd) Execute(args []core.Value, _ any) core.Result {
	if len(args) < 3 {
		return ARITY_ERROR(DIVIDE_SIGNATURE)
	}
	result, first := core.ValueToFloat(args[1])
	if result.Code != core.ResultCode_OK {
		return result
	}
	total := first
	for i := 2; i < len(args); i++ {
		result, f := core.ValueToFloat(args[i])
		if result.Code != core.ResultCode_OK {
			return result
		}
		total /= f
	}
	return core.OK(floatToValue(total))
}
func (divideCmd) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	return core.OK(core.STR(DIVIDE_SIGNATURE))
}
func registerMathCommands(scope *Scope) {
	scope.RegisterNamedCommand("+", addCmd{})
	scope.RegisterNamedCommand("-", subtractCmd{})
	scope.RegisterNamedCommand("*", multiplyCmd{})
	scope.RegisterNamedCommand("/", divideCmd{})
}
