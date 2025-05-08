package helena_dialect

import "helena/core"

func USAGE_ARGSPEC(name core.Value, def string, argspec ArgspecValue, options core.CommandHelpOptions) string {
	prefix := USAGE_PREFIX(name, def, options)
	var usage string
	if options.Skip > 0 {
		usage = argspec.Usage(options.Skip - 1)
	} else {
		usage = argspec.Usage(0)
	}
	if len(usage) > 0 {
		if len(prefix) > 0 {
			return prefix + " " + usage
		}
		return usage
	}
	return prefix
}

type Argspec struct {
	Args         []Argument
	Names        []string
	NbRequired   uint
	NbOptional   uint
	HasRemainder bool
	HasOptions   bool
	OptionSlots  map[string]uint
	HasGuards    bool
}

func NewArgspec(args []Argument) Argspec {
	nbRequired := uint(0)
	nbOptional := uint(0)
	hasRemainder := false
	hasOptions := false
	hasGuard := false
	var names []string = make([]string, len(args))
	var optionSlots map[string]uint
	for i, arg := range args {
		names[i] = arg.Name
		if arg.Option != nil {
			hasOptions = true
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
		if arg.Guard != nil {
			hasGuard = true
		}
	}
	return Argspec{args, names, nbRequired, nbOptional, hasRemainder, hasOptions, optionSlots, hasGuard}
}
func (argspec Argspec) IsVariadic() bool {
	return (argspec.NbOptional > 0) || argspec.HasRemainder
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
func ArgspecValueFromValue(value core.Value) (core.Result, ArgspecValue) {
	if v, ok := value.(ArgspecValue); ok {
		return core.OK(v), v
	}
	result, args := buildArguments(value)
	if result.Code != core.ResultCode_OK {
		return result, ArgspecValue{}
	}
	v := NewArgspecValue(NewArgspec(args))
	return core.OK(v), v
}
func (argspec ArgspecValue) Usage(skip uint) string {
	return BuildUsage(argspec.Argspec.Args, skip)
}
func (argspec ArgspecValue) CheckArity(values []core.Value, skip uint) bool {
	if argspec.Argspec.HasOptions {
		// There is no fast way to check arity without parsing all options, so
		// just check that there are enough to cover all the required ones
		return uint(len(values))-skip >= argspec.Argspec.NbRequired
	}
	return (uint(len(values))-skip >= argspec.Argspec.NbRequired &&
		(argspec.Argspec.HasRemainder ||
			uint(len(values))-skip <= argspec.Argspec.NbRequired+argspec.Argspec.NbOptional))
}
func (argspec ArgspecValue) CollectArguments(
	scope *Scope,
	args []core.Value,
	skip uint,
) (core.Result, []core.Value) {
	slotValues := make([]core.Value, len(argspec.Argspec.Args))
	return argspec.setSlotValues(scope, args, skip, slotValues), slotValues
}
func (argspec ArgspecValue) setSlotValues(
	scope *Scope,
	args []core.Value,
	skip uint,
	slotValues []core.Value,
) core.Result {
	if !argspec.Argspec.HasOptions {
		// Use faster algorithm for the common case with all positionals
		return argspec.setPositionalSlotValues(scope, args, skip, slotValues)
	}
	result, data := argspec.findSlots(args, skip)
	if result.Code != core.ResultCode_OK {
		return result
	}
	slots := data.slots
	remainders := data.remainders
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
			value = args[slots[slot]]
		case ArgumentType_OPTIONAL:
			if slots[slot] >= 0 {
				if arg.Option != nil && arg.Option.Type == OptionType_FLAG {
					value = core.TRUE
				} else {
					value = args[slots[slot]]
				}
			} else if arg.Option != nil && arg.Option.Type == OptionType_FLAG {
				value = core.FALSE
			} else if arg.Default != nil {
				if arg.Default.Type() == core.ValueType_SCRIPT {
					program := scope.CompileScriptValue(
						arg.Default.(core.ScriptValue),
					)
					result := scope.Execute(program, nil)
					switch result.Code {
					case core.ResultCode_OK:
						value = result.Value
					case core.ResultCode_ERROR:
						return result
					default:
						return core.ERROR("unexpected " + core.RESULT_CODE_NAME(result))
					}
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
						args[slots[slot] : slots[slot]+int(remainders)],
					)
				}
			}
		}
		result := argspec.setSlotValue(scope, arg, value, slot, slotValues)
		switch result.Code {
		case core.ResultCode_OK:
		case core.ResultCode_ERROR:
			return result
		default:
			return core.ERROR("unexpected " + core.RESULT_CODE_NAME(result))
		}
	}
	return core.OK(core.NIL)
}

type findSlotsResult = struct {
	slots      []int
	remainders uint
}

