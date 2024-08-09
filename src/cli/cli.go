package cli

import (
	"fmt"
	"helena/core"
	"helena/helena_dialect"
	"helena/native"
	"helena/picol_dialect"
	"os"

	"github.com/ergochat/readline"
	"github.com/fatih/color"
)

var moduleRegistry = helena_dialect.NewModuleRegistry(&helena_dialect.ModuleOptions{
	CaptureErrorStack: true,
	CapturePositions:  true,
})

func sourceFile(path string, scope *helena_dialect.Scope) core.Result {
	data, err := os.ReadFile(path)
	if err != nil {
		return core.ERROR("error reading file: " + fmt.Sprint(err))
	}
	tokens := core.Tokenizer{}.Tokenize(string(data))
	result := core.NewParser(nil).ParseTokens(tokens, nil)
	if !result.Success {
		return core.ERROR(result.Message)
	}
	program := scope.Compile(*result.Script)
	process := scope.PrepareProcess(program)
	return process.Run()
}

type sourceCmd struct{}

func (sourceCmd) Execute(args []core.Value, context any) core.Result {
	scope := context.(*helena_dialect.Scope)
	if len(args) != 2 {
		return helena_dialect.ARITY_ERROR("source path")
	}
	_, path := core.ValueToString(args[1])
	data, err := os.ReadFile(path)
	if err != nil {
		return core.ERROR("error reading file: " + fmt.Sprint(err))
	}
	input := core.NewStringStreamFromFile(string(data), path)
	output := core.NewArrayTokenStream([]core.Token{}, input.Source())
	(&core.Tokenizer{}).TokenizeStream(input, output)
	result := core.NewParser(&core.ParserOptions{
		CapturePositions: true,
	}).Parse(output)
	if !result.Success {
		return core.ERROR(result.Message)
	}
	program := scope.Compile(*result.Script)
	return helena_dialect.CreateContinuationValue(scope, program, nil)
}

type exitCmd struct{}

func (exitCmd) Execute(args []core.Value, _ any) core.Result {
	if len(args) != 1 {
		return helena_dialect.ARITY_ERROR("exit")
	}
	os.Exit(0)
	panic("UNREACHABLE")
}

type picolCmd struct{}

func (picolCmd) Execute(args []core.Value, _ any) core.Result {
	if len(args) != 2 {
		return helena_dialect.ARITY_ERROR("picol script")
	}
	if args[1].Type() != core.ValueType_SCRIPT {
		return core.ERROR("invalid script")
	}
	script := args[1].(core.ScriptValue).Script
	scope := picol_dialect.NewPicolScope(nil)
	picol_dialect.InitPicolCommands(scope)
	return scope.Evaluator.EvaluateScript(script)
}

func initScope() *helena_dialect.Scope {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	rootScope := helena_dialect.NewRootScope(&helena_dialect.ScopeOptions{
		CaptureErrorStack: true,
		CapturePositions:  true,
	})
	helena_dialect.InitCommandsForModule(rootScope, moduleRegistry, cwd)
	rootScope.RegisterNamedCommand("source", sourceCmd{})
	rootScope.RegisterNamedCommand("exit", exitCmd{})
	rootScope.RegisterNamedCommand("picol", picolCmd{})
	return rootScope
}

func source(path string) {
	rootScope := initScope()
	result := sourceFile(path, rootScope)
	value, err := processResult(result)
	if err == nil {
		os.Stdout.WriteString(resultWriter(value) + "\n")
		os.Exit(0)
	} else {
		os.Stderr.WriteString(resultWriter(err) + "\n")
		os.Exit(-1)
	}
}

// function registerNativeModule(
//   moduleName: string,
//   exportName: string,
//   command: Command
// ) {
//   const scope = new Scope();
//   const exports = new Map();
//   scope.registerNamedCommand(exportName, command);
//   exports.set(exportName, STR(exportName));
//   moduleRegistry.register(moduleName, new Module(scope, exports));
// }

func prompt() {
	rl, err := readline.NewEx(&readline.Config{
		Prompt: "> ",
	})
	if err != nil {
		panic(err)
	}
	defer rl.Close()

	rootScope := initScope()
	registerNativeModule("go:slog", "slog", native.SlogCmd{})

	cmd := ""
	for {
		line, err := rl.Readline()
		if err != nil {
			break
		}
		cmd += line
		value, err := run(rootScope, cmd)
		if err != nil {
			if _, ok := err.(recoverableError); ok {
				rl.SetPrompt("... ")
				cmd += "\n"
				continue
			}
			os.Stdout.WriteString(resultWriter(err) + "\n")
		} else {
			os.Stdout.WriteString(resultWriter(value) + "\n")
		}
		rl.SetPrompt("> ")
		cmd = ""
	}

}

