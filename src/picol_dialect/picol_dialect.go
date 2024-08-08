package picol_dialect

import (
	"strconv"

	"helena/core"
)

type PicolScope struct {
	Parent    *PicolScope
	Variables map[string]core.Value
	Commands  map[string]core.Command
	Evaluator core.Evaluator
}

type variableResolver struct{ scope *PicolScope }

func (resolver variableResolver) Resolve(name string) core.Value {
	return resolver.scope.resolveVariable(name)
}

type commandResolver struct{ scope *PicolScope }

func (resolver commandResolver) Resolve(name core.Value) core.Command {
	return resolver.scope.resolveCommand(name)
}

func NewPicolScope(parent *PicolScope) *PicolScope {
	scope := &PicolScope{
		Parent:    parent,
		Variables: map[string]core.Value{},
		Commands:  map[string]core.Command{},
	}
	scope.Evaluator = core.NewCompilingEvaluator(
		variableResolver{scope},
		commandResolver{scope},
		nil,
		scope,
	)
	return scope
}

func (scope *PicolScope) resolveVariable(name string) core.Value {
	return scope.Variables[name]
}
func (scope *PicolScope) resolveCommand(name core.Value) core.Command {
	_, s := core.ValueToString(name)
	return scope.resolveNamedCommand(s)
}
func (scope *PicolScope) resolveNamedCommand(name string) core.Command {
	cmd := scope.Commands[name]
	if cmd != nil {
		return cmd
	}
	if scope.Parent != nil {
		return scope.Parent.resolveNamedCommand(name)
	}
	return nil
}

func asString(value core.Value) (s string) { _, s = core.ValueToString(value); return }

var EMPTY = core.OK(core.STR(""))

func ARITY_ERROR(signature string) core.Result {
	return core.ERROR(`wrong # args: should be "` + signature + `"`)
}

type simpleCommand struct {
	fn func(args []core.Value, context any) core.Result
}

func (command simpleCommand) Execute(args []core.Value, context any) core.Result {
	return command.fn(args, context)
}

func makeCommand(executeFn func(args []core.Value, context any) core.Result) core.Command {
	return simpleCommand{executeFn}
}

var addCmd = makeCommand(
	func(args []core.Value, _ any) core.Result {
		if len(args) < 2 {
			return ARITY_ERROR("+ arg ?arg ...?")
		}
		result, first := core.ValueToFloat(args[1])
		if result.Code != core.ResultCode_OK {
			return result
		}
		if len(args) == 2 {
			return core.OK(core.REAL(first))
		}
		total := first
		for i := 2; i < len(args); i++ {
			result, n := core.ValueToFloat(args[i])
			if result.Code != core.ResultCode_OK {
				return result
			}
			total += n
		}
		return core.OK(core.REAL(total))
	},
)
var subtractCmd = makeCommand(
	func(args []core.Value, _ any) core.Result {
		if len(args) < 2 {
			return ARITY_ERROR("- arg ?arg ...?")
		}
		result, first := core.ValueToFloat(args[1])
		if result.Code != core.ResultCode_OK {
			return result
		}
		if len(args) == 2 {
			return core.OK(core.REAL(-first))
		}
		total := first
		for i := 2; i < len(args); i++ {
			result, n := core.ValueToFloat(args[i])
			if result.Code != core.ResultCode_OK {
				return result
			}
			total -= n
		}
		return core.OK(core.REAL(total))
	},
)
var multiplyCmd = makeCommand(
	func(args []core.Value, _ any) core.Result {
		if len(args) < 2 {
			return ARITY_ERROR("* arg ?arg ...?")
		}
		result, first := core.ValueToFloat(args[1])
		if result.Code != core.ResultCode_OK {
			return result
		}
		if len(args) == 2 {
			return core.OK(core.REAL(first))
		}
		total := first
		for i := 2; i < len(args); i++ {
			result, n := core.ValueToFloat(args[i])
			if result.Code != core.ResultCode_OK {
				return result
			}
			total *= n
		}
		return core.OK(core.REAL(total))
	},
)
var divideCmd = makeCommand(
	func(args []core.Value, _ any) core.Result {
		if len(args) < 3 {
			return ARITY_ERROR("/ arg arg ?arg ...?")
		}
		result, first := core.ValueToFloat(args[1])
		if result.Code != core.ResultCode_OK {
			return result
		}
		total := first
		for i := 2; i < len(args); i++ {
			result, n := core.ValueToFloat(args[i])
			if result.Code != core.ResultCode_OK {
				return result
			}
			total /= n
		}
		return core.OK(core.REAL(total))
	},
)

