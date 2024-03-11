package helena_dialect

import "helena/core"

// /* eslint-disable jsdoc/require-jsdoc */ // TODO
// import { Command } from "../core/command";
// import { Compiler, Executor, Program } from "../core/compiler";
// import {
//   defaultDisplayFunction,
//   DisplayFunction,
//   displayList,
// } from "../core/display";
// import { ERROR, OK, Result, ResultCode, YIELD } from "../core/results";
// import {
//   INT,
//   IntegerValue,
//   LIST,
//   ListValue,
//   NIL,
//   ScriptValue,
//   STR,
//   TupleValue,
//   Value,
//   ValueType,
// } from "../core/values";
// import { ArgspecValue } from "./argspecs";
// import { ARITY_ERROR } from "./arguments";
// import { destructureValue, Process, Scope } from "./core";
// import { EnsembleCommand } from "./ensembles";

// class ListCommand implements Command {
//   scope: Scope;
//   ensemble: EnsembleCommand;
//   constructor(scope: Scope) {
//     this.scope = new Scope(scope);
//     const { data: argspec } = ArgspecValue.fromValue(LIST([STR("value")]));
//     this.ensemble = new EnsembleCommand(this.scope, argspec);
//   }
//   execute(args: Value[], scope: Scope): Result {
//     if (args.length == 2) return valueToList(args[1]);
//     return this.ensemble.execute(args, scope);
//   }
//   help(args) {
//     return this.ensemble.help(args);
//   }
// }

// const LIST_LENGTH_SIGNATURE = "list value length";
// const listLengthCmd: Command = {
//   execute(args) {
//     if (args.length != 2) return ARITY_ERROR(LIST_LENGTH_SIGNATURE);
//     const { data: values, ...result } = valueToArray(args[1]);
//     if (result.code != core.ResultCode_OK) {return result;}
//     return OK(INT(values.length));
//   },
//   help(args) {
//     if (args.length > 2) return ARITY_ERROR(LIST_LENGTH_SIGNATURE);
//     return OK(STR(LIST_LENGTH_SIGNATURE));
//   },
// };

// const LIST_AT_SIGNATURE = "list value at index ?default?";
// const listAtCmd: Command = {
//   execute(args) {
//     if (args.length != 3 && args.length != 4)
//       return ARITY_ERROR(LIST_AT_SIGNATURE);
//     const { data: values, ...result } = valueToArray(args[1]);
//     if (result.code != core.ResultCode_OK) {return result;}
//     return ListValue.at(values, args[2], args[3]);
//   },
//   help(args) {
//     if (args.length > 4) return ARITY_ERROR(LIST_AT_SIGNATURE);
//     return OK(STR(LIST_AT_SIGNATURE));
//   },
// };

// const LIST_RANGE_SIGNATURE = "list value range first ?last?";
// const listRangeCmd: Command = {
//   execute(args) {
//     if (args.length != 3 && args.length != 4)
//       return ARITY_ERROR(LIST_RANGE_SIGNATURE);
//     const { data: values, ...result } = valueToArray(args[1]);
//     if (result.code != core.ResultCode_OK) {return result;}
//     const firstResult = IntegerValue.toInteger(args[2]);
//     if (firstResult.code != core.ResultCode_OK) {return firstResult;}
//     const first = Math.max(0, firstResult.data);
//     if (args.length == 3) {
//       if (first >= values.length) return OK(LIST([]));
//       return OK(LIST(values.slice(first)));
//     } else {
//       const lastResult = IntegerValue.toInteger(args[3]);
//       if (lastResult.code != core.ResultCode_OK) {return lastResult;}
//       const last = lastResult.data;
//       if (first >= values.length || last < first || last < 0)
//         return OK(LIST([]));
//       return OK(LIST(values.slice(first, last + 1)));
//     }
//   },
//   help(args) {
//     if (args.length > 4) return ARITY_ERROR(LIST_RANGE_SIGNATURE);
//     return OK(STR(LIST_RANGE_SIGNATURE));
//   },
// };

