package helena_dialect

import (
	"helena/core"
	"sync"
)

const LOOP_SIGNATURE = "loop ?index? ?value source ...? body"

type loopCmd struct{}
type LoopSourceFn = func(i int, data any, callback ContinuationCallback) core.Result
type loopCmdState struct {
	scope         *Scope
	bodyProgram   *core.Program
	lastResult    core.Result
	hasIndex      bool
	index         string
	varnames      []core.Value
	sources       []LoopSourceFn
	activeSources int
	i             int
	iSource       int
}

var loopCmdStatePool = sync.Pool{
	New: func() any {
		return &loopCmdState{}
	},
}

func (loopCmd) Execute(args []core.Value, context any) core.Result {
	scope := context.(*Scope)
	if len(args) < 2 {
		return ARITY_ERROR(LOOP_SIGNATURE)
	}
	var index string
	var hasIndex = (len(args)%2 == 1)
	if hasIndex {
		result, name := core.ValueToString(args[1])
		if result.Code != core.ResultCode_OK {
			return core.ERROR("invalid index name")
		}
		index = name
	}
	body := args[len(args)-1]
	if body.Type() != core.ValueType_SCRIPT {
		return core.ERROR("body must be a script")
	}
	nbSources := (len(args) - 2) / 2
	var firstSource int
	if hasIndex {
		firstSource = 2
	} else {
		firstSource = 1
	}
	slots := map[string]uint{}
	if hasIndex {
		slots[index] = 0
	}
	varnames := make([]core.Value, nbSources)
	for i, iSource := firstSource, 0; i < len(args)-1; i, iSource = i+2, iSource+1 {
		varname := args[i]
		varnames[iSource] = varname
		result := DestructureLocalSlots(varname, slots)
		if result.Code != core.ResultCode_OK {
			return result
		}
	}
	subscope := scope.NewLocalScope(slots, nil)
	sources := make([]LoopSourceFn, nbSources)
	for i, iSource := firstSource, 0; i < len(args)-1; i, iSource = i+2, iSource+1 {
		source := args[i+1]
		switch source.Type() {
		case core.ValueType_LIST:
			{
				list := source.(core.ListValue)
				sources[iSource] = func(i int, data any, callback ContinuationCallback) core.Result {
					if i >= len(list.Values) {
						return callback(core.BREAK(nil), data)
					}
					value := list.Values[i]
					return callback(core.OK(value), data)
				}
			}
		case core.ValueType_DICTIONARY:
			{
				dictionary := source.(core.DictionaryValue)
				entries := make([][2]core.Value, len(dictionary.Map))
				j := 0
				for key, value := range dictionary.Map {
					entries[j] = [2]core.Value{core.STR(key), value}
					j++
				}
				sources[iSource] = func(i int, data any, callback ContinuationCallback) core.Result {
					if i >= len(entries) {
						return callback(core.BREAK(nil), data)
					}
					value := core.TUPLE(entries[i][:])
					return callback(core.OK(value), data)
				}
			}
		case core.ValueType_SCRIPT:
			{
				program := subscope.CompileScriptValue(source.(core.ScriptValue))
				sources[iSource] = func(i int, data any, callback ContinuationCallback) core.Result {
					return CreateContinuationValueWithCallback(subscope, program, data, callback)
				}
			}
		default:
			{
				if scope.ResolveCommand(source) == nil {
					return core.ERROR("invalid source")
				}
				sources[iSource] = func(i int, data any, callback ContinuationCallback) core.Result {
					program := subscope.CompilePair(source, core.INT(int64(i)))
					return CreateContinuationValueWithCallback(subscope, program, data, callback)
				}
			}
		}
	}

	state := loopCmdStatePool.Get().(*loopCmdState)
	state.scope = subscope
	state.bodyProgram = subscope.CompileScriptValue(body.(core.ScriptValue))
	state.lastResult = core.OK(core.NIL)
	state.hasIndex = hasIndex
	state.index = index
	state.varnames = varnames
	state.sources = sources
	state.activeSources = len(sources)
	state.i = 0
	state.iSource = -1
	return loopCmdNextIteration(state)
}
func (loopCmd) Help(_ []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	return core.OK(core.STR(LOOP_SIGNATURE))
}
func loopCmdNextIteration(state *loopCmdState) core.Result {
	state.scope.ClearLocals()
	if state.hasIndex {
		state.scope.SetNamedLocal(state.index, core.INT(int64(state.i)))
	}
	state.iSource = -1
	return loopCmdNextSource(state)
}
func loopCmdNextSource(state *loopCmdState) core.Result {
	if state.activeSources == 0 && len(state.sources) > 0 {
		loopCmdStatePool.Put(state)
		return state.lastResult
	}
	state.iSource++
	if state.iSource >= len(state.sources) {
		return loopCmdBody(state)
	}
	source := state.sources[state.iSource]
	if source == nil {
		return loopCmdNextSource(state)
	}
	varname := state.varnames[state.iSource]
	return source(state.i, state, func(result core.Result, data any) core.Result {
		switch result.Code {
		case core.ResultCode_BREAK:
			state.sources[state.iSource] = nil
			state.activeSources--
			return loopCmdNextSource(state)
		case core.ResultCode_CONTINUE:
			return loopCmdNextSource(state)
		case core.ResultCode_OK:
		default:
			loopCmdStatePool.Put(state)
			return result
		}
		value := result.Value
		result2 := DestructureValue(
			func(name core.Value, value core.Value, check bool) core.Result {
				return state.scope.DestructureLocal(name, value, check)
			},
			varname,
			value,
		)
		if result2.Code != core.ResultCode_OK {
			loopCmdStatePool.Put(state)
			return result2
		}
		return loopCmdNextSource(state)
	})
}
func loopCmdBody(state *loopCmdState) core.Result {
	state.i++
	return CreateContinuationValueWithCallback(state.scope, state.bodyProgram, state, func(result core.Result, data any) core.Result {
		state := data.(*loopCmdState)
		switch result.Code {
		case core.ResultCode_BREAK:
			loopCmdStatePool.Put(state)
			return state.lastResult
		case core.ResultCode_CONTINUE:
		case core.ResultCode_OK:
			state.lastResult = result
		default:
			loopCmdStatePool.Put(state)
			return result
		}
		return loopCmdNextIteration(state)
	})
}

