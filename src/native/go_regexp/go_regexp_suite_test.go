package go_regexp_test

import (
	"helena/core"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGoRegexp(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Go Regexp Suite")
}

//
// Helpers
//

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
var ERROR = core.ERROR

//
// Helpers
//

func asString(value core.Value) (s string) { _, s = core.ValueToString(value); return }

type mockVariableResolver struct {
	variables map[string]core.Value
}

func newMockVariableResolver() *mockVariableResolver {
	return &mockVariableResolver{
		variables: map[string]core.Value{},
	}
}
func (resolver *mockVariableResolver) Resolve(name string) core.Value {
	return resolver.variables[name]
}
func (resolver *mockVariableResolver) register(name string, value core.Value) {
	resolver.variables[name] = value
}

type mockCommandResolver struct {
	commands map[string]core.Command
}

func newMockCommandResolver() *mockCommandResolver {
	return &mockCommandResolver{
		commands: map[string]core.Command{},
	}
}

func (resolver *mockCommandResolver) Resolve(name core.Value) core.Command {
	return resolver.commands[asString(name)]
}
func (resolver *mockCommandResolver) register(name string, command core.Command) {
	resolver.commands[name] = command
}
