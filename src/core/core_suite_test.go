package core_test

import (
	"strconv"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "helena/core"
)

func TestCore(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Core Suite")
}

//
// Helpers
//

func asString(value Value) string { return ValueToString(value).Data }

type mockVariableResolver struct {
	variables map[string]Value
}

func newMockVariableResolver() *mockVariableResolver {
	return &mockVariableResolver{
		variables: map[string]Value{},
	}
}
func (resolver *mockVariableResolver) Resolve(name string) Value {
	return resolver.variables[name]
}
func (resolver *mockVariableResolver) register(name string, value Value) {
	resolver.variables[name] = value
}

type intCommand struct {
}

func (command intCommand) Execute(args []Value, context any) Result {
	return OK(args[0])
}

type mockCommandResolver struct {
	commands map[string]Command
}

func newMockCommandResolver() *mockCommandResolver {
	return &mockCommandResolver{
		commands: map[string]Command{},
	}
}

func (resolver *mockCommandResolver) Resolve(name Value) Command {
	if name.Type() == ValueType_INTEGER {
		return intCommand{}
	}
	if _, err := strconv.ParseInt(asString(name), 10, 64); err == nil {
		return intCommand{}
	}
	return resolver.commands[asString(name)]
}
func (resolver *mockCommandResolver) register(name string, command Command) {
	resolver.commands[name] = command
}

type functionCommand struct {
	fn func(args []Value) Value
}

func (command functionCommand) Execute(args []Value, context any) Result {
	return OK(command.fn(args))
}

type simpleCommand struct {
	fn func(args []Value) Result
}

func (command simpleCommand) Execute(args []Value, context any) Result {
	return command.fn(args)
}

type resumableCommand struct {
	exec   func(args []Value, context any) Result
	resume func(result Result, context any) Result
}

func (command resumableCommand) Execute(args []Value, context any) Result {
	return command.exec(args, context)
}
func (command resumableCommand) Resume(result Result, context any) Result {
	return command.resume(result, context)
}

type captureContextCommand struct {
	context any
}

func (command *captureContextCommand) Execute(args []Value, context any) Result {
	command.context = context
	return OK(NIL)
}

type builderFn func(rules []Value) TypedResult[Selector]
type mockSelectorResolver struct {
	builder builderFn
}

func newMockSelectorResolver() *mockSelectorResolver {
	return &mockSelectorResolver{
		builder: func(rules []Value) TypedResult[Selector] { return OK_T[Selector](NIL, nil) },
	}
}

func (resolver *mockSelectorResolver) Resolve(rules []Value) TypedResult[Selector] {
	return resolver.builder(rules)
}
func (resolver *mockSelectorResolver) register(builder builderFn) {
	resolver.builder = builder
}

type lastSelector struct{}

func (selector lastSelector) Apply(value Value) Result {
	if selectable, ok := value.(Selectable); ok {
		return selectable.Select(selector)
	}
	list, ok := value.(ListValue)
	if !ok {
		return ERROR("value is not a list")
	}
	return OK(list.Values[len(list.Values)-1])
}
