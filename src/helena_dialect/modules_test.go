package helena_dialect_test

import (
	"path"
	"path/filepath"
	"runtime"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"helena/core"
	. "helena/helena_dialect"
)

var _ = Describe("Helena modules", func() {
	var rootScope *Scope
	var moduleRegistry *ModuleRegistry
	var dirname string

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
		moduleRegistry = NewModuleRegistry()
		_, filename, _, _ := runtime.Caller(0)
		dirname = filepath.Dir(filename)
		InitCommandsForModule(rootScope, moduleRegistry, dirname)

		tokenizer = core.Tokenizer{}
		parser = &core.Parser{}
	}

	BeforeEach(init)

	Describe("module", func() {
		Describe("Specifications", func() {
			Specify("usage", func() {
				Expect(evaluate("help module")).To(Equal(STR("module ?name? body")))
				Expect(evaluate("help module {}")).To(Equal(STR("module ?name? body")))
				Expect(evaluate("help module cmd {}")).To(Equal(
					STR("module ?name? body"),
				))
			})

			It("should define a new command", func() {
				evaluate("module cmd {}")
				Expect(rootScope.Context.Commands["cmd"]).NotTo(BeNil())
			})
			It("should replace existing commands", func() {
				evaluate("module cmd {}")
				Expect(execute("module cmd {}").Code).To(Equal(core.ResultCode_OK))
			})
			It("should return a command object", func() {
				Expect(evaluate("module {}").Type()).To(Equal(core.ValueType_COMMAND))
				Expect(evaluate("module cmd  {}").Type()).To(Equal(core.ValueType_COMMAND))
			})
			Specify("the named command should return its command object", func() {
				value := evaluate("module cmd {}")
				Expect(evaluate("cmd")).To(Equal(value))
			})
			Specify("the command object should return itself", func() {
				value := evaluate("set cmd [module {}]")
				Expect(evaluate("$cmd")).To(Equal(value))
			})
		})

		Describe("Exceptions", func() {
			Specify("wrong arity", func() {
				Expect(execute("module")).To(Equal(
					ERROR(`wrong # args: should be "module ?name? body"`),
				))
				Expect(execute("module a b c")).To(Equal(
					ERROR(`wrong # args: should be "module ?name? body"`),
				))
				Expect(execute("help module a b c")).To(Equal(
					ERROR(`wrong # args: should be "module ?name? body"`),
				))
			})
			Specify("invalid `name`", func() {
				Expect(execute("module [] {}")).To(Equal(ERROR("invalid command name")))
			})
			Specify("non-script body", func() {
				Expect(execute("module a")).To(Equal(ERROR("body must be a script")))
				Expect(execute("module a b")).To(Equal(ERROR("body must be a script")))
			})
		})

		Describe("`body`", func() {
			It("should be executed", func() {
				Expect(execute("module {macro cmd {} {error message}; cmd}")).To(Equal(
					ERROR("message"),
				))
			})
			It("should not access outer commands", func() {
				evaluate("closure cmd {} {unreachable}")
				Expect(execute("module {cmd}")).To(Equal(
					ERROR(`cannot resolve command "cmd"`),
				))
			})
			It("should not define outer commands", func() {
				evaluate("closure cmd {} {idem outer}")
				evaluate("module {closure cmd {} {idem outer}}")
				Expect(evaluate("cmd")).To(Equal(STR("outer")))
			})
			It("should not access outer variables", func() {
				evaluate("set var val")
				Expect(execute("module {idem $var}")).To(Equal(
					ERROR(`cannot resolve variable "var"`),
				))
			})
			It("should not set outer variables", func() {
				evaluate("set var val")
				evaluate("module {set var val2; let cst val3}")
				Expect(rootScope.Context.Variables["var"]).To(Equal(STR("val")))
				Expect(rootScope.Context.Constants["cst"]).To(BeNil())
			})

			Describe("Control flow", func() {
				Describe("`return`", func() {
					It("should interrupt the body with `ERROR` Code", func() {
						Expect(execute("module {return value}")).To(Equal(
							ERROR("unexpected return"),
						))
					})
					It("should not define the module command", func() {
						evaluate("module cmd {return value}")
						Expect(rootScope.Context.Commands["cmd"]).To(BeNil())
					})
				})
				Describe("`tailcall`", func() {
					It("should interrupt the body with `ERROR` Code", func() {
						Expect(execute("module {tailcall {idem value}}")).To(Equal(
							ERROR("unexpected return"),
						))
					})
					It("should not define the module command", func() {
						evaluate("module cmd {return value}")
						Expect(rootScope.Context.Commands["cmd"]).To(BeNil())
					})
				})
				Describe("`yield`", func() {
					It("should interrupt the body with `ERROR` Code", func() {
						Expect(execute("module {yield value}")).To(Equal(
							ERROR("unexpected yield"),
						))
					})
					It("should not define the module command", func() {
						evaluate("module cmd {yield value}")
						Expect(rootScope.Context.Commands["cmd"]).To(BeNil())
					})
				})
				Describe("`error`", func() {
					It("should interrupt the body with `ERROR` Code", func() {
						Expect(execute("module {error message}")).To(Equal(ERROR("message")))
					})
					It("should not define the module command", func() {
						evaluate("module cmd {error message}")
						Expect(rootScope.Context.Commands["cmd"]).To(BeNil())
					})
				})
				Describe("`break`", func() {
					It("should interrupt the body with `ERROR` Code", func() {
						Expect(execute("module {break}")).To(Equal(ERROR("unexpected break")))
					})
					It("should not define the module command", func() {
						evaluate("module cmd {break}")
						Expect(rootScope.Context.Commands["cmd"]).To(BeNil())
					})
				})
				Describe("`continue`", func() {
					It("should interrupt the body with `ERROR` Code", func() {
						Expect(execute("module {continue}")).To(Equal(
							ERROR("unexpected continue"),
						))
					})
					It("should not define the module command", func() {
						evaluate("module cmd {continue}")
						Expect(rootScope.Context.Commands["cmd"]).To(BeNil())
					})
				})
				Describe("`pass`", func() {
					It("should interrupt the body with `ERROR` Code", func() {
						Expect(execute("module {pass}")).To(Equal(ERROR("unexpected pass")))
					})
					It("should not define the module command", func() {
						evaluate("module cmd {pass}")
						Expect(rootScope.Context.Commands["cmd"]).To(BeNil())
					})
				})
			})
		})

		Describe("Subcommands", func() {
			Describe("`subcommands`", func() {
				It("should return list of subcommands", func() {
					Expect(evaluate("[module {}] subcommands")).To(Equal(
						evaluate("list (subcommands exports import)"),
					))
				})

				Describe("Exceptions", func() {
					Specify("wrong arity", func() {
						Expect(execute("[module {}] subcommands a")).To(Equal(
							ERROR(`wrong # args: should be "<module> subcommands"`),
						))
					})
				})
			})

			Describe("`exports`", func() {
				It("should return a list", func() {
					Expect(evaluate("[module {}] exports")).To(Equal(LIST([]core.Value{})))
				})
				It("should return the list of module exports", func() {
					// Expect(
					// 	TODO preserve order?
					// 	evaluate("[module {export a; export b; export c}] exports"),
					// ).To(Equal(evaluate("list (a b c)")))
				})

				Describe("Exceptions", func() {
					Specify("wrong arity", func() {
						Expect(execute("[module {}] exports a")).To(Equal(
							ERROR(`wrong # args: should be "<module> exports"`),
						))
					})
				})
			})

			Describe("`import`", func() {
				It("should declare imported commands in the calling scope", func() {
					evaluate(`module mod {macro cmd {} {idem value}; export cmd}`)
					evaluate("mod import cmd")
					Expect(evaluate("cmd")).To(Equal(STR("value")))
				})
				It("should return nil", func() {
					evaluate(`module mod {macro cmd {} {idem value}; export cmd}`)
					Expect(execute("mod import cmd")).To(Equal(OK(NIL)))
				})
				It("should replace existing commands", func() {
					evaluate("closure cmd {} {idem val1} ")
					Expect(evaluate("cmd")).To(Equal(STR("val1")))
					evaluate(`module mod {macro cmd {} {idem val2}; export cmd}`)
					evaluate("mod import cmd")
					Expect(evaluate("cmd")).To(Equal(STR("val2")))
				})
				It("should evaluate macros in the caller scope", func() {
					evaluate(`module mod {macro cmd {} {set var val}; export cmd}`)
					evaluate("mod import cmd")
					evaluate("cmd")
					Expect(evaluate("get var")).To(Equal(STR("val")))
				})
				It("should evaluate closures in their scope", func() {
					evaluate(
						`module mod {set var val; closure cmd {} {get var}; export cmd}`,
					)
					evaluate("mod import cmd")
					Expect(evaluate("cmd")).To(Equal(STR("val")))
					Expect(execute("get var").Code).To(Equal(core.ResultCode_ERROR))
				})
				It("should resolve exported commands at call time", func() {
					evaluate(`
						module mod {
							closure cmd {} {idem val1}
							export cmd
							closure redefine {} {
								closure cmd {} {idem val2}
							}
							export redefine
						}
					`)
					Expect(evaluate("mod import cmd; cmd")).To(Equal(STR("val1")))
					evaluate("mod import redefine; redefine")
					Expect(evaluate("cmd")).To(Equal(STR("val1")))
					Expect(evaluate("mod import cmd; cmd")).To(Equal(STR("val2")))
				})
				It("should accept an optional alias name", func() {
					evaluate("macro cmd {} {idem original}")
					evaluate(`module mod {macro cmd {} {idem imported}; export cmd}`)
					evaluate("mod import cmd cmd2")
					Expect(evaluate("cmd")).To(Equal(STR("original")))
					Expect(evaluate("cmd2")).To(Equal(STR("imported")))
				})

				Describe("Exceptions", func() {
					Specify("wrong arity", func() {
						Expect(execute("[module {}] import")).To(Equal(
							ERROR(`wrong # args: should be "<module> import name ?alias?"`),
						))
						Expect(execute("[module {}] import a b c")).To(Equal(
							ERROR(`wrong # args: should be "<module> import name ?alias?"`),
						))
					})
					Specify("unknown export", func() {
						Expect(execute("[module {}] import a")).To(Equal(
							ERROR(`unknown export "a"`),
						))
					})
					Specify("unresolved export", func() {
						Expect(execute("[module {export a}] import a")).To(Equal(
							ERROR(`cannot resolve export "a"`),
						))
					})
					Specify("invalid import name", func() {
						Expect(execute("[module {}] import []")).To(Equal(
							ERROR("invalid import name"),
						))
					})
					Specify("invalid alias name", func() {
						Expect(execute("[module {}] import a []")).To(Equal(
							ERROR("invalid alias name"),
						))
					})
				})
			})

			Describe("Exceptions", func() {
				Specify("unknown subcommand", func() {
					Expect(execute("[module {}] unknownSubcommand")).To(Equal(
						ERROR(`unknown subcommand "unknownSubcommand"`),
					))
				})
				Specify("invalid subcommand name", func() {
					Expect(execute("[module {}] []")).To(Equal(
						ERROR("invalid subcommand name"),
					))
				})
			})
		})
	})

	Describe("export", func() {
		Describe("Specifications", func() {
			Specify("usage", func() {
				evaluate(`
					[module {
						closure usage {*args} {help $*args}
						export usage
					}] import usage
				`)
				Expect(evaluate("usage export")).To(Equal(STR("export name")))
				Expect(evaluate("usage export cmd")).To(Equal(STR("export name")))
			})

			It("should not exist in non-module scope", func() {
				Expect(execute("export")).To(Equal(
					ERROR(`cannot resolve command "export"`),
				))
			})
			It("should exist in module scope", func() {
				Expect(execute("module {export foo}").Code).To(Equal(core.ResultCode_OK))
			})
			It("should return nil", func() {
				evaluate(`
					module mod {
						set result [export cmd]
						closure cmd {} {get result}
					}
				`)
				evaluate("mod import cmd")
				Expect(execute("cmd")).To(Equal(OK(NIL)))
			})
			It("should add command name to exports", func() {
				evaluate("module mod {macro cmd {} {}; export cmd}")
				Expect(evaluate("mod exports")).To(Equal(evaluate("list (cmd)")))
			})
			It("should allow non-existing command names", func() {
				evaluate("module mod {export cmd}")
				Expect(evaluate("mod exports")).To(Equal(evaluate("list (cmd)")))
			})
		})

		Describe("Exceptions", func() {
			Specify("wrong arity", func() {
				Expect(execute("module {export}")).To(Equal(
					ERROR(`wrong # args: should be "export name"`),
				))
				Expect(execute("module {export a b}")).To(Equal(
					ERROR(`wrong # args: should be "export name"`),
				))
				Expect(execute("module {help export a b}")).To(Equal(
					ERROR(`wrong # args: should be "export name"`),
				))
			})
			Specify("invalid `name`", func() {
				Expect(execute("module {export []}")).To(Equal(
					ERROR("invalid export name"),
				))
			})
		})
	})

	Describe("import", func() {
		const moduleAPathRel = "tests/module-a.lna"
		var moduleAPathAbs = `"""` + path.Join(dirname, moduleAPathRel) + `"""`
		const moduleBPath = `"tests/module-b.lna"`
		const moduleCPath = `"tests/module-c.lna"`
		const moduleDPath = `"tests/module-d.lna"`
		const errorPath = `"tests/error.txt"`

		Describe("Specifications", func() {
			Specify("usage", func() {
				Expect(evaluate("help import")).To(Equal(
					STR("import path ?name|imports?"),
				))
				Expect(evaluate("help import /a")).To(Equal(
					STR("import path ?name|imports?"),
				))
				Expect(evaluate("help import /a b")).To(Equal(
					STR("import path ?name|imports?"),
				))
			})

			It("should return a module object", func() {
				value := evaluate(`set cmd [import ` + moduleAPathAbs + `]`)
				Expect(value.Type()).To(Equal(core.ValueType_COMMAND))
				Expect(evaluate("$cmd exports")).To(Equal(LIST([]core.Value{STR("name")})))
			})
			Specify(
				"relative paths should resolve relatively to the working directory",
				func() {
					evaluate(`import ` + moduleAPathRel + ` (name)`)
					Expect(evaluate("name")).To(Equal(STR("module-a")))
				},
			)
			Specify(
				"in-module relative paths should resolve relatively to the module path",
				func() {
					evaluate(`import ` + moduleCPath + ` (name)`)
					Expect(evaluate("name")).To(Equal(STR("module-a")))
				},
			)
			Specify("multiple imports should resolve to the same object", func() {
				value := evaluate(`import ` + moduleAPathAbs)
				Expect(evaluate(`import ` + moduleAPathAbs)).To(Equal(value))
				Expect(evaluate(`import ` + moduleAPathRel)).To(Equal(value))
			})
			It("should not support circular imports", func() {
				Expect(execute(`import ` + moduleDPath)).To(Equal(
					ERROR("circular imports are forbidden"),
				))
			})
			It("should support named modules", func() {
				foo := evaluate(
					`module {macro name {} {idem "foo module"}; export name}`,
				)
				Expect(foo.Type()).To(Equal(core.ValueType_COMMAND))
				moduleRegistry.Register("foo", foo.(core.CommandValue).Command().(*Module))
				Expect(evaluate("import foo")).To(Equal(foo))
				Expect(evaluate("import foo (name); name")).To(Equal(STR("foo module")))
			})

			Describe("`name`", func() {
				It("should define a new command", func() {
					evaluate(`import ` + moduleAPathRel + ` cmd`)
					Expect(rootScope.Context.Commands["cmd"]).NotTo(BeNil())
				})
				It("should replace existing commands", func() {
					evaluate("macro cmd {}")
					Expect(execute(`import ` + moduleAPathRel + ` cmd`).Code).To(Equal(
						core.ResultCode_OK,
					))
				})
				Specify("the named command should return its command object", func() {
					value := evaluate(`import ` + moduleAPathRel + ` cmd`)
					Expect(evaluate("cmd")).To(Equal(value))
				})
			})

			Describe("`imports`", func() {
				It("should declare imported commands in the calling scope", func() {
					evaluate(`import ` + moduleAPathRel + ` (name)`)
					Expect(evaluate("name")).To(Equal(STR("module-a")))
				})
				It("should accept tuples", func() {
					evaluate(`import ` + moduleAPathRel + ` (name)`)
					Expect(evaluate("name")).To(Equal(STR("module-a")))
				})
				It("should accept lists", func() {
					evaluate(`import ` + moduleAPathRel + ` [list (name)]`)
					Expect(evaluate("name")).To(Equal(STR("module-a")))
				})
				It("should accept blocks", func() {
					evaluate(`import ` + moduleAPathRel + ` {name}`)
					Expect(evaluate("name")).To(Equal(STR("module-a")))
				})
				It("should accept (name alias) tuples", func() {
					evaluate("macro name {} {idem original}")
					evaluate(`import ` + moduleAPathRel + ` ( (name name2) )`)
					Expect(evaluate("name")).To(Equal(STR("original")))
					Expect(evaluate("name2")).To(Equal(STR("module-a")))
				})

				Describe("Exceptions", func() {
					Specify("unknown export", func() {
						Expect(execute(`import ` + moduleAPathRel + ` (a)`)).To(Equal(
							ERROR(`unknown export "a"`),
						))
					})
					Specify("unresolved export", func() {
						Expect(execute(`import ` + moduleBPath + ` (unresolved)`)).To(Equal(
							ERROR(`cannot resolve export "unresolved"`),
						))
					})
					Specify("invalid import name", func() {
						Expect(execute(`import ` + moduleBPath + ` ([])`)).To(Equal(
							ERROR("invalid import name"),
						))
						Expect(execute(`import ` + moduleBPath + ` ( ([] a) )`)).To(Equal(
							ERROR("invalid import name"),
						))
					})
					Specify("invalid alias name", func() {
						Expect(execute(`import ` + moduleBPath + ` ( (name []) )`)).To(Equal(
							ERROR("invalid alias name"),
						))
					})
					Specify("invalid name tuple", func() {
						Expect(execute(`import ` + moduleBPath + ` ( () )`)).To(Equal(
							ERROR("invalid (name alias) tuple"),
						))
						Expect(execute(`import ` + moduleBPath + ` ( (a b c) )`)).To(Equal(
							ERROR("invalid (name alias) tuple"),
						))
					})
				})
			})
		})

		Describe("Exceptions", func() {
			Specify("wrong arity", func() {
				Expect(execute("import")).To(Equal(
					ERROR(`wrong # args: should be "import path ?name|imports?"`),
				))
				Expect(execute("import a b c")).To(Equal(
					ERROR(`wrong # args: should be "import path ?name|imports?"`),
				))
				Expect(execute("help import a b c")).To(Equal(
					ERROR(`wrong # args: should be "import path ?name|imports?"`),
				))
			})
			Specify("invalid path", func() {
				Expect(execute(`import []`)).To(Equal(ERROR("invalid path")))
			})
			Specify("unknown file", func() {
				result := execute(`import /unknownFile`)
				Expect(result.Code).To(Equal(core.ResultCode_ERROR))
				Expect(asString(result.Value)).To(ContainSubstring("error reading module"))
			})
			Specify("invalid file", func() {
				result := execute(`import /`)
				Expect(result.Code).To(Equal(core.ResultCode_ERROR))
				Expect(asString(result.Value)).To(ContainSubstring("error reading module"))
			})
			Specify("parsing error", func() {
				Expect(execute(`import ` + errorPath)).To(Equal(
					ERROR("unmatched left brace"),
				))
			})
		})
	})
})