const WHILE_SIGNATURE = "while test body"

type whileCmd struct{}
type whileCmdState struct {
	scope       *Scope
	testProgram *core.Program
	bodyProgram *core.Program
	lastResult  core.Result
}

var whileCmdStatePool = sync.Pool{
	New: func() any {
		return &whileCmdState{}
	},
}

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
	if test.Type() == core.ValueType_SCRIPT {
		state := whileCmdStatePool.Get().(*whileCmdState)
		state.scope = scope
		state.bodyProgram = scope.CompileScriptValue(body.(core.ScriptValue))
		state.testProgram = scope.CompileScriptValue(test.(core.ScriptValue))
		state.lastResult = core.OK(core.NIL)
		return whileCmdTest(state)
	} else {
		result, b := core.ValueToBoolean(test)
		if result.Code != core.ResultCode_OK {
			return result
		}
		if !b {
			return core.OK(core.NIL)
		}
		state := whileCmdStatePool.Get().(*whileCmdState)
		state.scope = scope
		state.bodyProgram = scope.CompileScriptValue(body.(core.ScriptValue))
		state.testProgram = nil
		state.lastResult = core.OK(core.NIL)
		return whileCmdLoop(state)
	}
}
func (whileCmd) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	if len(args) > 3 {
		return ARITY_ERROR(WHILE_SIGNATURE)
	}
	return core.OK(core.STR(WHILE_SIGNATURE))
}
func whileCmdLoop(state *whileCmdState) core.Result {
	return CreateContinuationValueWithCallback(state.scope, state.bodyProgram, state, func(result core.Result, data any) core.Result {
		state := data.(*whileCmdState)
		switch result.Code {
		case core.ResultCode_BREAK:
			whileCmdStatePool.Put(state)
			return state.lastResult
		case core.ResultCode_CONTINUE:
		case core.ResultCode_OK:
			state.lastResult = result
		default:
			whileCmdStatePool.Put(state)
			return result
		}
		return whileCmdLoop(state)
	})
}
func whileCmdTest(state *whileCmdState) core.Result {
	return CreateContinuationValueWithCallback(state.scope, state.testProgram, state, func(result core.Result, data any) core.Result {
		state := data.(*whileCmdState)
		if result.Code != core.ResultCode_OK {
			whileCmdStatePool.Put(state)
			return result
		}
		result2, b := core.ValueToBoolean(result.Value)
		if result2.Code != core.ResultCode_OK {
			whileCmdStatePool.Put(state)
			return result2
		}
		if !b {
			whileCmdStatePool.Put(state)
			return state.lastResult
		}
		return CreateContinuationValueWithCallback(state.scope, state.bodyProgram, state, func(result core.Result, data any) core.Result {
			state := data.(*whileCmdState)
			switch result.Code {
			case core.ResultCode_BREAK:
				whileCmdStatePool.Put(state)
				return state.lastResult
			case core.ResultCode_CONTINUE:
			case core.ResultCode_OK:
				state.lastResult = result
			default:
				whileCmdStatePool.Put(state)
				return result
			}
			return whileCmdTest(state)
		})
	})
}

