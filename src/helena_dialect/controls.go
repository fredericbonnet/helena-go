package helena_dialect

import "helena/core"

type whileStateStep uint8

const (
	whileStateStep_beforeTest whileStateStep = iota
	whileStateStep_inTest
	whileStateStep_afterTest
	whileStateStep_beforeBody
	whileStateStep_inBody
)

type whileState struct {
	step        whileStateStep
	test        core.Value
	testProgram *core.Program
	result      core.Result
	program     *core.Program
	process     *Process
	lastResult  core.Result
}

const WHILE_SIGNATURE = "while test body"

type whileCmd struct{}

func (cmd whileCmd) Execute(args []core.Value, context any) core.Result {
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
	var testProgram *core.Program
	if test.Type() == core.ValueType_SCRIPT {
		testProgram = scope.Compile(test.(core.ScriptValue).Script)
	}
	program := scope.Compile(body.(core.ScriptValue).Script)
	return cmd.run(
		&whileState{
			step:        whileStateStep_beforeTest,
			test:        test,
			testProgram: testProgram,
			program:     program,
			lastResult:  core.OK(core.NIL),
		},
		scope,
	)
}
func (cmd whileCmd) Resume(result core.Result, context any) core.Result {
	scope := context.(*Scope)
	state := result.Data.(*whileState)
	state.process.YieldBack(result.Value)
	return cmd.run(state, scope)
}
func (whileCmd) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	if len(args) > 3 {
		return ARITY_ERROR(WHILE_SIGNATURE)
	}
	return core.OK(core.STR(WHILE_SIGNATURE))
}
func (cmd whileCmd) run(state *whileState, scope *Scope) core.Result {
	for {
		switch state.step {
		case whileStateStep_beforeTest:
			{
				state.result = cmd.executeTest(state, scope)
				state.step = whileStateStep_inTest
				if state.result.Code == core.ResultCode_YIELD {
					return core.YIELD_STATE(state.result.Value, state)
				}
				state.step = whileStateStep_afterTest
			}
		case whileStateStep_inTest:
			{
				state.result = cmd.resumeTest(state)
				if state.result.Code == core.ResultCode_YIELD {
					return core.YIELD_STATE(state.result.Value, state)
				}
				state.step = whileStateStep_afterTest
			}
		case whileStateStep_afterTest:
			if state.result.Code != core.ResultCode_OK {
				return state.result
			}
			if !state.result.Value.(core.BooleanValue).Value {
				return state.lastResult
			}
			state.step = whileStateStep_beforeBody
		case whileStateStep_beforeBody:
			state.process = scope.PrepareProcess(state.program)
			state.step = whileStateStep_inBody
		case whileStateStep_inBody:
			{
				result := state.process.Run()
				if result.Code == core.ResultCode_YIELD {
					return core.YIELD_STATE(result.Value, state)
				}
				state.step = whileStateStep_beforeTest
				if result.Code == core.ResultCode_BREAK {
					return state.lastResult
				}
				if result.Code == core.ResultCode_CONTINUE {
					continue
				}
				if result.Code != core.ResultCode_OK {
					return result
				}
				state.lastResult = result
			}
		}
	}
}
func (whileCmd) executeTest(state *whileState, scope *Scope) core.Result {
	if state.test.Type() == core.ValueType_SCRIPT {
		state.process = scope.PrepareProcess(state.testProgram)
		result := state.process.Run()
		if result.Code != core.ResultCode_OK {
			return result
		}
		return core.BooleanValueFromValue(result.Value).AsResult()
	}
	return core.BooleanValueFromValue(state.test).AsResult()
}
func (whileCmd) resumeTest(state *whileState) core.Result {
	result := state.process.Run()
	if result.Code != core.ResultCode_OK {
		return result
	}
	return core.BooleanValueFromValue(result.Value).AsResult()
}

const IF_SIGNATURE = "if test body ?elseif test body ...? ?else? ?body?"

type ifStateStep uint8

const (
	ifStateStep_beforeTest ifStateStep = iota
	ifStateStep_inTest
	ifStateStep_afterTest
	ifStateStep_beforeBody
	ifStateStep_inBody
)

type ifState struct {
	args    []core.Value
	i       int
	step    ifStateStep
	result  core.Result
	process *Process
}

type ifCmd struct{}

