package helena_dialect

import "helena/core"

const PARSE_SIGNATURE = "parse source"

type parseCmd struct{}

func (parseCmd) Execute(args []core.Value, _ any) core.Result {
	if len(args) != 2 {
		return ARITY_ERROR(PARSE_SIGNATURE)
	}
	result := core.ValueToString(args[1])
	if result.Code != core.ResultCode_OK {
		return result.AsResult()
	}
	source := result.Data
	tokenizer := core.Tokenizer{}
	parser := core.NewParser(nil)
	parseResult := parser.Parse(
		tokenizer.Tokenize(source),
	)
	if !parseResult.Success {
		return core.ERROR(parseResult.Message)
	}
	return core.OK(core.NewScriptValue(*parseResult.Script, source))
}
func (parseCmd) Help(args []core.Value, options core.CommandHelpOptions, _ any) core.Result {
	if len(args) > 2 {
		return ARITY_ERROR(PARSE_SIGNATURE)
	}
	return core.OK(core.STR(PARSE_SIGNATURE))
}

type scriptCommand struct {
	scope    *Scope
	ensemble *EnsembleCommand
}

func newScriptCommand(scope *Scope) *scriptCommand {
	list := &scriptCommand{}
	list.scope = scope.NewChildScope()
	argspec := ArgspecValueFromValue(core.LIST([]core.Value{core.STR("value")})).Data
	list.ensemble = NewEnsembleCommand(list.scope, argspec)
	return list
}
func (script *scriptCommand) Execute(args []core.Value, context any) core.Result {
	if len(args) == 2 {
		return valueToScript(args[1])
	}
	return script.ensemble.Execute(args, context)
}
func (script *scriptCommand) Help(args []core.Value, options core.CommandHelpOptions, context any) core.Result {
	return script.ensemble.Help(args, options, context)
}

const SCRIPT_LENGTH_SIGNATURE = "script value length"

type scriptLengthCmd struct{}

func (scriptLengthCmd) Execute(args []core.Value, context any) core.Result {
	if len(args) != 2 {
		return ARITY_ERROR(SCRIPT_LENGTH_SIGNATURE)
	}
	result := valueToScript(args[1])
	if result.Code != core.ResultCode_OK {
		return result
	}
	return core.OK(core.INT(int64(len(result.Value.(core.ScriptValue).Script.Sentences))))
}
func (scriptLengthCmd) Help(args []core.Value, options core.CommandHelpOptions, context any) core.Result {
	if len(args) > 2 {
		return ARITY_ERROR(SCRIPT_LENGTH_SIGNATURE)
	}
	return core.OK(core.STR(SCRIPT_LENGTH_SIGNATURE))
}

const SCRIPT_APPEND_SIGNATURE = "script value append ?script ...?"

type scriptAppendCmd struct{}

func (scriptAppendCmd) Execute(args []core.Value, context any) core.Result {
	result := valueToScript(args[1])
	if result.Code != core.ResultCode_OK {
		return result
	}
	if len(args) == 2 {
		return result
	}
	script := core.Script{}
	script.Sentences = append(script.Sentences, result.Value.(core.ScriptValue).Script.Sentences...)
	for i := 2; i < len(args); i++ {
		result2 := valueToScript(args[i])
		if result2.Code != core.ResultCode_OK {
			return result2
		}
		script.Sentences = append(script.Sentences, result2.Value.(core.ScriptValue).Script.Sentences...)
	}
	return core.OK(core.NewScriptValueWithNoSource(script))
}
func (scriptAppendCmd) Help(args []core.Value, options core.CommandHelpOptions, context any) core.Result {
	return core.OK(core.STR(SCRIPT_APPEND_SIGNATURE))
}

const SCRIPT_SPLIT_SIGNATURE = "script value split"

type scriptSplitCmd struct{}

func (scriptSplitCmd) Execute(args []core.Value, context any) core.Result {
	if len(args) != 2 {
		return ARITY_ERROR(SCRIPT_SPLIT_SIGNATURE)
	}
	result := valueToScript(args[1])
	if result.Code != core.ResultCode_OK {
		return result
	}
	sentences := []core.Value{}
	for _, sentence := range result.Value.(core.ScriptValue).Script.Sentences {
		script := core.Script{}
		script.Sentences = append(script.Sentences, sentence)
		sentences = append(sentences, core.NewScriptValueWithNoSource(script))
	}
	return core.OK(core.LIST(sentences))
}
func (scriptSplitCmd) Help(args []core.Value, options core.CommandHelpOptions, context any) core.Result {
	if len(args) > 2 {
		return ARITY_ERROR(SCRIPT_SPLIT_SIGNATURE)
	}
	return core.OK(core.STR(SCRIPT_SPLIT_SIGNATURE))
}

func valueToScript(value core.Value) core.Result {
	switch value.Type() {
	case core.ValueType_SCRIPT:
		return core.OK(value)
	case core.ValueType_TUPLE:
		return core.OK(tupleToScript(value.(core.TupleValue)))
	default:
		return core.ERROR("value must be a script or tuple")
	}
}

func tupleToScript(tuple core.TupleValue) core.ScriptValue {
	script := core.Script{}
	if len(tuple.Values) != 0 {
		sentence := core.Sentence{}
		for _, value := range tuple.Values {
			sentence.Words = append(sentence.Words, core.WordOrValue{Value: value})
		}
		script.Sentences = append(script.Sentences, sentence)
	}
	return core.NewScriptValueWithNoSource(script)
}

func registerScriptCommands(scope *Scope) {
	scope.RegisterNamedCommand("parse", parseCmd{})
	command := newScriptCommand(scope)
	scope.RegisterNamedCommand("script", command)
	command.scope.RegisterNamedCommand("length", scriptLengthCmd{})
	command.scope.RegisterNamedCommand("append", scriptAppendCmd{})
	command.scope.RegisterNamedCommand("split", scriptSplitCmd{})
}
