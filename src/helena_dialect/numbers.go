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
	return numberSubcommands.Dispatch(args[1], SubcommandHandlers{
		"subcommands": func() core.Result {
			if len(args) != 2 {
				return ARITY_ERROR("<number> subcommands")
			}
			return core.OK(numberSubcommands.List)
		},

		"+": func() core.Result { return arithmetics(args, operand1) },
		"-": func() core.Result { return arithmetics(args, operand1) },
		"*": func() core.Result { return arithmetics(args, operand1) },
		"/": func() core.Result { return arithmetics(args, operand1) },

		"==": func() core.Result { return eqOp(args, operand1) },
		"!=": func() core.Result { return neOp(args, operand1) },
		">":  func() core.Result { return gtOp(args, operand1) },
		">=": func() core.Result { return geOp(args, operand1) },
		"<":  func() core.Result { return ltOp(args, operand1) },
		"<=": func() core.Result { return leOp(args, operand1) },
	})
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

// class IntCommand implements Command {
//   scope: Scope;
//   ensemble: EnsembleCommand;
//   constructor(scope: Scope) {
//     this.scope = new Scope(scope);
//     const { data: argspec } = ArgspecValue.fromValue(LIST([STR("value")]));
//     this.ensemble = new EnsembleCommand(this.scope, argspec);
//   }
//   execute(args: Value[], scope: Scope): Result {
//     if (len(args) == 2) return IntegerValue.fromValue(args[1]);
//     return this.ensemble.execute(args, scope);
//   }
//   help(args) {
//     return this.ensemble.help(args);
//   }
// }

// class RealCommand implements Command {
//   scope: Scope;
//   ensemble: EnsembleCommand;
//   constructor(scope: Scope) {
//     this.scope = new Scope(scope);
//     const { data: argspec } = ArgspecValue.fromValue(LIST([STR("value")]));
//     this.ensemble = new EnsembleCommand(this.scope, argspec);
//   }
//   execute(args: Value[], scope: Scope): Result {
//     if (len(args) == 2) return RealValue.fromValue(args[1]);
//     return this.ensemble.execute(args, scope);
//   }
//   help(args) {
//     return this.ensemble.help(args);
//   }
// }

// export function registerNumberCommands(scope: Scope) {
//   const intCommand = new IntCommand(scope);
//   scope.registerNamedCommand("int", intCommand);
//   const realCommand = new RealCommand(scope);
//   scope.registerNamedCommand("real", realCommand);
// }
