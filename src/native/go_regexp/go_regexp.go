package go_regexp

import (
	"helena/core"
	"regexp"
)

func asString(value core.Value) (s string, ok bool) {
	result, s := core.ValueToString(value)
	if result.Code == core.ResultCode_OK {
		return s, true
	}
	return "", false
}

type RegexpValue struct {
	regexp *regexp.Regexp
}

var RegexpValueType core.CustomValueType = core.CustomValueType{Name: "go:Regexp"}

func NewRegexpValue(regexp *regexp.Regexp) RegexpValue {
	return RegexpValue{regexp: regexp}
}

func (RegexpValue) Type() core.ValueType {
	return core.ValueType_CUSTOM
}

func (RegexpValue) CustomType() core.CustomValueType {
	return RegexpValueType
}

func (value RegexpValue) Display(fn core.DisplayFunction) string {
	if fn != nil {
		return fn(value)
	}
	return core.UndisplayableValueWithLabel("Regexp " + value.regexp.String())
}

type RegexpCmd struct{}

func (RegexpCmd) Execute(args []core.Value, _ any) core.Result {
	if len(args) < 2 {
		return core.ERROR(`wrong # args: should be "regexp method ?arg ...?"`)
	}
	method, ok := asString(args[1])
	if !ok {
		return core.ERROR("invalid method name")
	}

	// https://pkg.go.dev/regexp
	switch method {
	case "QuoteMeta":
		// https://pkg.go.dev/regexp#QuoteMeta
		if len(args) != 3 {
			return core.ERROR(`wrong # args: should be "regexp QuoteMeta s"`)
		}
		result, s := core.ValueToString(args[2])
		if result.Code != core.ResultCode_OK {
			return result
		}
		return core.OK(core.STR(regexp.QuoteMeta(s)))

	case "Compile":
		// https://pkg.go.dev/regexp#Compile
		if len(args) != 3 {
			return core.ERROR(`wrong # args: should be "regexp Compile expr"`)
		}
		result, expr := core.ValueToString(args[2])
		if result.Code != core.ResultCode_OK {
			return result
		}
		re, err := regexp.Compile(expr)
		if err != nil {
			return core.ERROR(err.Error())
		}
		return core.OK(NewRegexpValue(re))

	case "CompilePOSIX":
		// https://pkg.go.dev/regexp#CompilePOSIX
		if len(args) != 3 {
			return core.ERROR(`wrong # args: should be "regexp CompilePOSIX expr"`)
		}
		result, expr := core.ValueToString(args[2])
		if result.Code != core.ResultCode_OK {
			return result
		}
		re, err := regexp.CompilePOSIX(expr)
		if err != nil {
			return core.ERROR(err.Error())
		}
		return core.OK(NewRegexpValue(re))

	case "FindAllString":
		// https://pkg.go.dev/regexp#Regexp.FindAllString
		if len(args) != 5 {
			return core.ERROR(`wrong # args: should be "regexp FindAllString re s n"`)
		}
		re, ok := args[2].(RegexpValue)
		if !ok {
			return core.ERROR("invalid regexp value")
		}
		result, s := core.ValueToString(args[3])
		if result.Code != core.ResultCode_OK {
			return result
		}
		result, n := core.ValueToInteger(args[4])
		if result.Code != core.ResultCode_OK {
			return result
		}
		matches := re.regexp.FindAllString(s, int(n))
		if matches == nil {
			return core.OK(core.NIL)
		}
		values := make([]core.Value, len(matches))
		for i, m := range matches {
			values[i] = core.STR(m)
		}
		return core.OK(core.LIST(values))

	case "FindAllStringIndex":
		// https://pkg.go.dev/regexp#Regexp.FindAllStringIndex
		if len(args) != 5 {
			return core.ERROR(`wrong # args: should be "regexp FindAllStringIndex re s n"`)
		}
		re, ok := args[2].(RegexpValue)
		if !ok {
			return core.ERROR("invalid regexp value")
		}
		result, s := core.ValueToString(args[3])
		if result.Code != core.ResultCode_OK {
			return result
		}
		result, n := core.ValueToInteger(args[4])
		if result.Code != core.ResultCode_OK {
			return result
		}
		matches := re.regexp.FindAllStringIndex(s, int(n))
		if matches == nil {
			return core.OK(core.NIL)
		}
		values := make([]core.Value, len(matches))
		for i, m := range matches {
			values[i] = core.LIST([]core.Value{
				core.INT(int64(m[0])),
				core.INT(int64(m[1])),
			})
		}
		return core.OK(core.LIST(values))

	case "FindAllStringSubmatch":
		// https://pkg.go.dev/regexp#Regexp.FindAllStringSubmatch
		if len(args) != 5 {
			return core.ERROR(`wrong # args: should be "regexp FindAllStringSubmatch re s n"`)
		}
		re, ok := args[2].(RegexpValue)
		if !ok {
			return core.ERROR("invalid regexp value")
		}
		result, s := core.ValueToString(args[3])
		if result.Code != core.ResultCode_OK {
			return result
		}
		result, n := core.ValueToInteger(args[4])
		if result.Code != core.ResultCode_OK {
			return result
		}
		matches := re.regexp.FindAllStringSubmatch(s, int(n))
		if matches == nil {
			return core.OK(core.NIL)
		}
		values := make([]core.Value, len(matches))
		for i, m := range matches {
			submatches := make([]core.Value, len(m))
			for j, sm := range m {
				submatches[j] = core.STR(sm)
			}
			values[i] = core.LIST(submatches)
		}
		return core.OK(core.LIST(values))

	case "FindAllStringSubmatchIndex":
		// https://pkg.go.dev/regexp#Regexp.FindAllStringSubmatchIndex
		if len(args) != 5 {
			return core.ERROR(`wrong # args: should be "regexp FindAllStringSubmatchIndex re s n"`)
		}
		re, ok := args[2].(RegexpValue)
		if !ok {
			return core.ERROR("invalid regexp value")
		}
		result, s := core.ValueToString(args[3])
		if result.Code != core.ResultCode_OK {
			return result
		}
		result, n := core.ValueToInteger(args[4])
		if result.Code != core.ResultCode_OK {
			return result
		}
		matches := re.regexp.FindAllStringSubmatchIndex(s, int(n))
		if matches == nil {
			return core.OK(core.NIL)
		}
		values := make([]core.Value, len(matches))
		for i, m := range matches {
			submatches := make([]core.Value, len(m))
			for j, sm := range m {
				submatches[j] = core.INT(int64(sm))
			}
			values[i] = core.LIST(submatches)
		}
		return core.OK(core.LIST(values))

	case "FindString":
		// https://pkg.go.dev/regexp#Regexp.FindString
		if len(args) != 4 {
			return core.ERROR(`wrong # args: should be "regexp FindString re s"`)
		}
		re, ok := args[2].(RegexpValue)
		if !ok {
			return core.ERROR("invalid regexp value")
		}
		result, s := core.ValueToString(args[3])
		if result.Code != core.ResultCode_OK {
			return result
		}
		match := re.regexp.FindString(s)
		return core.OK(core.STR(match))

	case "FindStringIndex":
		// https://pkg.go.dev/regexp#Regexp.FindStringIndex
		if len(args) != 4 {
			return core.ERROR(`wrong # args: should be "regexp FindStringIndex re s"`)
		}
		re, ok := args[2].(RegexpValue)
		if !ok {
			return core.ERROR("invalid regexp value")
		}
		result, s := core.ValueToString(args[3])
		if result.Code != core.ResultCode_OK {
			return result
		}
		match := re.regexp.FindStringIndex(s)
		if match == nil {
			return core.OK(core.NIL)
		}
		return core.OK(core.LIST([]core.Value{
			core.INT(int64(match[0])),
			core.INT(int64(match[1])),
		}))

	case "FindStringSubmatch":
		if len(args) != 4 {
			return core.ERROR(`wrong # args: should be "regexp FindStringSubmatch re s"`)
		}
		re, ok := args[2].(RegexpValue)
		if !ok {
			return core.ERROR("invalid regexp value")
		}
		result, s := core.ValueToString(args[3])
		if result.Code != core.ResultCode_OK {
			return result
		}
		match := re.regexp.FindStringSubmatch(s)
		if match == nil {
			return core.OK(core.NIL)
		}
		values := make([]core.Value, len(match))
		for i, m := range match {
			values[i] = core.STR(m)
		}
		return core.OK(core.LIST(values))

	case "FindStringSubmatchIndex":
		if len(args) != 4 {
			return core.ERROR(`wrong # args: should be "regexp FindStringSubmatchIndex re s"`)
		}
		re, ok := args[2].(RegexpValue)
		if !ok {
			return core.ERROR("invalid regexp value")
		}
		result, s := core.ValueToString(args[3])
		if result.Code != core.ResultCode_OK {
			return result
		}
		match := re.regexp.FindStringSubmatchIndex(s)
		if match == nil {
			return core.OK(core.NIL)
		}
		values := make([]core.Value, len(match))
		for i, m := range match {
			values[i] = core.INT(int64(m))
		}
		return core.OK(core.LIST(values))

	case "LiteralPrefix":
		// https://pkg.go.dev/regexp#Regexp.LiteralPrefix
		if len(args) != 3 {
			return core.ERROR(`wrong # args: should be "regexp LiteralPrefix re"`)
		}
		re, ok := args[2].(RegexpValue)
		if !ok {
			return core.ERROR("invalid regexp value")
		}
		prefix, complete := re.regexp.LiteralPrefix()
		return core.OK(core.TUPLE([]core.Value{core.STR(prefix), core.BOOL(complete)}))

	case "Longest":
		// https://pkg.go.dev/regexp#Regexp.Longest
		if len(args) != 3 {
			return core.ERROR(`wrong # args: should be "regexp Longest re"`)
		}
		re, ok := args[2].(RegexpValue)
		if !ok {
			return core.ERROR("invalid regexp value")
		}
		re.regexp.Longest()
		return core.OK(core.NIL)

	case "MatchString":
		// https://pkg.go.dev/regexp#Regexp.MatchString
		if len(args) != 4 {
			return core.ERROR(`wrong # args: should be "regexp MatchString re s"`)
		}
		re, ok := args[2].(RegexpValue)
		if !ok {
			return core.ERROR("invalid regexp value")
		}
		result, s := core.ValueToString(args[3])
		if result.Code != core.ResultCode_OK {
			return result
		}
		match := re.regexp.MatchString(s)
		return core.OK(core.BOOL(match))

	case "NumSubexp":
		// https://pkg.go.dev/regexp#Regexp.NumSubexp
		if len(args) != 3 {
			return core.ERROR(`wrong # args: should be "regexp NumSubexp re"`)
		}
		re, ok := args[2].(RegexpValue)
		if !ok {
			return core.ERROR("invalid regexp value")
		}
		nb := re.regexp.NumSubexp()
		return core.OK(core.INT(int64(nb)))

	case "ReplaceAllLiteralString":
		// https://pkg.go.dev/regexp#Regexp.ReplaceAllLiteralString
		if len(args) != 5 {
			return core.ERROR(`wrong # args: should be "regexp ReplaceAllLiteralString re src repl"`)
		}
		re, ok := args[2].(RegexpValue)
		if !ok {
			return core.ERROR("invalid regexp value")
		}
		result, src := core.ValueToString(args[3])
		if result.Code != core.ResultCode_OK {
			return result
		}
		result, repl := core.ValueToString(args[4])
		if result.Code != core.ResultCode_OK {
			return result
		}
		s := re.regexp.ReplaceAllLiteralString(src, repl)
		return core.OK(core.STR(s))

	case "ReplaceAllString":
		// https://pkg.go.dev/regexp#Regexp.ReplaceAllString
		if len(args) != 5 {
			return core.ERROR(`wrong # args: should be "regexp ReplaceAllString re src repl"`)
		}
		re, ok := args[2].(RegexpValue)
		if !ok {
			return core.ERROR("invalid regexp value")
		}
		result, src := core.ValueToString(args[3])
		if result.Code != core.ResultCode_OK {
			return result
		}
		result, repl := core.ValueToString(args[4])
		if result.Code != core.ResultCode_OK {
			return result
		}
		s := re.regexp.ReplaceAllString(src, repl)
		return core.OK(core.STR(s))

	case "ReplaceAllStringFunc":
		// https://pkg.go.dev/regexp#Regexp.ReplaceAllStringFunc
		return core.ERROR("not implemented")

	case "Split":
		// https://pkg.go.dev/regexp#Regexp.Split
		if len(args) != 5 {
			return core.ERROR(`wrong # args: should be "regexp Split re s n"`)
		}
		re, ok := args[2].(RegexpValue)
		if !ok {
			return core.ERROR("invalid regexp value")
		}
		result, s := core.ValueToString(args[3])
		if result.Code != core.ResultCode_OK {
			return result
		}
		result, n := core.ValueToInteger(args[4])
		if result.Code != core.ResultCode_OK {
			return result
		}
		strings := re.regexp.Split(s, int(n))
		values := make([]core.Value, len(strings))
		for i, m := range strings {
			values[i] = core.STR(m)
		}
		return core.OK(core.LIST(values))

	case "String":
		// https://pkg.go.dev/regexp#Regexp.String
		if len(args) != 3 {
			return core.ERROR(`wrong # args: should be "regexp String re"`)
		}
		re, ok := args[2].(RegexpValue)
		if !ok {
			return core.ERROR("invalid regexp value")
		}
		return core.OK(core.STR(re.regexp.String()))

	case "SubexpIndex":
		// https://pkg.go.dev/regexp#Regexp.SubexpIndex
		if len(args) != 4 {
			return core.ERROR(`wrong # args: should be "regexp SubexpIndex re s"`)
		}
		re, ok := args[2].(RegexpValue)
		if !ok {
			return core.ERROR("invalid regexp value")
		}
		result, s := core.ValueToString(args[3])
		if result.Code != core.ResultCode_OK {
			return result
		}
		index := re.regexp.SubexpIndex(s)
		return core.OK(core.INT(int64(index)))

	case "SubexpNames":
		// https://pkg.go.dev/regexp#Regexp.SubexpNames
		if len(args) != 3 {
			return core.ERROR(`wrong # args: should be "regexp SubexpNames re"`)
		}
		re, ok := args[2].(RegexpValue)
		if !ok {
			return core.ERROR("invalid regexp value")
		}
		names := re.regexp.SubexpNames()
		values := make([]core.Value, len(names))
		for i, m := range names {
			values[i] = core.STR(m)
		}
		return core.OK(core.LIST(values))

	default:
		return core.ERROR(`unknown method "` + method + `"`)
	}
}