func run(scope *helena_dialect.Scope, cmd string) (core.Value, error) {
	input := core.NewStringStream(cmd)
	tokens := []core.Token{}
	output := core.NewArrayTokenStream(tokens, input.Source())
	(&core.Tokenizer{}).TokenizeStream(input, output)
	if len(tokens) > 0 &&
		tokens[len(tokens)-1].Type == core.TokenType_CONTINUATION {
		// Continuation, wait for next line
		return nil, recoverableError{"continuation"}
	}

	parser := core.NewParser(&core.ParserOptions{CapturePositions: true})
	parseResult := parser.ParseStream(output)
	if !parseResult.Success {
		// Parse error
		return nil, basicError{parseResult.Message}
	}

	parseResult = parser.CloseStream()
	if !parseResult.Success {
		// Incomplete script, wait for new line
		return nil, recoverableError{parseResult.Message}
	}

	program := scope.Compile(*parseResult.Script)
	process := scope.PrepareProcess(program)
	result := process.Run()
	if result.Code == core.ResultCode_ERROR {
		printErrorStack(result.Data.(*core.ErrorStack))
	}
	return processResult(result)
}

type basicError struct {
	message string
}

func (err basicError) Error() string {
	return err.message
}

type recoverableError struct {
	message string
}

func (err recoverableError) Error() string {
	return err.message
}
func printErrorStack(errorStack *core.ErrorStack) {
	for level := uint(0); level < errorStack.Depth(); level++ {
		l := errorStack.Level(level)
		log := fmt.Sprintf(`[%v] `, level)
		if l.Source != nil && l.Source.Filename != nil {
			log += *l.Source.Filename
		} else {
			log += "(script)"
		}
		if l.Position != nil {
			log += fmt.Sprintf(`:%v:%v: `, l.Position.Line+1, l.Position.Column+1)
		} else {
			log += ` `
		}
		if l.Frame != nil {
			for i, arg := range *l.Frame {
				if i > 0 {
					log += " "
				}
				log += displayErrorFrameArg(arg)
			}
		}
		os.Stdout.WriteString(grey.Sprintln(log))
	}
}
func displayErrorFrameArg(displayable any) string {
	if _, ok := displayable.(core.ListValue); ok {
		return `[list (...)]`
	}
	if _, ok := displayable.(core.DictionaryValue); ok {
		return `[dict (...)]`
	}
	if _, ok := displayable.(core.ScriptValue); ok {
		return `{...}`
	}
	return core.Display(displayable, displayErrorFrameArg)
}

func processResult(result core.Result) (core.Value, error) {
	switch result.Code {
	case core.ResultCode_OK:
		return result.Value, nil
	case core.ResultCode_ERROR:
		_, s := core.ValueToString(result.Value)
		return nil, basicError{s}
	default:
		return nil, (basicError{"unexpected " + core.RESULT_CODE_NAME(result)})
	}
}

var grey = color.New(color.FgBlack)
var italicGrey = color.New(color.FgBlack).Add(color.Italic)

func displayResult(displayable any) string {
	if v, ok := displayable.(core.ListValue); ok {
		return helena_dialect.DisplayListValue(v, displayResult)
	}
	if v, ok := displayable.(core.DictionaryValue); ok {
		return helena_dialect.DisplayDictionaryValue(v, displayResult)
	}
	return core.DefaultDisplayFunction(displayable)
}
func resultWriter(output any) string {
	if err, ok := output.(error); ok {
		return color.RedString(err.Error())
	}
	value := core.Display(output, displayResult)
	var type_ string
	if v, ok := output.(core.Value); ok {
		if v.Type() == core.ValueType_CUSTOM {
			type_ = `CUSTOM[` + output.(core.CustomValue).CustomType().Name + `]`
		} else {
			type_ = fmt.Sprint(v.Type())
		}
	}
	if len(type_) > 0 {
		return color.GreenString(value) + italicGrey.Sprint(" # "+type_)
	} else {
		return color.GreenString(value)
	}
}

func registerNativeModule(
	moduleName string,
	exportName string,
	command core.Command,
) {
	scope := helena_dialect.NewRootScope(nil)
	exports := &helena_dialect.Exports{}
	scope.RegisterNamedCommand(exportName, command)
	(*exports)[exportName] = core.STR(exportName)
	moduleRegistry.Register(moduleName, helena_dialect.NewModule(scope, exports))

}

func Cli() {
	if len(os.Args) > 2 {
		os.Stderr.WriteString("Usage: helena [script]\n")
		os.Exit(0)
	} else if len(os.Args) == 2 {
		source(os.Args[1])
	} else {
		prompt()
	}
}
