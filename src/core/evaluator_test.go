package core_test

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "helena/core"
)

func mapValues(values []Value) any {
	result := make([]any, len(values))
	for i, value := range values {
		result[i] = mapValue(value)
	}
	return result
}
func mapValue(value Value) any {
	switch v := value.(type) {
	case NilValue:
		return NIL
	case StringValue:
		return v.Value
	case IntegerValue:
		return v.Value
	case ListValue:
		return mapValues(v.Values)
		//   if (value instanceof DictionaryValue) {
		//     const result = {};
		//     value.map.forEach((v, k) => {
		//       result[k] = mapValue(v);
		//     });
		//     return result;
		//   }
	case TupleValue:
		return mapValues(v.Values)
	case ScriptValue:
		return v.Script
	case QualifiedValue:
		return map[any]any{
			"source":    mapValue(v.Source),
			"selectors": mapSelectors(v.Selectors),
		}
	}
	panic("CANTHAPPEN")
}
func mapSelectors(selectors []Selector) [](map[any]any) {
	result := make([](map[any]any), len(selectors))
	for i, selector := range selectors {
		result[i] = mapSelector(selector)
	}
	return result
}
func mapSelector(selector Selector) map[any]any {
	switch s := selector.(type) {
	case IndexedSelector:
		return map[any]any{"index": mapValue(s.Index)}
	case KeyedSelector:
		return map[any]any{"keys": mapValues(s.Keys)}
	case GenericSelector:
		return map[any]any{"rules": mapValues(s.Rules)}
	}
	return map[any]any{"custom": selector}
}

