package helena_dialect_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"helena/core"
	. "helena/helena_dialect"
)

var _ = Describe("Helena dialect", func() {
	var rootScope *Scope

	var tokenizer core.Tokenizer
	var parser *core.Parser

	parse := func(script string) *core.Script {
		return parser.Parse(tokenizer.Tokenize(script)).Script
	}
	execute := func(script string) core.Result {
		return rootScope.ExecuteScript(*parse(script))
	}
	evaluate := func(script string) core.Value {
		return execute(script).Value
	}

	BeforeEach(func() {
		rootScope = NewScope(nil, false)
		InitCommands(rootScope)

		tokenizer = core.Tokenizer{}
		parser = &core.Parser{}
	})

	Describe("leading tuple auto-expansion", func() {
		Specify("zero", func() {
			Expect(evaluate("()")).To(Equal(NIL))
			Expect(evaluate("() idem val")).To(Equal(STR("val")))
			Expect(evaluate("(())")).To(Equal(NIL))
			Expect(evaluate("(()) idem val")).To(Equal(STR("val")))
			Expect(evaluate("((())) idem val")).To(Equal(STR("val")))
		})
		Specify("one", func() {
			Expect(evaluate("(idem) val")).To(Equal(STR("val")))
			Expect(evaluate("((idem)) val")).To(Equal(STR("val")))
			Expect(evaluate("((idem)) val")).To(Equal(STR("val")))
			Expect(evaluate("(((idem))) val")).To(Equal(STR("val")))
		})
		Specify("two", func() {
			Expect(evaluate("(idem val)")).To(Equal(STR("val")))
			Expect(evaluate("((idem) val)")).To(Equal(STR("val")))
			Expect(evaluate("(((idem val)))")).To(Equal(STR("val")))
			Expect(evaluate("(((idem) val))")).To(Equal(STR("val")))
			Expect(evaluate("((() idem) val)")).To(Equal(STR("val")))
		})
		Specify("multiple", func() {
			Expect(evaluate("(1)")).To(Equal(INT(1)))
			Expect(evaluate("(+ 1 2)")).To(Equal(INT(3)))
			Expect(evaluate("(+ 1) 2 3")).To(Equal(INT(6)))
			Expect(evaluate("(+ 1 2 3) 4")).To(Equal(INT(10)))
			Expect(evaluate("((+ 1) 2 3) 4 5")).To(Equal(INT(15)))
		})
		// Specify("indirect", func() {
		// 	evaluate("let mac [macro {*args} {+ $*args}]")
		// 	evaluate("let sum ([$mac] 1)")
		// 	Expect(evaluate("$sum 2 3")).To(Equal(INT(6)))
		// })
		Specify("currying", func() {
			evaluate("let double (* 2)")
			evaluate("let quadruple ($double 2)")
			Expect(evaluate("$double 5")).To(Equal(INT(10)))
			Expect(evaluate("$quadruple 3")).To(Equal(INT(12)))
		})
		// Describe("yield", func() {
		// 	It("should provide a resumable state", func() {
		// 		evaluate("macro cmd {*} {yield val1; idem val2}")
		// 		process := rootScope.PrepareScript(*parse("cmd a b c"))

		// 		result := process.Run()
		// 		Expect(result.Code).To(Equal(core.ResultCode_YIELD))
		// 		Expect(result.Value).To(Equal(STR("val1")))

		// 		result = process.Run()
		// 		Expect(result).To(Equal(OK(STR("val2"))))
		// 	})
		// 	It("should work on several levels", func() {
		// 		evaluate("macro cmd2 {*} {yield val2}")
		// 		evaluate("macro cmd {*} {yield val1; yield [cmd2]; idem val4}")
		// 		process := rootScope.PrepareScript(*parse("(((cmd) a) b) c"))

		// 		result := process.Run()
		// 		Expect(result.Code).To(Equal(core.ResultCode_YIELD))
		// 		Expect(result.Value).To(Equal(STR("val1")))

		// 		result = process.Run()
		// 		Expect(result.Code).To(Equal(core.ResultCode_YIELD))
		// 		Expect(result.Value).To(Equal(STR("val2")))

		// 		process.YieldBack(STR("val3"))
		// 		result = process.Run()
		// 		Expect(result.Code).To(Equal(core.ResultCode_YIELD))
		// 		Expect(result.Value).To(Equal(STR("val3")))

		// 		result = process.Run()
		// 		Expect(result).To(Equal(OK(STR("val4"))))
		// 	})
		// })
		Specify("error", func() {
			Expect(execute("(a)")).To(Equal(ERROR(`cannot resolve command "a"`)))
			Expect(execute("() a")).To(Equal(ERROR(`cannot resolve command "a"`)))
			Expect(execute("(()) a")).To(Equal(ERROR(`cannot resolve command "a"`)))
			Expect(execute("([]) a")).To(Equal(ERROR("invalid command name")))
		})
	})

	// TODO example scripts
})
