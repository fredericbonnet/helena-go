package helena_dialect

import "helena/core"

const WHILE_SIGNATURE = "while test body"

type whileCmd struct{}

func (whileCmd) Execute(args []core.Value, context any) core.Result {
	scope := context.(*Scope)
	var test, body core.Value
	switch len(args) {
	case 3:
		test, body = args[1], args[2]
	default:
		return ARITY_ERROR(WHILE_SIGNATURE)
	}
	if body.Type() != core.ValueType_SCRIPT {
		return core.ERROR("body must be a script")
	}

	lastResult := core.OK(core.NIL)
	var callTest func() core.Result
	var callBody func() core.Result
	if test.Type() == core.ValueType_SCRIPT {
		testProgram := scope.Compile((test.(core.ScriptValue)).Script)
		callTest = func() core.Result {
			return CreateContinuationValue(scope, testProgram, func(result core.Result) core.Result {
				if result.Code != core.ResultCode_OK {
					return result
				}
				result2 := core.ValueToBoolean(result.Value)
				if result2.Code != core.ResultCode_OK {
					return result2.AsResult()
				}
				if !result2.Data {
					return lastResult
				}
				return callBody()
			})
		}
	} else {
		result := core.ValueToBoolean(test)
		if result.Code != core.ResultCode_OK {
			return result.AsResult()
		}
		if !result.Data {
			return lastResult
		}
		callTest = callBody
	}
	program := scope.Compile((body.(core.ScriptValue)).Script)
	callBody = func() core.Result {
		return CreateContinuationValue(scope, program, func(result core.Result) core.Result {
			switch result.Code {
			case core.ResultCode_BREAK:
				return lastResult
			case core.ResultCode_CONTINUE:
			case core.ResultCode_OK:
				lastResult = result
			default:
				return result
			}
			return callTest()
		})
	}
	return callTest()
}
func (whileCmd) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	if len(args) > 3 {
		return ARITY_ERROR(WHILE_SIGNATURE)
	}
	return core.OK(core.STR(WHILE_SIGNATURE))
}

const IF_SIGNATURE = "if test body ?elseif test body ...? ?else? ?body?"

type ifCmd struct{}

func (cmd ifCmd) Execute(args []core.Value, context any) core.Result {
	scope := context.(*Scope)
	checkResult := cmd.checkArgs(args)
	if checkResult.Code != core.ResultCode_OK {
		return checkResult
	}
	i := 0
	var callTest func() core.Result
	var callBody func() core.Result
	callTest = func() core.Result {
		if i >= len(args) {
			return core.OK(core.NIL)
		}
		keyword := core.ValueToString(args[i]).Data
		if keyword == "else" {
			return callBody()
		}
		test := args[i+1]
		if test.Type() == core.ValueType_SCRIPT {
			program := scope.CompileScriptValue(test.(core.ScriptValue))
			return CreateContinuationValue(scope, program, func(result core.Result) core.Result {
				if result.Code != core.ResultCode_OK {
					return result
				}
				result2 := core.ValueToBoolean(result.Value)
				if result2.Code != core.ResultCode_OK {
					return result2.AsResult()
				}
				if result2.Data {
					return callBody()
				}
				i += 3
				return callTest()
			})
		} else {
			result := core.ValueToBoolean(test)
			if result.Code != core.ResultCode_OK {
				return result.AsResult()
			}
			if result.Data {
				return callBody()
			}
			i += 3
			return callTest()
		}
	}
	callBody = func() core.Result {
		var body core.Value
		if core.ValueToString(args[i]).Data == "else" {
			body = args[i+1]
		} else {
			body = args[i+2]
		}
		if body.Type() != core.ValueType_SCRIPT {
			return core.ERROR("body must be a script")
		}
		program := scope.CompileScriptValue(body.(core.ScriptValue))
		return CreateContinuationValue(scope, program, nil)
	}
	return callTest()
}
func (ifCmd) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	return core.OK(core.STR(IF_SIGNATURE))
}
func (ifCmd) checkArgs(args []core.Value) core.Result {
	if len(args) == 2 {
		return core.ERROR("wrong # args: missing if body")
	}
	i := 3
	for i < len(args) {
		result := core.ValueToString(args[i])
		if result.Code != core.ResultCode_OK {
			return core.ERROR("invalid keyword")
		}
		keyword := result.Data
		switch keyword {
		case "elseif":
			switch len(args) - i {
			case 1:
				return core.ERROR("wrong # args: missing elseif test")
			case 2:
				return core.ERROR("wrong # args: missing elseif body")
			default:
				i += 3
			}
		case "else":
			switch len(args) - i {
			case 1:
				return core.ERROR("wrong # args: missing else body")
			default:
				i += 2
			}
		default:
			return core.ERROR(`invalid keyword "` + keyword + `"`)
		}
	}
	if i == len(args) {
		return core.OK(core.NIL)
	}
	return ARITY_ERROR(IF_SIGNATURE)
}

