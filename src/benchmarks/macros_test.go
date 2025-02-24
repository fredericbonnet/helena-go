package benchmarks

import "testing"

func BenchmarkMacro(b *testing.B) {
	scope := initScope()
	runScript(scope, "macro cmd {} {}")
	runBenchmarkInScope(scope, "macro", "cmd", b.N)
}
