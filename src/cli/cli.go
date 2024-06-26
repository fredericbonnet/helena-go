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

var moduleRegistry = helena_dialect.NewModuleRegistry()

func sourceFile(path string, scope *helena_dialect.Scope) core.Result {
	data, err := os.ReadFile(path)
	if err != nil {
		return core.ERROR("error reading file: " + fmt.Sprint(err))
	}
	tokens := core.Tokenizer{}.Tokenize(string(data))
	result := (&core.Parser{}).Parse(tokens)
	if !result.Success {
		return core.ERROR(result.Message)
	}
	return scope.ExecuteScript(*result.Script)
}

type sourceCmd struct{}

func (sourceCmd) Execute(args []core.Value, context any) core.Result {
	scope := context.(*helena_dialect.Scope)
	if len(args) != 2 {
		return helena_dialect.ARITY_ERROR("source path")
	}
	path := core.ValueToString(args[1]).Data
	return sourceFile(path, scope)
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
	rootScope := helena_dialect.NewScope(nil, false)
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
	tokens := core.Tokenizer{}.Tokenize(cmd)
	if len(tokens) > 0 &&
		tokens[len(tokens)-1].Type == core.TokenType_CONTINUATION {
		// Continuation, wait for next line
		return nil, recoverableError{"continuation"}
	}

	stream := core.NewArrayTokenStream(tokens)
	parser := &core.Parser{}
	parseResult := parser.ParseStream(stream)
	if !parseResult.Success {
		// Parse error
		return nil, basicError{parseResult.Message}
	}

	parseResult = parser.CloseStream()
	if !parseResult.Success {
		// Incomplete script, wait for new line
		return nil, recoverableError{parseResult.Message}
	}

	result := scope.ExecuteScript(*parseResult.Script)
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

func processResult(result core.Result) (core.Value, error) {
	switch result.Code {
	case core.ResultCode_OK:
		return result.Value, nil
	case core.ResultCode_ERROR:
		return nil, basicError{core.ValueToString(result.Value).Data}
	default:
		return nil, (basicError{"unexpected " + core.RESULT_CODE_NAME(result)})
	}
}

var italicGrey = color.New(color.FgBlack).Add(color.Italic)

func resultWriter(output any) string {
	if err, ok := output.(error); ok {
		return color.RedString(err.Error())
	}
	value := core.Display(output, func(displayable any) string {
		if v, ok := displayable.(core.ListValue); ok {
			return helena_dialect.DisplayListValue(v, nil)
		}
		if v, ok := displayable.(core.DictionaryValue); ok {
			return helena_dialect.DisplayDictionaryValue(v, nil)
		}
		return core.DefaultDisplayFunction(displayable)
	})
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
	scope := helena_dialect.NewScope(nil, false)
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
