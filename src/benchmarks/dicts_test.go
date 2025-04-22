package benchmarks

import "testing"

func BenchmarkDict(b *testing.B) {
	runBenchmark(
		"dict",
		"dict (a 1 b 2 c 3)",
		b.N)
}
func BenchmarkDictAdd(b *testing.B) {
	scope := initScope()
	runScript(scope, "set d [dict (a 1 b 2 c 3)]")
	runBenchmarkInScope(
		scope,
		"dict_add",
		"dict $d add d 4",
		b.N)
}
func BenchmarkDictMerge(b *testing.B) {
	scope := initScope()
	runScript(scope, `
		set d [dict (a 1 b 2 c 3)]
		set d1 [dict (a 4 b 5 c 6)]
		set d2 [dict (d 7 e 8 f 9 g 10)]
		set d3 [dict (a 11 d 12 g 13)]
	`)
	runBenchmarkInScope(
		scope,
		"dict_merge",
		"dict $d merge $d1 $d2 $d3",
		b.N)
}