const IF_SIGNATURE = "if test body ?elseif test body ...? ?else? ?body?"

type ifCmd struct{}
type ifCmdState struct {
	scope *Scope
	i     int
	args  []core.Value
}

var ifCmdStatePool = sync.Pool{
	New: func() any {
		return &ifCmdState{}
	},
}

func (cmd ifCmd) Execute(args []core.Value, context any) core.Result {
	scope := context.(*Scope)
	checkResult := cmd.checkArgs(args)
	if checkResult.Code != core.ResultCode_OK {
		return checkResult
	}
	i := 0
	return ifCmdTest(scope, args, i)
}
func (ifCmd) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	return core.OK(core.STR(IF_SIGNATURE))
}
func ifCmdTest(scope *Scope, args []core.Value, i int) core.Result {
	if i >= len(args) {
		return core.OK(core.NIL)
	}
	_, keyword := core.ValueToString(args[i])
	if keyword == "else" {
		return ifCmdBody(scope, args, i)
	}
	test := args[i+1]
	if test.Type() == core.ValueType_SCRIPT {
		program := scope.CompileScriptValue(test.(core.ScriptValue))
		state := ifCmdStatePool.Get().(*ifCmdState)
		state.scope = scope
		state.i = i
		state.args = args
		return CreateContinuationValueWithCallback(scope, program, state, func(result core.Result, data any) core.Result {
			state := data.(*ifCmdState)
			ifCmdStatePool.Put(state)
			if result.Code != core.ResultCode_OK {
				return result
			}
			return ifCmdNext(result.Value, state.scope, state.args, state.i)
		})
	} else {
		return ifCmdNext(test, scope, args, i)
	}
}
func ifCmdNext(value core.Value, scope *Scope, args []core.Value, i int) core.Result {
	result, b := core.ValueToBoolean(value)
	if result.Code != core.ResultCode_OK {
		return result
	}
	if b {
		return ifCmdBody(scope, args, i)
	}
	i += 3
	return ifCmdTest(scope, args, i)
}
func ifCmdBody(scope *Scope, args []core.Value, i int) core.Result {
	var body core.Value
	if _, s := core.ValueToString(args[i]); s == "else" {
		body = args[i+1]
	} else {
		body = args[i+2]
	}
	if body.Type() != core.ValueType_SCRIPT {
		return core.ERROR("body must be a script")
	}
	program := scope.CompileScriptValue(body.(core.ScriptValue))
	return CreateContinuationValue(scope, program)
}
func (ifCmd) checkArgs(args []core.Value) core.Result {
	if len(args) == 2 {
		return core.ERROR("wrong # args: missing if body")
	}
	i := 3
	for i < len(args) {
		result, keyword := core.ValueToString(args[i])
		if result.Code != core.ResultCode_OK {
			return core.ERROR("invalid keyword")
		}
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
type whenCmdState struct {
	scope      *Scope
	hasCommand bool
	command    core.Value
	i          int
	cases      []core.Value
}

var whenCmdStatePool = sync.Pool{
	New: func() any {
		return &whenCmdState{}
	},
}

func (whenCmd) Execute(args []core.Value, context any) core.Result {
	scope := context.(*Scope)
	hasCommand := false
	var command, casesBody core.Value
	switch len(args) {
	case 2:
		casesBody = args[1]
	case 3:
		hasCommand = true
		command, casesBody = args[1], args[2]
	default:
		return ARITY_ERROR(WHEN_SIGNATURE)
	}
	result, cases := ValueToArray(casesBody)
	if result.Code != core.ResultCode_OK {
		return result
	}
	if len(cases) == 0 {
		return core.OK(core.NIL)
	}
	state := whenCmdStatePool.Get().(*whenCmdState)
	state.scope = scope
	state.hasCommand = hasCommand
	state.command = command
	state.i = 0
	state.cases = cases
	return whenCmdCommand(state)
}
func whenCmdCommand(state *whenCmdState) core.Result {
	if state.i >= len(state.cases) {
		whenCmdStatePool.Put(state)
		return core.OK(core.NIL)
	}
	if state.i == len(state.cases)-1 {
		return whenCmdBody(state)
	}
	if !state.hasCommand {
		return whenCmdTest(core.NIL, state)
	}
	if state.command.Type() == core.ValueType_SCRIPT {
		program := state.scope.CompileScriptValue(state.command.(core.ScriptValue))
		return CreateContinuationValueWithCallback(state.scope, program, state, func(result core.Result, data any) core.Result {
			state := data.(*whenCmdState)
			if result.Code != core.ResultCode_OK {
				whenCmdStatePool.Put(state)
				return result
			}
			return whenCmdTest(result.Value, state)
		})
	} else {
		return whenCmdTest(state.command, state)
	}
}
func whenCmdTest(command core.Value, state *whenCmdState) core.Result {
	test := state.cases[state.i]
	if state.hasCommand {
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
			program = state.scope.CompileScriptValue(test.(core.ScriptValue))
		}
	case core.ValueType_TUPLE:
		{
			program = state.scope.CompileTupleValue(test.(core.TupleValue))
		}
	default:
		{
			result, b := core.ValueToBoolean(test)
			if result.Code != core.ResultCode_OK {
				whenCmdStatePool.Put(state)
				return result
			}
			if b {
				return whenCmdBody(state)
			}
			state.i += 2
			return whenCmdCommand(state)
		}
	}
	return CreateContinuationValueWithCallback(state.scope, program, state, func(result core.Result, data any) core.Result {
		state := data.(*whenCmdState)
		if result.Code != core.ResultCode_OK {
			whenCmdStatePool.Put(state)
			return result
		}
		result2, b := core.ValueToBoolean(result.Value)
		if result2.Code != core.ResultCode_OK {
			whenCmdStatePool.Put(state)
			return result2
		}
		if b {
			return whenCmdBody(state)
		}
		state.i += 2
		return whenCmdCommand(state)
	})
}
func whenCmdBody(state *whenCmdState) core.Result {
	whenCmdStatePool.Put(state)
	var body core.Value
	if state.i == len(state.cases)-1 {
		body = state.cases[state.i]
	} else {
		body = state.cases[state.i+1]
	}
	if body.Type() != core.ValueType_SCRIPT {
		return core.ERROR("body must be a script")
	}
	program := state.scope.CompileScriptValue(body.(core.ScriptValue))
	return CreateContinuationValue(state.scope, program)
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
		program := scope.CompileScriptValue(body.(core.ScriptValue))
		result := scope.Execute(program, nil)
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
				program := scope.CompileScriptValue(body.(core.ScriptValue)) // TODO check type
				state.bodyProcess = scope.PrepareProcess(program)
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
						_, varname := core.ValueToString(state.args[i+1])
						handler := state.args[i+2]
						subscope := scope.NewLocalScope(
							map[string]uint{varname: 0},
							[]core.Value{state.bodyResult.Value},
						)
						program := subscope.CompileScriptValue(
							handler.(core.ScriptValue),
						) // TODO check type
						state.process = subscope.PrepareProcess(program)
					}
				case core.ResultCode_BREAK,
					core.ResultCode_CONTINUE:
					{
						handler := state.args[i+1]
						program := scope.CompileScriptValue(handler.(core.ScriptValue)) // TODO check type
						state.process = scope.PrepareProcess(program)
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
				program := scope.CompileScriptValue(handler.(core.ScriptValue)) // TODO check type
				state.process = scope.PrepareProcess(program)
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
		_, keyword := core.ValueToString(args[i])
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
		_, keyword := core.ValueToString(args[i])
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
		result, keyword := core.ValueToString(args[i])
		if result.Code != core.ResultCode_OK {
			return core.ERROR("invalid keyword")
		}
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
					if result, _ := core.ValueToString(args[i+1]); result.Code != core.ResultCode_OK {
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
	scope.RegisterNamedCommand("loop", loopCmd{})
	scope.RegisterNamedCommand("while", whileCmd{})
	scope.RegisterNamedCommand("if", ifCmd{})
	scope.RegisterNamedCommand("when", whenCmd{})
	scope.RegisterNamedCommand("catch", catchCmd{})
	scope.RegisterNamedCommand("pass", passCmd{})
}
