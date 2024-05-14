package helena_dialect

import "helena/core"

type Argspec struct {
	Args         []Argument
	NbRequired   uint
	NbOptional   uint
	HasRemainder bool
	OptionSlots  map[string]uint
}

func NewArgspec(args []Argument) Argspec {
	nbRequired := uint(0)
	nbOptional := uint(0)
	hasRemainder := false
	var optionSlots map[string]uint
	for i, arg := range args {
		if arg.Option != nil {
			if arg.Type == ArgumentType_REQUIRED {
				nbRequired += 2
			}
			if optionSlots == nil {
				optionSlots = map[string]uint{}
			}
			for _, name := range arg.Option.Names {
				optionSlots[name] = uint(i)
			}
		} else {
			switch arg.Type {
			case ArgumentType_REQUIRED:
				nbRequired++
			case ArgumentType_OPTIONAL:
				nbOptional++
			case ArgumentType_REMAINDER:
				hasRemainder = true
			}
		}
	}
	return Argspec{args, nbRequired, nbOptional, hasRemainder, optionSlots}
}
func (argspec Argspec) IsVariadic() bool {
	return (argspec.NbOptional > 0) || argspec.HasRemainder
}
func (argspec Argspec) HasOptions() bool {
	return len(argspec.OptionSlots) > 0
}

type ArgspecValue struct {
	Argspec Argspec
}

