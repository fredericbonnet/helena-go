package helena_dialect

import "helena/core"

const LET_SIGNATURE = "let constname value"

type letCmd struct{}

func (letCmd) Execute(args []core.Value, context any) core.Result {
	scope := context.(*Scope)
	switch len(args) {
	case 3:
		return DestructureValue(
			func(name core.Value, value core.Value, check bool) core.Result {
				return scope.DestructureConstant(name, value, check)
			},
			args[1],
			args[2],
		)
	default:
		return ARITY_ERROR(LET_SIGNATURE)
	}
}
func (letCmd) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	if len(args) > 3 {
		return ARITY_ERROR(LET_SIGNATURE)
	}
	return core.OK(core.STR(LET_SIGNATURE))
}

const SET_SIGNATURE = "set varname value"

type setCmd struct{}

func (setCmd) Execute(args []core.Value, context any) core.Result {
	scope := context.(*Scope)
	switch len(args) {
	case 3:
		return DestructureValue(
			func(name core.Value, value core.Value, check bool) core.Result {
				return scope.DestructureVariable(name, value, check)
			},
			args[1],
			args[2],
		)
	default:
		return ARITY_ERROR(SET_SIGNATURE)
	}
}
func (setCmd) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	if len(args) > 3 {
		return ARITY_ERROR(SET_SIGNATURE)
	}
	return core.OK(core.STR(SET_SIGNATURE))
}

const GET_SIGNATURE = "get varname ?default?"

type getCmd struct{}

func (getCmd) Execute(args []core.Value, context any) core.Result {
	scope := context.(*Scope)
	switch len(args) {
	case 2:
		switch args[1].Type() {
		case core.ValueType_TUPLE,
			core.ValueType_QUALIFIED:
			return scope.ResolveValue(args[1])
		default:
			return scope.GetVariable(args[1], nil)
		}
	case 3:
		switch args[1].Type() {
		case core.ValueType_TUPLE:
			return core.ERROR("cannot use default with name tuples")
		case core.ValueType_QUALIFIED:
			{
				result := scope.ResolveValue(args[1])
				if result.Code == core.ResultCode_OK {
					return result
				}
				return core.OK(args[2])
			}
		default:
			return scope.GetVariable(args[1], args[2])
		}
	default:
		return ARITY_ERROR(GET_SIGNATURE)
	}
}
func (getCmd) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	if len(args) > 3 {
		return ARITY_ERROR(GET_SIGNATURE)
	}
	return core.OK(core.STR(GET_SIGNATURE))
}

const EXISTS_SIGNATURE = "exists varname"

type existsCmd struct{}

func (existsCmd) Execute(args []core.Value, context any) core.Result {
	scope := context.(*Scope)
	switch len(args) {
	case 2:
		switch args[1].Type() {
		case core.ValueType_TUPLE:
			return core.ERROR("invalid value")
		case core.ValueType_QUALIFIED:
			{
				return core.OK(core.BOOL(scope.ResolveValue(args[1]).Code == core.ResultCode_OK))
			}
		default:
			return core.OK(core.BOOL(scope.GetVariable(args[1], nil).Code == core.ResultCode_OK))
		}
	default:
		return ARITY_ERROR(EXISTS_SIGNATURE)
	}
}
func (existsCmd) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	if len(args) > 2 {
		return ARITY_ERROR(EXISTS_SIGNATURE)
	}
	return core.OK(core.STR(EXISTS_SIGNATURE))
}

const UNSET_SIGNATURE = "unset varname"

type unsetCmd struct{}

func (unsetCmd) Execute(args []core.Value, context any) core.Result {
	scope := context.(*Scope)
	switch len(args) {
	case 2:
		return unset(scope, args[1], false)
	default:
		return ARITY_ERROR(UNSET_SIGNATURE)
	}
}
func (unsetCmd) Help(args []core.Value, _ core.CommandHelpOptions, _ any) core.Result {
	if len(args) > 2 {
		return ARITY_ERROR(UNSET_SIGNATURE)
	}
	return core.OK(core.STR(UNSET_SIGNATURE))
}

func unset(scope *Scope, name core.Value, check bool) core.Result {
	if name.Type() != core.ValueType_TUPLE {
		return scope.UnsetVariable(name, check)
	}
	variables := name.(core.TupleValue)
	// First pass for error checking
	for i := 0; i < len(variables.Values); i++ {
		result := unset(scope, variables.Values[i], true)
		if result.Code != core.ResultCode_OK {
			return result
		}
	}
	if check {
		return core.OK(core.NIL)
	}
	// Second pass for actual setting
	for i := 0; i < len(variables.Values); i++ {
		unset(scope, variables.Values[i], false)
	}
	return core.OK(core.NIL)
}

func registerVariableCommands(scope *Scope) {
	scope.RegisterNamedCommand("let", letCmd{})
	scope.RegisterNamedCommand("set", setCmd{})
	scope.RegisterNamedCommand("get", getCmd{})
	scope.RegisterNamedCommand("exists", existsCmd{})
	scope.RegisterNamedCommand("unset", unsetCmd{})
}