func (cmd ifCmd) Execute(args []core.Value, context any) core.Result {
	scope := context.(*Scope)
	checkResult := cmd.checkArgs(args)
	if checkResult.Code != core.ResultCode_OK {
		return checkResult
	}
	return cmd.run(&ifState{args: args, i: 0, step: ifStateStep_beforeTest}, scope)
}
func (cmd ifCmd) Resume(result core.Result, context any) core.Result {
	scope := context.(*Scope)
	state := result.Data.(*ifState)
	state.process.YieldBack(result.Value)
	return cmd.run(state, scope)
}
func (ifCmd) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	return core.OK(core.STR(IF_SIGNATURE))
}
func (cmd ifCmd) run(state *ifState, scope *Scope) core.Result {
	for state.i < len(state.args) {
		switch state.step {
		case ifStateStep_beforeTest:
			if core.ValueToString(state.args[state.i]).Data == "else" {
				state.step = ifStateStep_beforeBody
			} else {
				state.result = cmd.executeTest(state, scope)
				state.step = ifStateStep_inTest
				if state.result.Code == core.ResultCode_YIELD {
					return core.YIELD_STATE(state.result.Value, state)
				}
				state.step = ifStateStep_afterTest
			}
		case ifStateStep_inTest:
			state.result = cmd.resumeTest(state)
			if state.result.Code == core.ResultCode_YIELD {
				return core.YIELD_STATE(state.result.Value, state)
			}
			state.step = ifStateStep_afterTest
		case ifStateStep_afterTest:
			if state.result.Code != core.ResultCode_OK {
				return state.result
			}
			if !state.result.Value.(core.BooleanValue).Value {
				state.step = ifStateStep_beforeTest
				state.i += 3
				continue
			}
			state.step = ifStateStep_beforeBody
		case ifStateStep_beforeBody:
			{
				var body core.Value
				if core.ValueToString(state.args[state.i]).Data == "else" {
					body = state.args[state.i+1]
				} else {
					body = state.args[state.i+2]
				}
				if body.Type() != core.ValueType_SCRIPT {
					return core.ERROR("body must be a script")
				}
				state.process = scope.PrepareScriptValue(body.(core.ScriptValue))
				state.step = ifStateStep_inBody
			}
		case ifStateStep_inBody:
			{
				result := state.process.Run()
				if result.Code == core.ResultCode_YIELD {
					return core.YIELD_STATE(result.Value, state)
				}
				return result
			}
		}
	}
	return core.OK(core.NIL)
}
func (ifCmd) executeTest(state *ifState, scope *Scope) core.Result {
	test := state.args[state.i+1]
	if test.Type() == core.ValueType_SCRIPT {
		state.process = scope.PrepareScriptValue(test.(core.ScriptValue))
		result := state.process.Run()
		if result.Code != core.ResultCode_OK {
			return result
		}
		return core.BooleanValueFromValue(result.Value).AsResult()
	}
	return core.BooleanValueFromValue(test).AsResult()
}
func (ifCmd) resumeTest(state *ifState) core.Result {
	result := state.process.Run()
	if result.Code != core.ResultCode_OK {
		return result
	}
	return core.BooleanValueFromValue(result.Value).AsResult()
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

type whenStateStep uint8

const (
	whenStateStep_beforeCommand whenStateStep = iota
	whenStateStep_inCommand
	whenStateStep_afterCommand
	whenStateStep_beforeTest
	whenStateStep_inTest
	whenStateStep_afterTest
	whenStateStep_beforeBody
	whenStateStep_inBody
)

type whenState struct {
	command core.Value
	cases   []core.Value
	i       int
	step    whenStateStep
	process *Process
	result  core.Result
}

type whenCmd struct{}

func (cmd whenCmd) Execute(args []core.Value, context any) core.Result {
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
	return cmd.run(&whenState{
		command: command,
		cases:   cases,
		i:       0,
		step:    whenStateStep_beforeCommand,
	}, scope)
}
func (cmd whenCmd) Resume(result core.Result, context any) core.Result {
	scope := context.(*Scope)
	state := result.Data.(*whenState)
	state.process.YieldBack(result.Value)
	return cmd.run(state, scope)
}
func (whenCmd) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	if len(args) > 3 {
		return ARITY_ERROR(WHEN_SIGNATURE)
	}
	return core.OK(core.STR(WHEN_SIGNATURE))
}
func (cmd whenCmd) run(state *whenState, scope *Scope) core.Result {
	for state.i < len(state.cases) {
		switch state.step {
		case whenStateStep_beforeCommand:
			if state.i == len(state.cases)-1 {
				state.step = whenStateStep_beforeBody
			} else {
				state.result = cmd.getCommand(state, scope)
				state.step = whenStateStep_inCommand
				if state.result.Code == core.ResultCode_YIELD {
					return core.YIELD_STATE(state.result.Value, state)
				}
				state.step = whenStateStep_afterCommand
			}
		case whenStateStep_inCommand:
			state.result = state.process.Run()
			if state.result.Code == core.ResultCode_YIELD {
				return core.YIELD_STATE(state.result.Value, state)
			}
			state.step = whenStateStep_afterCommand
		case whenStateStep_afterCommand:
			if state.result.Code != core.ResultCode_OK {
				return state.result
			}
			state.result = cmd.getTest(state, state.result.Value)
			state.step = whenStateStep_beforeTest
		case whenStateStep_beforeTest:
			state.result = cmd.executeTest(state.result.Value, state, scope)
			state.step = whenStateStep_inTest
			if state.result.Code == core.ResultCode_YIELD {
				return core.YIELD_STATE(state.result.Value, state)
			}
			state.step = whenStateStep_afterTest
		case whenStateStep_inTest:
			state.result = cmd.resumeTest(state)
			if state.result.Code == core.ResultCode_YIELD {
				return core.YIELD_STATE(state.result.Value, state)
			}
			state.step = whenStateStep_afterTest
		case whenStateStep_afterTest:
			if state.result.Code != core.ResultCode_OK {
				return state.result
			}
			if !state.result.Value.(core.BooleanValue).Value {
				state.step = whenStateStep_beforeCommand
				state.i += 2
				continue
			}
			state.step = whenStateStep_beforeBody
		case whenStateStep_beforeBody:
			{
				var body core.Value
				if state.i == len(state.cases)-1 {
					body = state.cases[state.i]
				} else {
					body = state.cases[state.i+1]
				}
				if body.Type() != core.ValueType_SCRIPT {
					return core.ERROR("body must be a script")
				}
				state.process = scope.PrepareScriptValue(body.(core.ScriptValue))
				state.step = whenStateStep_inBody
			}
		case whenStateStep_inBody:
			{
				result := state.process.Run()
				if result.Code == core.ResultCode_YIELD {
					return core.YIELD_STATE(result.Value, state)
				}
				return result
			}
		}
	}
	return core.OK(core.NIL)
}
func (whenCmd) getCommand(state *whenState, scope *Scope) core.Result {
	if state.command == nil {
		return core.OK(core.NIL)
	}
	if state.command.Type() == core.ValueType_SCRIPT {
		state.process = scope.PrepareScriptValue(state.command.(core.ScriptValue))
		return state.process.Run()
	}
	return core.OK(state.command)
}
func (whenCmd) getTest(state *whenState, command core.Value) core.Result {
	test := state.cases[state.i]
	if command == core.NIL {
		return core.OK(test)
	}
	switch test.Type() {
	case core.ValueType_TUPLE:
		return core.OK(core.TUPLE(append([]core.Value{command}, test.(core.TupleValue).Values...)))
	default:
		return core.OK(core.TUPLE([]core.Value{command, test}))
	}
}
func (whenCmd) executeTest(test core.Value, state *whenState, scope *Scope) core.Result {
	switch test.Type() {
	case core.ValueType_SCRIPT:
		{
			state.process = scope.PrepareScriptValue(test.(core.ScriptValue))
			result := state.process.Run()
			if result.Code != core.ResultCode_OK {
				return result
			}
			return core.BooleanValueFromValue(result.Value).AsResult()
		}
	case core.ValueType_TUPLE:
		{
			state.process = scope.PrepareTupleValue(test.(core.TupleValue))
			result := state.process.Run()
			if result.Code != core.ResultCode_OK {
				return result
			}
			return core.BooleanValueFromValue(result.Value).AsResult()
		}
	default:
		return core.BooleanValueFromValue(test).AsResult()
	}
}
func (whenCmd) resumeTest(state *whenState) core.Result {
	result := state.process.Run()
	if result.Code != core.ResultCode_OK {
		return result
	}
	return core.BooleanValueFromValue(result.Value).AsResult()
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
				// if state.result.Code == passResultCode {
				// 	if state.bodyResult.Code == core.ResultCode_YIELD {
				// 		state.step = catchStateStep_inBody
				// 		return core.YIELD_STATE(state.bodyResult.Value, state)
				// 	}
				// 	state.result = state.bodyResult
				// 	state.step = catchStateStep_beforeFinally
				// 	continue
				// }
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
	// code: ResultCode | CustomResultCode,
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

// const PASS_SIGNATURE = "pass";
// const passResultCode: CustomResultCode = { name: "pass" };
// const passCmd: Command = {
//   execute(args) {
//     if (len(args) != 1) return ARITY_ERROR(PASS_SIGNATURE);
//     return CUSTOM_RESULT(passResultCode);
//   },
//   help(args) {
//     if (len(args) != 1) return ARITY_ERROR(PASS_SIGNATURE);
//     return OK(STR(PASS_SIGNATURE));
//   },
// };

func registerControlCommands(scope *Scope) {
	scope.RegisterNamedCommand("while", whileCmd{})
	scope.RegisterNamedCommand("if", ifCmd{})
	scope.RegisterNamedCommand("when", whenCmd{})
	scope.RegisterNamedCommand("catch", catchCmd{})
	// scope.RegisterNamedCommand("pass", passCmd);
}
