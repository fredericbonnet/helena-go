//
// Helena results
//

package core

// import { Value, NIL, STR } from "./values";

// Helena standard result codes
type ResultCode int8

const (
	ResultCode_OK ResultCode = iota
	ResultCode_RETURN
	ResultCode_YIELD
	ResultCode_ERROR
	ResultCode_BREAK
	ResultCode_CONTINUE
)

// /** Helena custom result code */
// export interface CustomResultCode {
//   /** Custom code name */
//   name: string;
// }

// Helena result
type Result struct {
	// Result code
	// readonly code: ResultCode | CustomResultCode;
	Code ResultCode

	// Result value
	Value Value
}
type TypedResult[T any] struct {
	Result

	// Extra data
	Data T
}

func ResultAs[To any](result Result) TypedResult[To] {
	return TypedResult[To]{Result: result}
}

// /**
//  * Convenience functions for results
//  */

// /* eslint-disable jsdoc/require-jsdoc */
// export const OK = <T = unknown>(value: Value, data?: T): Result<T> => ({
//   code: ResultCode.OK,
//   value,
//   data,
// });
func OK(value Value) Result {
	return Result{
		Code:  ResultCode_OK,
		Value: value,
	}
}
func OK_T[T any](value Value, data T) TypedResult[T] {
	return TypedResult[T]{
		Result: OK(value),
		Data:   data,
	}
}

// export const RETURN = (value: Value = NIL): Result => ({
//   code: ResultCode.RETURN,
//   value,
// });
// export const YIELD = (value: Value = NIL, state?): Result => ({
//   code: ResultCode.YIELD,
//   value,
//   data: state,
// });
func ERROR(message string) Result {
	return Result{ //Result<never>
		Code:  ResultCode_ERROR,
		Value: STR(message),
	}
}
func ERROR_T[T any](message string) TypedResult[T] {
	return TypedResult[T]{
		Result: ERROR(message),
	}
}

// export const BREAK = (value: Value = NIL): Result => ({
//   code: ResultCode.BREAK,
//   value,
// });
// export const CONTINUE = (value: Value = NIL): Result => ({
//   code: ResultCode.CONTINUE,
//   value,
// });
// export const CUSTOM_RESULT = (
//   code: CustomResultCode,
//   value: Value = NIL
// ): Result => ({
//   code,
//   value,
// });

// export const RESULT_CODE_NAME = (code: ResultCode | CustomResultCode) => {
//   switch (code) {
//     case ResultCode.OK:
//       return "ok";
//     case ResultCode.RETURN:
//       return "return";
//     case ResultCode.YIELD:
//       return "yield";
//     case ResultCode.ERROR:
//       return "error";
//     case ResultCode.BREAK:
//       return "break";
//     case ResultCode.CONTINUE:
//       return "continue";
//     default:
//       return (code as CustomResultCode).name;
//   }
// };
// /* eslint-enable jsdoc/require-jsdoc */
