package helena_dialect_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"helena/core"
	. "helena/helena_dialect"
)

var _ = Describe("Helena constants and variables", func() {
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

	Describe("let", func() {
		Describe("Specifications", func() {
			Specify("usage", func() {
				Expect(evaluate("help let")).To(Equal(STR("let constname value")))
				Expect(evaluate("help let name")).To(Equal(STR("let constname value")))
				Expect(evaluate("help let name val")).To(Equal(
					STR("let constname value"),
				))
			})

			It("should define the value of a new constant", func() {
				evaluate("let cst val")
				Expect(rootScope.Context.Constants["cst"]).To(Equal(STR("val")))
			})
			It("should return the constant value", func() {
				Expect(evaluate("let cst val")).To(Equal(STR("val")))
			})

			Describe("Tuple destructuring", func() {
				It("should define several constants at once", func() {
					Expect(execute("let (var1 var2 var3) (val1 val2 val3)")).To(Equal(
						execute("idem (val1 val2 val3)"),
					))
					Expect(evaluate("get var1")).To(Equal(STR("val1")))
					Expect(evaluate("get var2")).To(Equal(STR("val2")))
					Expect(evaluate("get var3")).To(Equal(STR("val3")))
				})
				It("should set duplicate constants to their last value", func() {
					Expect(execute("let (var1 var2 var1) (val1 val2 val3)")).To(Equal(
						execute("idem (val1 val2 val3)"),
					))
					Expect(evaluate("get var1")).To(Equal(STR("val3")))
					Expect(evaluate("get var2")).To(Equal(STR("val2")))
				})
				It("should work recursively", func() {
					Expect(execute("let (var1 (var2 var3)) (val1 (val2 val3))")).To(Equal(
						execute("idem (val1 (val2 val3))"),
					))
					Expect(evaluate("get var1")).To(Equal(STR("val1")))
					Expect(evaluate("get var2")).To(Equal(STR("val2")))
					Expect(evaluate("get var3")).To(Equal(STR("val3")))
				})
				It("should support setting a constant to a tuple value", func() {
					Expect(execute("let (var1 var2) (val1 (val2 val3))")).To(Equal(
						execute("idem (val1 (val2 val3))"),
					))
					Expect(evaluate("get var1")).To(Equal(STR("val1")))
					Expect(evaluate("get var2")).To(Equal(evaluate("idem (val2 val3)")))
				})
				It("should not define constants in case of missing value", func() {
					Expect(execute("let (var1 var2) (val1)")).To(Equal(
						ERROR("bad value shape"),
					))
					Expect(rootScope.Context.Variables["var1"]).To(BeNil())
					Expect(rootScope.Context.Variables["var2"]).To(BeNil())
				})
				It("should not define constants in case of missing subvalue", func() {
					Expect(execute("let (var1 (var2 var3)) (val1 (val2))")).To(Equal(
						ERROR("bad value shape"),
					))
					Expect(rootScope.Context.Constants["var1"]).To(BeNil())
					Expect(rootScope.Context.Constants["var2"]).To(BeNil())
					Expect(rootScope.Context.Constants["var3"]).To(BeNil())
				})
				It("should not define constants in case of bad shape", func() {
					Expect(execute("let (var1 (var2)) (val1 val2)")).To(Equal(
						ERROR("bad value shape"),
					))
					Expect(rootScope.Context.Constants["var1"]).To(BeNil())
					Expect(rootScope.Context.Constants["var2"]).To(BeNil())
				})
			})
		})

		Describe("Exceptions", func() {
			Specify("wrong arity", func() {
				Expect(execute("let")).To(Equal(
					ERROR(`wrong # args: should be "let constname value"`),
				))
				Expect(execute("let a")).To(Equal(
					ERROR(`wrong # args: should be "let constname value"`),
				))
				Expect(execute("let a b c")).To(Equal(
					ERROR(`wrong # args: should be "let constname value"`),
				))
				Expect(execute("help let a b c")).To(Equal(
					ERROR(`wrong # args: should be "let constname value"`),
				))
			})
			Specify("invalid `constname`", func() {
				Expect(execute("let [] val")).To(Equal(ERROR("invalid constant name")))
				Expect(execute("let ([]) (val)")).To(Equal(
					ERROR("invalid constant name"),
				))
			})
			Specify("bad `constname` tuple shape", func() {
				Expect(execute("let (a) b")).To(Equal(ERROR("bad value shape")))
				Expect(execute("let ((a)) (b)")).To(Equal(ERROR("bad value shape")))
				Expect(execute("let (a) ()")).To(Equal(ERROR("bad value shape")))
			})
			Specify("existing constant", func() {
				rootScope.Context.Constants["cst"] = STR("old")
				Expect(execute("let cst val")).To(Equal(
					ERROR(`cannot redefine constant "cst"`),
				))
			})
			Specify("existing variable", func() {
				rootScope.Context.Variables["var"] = STR("old")
				Expect(execute("let var val")).To(Equal(
					ERROR(`cannot define constant "var": variable already exists`),
				))
			})
		})
	})

	Describe("set", func() {
		Describe("Specifications", func() {
			Specify("usage", func() {
				Expect(evaluate("help set")).To(Equal(STR("set varname value")))
				Expect(evaluate("help set name")).To(Equal(STR("set varname value")))
				Expect(evaluate("help set name val")).To(Equal(STR("set varname value")))
			})

			It("should set the value of a new variable", func() {
				evaluate("set var val")
				Expect(rootScope.Context.Variables["var"]).To(Equal(STR("val")))
			})
			It("should redefine the value of an existing variable", func() {
				rootScope.Context.Variables["var"] = STR("old")
				evaluate("set var val")
				Expect(rootScope.Context.Variables["var"]).To(Equal(STR("val")))
			})
			It("should return the set value", func() {
				Expect(evaluate("set var val")).To(Equal(STR("val")))
			})

			Describe("Tuple destructuring", func() {
				It("should set several variables at once", func() {
					Expect(execute("set (var1 var2 var3) (val1 val2 val3)")).To(Equal(
						execute("idem (val1 val2 val3)"),
					))
					Expect(evaluate("get var1")).To(Equal(STR("val1")))
					Expect(evaluate("get var2")).To(Equal(STR("val2")))
					Expect(evaluate("get var3")).To(Equal(STR("val3")))
				})
				It("should set duplicate values to their last value", func() {
					Expect(execute("set (var1 var2 var1) (val1 val2 val3)")).To(Equal(
						execute("idem (val1 val2 val3)"),
					))
					Expect(evaluate("get var1")).To(Equal(STR("val3")))
					Expect(evaluate("get var2")).To(Equal(STR("val2")))
				})
				It("should work recursively", func() {
					Expect(execute("set (var1 (var2 var3)) (val1 (val2 val3))")).To(Equal(
						execute("idem (val1 (val2 val3))"),
					))
					Expect(evaluate("get var1")).To(Equal(STR("val1")))
					Expect(evaluate("get var2")).To(Equal(STR("val2")))
					Expect(evaluate("get var3")).To(Equal(STR("val3")))
				})
				It("should support setting a variable to a tuple value", func() {
					Expect(execute("set (var1 var2) (val1 (val2 val3))")).To(Equal(
						execute("idem (val1 (val2 val3))"),
					))
					Expect(evaluate("get var1")).To(Equal(STR("val1")))
					Expect(evaluate("get var2")).To(Equal(evaluate("idem (val2 val3)")))
				})
				It("should not set variables in case of missing value", func() {
					Expect(execute("set (var1 var2) (val1)")).To(Equal(
						ERROR("bad value shape"),
					))
					Expect(rootScope.Context.Variables["var1"]).To(BeNil())
					Expect(rootScope.Context.Variables["var2"]).To(BeNil())
				})
				It("should not set variables in case of missing subvalue", func() {
					Expect(execute("set (var1 (var2 var3)) (val1 (val2))")).To(Equal(
						ERROR("bad value shape"),
					))
					Expect(rootScope.Context.Variables["var1"]).To(BeNil())
					Expect(rootScope.Context.Variables["var2"]).To(BeNil())
					Expect(rootScope.Context.Variables["var3"]).To(BeNil())
				})
				It("should not set variables in case of bad shape", func() {
					Expect(execute("set (var1 (var2)) (val1 val2)")).To(Equal(
						ERROR("bad value shape"),
					))
					Expect(rootScope.Context.Variables["var1"]).To(BeNil())
					Expect(rootScope.Context.Variables["var2"]).To(BeNil())
				})
			})
		})

		Describe("Exceptions", func() {
			Specify("wrong arity", func() {
				Expect(execute("set")).To(Equal(
					ERROR(`wrong # args: should be "set varname value"`),
				))
				Expect(execute("set a")).To(Equal(
					ERROR(`wrong # args: should be "set varname value"`),
				))
				Expect(execute("set a b c")).To(Equal(
					ERROR(`wrong # args: should be "set varname value"`),
				))
				Expect(execute("help set a b c")).To(Equal(
					ERROR(`wrong # args: should be "set varname value"`),
				))
			})
			Specify("invalid `varname`", func() {
				Expect(execute("set [] val")).To(Equal(ERROR("invalid variable name")))
				Expect(execute("set ([]) (val)")).To(Equal(
					ERROR("invalid variable name"),
				))
			})
			Specify("bad `varname` tuple shape", func() {
				Expect(execute("set (a) b")).To(Equal(ERROR("bad value shape")))
				Expect(execute("set ((a)) (b)")).To(Equal(ERROR("bad value shape")))
				Expect(execute("set (a) ()")).To(Equal(ERROR("bad value shape")))
			})
			Specify("existing constant", func() {
				rootScope.Context.Constants["cst"] = STR("old")
				Expect(execute("set cst val")).To(Equal(
					ERROR(`cannot redefine constant "cst"`),
				))
			})
		})
	})

	Describe("get", func() {
		Describe("Specifications", func() {
			Specify("usage", func() {
				Expect(evaluate("help get")).To(Equal(STR("get varname ?default?")))
				Expect(evaluate("help get name")).To(Equal(STR("get varname ?default?")))
				Expect(evaluate("help get name val")).To(Equal(
					STR("get varname ?default?"),
				))
			})

			It("should return the value of an existing variable", func() {
				evaluate("let cst val")
				Expect(evaluate("get cst")).To(Equal(STR("val")))
			})
			It("should return the value of an existing constant", func() {
				evaluate("set var val")
				Expect(evaluate("get var")).To(Equal(STR("val")))
			})
			It("should return the default value for a unknown variable", func() {
				Expect(evaluate("get var default")).To(Equal(STR("default")))
				Expect(evaluate("get var(key) default")).To(Equal(STR("default")))
				Expect(evaluate("get var[1] default")).To(Equal(STR("default")))
			})

			Describe("Qualified names", func() {
				Specify("indexed selector", func() {
					rootScope.SetNamedVariable("var", LIST([]core.Value{STR("val1"), STR("val2")}))
					Expect(evaluate("get var[1]")).To(Equal(STR("val2")))
				})
				Specify("keyed selector", func() {
					rootScope.SetNamedVariable("var", DICT(map[string]core.Value{"key": STR("val")}))
					Expect(evaluate("get var(key)")).To(Equal(STR("val")))
				})
				Specify("should work recursively", func() {
					rootScope.SetNamedVariable(
						"var",
						DICT(map[string]core.Value{"key": LIST([]core.Value{STR("val1"), STR("val2")})}),
					)
					Expect(evaluate("get var(key)[1]")).To(Equal(STR("val2")))
				})
				It("should return the default value when a selector fails", func() {
					rootScope.SetNamedConstant("l", LIST([]core.Value{}))
					rootScope.SetNamedConstant("m", DICT(map[string]core.Value{}))
					Expect(evaluate("get l[1] default")).To(Equal(STR("default")))
					Expect(evaluate("get l(key) default")).To(Equal(STR("default")))
					Expect(evaluate("get l[0](key) default")).To(Equal(STR("default")))
					Expect(evaluate("get m[1] default")).To(Equal(STR("default")))
					Expect(evaluate("get m(key) default")).To(Equal(STR("default")))
				})
			})

			Describe("Tuple destructuring", func() {
				It("should get several variables at once", func() {
					evaluate("set var1 val1")
					evaluate("set var2 val2")
					evaluate("set var3 val3")
					Expect(execute("get (var1 var2 var3)")).To(Equal(
						execute("idem (val1 val2 val3)"),
					))
				})
				It("should work recursively", func() {
					evaluate("set var1 val1")
					evaluate("set var2 val2")
					evaluate("set var3 val3")
					Expect(execute("get (var1 (var2 var3))")).To(Equal(
						execute("idem (val1 (val2 val3))"),
					))
				})
				It("should support qualified names", func() {
					rootScope.SetNamedVariable("var1", LIST([]core.Value{STR("val1"), STR("val2")}))
					rootScope.SetNamedVariable("var2", LIST([]core.Value{STR("val3"), STR("val4")}))
					rootScope.SetNamedVariable("var3", LIST([]core.Value{STR("val5"), STR("val6")}))
					Expect(evaluate("get (var1 (var2 var3))[1]")).To(Equal(
						evaluate("idem (val2 (val4 val6))"),
					))
					Expect(evaluate("get (var1[1])")).To(Equal(evaluate("idem (val2)")))
				})
			})
		})

		Describe("Exceptions", func() {
			Specify("wrong arity", func() {
				Expect(execute("get")).To(Equal(
					ERROR(`wrong # args: should be "get varname ?default?"`),
				))
				Expect(execute("get a b c")).To(Equal(
					ERROR(`wrong # args: should be "get varname ?default?"`),
				))
				Expect(execute("help get a b c")).To(Equal(
					ERROR(`wrong # args: should be "get varname ?default?"`),
				))
			})
			Specify("tuple `varname` with default", func() {
				Expect(execute("get (var) default")).To(Equal(
					ERROR("cannot use default with name tuples"),
				))
			})
			Specify("unknown variable", func() {
				Expect(execute("get unknownVariable")).To(Equal(
					ERROR(`cannot get "unknownVariable": no such variable`),
				))
			})
			Specify("bad selector", func() {
				rootScope.SetNamedConstant("l", LIST([]core.Value{}))
				rootScope.SetNamedConstant("m", DICT(map[string]core.Value{}))
				Expect(execute("get l[1]").Code).To(Equal(core.ResultCode_ERROR))
				Expect(execute("get l(key)").Code).To(Equal(core.ResultCode_ERROR))
				Expect(execute("get m[1]").Code).To(Equal(core.ResultCode_ERROR))
				Expect(execute("get m(key)").Code).To(Equal(core.ResultCode_ERROR))
			})
		})
	})

	Describe("exists", func() {
		Describe("Specifications", func() {
			Specify("usage", func() {
				Expect(evaluate("help exists")).To(Equal(STR("exists varname")))
				Expect(evaluate("help exists name")).To(Equal(STR("exists varname")))
			})

			It("should return `true` for an existing variable", func() {
				evaluate("let cst val")
				Expect(evaluate("exists cst")).To(Equal(TRUE))
			})
			It("should return `true` for an existing constant", func() {
				evaluate("set var val")
				Expect(evaluate("exists var")).To(Equal(TRUE))
			})
			It("should return `false` for a unknown variable", func() {
				Expect(evaluate("exists var")).To(Equal(FALSE))
				Expect(evaluate("exists var(key)")).To(Equal(FALSE))
				Expect(evaluate("exists var[1]")).To(Equal(FALSE))
			})

			Describe("Qualified names", func() {
				Specify("indexed selector", func() {
					rootScope.SetNamedVariable("var", LIST([]core.Value{STR("val1"), STR("val2")}))
					Expect(evaluate("exists var[1]")).To(Equal(TRUE))
				})
				Specify("keyed selector", func() {
					rootScope.SetNamedVariable("var", DICT(map[string]core.Value{"key": STR("val")}))
					Expect(evaluate("exists var(key)")).To(Equal(TRUE))
				})
				Specify("recursive selectors", func() {
					rootScope.SetNamedVariable(
						"var",
						DICT(map[string]core.Value{"key": LIST([]core.Value{STR("val1"), STR("val2")})}),
					)
					Expect(evaluate("exists var(key)[1]")).To(Equal(TRUE))
				})
				It("should return `false` for a unknown variable", func() {
					Expect(evaluate("exists var[1]")).To(Equal(FALSE))
					Expect(evaluate("exists var(key)")).To(Equal(FALSE))
				})
				It("should return `false` when a selector fails", func() {
					rootScope.SetNamedConstant("l", LIST([]core.Value{}))
					rootScope.SetNamedConstant("m", DICT(map[string]core.Value{}))
					Expect(evaluate("exists l[1]")).To(Equal(FALSE))
					Expect(evaluate("exists l(key)")).To(Equal(FALSE))
					Expect(evaluate("exists m[1]")).To(Equal(FALSE))
					Expect(evaluate("exists m(key)")).To(Equal(FALSE))
				})
			})
		})

		Describe("Exceptions", func() {
			Specify("wrong arity", func() {
				Expect(execute("exists")).To(Equal(
					ERROR(`wrong # args: should be "exists varname"`),
				))
				Expect(execute("exists a b")).To(Equal(
					ERROR(`wrong # args: should be "exists varname"`),
				))
				Expect(execute("help exists a b")).To(Equal(
					ERROR(`wrong # args: should be "exists varname"`),
				))
			})
			Specify("tuple `varname`", func() {
				Expect(execute("exists (var)")).To(Equal(ERROR("invalid value")))
			})
		})
	})

	Describe("unset", func() {
		Describe("Specifications", func() {
			Specify("usage", func() {
				Expect(evaluate("help unset")).To(Equal(STR("unset varname")))
				Expect(evaluate("help unset name")).To(Equal(STR("unset varname")))
			})

			It("should unset an existing variable", func() {
				evaluate("set var val")
				Expect(evaluate("exists var")).To(Equal(TRUE))
				evaluate("unset var")
				Expect(evaluate("exists var")).To(Equal(FALSE))
			})
			It("should return nil", func() {
				evaluate("set var val")
				Expect(evaluate("unset var")).To(Equal(NIL))
			})

			Describe("Tuples", func() {
				It("should unset several variables at once", func() {
					evaluate("set var1 val1")
					evaluate("set var2 val2")
					evaluate("set var3 val3")
					Expect(rootScope.Context.Variables["var1"]).NotTo(BeNil())
					Expect(rootScope.Context.Variables["var2"]).NotTo(BeNil())
					Expect(rootScope.Context.Variables["var3"]).NotTo(BeNil())
					Expect(evaluate("unset (var1 var2 var3)")).To(Equal(NIL))
					Expect(rootScope.Context.Variables["var1"]).To(BeNil())
					Expect(rootScope.Context.Variables["var2"]).To(BeNil())
					Expect(rootScope.Context.Variables["var3"]).To(BeNil())
				})
				It("should work recursively", func() {
					evaluate("set var1 val1")
					evaluate("set var2 val2")
					evaluate("set var3 val3")
					Expect(rootScope.Context.Variables["var1"]).NotTo(BeNil())
					Expect(rootScope.Context.Variables["var2"]).NotTo(BeNil())
					Expect(rootScope.Context.Variables["var3"]).NotTo(BeNil())
					Expect(evaluate("unset (var1 (var2 var3))")).To(Equal(NIL))
					Expect(rootScope.Context.Variables["var1"]).To(BeNil())
					Expect(rootScope.Context.Variables["var2"]).To(BeNil())
					Expect(rootScope.Context.Variables["var3"]).To(BeNil())
				})
				It("should not unset variables in case the name tuple contains unknown variables", func() {
					evaluate("set var1 val1")
					evaluate("set var2 val2")
					Expect(rootScope.Context.Variables["var1"]).NotTo(BeNil())
					Expect(rootScope.Context.Variables["var2"]).NotTo(BeNil())
					Expect(execute("unset (var1 (var2 var3))").Code).To(Equal(
						core.ResultCode_ERROR,
					))
					Expect(rootScope.Context.Variables["var1"]).NotTo(BeNil())
					Expect(rootScope.Context.Variables["var2"]).NotTo(BeNil())
				})
				It("should not unset variables in case the name tuple contains qualified names", func() {
					rootScope.SetNamedVariable("var", LIST([]core.Value{STR("val1"), STR("val2")}))
					Expect(rootScope.Context.Variables["var"]).NotTo(BeNil())
					Expect(execute("unset (var[1])").Code).To(Equal(core.ResultCode_ERROR))
					Expect(rootScope.Context.Variables["var"]).NotTo(BeNil())
				})
				It("should not unset variables in case the name tuple contains invalid variables", func() {
					evaluate("set var1 val1")
					evaluate("set var2 val2")
					Expect(rootScope.Context.Variables["var1"]).NotTo(BeNil())
					Expect(rootScope.Context.Variables["var2"]).NotTo(BeNil())
					Expect(execute("unset (var1 (var2 []))").Code).To(Equal(
						core.ResultCode_ERROR,
					))
					Expect(rootScope.Context.Variables["var1"]).NotTo(BeNil())
					Expect(rootScope.Context.Variables["var2"]).NotTo(BeNil())
				})
			})
		})

		Describe("Exceptions", func() {
			Specify("wrong arity", func() {
				Expect(execute("unset")).To(Equal(
					ERROR(`wrong # args: should be "unset varname"`),
				))
				Expect(execute("unset a b")).To(Equal(
					ERROR(`wrong # args: should be "unset varname"`),
				))
				Expect(execute("help unset a b")).To(Equal(
					ERROR(`wrong # args: should be "unset varname"`),
				))
			})
			Specify("invalid `varname`", func() {
				Expect(execute("unset []")).To(Equal(ERROR("invalid variable name")))
			})
			Specify("qualified `varname`", func() {
				rootScope.SetNamedVariable("var", LIST([]core.Value{STR("val1"), STR("val2")}))
				rootScope.SetNamedVariable("var", DICT(map[string]core.Value{"key": STR("val")}))
				Expect(execute("unset var[1]")).To(Equal(ERROR("invalid variable name")))
				Expect(execute("unset var(key)")).To(Equal(
					ERROR("invalid variable name"),
				))
				Expect(execute("unset (var[1] var(key))")).To(Equal(
					ERROR("invalid variable name"),
				))
			})
			Specify("existing constant", func() {
				rootScope.Context.Constants["cst"] = STR("old")
				Expect(execute("unset cst")).To(Equal(
					ERROR(`cannot unset constant "cst"`),
				))
			})
			Specify("unknown variable", func() {
				Expect(execute("unset unknownVariable")).To(Equal(
					ERROR(`cannot unset "unknownVariable": no such variable`),
				))
			})
		})
	})
})