const WHEN_SIGNATURE = "when ?command? {?test body ...? ?default?}"

type whenCmd struct{}

func (whenCmd) Execute(args []core.Value, context any) core.Result {
	scope := context.(*Scope)
	var command, casesBody core.Value
	switch len(args) {
	case 2:
		casesBody = args[1]
	case 3:
		command, casesBody = args[1], args[2]
	default:
		return ARITY_ERROR(WHEN_SIGNATURE)
	}
	result := ValueToArray(casesBody)
	if result.Code != core.ResultCode_OK {
		return result.AsResult()
	}
	cases := result.Data
	if len(cases) == 0 {
		return core.OK(core.NIL)
	}
	i := 0
	var callCommand func() core.Result
	var callTest func(command core.Value) core.Result
	var callBody func() core.Result
	callCommand = func() core.Result {
		if i >= len(cases) {
			return core.OK(core.NIL)
		}
		if i == len(cases)-1 {
			return callBody()
		}
		if command == nil {
			return callTest(core.NIL)
		}
		if command.Type() == core.ValueType_SCRIPT {
			program := scope.CompileScriptValue(command.(core.ScriptValue))
			return CreateContinuationValue(scope, program, func(result core.Result) core.Result {
				if result.Code != core.ResultCode_OK {
					return result
				}
				return callTest(result.Value)
			})
		} else {
			return callTest(command)
		}
	}
	callTest = func(command core.Value) core.Result {
		test := cases[i]
		if command != core.NIL {
			switch test.Type() {
			case core.ValueType_TUPLE:
				test = core.TUPLE(append([]core.Value{command}, test.(core.TupleValue).Values...))
			default:
				test = core.TUPLE([]core.Value{command, test})
			}
		}
		var program *core.Program
		switch test.Type() {
		case core.ValueType_SCRIPT:
			{
				program = scope.CompileScriptValue(test.(core.ScriptValue))
			}
		case core.ValueType_TUPLE:
			{
				program = scope.CompileTupleValue(test.(core.TupleValue))
			}
		default:
			{
				result := core.ValueToBoolean(test)
				if result.Code != core.ResultCode_OK {
					return result.AsResult()
				}
				if result.Data {
					return callBody()
				}
				i += 2
				return callCommand()
			}
		}
		return CreateContinuationValue(scope, program, func(result core.Result) core.Result {
			if result.Code != core.ResultCode_OK {
				return result
			}
			result2 := core.ValueToBoolean(result.Value)
			if result2.Code != core.ResultCode_OK {
				return result2.AsResult()
			}
			if result2.Data {
				return callBody()
			}
			i += 2
			return callCommand()
		})
	}
	callBody = func() core.Result {
		var body core.Value
		if i == len(cases)-1 {
			body = cases[i]
		} else {
			body = cases[i+1]
		}
		if body.Type() != core.ValueType_SCRIPT {
			return core.ERROR("body must be a script")
		}
		program := scope.CompileScriptValue(body.(core.ScriptValue))
		return CreateContinuationValue(scope, program, nil)
	}
	return callCommand()
}
func (whenCmd) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	if len(args) > 3 {
		return ARITY_ERROR(WHEN_SIGNATURE)
	}
	return core.OK(core.STR(WHEN_SIGNATURE))
}

const CATCH_SIGNATURE = "catch body ?return value handler? ?yield value handler? ?error message handler? ?break handler? ?continue handler? ?finally handler?"

