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

//   usage(skip = 0): string {
//     return buildUsage(this.argspec.args, skip);
//   }
//   checkArity(values: Value[], skip: number) {
//     return (
//       values.length - skip >= this.argspec.nbRequired &&
//       (this.argspec.hasRemainder ||
//         values.length - skip <=
//           this.argspec.nbRequired + this.argspec.nbOptional)
//     );
//   }
//   applyArguments(
//     scope: Scope,
//     values: Value[],
//     skip: number,
//     setArgument: (name: string, value: Value) => Result
//   ): Result {
//     const nonRequired = values.length - skip - this.argspec.nbRequired;
//     let optionals = Math.min(this.argspec.nbOptional, nonRequired);
//     const remainders = nonRequired - optionals;
//     let i = skip;
//     for (const arg of this.argspec.args) {
//       let value: Value;
//       switch (arg.type) {
//         case "required":
//           value = values[i++];
//           break;
//         case "optional":
//           if (optionals > 0) {
//             optionals--;
//             value = values[i++];
//           } else if (arg.default) {
//             if (arg.default.type == ValueType.SCRIPT) {
//               const body = arg.default as ScriptValue;
//               const result = scope.executeScriptValue(body);
//               // TODO handle YIELD?
//               if (result.code != ResultCode.OK) return result;
//               value = result.value;
//             } else {
//               value = arg.default;
//             }
//           } else continue; // Skip missing optional
//           break;
//         case "remainder":
//           value = TUPLE(values.slice(i, i + remainders));
//           i += remainders;
//           break;
//       }
//       if (arg.guard) {
//         const process = scope.prepareTupleValue(TUPLE([arg.guard, value]));
//         const result = process.run();
//         // TODO handle YIELD?
//         if (result.code != ResultCode.OK) return result;
//         value = result.value;
//       }
//       const result = setArgument(arg.name, value);
//       // TODO handle YIELD?
//       if (result.code != ResultCode.OK) return result;
//     }
//     return OK(NIL);
//   }

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
//     if (result.code != ResultCode.OK) return result;
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
//     if (result.code != ResultCode.OK) return result;
//     const { data: values, ...result2 } = valueToArray(args[2]);
//     if (result2.code != ResultCode.OK) return result2;
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