var _ = Describe("CompilingEvaluator", func() {
	var tokenizer Tokenizer
	var parser *Parser
	var variableResolver *mockVariableResolver
	var commandResolver *mockCommandResolver
	var selectorResolver *mockSelectorResolver
	var evaluator Evaluator

	parse := func(script string) *Script {
		return parser.Parse(tokenizer.Tokenize(script)).Script
	}
	firstSentence := func(script *Script) Sentence { return script.Sentences[0] }
	firstWord := func(script *Script) Word {
		return firstSentence(script).Words[0]
	}

	BeforeEach(func() {
		tokenizer = Tokenizer{}
		parser = &Parser{}
		variableResolver = newMockVariableResolver()
		commandResolver = newMockCommandResolver()
		selectorResolver = newMockSelectorResolver()
		evaluator = NewCompilingEvaluator(
			variableResolver,
			commandResolver,
			selectorResolver,
			nil,
		)
	})

	Describe("words", func() {
		Describe("roots", func() {
			Specify("literals", func() {
				word := firstWord(parse("word"))
				value := evaluator.EvaluateWord(word).Value
				Expect(mapValue(value)).To(Equal("word"))
			})

			Describe("expressions", func() {
				Specify("empty expression", func() {
					word := firstWord(parse("[]"))
					value := evaluator.EvaluateWord(word).Value
					Expect(value).To(Equal(NIL))
				})
			})
		})

		Describe("qualified words", func() {
			Describe("scalars", func() {
				Specify("indexed selector", func() {
					word := firstWord(parse("var[123]"))
					value := evaluator.EvaluateWord(word).Value
					Expect(mapValue(value)).To(Equal(map[any]any{
						"source":    "var",
						"selectors": [](map[any]any){{"index": "123"}},
					}))
				})
				Specify("keyed selector", func() {
					word := firstWord(parse("var(key)"))
					value := evaluator.EvaluateWord(word).Value
					Expect(mapValue(value)).To(Equal(map[any]any{
						"source":    "var",
						"selectors": [](map[any]any){{"keys": []any{"key"}}},
					}))
				})
				Describe("generic selectors", func() {
					BeforeEach(func() {
						selectorResolver.register(func(rules []Value) TypedResult[Selector] {
							return CreateGenericSelector(rules)
						})
					})
					Specify("simple rule", func() {
						word := firstWord(parse("var{rule}"))
						value := evaluator.EvaluateWord(word).Value
						Expect(mapValue(value)).To(Equal(map[any]any{
							"source":    "var",
							"selectors": [](map[any]any){{"rules": []any{[]any{"rule"}}}},
						}))
					})
					Specify("rule with literal arguments", func() {
						word := firstWord(parse("var{rule arg1 arg2}"))
						value := evaluator.EvaluateWord(word).Value
						Expect(mapValue(value)).To(Equal(map[any]any{
							"source":    "var",
							"selectors": [](map[any]any){{"rules": []any{[]any{"rule", "arg1", "arg2"}}}},
						}))
					})
					Specify("multiple rules", func() {
						word := firstWord(parse("var{rule1;rule2}"))
						value := evaluator.EvaluateWord(word).Value
						Expect(mapValue(value)).To(Equal(map[any]any{
							"source":    "var",
							"selectors": [](map[any]any){{"rules": []any{[]any{"rule1"}, []any{"rule2"}}}},
						}))
					})
					Specify("successive selectors", func() {
						word := firstWord(parse("var{rule1}{rule2}"))
						value := evaluator.EvaluateWord(word).Value
						Expect(mapValue(value)).To(Equal(map[any]any{
							"source":    "var",
							"selectors": [](map[any]any){{"rules": []any{[]any{"rule1"}}}, {"rules": []any{[]any{"rule2"}}}},
						}))
					})
					Specify("indirect selector", func() {
						variableResolver.register("var2", STR("rule"))
						word := firstWord(parse("var1{$var2}"))
						value := evaluator.EvaluateWord(word).Value
						Expect(mapValue(value)).To(Equal(map[any]any{
							"source":    "var1",
							"selectors": [](map[any]any){{"rules": []any{[]any{"rule"}}}},
						}))
					})
				})
				Describe("custom selectors", func() {
					builder := func(selector Selector) builderFn {
						return func(_ []Value) TypedResult[Selector] {
							return OK_T(NIL, selector)
						}
					}
					BeforeEach(func() {
						selectorResolver.register(builder(lastSelector{}))
					})
					Specify("simple rule", func() {
						word := firstWord(parse("var{last}"))
						value := evaluator.EvaluateWord(word).Value
						Expect(mapValue(value)).To(Equal(map[any]any{
							"source":    "var",
							"selectors": [](map[any]any){{"custom": lastSelector{}}},
						}))
					})
					Specify("indirect selector", func() {
						variableResolver.register("var2", STR("last"))
						word := firstWord(parse("var1{$var2}"))
						value := evaluator.EvaluateWord(word).Value
						Expect(mapValue(value)).To(Equal(map[any]any{
							"source":    "var1",
							"selectors": [](map[any]any){{"custom": lastSelector{}}},
						}))
					})
				})

				Specify("multiple selectors", func() {
					selectorResolver.register(func(rules []Value) TypedResult[Selector] {
						return CreateGenericSelector(rules)
					})
					word := firstWord(
						parse(
							"var(key1 key2 key3){rule1;rule2}[1]{rule3}{rule4}[2](key4)",
						),
					)
					value := evaluator.EvaluateWord(word).Value
					Expect(mapValue(value)).To(Equal(map[any]any{
						"source": "var",
						"selectors": [](map[any]any){
							{"keys": []any{"key1", "key2", "key3"}},
							{"rules": []any{[]any{"rule1"}, []any{"rule2"}}},
							{"index": "1"},
							{"rules": []any{[]any{"rule3"}}},
							{"rules": []any{[]any{"rule4"}}},
							{"index": "2"},
							{"keys": []any{"key4"}},
						},
					}))
				})
				Describe("exceptions", func() {
					Specify("empty indexed selector", func() {
						word := firstWord(parse("var[1][]"))
						Expect(evaluator.EvaluateWord(word)).To(Equal(
							ERROR("invalid index"),
						))
					})
					Specify("empty keyed selector", func() {
						word := firstWord(parse("var(key)()"))
						Expect(evaluator.EvaluateWord(word)).To(Equal(
							ERROR("empty selector"),
						))
					})
					Specify("empty generic selector", func() {
						selectorResolver.register(func(rules []Value) TypedResult[Selector] {
							return CreateGenericSelector(rules)
						})
						word := firstWord(parse("var{rule}{}"))
						Expect(evaluator.EvaluateWord(word)).To(Equal(
							ERROR("empty selector"),
						))
					})
					Specify("invalid trailing morphemes", func() {
						word := firstWord(parse("var(key1)2"))
						Expect(evaluator.EvaluateWord(word)).To(Equal(
							ERROR("invalid word structure"),
						))
					})
				})
			})
			Describe("tuples", func() {
				Specify("indexed selector", func() {
					word := firstWord(parse("(var1 var2 (var3 var4))[123]"))
					value := evaluator.EvaluateWord(word).Value
					Expect(mapValue(value)).To(Equal(map[any]any{
						"source":    []any{"var1", "var2", []any{"var3", "var4"}},
						"selectors": [](map[any]any){{"index": "123"}},
					}))
				})
				Specify("keyed selector", func() {
					word := firstWord(parse("(var1 (var2) var3)(key)"))
					value := evaluator.EvaluateWord(word).Value
					Expect(mapValue(value)).To(Equal(map[any]any{
						"source":    []any{"var1", []any{"var2"}, "var3"},
						"selectors": [](map[any]any){{"keys": []any{"key"}}},
					}))
				})
				Describe("generic selectors", func() {
					BeforeEach(func() {
						selectorResolver.register(func(rules []Value) TypedResult[Selector] {
							return CreateGenericSelector(rules)
						})
					})
					Specify("simple rule", func() {
						word := firstWord(parse("(var1 (var2) var3){rule}"))
						value := evaluator.EvaluateWord(word).Value
						Expect(mapValue(value)).To(Equal(map[any]any{
							"source":    []any{"var1", []any{"var2"}, "var3"},
							"selectors": [](map[any]any){{"rules": []any{[]any{"rule"}}}},
						}))
					})
					Specify("rule with literal arguments", func() {
						word := firstWord(
							parse("(var1 (var2) var3){rule arg1 arg2}"),
						)
						value := evaluator.EvaluateWord(word).Value
						Expect(mapValue(value)).To(Equal(map[any]any{
							"source":    []any{"var1", []any{"var2"}, "var3"},
							"selectors": [](map[any]any){{"rules": []any{[]any{"rule", "arg1", "arg2"}}}},
						}))
					})
					Specify("multiple rules", func() {
						word := firstWord(parse("(var1 (var2) var3){rule1;rule2}"))
						value := evaluator.EvaluateWord(word).Value
						Expect(mapValue(value)).To(Equal(map[any]any{
							"source":    []any{"var1", []any{"var2"}, "var3"},
							"selectors": [](map[any]any){{"rules": []any{[]any{"rule1"}, []any{"rule2"}}}},
						}))
					})
					Specify("successive selectors", func() {
						word := firstWord(parse("(var1 (var2) var3){rule1}{rule2}"))
						value := evaluator.EvaluateWord(word).Value
						Expect(mapValue(value)).To(Equal(map[any]any{
							"source":    []any{"var1", []any{"var2"}, "var3"},
							"selectors": [](map[any]any){{"rules": []any{[]any{"rule1"}}}, {"rules": []any{[]any{"rule2"}}}},
						}))
					})
					Specify("indirect selector", func() {
						variableResolver.register("var4", STR("rule"))
						word := firstWord(parse("(var1 (var2) var3){$var4}"))
						value := evaluator.EvaluateWord(word).Value
						Expect(mapValue(value)).To(Equal(map[any]any{
							"source":    []any{"var1", []any{"var2"}, "var3"},
							"selectors": [](map[any]any){{"rules": []any{[]any{"rule"}}}},
						}))
					})
				})
				Specify("multiple selectors", func() {
					selectorResolver.register(func(rules []Value) TypedResult[Selector] {
						return CreateGenericSelector(rules)
					})
					word := firstWord(
						parse(
							"((var))(key1 key2 key3){rule1;rule2}[1]{rule3}{rule4}[2](key4)",
						),
					)
					value := evaluator.EvaluateWord(word).Value
					Expect(mapValue(value)).To(Equal(map[any]any{
						"source": []any{[]any{"var"}},
						"selectors": [](map[any]any){
							{"keys": []any{"key1", "key2", "key3"}},
							{"rules": []any{[]any{"rule1"}, []any{"rule2"}}},
							{"index": "1"},
							{"rules": []any{[]any{"rule3"}}},
							{"rules": []any{[]any{"rule4"}}},
							{"index": "2"},
							{"keys": []any{"key4"}},
						},
					}))
				})
				Describe("exceptions", func() {
					Specify("empty indexed selector", func() {
						word := firstWord(parse("(var1 var2 (var3 var4))[1][]"))
						Expect(evaluator.EvaluateWord(word)).To(Equal(
							ERROR("invalid index"),
						))
					})
					Specify("empty keyed selector", func() {
						word := firstWord(parse("(var1 var2 (var3 var4))(key)()"))
						Expect(evaluator.EvaluateWord(word)).To(Equal(
							ERROR("empty selector"),
						))
					})
					Specify("empty generic selector", func() {
						selectorResolver.register(func(rules []Value) TypedResult[Selector] {
							return CreateGenericSelector(rules)
						})
						word := firstWord(parse("(var1 var2 (var3 var4)){rule}{}"))
						Expect(evaluator.EvaluateWord(word)).To(Equal(
							ERROR("empty selector"),
						))
					})
					Specify("invalid trailing morphemes", func() {
						word := firstWord(parse("(var1 var2)(key1)2"))
						Expect(evaluator.EvaluateWord(word)).To(Equal(
							ERROR("invalid word structure"),
						))
					})
				})
			})
		})

		Describe("substitutions", func() {
			Describe("scalars", func() {
				Specify("simple substitution", func() {
					variableResolver.register("var", STR("value"))
					word := firstWord(parse("$var"))
					value := evaluator.EvaluateWord(word).Value
					Expect(mapValue(value)).To(Equal("value"))
				})
				Specify("double substitution", func() {
					variableResolver.register("var1", STR("var2"))
					variableResolver.register("var2", STR("value"))
					word := firstWord(parse("$$var1"))
					value := evaluator.EvaluateWord(word).Value
					Expect(mapValue(value)).To(Equal("value"))
				})
				Specify("triple substitution", func() {
					variableResolver.register("var1", STR("var2"))
					variableResolver.register("var2", STR("var3"))
					variableResolver.register("var3", STR("value"))
					word := firstWord(parse("$$$var1"))
					value := evaluator.EvaluateWord(word).Value
					Expect(mapValue(value)).To(Equal("value"))
				})
			})

			Describe("tuples", func() {
				Specify("single variable", func() {
					variableResolver.register("var", STR("value"))
					word := firstWord(parse("$(var)"))
					value := evaluator.EvaluateWord(word).Value
					Expect(mapValue(value)).To(Equal([]any{"value"}))
				})
				Specify("multiple variables", func() {
					variableResolver.register("var1", STR("value1"))
					variableResolver.register("var2", STR("value2"))
					word := firstWord(parse("$(var1 var2)"))
					value := evaluator.EvaluateWord(word).Value
					Expect(mapValue(value)).To(Equal([]any{"value1", "value2"}))
				})
				Specify("double substitution", func() {
					variableResolver.register("var1", STR("var2"))
					variableResolver.register("var2", STR("value"))
					word := firstWord(parse("$$(var1)"))
					value := evaluator.EvaluateWord(word).Value
					Expect(mapValue(value)).To(Equal([]any{"value"}))
				})
				Specify("nested tuples", func() {
					variableResolver.register("var1", STR("value1"))
					variableResolver.register("var2", STR("value2"))
					word := firstWord(parse("$(var1 (var2))"))
					value := evaluator.EvaluateWord(word).Value
					Expect(mapValue(value)).To(Equal([]any{"value1", []any{"value2"}}))
				})
				Specify("nested double substitution", func() {
					variableResolver.register("var1", STR("var2"))
					variableResolver.register("var2", STR("value"))
					word := firstWord(parse("$$((var1))"))
					value := evaluator.EvaluateWord(word).Value
					Expect(mapValue(value)).To(Equal([]any{[]any{"value"}}))
				})
				Specify("nested qualified words", func() {
					variableResolver.register(
						"var1",
						LIST([]Value{STR("value1"), STR("value2")}),
					)
					variableResolver.register("var2", DICT(map[string]Value{"key": STR("value3")}))
					word := firstWord(parse("$(var1[0] var2(key))"))
					value := evaluator.EvaluateWord(word).Value
					Expect(mapValue(value)).To(Equal([]any{"value1", "value3"}))
				})
			})

			Describe("blocks", func() {
				Specify("varname with spaces", func() {
					variableResolver.register("variable name", STR("value"))
					word := firstWord(parse("${variable name}"))
					value := evaluator.EvaluateWord(word).Value
					Expect(mapValue(value)).To(Equal("value"))
				})
				Specify("varname with special characters", func() {
					variableResolver.register(`variable " " name`, STR("value"))
					word := firstWord(parse(`${variable " " name}`))
					value := evaluator.EvaluateWord(word).Value
					Expect(mapValue(value)).To(Equal("value"))
				})
				Specify("varname with continuations", func() {
					variableResolver.register("variable name", STR("value"))
					word := firstWord(parse("${variable\\\n \t\r     name}"))
					value := evaluator.EvaluateWord(word).Value
					Expect(mapValue(value)).To(Equal("value"))
				})
			})

			Describe("expressions", func() {
				Specify("simple substitution", func() {
					commandResolver.register(
						"cmd",
						functionCommand{func(_ []Value) Value { return STR("value") }},
					)
					word := firstWord(parse("$[cmd]"))
					value := evaluator.EvaluateWord(word).Value
					Expect(mapValue(value)).To(Equal("value"))
				})
				Specify("double substitution, scalar", func() {
					commandResolver.register(
						"cmd",
						functionCommand{func(_ []Value) Value { return STR("var") }},
					)
					variableResolver.register("var", STR("value"))
					word := firstWord(parse("$$[cmd]"))
					value := evaluator.EvaluateWord(word).Value
					Expect(mapValue(value)).To(Equal("value"))
				})
				Specify("double substitution, tuple", func() {
					commandResolver.register(
						"cmd",
						functionCommand{func(_ []Value) Value { return TUPLE([]Value{STR("var1"), STR("var2")}) }},
					)
					variableResolver.register("var1", STR("value1"))
					variableResolver.register("var2", STR("value2"))
					word := firstWord(parse("$$[cmd]"))
					value := evaluator.EvaluateWord(word).Value
					Expect(mapValue(value)).To(Equal([]any{"value1", "value2"}))
				})
			})

			Describe("indexed selectors", func() {
				Specify("simple substitution", func() {
					variableResolver.register(
						"var",
						LIST([]Value{STR("value1"), STR("value2")}),
					)
					word := firstWord(parse("$var[1]"))
					value := evaluator.EvaluateWord(word).Value
					Expect(mapValue(value)).To(Equal("value2"))
				})
				Specify("double substitution", func() {
					variableResolver.register("var1", LIST([]Value{STR("var2")}))
					variableResolver.register("var2", STR("value"))
					word := firstWord(parse("$$var1[0]"))
					value := evaluator.EvaluateWord(word).Value
					Expect(mapValue(value)).To(Equal("value"))
				})
				Specify("successive indexes", func() {
					variableResolver.register(
						"var",
						LIST([]Value{STR("value1"), LIST([]Value{STR("value2_1"), STR("value2_2")})}),
					)
					word := firstWord(parse("$var[1][0]"))
					value := evaluator.EvaluateWord(word).Value
					Expect(mapValue(value)).To(Equal("value2_1"))
				})
				Specify("indirect index", func() {
					variableResolver.register(
						"var1",
						LIST([]Value{STR("value1"), STR("value2"), STR("value3")}),
					)
					variableResolver.register("var2", STR("1"))
					word := firstWord(parse("$var1[$var2]"))
					value := evaluator.EvaluateWord(word).Value
					Expect(mapValue(value)).To(Equal("value2"))
				})
				Specify("command index", func() {
					commandResolver.register(
						"cmd",
						functionCommand{func(_ []Value) Value { return STR("1") }},
					)
					variableResolver.register(
						"var",
						LIST([]Value{STR("value1"), STR("value2"), STR("value3")}),
					)
					word := firstWord(parse("$var[cmd]"))
					value := evaluator.EvaluateWord(word).Value
					Expect(mapValue(value)).To(Equal("value2"))
				})
				Specify("tuple", func() {
					variableResolver.register(
						"var1",
						LIST([]Value{STR("value1"), STR("value2")}),
					)
					variableResolver.register(
						"var2",
						LIST([]Value{STR("value3"), STR("value4")}),
					)
					word := firstWord(parse("$(var1 var2)[1]"))
					value := evaluator.EvaluateWord(word).Value
					Expect(mapValue(value)).To(Equal([]any{"value2", "value4"}))
				})
				Specify("recursive tuple", func() {
					variableResolver.register(
						"var1",
						LIST([]Value{STR("value1"), STR("value2")}),
					)
					variableResolver.register(
						"var2",
						LIST([]Value{STR("value3"), STR("value4")}),
					)
					word := firstWord(parse("$(var1 (var2))[1]"))
					value := evaluator.EvaluateWord(word).Value
					Expect(mapValue(value)).To(Equal([]any{"value2", []any{"value4"}}))
				})
				Specify("tuple with double substitution", func() {
					variableResolver.register("var1", LIST([]Value{STR("var3"), STR("var4")}))
					variableResolver.register("var2", LIST([]Value{STR("var5"), STR("var6")}))
					variableResolver.register("var4", STR("value1"))
					variableResolver.register("var6", STR("value2"))
					word := firstWord(parse("$$(var1 var2)[1]"))
					value := evaluator.EvaluateWord(word).Value
					Expect(mapValue(value)).To(Equal([]any{"value1", "value2"}))
				})
				Specify("scalar expression", func() {
					commandResolver.register(
						"cmd",
						functionCommand{func(_ []Value) Value { return LIST([]Value{STR("value")}) }},
					)
					word := firstWord(parse("$[cmd][0]"))
					value := evaluator.EvaluateWord(word).Value
					Expect(mapValue(value)).To(Equal("value"))
				})
				Specify("tuple expression", func() {
					commandResolver.register(
						"cmd",
						functionCommand{func(_ []Value) Value {
							return TUPLE([]Value{LIST([]Value{STR("value1")}), LIST([]Value{STR("value2")})})
						}},
					)
					word := firstWord(parse("$[cmd][0]"))
					value := evaluator.EvaluateWord(word).Value
					Expect(mapValue(value)).To(Equal([]any{"value1", "value2"}))
				})
			})

			Describe("keyed selectors", func() {
				Specify("simple substitution", func() {
					variableResolver.register("var", DICT(map[string]Value{"key": STR("value")}))
					word := firstWord(parse("$var(key)"))
					value := evaluator.EvaluateWord(word).Value
					Expect(mapValue(value)).To(Equal("value"))
				})
				Specify("double substitution", func() {
					variableResolver.register("var1", DICT(map[string]Value{"key": STR("var2")}))
					variableResolver.register("var2", STR("value"))
					word := firstWord(parse("$$var1(key)"))
					value := evaluator.EvaluateWord(word).Value
					Expect(mapValue(value)).To(Equal("value"))
				})
				Specify("recursive keys", func() {
					variableResolver.register(
						"var",
						DICT(map[string]Value{
							"key1": DICT(map[string]Value{"key2": STR("value")}),
						}),
					)
					word := firstWord(parse("$var(key1 key2)"))
					value := evaluator.EvaluateWord(word).Value
					Expect(mapValue(value)).To(Equal("value"))
				})
				Specify("successive keys", func() {
					variableResolver.register(
						"var",
						DICT(map[string]Value{
							"key1": DICT(map[string]Value{"key2": STR("value")}),
						}),
					)
					word := firstWord(parse("$var(key1)(key2)"))
					value := evaluator.EvaluateWord(word).Value
					Expect(mapValue(value)).To(Equal("value"))
				})
				Specify("indirect key", func() {
					variableResolver.register(
						"var1",
						DICT(map[string]Value{
							"key": STR("value"),
						}),
					)
					variableResolver.register("var2", STR("key"))
					word := firstWord(parse("$var1($var2)"))
					value := evaluator.EvaluateWord(word).Value
					Expect(mapValue(value)).To(Equal("value"))
				})
				Specify("string key", func() {
					variableResolver.register(
						"var",
						DICT(map[string]Value{
							"arbitrary key": STR("value"),
						}),
					)
					word := firstWord(parse(`$var("arbitrary key")`))
					value := evaluator.EvaluateWord(word).Value
					Expect(mapValue(value)).To(Equal("value"))
				})
				Specify("block key", func() {
					variableResolver.register(
						"var",
						DICT(map[string]Value{
							"arbitrary key": STR("value"),
						}),
					)
					word := firstWord(parse("$var({arbitrary key})"))
					value := evaluator.EvaluateWord(word).Value
					Expect(mapValue(value)).To(Equal("value"))
				})
				Specify("tuple", func() {
					variableResolver.register("var1", DICT(map[string]Value{"key": STR("value1")}))
					variableResolver.register("var2", DICT(map[string]Value{"key": STR("value2")}))
					word := firstWord(parse("$(var1 var2)(key)"))
					value := evaluator.EvaluateWord(word).Value
					Expect(mapValue(value)).To(Equal([]any{"value1", "value2"}))
				})
				Specify("recursive tuple", func() {
					variableResolver.register("var1", DICT(map[string]Value{"key": STR("value1")}))
					variableResolver.register("var2", DICT(map[string]Value{"key": STR("value2")}))
					word := firstWord(parse("$(var1 (var2))(key)"))
					value := evaluator.EvaluateWord(word).Value
					Expect(mapValue(value)).To(Equal([]any{"value1", []any{"value2"}}))
				})
				Specify("tuple with double substitution", func() {
					variableResolver.register("var1", DICT(map[string]Value{"key": STR("var3")}))
					variableResolver.register("var2", DICT(map[string]Value{"key": STR("var4")}))
					variableResolver.register("var3", STR("value3"))
					variableResolver.register("var4", STR("value4"))
					word := firstWord(parse("$$(var1 var2)(key)"))
					value := evaluator.EvaluateWord(word).Value
					Expect(mapValue(value)).To(Equal([]any{"value3", "value4"}))
				})
				Specify("scalar expression", func() {
					commandResolver.register(
						"cmd",
						functionCommand{func(_ []Value) Value { return DICT(map[string]Value{"key": STR("value")}) }},
					)
					word := firstWord(parse("$[cmd](key)"))
					value := evaluator.EvaluateWord(word).Value
					Expect(mapValue(value)).To(Equal("value"))
				})
				Specify("tuple expression", func() {
					commandResolver.register(
						"cmd",
						functionCommand{func(_ []Value) Value {
							return TUPLE([]Value{
								DICT(map[string]Value{"key": STR("value1")}),
								DICT(map[string]Value{"key": STR("value2")}),
							})
						}},
					)
					word := firstWord(parse("$[cmd](key)"))
					value := evaluator.EvaluateWord(word).Value
					Expect(mapValue(value)).To(Equal([]any{"value1", "value2"}))
				})
				Describe("exceptions", func() {
					Specify("empty selector", func() {
						variableResolver.register("var", DICT(map[string]Value{"key": STR("value")}))
						word := firstWord(parse("$var()"))
						Expect(evaluator.EvaluateWord(word)).To(Equal(
							ERROR("empty selector"),
						))
					})
				})
			})

			Describe("custom selectors", func() {
				builder := func(selector Selector) builderFn {
					return func(_ []Value) TypedResult[Selector] {
						return OK_T(NIL, selector)
					}
				}
				BeforeEach(func() {
					selectorResolver.register(builder(lastSelector{}))
				})
				Specify("simple substitution", func() {
					variableResolver.register(
						"var",
						LIST([]Value{STR("value1"), STR("value2"), STR("value3")}),
					)
					word := firstWord(parse("$var{last}"))
					value := evaluator.EvaluateWord(word).Value
					Expect(mapValue(value)).To(Equal("value3"))
				})
				Specify("double substitution", func() {
					variableResolver.register("var1", LIST([]Value{STR("var2"), STR("var3")}))
					variableResolver.register("var3", STR("value"))
					word := firstWord(parse("$$var1{last}"))
					value := evaluator.EvaluateWord(word).Value
					Expect(mapValue(value)).To(Equal("value"))
				})
				Specify("successive selectors", func() {
					variableResolver.register(
						"var",
						LIST([]Value{STR("value1"), LIST([]Value{STR("value2_1"), STR("value2_2")})}),
					)
					word := firstWord(parse("$var{last}{last}"))
					value := evaluator.EvaluateWord(word).Value
					Expect(mapValue(value)).To(Equal("value2_2"))
				})
				Specify("indirect selector", func() {
					variableResolver.register(
						"var1",
						LIST([]Value{STR("value1"), STR("value2"), STR("value3")}),
					)
					variableResolver.register("var2", STR("last"))
					word := firstWord(parse("$var1{$var2}"))
					value := evaluator.EvaluateWord(word).Value
					Expect(mapValue(value)).To(Equal("value3"))
				})
				Specify("tuple", func() {
					variableResolver.register(
						"var1",
						LIST([]Value{STR("value1"), STR("value2")}),
					)
					variableResolver.register(
						"var2",
						LIST([]Value{STR("value3"), STR("value4"), STR("value5")}),
					)
					word := firstWord(parse("$(var1 var2){last}"))
					value := evaluator.EvaluateWord(word).Value
					Expect(mapValue(value)).To(Equal([]any{"value2", "value5"}))
				})
				Specify("recursive tuple", func() {
					variableResolver.register(
						"var1",
						LIST([]Value{STR("value1"), STR("value2")}),
					)
					variableResolver.register(
						"var2",
						LIST([]Value{STR("value3"), STR("value4")}),
					)
					word := firstWord(parse("$(var1 (var2))[1]"))
					value := evaluator.EvaluateWord(word).Value
					Expect(mapValue(value)).To(Equal([]any{"value2", []any{"value4"}}))
				})
				Specify("tuple with double substitution", func() {
					variableResolver.register("var1", LIST([]Value{STR("var3"), STR("var4")}))
					variableResolver.register("var2", LIST([]Value{STR("var5"), STR("var6")}))
					variableResolver.register("var4", STR("value1"))
					variableResolver.register("var6", STR("value2"))
					word := firstWord(parse("$$(var1 var2)[1]"))
					value := evaluator.EvaluateWord(word).Value
					Expect(mapValue(value)).To(Equal([]any{"value1", "value2"}))
				})
				Specify("expression", func() {
					commandResolver.register(
						"cmd",
						functionCommand{func(_ []Value) Value { return LIST([]Value{STR("value1"), STR("value2")}) }},
					)
					word := firstWord(parse("$[cmd]{last}"))
					value := evaluator.EvaluateWord(word).Value
					Expect(mapValue(value)).To(Equal("value2"))
				})
			})

			Describe("exceptions", func() {
				Specify("unknown variable", func() {
					word := firstWord(parse("$var"))
					Expect(evaluator.EvaluateWord(word)).To(Equal(
						ERROR(`cannot resolve variable "var"`),
					))
				})
			})
		})

		Describe("ignored words", func() {
			Specify("line comments", func() {
				word := firstWord(parse("# this is a comment"))
				value := evaluator.EvaluateWord(word).Value
				Expect(value).To(Equal(NIL))
			})
			Specify("block comments", func() {
				word := firstWord(parse("#{ this is\n a\nblock comment }#"))
				value := evaluator.EvaluateWord(word).Value
				Expect(value).To(Equal(NIL))
			})
		})

		Specify("complex case", func() {
			variableResolver.register("var", DICT(map[string]Value{"key": STR("value")}))
			commandResolver.register(
				"cmd",
				functionCommand{func(args []Value) Value {
					return STR(asString(args[1]) + asString(args[2]))
				}},
			)
			word := firstWord(
				parse("prefix_${var}(key)_infix_[cmd a b]_suffix"),
			)
			value := evaluator.EvaluateWord(word).Value
			Expect(mapValue(value)).To(Equal("prefix_value_infix_ab_suffix"))
		})
	})

	Describe("word expansion", func() {
		Describe("tuple words", func() {
			Specify("empty string", func() {
				variableResolver.register("var", STR(""))
				word := firstWord(parse("(prefix $*var suffix)"))
				value := evaluator.EvaluateWord(word).Value
				Expect(mapValue(value)).To(Equal([]any{"prefix", "", "suffix"}))
			})
			Specify("scalar variable", func() {
				variableResolver.register("var", STR("value"))
				word := firstWord(parse("(prefix $*var suffix)"))
				value := evaluator.EvaluateWord(word).Value
				Expect(mapValue(value)).To(Equal([]any{"prefix", "value", "suffix"}))
			})
			Specify("empty tuple variable", func() {
				variableResolver.register("var", TUPLE([]Value{}))
				word := firstWord(parse("(prefix $*var suffix)"))
				value := evaluator.EvaluateWord(word).Value
				Expect(mapValue(value)).To(Equal([]any{"prefix", "suffix"}))
			})
			Specify("tuple variable", func() {
				variableResolver.register(
					"var",
					TUPLE([]Value{STR("value1"), STR("value2")}),
				)
				word := firstWord(parse("(prefix $*var suffix)"))
				value := evaluator.EvaluateWord(word).Value
				Expect(mapValue(value)).To(Equal([]any{
					"prefix",
					"value1",
					"value2",
					"suffix",
				}))
			})
			Specify("scalar expression", func() {
				commandResolver.register(
					"cmd",
					functionCommand{func(_ []Value) Value { return STR("value") }},
				)
				word := firstWord(parse("(prefix $*[cmd] suffix)"))
				value := evaluator.EvaluateWord(word).Value
				Expect(mapValue(value)).To(Equal([]any{"prefix", "value", "suffix"}))
			})
			Specify("tuple expression", func() {
				commandResolver.register(
					"cmd",
					functionCommand{func(_ []Value) Value { return TUPLE([]Value{STR("value1"), STR("value2")}) }},
				)
				word := firstWord(parse("(prefix $*[cmd] suffix)"))
				value := evaluator.EvaluateWord(word).Value
				Expect(mapValue(value)).To(Equal([]any{
					"prefix",
					"value1",
					"value2",
					"suffix",
				}))
			})
		})
		Describe("sentences", func() {
			BeforeEach(func() {
				commandResolver.register(
					"cmd",
					functionCommand{func(args []Value) Value { return TUPLE(args) }},
				)
			})
			Specify("empty string", func() {
				variableResolver.register("var", STR(""))
				sentence := firstSentence(parse("cmd $*var arg"))
				value := evaluator.EvaluateSentence(sentence).Value
				Expect(mapValue(value)).To(Equal([]any{"cmd", "", "arg"}))
			})
			Specify("scalar variable", func() {
				variableResolver.register("var", STR("value"))
				sentence := firstSentence(parse("cmd $*var arg"))
				value := evaluator.EvaluateSentence(sentence).Value
				Expect(mapValue(value)).To(Equal([]any{"cmd", "value", "arg"}))
			})
			Specify("empty tuple variable", func() {
				variableResolver.register("var", TUPLE([]Value{}))
				sentence := firstSentence(parse("cmd $*var arg"))
				value := evaluator.EvaluateSentence(sentence).Value
				Expect(mapValue(value)).To(Equal([]any{"cmd", "arg"}))
			})
			Specify("tuple variable", func() {
				variableResolver.register(
					"var",
					TUPLE([]Value{STR("value1"), STR("value2")}),
				)
				sentence := firstSentence(parse("cmd $*var arg"))
				value := evaluator.EvaluateSentence(sentence).Value
				Expect(mapValue(value)).To(Equal([]any{"cmd", "value1", "value2", "arg"}))
			})
			Specify("multiple variables", func() {
				variableResolver.register("var1", STR("value1"))
				variableResolver.register("var2", STR("value2"))
				sentence := firstSentence(parse("cmd $*(var1 var2) arg"))
				value := evaluator.EvaluateSentence(sentence).Value
				Expect(mapValue(value)).To(Equal([]any{"cmd", "value1", "value2", "arg"}))
			})
			Specify("scalar expression", func() {
				commandResolver.register(
					"cmd2",
					functionCommand{func(_ []Value) Value { return STR("value") }},
				)
				sentence := firstSentence(parse("cmd $*[cmd2] arg"))
				value := evaluator.EvaluateSentence(sentence).Value
				Expect(mapValue(value)).To(Equal([]any{"cmd", "value", "arg"}))
			})
			Specify("tuple expression", func() {
				commandResolver.register(
					"cmd2",
					functionCommand{func(_ []Value) Value { return TUPLE([]Value{STR("value1"), STR("value2")}) }},
				)
				sentence := firstSentence(parse("cmd $*[cmd2] arg"))
				value := evaluator.EvaluateSentence(sentence).Value
				Expect(mapValue(value)).To(Equal([]any{"cmd", "value1", "value2", "arg"}))
			})
		})
	})

	Describe("comments", func() {
		Describe("line comments", func() {
			Specify("empty sentence", func() {
				sentence := firstSentence(parse("# this is a comment"))
				value := evaluator.EvaluateSentence(sentence).Value
				Expect(value).To(Equal(NIL))
			})
			Specify("command", func() {
				commandResolver.register(
					"cmd",
					functionCommand{func(args []Value) Value { return TUPLE(args) }},
				)
				sentence := firstSentence(parse("cmd arg # this is a comment"))
				value := evaluator.EvaluateSentence(sentence).Value
				Expect(mapValue(value)).To(Equal([]any{"cmd", "arg"}))
			})
		})
		Describe("block comments", func() {
			Specify("empty sentence", func() {
				sentence := firstSentence(
					parse("#{ this is\na\nblock comment }#"),
				)
				value := evaluator.EvaluateSentence(sentence).Value
				Expect(value).To(Equal(NIL))
			})
			Specify("command", func() {
				commandResolver.register(
					"cmd",
					functionCommand{func(args []Value) Value { return TUPLE(args) }},
				)
				sentence := firstSentence(
					parse("cmd #{ this is\na\nblock comment }# arg"),
				)
				value := evaluator.EvaluateSentence(sentence).Value
				Expect(mapValue(value)).To(Equal([]any{"cmd", "arg"}))
			})
			Specify("tuple", func() {
				word := firstWord(
					parse("(prefix #{ this is\na\nblock comment }# suffix)"),
				)
				value := evaluator.EvaluateWord(word).Value
				Expect(mapValue(value)).To(Equal([]any{"prefix", "suffix"}))
			})
		})
	})

	Describe("scripts", func() {
		Specify("conditional evaluation", func() {
			commandResolver.register(
				"if",
				simpleCommand{func(args []Value) Result {
					condition := args[1]
					var block Value
					if asString(condition) == "true" {
						block = args[2]
					} else {
						block = args[4]
					}
					return evaluator.EvaluateScript(block.(ScriptValue).Script)
				}},
			)
			called := map[string]uint{}
			fn := functionCommand{func(args []Value) Value {
				cmd := asString(args[0])
				called[cmd] += 1
				return args[1]
			}}
			commandResolver.register("cmd1", fn)
			commandResolver.register("cmd2", fn)
			script1 := parse("if true {cmd1 a} else {cmd2 b}")
			value1 := evaluator.EvaluateScript(*script1).Value
			Expect(mapValue(value1)).To(Equal("a"))
			Expect(called).To(Equal(map[string]uint{"cmd1": 1}))
			script2 := parse("if false {cmd1 a} else {cmd2 b}")
			value2 := evaluator.EvaluateScript(*script2).Value
			Expect(mapValue(value2)).To(Equal("b"))
			Expect(called).To(Equal(map[string]uint{"cmd1": 1, "cmd2": 1}))
		})
		Specify("loop", func() {
			commandResolver.register(
				"repeat",
				functionCommand{func(args []Value) Value {
					nb := ValueToInteger(args[1]).Data
					block := args[2]
					var value Value = NIL
					for i := 0; int64(i) < nb; i++ {
						value = evaluator.EvaluateScript(
							block.(ScriptValue).Script,
						).Value
					}
					return value
				}},
			)
			counter := 0
			acc := ""
			commandResolver.register(
				"cmd",
				functionCommand{func(args []Value) Value {
					value := asString(args[1])
					acc += value
					counter++
					return INT(int64(counter - 1))
				}},
			)
			script := parse("repeat 10 {cmd foo}")
			value := evaluator.EvaluateScript(*script).Value
			Expect(mapValue(value)).To(Equal(int64(9)))
			Expect(counter).To(Equal(10))
			Expect(acc).To(Equal(strings.Repeat("foo", 10)))
		})
	})

	Describe("result codes", func() {
		Describe("return", func() {
			It("should interrupt script evaluation", func() {
				commandResolver.register("return", simpleCommand{
					func(args []Value) Result { return RETURN(args[1]) },
				})
				script := parse("return a [return b]; return c")
				result := evaluator.EvaluateScript(*script)
				Expect(result.Code).To(Equal(ResultCode_RETURN))
				Expect(mapValue(result.Value)).To(Equal("b"))
			})
			It("should interrupt sentence evaluation", func() {
				commandResolver.register("cmd", simpleCommand{
					func(_ []Value) Result { return RETURN(STR("value")) },
				})
				commandResolver.register("return", simpleCommand{
					func(args []Value) Result { return RETURN(args[1]) },
				})
				script := parse("cmd [return a [return b]]")
				result := evaluator.EvaluateScript(*script)
				Expect(result.Code).To(Equal(ResultCode_RETURN))
				Expect(mapValue(result.Value)).To(Equal("b"))
			})
			It("should interrupt tuple evaluation", func() {
				commandResolver.register("cmd", simpleCommand{
					func(_ []Value) Result { return OK(STR("value")) },
				})
				commandResolver.register("return", simpleCommand{
					func(args []Value) Result { return RETURN(args[1]) },
				})
				script := parse("cmd ([return a [return b]])")
				result := evaluator.EvaluateScript(*script)
				Expect(result.Code).To(Equal(ResultCode_RETURN))
				Expect(mapValue(result.Value)).To(Equal("b"))
			})
			It("should interrupt expression evaluation", func() {
				commandResolver.register("cmd", simpleCommand{
					func(_ []Value) Result { return RETURN(STR("value")) },
				})
				commandResolver.register("cmd2", simpleCommand{
					func(_ []Value) Result {
						panic("CANTHAPPEN")
					},
				})
				commandResolver.register("return", simpleCommand{
					func(args []Value) Result { return RETURN(args[1]) },
				})
				script := parse("cmd [return a [return b]; cmd2] ")
				result := evaluator.EvaluateScript(*script)
				Expect(result.Code).To(Equal(ResultCode_RETURN))
				Expect(mapValue(result.Value)).To(Equal("b"))
			})
			It("should interrupt keyed selector evaluation", func() {
				commandResolver.register("cmd", simpleCommand{
					func(_ []Value) Result { return RETURN(STR("value")) },
				})
				commandResolver.register("return", simpleCommand{
					func(args []Value) Result { return RETURN(args[1]) },
				})
				variableResolver.register("var", DICT(map[string]Value{"key": STR("value")}))
				script := parse("cmd $var([return a [return b]])")
				result := evaluator.EvaluateScript(*script)
				Expect(result.Code).To(Equal(ResultCode_RETURN))
				Expect(mapValue(result.Value)).To(Equal("b"))
			})
			It("should interrupt indexed selector evaluation", func() {
				commandResolver.register("cmd", simpleCommand{
					func(_ []Value) Result { return RETURN(STR("value")) },
				})
				commandResolver.register("return", simpleCommand{
					func(args []Value) Result { return RETURN(args[1]) },
				})
				variableResolver.register(
					"var",
					LIST([]Value{STR("value1"), STR("value2")}),
				)
				script := parse("cmd $var[return a [return b]]")
				result := evaluator.EvaluateScript(*script)
				Expect(result.Code).To(Equal(ResultCode_RETURN))
				Expect(mapValue(result.Value)).To(Equal("b"))
			})
			It("should interrupt generic selector evaluation", func() {
				commandResolver.register("cmd", simpleCommand{
					func(_ []Value) Result { return RETURN(STR("value")) },
				})
				commandResolver.register("return", simpleCommand{
					func(args []Value) Result { return RETURN(args[1]) },
				})
				variableResolver.register("var", STR("value"))
				script := parse("cmd $var{[return a [return b]]}")
				result := evaluator.EvaluateScript(*script)
				Expect(result.Code).To(Equal(ResultCode_RETURN))
				Expect(mapValue(result.Value)).To(Equal("b"))
			})
		})
	})

	Describe("command context", func() {
		Specify("evaluateScript", func() {
			script := parse("cmd")

			cmd := &captureContextCommand{}
			commandResolver.register("cmd", cmd)

			var context struct{}
			evaluator = NewCompilingEvaluator(
				variableResolver,
				commandResolver,
				selectorResolver,
				context,
			)
			evaluator.EvaluateScript(*script)
			Expect(cmd.context).To(BeIdenticalTo(context))
		})
		Specify("evaluateSentence", func() {
			script := parse("cmd")
			sentence := firstSentence(script)

			cmd := &captureContextCommand{}
			commandResolver.register("cmd", cmd)

			var context struct{}
			evaluator = NewCompilingEvaluator(
				variableResolver,
				commandResolver,
				selectorResolver,
				context,
			)
			evaluator.EvaluateSentence(sentence)
			Expect(cmd.context).To(BeIdenticalTo(context))
		})
		Specify("evaluateWord", func() {
			script := parse("[cmd]")
			word := firstWord(script)

			cmd := &captureContextCommand{}
			commandResolver.register("cmd", cmd)

			var context struct{}
			evaluator = NewCompilingEvaluator(
				variableResolver,
				commandResolver,
				selectorResolver,
				context,
			)
			evaluator.EvaluateWord(word)
			Expect(cmd.context).To(BeIdenticalTo(context))
		})
	})

	Describe("exceptions", func() {
		Specify("invalid command name", func() {
			script := parse("[]")
			Expect(evaluator.EvaluateScript(*script)).To(Equal(
				ERROR("invalid command name"),
			))
		})
		Specify("invalid variable name", func() {
			script := parse("$([])")
			Expect(evaluator.EvaluateScript(*script)).To(Equal(
				ERROR("invalid variable name"),
			))
		})
		Specify("variable substitution with no string representation", func() {
			script := parse(`"$var"`)

			variableResolver.register("var", NIL)

			Expect(evaluator.EvaluateScript(*script)).To(Equal(
				ERROR("value has no string representation"),
			))
		})
		Specify("command substitution with no string representation", func() {
			script := parse(`"[]"`)

			Expect(evaluator.EvaluateScript(*script)).To(Equal(
				ERROR("value has no string representation"),
			))
		})
	})
})