func (argspec ArgspecValue) findSlots(
	args []core.Value,
	skip uint,
) (core.Result, findSlotsResult) {
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
	for i < len(args) {
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
		for i < len(args) && slot < lastSlot {
			arg := argspec.Argspec.Args[slot]
			remaining := len(args) - i
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
		if i >= len(args) {
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
		for i < len(args) && nbOptions < lastSlot-firstSlot {
			result, optname := core.ValueToString(args[i])
			if result.Code != core.ResultCode_OK {
				if requiredOptions == 0 {
					break
				}
				return core.ERROR("invalid option"), findSlotsResult{}
			}
			if optname == "--" {
				if requiredOptions == 0 {
					break
				}
				return core.ERROR("unexpected option terminator"), findSlotsResult{}
			}
			if _, ok := argspec.Argspec.OptionSlots[optname]; !ok {
				if requiredOptions == 0 {
					break
				}
				return core.ERROR(`unknown option "` + optname + `"`), findSlotsResult{}
			}
			optionSlot := argspec.Argspec.OptionSlots[optname]
			if optionSlot < firstSlot || optionSlot >= lastSlot {
				return core.ERROR(`unexpected option "` + optname + `"`), findSlotsResult{}
			}
			arg := argspec.Argspec.Args[optionSlot]
			if slots[optionSlot] >= 0 {
				return core.ERROR(
					`duplicate values for option "` + OptionName(arg.Option.Names) + `"`,
				), findSlotsResult{}
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
		if i < len(args) {
			// Skip first trailing terminator
			result, optname := core.ValueToString(args[i])
			if result.Code == core.ResultCode_OK && optname == "--" {
				i++
			}
		}
		slot = lastSlot
		if slot >= nbArgs {
			break
		}
	}
	if i < len(args) {
		return core.ERROR("extra values after arguments"), findSlotsResult{}
	}
	return core.OK(core.NIL), findSlotsResult{slots, remainders}
}
func (argspec ArgspecValue) setPositionalSlotValues(
	scope *Scope,
	args []core.Value,
	skip uint,
	slotValues []core.Value,
) core.Result {
	total := uint(len(args)) - skip
	nbNonRequired := total - argspec.Argspec.NbRequired
	nbOptional := min(argspec.Argspec.NbOptional, nbNonRequired)
	remainders := nbNonRequired - nbOptional
	i := skip
	for slot, arg := range argspec.Argspec.Args {
		var value core.Value
		switch arg.Type {
		case ArgumentType_REQUIRED:
			value = args[i]
			i++
		case ArgumentType_OPTIONAL:
			if nbOptional > 0 {
				nbOptional--
				value = args[i]
				i++
			} else if arg.Default != nil {
				if arg.Default.Type() == core.ValueType_SCRIPT {
					program := scope.CompileScriptValue(
						arg.Default.(core.ScriptValue),
					)
					result := scope.Execute(program, nil)
					switch result.Code {
					case core.ResultCode_OK:
						value = result.Value
					case core.ResultCode_ERROR:
						return result
					default:
						return core.ERROR("unexpected " + core.RESULT_CODE_NAME(result))
					}
				} else {
					value = arg.Default
				}
			} else {
				continue // Skip missing optional
			}
		case ArgumentType_REMAINDER:
			value = core.TUPLE(append([]core.Value{}, args[i:i+remainders]...))
			i += remainders
		}
		result := argspec.setSlotValue(scope, arg, value, slot, slotValues)
		switch result.Code {
		case core.ResultCode_OK:
		case core.ResultCode_ERROR:
			return result
		default:
			return core.ERROR("unexpected " + core.RESULT_CODE_NAME(result))
		}
	}
	return core.OK(core.NIL)
}
func (ArgspecValue) setSlotValue(
	scope *Scope,
	arg Argument,
	value core.Value,
	slot int,
	slotValues []core.Value,
) core.Result {
	if arg.Guard != nil {
		program := scope.CompilePair(arg.Guard, value)
		process := scope.PrepareProcess(program)
		result := process.Run()
		// TODO handle YIELD?
		if result.Code != core.ResultCode_OK {
			return result
		}
		value = result.Value
	}
	slotValues[slot] = value
	return core.OK(core.NIL)
}

type argspecCommand struct {
	scope    *Scope
	ensemble *EnsembleCommand
}

func newArgspecCommand(scope *Scope) argspecCommand {
	subscope := scope.NewChildScope()
	_, argspec := ArgspecValueFromValue(core.LIST([]core.Value{core.STR("value")}))
	ensemble := NewEnsembleCommand(subscope, argspec)
	return argspecCommand{subscope, ensemble}
}
func (argspec argspecCommand) Execute(args []core.Value, context any) core.Result {
	scope := context.(*Scope)
	if len(args) == 2 {
		result, _ := ArgspecValueFromValue(args[1])
		return result
	}
	return argspec.ensemble.Execute(args, scope)
}
func (argspec argspecCommand) Resume(result core.Result, context any) core.Result {
	return argspec.ensemble.Resume(result, context)
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
	result, value := ArgspecValueFromValue(args[1])
	if result.Code != core.ResultCode_OK {
		return result
	}
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
	result, argspec := ArgspecValueFromValue(args[1])
	if result.Code != core.ResultCode_OK {
		return result
	}
	result2, values := ValueToArray(args[2])
	if result2.Code != core.ResultCode_OK {
		return result2
	}
	if !argspec.CheckArity(values, 0) {
		return core.ERROR(`wrong # values: should be "` + argspec.Usage(0) + `"`)
	}
	result3, values2 := argspec.CollectArguments(scope, values, 0)
	if result3.Code != core.ResultCode_OK {
		return result3
	}
	return scope.SetNamedVariables(argspec.Argspec.Names, values2)
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