func compareValuesCmd(
	name string,
	fn func(op1 core.Value, op2 core.Value) bool,
) core.Command {
	return makeCommand(func(args []core.Value, context any) core.Result {
		if len(args) != 3 {
			return ARITY_ERROR(name + ` arg arg`)
		}
		if fn(args[1], args[2]) {
			return core.OK(core.TRUE)
		} else {
			return core.OK(core.FALSE)
		}
	})
}

var eqCmd = compareValuesCmd(
	"==",
	func(op1 core.Value, op2 core.Value) bool { return op1 == op2 || asString(op1) == asString(op2) },
)
var neCmd = compareValuesCmd(
	"!=",
	func(op1 core.Value, op2 core.Value) bool { return op1 != op2 && asString(op1) != asString(op2) },
)

func compareNumbersCmd(
	name string,
	fn func(op1 float64, op2 float64) bool,
) core.Command {
	return makeCommand(func(args []core.Value, context any) core.Result {
		if len(args) != 3 {
			return ARITY_ERROR(name + ` arg arg`)
		}
		result1, op1 := core.ValueToFloat(args[1])
		if result1.Code != core.ResultCode_OK {
			return result1
		}
		result2, op2 := core.ValueToFloat(args[2])
		if result2.Code != core.ResultCode_OK {
			return result2
		}
		if fn(op1, op2) {
			return core.OK(core.TRUE)
		} else {
			return core.OK(core.FALSE)
		}
	})
}

var gtCmd = compareNumbersCmd(">", func(op1 float64, op2 float64) bool { return op1 > op2 })
var geCmd = compareNumbersCmd(">=", func(op1 float64, op2 float64) bool { return op1 >= op2 })
var ltCmd = compareNumbersCmd("<", func(op1 float64, op2 float64) bool { return op1 < op2 })
var leCmd = compareNumbersCmd("<=", func(op1 float64, op2 float64) bool { return op1 <= op2 })

var notCmd = makeCommand(
	func(args []core.Value, context any) core.Result {
		scope := context.(*PicolScope)
		if len(args) != 2 {
			return ARITY_ERROR("! arg")
		}
		result := evaluateCondition(args[1], scope)
		if result.Code != core.ResultCode_OK {
			return result
		}
		if result.Value.(core.BooleanValue).Value {
			return core.OK(core.FALSE)
		} else {
			return core.OK(core.TRUE)
		}
	},
)
var andCmd = makeCommand(
	func(args []core.Value, context any) core.Result {
		scope := context.(*PicolScope)
		if len(args) < 2 {
			return ARITY_ERROR("&& arg ?arg ...?")
		}
		r := true
		for i := 1; i < len(args); i++ {
			result := evaluateCondition(args[i], scope)
			if result.Code != core.ResultCode_OK {
				return result
			}
			if !result.Value.(core.BooleanValue).Value {
				r = false
				break
			}
		}

		if r {
			return core.OK(core.TRUE)
		} else {
			return core.OK(core.FALSE)
		}
	},
)
var orCmd = makeCommand(
	func(args []core.Value, context any) core.Result {
		scope := context.(*PicolScope)
		if len(args) < 2 {
			return ARITY_ERROR("|| arg ?arg ...?")
		}
		r := false
		for i := 1; i < len(args); i++ {
			result := evaluateCondition(args[i], scope)
			if result.Code != core.ResultCode_OK {
				return result
			}
			if result.Value.(core.BooleanValue).Value {
				r = true
				break
			}
		}

		if r {
			return core.OK(core.TRUE)
		} else {
			return core.OK(core.FALSE)
		}
	},
)

