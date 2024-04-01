package helena_dialect

import "helena/core"

func OPERATOR_ARITY_ERROR(operator string) core.Result {
	return core.ERROR(`wrong # operands: should be "operand1 ` + operator + ` operand2"`)
}

var numberSubcommands = NewSubcommands([]string{
	"subcommands",
	"+",
	"-",
	"*",
	"/",
	"==",
	"!=",
	">",
	">=",
	"<",
	"<=",
})

type numberCommand struct{}

func (numberCommand) Execute(args []core.Value, _ any) core.Result {
	result := core.ValueToFloat(args[0])
	if result.Code != core.ResultCode_OK {
		return result.AsResult()
	}
	operand1 := result.Data
	if len(args) == 1 {
		return core.OK(floatToValue(operand1))
	}

	result2 := core.ValueToString(args[1])
	if result2.Code != core.ResultCode_OK {
		return INVALID_SUBCOMMAND_ERROR()
	}
	subcommand := result2.Data
	switch subcommand {
	case "subcommands":
		if len(args) != 2 {
			return ARITY_ERROR("<number> subcommands")
		}
		return core.OK(numberSubcommands.List)

	case "+",
		"-",
		"*",
		"/":
		return arithmetics(args, operand1)

	case "==":
		return eqOp(args, operand1)
	case "!=":
		return neOp(args, operand1)
	case ">":
		return gtOp(args, operand1)
	case ">=":
		return geOp(args, operand1)
	case "<":
		return ltOp(args, operand1)
	case "<=":
		return leOp(args, operand1)

	default:
		return UNKNOWN_SUBCOMMAND_ERROR(subcommand)
	}
}

var numberCmd = numberCommand{}

func arithmetics(args []core.Value, operand1 float64) core.Result {
	if len(args)%2 == 0 {
		return core.ERROR(
			`wrong # operands: should be "operand ?operator operand? ?...?"`,
		)
	}
	total := 0.0
	last := operand1
	for i := 1; i < len(args); i += 2 {
		result := core.ValueToString(args[i])
		if result.Code != core.ResultCode_OK {
			return core.ERROR(`invalid operator`)
		}
		operator := result.Data
		switch operator {
		case "+":
			{
				result := core.ValueToFloat(args[i+1])
				if result.Code != core.ResultCode_OK {
					return result.AsResult()
				}
				operand2 := result.Data
				total += last
				last = operand2
			}
		case "-":
			{
				result := core.ValueToFloat(args[i+1])
				if result.Code != core.ResultCode_OK {
					return result.AsResult()
				}
				operand2 := result.Data
				total += last
				last = -operand2
			}
		case "*":
			{
				result := core.ValueToFloat(args[i+1])
				if result.Code != core.ResultCode_OK {
					return result.AsResult()
				}
				operand2 := result.Data
				last *= operand2
			}
		case "/":
			{
				result := core.ValueToFloat(args[i+1])
				if result.Code != core.ResultCode_OK {
					return result.AsResult()
				}
				operand2 := result.Data
				last /= operand2
			}
		default:
			return core.ERROR(`invalid operator "` + operator + `"`)
		}
	}
	total += last
	return core.OK(floatToValue(total))
}

func binaryOp(
	operator string,
	whenEqual bool,
	fn func(op1 float64, op2 float64) bool,
) func(args []core.Value, operand1 float64) core.Result {
	return func(args []core.Value, operand1 float64) core.Result {
		if len(args) != 3 {
			return OPERATOR_ARITY_ERROR(operator)
		}
		if args[0] == args[2] {
			return core.OK(core.BOOL(whenEqual))
		}
		result := core.ValueToFloat(args[2])
		if result.Code != core.ResultCode_OK {
			return result.AsResult()
		}
		operand2 := result.Data
		return core.OK(core.BOOL(fn(operand1, operand2)))
	}
}

var eqOp = binaryOp("==", true, func(op1 float64, op2 float64) bool { return op1 == op2 })
var neOp = binaryOp("!=", false, func(op1 float64, op2 float64) bool { return op1 != op2 })
var gtOp = binaryOp(">", false, func(op1 float64, op2 float64) bool { return op1 > op2 })
var geOp = binaryOp(">=", true, func(op1 float64, op2 float64) bool { return op1 >= op2 })
var ltOp = binaryOp("<", false, func(op1 float64, op2 float64) bool { return op1 < op2 })
var leOp = binaryOp("<=", true, func(op1 float64, op2 float64) bool { return op1 <= op2 })

func floatToValue(f float64) core.Value {
	i := int64(f)
	if float64(i) == f {
		return core.INT(i)
	} else {
		return core.REAL(f)
	}
}

type intCommand struct {
	scope    *Scope
	ensemble *EnsembleCommand
}

func newIntCommand(scope *Scope) *intCommand {
	cmd := &intCommand{}
	cmd.scope = NewScope(scope, false)
	argspec := ArgspecValueFromValue(core.LIST([]core.Value{core.STR("value")})).Data
	cmd.ensemble = NewEnsembleCommand(cmd.scope, argspec)
	return cmd
}
func (cmd *intCommand) Execute(args []core.Value, context any) core.Result {
	if len(args) == 2 {
		return core.IntegerValueFromValue(args[1]).AsResult()
	}
	return cmd.ensemble.Execute(args, context)
}
func (cmd *intCommand) Help(args []core.Value, options core.CommandHelpOptions, context any) core.Result {
	return cmd.ensemble.Help(args, options, context)
}

type realCommand struct {
	scope    *Scope
	ensemble *EnsembleCommand
}

func newRealCommand(scope *Scope) *realCommand {
	cmd := &realCommand{}
	cmd.scope = NewScope(scope, false)
	argspec := ArgspecValueFromValue(core.LIST([]core.Value{core.STR("value")})).Data
	cmd.ensemble = NewEnsembleCommand(cmd.scope, argspec)
	return cmd
}
func (cmd *realCommand) Execute(args []core.Value, context any) core.Result {
	if len(args) == 2 {
		return core.RealValueFromValue(args[1]).AsResult()
	}
	return cmd.ensemble.Execute(args, context)
}
func (cmd *realCommand) Help(args []core.Value, options core.CommandHelpOptions, context any) core.Result {
	return cmd.ensemble.Help(args, options, context)
}

func registerNumberCommands(scope *Scope) {
	intCommand := newIntCommand(scope)
	scope.RegisterNamedCommand("int", intCommand)
	realCommand := newRealCommand(scope)
	scope.RegisterNamedCommand("real", realCommand)
}