// const LIST_APPEND_SIGNATURE = "list value append ?list ...?";
// const listAppendCmd: Command = {
//   execute(args) {
//     const { data: values, ...result } = valueToArray(args[1]);
//     if (result.code != core.ResultCode_OK) {return result;}
//     const values2 = [...values];
//     for (let i = 2; i < args.length; i++) {
//       const { data: values, ...result } = valueToArray(args[i]);
//       if (result.code != core.ResultCode_OK) {return result;}
//       values2.push(...values);
//     }
//     return OK(LIST(values2));
//   },
//   help() {
//     return OK(STR(LIST_APPEND_SIGNATURE));
//   },
// };

// const LIST_REMOVE_SIGNATURE = "list value remove first last";
// const listRemoveCmd: Command = {
//   execute(args) {
//     if (args.length != 4 && args.length != 5)
//       return ARITY_ERROR(LIST_REMOVE_SIGNATURE);
//     const { data: values, ...result } = valueToArray(args[1]);
//     if (result.code != core.ResultCode_OK) {return result;}
//     const firstResult = IntegerValue.toInteger(args[2]);
//     if (firstResult.code != core.ResultCode_OK) {return firstResult;}
//     const first = Math.max(0, firstResult.data);
//     const lastResult = IntegerValue.toInteger(args[3]);
//     if (lastResult.code != core.ResultCode_OK) {return lastResult;}
//     const last = lastResult.data;
//     const head = values.slice(0, first);
//     const tail = values.slice(Math.max(first, last + 1));
//     return OK(LIST([...head, ...tail]));
//   },
//   help(args) {
//     if (args.length > 4) return ARITY_ERROR(LIST_REMOVE_SIGNATURE);
//     return OK(STR(LIST_REMOVE_SIGNATURE));
//   },
// };

// const LIST_INSERT_SIGNATURE = "list value insert index value2";
// const listInsertCmd: Command = {
//   execute(args) {
//     if (args.length != 4) return ARITY_ERROR(LIST_INSERT_SIGNATURE);
//     const { data: values, ...result } = valueToArray(args[1]);
//     if (result.code != core.ResultCode_OK) {return result;}
//     const indexResult = IntegerValue.toInteger(args[2]);
//     if (indexResult.code != core.ResultCode_OK) {return indexResult;}
//     const index = Math.max(0, indexResult.data);
//     const { data: insert, ...result2 } = valueToArray(args[3]);
//     if (result2.code != core.ResultCode_OK) {return result2;}
//     const head = values.slice(0, index);
//     const tail = values.slice(index);
//     return OK(LIST([...head, ...insert, ...tail]));
//   },
//   help(args) {
//     if (args.length > 4) return ARITY_ERROR(LIST_INSERT_SIGNATURE);
//     return OK(STR(LIST_INSERT_SIGNATURE));
//   },
// };

// const LIST_REPLACE_SIGNATURE = "list value replace first last value2";
// const listReplaceCmd: Command = {
//   execute(args) {
//     if (args.length != 5) return ARITY_ERROR(LIST_REPLACE_SIGNATURE);
//     const { data: values, ...result } = valueToArray(args[1]);
//     if (result.code != core.ResultCode_OK) {return result;}
//     const firstResult = IntegerValue.toInteger(args[2]);
//     if (firstResult.code != core.ResultCode_OK) {return firstResult;}
//     const first = Math.max(0, firstResult.data);
//     const lastResult = IntegerValue.toInteger(args[3]);
//     if (lastResult.code != core.ResultCode_OK) {return lastResult;}
//     const last = lastResult.data;
//     const head = values.slice(0, first);
//     const tail = values.slice(Math.max(first, last + 1));
//     const { data: insert, ...result2 } = valueToArray(args[4]);
//     if (result2.code != core.ResultCode_OK) {return result2;}
//     return OK(LIST([...head, ...insert, ...tail]));
//   },
//   help(args) {
//     if (args.length > 5) return ARITY_ERROR(LIST_REPLACE_SIGNATURE);
//     return OK(STR(LIST_REPLACE_SIGNATURE));
//   },
// };

