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
var INT = core.INT
var REAL = core.REAL
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

func asString(value core.Value) (s string) { _, s = core.ValueToString(value); return }

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

type exampleSpec struct {
	script string
	result any
}
type exampleExecutor = func(spec exampleSpec) core.Result

func executeExample(executor exampleExecutor, spec exampleSpec) {
	result := executor(spec)
	switch r := spec.result.(type) {
	case nil:
		Expect(result.Code).To(Equal(core.ResultCode_OK))
	case core.Result:
		Expect(result).To(Equal(r))
	case core.Value:
		Expect(result).To(Equal(OK(r)))
	}
}
func specifyExample(executor exampleExecutor) func(any) {
	return func(specs any) {
		switch s := specs.(type) {
		case exampleSpec:
			executeExample(executor, s)
		case []exampleSpec:
			for _, spec := range s {
				executeExample(executor, spec)
			}
		}
	}
}
