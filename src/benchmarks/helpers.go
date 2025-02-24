package benchmarks

import (
	"helena/core"
	"helena/helena_dialect"
	"log"
	"os"
	"runtime/pprof"
)

func initScope() *helena_dialect.Scope {
	scope := helena_dialect.NewRootScope(nil)
	helena_dialect.InitCommands(scope)
	return scope
}
func runScript(scope *helena_dialect.Scope, script string) {
	tokens := core.Tokenizer{}.Tokenize(script)
	result := core.NewParser(nil).ParseTokens(tokens, nil)
	program := scope.Compile(*result.Script)
	process := scope.PrepareProcess(program)
	process.Run()
}
func runTest(label string, script string) {
	runTestInScope(initScope(), label, script)
}
func runTestInScope(scope *helena_dialect.Scope, label string, script string) {
	tokens := core.Tokenizer{}.Tokenize(script)
	result := core.NewParser(nil).ParseTokens(tokens, nil)
	f, err := os.Create("./" + label + ".prof")
	if err != nil {
		log.Fatal("could not create CPU profile: ", err)
	}
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()
	program := scope.Compile(*result.Script)
	process := scope.PrepareProcess(program)
	process.Run()
}

func runBenchmark(label string, script string, n int) {
	runBenchmarkInScope(initScope(), label, script, n)
}
func runBenchmarkInScope(scope *helena_dialect.Scope, label string, script string, n int) {
	// n = 100
	tokens := core.Tokenizer{}.Tokenize(script)
	result := core.NewParser(nil).ParseTokens(tokens, nil)
	f, err := os.Create("./" + label + ".prof")
	if err != nil {
		log.Fatal("could not create CPU profile: ", err)
	}
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()
	program := scope.Compile(*result.Script)
	for i := 0; i < n; i++ {
		process := scope.PrepareProcess(program)
		process.Run()
	}
}
