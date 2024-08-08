package native

import (
	"helena/core"
	"log/slog"
)

type SlogCmd struct{}

func asString(value core.Value) (s string, ok bool) {
	result, s := core.ValueToString(value)
	if result.Code == core.ResultCode_OK {
		return s, true
	}
	return "", false
}

func (SlogCmd) Execute(args []core.Value, _ any) core.Result {
	if len(args) < 2 {
		return core.ERROR(`wrong # args: should be "slog method ?arg ...?"`)
	}
	method, ok := asString(args[1])
	if !ok {
		return core.ERROR("invalid method name")
	}
	// TODO Attrs and Values https://pkg.go.dev/log/slog@go1.22.2#hdr-Attrs_and_Values
	switch method {
	case "Debug":
		if len(args) < 3 {
			return core.ERROR(`wrong # args: should be "slog Debug msg"`)
		}
		method, _ = asString((args[2]))
		slog.Debug(method)
	case "Error":
		if len(args) < 3 {
			return core.ERROR(`wrong # args: should be "slog Error msg"`)
		}
		method, _ = asString((args[2]))
		slog.Error(method, "toto", 1)
	case "Info":
		if len(args) < 3 {
			return core.ERROR(`wrong # args: should be "slog Info msg"`)
		}
		method, _ = asString((args[2]))
		slog.Info(method)
	case "Warn":
		if len(args) < 3 {
			return core.ERROR(`wrong # args: should be "slog Warn msg"`)
		}
		method, _ = asString((args[2]))
		slog.Warn(method)
	}
	return core.OK(core.NIL)
}