var ifCmd = makeCommand(
	func(args []core.Value, context any) core.Result {
		scope := context.(*PicolScope)
		if len(args) != 3 && len(args) != 5 {
			return ARITY_ERROR("if test script1 ?else script2?")
		}
		testResult := evaluateCondition(args[1], scope)
		if testResult.Code != core.ResultCode_OK {
			return testResult
		}
		var script core.ScriptValue
		if testResult.Value.(core.BooleanValue).Value {
			script = args[2].(core.ScriptValue)
		} else if len(args) == 3 {
			return EMPTY
		} else {
			script = args[4].(core.ScriptValue)
		}
		result := scope.Evaluator.EvaluateScript(script.Script)
		if result.Code != core.ResultCode_OK {
			return result
		}
		if result.Value == core.NIL {
			return EMPTY
		} else {
			return core.OK(result.Value)
		}
	},
)
var forCmd = makeCommand(
	func(args []core.Value, context any) core.Result {
		scope := context.(*PicolScope)
		if len(args) != 5 {
			return ARITY_ERROR("for start test next command")
		}
		start := args[1].(core.ScriptValue)
		test := args[2]
		next := args[3].(core.ScriptValue)
		script := args[4].(core.ScriptValue)
		var result core.Result
		result = scope.Evaluator.EvaluateScript(start.Script)
		if result.Code != core.ResultCode_OK {
			return result
		}
		for {
			result = evaluateCondition(test, scope)
			if result.Code != core.ResultCode_OK {
				return result
			}
			if !(result.Value.(core.BooleanValue)).Value {
				break
			}
			result = scope.Evaluator.EvaluateScript(script.Script)
			if result.Code == core.ResultCode_BREAK {
				break
			}
			if result.Code == core.ResultCode_CONTINUE {
				result = scope.Evaluator.EvaluateScript(next.Script)
				if result.Code != core.ResultCode_OK {
					return result
				}
				continue
			}
			if result.Code != core.ResultCode_OK {
				return result
			}
			result = scope.Evaluator.EvaluateScript(next.Script)
			if result.Code != core.ResultCode_OK {
				return result
			}
		}
		return EMPTY
	},
)
var whileCmd = makeCommand(
	func(args []core.Value, context any) core.Result {
		scope := context.(*PicolScope)
		if len(args) != 3 && len(args) != 5 {
			return ARITY_ERROR("while test script")
		}
		test := args[1]
		script := args[2].(core.ScriptValue)
		var result core.Result
		for {
			result = evaluateCondition(test, scope)
			if result.Code != core.ResultCode_OK {
				return result
			}
			if !(result.Value.(core.BooleanValue)).Value {
				break
			}
			result = scope.Evaluator.EvaluateScript(script.Script)
			if result.Code == core.ResultCode_BREAK {
				break
			}
			if result.Code == core.ResultCode_CONTINUE {
				continue
			}
			if result.Code != core.ResultCode_OK {
				return result
			}
		}
		return EMPTY
	},
)

func evaluateCondition(value core.Value, scope *PicolScope) core.Result {
	if value.Type() == core.ValueType_BOOLEAN {
		return core.OK(value)
	}
	if value.Type() == core.ValueType_INTEGER {
		return core.OK(core.BOOL(value.(core.IntegerValue).Value != 0))
	}
	if value.Type() == core.ValueType_SCRIPT {
		result := scope.Evaluator.EvaluateScript(value.(core.ScriptValue).Script)
		if result.Code != core.ResultCode_OK {
			return result
		}
		result, _ = core.BooleanValueFromValue(result.Value)
		return result
	}
	s := asString(value)
	if s == "true" || s == "yes" || s == "1" {
		return core.OK(core.TRUE)
	}
	if s == "false" || s == "no" || s == "0" {
		return core.OK(core.FALSE)
	}
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return core.ERROR(`invalid boolean "` + s + `"`)
	}
	return core.OK(core.BOOL(i != 0))
}