type catchStateStep uint8

const (
	catchStateStep_beforeBody catchStateStep = iota
	catchStateStep_inBody
	catchStateStep_beforeHandler
	catchStateStep_inHandler
	catchStateStep_beforeFinally
	catchStateStep_inFinally
)

type catchState struct {
	args        []core.Value
	step        catchStateStep
	bodyResult  core.Result
	bodyProcess *Process
	result      core.Result
	process     *Process
}

type catchCmd struct{}

func (cmd catchCmd) Execute(args []core.Value, context any) core.Result {
	scope := context.(*Scope)
	checkResult := cmd.checkArgs(args)
	if checkResult.Code != core.ResultCode_OK {
		return checkResult
	}
	if len(args) == 2 {
		body := args[1]
		if body.Type() != core.ValueType_SCRIPT {
			return core.ERROR("body must be a script")
		}
		result := scope.ExecuteScriptValue(body.(core.ScriptValue))
		codeName := core.STR(core.RESULT_CODE_NAME(result))
		switch result.Code {
		case core.ResultCode_OK,
			core.ResultCode_RETURN,
			core.ResultCode_YIELD,
			core.ResultCode_ERROR:
			return core.OK(core.TUPLE([]core.Value{codeName, result.Value}))
		default:
			return core.OK(core.TUPLE([]core.Value{core.STR(core.RESULT_CODE_NAME(result))}))
		}
	}
	return cmd.run(&catchState{step: catchStateStep_beforeBody, args: args}, scope)
}
func (cmd catchCmd) Resume(result core.Result, context any) core.Result {
	scope := context.(*Scope)
	state := result.Data.(*catchState)
	if state.step == catchStateStep_inBody {
		state.bodyProcess.YieldBack(result.Value)
	} else {
		state.process.YieldBack(result.Value)
	}
	return cmd.run(state, scope)
}
func (catchCmd) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	return core.OK(core.STR(CATCH_SIGNATURE))
}

