package benchmarks

import "testing"

func BenchmarkLoopBreak(b *testing.B) {
	runBenchmark(
		"loop_break",
		"loop {break}",
		b.N)
}
func BenchmarkLoop100(b *testing.B) {
	runBenchmark(
		"loop_100",
		"loop i {if [$i == 100] {break}}",
		b.N/100)
}

func BenchmarkIfScript(b *testing.B) {
	runBenchmark(
		"if_script",
		"if {true} {}",
		b.N)
}
func BenchmarkIfMultiple(b *testing.B) {
	runBenchmark(
		"if_multiple",
		"if {false} {} elseif false {} elseif {false} {} elseif false {} elseif {false} {} elseif false {} elseif {false} {} else {}",
		b.N)
}

func BenchmarkWhileTrue(b *testing.B) {
	runBenchmark(
		"while_true",
		"while true {break}",
		b.N)
}
func BenchmarkWhileScript(b *testing.B) {
	runBenchmark(
		"while_script",
		"while {true} {break}",
		b.N)
}

func BenchmarkWhenCommand(b *testing.B) {
	runBenchmark(
		"when_command",
		"when {true} {}",
		b.N)
}
func BenchmarkWhenTrue(b *testing.B) {
	runBenchmark(
		"when_true",
		"when {true {}}",
		b.N)
}
func BenchmarkWhenScript(b *testing.B) {
	runBenchmark(
		"when_script",
		"when {{true} {}}",
		b.N)
}
func BenchmarkWhenTuple(b *testing.B) {
	runBenchmark(
		"when_tuple",
		"when {(true) {}}",
		b.N)
}
func BenchmarkWhenMultiple(b *testing.B) {
	runBenchmark(
		"when_multiple",
		"when {1} {(== 2) {} (== 3) {} (> 4) {} {}}",
		b.N)
}
