package helena_dialect_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"helena/core"
)

func TestHelenaDialect(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Helena Dialect Suite")
}

var NIL = core.NIL
var TRUE = core.TRUE
var FALSE = core.FALSE
var STR = core.STR
var LIST = core.LIST
var DICT = core.DICT
var TUPLE = core.TUPLE

var OK = core.OK
var RETURN = core.RETURN
var YIELD = core.YIELD
var ERROR = core.ERROR
var BREAK = core.BREAK
var CONTINUE = core.CONTINUE

type simpleCommand struct {
	execute func(args []core.Value, context any) core.Result
}

func (command simpleCommand) Execute(args []core.Value, context any) core.Result {
	return command.execute(args, context)
}

type commandWithHelp struct {
	execute func(args []core.Value, context any) core.Result
	help    func(args []core.Value, options core.CommandHelpOptions, context any) core.Result
}

func (command commandWithHelp) Execute(args []core.Value, context any) core.Result {
	return command.execute(args, context)
}
func (command commandWithHelp) Help(args []core.Value, options core.CommandHelpOptions, context any) core.Result {
	return command.help(args, options, context)
}
