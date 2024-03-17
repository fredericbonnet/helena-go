package helena_dialect

import "helena/core"

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
func (value ArgspecValue) Usage(skip uint) string {
	return BuildUsage(value.Argspec.Args, skip)
}
func (value ArgspecValue) CheckArity(values []core.Value, skip uint) bool {
	return (uint(len(values))-skip >= value.Argspec.NbRequired &&
		(value.Argspec.HasRemainder ||
			uint(len(values))-skip <= value.Argspec.NbRequired+value.Argspec.NbOptional))
}
func (value ArgspecValue) ApplyArguments(
	scope *Scope,
	values []core.Value,
	skip uint,
	setArgument func(name string, value core.Value) core.Result,
) core.Result {
	nonRequired := uint(len(values)) - skip - value.Argspec.NbRequired
	optionals := min(value.Argspec.NbOptional, nonRequired)
	remainders := nonRequired - optionals
	i := skip
	for _, arg := range value.Argspec.Args {
		var value core.Value
		switch arg.Type {
		case ArgumentType_REQUIRED:
			value = values[i]
			i++
		case ArgumentType_OPTIONAL:
			if optionals > 0 {
				optionals--
				value = values[i]
				i++
			} else if arg.Default != nil {
				if arg.Default.Type() == core.ValueType_SCRIPT {
					body := arg.Default.(core.ScriptValue)
					result := scope.ExecuteScriptValue(body)
					// TODO handle YIELD?
					if result.Code != core.ResultCode_OK {
						return result
					}
					value = result.Value
				} else {
					value = arg.Default
				}
			} else {
				continue // Skip missing optional
			}
		case ArgumentType_REMAINDER:
			value = core.TUPLE(values[i : i+remainders])
			i += remainders
		}
		if arg.Guard != nil {
			process := scope.PrepareTupleValue(core.TUPLE([]core.Value{arg.Guard, value}).(core.TupleValue))
			result := process.Run()
			// TODO handle YIELD?
			if result.Code != core.ResultCode_OK {
				return result
			}
			value = result.Value
		}
		result := setArgument(arg.Name, value)
		// TODO handle YIELD?
		if result.Code != core.ResultCode_OK {
			return result
		}
	}
	return core.OK(core.NIL)
}

func (value ArgspecValue) SetArguments(values []core.Value, scope *Scope) core.Result {
	if !value.CheckArity(values, 0) {
		return core.ERROR(`wrong # values: should be "` + value.Usage(0) + `"`)
	}
	return value.ApplyArguments(scope, values, 0, func(name string, value core.Value) core.Result {
		return scope.SetNamedVariable(name, value)
	})
}

type argspecCommand struct {
	scope    *Scope
	ensemble *EnsembleCommand
}

func newArgspecCommand(scope *Scope) argspecCommand {
	subscope := NewScope(scope, false)
	argspec := ArgspecValueFromValue(core.LIST([]core.Value{core.STR("value")})).Data
	ensemble := NewEnsembleCommand(subscope, argspec)
	return argspecCommand{subscope, ensemble}
}
func (argspec argspecCommand) Execute(args []core.Value, context any) core.Result {
	scope := context.(*Scope)
	if len(args) == 2 {
		return ArgspecValueFromValue(args[1]).AsResult()
	}
	return argspec.ensemble.Execute(args, scope)
}
func (argspec argspecCommand) Help(args []core.Value, options core.CommandHelpOptions, context any) core.Result {
	return argspec.ensemble.Help(args, options, context)
}

const ARGSPEC_USAGE_SIGNATURE = "argspec value usage"

type argspecUsageCmd struct{}

func (argspecUsageCmd) Execute(args []core.Value, _ any) core.Result {
	if len(args) != 2 {
		return ARITY_ERROR(ARGSPEC_USAGE_SIGNATURE)
	}
	result := ArgspecValueFromValue(args[1])
	if result.Code != core.ResultCode_OK {
		return result.AsResult()
	}
	value := result.Data
	return core.OK(core.STR(value.Usage(0)))
}
func (argspecUsageCmd) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	if len(args) > 2 {
		return ARITY_ERROR(ARGSPEC_USAGE_SIGNATURE)
	}
	return core.OK(core.STR(ARGSPEC_USAGE_SIGNATURE))
}

const ARGSPEC_SET_SIGNATURE = "argspec value set values"

type argspecSetCmd struct{}

func (argspecSetCmd) Execute(args []core.Value, context any) core.Result {
	scope := context.(*Scope)
	if len(args) != 3 {
		return ARITY_ERROR(ARGSPEC_SET_SIGNATURE)
	}
	result := ArgspecValueFromValue(args[1])
	if result.Code != core.ResultCode_OK {
		return result.AsResult()
	}
	value := result.Data
	result2 := ValueToArray(args[2])
	if result2.Code != core.ResultCode_OK {
		return result2.AsResult()
	}
	values := result2.Data
	return value.SetArguments(values, scope)
}
func (argspecSetCmd) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	if len(args) > 2 {
		return ARITY_ERROR(ARGSPEC_SET_SIGNATURE)
	}
	return core.OK(core.STR(ARGSPEC_SET_SIGNATURE))
}

func registerArgspecCommands(scope *Scope) {
	command := newArgspecCommand(scope)
	scope.RegisterNamedCommand("argspec", command)
	command.scope.RegisterNamedCommand("usage", argspecUsageCmd{})
	command.scope.RegisterNamedCommand("set", argspecSetCmd{})
}
