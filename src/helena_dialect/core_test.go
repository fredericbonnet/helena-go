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
		return parser.Parse(tokenizer.Tokenize(script)).Script
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
			Expect(process.Run()).To(Equal(ERROR("msg")))
			Expect(process.ErrorStack.Depth()).To(Equal(uint(3)))
			Expect(process.ErrorStack.Level(0)).To(Equal(ErrorStackLevel{
				Frame: []core.Value{STR("error"), STR("msg")},
			}))
			Expect(process.ErrorStack.Level(1)).To(Equal(ErrorStackLevel{Frame: []core.Value{STR("cmd2")}}))
			Expect(process.ErrorStack.Level(2)).To(Equal(ErrorStackLevel{Frame: []core.Value{STR("cmd1")}}))
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
			Expect(process.Run()).To(Equal(ERROR("msg")))
			Expect(process.ErrorStack.Depth()).To(Equal(uint(3)))
			Expect(process.ErrorStack.Level(0)).To(Equal(ErrorStackLevel{
				Frame: []core.Value{STR("error"), STR("msg")},
			}))
			Expect(process.ErrorStack.Level(1)).To(Equal(ErrorStackLevel{
				Frame: []core.Value{STR("cmd2")},
			}))
			Expect(process.ErrorStack.Level(2)).To(Equal(ErrorStackLevel{
				Frame: []core.Value{STR("cmd1")},
			}))
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
			Expect(process.Run()).To(Equal(ERROR("msg")))
			Expect(process.ErrorStack.Depth()).To(Equal(uint(3)))
			Expect(process.ErrorStack.Level(0)).To(Equal(ErrorStackLevel{
				Frame:    []core.Value{STR("error"), STR("msg")},
				Position: &core.SourcePosition{Index: 37, Line: 2, Column: 15},
			}))
			Expect(process.ErrorStack.Level(1)).To(Equal(ErrorStackLevel{
				Frame:    []core.Value{STR("cmd2")},
				Position: &core.SourcePosition{Index: 16, Line: 1, Column: 15},
			}))
			Expect(process.ErrorStack.Level(2)).To(Equal(ErrorStackLevel{
				Frame:    []core.Value{STR("cmd1")},
				Position: &core.SourcePosition{Index: 48, Line: 3, Column: 0},
			}))
		})
	})
})
