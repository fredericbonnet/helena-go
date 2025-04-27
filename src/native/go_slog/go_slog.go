package go_slog

import (
	"helena/core"
	"log/slog"
)

func asString(value core.Value) (s string, ok bool) {
	result, s := core.ValueToString(value)
	if result.Code == core.ResultCode_OK {
		return s, true
	}
	return "", false
}

type SlogCmd struct{}

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
		msg, _ := asString((args[2]))
		slog.Debug(msg)
	case "Error":
		if len(args) < 3 {
			return core.ERROR(`wrong # args: should be "slog Error msg"`)
		}
		msg, _ := asString((args[2]))
		slog.Error(msg, "toto", 1)
	case "Info":
		if len(args) < 3 {
			return core.ERROR(`wrong # args: should be "slog Info msg"`)
		}
		msg, _ := asString((args[2]))
		slog.Info(msg)
	case "Warn":
		if len(args) < 3 {
			return core.ERROR(`wrong # args: should be "slog Warn msg"`)
		}
		msg, _ := asString((args[2]))
		slog.Warn(msg)
	}
	return core.OK(core.NIL)
}
