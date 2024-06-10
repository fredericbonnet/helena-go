package helena_dialect_test

import (
	"fmt"
	"math/rand"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"helena/core"
	. "helena/helena_dialect"
)

var _ = Describe("Helena math operations", func() {
	var rootScope *Scope

	var tokenizer core.Tokenizer
	var parser *core.Parser

	parse := func(script string) *core.Script {
		return parser.Parse(tokenizer.Tokenize(script)).Script
	}
	prepareScript := func(script string) *Process {
		return rootScope.PrepareProcess(rootScope.Compile(*parse(script)))
	}
	execute := func(script string) core.Result {
		return prepareScript(script).Run()
	}
	evaluate := func(script string) core.Value {
		return execute(script).Value
	}
	init := func() {
		rootScope = NewRootScope(nil)
		InitCommands(rootScope)

		tokenizer = core.Tokenizer{}
		parser = core.NewParser(nil)
	}

	BeforeEach(init)

	Describe("Prefix operators", func() {
		Describe("Arithmetic", func() {
			Describe("`+`", func() {
				Specify("usage", func() {
					Expect(evaluate("help +")).To(Equal(STR("+ number ?number ...?")))
				})

				It("should accept one number", func() {
					Expect(evaluate("+ 3")).To(Equal(INT(3)))
					Expect(evaluate("+ -1.2e-3")).To(Equal(REAL(-1.2e-3)))
				})
				It("should add two numbers", func() {
					Expect(evaluate("+ 6 23")).To(Equal(INT(6 + 23)))
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

				Describe("Exceptions", func() {
					Specify("wrong arity", func() {
						Expect(execute("+")).To(Equal(
							ERROR(`wrong # args: should be "+ number ?number ...?"`),
						))
					})
					Specify("invalid value", func() {
						Expect(execute("+ a")).To(Equal(ERROR(`invalid number "a"`)))
					})
				})
			})

			Describe("`-`", func() {
				Specify("usage", func() {
					Expect(evaluate("help -")).To(Equal(STR("- number ?number ...?")))
				})

				It("should negate one number", func() {
					Expect(evaluate("- 6")).To(Equal(INT(-6)))
					Expect(evaluate("- -3.4e-5")).To(Equal(REAL(3.4e-5)))
				})
				It("should subtract two numbers", func() {
					Expect(evaluate("- 4 12")).To(Equal(INT(4 - 12)))
					Expect(evaluate("- 12.3e-4 -56")).To(Equal(REAL(12.3e-4 + 56)))
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

				Describe("Exceptions", func() {
					Specify("wrong arity", func() {
						Expect(execute("-")).To(Equal(
							ERROR(`wrong # args: should be "- number ?number ...?"`),
						))
					})
					Specify("invalid value", func() {
						Expect(execute("- a")).To(Equal(ERROR(`invalid number "a"`)))
					})
				})
			})

			Describe("`*`", func() {
				Specify("usage", func() {
					Expect(evaluate("help *")).To(Equal(STR("* number ?number ...?")))
				})

				It("should accept one number", func() {
					Expect(evaluate("* 12")).To(Equal(INT(12)))
					Expect(evaluate("* -67.89")).To(Equal(REAL(-67.89)))
				})
				It("should multiply two numbers", func() {
					Expect(evaluate("* 45 67")).To(Equal(INT(45 * 67)))
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

				Describe("Exceptions", func() {
					Specify("wrong arity", func() {
						Expect(execute("*")).To(Equal(
							ERROR(`wrong # args: should be "* number ?number ...?"`),
						))
					})
					Specify("invalid value", func() {
						Expect(execute("* a")).To(Equal(ERROR(`invalid number "a"`)))
					})
				})
			})

			Describe("`/`", func() {
				Specify("usage", func() {
					Expect(evaluate("help /")).To(Equal(
						STR("/ number number ?number ...?"),
					))
				})

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

				Describe("Exceptions", func() {
					Specify("wrong arity", func() {
						Expect(execute("/")).To(Equal(
							ERROR(`wrong # args: should be "/ number number ?number ...?"`),
						))
						Expect(execute("/ 1")).To(Equal(
							ERROR(`wrong # args: should be "/ number number ?number ...?"`),
						))
					})
					Specify("invalid value", func() {
						Expect(execute("/ a 1")).To(Equal(ERROR(`invalid number "a"`)))
						Expect(execute("/ 2 b")).To(Equal(ERROR(`invalid number "b"`)))
					})
				})
			})
		})
	})
})