// const LIST_FOREACH_SIGNATURE = "list value foreach element body";
// type ListForeachState = {
//   varname: Value;
//   list: ListValue;
//   i: number;
//   step: "beforeBody" | "inBody";
//   program: Program;
//   scope: Scope;
//   process?: Process;
//   lastResult: Result;
// };
// class ListForeachCommand implements Command {
//   execute(args, scope: Scope) {
//     if (args.length != 4) return ARITY_ERROR(LIST_FOREACH_SIGNATURE);
//     const { data: list, ...result } = valueToList(args[1]);
//     if (result.code != core.ResultCode_OK) {return result;}
//     const varname = args[2];
//     const body = args[3];
//     if (body.type != core.ValueType_SCRIPT) return ERROR("body must be a script");
//     const program = scope.compile((body as ScriptValue).script);
//     const subscope = new Scope(scope, true);
//     return this.run({
//       varname,
//       list: list as ListValue,
//       i: 0,
//       step: "beforeBody",
//       program,
//       scope: subscope,
//       lastResult: OK(NIL),
//     });
//   }
//   resume(result: Result): Result {
//     const state = result.data as ListForeachState;
//     state.process.yieldBack(result.value);
//     return this.run(state);
//   }
//   help(args) {
//     if (args.length > 4) return ARITY_ERROR(LIST_FOREACH_SIGNATURE);
//     return OK(STR(LIST_FOREACH_SIGNATURE));
//   }
//   private run(state: ListForeachState) {
//     for (;;) {
//       switch (state.step) {
//         case "beforeBody": {
//           if (state.i == state.list.values.length) return state.lastResult;
//           const value = state.list.values[state.i++];
//           const result = destructureValue(
//             state.scope.destructureLocal.bind(state.scope),
//             state.varname,
//             value
//           );
//           if (result.code != core.ResultCode_OK) {return result;}
//           state.process = state.scope.prepareProcess(state.program);
//           state.step = "inBody";
//           break;
//         }
//         case "inBody": {
//           const result = state.process.run();
//           if (result.code == core.ResultCode_YIELD)
//             return YIELD(result.value, state);
//           state.step = "beforeBody";
//           if (result.code == core.ResultCode_BREAK) {return state.lastResult;}
//           if (result.code == core.ResultCode_CONTINUE) {continue;}
//           if (result.code != core.ResultCode_OK) {return result;}
//           state.lastResult = result;
//           break;
//         }
//       }
//     }
//   }
// }
// const listForeachCmd = new ListForeachCommand();

// export function valueToList(value: Value): Result {
//   if (value.type == core.ValueType_SCRIPT) {
//     const { data, ...result } = valueToArray(value);
//     if (result.code != core.ResultCode_OK) {return result;}
//     return OK(LIST(data));
//   }
//   return ListValue.fromValue(value);
// }

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

// export function displayListValue(
//   list: ListValue,
//   fn: DisplayFunction = defaultDisplayFunction
// ) {
//   return `[list (${displayList(list.values, fn)})]`;
// }

// export function registerListCommands(scope: Scope) {
//   const command = new ListCommand(scope);
//   scope.registerNamedCommand("list", command);
//   command.scope.registerNamedCommand("length", listLengthCmd);
//   command.scope.registerNamedCommand("at", listAtCmd);
//   command.scope.registerNamedCommand("range", listRangeCmd);
//   command.scope.registerNamedCommand("append", listAppendCmd);
//   command.scope.registerNamedCommand("remove", listRemoveCmd);
//   command.scope.registerNamedCommand("insert", listInsertCmd);
//   command.scope.registerNamedCommand("replace", listReplaceCmd);
//   command.scope.registerNamedCommand("foreach", listForeachCmd);
// }
