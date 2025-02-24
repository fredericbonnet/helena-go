package benchmarks

import (
	"testing"
)

func BenchmarkNotTrue(b *testing.B) {
	runBenchmark(
		"not_true",
		"! true",
		b.N)
}
func BenchmarkNotScript(b *testing.B) {
	runBenchmark(
		"not_script",
		"! {true}",
		b.N)
}

func BenchmarkAndTrue(b *testing.B) {
	runBenchmark(
		"and_true",
		"&& true",
		b.N)
}
func BenchmarkAndScript(b *testing.B) {
	runBenchmark(
		"and_script",
		"&& {true}",
		b.N)
}
func BenchmarkAndMultiple(b *testing.B) {
	runBenchmark(
		"and_multiple",
		"&& {true} true {true} true {true} {true} true",
		b.N)
}

func BenchmarkOrTrue(b *testing.B) {
	runBenchmark(
		"or_true",
		"|| true",
		b.N)
}
func BenchmarkOrScript(b *testing.B) {
	runBenchmark(
		"or_script",
		"|| {true}",
		b.N)
}
func BenchmarkOrMultiple(b *testing.B) {
	runBenchmark(
		"or_multiple",
		"|| {false} false {false} false {false} {false} false",
		b.N)
}
