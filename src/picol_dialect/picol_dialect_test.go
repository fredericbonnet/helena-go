package picol_dialect_test

import (
	"fmt"
	"math/rand"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	core "helena/core"
	. "helena/picol_dialect"
)

var TRUE = core.TRUE
var FALSE = core.FALSE
var INT = core.INT
var REAL = core.REAL
var STR = core.STR
var TUPLE = core.TUPLE

var OK = core.OK
var ERROR = core.ERROR

var _ = Describe("Picol dialect", func() {
	var rootScope *PicolScope

	var tokenizer core.Tokenizer
	var parser *core.Parser

	parse := func(script string) *core.Script {
		return parser.Parse(tokenizer.Tokenize(script)).Script
	}
	execute := func(script string) core.Result {
		return rootScope.Evaluator.EvaluateScript(*parse(script))
	}
	evaluate := func(script string) core.Value { return execute(script).Value }

	BeforeEach(func() {
		rootScope = NewPicolScope(nil)
		InitPicolCommands(rootScope)

		tokenizer = core.Tokenizer{}
		parser = &core.Parser{}
	})

	Describe("math", func() {
		Describe("+", func() {
			It("should accept one number", func() {
				Expect(evaluate("+ 3")).To(Equal(REAL(3)))
				Expect(evaluate("+ -1.2e3")).To(Equal(REAL(-1.2e3)))
			})
			It("should add two numbers", func() {
				Expect(evaluate("+ 6 23")).To(Equal(REAL(6 + 23)))
				Expect(evaluate("+ 4.5e-3 -6")).To(Equal(REAL(4.5e-3 - 6)))
			})
			It("should add several numbers", func() {
				expr := "+"
				numbers := make([]float64, 10)
				var total float64 = 0
				for i := 0; i < 10; i++ {
					v := rand.Float64()
					expr += " " + fmt.Sprint(v)
					numbers[i] = v
					total += v
				}
				Expect(evaluate(expr)).To(Equal(REAL(float64(total))))
			})
			Describe("exceptions", func() {
				Specify("wrong arity", func() {
					Expect(execute("+")).To(Equal(
						ERROR(`wrong # args: should be "+ arg ?arg ...?"`),
					))
				})
				Specify("invalid value", func() {
					Expect(execute("+ a")).To(Equal(ERROR(`invalid number "a"`)))
				})
			})
		})
		Describe("-", func() {
			It("should negate one number", func() {
				Expect(evaluate("- 6")).To(Equal(REAL(-6)))
				Expect(evaluate("- -3.4e5")).To(Equal(REAL(3.4e5)))
			})
			It("should subtract two numbers", func() {
				Expect(evaluate("- 4 12")).To(Equal(REAL(4 - 12)))
				Expect(evaluate("- 12.3e4 -56")).To(Equal(REAL(12.3e4 + 56)))
			})
			It("should subtract several numbers", func() {
				expr := "-"
				numbers := make([]float64, 10)
				var total float64 = 0
				for i := 0; i < 10; i++ {
					v := rand.Float64()
					expr += " " + fmt.Sprint(v)
					numbers[i] = v
					if i == 0 {
						total = v
					} else {
						total -= v
					}
				}
				Expect(evaluate(expr)).To(Equal(REAL(total)))
			})
			Describe("exceptions", func() {
				Specify("wrong arity", func() {
					Expect(execute("-")).To(Equal(
						ERROR(`wrong # args: should be "- arg ?arg ...?"`),
					))
				})
				Specify("invalid value", func() {
					Expect(execute("- a")).To(Equal(ERROR(`invalid number "a"`)))
				})
			})
		})
		Describe("*", func() {
			It("should accept one number", func() {
				Expect(evaluate("* 12")).To(Equal(REAL(12)))
				Expect(evaluate("* -67.89")).To(Equal(REAL(-67.89)))
			})
			It("should multiply two numbers", func() {
				Expect(evaluate("* 45 67")).To(Equal(REAL(45 * 67)))
				Expect(evaluate("* 1.23e-4 -56")).To(Equal(REAL(1.23e-4 * -56)))
			})
			It("should multiply several numbers", func() {
				expr := "*"
				numbers := make([]float64, 10)
				var total float64 = 1
				for i := 0; i < 10; i++ {
					v := rand.Float64()
					expr += " " + fmt.Sprint(v)
					numbers[i] = v
					total *= v
				}
				Expect(evaluate(expr)).To(Equal(REAL(total)))
			})
			Describe("exceptions", func() {
				Specify("wrong arity", func() {
					Expect(execute("*")).To(Equal(
						ERROR(`wrong # args: should be "* arg ?arg ...?"`),
					))
				})
				Specify("invalid value", func() {
					Expect(execute("* a")).To(Equal(ERROR(`invalid number "a"`)))
				})
			})
		})
		Describe("/", func() {
			It("should divide two numbers", func() {
				Expect(evaluate("/ 12 -34")).To(Equal(REAL(12.0 / -34.0)))
				Expect(evaluate("/ 45.67e8 -123")).To(Equal(REAL(45.67e8 / -123)))
			})
			It("should divide several numbers", func() {
				expr := "/"
				numbers := make([]float64, 10)
				var total float64
				for i := 0; i < 10; i++ {
					v := rand.Float64()
					expr += " " + fmt.Sprint(v)
					numbers[i] = v
					if i == 0 {
						total = v
					} else {
						total /= v
					}
				}
				Expect(evaluate(expr)).To(Equal(REAL(total)))
			})
			Describe("exceptions", func() {
				Specify("wrong arity", func() {
					Expect(execute("/")).To(Equal(
						ERROR(`wrong # args: should be "/ arg arg ?arg ...?"`),
					))
					Expect(execute("/ 1")).To(Equal(
						ERROR(`wrong # args: should be "/ arg arg ?arg ...?"`),
					))
				})
				Specify("invalid value", func() {
					Expect(execute("/ a 1")).To(Equal(ERROR(`invalid number "a"`)))
					Expect(execute("/ 2 b")).To(Equal(ERROR(`invalid number "b"`)))
				})
			})
		})
	})
	Describe("comparisons", func() {
		Describe("==", func() {
			It("should compare two values", func() {
				Expect(evaluate(`== "123" -34`)).To(Equal(FALSE))
				Expect(evaluate(`== 56 "56"`)).To(Equal(TRUE))
				Expect(evaluate(`== abc "abc"`)).To(Equal(TRUE))
			})
			Describe("exceptions", func() {
				Specify("wrong arity", func() {
					Expect(execute("==")).To(Equal(
						ERROR(`wrong # args: should be "== arg arg"`),
					))
					Expect(execute("== a")).To(Equal(
						ERROR(`wrong # args: should be "== arg arg"`),
					))
				})
			})
		})
		Describe("!=", func() {
			It("should compare two values", func() {
				Expect(evaluate(`!= "123" -34`)).To(Equal(TRUE))
				Expect(evaluate(`!= 56 "56"`)).To(Equal(FALSE))
				Expect(evaluate(`!= abc "abc"`)).To(Equal(FALSE))
			})
			Describe("exceptions", func() {
				Specify("wrong arity", func() {
					Expect(execute("!=")).To(Equal(
						ERROR(`wrong # args: should be "!= arg arg"`),
					))
					Expect(execute("!= a")).To(Equal(
						ERROR(`wrong # args: should be "!= arg arg"`),
					))
				})
			})
		})
		Describe(">", func() {
			It("should compare two numbers", func() {
				Expect(evaluate("> 12 -34")).To(Equal(TRUE))
				Expect(evaluate("> 56 56")).To(Equal(FALSE))
				Expect(evaluate("> -45.6e7 890")).To(Equal(FALSE))
			})
			Describe("exceptions", func() {
				Specify("wrong arity", func() {
					Expect(execute(">")).To(Equal(
						ERROR(`wrong # args: should be "> arg arg"`),
					))
					Expect(execute("> a")).To(Equal(
						ERROR(`wrong # args: should be "> arg arg"`),
					))
				})
				Specify("invalid value", func() {
					Expect(execute("> a 1")).To(Equal(ERROR(`invalid number "a"`)))
					Expect(execute("> 2 b")).To(Equal(ERROR(`invalid number "b"`)))
				})
			})
		})
		Describe(">=", func() {
			It("should compare two numbers", func() {
				Expect(evaluate(">= 12 -34")).To(Equal(TRUE))
				Expect(evaluate(">= 56 56")).To(Equal(TRUE))
				Expect(evaluate(">= -45.6e7 890")).To(Equal(FALSE))
			})
			Describe("exceptions", func() {
				Specify("wrong arity", func() {
					Expect(execute(">=")).To(Equal(
						ERROR(`wrong # args: should be ">= arg arg"`),
					))
					Expect(execute(">= a")).To(Equal(
						ERROR(`wrong # args: should be ">= arg arg"`),
					))
				})
				Specify("invalid value", func() {
					Expect(execute(">= a 1")).To(Equal(ERROR(`invalid number "a"`)))
					Expect(execute(">= 2 b")).To(Equal(ERROR(`invalid number "b"`)))
				})
			})
		})
		Describe("<", func() {
			It("should compare two numbers", func() {
				Expect(evaluate("< 12 -34")).To(Equal(FALSE))
				Expect(evaluate("< 56 56")).To(Equal(FALSE))
				Expect(evaluate("< -45.6e7 890")).To(Equal(TRUE))
			})
			Describe("exceptions", func() {
				Specify("wrong arity", func() {
					Expect(execute("<")).To(Equal(
						ERROR(`wrong # args: should be "< arg arg"`),
					))
					Expect(execute("< a")).To(Equal(
						ERROR(`wrong # args: should be "< arg arg"`),
					))
				})
				Specify("invalid value", func() {
					Expect(execute("< a 1")).To(Equal(ERROR(`invalid number "a"`)))
					Expect(execute("< 2 b")).To(Equal(ERROR(`invalid number "b"`)))
				})
			})
		})
		Describe("<=", func() {
			It("should compare two numbers", func() {
				Expect(evaluate("<= 12 -34")).To(Equal(FALSE))
				Expect(evaluate("<= 56 56")).To(Equal(TRUE))
				Expect(evaluate("<= -45.6e7 890")).To(Equal(TRUE))
			})
			Describe("exceptions", func() {
				Specify("wrong arity", func() {
					Expect(execute("<=")).To(Equal(
						ERROR(`wrong # args: should be "<= arg arg"`),
					))
					Expect(execute("<= a")).To(Equal(
						ERROR(`wrong # args: should be "<= arg arg"`),
					))
				})
				Specify("invalid value", func() {
					Expect(execute("<= a 1")).To(Equal(ERROR(`invalid number "a"`)))
					Expect(execute("<= 2 b")).To(Equal(ERROR(`invalid number "b"`)))
				})
			})
		})
	})
	Describe("logic", func() {
		Describe("!", func() {
			It("should invert boolean values", func() {
				Expect(evaluate("! true")).To(Equal(FALSE))
				Expect(evaluate("! false")).To(Equal(TRUE))
			})
			It("should invert integer values", func() {
				Expect(evaluate("! 1")).To(Equal(FALSE))
				Expect(evaluate("! 123")).To(Equal(FALSE))
				Expect(evaluate("! 0")).To(Equal(TRUE))
			})
			It("should accept block expressions", func() {
				Expect(evaluate("! {!= 1 2}")).To(Equal(FALSE))
				Expect(evaluate("! {== 1 2}")).To(Equal(TRUE))
			})
			It("should propagate return", func() {
				Expect(
					evaluate("proc cmd {} {! {return value}; error}; cmd"),
				).To(Equal(STR("value")))
			})
			Describe("exceptions", func() {
				Specify("wrong arity", func() {
					Expect(execute("!")).To(Equal(
						ERROR(`wrong # args: should be "! arg"`),
					))
				})
				Specify("invalid value", func() {
					Expect(execute("! a")).To(Equal(ERROR(`invalid boolean "a"`)))
				})
			})
		})
		Describe("&&", func() {
			It("should accept one boolean", func() {
				Expect(evaluate("&& false")).To(Equal(FALSE))
				Expect(evaluate("&& true")).To(Equal(TRUE))
			})
			It("should accept two booleans", func() {
				Expect(evaluate("&& false false")).To(Equal(FALSE))
				Expect(evaluate("&& false true")).To(Equal(FALSE))
				Expect(evaluate("&& true false")).To(Equal(FALSE))
				Expect(evaluate("&& true true")).To(Equal(TRUE))
			})
			It("should accept several booleans", func() {
				Expect(evaluate("&&" + strings.Repeat(" true", 3))).To(Equal(TRUE))
				Expect(evaluate("&&" + strings.Repeat(" true", 3) + " false")).To(Equal(FALSE))
			})
			It("should accept block expressions", func() {
				Expect(evaluate("&& {!= 1 2}")).To(Equal(TRUE))
				Expect(evaluate("&& {== 1 2}")).To(Equal(FALSE))
			})
			It("should short-circuit on false", func() {
				Expect(evaluate("&& false {error}")).To(Equal(FALSE))
			})
			It("should propagate return", func() {
				Expect(
					evaluate("proc cmd {} {&& {return value}; error}; cmd"),
				).To(Equal(STR("value")))
			})
			Describe("exceptions", func() {
				Specify("wrong arity", func() {
					Expect(execute("&&")).To(Equal(
						ERROR(`wrong # args: should be "&& arg ?arg ...?"`),
					))
				})
				Specify("invalid value", func() {
					Expect(execute("&& a")).To(Equal(ERROR(`invalid boolean "a"`)))
				})
			})
		})
		Describe("||", func() {
			It("should accept two booleans", func() {
				Expect(evaluate("|| false false")).To(Equal(FALSE))
				Expect(evaluate("|| false true")).To(Equal(TRUE))
				Expect(evaluate("|| true false")).To(Equal(TRUE))
				Expect(evaluate("|| true true")).To(Equal(TRUE))
			})
			It("should accept several booleans", func() {
				Expect(evaluate("||" + strings.Repeat(" false", 3))).To(Equal(FALSE))
				Expect(evaluate("||" + strings.Repeat(" false", 3) + " true")).To(Equal(TRUE))
			})
			It("should accept block expressions", func() {
				Expect(evaluate("|| {!= 1 2}")).To(Equal(TRUE))
				Expect(evaluate("|| {== 1 2}")).To(Equal(FALSE))
			})
			It("should short-circuit on true", func() {
				Expect(evaluate("|| true {error}")).To(Equal(TRUE))
			})
			It("should propagate return", func() {
				Expect(
					evaluate("proc cmd {} {|| {return value}; error}; cmd"),
				).To(Equal(STR("value")))
			})
			Describe("exceptions", func() {
				Specify("wrong arity", func() {
					Expect(execute("||")).To(Equal(
						ERROR(`wrong # args: should be "|| arg ?arg ...?"`),
					))
				})
				Specify("invalid value", func() {
					Expect(execute("|| a")).To(Equal(ERROR(`invalid boolean "a"`)))
				})
			})
		})
	})

	Describe("control flow", func() {
		Describe("if", func() {
			It("should evaluate the if branch when test is true", func() {
				Expect(evaluate("if true {set var if}")).To(Equal(STR("if")))
				Expect(evaluate("if 1 {set var if}")).To(Equal(STR("if")))
			})
			It("should evaluate the else branch when test is false", func() {
				Expect(
					evaluate("if false {set var if} else {set var else}"),
				).To(Equal(STR("else")))
				Expect(evaluate("if 0 {set var if} else {set var else}")).To(Equal(
					STR("else"),
				))
			})
			It("should accept block expressions", func() {
				Expect(
					evaluate("if {!= 1 2} {set var if} else {set var else}"),
				).To(Equal(STR("if")))
				Expect(
					evaluate("if {== 1 2} {set var if} else {set var else}"),
				).To(Equal(STR("else")))
			})
			It("should return empty when test is false and there is no else branch", func() {
				Expect(evaluate("if false {set var if}")).To(Equal(STR("")))
			})
			It("should propagate return", func() {
				evaluate(
					"proc cmd {expr} {if $expr {return if} else {return else}; error}",
				)
				Expect(evaluate("cmd true")).To(Equal(STR("if")))
				Expect(evaluate("cmd false")).To(Equal(STR("else")))
				Expect(evaluate("cmd {!= 1 2}")).To(Equal(STR("if")))
				Expect(evaluate("cmd {== 1 2}")).To(Equal(STR("else")))
				Expect(evaluate("cmd {return test}")).To(Equal(STR("test")))
			})
			Describe("exceptions", func() {
				Specify("wrong arity", func() {
					Expect(execute("if")).To(Equal(
						ERROR(
							`wrong # args: should be "if test script1 ?else script2?"`,
						),
					))
					Expect(execute("if true")).To(Equal(
						ERROR(
							`wrong # args: should be "if test script1 ?else script2?"`,
						),
					))
					Expect(execute("if true {} else")).To(Equal(
						ERROR(
							`wrong # args: should be "if test script1 ?else script2?"`,
						),
					))
					Expect(execute("if true {} else {} {}")).To(Equal(
						ERROR(
							`wrong # args: should be "if test script1 ?else script2?"`,
						),
					))
				})
				Specify("invalid condition", func() {
					Expect(execute("if a {}")).To(Equal(ERROR(`invalid boolean "a"`)))
				})
			})
		})
		Describe("for", func() {
			It("should always evaluate the start segment", func() {
				Expect(evaluate("for {set var start} false {} {}; set var")).To(Equal(
					STR("start"),
				))
				Expect(
					evaluate(
						"for {set var start} {== $var start} {set var ${var}2} {}; set var",
					),
				).To(Equal(STR("start2")))
			})
			It("should skip the body when test is false", func() {
				Expect(
					evaluate(
						"set var before; for {} false {set var next} {set var body}; set var",
					),
				).To(Equal(STR("before")))
			})
			It("should skip next statement when test is false", func() {
				Expect(
					evaluate(
						"set var before; for {} false {set var next} {}; set var",
					),
				).To(Equal(STR("before")))
			})
			It("should loop over the body while test is true", func() {
				Expect(
					evaluate("for {set i 0} {< $i 10} {incr i} {set var $i}; set var"),
				).To(Equal(INT(9)))
			})
			It("should return empty", func() {
				Expect(
					evaluate("for {set i 0} {< $i 10} {incr i} {set var $i}"),
				).To(Equal(STR("")))
			})
			It("should propagate return", func() {
				evaluate(
					"proc cmd {start test next body} {for $start $test $next $body; set var val}",
				)
				Expect(evaluate("cmd {return start} {} {} {}")).To(Equal(
					STR("start"),
				))
				Expect(evaluate("cmd {} {return test} {} {}")).To(Equal(STR("test")))
				Expect(evaluate("cmd {} true {return next} {}")).To(Equal(
					STR("next"),
				))
				Expect(evaluate("cmd {} true {} {return body}")).To(Equal(
					STR("body"),
				))
			})
			Describe("exceptions", func() {
				Specify("wrong arity", func() {
					Expect(execute("for")).To(Equal(
						ERROR(`wrong # args: should be "for start test next command"`),
					))
					Expect(execute("for a b c")).To(Equal(
						ERROR(`wrong # args: should be "for start test next command"`),
					))
					Expect(execute("for a b c d e")).To(Equal(
						ERROR(`wrong # args: should be "for start test next command"`),
					))
				})
				Specify("invalid condition", func() {
					Expect(execute("for {} a {} {} ")).To(Equal(
						ERROR(`invalid boolean "a"`)),
					)
				})
			})
		})
		Describe("while", func() {
			It("should skip the body when test is false", func() {
				Expect(
					evaluate("set var before; while false {set var body}; set var"),
				).To(Equal(STR("before")))
			})
			It("should loop over the body while test is true", func() {
				Expect(evaluate("set i 0; while {< $i 10} {incr i}; set i")).To(Equal(
					INT(10),
				))
			})
			It("should return empty", func() {
				Expect(evaluate("set i 0; while {< $i 10} {incr i}")).To(Equal(
					STR(""),
				))
			})
			It("should propagate return", func() {
				evaluate(
					"proc cmd {test} {while $test {return body; error}; set var val}",
				)
				Expect(evaluate("cmd true")).To(Equal(STR("body")))
				Expect(evaluate("cmd false")).To(Equal(STR("val")))
				Expect(evaluate("cmd {!= 1 2}")).To(Equal(STR("body")))
				Expect(evaluate("cmd {== 1 2}")).To(Equal(STR("val")))
				Expect(evaluate("cmd {return test}")).To(Equal(STR("test")))
			})
			Describe("exceptions", func() {
				Specify("wrong arity", func() {
					Expect(execute("while")).To(Equal(
						ERROR(`wrong # args: should be "while test script"`),
					))
					Expect(execute("while true")).To(Equal(
						ERROR(`wrong # args: should be "while test script"`),
					))
					Expect(execute("while true a b")).To(Equal(
						ERROR(`wrong # args: should be "while test script"`),
					))
				})
				Specify("invalid condition", func() {
					Expect(execute("while a {}")).To(Equal(
						ERROR(`invalid boolean "a"`)),
					)
				})
			})
		})
		Describe("return", func() {
			It("should return empty by default", func() {
				Expect(evaluate("return")).To(Equal(STR("")))
			})
			It("should return an optional result", func() {
				Expect(evaluate("return value")).To(Equal(STR("value")))
			})
			It("should interrupt a proc", func() {
				Expect(evaluate("proc cmd {} {return; set var val}; cmd")).To(Equal(
					STR(""),
				))
			})
			It("should return result from a proc", func() {
				Expect(
					evaluate("proc cmd {} {return result; set var val}; cmd"),
				).To(Equal(STR("result")))
			})
			Describe("exceptions", func() {
				Specify("wrong arity", func() {
					Expect(execute("return a b")).To(Equal(
						ERROR(`wrong # args: should be "return ?result?"`),
					))
				})
			})
		})
		Describe("break", func() {
			It("should interrupt a for loop", func() {
				Expect(
					evaluate(
						"for {set i 0} {< $i 10} {incr i} {set var before$i; break; set var after$i}; set var",
					),
				).To(Equal(STR("before0")))
			})
			It("should interrupt a while loop", func() {
				Expect(
					evaluate(
						"while true {set var before; break; set var after}; set var",
					),
				).To(Equal(STR("before")))
			})
			It("should not interrupt a proc", func() {
				Expect(
					evaluate(
						"proc cmd {} {for {set i 0} {< $i 10} {incr i} {break}; set i}; cmd",
					),
				).To(Equal(STR("0")))
			})
			Describe("exceptions", func() {
				Specify("wrong arity", func() {
					Expect(execute("break a")).To(Equal(
						ERROR(`wrong # args: should be "break"`),
					))
				})
			})
		})
		Describe("continue", func() {
			It("should interrupt a for loop iteration", func() {
				Expect(
					evaluate(
						"for {set i 0} {< $i 10} {incr i} {set var before$i; continue; set var after$i}; set var",
					),
				).To(Equal(STR("before9")))
			})
			It("should interrupt a while loop iteration", func() {
				Expect(
					evaluate(
						"set i 0; while {< $i 10} {incr i; set var before$i; continue; set var after$i}; set var",
					),
				).To(Equal(STR("before10")))
			})
			It("should not interrupt a proc", func() {
				Expect(
					evaluate(
						"proc cmd {} {for {set i 0} {< $i 10} {incr i} {continue}; set i}; cmd",
					),
				).To(Equal(INT(10)))
			})
			Describe("exceptions", func() {
				Specify("wrong arity", func() {
					Expect(execute("continue a")).To(Equal(
						ERROR(`wrong # args: should be "continue"`),
					))
				})
			})
		})
		Describe("error", func() {
			It("should interrupt a for loop", func() {
				Expect(
					execute(
						"for {set i 0} {< $i 10} {incr i} {set var before$i; error message; set var after$i}; set var",
					),
				).To(Equal(ERROR("message")))
			})
			It("should interrupt a while loop", func() {
				Expect(
					execute(
						"while true {set var before; error message; set var after}; set var",
					),
				).To(Equal(ERROR("message")))
			})
			It("should interrupt a proc", func() {
				Expect(
					execute("proc cmd {} {error message; set var val}; cmd"),
				).To(Equal(ERROR("message")))
			})
			Describe("exceptions", func() {
				Specify("wrong arity", func() {
					Expect(execute("error")).To(Equal(
						ERROR(`wrong # args: should be "error message"`),
					))
				})
			})
		})
	})

	Describe("set", func() {
		It("should return the value of an existing variable", func() {
			rootScope.Variables["var"] = STR("val")
			Expect(evaluate("set var")).To(Equal(STR("val")))
		})
		It("should set the value of a new variable", func() {
			evaluate("set var val")
			Expect(rootScope.Variables["var"]).To(Equal(STR("val")))
		})
		It("should overwrite the value of an existing variable", func() {
			rootScope.Variables["var"] = STR("old")
			evaluate("set var val")
			Expect(rootScope.Variables["var"]).To(Equal(STR("val")))
		})
		It("should return the set value", func() {
			Expect(evaluate("set var val")).To(Equal(STR("val")))
		})
		Describe("exceptions", func() {
			Specify("non-existing variable", func() {
				Expect(execute("set unknownVariable")).To(Equal(
					ERROR(`can't read "unknownVariable": no such variable`),
				))
			})
			Specify("wrong arity", func() {
				Expect(execute("set")).To(Equal(
					ERROR(`wrong # args: should be "set varName ?newValue?"`),
				))
				Expect(execute("set a b c")).To(Equal(
					ERROR(`wrong # args: should be "set varName ?newValue?"`),
				))
			})
		})
	})
	Describe("incr", func() {
		It("should set new variables to the increment", func() {
			evaluate("incr var 5")
			Expect(rootScope.Variables["var"]).To(Equal(INT(5)))
		})
		It("should increment existing variables by the increment", func() {
			rootScope.Variables["var"] = INT(2)
			evaluate("incr var 4")
			Expect(rootScope.Variables["var"]).To(Equal(INT(6)))
		})
		Specify("increment should default to 1", func() {
			rootScope.Variables["var"] = INT(1)
			evaluate("incr var")
			Expect(rootScope.Variables["var"]).To(Equal(INT(2)))
		})
		It("should return the new value", func() {
			Expect(evaluate("set var 1; incr var")).To(Equal(INT(2)))
		})
		Describe("exceptions", func() {
			Specify("wrong arity", func() {
				Expect(execute("incr")).To(Equal(
					ERROR(`wrong # args: should be "incr varName ?increment?"`),
				))
				Expect(execute("incr a 1 2")).To(Equal(
					ERROR(`wrong # args: should be "incr varName ?increment?"`),
				))
			})
			Specify("invalid variable value", func() {
				Expect(execute("set var a; incr var")).To(Equal(
					ERROR(`invalid integer "a"`)),
				)
			})
			Specify("invalid increment", func() {
				Expect(execute("incr var a")).To(Equal(ERROR(`invalid integer "a"`)))
			})
		})
	})

	Describe("proc", func() {
		It("should define a new command", func() {
			evaluate("proc cmd {} {}")
			Expect(rootScope.Commands["cmd"]).NotTo(BeNil())
		})
		It("should replace existing commands", func() {
			evaluate("proc cmd {} {}")
			evaluate("proc cmd {} {}")
			Expect(execute("proc cmd {} {}").Code).To(Equal(core.ResultCode_OK))
		})
		It("should return empty", func() {
			Expect(evaluate("proc cmd {} {}")).To(Equal(STR("")))
		})
		Describe("calls", func() {
			It("should return empty string for empty body", func() {
				evaluate("proc cmd {} {}")
				Expect(evaluate("cmd")).To(Equal(STR("")))
			})
			It("should return the result of the last command", func() {
				evaluate("proc cmd {} {set var val}")
				Expect(evaluate("cmd")).To(Equal(STR("val")))
			})
			It("should access global commands", func() {
				evaluate("proc cmd2 {} {set var val}")
				evaluate("proc cmd {} {cmd2}")
				Expect(evaluate("cmd")).To(Equal(STR("val")))
			})
			It("should not access global variables", func() {
				evaluate("set var val")
				evaluate("proc cmd {} {set var}")
				Expect(execute("cmd").Code).To(Equal(core.ResultCode_ERROR))
			})
			It("should not set global variables", func() {
				evaluate("set var val")
				evaluate("proc cmd {} {set var val2}")
				evaluate("cmd")
				Expect(rootScope.Variables["var"]).To(Equal(STR("val")))
			})
			It("should set local variables", func() {
				evaluate("set var val")
				evaluate("proc cmd {} {set var2 val}")
				evaluate("cmd")
				Expect(rootScope.Variables["var2"]).To(BeNil())
			})
			It("should map arguments to local variables", func() {
				evaluate("proc cmd {param} {set param}")
				Expect(evaluate("cmd arg")).To(Equal(STR("arg")))
				Expect(rootScope.Variables["param"]).To(BeNil())
			})
			It("should accept default argument values", func() {
				evaluate(
					"proc cmd {param1 {param2 def}} {set var ($param1 $param2)}",
				)
				Expect(evaluate("cmd arg")).To(Equal(TUPLE([]core.Value{STR("arg"), STR("def")})))
			})
			It("should accept remaining args", func() {
				evaluate("proc cmd {param1 param2 args} {set var $args}")
				Expect(evaluate("cmd 1 2")).To(Equal(TUPLE([]core.Value{})))
				Expect(evaluate("cmd 1 2 3 4")).To(Equal(TUPLE([]core.Value{STR("3"), STR("4")})))
			})
			It("should accept both default and remaining args", func() {
				evaluate(
					"proc cmd {param1 {param2 def} args} {set var ($param1 $param2 $*args)}",
				)
				Expect(evaluate("cmd 1 2")).To(Equal(TUPLE([]core.Value{STR("1"), STR("2")})))
				Expect(evaluate("cmd 1 2 3 4")).To(Equal(
					TUPLE([]core.Value{STR("1"), STR("2"), STR("3"), STR("4")}),
				))
			})
			Describe("exceptions", func() {
				Specify("not enough arguments", func() {
					Expect(execute("proc cmd {a b} {}; cmd 1")).To(Equal(
						ERROR(`wrong # args: should be "cmd a b"`),
					))
					Expect(execute("proc cmd {a b args} {}; cmd 1")).To(Equal(
						ERROR(`wrong # args: should be "cmd a b ?arg ...?"`),
					))
				})
				Specify("too many arguments", func() {
					Expect(execute("proc cmd {} {}; cmd 1 2")).To(Equal(
						ERROR(`wrong # args: should be "cmd"`),
					))
					Expect(execute("proc cmd {a} {}; cmd 1 2")).To(Equal(
						ERROR(`wrong # args: should be "cmd a"`),
					))
					Expect(execute("proc cmd {a {b 1}} {}; cmd 1 2 3")).To(Equal(
						ERROR(`wrong # args: should be "cmd a ?b?"`),
					))
				})
			})
		})
		Describe("exceptions", func() {
			Specify("wrong arity", func() {
				Expect(execute("proc")).To(Equal(
					ERROR(`wrong # args: should be "proc name args body"`),
				))
				Expect(execute("proc a")).To(Equal(
					ERROR(`wrong # args: should be "proc name args body"`),
				))
				Expect(execute("proc a b")).To(Equal(
					ERROR(`wrong # args: should be "proc name args body"`),
				))
				Expect(execute("proc a b c d")).To(Equal(
					ERROR(`wrong # args: should be "proc name args body"`),
				))
			})
			Specify("argument with no name", func() {
				Expect(execute("proc cmd {{}} {}")).To(Equal(
					ERROR("argument with no name"),
				))
				Expect(execute("proc cmd {{{} def}} {}")).To(Equal(
					ERROR("argument with no name"),
				))
			})
			Specify("wrong argument specifier format", func() {
				Expect(execute("proc cmd {{a b c}} {}")).To(Equal(
					ERROR(`too many fields in argument specifier "a b c"`),
				))
			})
		})
	})
})
