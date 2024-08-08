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
	ResultCode_CUSTOM
)

// Helena custom result code
type CustomResultCode struct {
	// Custom code name
	Name string
}

// Helena result
type Result struct {
	// Result code
	Code ResultCode

	// Result value
	Value Value

	// Extra data
	Data any
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
func ERROR_STACK(message string, errorStack *ErrorStack) Result {
	return Result{
		Code:  ResultCode_ERROR,
		Value: STR(message),
		Data:  errorStack,
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

func CUSTOM_RESULT(
	code CustomResultCode,
	value Value,
) Result {
	return Result{
		Code:  ResultCode_CUSTOM,
		Value: value,
		Data:  code,
	}
}

func RESULT_CODE_NAME(result Result) string {
	switch result.Code {
	case ResultCode_OK:
		return "ok"
	case ResultCode_RETURN:
		return "return"
	case ResultCode_YIELD:
		return "yield"
	case ResultCode_ERROR:
		return "error"
	case ResultCode_BREAK:
		return "break"
	case ResultCode_CONTINUE:
		return "continue"
	case ResultCode_CUSTOM:
		return result.Data.(CustomResultCode).Name

	default:
		panic("CANTHAPPEN")
	}
}

// Report whether result code is a custom code of the given type
func IsCustomResult(
	result Result,
	customType CustomResultCode,
) bool {
	return result.Code == ResultCode_CUSTOM && result.Data.(CustomResultCode) == customType
}