func (cmd catchCmd) run(state *catchState, scope *Scope) core.Result {
	for {
		switch state.step {
		case catchStateStep_beforeBody:
			{
				body := state.args[1]
				// TODO check type
				state.bodyProcess = scope.PrepareScriptValue(body.(core.ScriptValue)) // TODO check type
				state.step = catchStateStep_inBody
			}
		case catchStateStep_inBody:
			{
				state.bodyResult = state.bodyProcess.Run()
				state.step = catchStateStep_beforeHandler
			}
		case catchStateStep_beforeHandler:
			{
				if state.bodyResult.Code == core.ResultCode_OK {
					state.result = state.bodyResult
					state.step = catchStateStep_beforeFinally
				}
				i := cmd.findHandlerIndex(state.bodyResult.Code, state.args)
				if i >= len(state.args)-1 {
					state.result = state.bodyResult
					state.step = catchStateStep_beforeFinally
					continue
				}
				switch state.bodyResult.Code {
				case core.ResultCode_RETURN,
					core.ResultCode_YIELD,
					core.ResultCode_ERROR:
					{
						varname := core.ValueToString(state.args[i+1]).Data
						handler := state.args[i+2]
						subscope := NewScope(scope, true)
						subscope.SetNamedLocal(varname, state.bodyResult.Value)
						state.process = subscope.PrepareScriptValue(
							handler.(core.ScriptValue),
						) // TODO check type
					}
				case core.ResultCode_BREAK,
					core.ResultCode_CONTINUE:
					{
						handler := state.args[i+1]
						state.process = scope.PrepareScriptValue(handler.(core.ScriptValue)) // TODO check type
					}
				default:
					panic("CANTHAPPEN")
				}
				state.step = catchStateStep_inHandler
			}
		case catchStateStep_inHandler:
			{
				state.result = state.process.Run()
				if state.result.Code == core.ResultCode_YIELD {
					return core.YIELD_STATE(state.result.Value, state)
				}
				if core.IsCustomResult(state.result, passResultCode) {
					if state.bodyResult.Code == core.ResultCode_YIELD {
						state.step = catchStateStep_inBody
						return core.YIELD_STATE(state.bodyResult.Value, state)
					}
					state.result = state.bodyResult
					state.step = catchStateStep_beforeFinally
					continue
				}
				if state.result.Code != core.ResultCode_OK {
					return state.result
				}
				state.step = catchStateStep_beforeFinally
			}
		case catchStateStep_beforeFinally:
			{
				i := cmd.findFinallyIndex(state.args)
				if i >= len(state.args)-1 {
					return state.result
				}
				handler := state.args[i+1]
				state.process = scope.PrepareScriptValue(handler.(core.ScriptValue)) // TODO check type
				state.step = catchStateStep_inFinally
			}
		case catchStateStep_inFinally:
			{
				result := state.process.Run()
				if result.Code == core.ResultCode_YIELD {
					return core.YIELD_STATE(result.Value, state)
				}
				if result.Code != core.ResultCode_OK {
					return result
				}
				return state.result
			}
		}
	}
}
func (catchCmd) findHandlerIndex(
	code core.ResultCode,
	args []core.Value,
) int {
	i := 2
	for i < len(args) {
		keyword := core.ValueToString(args[i]).Data
		switch keyword {
		case "return":
			if code == core.ResultCode_RETURN {
				return i
			}
			i += 3
		case "yield":
			if code == core.ResultCode_YIELD {
				return i
			}
			i += 3
		case "error":
			if code == core.ResultCode_ERROR {
				return i
			}
			i += 3
		case "break":
			if code == core.ResultCode_BREAK {
				return i
			}
			i += 2
		case "continue":
			if code == core.ResultCode_CONTINUE {
				return i
			}
			i += 2
		case "finally":
			i += 2
		}
	}
	return i
}
func (catchCmd) findFinallyIndex(args []core.Value) int {
	i := 2
	for i < len(args) {
		keyword := core.ValueToString(args[i]).Data
		switch keyword {
		case "return",
			"yield",
			"error":
			i += 3
		case "break",
			"continue":
			i += 2
		case "finally":
			return i
		}
	}
	return i
}
func (catchCmd) checkArgs(args []core.Value) core.Result {
	i := 2
	for i < len(args) {
		result := core.ValueToString(args[i])
		if result.Code != core.ResultCode_OK {
			return core.ERROR("invalid keyword")
		}
		keyword := result.Data
		switch keyword {
		case "return",
			"yield",
			"error":
			switch len(args) - i {
			case 1:
				return core.ERROR(`wrong #args: missing ` + keyword + ` handler parameter`)
			case 2:
				return core.ERROR(`wrong #args: missing ` + keyword + ` handler body`)
			default:
				{
					if core.ValueToString(args[i+1]).Code != core.ResultCode_OK {
						return core.ERROR(`invalid ` + keyword + ` handler parameter name`)
					}
					i += 3
				}
			}
		case "break",
			"continue",
			"finally":
			switch len(args) - i {
			case 1:
				return core.ERROR(`wrong #args: missing ` + keyword + ` handler body`)
			default:
				i += 2
			}
		default:
			return core.ERROR(`invalid keyword "` + keyword + `"`)
		}
	}
	if i == len(args) {
		return core.OK(core.NIL)
	}
	return ARITY_ERROR(CATCH_SIGNATURE)
}

const PASS_SIGNATURE = "pass"

var passResultCode = core.CustomResultCode{Name: "pass"}

type passCmd struct{}

func (passCmd) Execute(args []core.Value, _ any) core.Result {
	if len(args) != 1 {
		return ARITY_ERROR(PASS_SIGNATURE)
	}
	return core.CUSTOM_RESULT(passResultCode, core.NIL)
}
func (passCmd) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	if len(args) != 1 {
		return ARITY_ERROR(PASS_SIGNATURE)
	}
	return core.OK(core.STR(PASS_SIGNATURE))
}

func registerControlCommands(scope *Scope) {
	scope.RegisterNamedCommand("while", whileCmd{})
	scope.RegisterNamedCommand("if", ifCmd{})
	scope.RegisterNamedCommand("when", whenCmd{})
	scope.RegisterNamedCommand("catch", catchCmd{})
	scope.RegisterNamedCommand("pass", passCmd{})
}
