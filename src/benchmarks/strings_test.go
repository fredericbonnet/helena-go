package benchmarks

import "testing"

func BenchmarkString(b *testing.B) {
	runBenchmark(
		"string",
		"string \"hello\"",
		b.N)
}
func BenchmarkStringLength(b *testing.B) {
	runBenchmark(
		"string_length",
		"string \"hello\" length",
		b.N)
}
