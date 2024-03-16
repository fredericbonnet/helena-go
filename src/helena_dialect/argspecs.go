package helena_dialect

import "helena/core"

// /* eslint-disable jsdoc/require-jsdoc */ // TODO
// import { Result, OK, ERROR, ResultCode } from "../core/results";
// import { Command } from "../core/command";
// import {
//   Value,
//   ValueType,
//   ScriptValue,
//   NIL,
//   STR,
//   TUPLE,
//   LIST,
// } from "../core/values";
// import { Argument, ARITY_ERROR, buildArguments, buildUsage } from "./arguments";
// import { Scope } from "./core";
// import { valueToArray } from "./lists";
// import { EnsembleCommand } from "./ensembles";

type Argspec struct {
	Args         []Argument
	NbRequired   uint
	NbOptional   uint
	HasRemainder bool
}

func NewArgspec(args []Argument) Argspec {
	nbRequired := uint(0)
	nbOptional := uint(0)
	hasRemainder := false
	for _, arg := range args {
		switch arg.Type {
		case ArgumentType_REQUIRED:
			nbRequired++
		case ArgumentType_OPTIONAL:
			nbOptional++
		case ArgumentType_REMAINDER:
			hasRemainder = true
		}
	}
	return Argspec{args, nbRequired, nbOptional, hasRemainder}
}
func (argspec Argspec) IsVariadic() bool {
	return (argspec.NbOptional > 0) || argspec.HasRemainder
}

type ArgspecValue struct {
	//   readonly type = { name: "argspec" };
	Argspec Argspec
}

func (ArgspecValue) Type() core.ValueType {
	return -1
}
func NewArgspecValue(argspec Argspec) ArgspecValue {
	return ArgspecValue{argspec}
}
func ArgspecValueFromValue(value core.Value) core.TypedResult[ArgspecValue] {
	if v, ok := value.(ArgspecValue); ok {
		return core.OK_T(v, v)
	}
	result := buildArguments(value)
	if result.Code != core.ResultCode_OK {
		return core.ResultAs[ArgspecValue](result.AsResult())
	}
	args := result.Data
	v := NewArgspecValue(NewArgspec(args))
	return core.OK_T(v, v)
}
func (value ArgspecValue) Usage(skip uint) string {
	return BuildUsage(value.Argspec.Args, skip)
}

//   checkArity(values: Value[], skip: number) {
//     return (
//       values.length - skip >= this.argspec.nbRequired &&
//       (this.argspec.hasRemainder ||
//         values.length - skip <=
//           this.argspec.nbRequired + this.argspec.nbOptional)
//     );
//   }
func (value ArgspecValue) ApplyArguments(
	scope *Scope,
	values []core.Value,
	skip uint,
	setArgument func(name string, value core.Value) core.Result,
) core.Result {
	nonRequired := uint(len(values)) - skip - value.Argspec.NbRequired
	optionals := min(value.Argspec.NbOptional, nonRequired)
	remainders := nonRequired - optionals
	i := skip
	for _, arg := range value.Argspec.Args {
		var value core.Value
		switch arg.Type {
		case ArgumentType_REQUIRED:
			value = values[i]
			i++
		case ArgumentType_OPTIONAL:
			if optionals > 0 {
				optionals--
				value = values[i]
				i++
			} else if arg.Default != nil {
				if arg.Default.Type() == core.ValueType_SCRIPT {
					body := arg.Default.(core.ScriptValue)
					result := scope.ExecuteScriptValue(body)
					// TODO handle YIELD?
					if result.Code != core.ResultCode_OK {
						return result
					}
					value = result.Value
				} else {
					value = arg.Default
				}
			} else {
				continue // Skip missing optional
			}
		case ArgumentType_REMAINDER:
			value = core.TUPLE(values[i : i+remainders])
			i += remainders
		}
		if arg.Guard != nil {
			process := scope.PrepareTupleValue(core.TUPLE([]core.Value{arg.Guard, value}).(core.TupleValue))
			result := process.Run()
			// TODO handle YIELD?
			if result.Code != core.ResultCode_OK {
				return result
			}
			value = result.Value
		}
		result := setArgument(arg.Name, value)
		// TODO handle YIELD?
		if result.Code != core.ResultCode_OK {
			return result
		}
	}
	return core.OK(core.NIL)
}

//   setArguments(values: Value[], scope: Scope): Result {
//     if (!this.checkArity(values, 0))
//       return ERROR(`wrong # values: should be "${this.usage()}"`);
//     return this.applyArguments(scope, values, 0, (name, value) =>
//       scope.setNamedVariable(name, value)
//     );
//   }
// }

type argspecCommand struct {
	scope *Scope
	//   ensemble: EnsembleCommand;
}

func newArgspecCommand(scope *Scope) argspecCommand {
	subscope := NewScope(scope, false)
	//     const { data: argspec } = ArgspecValue.fromValue(LIST([STR("value")]));
	//     this.ensemble = new EnsembleCommand(this.scope, argspec);
	return argspecCommand{subscope}
}
func (argspec argspecCommand) Execute(args []core.Value, context any) core.Result {
	if len(args) == 2 {
		return ArgspecValueFromValue(args[1]).AsResult()
	}
	//     return this.ensemble.execute(args, scope);
	return core.ERROR("TODO1")
}
func (argspecCommand) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	//     return this.ensemble.help(args);
	return core.ERROR("TODO2")
}

// const ARGSPEC_USAGE_SIGNATURE = "argspec value usage";
// const argspecUsageCmd: Command = {
//   execute(args) {
//     if (len(args) != 2) return ARITY_ERROR(ARGSPEC_USAGE_SIGNATURE);
//     const { data: value, ...result } = ArgspecValue.fromValue(args[1]);
//     if (result.code != core.ResultCode_OK) return result;
//     return OK(STR(value.usage()));
//   },
//   help(args) {
//     if (len(args) > 2) return ARITY_ERROR(ARGSPEC_USAGE_SIGNATURE);
//     return OK(STR(ARGSPEC_USAGE_SIGNATURE));
//   },
// };

// const ARGSPEC_SET_SIGNATURE = "argspec value set values";
// const argspecSetCmd: Command = {
//   execute(args, scope: Scope) {
//     if (len(args) != 3) return ARITY_ERROR(ARGSPEC_SET_SIGNATURE);
//     const { data: value, ...result } = ArgspecValue.fromValue(args[1]);
//     if (result.code != core.ResultCode_OK) return result;
//     const { data: values, ...result2 } = valueToArray(args[2]);
//     if (result2.code != core.ResultCode_OK) return result2;
//     return value.setArguments(values, scope);
//   },
//   help(args) {
//     if (len(args) > 2) return ARITY_ERROR(ARGSPEC_SET_SIGNATURE);
//     return OK(STR(ARGSPEC_SET_SIGNATURE));
//   },
// };

func registerArgspecCommands(scope *Scope) {
	command := newArgspecCommand(scope)
	scope.RegisterNamedCommand("argspec", command)
	//   command.scope.registerNamedCommand("usage", argspecUsageCmd);
	//   command.scope.registerNamedCommand("set", argspecSetCmd);
}
