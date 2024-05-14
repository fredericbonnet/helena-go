package helena_dialect

import (
	"helena/core"
	"strings"
)

func ARITY_ERROR(signature string) core.Result {
	return core.ERROR(`wrong # args: should be "` + signature + `"`)
}

type ArgumentType uint8

const (
	ArgumentType_REQUIRED ArgumentType = iota
	ArgumentType_OPTIONAL
	ArgumentType_REMAINDER
)

type Argument struct {
	Name    string
	Type    ArgumentType
	Default core.Value
	Guard   core.Value
	Option  *Option
}

type OptionType uint8

const (
	OptionType_FLAG OptionType = iota
	OptionType_OPTION
)

type Option struct {
	Names []string
	Type  OptionType
}

func OptionName(names []string) string {
	return strings.Join(names, "|")
}

func buildArguments(specs core.Value) core.TypedResult[[]Argument] {
	args := []Argument{}
	argnames := map[string]struct{}{}
	optnames := map[string]struct{}{}
	hasRemainder := false
	result := ValueToArray(specs)
	if result.Code != core.ResultCode_OK {
		return core.ERROR_T[[]Argument]("invalid argument list")
	}
	values := result.Data
	var lastOption *Option = nil
	for _, value := range values {
		result := isOption(value)
		if result.Code != core.ResultCode_OK {
			return core.ResultAs[[]Argument](result.AsResult())
		}
		option := result.Data
		if option != nil {
			if hasRemainder {
				return core.ERROR_T[[]Argument]("cannot use remainder argument before options")
			}
			for _, optname := range option.Names {
				if _, ok := optnames[optname]; ok {
					return core.ERROR_T[[]Argument](`duplicate option "` + optname + `"`)
				}
				optnames[optname] = struct{}{}
			}
			lastOption = option
			continue
		}
		result2 := buildArgument(value)
		if result2.Code != core.ResultCode_OK {
			return core.ResultAs[[]Argument](result2.AsResult())
		}
		arg := result2.Data
		if lastOption != nil {
			if lastOption.Type == OptionType_FLAG && arg.Type != ArgumentType_OPTIONAL {
				return core.ERROR_T[[]Argument](
					`argument for flag "` + OptionName(lastOption.Names) + `" must be optional`,
				)
			}
			arg.Option = lastOption
			args = append(args, arg)
			lastOption = nil
			continue
		}

		if arg.Type == ArgumentType_REMAINDER && hasRemainder {
			return core.ERROR_T[[]Argument]("only one remainder argument is allowed")
		}
		if _, ok := argnames[arg.Name]; ok {
			return core.ERROR_T[[]Argument](`duplicate argument "` + arg.Name + `"`)
		}
		hasRemainder = arg.Type == ArgumentType_REMAINDER
		argnames[arg.Name] = struct{}{}
		args = append(args, arg)
	}
	if lastOption != nil {
		return core.ERROR_T[[]Argument](`missing argument for option "` + OptionName(lastOption.Names) + `"`)
	}
	return core.OK_T(core.NIL, args)
}
func isOption(value core.Value) core.TypedResult[*Option] {
	var options []core.Value
	switch value.Type() {
	case core.ValueType_LIST,
		core.ValueType_TUPLE,
		core.ValueType_SCRIPT:
		{
			result := ValueToArray(value)
			if result.Code != core.ResultCode_OK {
				return core.ResultAs[*Option](result.AsResult())
			}
			options = result.Data
		}
	default:
		options = []core.Value{value}
	}
	if len(options) == 0 {
		return core.OK_T[*Option](core.NIL, nil)
	}

	var type_ OptionType
	names := []string{}
	for _, option := range options {
		result := core.ValueToString(option)
		if result.Code != core.ResultCode_OK {
			break
		}
		name := result.Data
		if len(name) < 1 {
			break
		}
		if name[0] == '-' {
			// Option
			if len(name) < 2 {
				break
			}
			if name == "--" {
				return core.ERROR_T[*Option]("cannot use option terminator as option name")
			}
			if len(names) > 0 && type_ != OptionType_OPTION {
				break
			}
			type_ = OptionType_OPTION
			names = append(names, name)
		} else if name[0] == '?' {
			// Flag
			if len(name) < 3 {
				break
			}
			if name[1] != '-' {
				break
			}
			if name == "?--" {
				return core.ERROR_T[*Option]("cannot use option terminator as option name")
			}
			if len(names) > 0 && type_ != OptionType_FLAG {
				break
			}
			type_ = OptionType_FLAG
			names = append(names, name[1:])
		} else {
			break
		}
	}
	if len(names) == 0 {
		return core.OK_T[*Option](core.NIL, nil)
	}
	if len(names) != len(options) {
		return core.ERROR_T[*Option](`incompatible aliases for option "` + OptionName(names) + `"`)
	}
	return core.OK_T(core.NIL, &Option{names, type_})
}
func buildArgument(value core.Value) core.TypedResult[Argument] {
	switch value.Type() {
	case core.ValueType_LIST,
		core.ValueType_TUPLE,
		core.ValueType_SCRIPT:
		{
			result := ValueToArray(value)
			if result.Code != core.ResultCode_OK {
				return core.ResultAs[Argument](result.AsResult())
			}
			specs := result.Data
			switch len(specs) {
			case 0:
				return core.ERROR_T[Argument]("empty argument specifier")
			case 1:
				{
					result := core.ValueToString(specs[0])
					if result.Code != core.ResultCode_OK {
						return core.ERROR_T[Argument]("invalid argument name")
					}
					name := result.Data
					if name == "" || name == "?" {
						return core.ERROR_T[Argument]("empty argument name")
					}
					if name[0] == '?' {
						return core.OK_T(core.NIL, Argument{Name: name[1:], Type: ArgumentType_OPTIONAL})
					} else {
						return core.OK_T(core.NIL, Argument{Name: name, Type: ArgumentType_REQUIRED})
					}
				}
			case 2:
				{
					result1 := core.ValueToString(specs[0])
					result2 := core.ValueToString(specs[1])
					if result1.Code != core.ResultCode_OK && result2.Code != core.ResultCode_OK {
						return core.ERROR_T[Argument]("invalid argument name")
					}
					nameOrGuard := result1.Data
					nameOrDefault := result2.Data
					if (nameOrGuard == "" || nameOrGuard == "?") &&
						(nameOrDefault == "" || nameOrDefault == "?") {
						return core.ERROR_T[Argument]("empty argument name")
					}
					if result1.Code == core.ResultCode_OK && nameOrGuard[0] == '?' {
						return core.OK_T(core.NIL, Argument{
							Name:    nameOrGuard[1:],
							Type:    ArgumentType_OPTIONAL,
							Default: specs[1],
						})
					} else if nameOrDefault[0] == '?' {
						return core.OK_T(core.NIL, Argument{
							Name:  nameOrDefault[1:],
							Type:  ArgumentType_OPTIONAL,
							Guard: specs[0],
						})
					} else {
						return core.OK_T[Argument](core.NIL, Argument{
							Name:  nameOrDefault,
							Type:  ArgumentType_REQUIRED,
							Guard: specs[0],
						})
					}
				}
			case 3:
				{
					result := core.ValueToString(specs[1])
					if result.Code != core.ResultCode_OK {
						return core.ERROR_T[Argument]("invalid argument name")
					}
					name := result.Data
					if name == "" || name == "?" {
						return core.ERROR_T[Argument]("empty argument name")
					}
					if name[0] != '?' {
						return core.ERROR_T[Argument](`default argument "` + name + `" must be optional`)
					}
					return core.OK_T(core.NIL, Argument{
						Name:    name[1:],
						Type:    ArgumentType_OPTIONAL,
						Default: specs[2],
						Guard:   specs[0],
					})
				}
			default:
				{
					result := core.ValueToString(specs[0])
					if result.Code != core.ResultCode_OK {
						return core.ERROR_T[Argument]("invalid argument name")
					}
					name := result.Data
					return core.ERROR_T[Argument](`too many specifiers for argument "` + name + `"`)
				}
			}
		}
	default:
		{
			result := core.ValueToString(value)
			if result.Code != core.ResultCode_OK {
				return core.ERROR_T[Argument]("invalid argument name")
			}
			name := result.Data
			if name == "" || name == "?" {
				return core.ERROR_T[Argument]("empty argument name")
			}
			if name[0] == '*' {
				if len(name) == 1 {
					return core.OK_T(core.NIL, Argument{Name: name, Type: ArgumentType_REMAINDER})
				} else {
					return core.OK_T(core.NIL, Argument{Name: name[1:], Type: ArgumentType_REMAINDER})
				}
			} else if name[0] == '?' {
				return core.OK_T(core.NIL, Argument{Name: name[1:], Type: ArgumentType_OPTIONAL})
			} else {
				return core.OK_T(core.NIL, Argument{Name: name, Type: ArgumentType_REQUIRED})
			}
		}
	}
}

func BuildUsage(args []Argument, skip uint) string {
	result := ""
	for i, arg := range args[skip:] {
		if i != 0 {
			result += " "
		}
		if arg.Option != nil {
			name := OptionName(arg.Option.Names)
			switch arg.Option.Type {
			case OptionType_FLAG:
				result += `?` + name + `?`
			case OptionType_OPTION:
				switch arg.Type {
				case ArgumentType_REQUIRED:
					result += name + " " + arg.Name
				case ArgumentType_OPTIONAL:
					result += `?` + name + " " + arg.Name + `?`
				default:
					panic("CANTHAPPEN")
				}
			}
		} else {
			switch arg.Type {
			case ArgumentType_REQUIRED:
				result += arg.Name
			case ArgumentType_OPTIONAL:
				result += `?` + arg.Name + `?`
			case ArgumentType_REMAINDER:
				if arg.Name == "*" {
					result += `?arg ...?`
				} else {
					result += `?` + arg.Name + ` ...?`
				}
			}
		}
	}
	return result
}