var setCmd = makeCommand(
	func(args []core.Value, context any) core.Result {
		scope := context.(*PicolScope)
		switch len(args) {
		case 2:
			{
				name := asString(args[1])
				value := scope.resolveVariable(name)
				if value != nil {
					return core.OK(value)
				}
				return core.ERROR(`can't read "` + name + `": no such variable`)
			}
		case 3:
			scope.Variables[asString(args[1])] = args[2]
			return core.OK(args[2])
		default:
			return ARITY_ERROR("set varName ?newValue?")
		}
	},
)
var incrCmd = makeCommand(
	func(args []core.Value, context any) core.Result {
		scope := context.(*PicolScope)
		var increment int64
		switch len(args) {
		case 2:
			increment = 1
		case 3:
			{
				result, i := core.ValueToInteger(args[2])
				if result.Code != core.ResultCode_OK {
					return result
				}
				increment = i
			}
		default:
			return ARITY_ERROR("incr varName ?increment?")
		}
		varName := asString(args[1])
		value, ok := scope.Variables[varName]
		var incremented core.Value
		if ok {
			result, i := core.ValueToInteger(value)
			if result.Code != core.ResultCode_OK {
				return result
			}
			incremented = core.INT(i + increment)
		} else {
			incremented = core.INT(increment)
		}
		scope.Variables[varName] = incremented
		return core.OK(incremented)
	},
)

type argSpec struct {
	name     string
	default_ core.Value
}

type procCommand struct {
	argspecs []argSpec
	body     core.ScriptValue
}

func (proc procCommand) Execute(args []core.Value, context any) core.Result {
	parent := context.(*PicolScope)
	scope := NewPicolScope(parent)
	var a = 1
	for p := 0; p < len(proc.argspecs); p, a = p+1, a+1 {
		argspec := proc.argspecs[p]
		var value core.Value
		if p == len(proc.argspecs)-1 && argspec.name == "args" {
			value = core.TUPLE(append([]core.Value{}, args[a:]...))
			a = len(args) - 1
		} else if p < len(args)-1 {
			value = args[a]
		} else if argspec.default_ != nil {
			value = argspec.default_
		} else {
			return ARITY_ERROR(argspecsToSignature(args[0], proc.argspecs))
		}
		scope.Variables[argspec.name] = value
	}
	if a < len(args) {
		return ARITY_ERROR(argspecsToSignature(args[0], proc.argspecs))
	}

	result := scope.Evaluator.EvaluateScript(proc.body.Script)
	if result.Code == core.ResultCode_ERROR {
		return result
	}
	if result.Value == core.NIL {
		return EMPTY
	} else {
		return core.OK(result.Value)
	}
}

func valueToArray(value core.Value) (core.Result, []core.Value) {
	switch value.Type() {
	case core.ValueType_TUPLE:
		return core.OK(core.NIL), value.(core.TupleValue).Values
	case core.ValueType_SCRIPT:
		{
			evaluator := core.NewCompilingEvaluator(nil, nil, nil, nil)
			values := []core.Value{}
			for _, sentence := range value.(core.ScriptValue).Script.Sentences {
				for _, word := range sentence.Words {
					if word.Value == nil {
						result := evaluator.EvaluateWord(word.Word)
						if result.Code != core.ResultCode_OK {
							return result, nil
						}
						values = append(values, result.Value)
					} else {
						values = append(values, word.Value)
					}
				}
			}
			return core.OK(core.NIL), values
		}
	default:
		return core.ERROR("unsupported list format"), nil
	}
}

