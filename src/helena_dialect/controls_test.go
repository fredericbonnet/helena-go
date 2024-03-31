package helena_dialect_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"helena/core"
	. "helena/helena_dialect"
)

var _ = Describe("Helena control flow commands", func() {
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
	init := func() {
		rootScope = NewScope(nil, false)
		InitCommands(rootScope)

		tokenizer = core.Tokenizer{}
		parser = &core.Parser{}
	}

	BeforeEach(init)

	Describe("while", func() {
		Describe("Specifications", func() {
			Specify("usage", func() {
				Expect(evaluate("help while")).To(Equal(STR("while test body")))
				Expect(evaluate("help while val")).To(Equal(STR("while test body")))
			})

			It("should skip `body` when `test` is false", func() {
				Expect(execute("while false {unreachable}").Code).To(Equal(core.ResultCode_OK))
			})
			It("should loop over `body` while `test` is true", func() {
				evaluate("set i 0; while {$i < 10} {set i [+ $i 1]}")
				Expect(evaluate("get i")).To(Equal(INT(10)))
			})
			It("should return the result of the last command", func() {
				Expect(execute(" while false {}")).To(Equal(OK(NIL)))
				Expect(
					evaluate("set i 0; while {$i < 10} {set i [+ $i 1]; idem val$i}"),
				).To(Equal(STR("val10")))
			})
		})

		Describe("Exceptions", func() {
			Specify("wrong arity", func() {
				Expect(execute("while a")).To(Equal(
					ERROR(`wrong # args: should be "while test body"`),
				))
				Expect(execute("while a b c")).To(Equal(
					ERROR(`wrong # args: should be "while test body"`),
				))
				Expect(execute("help while a b c")).To(Equal(
					ERROR(`wrong # args: should be "while test body"`),
				))
			})
			Specify("non-script body", func() {
				Expect(execute("while a b")).To(Equal(ERROR("body must be a script")))
			})
		})

		Describe("Control flow", func() {
			Describe("`return`", func() {
				It("should interrupt the test with `RETURN` code", func() {
					Expect(
						execute("while {return val; unreachable} {unreachable}"),
					).To(Equal(RETURN(STR("val"))))
				})
				It("should interrupt the loop with `RETURN` code", func() {
					Expect(
						execute(
							"set i 0; while {$i < 10} {set i [+ $i 1]; return val; unreachable}",
						),
					).To(Equal(RETURN(STR("val"))))
					Expect(evaluate("get i")).To(Equal(INT(1)))
				})
			})
			Describe("`tailcall`", func() {
				It("should interrupt the test with `RETURN` code", func() {
					Expect(
						execute("while {tailcall {idem val}; unreachable} {unreachable}"),
					).To(Equal(RETURN(STR("val"))))
				})
				It("should interrupt the loop with `RETURN` code", func() {
					Expect(
						execute(
							"set i 0; while {$i < 10} {set i [+ $i 1]; tailcall {idem val}; unreachable}",
						),
					).To(Equal(RETURN(STR("val"))))
					Expect(evaluate("get i")).To(Equal(INT(1)))
				})
			})
			Describe("`yield`", func() {
				It("should interrupt the test with `YIELD` code", func() {
					Expect(execute("while {yield; unreachable} {}").Code).To(Equal(
						core.ResultCode_YIELD,
					))
				})
				It("should interrupt the body with `YIELD` code", func() {
					Expect(execute("while {true} {yield; unreachable}").Code).To(Equal(
						core.ResultCode_YIELD,
					))
				})
				It("should provide a resumable state", func() {
					process := rootScope.PrepareScript(
						*parse("while {yield test} {yield body}"),
					)

					result := process.Run()
					Expect(result.Code).To(Equal(core.ResultCode_YIELD))
					Expect(result.Value).To(Equal(STR("test")))
					Expect(result.Data).NotTo(BeNil())

					process.YieldBack(TRUE)
					result = process.Run()
					Expect(result.Code).To(Equal(core.ResultCode_YIELD))
					Expect(result.Value).To(Equal(STR("body")))
					Expect(result.Data).NotTo(BeNil())

					process.YieldBack(STR("step 1"))
					result = process.Run()
					Expect(result.Code).To(Equal(core.ResultCode_YIELD))
					Expect(result.Value).To(Equal(STR("test")))
					Expect(result.Data).NotTo(BeNil())

					process.YieldBack(TRUE)
					result = process.Run()
					Expect(result.Code).To(Equal(core.ResultCode_YIELD))
					Expect(result.Value).To(Equal(STR("body")))
					Expect(result.Data).NotTo(BeNil())

					process.YieldBack(STR("step 2"))
					result = process.Run()
					Expect(result.Code).To(Equal(core.ResultCode_YIELD))
					Expect(result.Value).To(Equal(STR("test")))
					Expect(result.Data).NotTo(BeNil())

					process.YieldBack(FALSE)
					result = process.Run()
					Expect(result).To(Equal(OK(STR("step 2"))))
				})
			})
			Describe("`error`", func() {
				It("should interrupt the test with `ERROR` code", func() {
					Expect(
						execute("while {error msg; set var val} {unreachable}"),
					).To(Equal(ERROR("msg")))
					Expect(execute("get var").Code).To(Equal(core.ResultCode_ERROR))
				})
				It("should interrupt the loop with `ERROR` code", func() {
					Expect(
						execute(
							"set i 0; while {$i < 10} {set i [+ $i 1]; error msg; set var val}",
						),
					).To(Equal(ERROR("msg")))
					Expect(evaluate("get i")).To(Equal(INT(1)))
					Expect(execute("get var").Code).To(Equal(core.ResultCode_ERROR))
				})
			})
			Describe("`break`", func() {
				It("should interrupt the test with `BREAK` code", func() {
					Expect(execute("while {break; unreachable} {unreachable}")).To(Equal(
						BREAK(NIL),
					))
				})
				It("should interrupt the body with `nil` result", func() {
					Expect(
						execute(
							"set i 0; while {$i < 10} {set i [+ $i 1]; break; unreachable}",
						),
					).To(Equal(OK(NIL)))
					Expect(evaluate("get i")).To(Equal(INT(1)))
				})
			})
			Describe("`continue`", func() {
				It("should interrupt the test with `CONTINUE` code", func() {
					Expect(execute("while {continue; unreachable} {unreachable}")).To(Equal(
						CONTINUE(NIL),
					))
				})
				It("should interrupt the body iteration", func() {
					Expect(
						execute(
							"set i 0; while {$i < 10} {set i [+ $i 1]; continue; unreachable}",
						),
					).To(Equal(OK(NIL)))
					Expect(evaluate("get i")).To(Equal(INT(10)))
				})
			})
		})
	})

	Describe("if", func() {
		Describe("Specifications", func() {
			Specify("usage", func() {
				Expect(evaluate("help if")).To(Equal(
					STR("if test body ?elseif test body ...? ?else? ?body?"),
				))
			})

			It("should return the result of the first true body", func() {
				Expect(evaluate("if true {1}")).To(Equal(INT(1)))
				Expect(evaluate("if true {1} else {2}")).To(Equal(INT(1)))
				Expect(evaluate("if true {1} elseif true {2} else {3}")).To(Equal(INT(1)))
				Expect(
					evaluate("if false {1} elseif true {2} elseif true {3} else {4}"),
				).To(Equal(INT(2)))
				Expect(evaluate("if false {1} elseif true {2} else {3}")).To(Equal(
					INT(2),
				))
				Expect(
					evaluate("if false {1} elseif true {2} elseif true {3} else {4}"),
				).To(Equal(INT(2)))
			})
			It("should return the result of the `else` body when all tests are false", func() {
				Expect(evaluate("if false {1} else {2}")).To(Equal(INT(2)))
				Expect(evaluate("if false {1} elseif false {2} else {3}")).To(Equal(
					INT(3),
				))
				Expect(
					evaluate("if false {1} elseif false {2} elseif false {3} else {4}"),
				).To(Equal(INT(4)))
			})
			It("should skip leading false bodies", func() {
				Expect(evaluate("if false {unreachable}")).To(Equal(NIL))
				Expect(
					evaluate("if false {unreachable} elseif false {unreachable}"),
				).To(Equal(NIL))
				Expect(
					evaluate(
						"if false {unreachable} elseif false {unreachable} elseif false {unreachable}",
					),
				).To(Equal(NIL))
			})
			It("should skip trailing tests and bodies", func() {
				Expect(evaluate("if true {1} else {unreachable}")).To(Equal(INT(1)))
				Expect(
					evaluate("if true {1} elseif {unreachable} {unreachable}"),
				).To(Equal(INT(1)))
				Expect(
					evaluate(
						"if true {1} elseif {unreachable} {unreachable} else {unreachable}",
					),
				).To(Equal(INT(1)))
				Expect(
					evaluate(
						"if false {1} elseif true {2} elseif {unreachable} {unreachable}",
					),
				).To(Equal(INT(2)))
			})
		})

		Describe("Exceptions", func() {
			Specify("wrong arity", func() {
				Expect(execute("if")).To(Equal(
					ERROR(
						`wrong # args: should be "if test body ?elseif test body ...? ?else? ?body?"`,
					),
				))
				Expect(execute("if a")).To(Equal(ERROR("wrong # args: missing if body")))
				Expect(execute("if a b else")).To(Equal(
					ERROR("wrong # args: missing else body"),
				))
				Expect(execute("if a b elseif")).To(Equal(
					ERROR("wrong # args: missing elseif test"),
				))
				Expect(execute("if a b elseif c")).To(Equal(
					ERROR("wrong # args: missing elseif body"),
				))
				Expect(execute("if a b elseif c d else")).To(Equal(
					ERROR("wrong # args: missing else body"),
				))
			})
			Specify("invalid keyword", func() {
				Expect(execute("if a b elif c d")).To(Equal(
					ERROR(`invalid keyword "elif"`),
				))
				Expect(execute("if a b fi")).To(Equal(ERROR(`invalid keyword "fi"`)))
				Expect(execute("if a b []")).To(Equal(ERROR("invalid keyword")))
			})
			Specify("invalid test", func() {
				Expect(execute("if a b")).To(Equal(ERROR(`invalid boolean "a"`)))
				Expect(execute("if false a elseif b c")).To(Equal(
					ERROR(`invalid boolean "b"`),
				))
				Expect(execute("if false a elseif false b elseif c d")).To(Equal(
					ERROR(`invalid boolean "c"`),
				))
			})
			Specify("non-script body", func() {
				Expect(execute("if true a")).To(Equal(ERROR("body must be a script")))
				Expect(execute("if false {} else a ")).To(Equal(
					ERROR("body must be a script"),
				))
				Expect(execute("if false {} elseif true a")).To(Equal(
					ERROR("body must be a script"),
				))
				Expect(execute("if false {} elseif false {} else a")).To(Equal(
					ERROR("body must be a script"),
				))
			})
		})

		Describe("Control flow", func() {
			Describe("`return`", func() {
				It("should interrupt tests with `RETURN` code", func() {
					Expect(execute("if {return val; unreachable} {unreachable}")).To(Equal(
						RETURN(STR("val")),
					))
					Expect(
						execute(
							"if false {} elseif {return val; unreachable} {unreachable}",
						),
					).To(Equal(RETURN(STR("val"))))
				})
				It("should interrupt bodies with `RETURN` code", func() {
					Expect(execute("if true {return val; unreachable}")).To(Equal(
						RETURN(STR("val")),
					))
					Expect(
						execute("if false {} elseif true {return val; unreachable}"),
					).To(Equal(RETURN(STR("val"))))
					Expect(
						execute(
							"if false {} elseif false {} else {return val; unreachable}",
						),
					).To(Equal(RETURN(STR("val"))))
				})
			})
			Describe("`tailcall`", func() {
				It("should interrupt tests with `RETURN` code", func() {
					Expect(
						execute("if {tailcall {idem val}; unreachable} {unreachable}"),
					).To(Equal(RETURN(STR("val"))))
					Expect(
						execute(
							"if false {} elseif {tailcall {idem val}; unreachable} {unreachable}",
						),
					).To(Equal(RETURN(STR("val"))))
				})
				It("should interrupt bodies with `RETURN` code", func() {
					Expect(execute("if true {tailcall {idem val}; unreachable}")).To(Equal(
						RETURN(STR("val")),
					))
					Expect(
						execute(
							"if false {} elseif true {tailcall {idem val}; unreachable}",
						),
					).To(Equal(RETURN(STR("val"))))
					Expect(
						execute(
							"if false {} elseif false {} else {tailcall {idem val}; unreachable}",
						),
					).To(Equal(RETURN(STR("val"))))
				})
			})
			Describe("`yield`", func() {
				It("should interrupt tests with `YIELD` code", func() {
					Expect(execute("if {yield; unreachable} {unreachable}").Code).To(Equal(
						core.ResultCode_YIELD,
					))
					Expect(
						execute("if false {} elseif {yield; unreachable} {unreachable}").Code,
					).To(Equal(core.ResultCode_YIELD))
				})
				It("should interrupt bodies with `YIELD` code", func() {
					Expect(execute("if true {yield; unreachable}").Code).To(Equal(
						core.ResultCode_YIELD,
					))
					Expect(
						execute("if false {} elseif true {yield; unreachable}").Code,
					).To(Equal(core.ResultCode_YIELD))
					Expect(
						execute("if false {} elseif false {} else {yield; unreachable}").Code,
					).To(Equal(core.ResultCode_YIELD))
				})
				Describe("should provide a resumable state", func() {
					var process *Process
					BeforeEach(func() {
						process = rootScope.PrepareScript(
							*parse(
								"if {yield test1} {yield body1} elseif {yield test2} {yield body2} else {yield body3}",
							),
						)
					})
					Specify("if", func() {
						result := process.Run()
						Expect(result.Code).To(Equal(core.ResultCode_YIELD))
						Expect(result.Value).To(Equal(STR("test1")))
						Expect(result.Data).NotTo(BeNil())

						process.YieldBack(TRUE)
						result = process.Run()
						Expect(result.Code).To(Equal(core.ResultCode_YIELD))
						Expect(result.Value).To(Equal(STR("body1")))
						Expect(result.Data).NotTo(BeNil())

						process.YieldBack(STR("result"))
						result = process.Run()
						Expect(result).To(Equal(OK(STR("result"))))
					})
					Specify("elseif", func() {
						result := process.Run()
						Expect(result.Code).To(Equal(core.ResultCode_YIELD))
						Expect(result.Value).To(Equal(STR("test1")))
						Expect(result.Data).NotTo(BeNil())

						process.YieldBack(FALSE)
						result = process.Run()
						Expect(result.Code).To(Equal(core.ResultCode_YIELD))
						Expect(result.Value).To(Equal(STR("test2")))
						Expect(result.Data).NotTo(BeNil())

						process.YieldBack(TRUE)
						result = process.Run()
						Expect(result.Code).To(Equal(core.ResultCode_YIELD))
						Expect(result.Value).To(Equal(STR("body2")))
						Expect(result.Data).NotTo(BeNil())

						process.YieldBack(STR("result"))
						result = process.Run()
						Expect(result).To(Equal(OK(STR("result"))))
					})
					Specify("else", func() {
						result := process.Run()
						Expect(result.Code).To(Equal(core.ResultCode_YIELD))
						Expect(result.Value).To(Equal(STR("test1")))
						Expect(result.Data).NotTo(BeNil())

						process.YieldBack(FALSE)
						result = process.Run()
						Expect(result.Code).To(Equal(core.ResultCode_YIELD))
						Expect(result.Value).To(Equal(STR("test2")))
						Expect(result.Data).NotTo(BeNil())

						process.YieldBack(FALSE)
						result = process.Run()
						Expect(result.Code).To(Equal(core.ResultCode_YIELD))
						Expect(result.Value).To(Equal(STR("body3")))
						Expect(result.Data).NotTo(BeNil())

						process.YieldBack(STR("result"))
						result = process.Run()
						Expect(result).To(Equal(OK(STR("result"))))
					})
				})
			})
			Describe("`error`", func() {
				It("should interrupt tests with `ERROR` code", func() {
					Expect(execute("if {error msg; unreachable} {unreachable}")).To(Equal(
						ERROR("msg"),
					))
					Expect(
						execute("if false {} elseif {error msg; unreachable} {unreachable}"),
					).To(Equal(ERROR("msg")))
				})
				It("should interrupt bodies with `ERROR` code", func() {
					Expect(execute("if true {error msg; unreachable}")).To(Equal(
						ERROR("msg"),
					))
					Expect(
						execute("if false {} elseif true {error msg; unreachable}"),
					).To(Equal(ERROR("msg")))
					Expect(
						execute("if false {} elseif false {} else {error msg; unreachable}"),
					).To(Equal(ERROR("msg")))
				})
			})
			Describe("`break`", func() {
				It("should interrupt tests with `BREAK` code", func() {
					Expect(execute("if {break; unreachable} {unreachable}")).To(Equal(
						BREAK(NIL),
					))
					Expect(
						execute("if false {} elseif {break; unreachable} {unreachable}"),
					).To(Equal(BREAK(NIL)))
				})
				It("should interrupt bodies with `BREAK` code", func() {
					Expect(execute("if true {break; unreachable}")).To(Equal(BREAK(NIL)))
					Expect(
						execute("if false {} elseif true {break; unreachable}"),
					).To(Equal(BREAK(NIL)))
					Expect(
						execute("if false {} elseif false {} else {break; unreachable}"),
					).To(Equal(BREAK(NIL)))
				})
			})
			Describe("`continue`", func() {
				It("should interrupt tests with `CONTINUE` code", func() {
					Expect(execute("if {continue; unreachable} {unreachable}")).To(Equal(
						CONTINUE(NIL),
					))
					Expect(
						execute("if false {} elseif {continue; unreachable} {unreachable}"),
					).To(Equal(CONTINUE(NIL)))
				})
				It("should interrupt bodies with `CONTINUE` code", func() {
					Expect(execute("if true {continue; unreachable}")).To(Equal(CONTINUE(NIL)))
					Expect(
						execute("if false {} elseif true {continue; unreachable}"),
					).To(Equal(CONTINUE(NIL)))
					Expect(
						execute("if false {} elseif false {} else {continue; unreachable}"),
					).To(Equal(CONTINUE(NIL)))
				})
			})
		})
	})

	Describe("when", func() {
		Describe("Specifications", func() {
			Specify("usage", func() {
				Expect(evaluate("help when")).To(Equal(
					STR("when ?command? {?test body ...? ?default?}"),
				))
			})

			It("should return nil with empty test list", func() {
				Expect(evaluate("when {}")).To(Equal(NIL))
			})
			It("should accept tuple case list", func() {
				Expect(evaluate("when ()")).To(Equal(NIL))
			})
			It("should return the result of the first true body", func() {
				Expect(evaluate("when {true {1}}")).To(Equal(INT(1)))
				Expect(evaluate("when {true {1} {2}}")).To(Equal(INT(1)))
				Expect(evaluate("when {true {1} true {2} {3}}")).To(Equal(INT(1)))
				Expect(evaluate("when {false {1} true {2} true {3} {4}}")).To(Equal(
					INT(2),
				))
				Expect(evaluate("when {false {1} true {2} {3}}")).To(Equal(INT(2)))
				Expect(evaluate("when {false {1} true {2} true {3}  {4}}")).To(Equal(
					INT(2),
				))
			})
			It("should skip leading false bodies", func() {
				Expect(evaluate("when {false {unreachable}}")).To(Equal(NIL))
				Expect(
					evaluate("when {false {unreachable} false {unreachable}}"),
				).To(Equal(NIL))
				Expect(
					evaluate(
						"when {false {unreachable} false {unreachable} false {unreachable}}",
					),
				).To(Equal(NIL))
			})
			It("should skip trailing tests and bodies", func() {
				Expect(evaluate("when {true {1} {unreachable}}")).To(Equal(INT(1)))
				Expect(evaluate("when {true {1} {unreachable} {unreachable}}")).To(Equal(
					INT(1),
				))
				Expect(
					evaluate("when {true {1} {unreachable} {unreachable} {unreachable}}"),
				).To(Equal(INT(1)))
				Expect(
					evaluate("when {false {1} true {2} {unreachable} {unreachable}}"),
				).To(Equal(INT(2)))
			})
			Describe("no command", func() {
				It("should evaluate tests as boolean conditions", func() {
					Expect(evaluate("when {true {1}}")).To(Equal(INT(1)))
					Expect(evaluate("when {{idem true} {1}}")).To(Equal(INT(1)))
				})
			})
			Describe("literal command", func() {
				It("should apply to tests", func() {
					Expect(evaluate("when ! {true {1}}")).To(Equal(NIL))
					Expect(evaluate("when ! {true {1} {2}}")).To(Equal(INT(2)))
					Expect(evaluate("when ! {true {1} true {2} {3}}")).To(Equal(INT(3)))
				})
				It("should be called on each test", func() {
					evaluate("macro test {v} {set count [+ $count 1]; idem $v}")
					evaluate("set count 0")
					Expect(evaluate("when test {false {1} false {2} {3}}")).To(Equal(
						INT(3),
					))
					Expect(evaluate("get count")).To(Equal(INT(2)))
				})
				It("should pass test literal as argument", func() {
					Expect(evaluate("when ! {false {1} true {2} true {3} {4}}")).To(Equal(
						evaluate("when {{! false} {1} {! true} {2} {! true} {3} {4}}"),
					))
					Expect(evaluate("when ! {true {1} false {2} {3}}")).To(Equal(
						evaluate("when {{! true} {1} {! false} {2} {3}}"),
					))
				})
				It("should pass test tuple values as arguments", func() {
					Expect(evaluate("when 1 {(== 2) {1} (!= 1) {2} {3}}")).To(Equal(
						evaluate("when {{1 == 2} {1} {1 != 1} {2} {3}}"),
					))
					Expect(evaluate("when true {(? false) {1} () {2} {3}}")).To(Equal(
						evaluate("when true {(? false) {1} () {2} {3}}"),
					))
				})
			})
			Describe("tuple command", func() {
				It("should apply to tests", func() {
					Expect(evaluate("when (1 ==) {2 {1} 1 {2} {3}}")).To(Equal(INT(2)))
				})
				It("should be called on each test", func() {
					evaluate("macro test {cmd v} {set count [+ $count 1]; $cmd $v}")
					evaluate("set count 0")
					Expect(
						evaluate("when (test (true ?)) {false {1} false {2} {3}}"),
					).To(Equal(INT(3)))
					Expect(evaluate("get count")).To(Equal(INT(2)))
				})
				It("should pass test literal as argument", func() {
					Expect(evaluate("when (1 ==) {2 {1} 3 {2} 1 {3} {4}}")).To(Equal(
						INT(3),
					))
				})
				It("should pass test tuple values as arguments", func() {
					Expect(evaluate("when () {false {1} true {2} {3}}")).To(Equal(INT(2)))
					Expect(evaluate("when (1) {(== 2) {1} (!= 1) {2} {3}}")).To(Equal(
						INT(3),
					))
					Expect(
						evaluate("when (&& true) {(true false) {1} (true) {2} {3}}"),
					).To(Equal(INT(2)))
				})
			})
			Describe("script command", func() {
				It("evaluation result should apply to tests", func() {
					evaluate("macro test {v} {idem $v}")
					Expect(evaluate("when {idem test} {false {1} true {2} {3}}")).To(Equal(
						INT(2),
					))
				})
				It("should be called on each test", func() {
					evaluate("macro test {cmd} {set count [+ $count 1]; idem $cmd}")
					evaluate("set count 0")
					Expect(evaluate("when {test !} {true {1} true {2} {3}}")).To(Equal(
						INT(3),
					))
					Expect(evaluate("get count")).To(Equal(INT(2)))
				})
				It("should pass test literal as argument", func() {
					evaluate("macro test {v} {1 == $v}")
					Expect(evaluate("when {idem test} {2 {1} 3 {2} 1 {3} {4}}")).To(Equal(
						INT(3),
					))
				})
				It("should pass test tuple values as arguments", func() {
					evaluate("macro test {v1 v2} {$v1 == $v2}")
					Expect(evaluate("when {idem test} {(1 2) {1} (1 1) {2} {3}}")).To(Equal(
						INT(2),
					))
					Expect(evaluate("when {1} {(== 2) {1} (!= 1) {2} {3}}")).To(Equal(
						INT(3),
					))
					Expect(
						evaluate("when {idem (&& true)} {(true false) {1} (true) {2} {3}}"),
					).To(Equal(INT(2)))
				})
			})
		})

		Describe("Exceptions", func() {
			Specify("wrong arity", func() {
				Expect(execute("when")).To(Equal(
					ERROR(
						`wrong # args: should be "when ?command? {?test body ...? ?default?}"`,
					),
				))
				Expect(execute("when a b c")).To(Equal(
					ERROR(
						`wrong # args: should be "when ?command? {?test body ...? ?default?}"`,
					),
				))
				Expect(execute("help when a b c")).To(Equal(
					ERROR(
						`wrong # args: should be "when ?command? {?test body ...? ?default?}"`,
					),
				))
			})
			Specify("invalid case list", func() {
				Expect(execute("when a")).To(Equal(ERROR("invalid list")))
				Expect(execute("when []")).To(Equal(ERROR("invalid list")))
				Expect(execute("when {$a}")).To(Equal(ERROR("invalid list")))
			})
		})

		Describe("Control flow", func() {
			Describe("`return`", func() {
				It("should interrupt tests with `RETURN` code", func() {
					Expect(
						execute("when {{return val; unreachable} {unreachable}}"),
					).To(Equal(RETURN(STR("val"))))
					Expect(
						execute("when {false {} {return val; unreachable} {unreachable}}"),
					).To(Equal(RETURN(STR("val"))))
				})
				It("should interrupt script command with `RETURN` code", func() {
					Expect(
						execute("when {return val; unreachable} {true {unreachable}}"),
					).To(Equal(RETURN(STR("val"))))
					Expect(
						execute(
							"set count 0; when {if {$count == 1} {return val; unreachable} else {set count [+ $count 1]; idem idem}} {false {unreachable} true {unreachable} {unreachable}}",
						),
					).To(Equal(RETURN(STR("val"))))
				})
				It("should interrupt bodies with `RETURN` code", func() {
					Expect(execute("when {true {return val; unreachable}}")).To(Equal(
						RETURN(STR("val")),
					))
					Expect(
						execute("when {false {} true {return val; unreachable}}"),
					).To(Equal(RETURN(STR("val"))))
					Expect(
						execute("when {false {} false {} {return val; unreachable}}"),
					).To(Equal(RETURN(STR("val"))))
				})
			})
			Describe("`tailcall`", func() {
				It("should interrupt tests with `RETURN` code", func() {
					Expect(
						execute("when {{tailcall {idem val}; unreachable} {unreachable}}"),
					).To(Equal(RETURN(STR("val"))))
					Expect(
						execute(
							"when {false {} {tailcall {idem val}; unreachable} {unreachable}}",
						),
					).To(Equal(RETURN(STR("val"))))
				})
				It("should interrupt script command with `RETURN` code", func() {
					Expect(
						execute(
							"when {tailcall {idem val}; unreachable} {true {unreachable}}",
						),
					).To(Equal(RETURN(STR("val"))))
					Expect(
						execute(
							"set count 0; when {if {$count == 1} {tailcall {idem val}; unreachable} else {set count [+ $count 1]; idem idem}} {false {unreachable} true {unreachable} {unreachable}}",
						),
					).To(Equal(RETURN(STR("val"))))
				})
				It("should interrupt bodies with `RETURN` code", func() {
					Expect(
						execute("when {true {tailcall {idem val}; unreachable}}"),
					).To(Equal(RETURN(STR("val"))))
					Expect(
						execute("when {false {} true {tailcall {idem val}; unreachable}}"),
					).To(Equal(RETURN(STR("val"))))
					Expect(
						execute(
							"when {false {} false {} {tailcall {idem val}; unreachable}}",
						),
					).To(Equal(RETURN(STR("val"))))
				})
			})
			Describe("`yield`", func() {
				It("should interrupt tests with `YIELD` code", func() {
					Expect(
						execute("when {{yield; unreachable} {unreachable}}").Code,
					).To(Equal(core.ResultCode_YIELD))
					Expect(
						execute("when {false {} {yield; unreachable} {unreachable}}").Code,
					).To(Equal(core.ResultCode_YIELD))
				})
				It("should interrupt script commands with YIELD code", func() {
					Expect(
						execute("when {yield; unreachable} {true {unreachable}}").Code,
					).To(Equal(core.ResultCode_YIELD))
					Expect(
						execute(
							"set count 0; when {if {$count == 1} {yield; unreachable} else {set count [+ $count 1]; idem idem}} {false {unreachable} true {unreachable} {unreachable}}",
						).Code,
					).To(Equal(core.ResultCode_YIELD))
				})
				It("should interrupt bodies with `YIELD` code", func() {
					Expect(execute("when {true {yield; unreachable}}").Code).To(Equal(
						core.ResultCode_YIELD,
					))
					Expect(
						execute("when {false {} true {yield; unreachable}}").Code,
					).To(Equal(core.ResultCode_YIELD))
					Expect(
						execute("when {false {} false {} {yield; unreachable}}").Code,
					).To(Equal(core.ResultCode_YIELD))
				})
				Describe("should provide a resumable state", func() {
					Describe("no command", func() {
						var process *Process
						BeforeEach(func() {
							process = rootScope.PrepareScript(
								*parse(
									"when {{yield test1} {yield body1} {yield test2} {yield body2} {yield body3}}",
								),
							)
						})
						Specify("first", func() {
							result := process.Run()
							Expect(result.Code).To(Equal(core.ResultCode_YIELD))
							Expect(result.Value).To(Equal(STR("test1")))
							Expect(result.Data).NotTo(BeNil())

							process.YieldBack(TRUE)
							result = process.Run()
							Expect(result.Code).To(Equal(core.ResultCode_YIELD))
							Expect(result.Value).To(Equal(STR("body1")))
							Expect(result.Data).NotTo(BeNil())

							process.YieldBack(STR("result"))
							result = process.Run()
							Expect(result).To(Equal(OK(STR("result"))))
						})
						Specify("second", func() {
							result := process.Run()
							Expect(result.Code).To(Equal(core.ResultCode_YIELD))
							Expect(result.Value).To(Equal(STR("test1")))
							Expect(result.Data).NotTo(BeNil())

							process.YieldBack(FALSE)
							result = process.Run()
							Expect(result.Code).To(Equal(core.ResultCode_YIELD))
							Expect(result.Value).To(Equal(STR("test2")))
							Expect(result.Data).NotTo(BeNil())

							process.YieldBack(TRUE)
							result = process.Run()
							Expect(result.Code).To(Equal(core.ResultCode_YIELD))
							Expect(result.Value).To(Equal(STR("body2")))
							Expect(result.Data).NotTo(BeNil())

							process.YieldBack(STR("result"))
							result = process.Run()
							Expect(result).To(Equal(OK(STR("result"))))
						})
						Specify("default", func() {
							result := process.Run()
							Expect(result.Code).To(Equal(core.ResultCode_YIELD))
							Expect(result.Value).To(Equal(STR("test1")))
							Expect(result.Data).NotTo(BeNil())

							process.YieldBack(FALSE)
							result = process.Run()
							Expect(result.Code).To(Equal(core.ResultCode_YIELD))
							Expect(result.Value).To(Equal(STR("test2")))
							Expect(result.Data).NotTo(BeNil())

							process.YieldBack(FALSE)
							result = process.Run()
							Expect(result.Code).To(Equal(core.ResultCode_YIELD))
							Expect(result.Value).To(Equal(STR("body3")))
							Expect(result.Data).NotTo(BeNil())

							process.YieldBack(STR("result"))
							result = process.Run()
							Expect(result).To(Equal(OK(STR("result"))))
						})
					})
					Describe("script command", func() {
						var process *Process
						BeforeEach(func() {
							evaluate("macro test {v} {yield $v}")
							process = rootScope.PrepareScript(
								*parse(
									"when {yield command} {test1 {yield body1} test2 {yield body2} {yield body3}}",
								),
							)
						})
						Specify("first", func() {
							result := process.Run()
							Expect(result.Code).To(Equal(core.ResultCode_YIELD))
							Expect(result.Value).To(Equal(STR("command")))
							Expect(result.Data).NotTo(BeNil())

							process.YieldBack(STR("test"))
							result = process.Run()
							Expect(result.Code).To(Equal(core.ResultCode_YIELD))
							Expect(result.Value).To(Equal(STR("test1")))
							Expect(result.Data).NotTo(BeNil())

							process.YieldBack(TRUE)
							result = process.Run()
							Expect(result.Code).To(Equal(core.ResultCode_YIELD))
							Expect(result.Value).To(Equal(STR("body1")))
							Expect(result.Data).NotTo(BeNil())

							process.YieldBack(STR("result"))
							result = process.Run()
							Expect(result).To(Equal(OK(STR("result"))))
						})
						Specify("second", func() {
							result := process.Run()
							Expect(result.Code).To(Equal(core.ResultCode_YIELD))
							Expect(result.Value).To(Equal(STR("command")))
							Expect(result.Data).NotTo(BeNil())

							process.YieldBack(STR("test"))
							result = process.Run()
							Expect(result.Code).To(Equal(core.ResultCode_YIELD))
							Expect(result.Value).To(Equal(STR("test1")))
							Expect(result.Data).NotTo(BeNil())

							process.YieldBack(FALSE)
							result = process.Run()
							Expect(result.Code).To(Equal(core.ResultCode_YIELD))
							Expect(result.Value).To(Equal(STR("command")))
							Expect(result.Data).NotTo(BeNil())

							process.YieldBack(STR("test"))
							result = process.Run()
							Expect(result.Code).To(Equal(core.ResultCode_YIELD))
							Expect(result.Value).To(Equal(STR("test2")))
							Expect(result.Data).NotTo(BeNil())

							process.YieldBack(TRUE)
							result = process.Run()
							Expect(result.Code).To(Equal(core.ResultCode_YIELD))
							Expect(result.Value).To(Equal(STR("body2")))
							Expect(result.Data).NotTo(BeNil())

							process.YieldBack(STR("result"))
							result = process.Run()
							Expect(result).To(Equal(OK(STR("result"))))
						})
						Specify("default", func() {
							result := process.Run()
							Expect(result.Code).To(Equal(core.ResultCode_YIELD))
							Expect(result.Value).To(Equal(STR("command")))
							Expect(result.Data).NotTo(BeNil())

							process.YieldBack(STR("test"))
							result = process.Run()
							Expect(result.Code).To(Equal(core.ResultCode_YIELD))
							Expect(result.Value).To(Equal(STR("test1")))
							Expect(result.Data).NotTo(BeNil())

							process.YieldBack(FALSE)
							result = process.Run()
							Expect(result.Code).To(Equal(core.ResultCode_YIELD))
							Expect(result.Value).To(Equal(STR("command")))
							Expect(result.Data).NotTo(BeNil())

							process.YieldBack(STR("test"))
							result = process.Run()
							Expect(result.Code).To(Equal(core.ResultCode_YIELD))
							Expect(result.Value).To(Equal(STR("test2")))
							Expect(result.Data).NotTo(BeNil())

							process.YieldBack(FALSE)
							result = process.Run()
							Expect(result.Code).To(Equal(core.ResultCode_YIELD))
							Expect(result.Value).To(Equal(STR("body3")))
							Expect(result.Data).NotTo(BeNil())

							process.YieldBack(STR("result"))
							result = process.Run()
							Expect(result).To(Equal(OK(STR("result"))))
						})
					})
				})
			})
			Describe("`error`", func() {
				It("should interrupt tests with `ERROR` code", func() {
					Expect(
						execute("when {{error msg; unreachable} {unreachable}}"),
					).To(Equal(ERROR("msg")))
					Expect(
						execute("when {false {} {error msg; unreachable} {unreachable}}"),
					).To(Equal(ERROR("msg")))
				})
				It("should interrupt script command with `ERROR` code", func() {
					Expect(
						execute("when {error msg; unreachable} {true {unreachable}}"),
					).To(Equal(ERROR("msg")))
					Expect(
						execute(
							"set count 0; when {if {$count == 1} {error msg; unreachable} else {set count [+ $count 1]; idem idem}} {false {unreachable} true {unreachable} {unreachable}}",
						),
					).To(Equal(ERROR("msg")))
				})
				It("should interrupt bodies with `ERROR` code", func() {
					Expect(execute("when {true {error msg; unreachable}}")).To(Equal(
						ERROR("msg"),
					))
					Expect(
						execute("when {false {} true {error msg; unreachable}}"),
					).To(Equal(ERROR("msg")))
					Expect(
						execute("when {false {} false {} {error msg; unreachable}}"),
					).To(Equal(ERROR("msg")))
				})
			})
			Describe("`break`", func() {
				It("should interrupt tests with `BREAK` code", func() {
					Expect(execute("when {{break; unreachable} {unreachable}}")).To(Equal(
						BREAK(NIL),
					))
					Expect(
						execute("when {false {} {break; unreachable} {unreachable}}"),
					).To(Equal(BREAK(NIL)))
				})
				It("should interrupt script command with `BREAK` code", func() {
					Expect(
						execute("when {break; unreachable} {true {unreachable}}"),
					).To(Equal(BREAK(NIL)))
					Expect(
						execute(
							"set count 0; when {if {$count == 1} {break; unreachable} else {set count [+ $count 1]; idem idem}} {false {unreachable} true {unreachable} {unreachable}}",
						),
					).To(Equal(BREAK(NIL)))
				})
				It("should interrupt bodies with `BREAK` code", func() {
					Expect(execute("when {true {break; unreachable}}")).To(Equal(BREAK(NIL)))
					Expect(execute("when {false {} true {break; unreachable}}")).To(Equal(
						BREAK(NIL),
					))
					Expect(
						execute("when {false {} false {} {break; unreachable}}"),
					).To(Equal(BREAK(NIL)))
				})
			})
			Describe("`continue`", func() {
				It("should interrupt tests with `CONTINUE` code", func() {
					Expect(
						execute("when {{continue; unreachable} {unreachable}}"),
					).To(Equal(CONTINUE(NIL)))
					Expect(
						execute("when {false {} {continue; unreachable} {unreachable}}"),
					).To(Equal(CONTINUE(NIL)))
				})
				It("should interrupt script command with `BREAK` code", func() {
					Expect(
						execute("when {continue; unreachable} {true {unreachable}}"),
					).To(Equal(CONTINUE(NIL)))
					Expect(
						execute(
							"set count 0; when {if {$count == 1} {continue; unreachable} else {set count [+ $count 1]; idem idem}} {false {unreachable} true {unreachable} {unreachable}}",
						),
					).To(Equal(CONTINUE(NIL)))
				})
				It("should interrupt bodies with `CONTINUE` code", func() {
					Expect(execute("when {true {continue; unreachable}}")).To(Equal(
						CONTINUE(NIL),
					))
					Expect(
						execute("when {false {} true {continue; unreachable}}"),
					).To(Equal(CONTINUE(NIL)))
					Expect(
						execute("when {false {} false {} {continue; unreachable}}"),
					).To(Equal(CONTINUE(NIL)))
				})
			})
		})
	})

	Describe("catch", func() {
		Describe("Specifications", func() {
			Specify("usage", func() {
				Expect(evaluate("help catch")).To(Equal(
					STR(
						"catch body ?return value handler? ?yield value handler? ?error message handler? ?break handler? ?continue handler? ?finally handler?",
					),
				))
			})
		})

		Describe("without handler", func() {
			Specify("`OK` code should return `(ok value)` tuple", func() {
				Expect(execute("catch {}")).To(Equal(execute("tuple (ok [])")))
				Expect(execute("catch {idem value}")).To(Equal(
					execute("tuple (ok value)"),
				))
			})
			Specify("`RETURN` code should return `(return value)` tuple", func() {
				Expect(execute("catch {return}")).To(Equal(execute("tuple (return [])")))
				Expect(execute("catch {return value}")).To(Equal(
					execute("tuple (return value)"),
				))
			})
			Specify("`YIELD` code should return `(yield value)` tuple", func() {
				Expect(execute("catch {yield}")).To(Equal(execute("tuple (yield [])")))
				Expect(execute("catch {yield}")).To(Equal(execute("tuple (yield [])")))
				Expect(execute("catch {yield value}")).To(Equal(
					execute("tuple (yield value)"),
				))
			})
			Specify("`ERROR` code should return `(error message)` tuple", func() {
				Expect(execute("catch {error value}")).To(Equal(
					execute("tuple (error value)"),
				))
				Expect(execute("catch {error value}")).To(Equal(
					execute("tuple (error value)"),
				))
			})
			Specify("`BREAK` code should return `(break)` tuple", func() {
				Expect(execute("catch {break}")).To(Equal(execute("tuple (break)")))
			})
			Specify("`CONTINUE` code should return `(continue)` tuple", func() {
				Expect(execute("catch {continue}")).To(Equal(execute("tuple (continue)")))
			})
			Specify("arbitrary errors", func() {
				Expect(execute("catch {idem}")).To(Equal(
					execute(`tuple (error "wrong # args: should be \"idem value\"")`),
				))
				Expect(execute("catch {get var}")).To(Equal(
					execute(`tuple (error "cannot get \"var\": no such variable")`),
				))
				Expect(execute("catch {cmd a b}")).To(Equal(
					execute(`tuple (error "cannot resolve command \"cmd\"")`),
				))
			})
		})

		Describe("`return` handler", func() {
			It("should catch `RETURN` code", func() {
				evaluate("catch {return} return res {set var handler}")
				Expect(evaluate("get var")).To(Equal(STR("handler")))
			})
			It("should let other codes pass through", func() {
				Expect(execute("catch {idem value} return res {unreachable}")).To(Equal(
					OK(STR("value")),
				))
				Expect(execute("catch {yield value} return res {unreachable}")).To(Equal(
					YIELD(STR("value")),
				))
				Expect(
					execute("catch {error message} return res {unreachable}"),
				).To(Equal(ERROR("message")))
				Expect(execute("catch {break} return res {unreachable}")).To(Equal(
					BREAK(NIL),
				))
				Expect(execute("catch {continue} return res {unreachable}")).To(Equal(
					CONTINUE(NIL),
				))
			})
			It("should return handler result", func() {
				Expect(evaluate("catch {return} return res {idem handler}")).To(Equal(
					STR("handler"),
				))
			})
			Specify("handler value should be handler-local", func() {
				Expect(evaluate("catch {return value} return res {idem _$res}")).To(Equal(
					STR("_value"),
				))
				Expect(evaluate("exists res")).To(Equal(FALSE))
			})
			Describe("Control flow", func() {
				Describe("`return`", func() {
					It("should interrupt handler with `RETURN` code", func() {
						Expect(
							execute(
								"catch {return val} return res {return handler; unreachable}",
							),
						).To(Equal(RETURN(STR("handler"))))
					})
					It("should bypass `finally` handler", func() {
						Expect(
							execute(
								"catch {return val} return res {return handler; unreachable} finally {unreachable}",
							),
						).To(Equal(RETURN(STR("handler"))))
					})
				})
				Describe("`tailcall`", func() {
					It("should interrupt handler with `RETURN` code", func() {
						Expect(
							execute(
								"catch {return val} return res {tailcall {idem handler}; unreachable}",
							),
						).To(Equal(RETURN(STR("handler"))))
					})
					It("should bypass `finally` handler", func() {
						Expect(
							execute(
								"catch {return val} return res {tailcall {idem handler}; unreachable} finally {unreachable}",
							),
						).To(Equal(RETURN(STR("handler"))))
					})
				})
				Describe("`yield`", func() {
					It("should interrupt handler with `YIELD` code", func() {
						Expect(
							execute("catch {return val} return res {yield; unreachable}").Code,
						).To(Equal(core.ResultCode_YIELD))
					})
					It("should provide a resumable state", func() {
						process := rootScope.PrepareScript(
							*parse("catch {return val} return res {idem _$[yield handler]}"),
						)

						result := process.Run()
						Expect(result.Code).To(Equal(core.ResultCode_YIELD))
						Expect(result.Value).To(Equal(STR("handler")))
						Expect(result.Data).NotTo(BeNil())

						process.YieldBack(STR("value"))
						result = process.Run()
						Expect(result).To(Equal(OK(STR("_value"))))
					})
					It("should not bypass `finally` handler", func() {
						process := rootScope.PrepareScript(
							*parse(
								"catch {return val} return res {yield; idem handler} finally {set var finally}",
							),
						)

						_ = process.Run()
						result := process.Run()
						Expect(result).To(Equal(OK(STR("handler"))))
						Expect(evaluate("get var")).To(Equal(STR("finally")))
					})
				})
				Describe("`error`", func() {
					It("should interrupt handler with `ERROR` code", func() {
						Expect(
							execute(
								"catch {return val} return res {error message; unreachable}",
							),
						).To(Equal(ERROR("message")))
					})
					It("should bypass `finally` handler", func() {
						Expect(
							execute(
								"catch {return val} return res {error message; unreachable} finally {unreachable}",
							),
						).To(Equal(ERROR("message")))
					})
				})
				Describe("`break`", func() {
					It("should interrupt handler with `BREAK` code", func() {
						Expect(
							execute("catch {return val} return res {break; unreachable}"),
						).To(Equal(BREAK(NIL)))
					})
					It("should bypass `finally` handler", func() {
						Expect(
							execute(
								"catch {return val} return res {break; unreachable} finally {unreachable}",
							),
						).To(Equal(BREAK(NIL)))
					})
				})
				Describe("`continue`", func() {
					It("should interrupt handler with `CONTINUE` code", func() {
						Expect(
							execute("catch {return val} return res {continue; unreachable}"),
						).To(Equal(CONTINUE(NIL)))
					})
					It("should bypass `finally` handler", func() {
						Expect(
							execute(
								"catch {return val} return res {continue; unreachable} finally {unreachable}",
							),
						).To(Equal(CONTINUE(NIL)))
					})
				})
			})

			Describe("Exceptions", func() {
				Specify("wrong arity", func() {
					Expect(execute("catch {} return")).To(Equal(
						ERROR("wrong #args: missing return handler parameter"),
					))
					Expect(execute("catch {} return a")).To(Equal(
						ERROR("wrong #args: missing return handler body"),
					))
				})
				Specify("invalid parameter name", func() {
					Expect(execute("catch {} return [] {}")).To(Equal(
						ERROR("invalid return handler parameter name"),
					))
				})
			})
		})

		Describe("`yield` handler", func() {
			It("should catch `YIELD` code", func() {
				evaluate("catch {yield} yield res {set var handler}")
				Expect(evaluate("get var")).To(Equal(STR("handler")))
			})
			It("should let other codes pass through", func() {
				Expect(execute("catch {idem value} yield res {unreachable}")).To(Equal(
					OK(STR("value")),
				))
				Expect(execute("catch {return value} yield res {unreachable}")).To(Equal(
					RETURN(STR("value")),
				))
				Expect(execute("catch {error message} yield res {unreachable}")).To(Equal(
					ERROR("message"),
				))
				Expect(execute("catch {break} yield res {unreachable}")).To(Equal(
					BREAK(NIL),
				))
				Expect(execute("catch {continue} yield res {unreachable}")).To(Equal(
					CONTINUE(NIL),
				))
			})
			It("should return handler result", func() {
				Expect(evaluate("catch {yield} yield res {idem handler}")).To(Equal(
					STR("handler"),
				))
			})
			Specify("handler value should be handler-local", func() {
				Expect(evaluate("catch {yield value} yield res {idem _$res}")).To(Equal(
					STR("_value"),
				))
				Expect(evaluate("exists res")).To(Equal(FALSE))
			})

			Describe("Control flow", func() {
				Describe("`return`", func() {
					It("should interrupt handler with `RETURN` code", func() {
						Expect(
							execute(
								"catch {yield val} yield res {return handler; unreachable}",
							),
						).To(Equal(RETURN(STR("handler"))))
					})
					It("should bypass `finally` handler", func() {
						Expect(
							execute(
								"catch {yield val} yield res {return handler; unreachable} finally {unreachable}",
							),
						).To(Equal(RETURN(STR("handler"))))
					})
				})
				Describe("`tailcall`", func() {
					It("should interrupt handler with `RETURN` code", func() {
						Expect(
							execute(
								"catch {yield val} yield res {tailcall {idem handler}; unreachable}",
							),
						).To(Equal(RETURN(STR("handler"))))
					})
					It("should bypass `finally` handler", func() {
						Expect(
							execute(
								"catch {yield val} yield res {tailcall {idem handler}; unreachable} finally {unreachable}",
							),
						).To(Equal(RETURN(STR("handler"))))
					})
				})
				Describe("`yield`", func() {
					It("should interrupt handler with `YIELD` code", func() {
						Expect(
							execute("catch {yield val} yield res {yield; unreachable}").Code,
						).To(Equal(core.ResultCode_YIELD))
					})
					It("should provide a resumable state", func() {
						process := rootScope.PrepareScript(
							*parse("catch {yield val} yield res {idem _$[yield handler]}"),
						)

						result := process.Run()
						Expect(result.Code).To(Equal(core.ResultCode_YIELD))
						Expect(result.Value).To(Equal(STR("handler")))
						Expect(result.Data).NotTo(BeNil())

						process.YieldBack(STR("value"))
						result = process.Run()
						Expect(result).To(Equal(OK(STR("_value"))))
					})
					It("should not bypass `finally` handler", func() {
						process := rootScope.PrepareScript(
							*parse(
								"catch {yield val} yield res {yield; idem handler} finally {set var finally}",
							),
						)

						_ = process.Run()
						result := process.Run()
						Expect(result).To(Equal(OK(STR("handler"))))
						Expect(evaluate("get var")).To(Equal(STR("finally")))
					})
				})
				Describe("`error`", func() {
					It("should interrupt handler with `ERROR` code", func() {
						Expect(
							execute(
								"catch {yield val} yield res {error message; unreachable}",
							),
						).To(Equal(ERROR("message")))
					})
					It("should bypass `finally` handler", func() {
						Expect(
							execute(
								"catch {yield val} yield res {error message; unreachable} finally {unreachable}",
							),
						).To(Equal(ERROR("message")))
					})
				})
				Describe("`break`", func() {
					It("should interrupt handler with `BREAK` code", func() {
						Expect(
							execute("catch {yield val} yield res {break; unreachable}"),
						).To(Equal(BREAK(NIL)))
					})
					It("should bypass `finally` handler", func() {
						Expect(
							execute(
								"catch {yield val} yield res {break; unreachable} finally {unreachable}",
							),
						).To(Equal(BREAK(NIL)))
					})
				})
				Describe("`continue`", func() {
					It("should interrupt handler with `CONTINUE` code", func() {
						Expect(
							execute("catch {yield val} yield res {continue; unreachable}"),
						).To(Equal(CONTINUE(NIL)))
					})
					It("should bypass `finally` handler", func() {
						Expect(
							execute(
								"catch {yield val} yield res {continue; unreachable} finally {unreachable}",
							),
						).To(Equal(CONTINUE(NIL)))
					})
				})
			})

			Describe("Exceptions", func() {
				Specify("wrong arity", func() {
					Expect(execute("catch {} yield")).To(Equal(
						ERROR("wrong #args: missing yield handler parameter"),
					))
					Expect(execute("catch {} yield a")).To(Equal(
						ERROR("wrong #args: missing yield handler body"),
					))
				})
				Specify("invalid parameter name", func() {
					Expect(execute("catch {} yield [] {}")).To(Equal(
						ERROR("invalid yield handler parameter name"),
					))
				})
			})
		})

		Describe("`error` handler", func() {
			It("should catch `ERROR` code", func() {
				evaluate("catch {error message} error msg {set var handler}")
				Expect(evaluate("get var")).To(Equal(STR("handler")))
			})
			It("should let other codes pass through", func() {
				Expect(execute("catch {idem value} error msg {unreachable}")).To(Equal(
					OK(STR("value")),
				))
				Expect(execute("catch {return value} error msg {unreachable}")).To(Equal(
					RETURN(STR("value")),
				))
				Expect(execute("catch {yield value} error msg {unreachable}")).To(Equal(
					YIELD(STR("value")),
				))
				Expect(execute("catch {break} error msg {unreachable}")).To(Equal(
					BREAK(NIL),
				))
				Expect(execute("catch {continue} error msg {unreachable}")).To(Equal(
					CONTINUE(NIL),
				))
			})
			It("should return handler result", func() {
				Expect(
					evaluate("catch {error message} error msg {idem handler}"),
				).To(Equal(STR("handler")))
			})
			Specify("handler value should be handler-local", func() {
				Expect(evaluate("catch {error message} error msg {idem _$msg}")).To(Equal(
					STR("_message"),
				))
				Expect(evaluate("exists msg")).To(Equal(FALSE))
			})
			Describe("Control flow", func() {
				Describe("`return`", func() {
					It("should interrupt handler with `RETURN` code", func() {
						Expect(
							execute(
								"catch {error message} error msg {return handler; unreachable}",
							),
						).To(Equal(RETURN(STR("handler"))))
					})
					It("should bypass `finally` handler", func() {
						Expect(
							execute(
								"catch {error message} error msg {return handler; unreachable} finally {unreachable}",
							),
						).To(Equal(RETURN(STR("handler"))))
					})
				})
				Describe("`tailcall`", func() {
					It("should interrupt handler with `RETURN` code", func() {
						Expect(
							execute(
								"catch {error message} error msg {tailcall {idem handler}; unreachable}",
							),
						).To(Equal(RETURN(STR("handler"))))
					})
					It("should bypass `finally` handler", func() {
						Expect(
							execute(
								"catch {error message} error msg {tailcall {idem handler}; unreachable} finally {unreachable}",
							),
						).To(Equal(RETURN(STR("handler"))))
					})
				})
				Describe("`yield`", func() {
					It("should interrupt handler with `YIELD` code", func() {
						Expect(
							execute("catch {error message} error msg {yield; unreachable}").Code,
						).To(Equal(core.ResultCode_YIELD))
					})
					It("should provide a resumable state", func() {
						process := rootScope.PrepareScript(
							*parse("catch {error message} error msg {idem _$[yield handler]}"),
						)

						result := process.Run()
						Expect(result.Code).To(Equal(core.ResultCode_YIELD))
						Expect(result.Value).To(Equal(STR("handler")))
						Expect(result.Data).NotTo(BeNil())

						process.YieldBack(STR("value"))
						result = process.Run()
						Expect(result).To(Equal(OK(STR("_value"))))
					})
					It("should not bypass `finally` handler", func() {
						process := rootScope.PrepareScript(
							*parse(
								"catch {error message} error msg {yield; idem handler} finally {set var finally}",
							),
						)

						_ = process.Run()
						result := process.Run()
						Expect(result).To(Equal(OK(STR("handler"))))
						Expect(evaluate("get var")).To(Equal(STR("finally")))
					})
				})
				Describe("`error`", func() {
					It("should interrupt handler with `ERROR` code", func() {
						Expect(
							execute(
								"catch {error message} error msg {error message; unreachable}",
							),
						).To(Equal(ERROR("message")))
					})
					It("should bypass `finally` handler", func() {
						Expect(
							execute(
								"catch {error message} error msg {error message; unreachable} finally {unreachable}",
							),
						).To(Equal(ERROR("message")))
					})
				})
				Describe("`break`", func() {
					It("should interrupt handler with `BREAK` code", func() {
						Expect(
							execute("catch {error message} error msg {break; unreachable}"),
						).To(Equal(BREAK(NIL)))
					})
					It("should bypass `finally` handler", func() {
						Expect(
							execute(
								"catch {error message} error msg {break; unreachable} finally {unreachable}",
							),
						).To(Equal(BREAK(NIL)))
					})
				})
				Describe("`continue`", func() {
					It("should interrupt handler with `CONTINUE` code", func() {
						Expect(
							execute("catch {error message} error msg {continue; unreachable}"),
						).To(Equal(CONTINUE(NIL)))
					})
					It("should bypass `finally` handler", func() {
						Expect(
							execute(
								"catch {error message} error msg {continue; unreachable} finally {unreachable}",
							),
						).To(Equal(CONTINUE(NIL)))
					})
				})
			})

			Describe("Exceptions", func() {
				Specify("wrong arity", func() {
					Expect(execute("catch {} error")).To(Equal(
						ERROR("wrong #args: missing error handler parameter"),
					))
					Expect(execute("catch {} error a")).To(Equal(
						ERROR("wrong #args: missing error handler body"),
					))
				})
				Specify("invalid parameter name", func() {
					Expect(execute("catch {} error [] {}")).To(Equal(
						ERROR("invalid error handler parameter name"),
					))
				})
			})
		})

		Describe("`break` handler", func() {
			It("should catch `BREAK` code", func() {
				evaluate("catch {break} break {set var handler}")
				Expect(evaluate("get var")).To(Equal(STR("handler")))
			})
			It("should let other codes pass through", func() {
				Expect(execute("catch {idem value} break {unreachable}")).To(Equal(
					OK(STR("value")),
				))
				Expect(execute("catch {return value} break {unreachable}")).To(Equal(
					RETURN(STR("value")),
				))
				Expect(execute("catch {yield value} break {unreachable}")).To(Equal(
					YIELD(STR("value")),
				))
				Expect(execute("catch {error message} break {unreachable}")).To(Equal(
					ERROR("message"),
				))
				Expect(execute("catch {continue} break {unreachable}")).To(Equal(
					CONTINUE(NIL),
				))
			})
			It("should return handler result", func() {
				Expect(evaluate("catch {break} break {idem handler}")).To(Equal(
					STR("handler"),
				))
			})
			Describe("Control flow", func() {
				Describe("`return`", func() {
					It("should interrupt handler with `RETURN` code", func() {
						Expect(
							execute("catch {break} break {return handler; unreachable}"),
						).To(Equal(RETURN(STR("handler"))))
					})
					It("should bypass `finally` handler", func() {
						Expect(
							execute(
								"catch {break} break {return handler; unreachable} finally {unreachable}",
							),
						).To(Equal(RETURN(STR("handler"))))
					})
				})
				Describe("`tailcall`", func() {
					It("should interrupt handler with `RETURN` code", func() {
						Expect(
							execute(
								"catch {break} break {tailcall {idem handler}; unreachable}",
							),
						).To(Equal(RETURN(STR("handler"))))
					})
					It("should bypass `finally` handler", func() {
						Expect(
							execute(
								"catch {break} break {tailcall {idem handler}; unreachable} finally {unreachable}",
							),
						).To(Equal(RETURN(STR("handler"))))
					})
				})
				Describe("`yield`", func() {
					It("should interrupt handler with `YIELD` code", func() {
						Expect(
							execute("catch {break} break {yield; unreachable}").Code,
						).To(Equal(core.ResultCode_YIELD))
					})
					It("should provide a resumable state", func() {
						process := rootScope.PrepareScript(
							*parse("catch {break} break {idem _$[yield handler]}"),
						)

						result := process.Run()
						Expect(result.Code).To(Equal(core.ResultCode_YIELD))
						Expect(result.Value).To(Equal(STR("handler")))
						Expect(result.Data).NotTo(BeNil())

						process.YieldBack(STR("value"))
						result = process.Run()
						Expect(result).To(Equal(OK(STR("_value"))))
					})
					It("should not bypass `finally` handler", func() {
						process := rootScope.PrepareScript(
							*parse(
								"catch {break} break {yield; idem handler} finally {set var finally}",
							),
						)

						_ = process.Run()
						result := process.Run()
						Expect(result).To(Equal(OK(STR("handler"))))
						Expect(evaluate("get var")).To(Equal(STR("finally")))
					})
				})
				Describe("`error`", func() {
					It("should interrupt handler with `ERROR` code", func() {
						Expect(
							execute("catch {break} break {error message; unreachable}"),
						).To(Equal(ERROR("message")))
					})
					It("should bypass `finally` handler", func() {
						Expect(
							execute(
								"catch {break} break {error message; unreachable} finally {unreachable}",
							),
						).To(Equal(ERROR("message")))
					})
				})
				Describe("`break`", func() {
					It("should interrupt handler with `BREAK` code", func() {
						Expect(execute("catch {break} break {break; unreachable}")).To(Equal(
							BREAK(NIL),
						))
					})
					It("should bypass `finally` handler", func() {
						Expect(
							execute(
								"catch {break} break {break; unreachable} finally {unreachable}",
							),
						).To(Equal(BREAK(NIL)))
					})
				})
				Describe("`continue`", func() {
					It("should interrupt handler with `CONTINUE` code", func() {
						Expect(
							execute("catch {break} break {continue; unreachable}"),
						).To(Equal(CONTINUE(NIL)))
					})
					It("should bypass `finally` handler", func() {
						Expect(
							execute(
								"catch {break} break {continue; unreachable} finally {unreachable}",
							),
						).To(Equal(CONTINUE(NIL)))
					})
				})
			})

			Describe("Exceptions", func() {
				Specify("wrong arity", func() {
					Expect(execute("catch {} break")).To(Equal(
						ERROR("wrong #args: missing break handler body"),
					))
				})
			})
		})

		Describe("`continue` handler", func() {
			It("should catch `CONTINUE` code", func() {
				evaluate("catch {continue} continue {set var handler}")
				Expect(evaluate("get var")).To(Equal(STR("handler")))
			})
			It("should let other codes pass through", func() {
				Expect(execute("catch {idem value} continue {unreachable}")).To(Equal(
					OK(STR("value")),
				))
				Expect(execute("catch {return value} continue {unreachable}")).To(Equal(
					RETURN(STR("value")),
				))
				Expect(execute("catch {yield value} continue {unreachable}")).To(Equal(
					YIELD(STR("value")),
				))
				Expect(execute("catch {error message} continue {unreachable}")).To(Equal(
					ERROR("message"),
				))
				Expect(execute("catch {break} continue {unreachable}")).To(Equal(BREAK(NIL)))
			})
			It("should return handler result", func() {
				Expect(evaluate("catch {continue} continue {idem handler}")).To(Equal(
					STR("handler"),
				))
			})
			Describe("Control flow", func() {
				Describe("`return`", func() {
					It("should interrupt handler with `RETURN` code", func() {
						Expect(
							execute("catch {continue} continue {return handler; unreachable}"),
						).To(Equal(RETURN(STR("handler"))))
					})
					It("should bypass `finally` handler", func() {
						Expect(
							execute(
								"catch {continue} continue {return handler; unreachable} finally {unreachable}",
							),
						).To(Equal(RETURN(STR("handler"))))
					})
				})
				Describe("`tailcall`", func() {
					It("should interrupt handler with `RETURN` code", func() {
						Expect(
							execute(
								"catch {continue} continue {tailcall {idem handler}; unreachable}",
							),
						).To(Equal(RETURN(STR("handler"))))
					})
					It("should bypass `finally` handler", func() {
						Expect(
							execute(
								"catch {continue} continue {tailcall {idem handler}; unreachable} finally {unreachable}",
							),
						).To(Equal(RETURN(STR("handler"))))
					})
				})
				Describe("`yield`", func() {
					It("should interrupt handler with `YIELD` code", func() {
						Expect(
							execute("catch {continue} continue {yield; unreachable}").Code,
						).To(Equal(core.ResultCode_YIELD))
					})
					It("should provide a resumable state", func() {
						process := rootScope.PrepareScript(
							*parse("catch {continue} continue {idem _$[yield handler]}"),
						)

						result := process.Run()
						Expect(result.Code).To(Equal(core.ResultCode_YIELD))
						Expect(result.Value).To(Equal(STR("handler")))
						Expect(result.Data).NotTo(BeNil())

						process.YieldBack(STR("value"))
						result = process.Run()
						Expect(result).To(Equal(OK(STR("_value"))))
					})
					It("should not bypass `finally` handler", func() {
						process := rootScope.PrepareScript(
							*parse(
								"catch {continue} continue {yield; idem handler} finally {set var finally}",
							),
						)

						_ = process.Run()
						result := process.Run()
						Expect(result).To(Equal(OK(STR("handler"))))
						Expect(evaluate("get var")).To(Equal(STR("finally")))
					})
				})
				Describe("`error`", func() {
					It("should interrupt handler with `ERROR` code", func() {
						Expect(
							execute("catch {continue} continue {error message; unreachable}"),
						).To(Equal(ERROR("message")))
					})
					It("should bypass `finally` handler", func() {
						Expect(
							execute(
								"catch {continue} continue {error message; unreachable} finally {unreachable}",
							),
						).To(Equal(ERROR("message")))
					})
				})
				Describe("`break`", func() {
					It("should interrupt handler with `BREAK` code", func() {
						Expect(
							execute("catch {continue} continue {break; unreachable}"),
						).To(Equal(BREAK(NIL)))
					})
					It("should bypass `finally` handler", func() {
						Expect(
							execute(
								"catch {continue} continue {break; unreachable} finally {unreachable}",
							),
						).To(Equal(BREAK(NIL)))
					})
				})
				Describe("`continue`", func() {
					It("should interrupt handler with `CONTINUE` code", func() {
						Expect(
							execute("catch {continue} continue {continue; unreachable}"),
						).To(Equal(CONTINUE(NIL)))
					})
					It("should bypass `finally` handler", func() {
						Expect(
							execute(
								"catch {continue} continue {continue; unreachable} finally {unreachable}",
							),
						).To(Equal(CONTINUE(NIL)))
					})
				})
			})

			Describe("Exceptions", func() {
				Specify("wrong arity", func() {
					Expect(execute("catch {} continue")).To(Equal(
						ERROR("wrong #args: missing continue handler body"),
					))
				})
			})
		})

		Describe("`finally` handler", func() {
			It("should execute for `OK` code", func() {
				evaluate("catch {idem value} finally {set var handler}")
				Expect(evaluate("get var")).To(Equal(STR("handler")))
			})
			It("should execute for `RETURN` code", func() {
				evaluate("catch {return} finally {set var handler}")
				Expect(evaluate("get var")).To(Equal(STR("handler")))
			})
			It("should execute for `YIELD` code", func() {
				evaluate("catch {yield} finally {set var handler}")
				Expect(evaluate("get var")).To(Equal(STR("handler")))
			})
			It("should execute for `ERROR` code", func() {
				evaluate("catch {error message} finally {set var handler}")
				Expect(evaluate("get var")).To(Equal(STR("handler")))
			})
			It("should execute for `BREAK` code", func() {
				evaluate("catch {break} finally {set var handler}")
				Expect(evaluate("get var")).To(Equal(STR("handler")))
			})
			It("should execute for `CONTINUE` code", func() {
				evaluate("catch {continue} finally {set var handler}")
				Expect(evaluate("get var")).To(Equal(STR("handler")))
			})
			It("should let all codes pass through", func() {
				Expect(execute("catch {idem value} finally {idem handler}")).To(Equal(
					OK(STR("value")),
				))
				Expect(execute("catch {return value} finally {idem handler}")).To(Equal(
					RETURN(STR("value")),
				))
				Expect(execute("catch {yield value} finally {idem handler}")).To(Equal(
					YIELD(STR("value")),
				))
				Expect(execute("catch {error message} finally {idem handler}")).To(Equal(
					ERROR("message"),
				))
				Expect(execute("catch {break} finally {idem handler}")).To(Equal(BREAK(NIL)))
				Expect(execute("catch {continue} finally {idem handler}")).To(Equal(
					CONTINUE(NIL),
				))
			})

			Describe("Control flow", func() {
				Describe("`return`", func() {
					It("should interrupt handler with `RETURN` code", func() {
						Expect(
							execute(
								"catch {error message} finally {return handler; unreachable}",
							),
						).To(Equal(RETURN(STR("handler"))))
					})
				})
				Describe("`tailcall`", func() {
					It("should interrupt handler with `RETURN` code", func() {
						Expect(
							execute(
								"catch {error message} finally {tailcall {idem handler}; unreachable}",
							),
						).To(Equal(RETURN(STR("handler"))))
					})
				})
				Describe("`yield`", func() {
					It("should interrupt handler with `YIELD` code", func() {
						Expect(
							execute("catch {error message} finally {yield; unreachable}").Code,
						).To(Equal(core.ResultCode_YIELD))
					})
					It("should provide a resumable state", func() {
						process := rootScope.PrepareScript(
							*parse("catch {error message} finally {idem _$[yield handler]}"),
						)

						result := process.Run()
						Expect(result.Code).To(Equal(core.ResultCode_YIELD))
						Expect(result.Value).To(Equal(STR("handler")))
						Expect(result.Data).NotTo(BeNil())

						process.YieldBack(STR("value"))
						result = process.Run()
						Expect(result).To(Equal(ERROR("message")))
					})
				})
				Describe("`error`", func() {
					It("should interrupt handler with `ERROR` code", func() {
						Expect(
							execute(
								"catch {error message} finally {error message; unreachable}",
							),
						).To(Equal(ERROR("message")))
					})
				})
				Describe("`break`", func() {
					It("should interrupt handler with `BREAK` code", func() {
						Expect(
							execute("catch {error message} finally {break; unreachable}"),
						).To(Equal(BREAK(NIL)))
					})
				})
				Describe("`continue`", func() {
					It("should interrupt handler with `CONTINUE` code", func() {
						Expect(
							execute("catch {error message} finally {continue; unreachable}"),
						).To(Equal(CONTINUE(NIL)))
					})
				})
			})

			Describe("Exceptions", func() {
				Specify("wrong arity", func() {
					Expect(execute("catch {} finally")).To(Equal(
						ERROR("wrong #args: missing finally handler body"),
					))
				})
			})
		})

		Describe("Exceptions", func() {
			Specify("wrong arity", func() {
				Expect(execute("catch")).To(Equal(
					ERROR(
						`wrong # args: should be "catch body ?return value handler? ?yield value handler? ?error message handler? ?break handler? ?continue handler? ?finally handler?"`,
					),
				))
			})
			Specify("non-script body", func() {
				Expect(execute("catch a")).To(Equal(ERROR("body must be a script")))
				Expect(execute("catch []")).To(Equal(ERROR("body must be a script")))
				Expect(execute("catch [1]")).To(Equal(ERROR("body must be a script")))
			})
			Specify("invalid keyword", func() {
				Expect(execute("catch {} foo {}")).To(Equal(
					ERROR(`invalid keyword "foo"`),
				))
				Expect(execute("catch {} [] {}")).To(Equal(ERROR("invalid keyword")))
			})
		})
	})

	Describe("pass", func() {
		Describe("Specifications", func() {
			Specify("usage", func() {
				Expect(evaluate("help pass")).To(Equal(STR("pass")))
			})

			Specify("result code should be the custom code `pass`", func() {
				Expect(core.RESULT_CODE_NAME(execute("pass"))).To(Equal("pass"))
			})
			Specify("`catch` should return `(pass)` tuple", func() {
				Expect(execute("catch {pass}")).To(Equal(execute("tuple (pass)")))
			})
			Specify("`catch` handlers should not handle it", func() {
				Expect(
					core.RESULT_CODE_NAME(
						execute(`
							catch {pass} \
								return value {unreachable} \
								yield value {unreachable} \
								error message {unreachable} \
								break {unreachable} \
								continue {unreachable} \
						`),
					),
				).To(Equal("pass"))
			})
			Describe("should interrupt `catch` handlers and let original result pass through", func() {
				Specify("`RETURN`", func() {
					Expect(
						execute("catch {return value} return res {pass; unreachable}"),
					).To(Equal(RETURN(STR("value"))))
				})
				Specify("`YIELD`", func() {
					result := execute(
						"catch {yield value} yield res {pass; unreachable}",
					)
					Expect(result.Code).To(Equal(core.ResultCode_YIELD))
					Expect(result.Value).To(Equal(STR("value")))
				})
				Specify("`ERROR`", func() {
					Expect(
						execute("catch {error message} error msg {pass; unreachable}"),
					).To(Equal(ERROR("message")))
				})
				Specify("`BREAK`", func() {
					Expect(execute("catch {break} break {pass; unreachable}")).To(Equal(
						BREAK(NIL),
					))
				})
				Specify("`CATCH`", func() {
					Expect(
						execute("catch {continue} continue {pass; unreachable}"),
					).To(Equal(CONTINUE(NIL)))
				})
			})
			Describe("should let `catch` `finally` handler execute", func() {
				Specify("`RETURN`", func() {
					Expect(
						execute(
							"catch {return value} return res {pass} finally {set var handler}",
						),
					).To(Equal(RETURN(STR("value"))))
					Expect(evaluate("get var")).To(Equal(STR("handler")))
				})
				Specify("`YIELD`", func() {
					process := rootScope.PrepareScript(
						*parse(
							"catch {yield value} yield res {pass} finally {set var handler}",
						),
					)

					result := process.Run()
					Expect(result.Code).To(Equal(core.ResultCode_YIELD))
					Expect(result.Value).To(Equal(STR("value")))

					process.Run()
					Expect(evaluate("get var")).To(Equal(STR("handler")))
				})
				Specify("`ERROR`", func() {
					Expect(
						execute(
							"catch {error message} error msg {pass} finally {set var handler}",
						),
					).To(Equal(ERROR("message")))
					Expect(evaluate("get var")).To(Equal(STR("handler")))
				})
				Specify("`BREAK`", func() {
					Expect(
						execute("catch {break} break {pass} finally {set var handler}"),
					).To(Equal(BREAK(NIL)))
					Expect(evaluate("get var")).To(Equal(STR("handler")))
				})
				Specify("`CONTINUE`", func() {
					Expect(
						execute(
							"catch {continue} continue {pass} finally {set var handler}",
						),
					).To(Equal(CONTINUE(NIL)))
					Expect(evaluate("get var")).To(Equal(STR("handler")))
				})
			})
			It("should resume yielded body", func() {
				process := rootScope.PrepareScript(
					*parse(
						"catch {set var [yield step1]; idem _$[yield step2]} yield res {pass}",
					),
				)

				result := process.Run()
				Expect(result.Code).To(Equal(core.ResultCode_YIELD))
				Expect(result.Value).To(Equal(STR("step1")))
				Expect(result.Data).NotTo(BeNil())

				process.YieldBack(STR("value1"))
				result = process.Run()
				Expect(result.Code).To(Equal(core.ResultCode_YIELD))
				Expect(result.Value).To(Equal(STR("step2")))
				Expect(result.Data).NotTo(BeNil())
				Expect(evaluate("get var")).To(Equal(STR("value1")))

				process.YieldBack(STR("value2"))
				result = process.Run()
				Expect(result).To(Equal(OK(STR("_value2"))))
			})
		})

		Describe("Exceptions", func() {
			Specify("wrong arity", func() {
				Expect(execute("pass a")).To(Equal(
					ERROR(`wrong # args: should be "pass"`),
				))
				Expect(execute("help pass a")).To(Equal(
					ERROR(`wrong # args: should be "pass"`),
				))
			})
			Specify("invalid `pass` handler", func() {
				Expect(execute("catch {pass} pass {}")).To(Equal(
					ERROR(`invalid keyword "pass"`),
				))
			})
		})
	})
})
