//
// Helena results
//

package core

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
type TypedResult[T any] struct {
	// Result code
	// readonly code: ResultCode | CustomResultCode;
	Code ResultCode

	// Result value
	Value Value

	// Extra data
	Data T
}
type Result TypedResult[any]

func (result TypedResult[any]) AsResult() Result {
	return Result{Code: result.Code, Value: result.Value, Data: nil}
}

func ResultAs[To any](result Result) TypedResult[To] {
	return TypedResult[To]{Code: result.Code, Value: result.Value}
}

//
// Convenience functions for results
//

func OK(value Value) Result {
	return Result{
		Code:  ResultCode_OK,
		Value: value,
	}
}
func OK_T[T any](value Value, data T) TypedResult[T] {
	return TypedResult[T]{
		Code:  ResultCode_OK,
		Value: value,
		Data:  data,
	}
}

func RETURN(value Value) Result {
	return Result{
		Code:  ResultCode_RETURN,
		Value: value,
	}
}

func YIELD(value Value) Result {
	return Result{
		Code:  ResultCode_YIELD,
		Value: value,
	}
}
func YIELD_STATE(value Value, state any) Result {
	return Result{
		Code:  ResultCode_YIELD,
		Value: value,
		Data:  state,
	}
}

func ERROR(message string) Result {
	return Result{
		Code:  ResultCode_ERROR,
		Value: STR(message),
	}
}
func ERROR_T[T any](message string) TypedResult[T] {
	return TypedResult[T]{
		Code:  ResultCode_ERROR,
		Value: STR(message),
	}
}

func BREAK(value Value) Result {
	return Result{
		Code:  ResultCode_BREAK,
		Value: value,
	}
}

func CONTINUE(value Value) Result {
	return Result{
		Code:  ResultCode_CONTINUE,
		Value: value,
	}
}

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