func valueToArgspec(value core.Value) (core.Result, argSpec) {
	switch value.Type() {
	case core.ValueType_SCRIPT:
		{
			result, values := valueToArray(value)
			if result.Code != core.ResultCode_OK {
				return result, argSpec{}
			}
			if len(values) == 0 {
				return core.ERROR("argument with no name"), argSpec{}
			}
			name := asString(values[0])
			if name == "" {
				return core.ERROR("argument with no name"), argSpec{}
			}
			switch len(values) {
			case 1:
				return core.OK(core.NIL), argSpec{name: name}
			case 2:
				return core.OK(core.NIL), argSpec{name: name, default_: values[1]}
			default:
				return core.ERROR(
					`too many fields in argument specifier "` + asString(value) + `"`,
				), argSpec{}
			}
		}
	default:
		return core.OK(core.NIL), argSpec{name: asString(value)}
	}
}
func valueToArgspecs(value core.Value) (core.Result, []argSpec) {
	result, values := valueToArray(value)
	if result.Code != core.ResultCode_OK {
		return result, nil
	}
	argspecs := make([]argSpec, len(values))
	for i, value := range values {
		result, argspec := valueToArgspec(value)
		if result.Code != core.ResultCode_OK {
			return result, nil
		}
		argspecs[i] = argspec
	}
	return core.OK(core.NIL), argspecs
}
func argspecsToSignature(name core.Value, argspecs []argSpec) string {
	result := asString(name)
	for i, argspec := range argspecs {
		result += " "
		if i == len(argspecs)-1 && argspec.name == "args" {
			result += "?arg ...?"
		} else if argspec.default_ != nil {
			result += "?" + argspec.name + "?"
		} else {
			result += argspec.name
		}
	}
	return result
}

var procCmd = makeCommand(
	func(args []core.Value, context any) core.Result {
		scope := context.(*PicolScope)
		if len(args) != 4 {
			return ARITY_ERROR("proc name args body")
		}
		name, _argspecs, body := args[1], args[2], args[3]
		result, argspecs := valueToArgspecs(_argspecs)
		if result.Code != core.ResultCode_OK {
			return result
		}
		scope.Commands[asString(name)] = procCommand{argspecs, body.(core.ScriptValue)}
		return EMPTY
	},
)

var returnCmd = makeCommand(
	func(args []core.Value, _ any) core.Result {
		if len(args) > 2 {
			return ARITY_ERROR("return ?result?")
		}
		if len(args) == 2 {
			return core.RETURN(args[1])
		} else {
			return core.RETURN(core.STR(""))
		}
	},
)
var breakCmd = makeCommand(
	func(args []core.Value, context any) core.Result {
		if len(args) != 1 {
			return ARITY_ERROR("break")
		}
		return core.BREAK(core.NIL)
	},
)
var continueCmd = makeCommand(
	func(args []core.Value, context any) core.Result {
		if len(args) != 1 {
			return ARITY_ERROR("continue")
		}
		return core.CONTINUE(core.NIL)
	},
)
var errorCmd = makeCommand(
	func(args []core.Value, context any) core.Result {
		if len(args) != 2 {
			return ARITY_ERROR("error message")
		}
		return core.Result{Code: core.ResultCode_ERROR, Value: args[1]}
	},
)

func InitPicolCommands(scope *PicolScope) {
	scope.Commands["+"] = addCmd
	scope.Commands["-"] = subtractCmd
	scope.Commands["*"] = multiplyCmd
	scope.Commands["/"] = divideCmd
	scope.Commands["=="] = eqCmd
	scope.Commands["!="] = neCmd
	scope.Commands[">"] = gtCmd
	scope.Commands[">="] = geCmd
	scope.Commands["<"] = ltCmd
	scope.Commands["<="] = leCmd
	scope.Commands["!"] = notCmd
	scope.Commands["&&"] = andCmd
	scope.Commands["||"] = orCmd
	scope.Commands["if"] = ifCmd
	scope.Commands["for"] = forCmd
	scope.Commands["while"] = whileCmd
	scope.Commands["set"] = setCmd
	scope.Commands["incr"] = incrCmd
	scope.Commands["proc"] = procCmd
	scope.Commands["return"] = returnCmd
	scope.Commands["break"] = breakCmd
	scope.Commands["continue"] = continueCmd
	scope.Commands["error"] = errorCmd
}
