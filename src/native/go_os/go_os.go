package go_os

import (
	"helena/core"
	"os"
)

func asString(value core.Value) (s string, ok bool) {
	result, s := core.ValueToString(value)
	if result.Code == core.ResultCode_OK {
		return s, true
	}
	return "", false
}

type OsCmd struct{}

func (OsCmd) Execute(args []core.Value, _ any) core.Result {
	if len(args) < 2 {
		return core.ERROR(`wrong # args: should be "os method ?arg ...?"`)
	}
	method, ok := asString(args[1])
	if !ok {
		return core.ERROR("invalid method name")
	}
	// https://pkg.go.dev/os
	switch method {
	case "ReadFile":
		if len(args) < 3 {
			return core.ERROR(`wrong # args: should be "os ReadFile name"`)
		}
		name, _ := asString((args[2]))
		data, err := os.ReadFile(name)
		if err != nil {
			return core.ERROR(err.Error())
		}
		return core.OK(core.STR(string(data)))

	default:
		return core.ERROR("unsupported method " + method)
	}
}