func (ArgspecValue) Type() core.ValueType {
	return core.ValueType_CUSTOM
}
func (ArgspecValue) CustomType() core.CustomValueType {
	return core.CustomValueType{Name: "argspec"}
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
func (argspec ArgspecValue) Usage(skip uint) string {
	return BuildUsage(argspec.Argspec.Args, skip)
}
func (argspec ArgspecValue) CheckArity(values []core.Value, skip uint) bool {
	if argspec.Argspec.HasOptions() {
		// There is no fast way to check arity without parsing all options, so
		// just check that there are enough to cover all the required ones
		return uint(len(values))-skip >= argspec.Argspec.NbRequired
	}
	return (uint(len(values))-skip >= argspec.Argspec.NbRequired &&
		(argspec.Argspec.HasRemainder ||
			uint(len(values))-skip <= argspec.Argspec.NbRequired+argspec.Argspec.NbOptional))
}
func (argspec ArgspecValue) ApplyArguments(
	scope *Scope,
	values []core.Value,
	skip uint,
	setArgument func(name string, value core.Value) core.Result,
) core.Result {
	if !argspec.Argspec.HasOptions() {
		// Use faster algorithm for the common case with all positionals
		return argspec.applyPositionals(scope, values, skip, setArgument)
	}
	result := argspec.findSlots(values, skip)
	if result.Code != core.ResultCode_OK {
		return result.AsResult()
	}
	slots := result.Data.slots
	remainders := result.Data.remainders
	for slot, arg := range argspec.Argspec.Args {
		var value core.Value
		switch arg.Type {
		case ArgumentType_REQUIRED:
			if slots[slot] < 0 {
				if arg.Option != nil {
					return core.ERROR(
						`missing value for option "` + OptionName(arg.Option.Names) + `"`,
					)
				} else {
					return core.ERROR(`missing value for argument "` + arg.Name + `"`)
				}
			}
			value = values[slots[slot]]
		case ArgumentType_OPTIONAL:
			if slots[slot] >= 0 {
				if arg.Option != nil && arg.Option.Type == OptionType_FLAG {
					value = core.TRUE
				} else {
					value = values[slots[slot]]
				}
			} else if arg.Option != nil && arg.Option.Type == OptionType_FLAG {
				value = core.FALSE
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
				// Skip missing optional
				continue
			}
		case ArgumentType_REMAINDER:
			{
				if slots[slot] < 0 {
					// No remainder
					value = core.TUPLE([]core.Value{})
				} else {
					value = core.TUPLE(
						values[slots[slot] : slots[slot]+int(remainders)],
					)
				}
			}
		}
		result := argspec.setArgument(scope, arg, value, setArgument)
		// TODO handle YIELD?
		if result.Code != core.ResultCode_OK {
			return result
		}
	}
	return core.OK(core.NIL)
}

type findSlotsResult = struct {
	slots      []int
	remainders uint
}

func (argspec ArgspecValue) findSlots(
	values []core.Value,
	skip uint,
) core.TypedResult[findSlotsResult] {
	nbRequired := argspec.Argspec.NbRequired
	nbOptional := argspec.Argspec.NbOptional
	remainders := uint(0)
	nbArgs := uint(len(argspec.Argspec.Args))
	slots := make([]int, nbArgs)
	for i := range slots {
		slots[i] = -1
	}
	// Consume positional arguments and options alternatively
	slot := uint(0)
	i := int(skip)
	for i < len(values) {
		// Positional arguments in order
		firstSlot := slot
		lastSlot := slot
		for lastSlot < nbArgs {
			arg := argspec.Argspec.Args[lastSlot]
			if arg.Option != nil {
				break
			}
			lastSlot++
		}
		for i < len(values) && slot < lastSlot {
			arg := argspec.Argspec.Args[slot]
			remaining := len(values) - i
			switch arg.Type {
			case ArgumentType_REQUIRED:
				nbRequired--
				slots[slot] = i
				i++
			case ArgumentType_OPTIONAL:
				if uint(remaining) > nbRequired {
					nbOptional--
					slots[slot] = i
					i++
				}
			case ArgumentType_REMAINDER:
				if uint(remaining) > nbRequired+nbOptional {
					remainders = uint(remaining) - nbRequired - nbOptional
					slots[slot] = i
					i += int(remainders)
				}
			}
			slot++
		}
		if i >= len(values) {
			break
		}

		// Options out-of-order
		requiredOptions := 0
		firstSlot = slot
		for lastSlot < nbArgs {
			arg := argspec.Argspec.Args[lastSlot]
			if arg.Option == nil {
				break
			}
			if arg.Type == ArgumentType_REQUIRED {
				requiredOptions++
			}
			lastSlot++
		}
		nbOptions := uint(0)
		for i < len(values) && nbOptions < lastSlot-firstSlot {
			result := core.ValueToString(values[i])
			if result.Code != core.ResultCode_OK {
				if requiredOptions == 0 {
					break
				}
				return core.ERROR_T[findSlotsResult]("invalid option")
			}
			optname := result.Data
			if optname == "--" {
				if requiredOptions == 0 {
					break
				}
				return core.ERROR_T[findSlotsResult]("unexpected option terminator")
			}
			if _, ok := argspec.Argspec.OptionSlots[optname]; !ok {
				if requiredOptions == 0 {
					break
				}
				return core.ERROR_T[findSlotsResult](`unknown option "` + optname + `"`)
			}
			optionSlot := argspec.Argspec.OptionSlots[optname]
			if optionSlot < firstSlot || optionSlot >= lastSlot {
				return core.ERROR_T[findSlotsResult](`unexpected option "` + optname + `"`)
			}
			arg := argspec.Argspec.Args[optionSlot]
			if slots[optionSlot] >= 0 {
				return core.ERROR_T[findSlotsResult](
					`duplicate values for option "` + OptionName(arg.Option.Names) + `"`,
				)
			}
			nbOptions++
			switch arg.Option.Type {
			case OptionType_FLAG:
				slots[optionSlot] = i
				i++
			case OptionType_OPTION:
				switch arg.Type {
				case ArgumentType_REQUIRED:
					nbRequired -= 2
					slots[optionSlot] = i + 1
					requiredOptions--
					i += 2
				case ArgumentType_OPTIONAL:
					slots[optionSlot] = i + 1
					i += 2
				default:
					panic("CANTHAPPEN")
				}
			}
		}
		if i < len(values) {
			// Skip first trailing terminator
			result := core.ValueToString(values[i])
			if result.Code == core.ResultCode_OK && result.Data == "--" {
				i++
			}
		}
		slot = lastSlot
		if slot >= nbArgs {
			break
		}
	}
	if i < len(values) {
		return core.ERROR_T[findSlotsResult]("extra values after arguments")
	}
	return core.OK_T(core.NIL, findSlotsResult{slots, remainders})
}
func (argspec ArgspecValue) applyPositionals(
	scope *Scope,
	values []core.Value,
	skip uint,
	setArgument func(name string, value core.Value) core.Result,
) core.Result {
	total := uint(len(values)) - skip
	nbNonRequired := total - argspec.Argspec.NbRequired
	nbOptional := min(argspec.Argspec.NbOptional, nbNonRequired)
	remainders := nbNonRequired - nbOptional
	i := skip
	for _, arg := range argspec.Argspec.Args {
		var value core.Value
		switch arg.Type {
		case ArgumentType_REQUIRED:
			value = values[i]
			i++
		case ArgumentType_OPTIONAL:
			if nbOptional > 0 {
				nbOptional--
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
			value = core.TUPLE(append([]core.Value{}, values[i:i+remainders]...))
			i += remainders
		}
		result := argspec.setArgument(scope, arg, value, setArgument)
		// TODO handle YIELD?
		if result.Code != core.ResultCode_OK {
			return result
		}
	}
	return core.OK(core.NIL)
}
func (ArgspecValue) setArgument(
	scope *Scope,
	arg Argument,
	value core.Value,
	setArgument func(name string, value core.Value) core.Result,
) core.Result {
	if arg.Guard != nil {
		process := scope.PrepareTupleValue(core.TUPLE([]core.Value{arg.Guard, value}))
		result := process.Run()
		// TODO handle YIELD?
		if result.Code != core.ResultCode_OK {
			return result
		}
		value = result.Value
	}
	return setArgument(arg.Name, value)
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
	if !value.CheckArity(values, 0) {
		return core.ERROR(`wrong # values: should be "` + value.Usage(0) + `"`)
	}
	return value.ApplyArguments(scope, values, 0, func(name string, value core.Value) core.Result {
		return scope.SetNamedVariable(name, value)
	})
}
func (argspecSetCmd) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	if len(args) > 3 {
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
