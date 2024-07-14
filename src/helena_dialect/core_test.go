package helena_dialect_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"helena/core"
	. "helena/helena_dialect"
)

var _ = Describe("Helena core internals", func() {
	var rootScope *Scope

	var tokenizer core.Tokenizer
	var parser *core.Parser

	parse := func(script string) *core.Script {
		return parser.ParseTokens(tokenizer.Tokenize(script), nil).Script
	}
	prepareScript := func(script string) *Process {
		return rootScope.PrepareProcess(rootScope.Compile(*parse(script)))
	}

	BeforeEach(func() {
		rootScope = NewRootScope(nil)
		InitCommands(rootScope)

		tokenizer = core.Tokenizer{}
		parser = core.NewParser(nil)
	})

	Describe("Process", func() {
		Specify("captureErrorStack", func() {
			source := `
macro cmd1 {} {cmd2}
macro cmd2 {} {error msg}
cmd1
`
			program := rootScope.Compile(*parse(source))
			process := NewProcess(rootScope, program, &ProcessOptions{
				CaptureErrorStack: true,
			})
			result := process.Run()
			Expect(result.Code).To(Equal(core.ResultCode_ERROR))
			Expect(result.Value).To(Equal(STR("msg")))
			errorStack := result.Data.(*core.ErrorStack)
			Expect(errorStack.Depth()).To(Equal(uint(3)))
			Expect(errorStack.Level(0)).To(Equal(core.ErrorStackLevel{
				Frame: &[]core.Value{STR("error"), STR("msg")},
			}))
			Expect(errorStack.Level(1)).To(Equal(core.ErrorStackLevel{Frame: &[]core.Value{STR("cmd2")}}))
			Expect(errorStack.Level(2)).To(Equal(core.ErrorStackLevel{Frame: &[]core.Value{STR("cmd1")}}))
		})
	})

	Describe("Scope", func() {
		Specify("captureErrorStack", func() {
			rootScope = NewRootScope(&ScopeOptions{
				CaptureErrorStack: true,
			})
			InitCommands(rootScope)

			source := `
macro cmd1 {} {cmd2}
macro cmd2 {} {error msg}
cmd1
`
			process := prepareScript(source)
			result := process.Run()
			Expect(result.Code).To(Equal(core.ResultCode_ERROR))
			Expect(result.Value).To(Equal(STR("msg")))
			errorStack := result.Data.(*core.ErrorStack)
			Expect(errorStack.Depth()).To(Equal(uint(3)))
			Expect(errorStack.Level(0)).To(Equal(core.ErrorStackLevel{
				Frame: &[]core.Value{STR("error"), STR("msg")},
			}))
			Expect(errorStack.Level(1)).To(Equal(core.ErrorStackLevel{Frame: &[]core.Value{STR("cmd2")}}))
			Expect(errorStack.Level(2)).To(Equal(core.ErrorStackLevel{Frame: &[]core.Value{STR("cmd1")}}))
		})
		Specify("captureErrorStack + capturePositions", func() {
			parser = core.NewParser(&core.ParserOptions{CapturePositions: true})
			rootScope = NewRootScope(&ScopeOptions{
				CapturePositions:  true,
				CaptureErrorStack: true,
			})
			InitCommands(rootScope)

			source := `
macro cmd1 {} {cmd2}
macro cmd2 {} {error msg}
cmd1
`
			process := prepareScript(source)
			result := process.Run()
			Expect(result.Code).To(Equal(core.ResultCode_ERROR))
			Expect(result.Value).To(Equal(STR("msg")))
			errorStack := result.Data.(*core.ErrorStack)
			Expect(errorStack.Depth()).To(Equal(uint(3)))
			Expect(errorStack.Level(0)).To(Equal(core.ErrorStackLevel{
				Frame:    &[]core.Value{STR("error"), STR("msg")},
				Position: &core.SourcePosition{Index: 37, Line: 2, Column: 15},
			}))
			Expect(errorStack.Level(1)).To(Equal(core.ErrorStackLevel{
				Frame:    &[]core.Value{STR("cmd2")},
				Position: &core.SourcePosition{Index: 16, Line: 1, Column: 15},
			}))
			Expect(errorStack.Level(2)).To(Equal(core.ErrorStackLevel{
				Frame:    &[]core.Value{STR("cmd1")},
				Position: &core.SourcePosition{Index: 48, Line: 3, Column: 0},
			}))
		})
		Describe("result error stack", func() {
			cmd := simpleCommand{
				execute: func(_ []core.Value, _ any) core.Result {
					errorStack := core.NewErrorStack()
					errorStack.Push(core.ErrorStackLevel{Frame: &[]core.Value{STR("foo")}})
					return core.ERROR_STACK("msg", errorStack)
				},
			}

			Specify("default options", func() {
				parser = core.NewParser(nil)
				rootScope = NewRootScope(nil)
				InitCommands(rootScope)
				rootScope.RegisterNamedCommand("cmd", cmd)
				source := `
macro mac {} {cmd}
mac
`
				process := prepareScript(source)
				result := process.Run()
				Expect(result.Code).To(Equal(core.ResultCode_ERROR))
				Expect(result.Value).To(Equal(STR("msg")))
				Expect(result.Data).To(BeNil())
			})
			Specify("captureErrorStacks", func() {
				parser = core.NewParser(nil)
				rootScope = NewRootScope(&ScopeOptions{
					CaptureErrorStack: true,
				})
				InitCommands(rootScope)
				rootScope.RegisterNamedCommand("cmd", cmd)
				source := `
macro mac {} {cmd}
mac
`
				process := prepareScript(source)
				result := process.Run()
				Expect(result.Code).To(Equal(core.ResultCode_ERROR))
				Expect(result.Value).To(Equal(STR("msg")))
				errorStack := result.Data.(*core.ErrorStack)
				Expect(errorStack.Depth()).To(Equal(uint(3)))
				Expect(errorStack.Level(0)).To(Equal(core.ErrorStackLevel{
					Frame: &[]core.Value{STR("foo")},
				}))
				Expect(errorStack.Level(1)).To(Equal(core.ErrorStackLevel{
					Frame: &[]core.Value{STR("cmd")},
				}))
				Expect(errorStack.Level(2)).To(Equal(core.ErrorStackLevel{
					Frame: &[]core.Value{STR("mac")},
				}))
			})
			Specify("captureErrorStacks + capturePositions", func() {
				parser = core.NewParser(&core.ParserOptions{CapturePositions: true})
				rootScope = NewRootScope(&ScopeOptions{
					CapturePositions:  true,
					CaptureErrorStack: true,
				})
				InitCommands(rootScope)
				rootScope.RegisterNamedCommand("cmd", cmd)
				source := `
macro mac {} {cmd}
mac
`
				process := prepareScript(source)
				result := process.Run()
				Expect(result.Code).To(Equal(core.ResultCode_ERROR))
				Expect(result.Value).To(Equal(STR("msg")))
				errorStack := result.Data.(*core.ErrorStack)
				Expect(errorStack.Depth()).To(Equal(uint(3)))
				Expect(errorStack.Level(0)).To(Equal(core.ErrorStackLevel{
					Frame: &[]core.Value{STR("foo")},
				}))
				Expect(errorStack.Level(1)).To(Equal(core.ErrorStackLevel{
					Frame:    &[]core.Value{STR("cmd")},
					Position: &core.SourcePosition{Index: 15, Line: 1, Column: 14},
				}))
				Expect(errorStack.Level(2)).To(Equal(core.ErrorStackLevel{
					Frame:    &[]core.Value{STR("mac")},
					Position: &core.SourcePosition{Index: 20, Line: 2, Column: 0},
				}))
			})
		})

	})
})
