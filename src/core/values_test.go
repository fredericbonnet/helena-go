package core_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "helena/core"
)

var _ = Describe("values", func() {
	Describe("NIL", func() {
		Specify("type should be NIL", func() {
			Expect(NIL.Type()).To(Equal(ValueType_NIL))
		})
		It("should be displayed as an empty expression", func() {
			Expect(NIL.Display(nil)).To(Equal("[]"))
		})
		It("should not be index-selectable", func() {
			Expect(func() { _ = Value(NIL).(IndexSelectable) }).To(Panic())
			Expect(NewIndexedSelector(NewIntegerValue(1)).Apply(NIL)).To(Equal(
				ERROR("value is not index-selectable"),
			))
		})
		It("should not be key-selectable", func() {
			Expect(func() { _ = Value(NIL).(KeySelectable) }).To(Panic())
			Expect(NewKeyedSelector([]Value{NewStringValue("key")}).Apply(NIL)).To(Equal(
				ERROR("value is not key-selectable"),
			))
		})
		It("should not be selectable", func() {
			Expect(func() { _ = Value(NIL).(Selectable) }).To(Panic())
			Expect(func() { _ = Value(NIL).(RulesSelectable) }).To(Panic())
			Expect(NewGenericSelector([]Value{NewStringValue("rule")}).Apply(NIL)).To(Equal(
				ERROR("value is not selectable"),
			))
		})
	})

	Describe("BooleanValue", func() {
		Specify("type should be BOOLEAN", func() {
			Expect(TRUE.Type()).To(Equal(ValueType_BOOLEAN))
			Expect(FALSE.Type()).To(Equal(ValueType_BOOLEAN))
		})
		It("should be displayed as true or false literals", func() {
			Expect(TRUE.Display(nil)).To(Equal("true"))
			Expect(FALSE.Display(nil)).To(Equal("false"))
		})
		Describe("BooleanValueFromValue()", func() {
			It("should return the passed BooleanValue", func() {
				Expect(BooleanValueFromValue(TRUE).Value).To(Equal(TRUE))
				Expect(BooleanValueFromValue(FALSE).Value).To(Equal(FALSE))
			})
			It("should accept boolean strings", func() {
				Expect(BooleanValueFromValue(NewStringValue("false")).Value).To(Equal(FALSE))
				Expect(BooleanValueFromValue(NewStringValue("true")).Value).To(Equal(TRUE))
			})
			It("should reject non-boolean strings", func() {
				Expect(BooleanValueFromValue(NIL)).To(Equal(
					ERROR_T[BooleanValue]("value has no string representation"),
				))
				Expect(BooleanValueFromValue(NewIntegerValue(0))).To(Equal(
					ERROR_T[BooleanValue](`invalid boolean "0"`),
				))
				Expect(BooleanValueFromValue(NewStringValue("1"))).To(Equal(
					ERROR_T[BooleanValue](`invalid boolean "1"`),
				))
				Expect(BooleanValueFromValue(NewStringValue("no"))).To(Equal(
					ERROR_T[BooleanValue](`invalid boolean "no"`),
				))
				Expect(BooleanValueFromValue(NewStringValue("yes"))).To(Equal(
					ERROR_T[BooleanValue](`invalid boolean "yes"`),
				))
				Expect(BooleanValueFromValue(NewStringValue("a"))).To(Equal(
					ERROR_T[BooleanValue](`invalid boolean "a"`),
				))
			})
		})
		It("should not be index-selectable", func() {
			Expect(func() { _ = Value(TRUE).(IndexSelectable) }).To(Panic())
			Expect(func() { _ = Value(FALSE).(IndexSelectable) }).To(Panic())
			Expect(NewIndexedSelector(NewIntegerValue(1)).Apply(TRUE)).To(Equal(
				ERROR("value is not index-selectable"),
			))
			Expect(NewIndexedSelector(NewIntegerValue(1)).Apply(FALSE)).To(Equal(
				ERROR("value is not index-selectable"),
			))
		})
		It("should not be key-selectable", func() {
			Expect(func() { _ = Value(TRUE).(KeySelectable) }).To(Panic())
			Expect(func() { _ = Value(FALSE).(KeySelectable) }).To(Panic())
			Expect(NewKeyedSelector([]Value{NewStringValue("key")}).Apply(TRUE)).To(Equal(
				ERROR("value is not key-selectable"),
			))
			Expect(NewKeyedSelector([]Value{NewStringValue("key")}).Apply(FALSE)).To(Equal(
				ERROR("value is not key-selectable"),
			))
		})
		It("should not be selectable", func() {
			Expect(func() { _ = Value(TRUE).(Selectable) }).To(Panic())
			Expect(func() { _ = Value(TRUE).(RulesSelectable) }).To(Panic())
			Expect(func() { _ = Value(FALSE).(Selectable) }).To(Panic())
			Expect(func() { _ = Value(FALSE).(RulesSelectable) }).To(Panic())
			Expect(NewGenericSelector([]Value{NewStringValue("rule")}).Apply(TRUE)).To(Equal(
				ERROR("value is not selectable"),
			))
			Expect(
				NewGenericSelector([]Value{NewStringValue("rule")}).Apply(FALSE)).To(Equal(
				ERROR("value is not selectable"),
			))
		})
	})

	Describe("IntegerValue", func() {
		Specify("type should be INTEGER", func() {
			value := NewIntegerValue(123)
			Expect(value.Type()).To(Equal(ValueType_INTEGER))
		})
		It("should be displayed as a literal decimal value", func() {
			const integer = 0x1234
			value := NewIntegerValue(integer)
			Expect(value.Display(nil)).To(Equal("4660"))
		})
		Describe("IntegerValueFromValue()", func() {
			It("should return the passed IntegerValue", func() {
				value := NewIntegerValue(1234)
				Expect(IntegerValueFromValue(value).Value).To(Equal(value))
			})
			It("should accept integer strings", func() {
				value := NewStringValue("1234")
				Expect(IntegerValueFromValue(value).Data.Value).To(Equal(int64(1234)))
			})
			It("should accept round reals", func() {
				value := NewRealValue(1)
				Expect(IntegerValueFromValue(value).Data.Value).To(Equal(int64(1)))
			})
			It("should reject non-integer strings", func() {
				Expect(IntegerValueFromValue(NIL)).To(Equal(
					ERROR_T[IntegerValue]("value has no string representation")),
				)
				Expect(IntegerValueFromValue(NewRealValue(1e100))).To(Equal(
					ERROR_T[IntegerValue](`invalid integer "1e+100"`)),
				)
				Expect(IntegerValueFromValue(NewRealValue(1.1))).To(Equal(
					ERROR_T[IntegerValue](`invalid integer "1.1"`)),
				)
				Expect(IntegerValueFromValue(NewStringValue("a"))).To(Equal(
					ERROR_T[IntegerValue](`invalid integer "a"`)),
				)
				Expect(IntegerValueFromValue(NewStringValue("1.2"))).To(Equal(
					ERROR_T[IntegerValue](`invalid integer "1.2"`)),
				)
			})
		})
		It("should not be index-selectable", func() {
			value := NewIntegerValue(0)
			Expect(func() { _ = Value(value).(IndexSelectable) }).To(Panic())
			Expect(NewIndexedSelector(NewIntegerValue(1)).Apply(value)).To(Equal(
				ERROR("value is not index-selectable"),
			))
		})
		It("should not be key-selectable", func() {
			value := NewIntegerValue(0)
			Expect(func() { _ = Value(value).(KeySelectable) }).To(Panic())
			Expect(NewKeyedSelector([]Value{NewStringValue("key")}).Apply(value)).To(Equal(
				ERROR("value is not key-selectable"),
			))
		})
		It("should not be selectable", func() {
			value := NewIntegerValue(0)
			Expect(func() { _ = Value(value).(Selectable) }).To(Panic())
			Expect(func() { _ = Value(value).(RulesSelectable) }).To(Panic())
			Expect(
				NewGenericSelector([]Value{NewStringValue("rule")}).Apply(value)).To(Equal(
				ERROR("value is not selectable"),
			))
		})
	})

	Describe("RealValue", func() {
		Specify("type should be REAL", func() {
			value := NewRealValue(12.3)
			Expect(value.Type()).To(Equal(ValueType_REAL))
		})
		It("should be displayed as a literal decimal value", func() {
			const real = 123.4
			value := NewRealValue(real)
			Expect(value.Display(nil)).To(Equal("123.4"))
		})
		Describe("RealValueFromValue()", func() {
			It("should return the passed RealValue", func() {
				value := NewRealValue(12.34)
				Expect(RealValueFromValue(value).Value).To(Equal(value))
			})
			It("should accept integer values", func() {
				value := NewIntegerValue(4567)
				Expect(RealValueFromValue(value).Data.Value).To(Equal(float64(4567)))
			})
			It("should accept float strings", func() {
				value := NewStringValue("12.34")
				Expect(RealValueFromValue(value).Data.Value).To(Equal(12.34))
			})
			It("should reject non-number strings", func() {
				Expect(RealValueFromValue(NIL)).To(Equal(
					ERROR_T[RealValue]("value has no string representation"),
				))
				Expect(RealValueFromValue(NewStringValue("a"))).To(Equal(
					ERROR_T[RealValue](`invalid number "a"`),
				))
			})
		})
		It("should not be index-selectable", func() {
			value := NewRealValue(0)
			Expect(func() { _ = Value(value).(IndexSelectable) }).To(Panic())
			Expect(NewIndexedSelector(NewIntegerValue(1)).Apply(value)).To(Equal(
				ERROR("value is not index-selectable"),
			))
		})
		It("should not be key-selectable", func() {
			value := NewRealValue(0)
			Expect(func() { _ = Value(value).(KeySelectable) }).To(Panic())
			Expect(NewKeyedSelector([]Value{NewStringValue("key")}).Apply(value)).To(Equal(
				ERROR("value is not key-selectable"),
			))
		})
		It("should not be selectable", func() {
			value := NewRealValue(0)
			Expect(func() { _ = Value(value).(Selectable) }).To(Panic())
			Expect(func() { _ = Value(value).(RulesSelectable) }).To(Panic())
			Expect(
				NewGenericSelector([]Value{NewStringValue("rule")}).Apply(value)).To(Equal(
				ERROR("value is not selectable"),
			))
		})
	})

	Describe("StringValue", func() {
		Specify("type should be STRING", func() {
			value := StringValue{"some string"}
			Expect(value.Type()).To(Equal(ValueType_STRING))
		})
		Describe("should be displayed as a Helena string", func() {
			Specify("empty string", func() {
				const str = ""
				value := StringValue{str}
				Expect(value.Display(nil)).To(Equal(`""`))
			})
			Specify("simple string", func() {
				const str = "some string"
				value := StringValue{str}
				Expect(value.Display(nil)).To(Equal(`"some string"`))
			})
			Specify("string with special characters", func() {
				const str = `some \"$[${$( $string`
				value := StringValue{str}
				Expect(value.Display(nil)).To(Equal(
					`"some \\\"\$\[\$\{\$\( \$string"`,
				))
			})
		})
		Describe("StringValueFromValue()", func() {
			It("should return the passed StringValue", func() {
				value := NewStringValue("some string")
				Expect(StringValueFromValue(value).Value).To(Equal(value))
			})
			It("should accept booleans as true/false strings", func() {
				Expect(StringValueFromValue(FALSE).Data.Value).To(Equal("false"))
				Expect(StringValueFromValue(TRUE).Data.Value).To(Equal("true"))
				Expect(
					StringValueFromValue(NewBooleanValue(false)).Data.Value,
				).To(Equal("false"))
				Expect(
					StringValueFromValue(NewBooleanValue(true)).Data.Value,
				).To(Equal("true"))
			})
			It("should accept integers as decimal strings", func() {
				value := NewIntegerValue(1234)
				Expect(StringValueFromValue(value).Data.Value).To(Equal("1234"))
			})
			It("should accept reals as decimal strings", func() {
				value := NewRealValue(1.1)
				Expect(StringValueFromValue(value).Data.Value).To(Equal("1.1"))
			})
			It("should accept scripts with source", func() {
				value := NewScriptValue(Script{}, "source")
				Expect(StringValueFromValue(value).Data.Value).To(Equal("source"))
			})
			It("should reject other value types", func() {
				Expect(StringValueFromValue(NIL)).To(Equal(
					ERROR_T[StringValue]("value has no string representation")),
				)
				Expect(StringValueFromValue(NewListValue([]Value{}))).To(Equal(
					ERROR_T[StringValue]("value has no string representation")),
				)
				Expect(StringValueFromValue(NewDictionaryValue(map[string]Value{}))).To(Equal(
					ERROR_T[StringValue]("value has no string representation")),
				)
				Expect(StringValueFromValue(NewTupleValue([]Value{}))).To(Equal(
					ERROR_T[StringValue]("value has no string representation")),
				)
				Expect(StringValueFromValue(NewScriptValueWithNoSource(Script{}))).To(Equal(
					ERROR_T[StringValue]("value has no string representation")),
				)
				Expect(StringValueFromValue(NewQualifiedValue(NewStringValue("name"), []Selector{}))).To(Equal(
					ERROR_T[StringValue]("value has no string representation")),
				)
			})
		})
		Describe("indexed selectors", func() {
			It("should select characters by index", func() {
				const str = "some string"
				value := NewStringValue(str)
				index := NewIntegerValue(2)
				Expect(value.SelectIndex(index)).To(Equal(OK(StringValue{string(str[2])})))
			})
			It("should accept integer strings", func() {
				const str = "some string"
				value := NewStringValue(str)
				index := NewStringValue("2")
				Expect(value.SelectIndex(index)).To(Equal(OK(NewStringValue(string(str[2])))))
			})
			Describe("exceptions", func() {
				Specify("invalid index type", func() {
					value := NewStringValue("some string")
					Expect(value.SelectIndex(NIL)).To(Equal(
						ERROR("value has no string representation"),
					))
				})
				Specify("invalid index value", func() {
					value := NewStringValue("some string")
					index := NewStringValue("foo")
					Expect(value.SelectIndex(index)).To(Equal(
						ERROR(`invalid integer "foo"`),
					))
				})
				Specify("index out of range", func() {
					const str = "some string"
					value := NewStringValue(str)
					Expect(value.SelectIndex(NewIntegerValue(-1))).To(Equal(
						ERROR(`index out of range "-1"`),
					))
					Expect(value.SelectIndex(NewIntegerValue(int64(len(str))))).To(Equal(
						ERROR(`index out of range "` + fmt.Sprint(len(str)) + `"`),
					))
				})
			})
		})
		It("should not be key-selectable", func() {
			value := NewStringValue("some string")
			Expect(func() { _ = Value(value).(KeySelectable) }).To(Panic())
			Expect(NewKeyedSelector([]Value{NewStringValue("key")}).Apply(value)).To(Equal(
				ERROR("value is not key-selectable"),
			))
		})
		It("should not be selectable", func() {
			value := NewStringValue("some string")
			Expect(func() { _ = Value(value).(Selectable) }).To(Panic())
			Expect(func() { _ = Value(value).(RulesSelectable) }).To(Panic())
			Expect(
				NewGenericSelector([]Value{NewStringValue("rule")}).Apply(value)).To(Equal(
				ERROR("value is not selectable"),
			))
		})
	})

	Describe("ListValue", func() {
		Specify("type should be LIST", func() {
			value := NewListValue([]Value{})
			Expect(value.Type()).To(Equal(ValueType_LIST))
		})
		Describe("ListValueFromValue()", func() {
			It("should return the passed ListValue", func() {
				value := NewListValue([]Value{})
				Expect(ListValueFromValue(value).Value).To(Equal(value))
			})
			It("should accept tuples", func() {
				value := NewTupleValue([]Value{
					NewStringValue("a"),
					TRUE,
					NewIntegerValue(10),
				})
				Expect(ListValueFromValue(value).Value).To(Equal(
					NewListValue(value.Values),
				))
			})
			It("should reject other value types", func() {
				Expect(ListValueFromValue(TRUE)).To(Equal(
					ERROR_T[ListValue]("invalid list"),
				))
				Expect(ListValueFromValue(NewStringValue("a"))).To(Equal(
					ERROR_T[ListValue]("invalid list"),
				))
				Expect(ListValueFromValue(NewIntegerValue(10))).To(Equal(
					ERROR_T[ListValue]("invalid list"),
				))
				Expect(ListValueFromValue(NewRealValue(10))).To(Equal(
					ERROR_T[ListValue]("invalid list"),
				))
				Expect(ListValueFromValue(NewScriptValue(Script{}, ""))).To(Equal(
					ERROR_T[ListValue]("invalid list"),
				))
				Expect(ListValueFromValue(NewDictionaryValue(map[string]Value{}))).To(Equal(
					ERROR_T[ListValue]("invalid list"),
				))
			})
		})
		Describe("indexed selectors", func() {
			It("should select elements by index", func() {
				values := []Value{NewStringValue("value1"), NewStringValue("value2")}
				value := NewListValue(values)
				index := NewIntegerValue(1)
				Expect(value.SelectIndex(index)).To(Equal(OK(values[1])))
			})
			It("should accept integer strings", func() {
				values := []Value{NewStringValue("value1"), NewStringValue("value2")}
				value := NewListValue(values)
				index := NewStringValue("0")
				Expect(value.SelectIndex(index)).To(Equal(OK(values[0])))
			})
			Describe("exceptions", func() {
				Specify("invalid index type", func() {
					value := NewListValue([]Value{})
					Expect(value.SelectIndex(NIL)).To(Equal(
						ERROR("value has no string representation"),
					))
				})
				Specify("invalid index value", func() {
					value := NewListValue([]Value{})
					index := NewStringValue("foo")
					Expect(value.SelectIndex(index)).To(Equal(
						ERROR(`invalid integer "foo"`),
					))
				})
				Specify("index out of range", func() {
					values := []Value{NewStringValue("value1"), NewStringValue("value2")}
					value := NewListValue(values)
					Expect(value.SelectIndex(NewIntegerValue(-1))).To(Equal(
						ERROR(`index out of range "-1"`),
					))
					Expect(value.SelectIndex(NewIntegerValue(int64(len(values))))).To(Equal(
						ERROR(`index out of range "` + fmt.Sprint(len(values)) + `"`),
					))
				})
			})
		})
		It("should not be key-selectable", func() {
			value := NewListValue([]Value{})
			Expect(func() { _ = Value(value).(KeySelectable) }).To(Panic())
			Expect(NewKeyedSelector([]Value{NewStringValue("key")}).Apply(value)).To(Equal(
				ERROR("value is not key-selectable"),
			))
		})
		It("should not be selectable", func() {
			value := NewListValue([]Value{})
			Expect(func() { _ = Value(value).(Selectable) }).To(Panic())
			Expect(func() { _ = Value(value).(RulesSelectable) }).To(Panic())
			Expect(
				NewGenericSelector([]Value{NewStringValue("rule")}).Apply(value)).To(Equal(
				ERROR("value is not selectable"),
			))
		})
	})

	Describe("DictionaryValue", func() {
		Specify("type should be DICTIONARY", func() {
			value := NewDictionaryValue(map[string]Value{})
			Expect(value.Type()).To(Equal(ValueType_DICTIONARY))
		})
		It("should not be index-selectable", func() {
			value := NewDictionaryValue(map[string]Value{})
			Expect(func() { _ = Value(value).(IndexSelectable) }).To(Panic())
			Expect(NewIndexedSelector(NewIntegerValue(1)).Apply(value)).To(Equal(
				ERROR("value is not index-selectable"),
			))
		})
		Describe("keyed selectors", func() {
			It("should select elements by key", func() {
				values := map[string]Value{
					"key1": NewStringValue("value1"),
					"key2": NewStringValue("value2"),
				}
				value := NewDictionaryValue(values)
				key := NewStringValue("key1")
				Expect(value.SelectKey(key)).To(Equal(OK(values["key1"])))
			})
			Describe("exceptions", func() {
				Specify("invalid key type", func() {
					value := NewDictionaryValue(map[string]Value{})
					Expect(value.SelectKey(NIL)).To(Equal(ERROR("invalid key")))
				})
				Specify("unknown key value", func() {
					value := NewDictionaryValue(map[string]Value{})
					key := NewStringValue("foo")
					Expect(value.SelectKey(key)).To(Equal(ERROR("unknown key")))
				})
			})
		})
		It("should not be selectable", func() {
			value := NewDictionaryValue(map[string]Value{})
			Expect(func() { _ = Value(value).(Selectable) }).To(Panic())
			Expect(func() { _ = Value(value).(RulesSelectable) }).To(Panic())
			Expect(
				NewGenericSelector([]Value{NewStringValue("rule")}).Apply(value)).To(Equal(
				ERROR("value is not selectable"),
			))
		})
	})

	Describe("TupleValue", func() {
		Specify("type should be TUPLE", func() {
			value := TupleValue{[]Value{}}
			Expect(value.Type()).To(Equal(ValueType_TUPLE))
		})
		Describe("should be displayed as a Helena tuple", func() {
			Specify("empty tuple", func() {
				value := NewTupleValue([]Value{})
				Expect(value.Display(nil)).To(Equal(`()`))
			})
			Specify("simple tuple", func() {
				value := NewTupleValue([]Value{
					NewStringValue("some string"),
					NIL,
					NewIntegerValue(1),
				})
				Expect(value.Display(nil)).To(Equal(`("some string" [] 1)`))
			})
			Specify("undisplayable elements", func() {
				value := NewTupleValue([]Value{
					NewListValue([]Value{}),
					NewDictionaryValue(map[string]Value{}),
				})
				Expect(value.Display(nil)).To(Equal(
					`({#{undisplayable value}#} {#{undisplayable value}#})`,
				))
			})
			Specify("custom function", func() {
				value := NewTupleValue([]Value{
					NewListValue([]Value{}),
					NewDictionaryValue(map[string]Value{}),
				})
				Expect(
					value.Display(func(v any) string {
						switch v.(type) {
						case ListValue:
							return UndisplayableValueWithLabel("ListValue")
						case DictionaryValue:
							return UndisplayableValueWithLabel("DictionaryValue")
						default:
							return UndisplayableValue()
						}
					})).To(Equal(`({#{ListValue}#} {#{DictionaryValue}#})`))
			})
			Specify("recursive tuple", func() {
				value := NewTupleValue([]Value{NewTupleValue([]Value{NewTupleValue([]Value{})})})
				Expect(value.Display(nil)).To(Equal(`((()))`))
			})
		})
		Describe("indexed selectors", func() {
			It("should apply to elements", func() {
				values := []Value{
					NewListValue([]Value{NewStringValue("value1"), NewStringValue("value2")}),
					NewStringValue("12345"),
				}
				value := NewTupleValue(values)
				index := NewIntegerValue(1)
				Expect(value.SelectIndex(index)).To(Equal(
					OK(NewTupleValue([]Value{NewStringValue("value2"), NewStringValue("2")})),
				))
			})
			It("should recurse into tuples", func() {
				values := []Value{
					NewListValue([]Value{NewStringValue("value1"), NewStringValue("value2")}),
					NewTupleValue([]Value{NewStringValue("12345"), NewStringValue("678")}),
				}
				value := NewTupleValue(values)
				index := NewIntegerValue(1)
				Expect(value.SelectIndex(index)).To(Equal(
					OK(
						NewTupleValue([]Value{
							NewStringValue("value2"),
							NewTupleValue([]Value{NewStringValue("2"), NewStringValue("7")}),
						}),
					),
				))
			})
			Describe("exceptions", func() {
				Specify("non-selectable element", func() {
					values := []Value{NewIntegerValue(0)}
					value := NewTupleValue(values)
					index := NewIntegerValue(0)
					Expect(value.SelectIndex(index)).To(Equal(
						ERROR("value is not index-selectable"),
					))
				})
			})
		})
		Describe("keyed selectors", func() {
			It("should apply to elements", func() {
				values := []Value{
					NewDictionaryValue(map[string]Value{
						"key1": NewStringValue("value1"),
						"key2": NewStringValue("value2"),
					}),
					NewDictionaryValue(map[string]Value{
						"key2": NewStringValue("value3"),
						"key3": NewStringValue("value4"),
					}),
				}
				value := NewTupleValue(values)
				key := NewStringValue("key2")
				Expect(value.SelectKey(key)).To(Equal(
					OK(
						NewTupleValue([]Value{
							NewStringValue("value2"),
							NewStringValue("value3"),
						}),
					),
				))
			})
			It("should recurse into tuples", func() {
				values := []Value{
					NewDictionaryValue(map[string]Value{
						"key1": NewStringValue("value1"),
						"key2": NewStringValue("value2"),
					}),
					NewTupleValue([]Value{
						NewDictionaryValue(map[string]Value{
							"key2": NewStringValue("value3"),
							"key3": NewStringValue("value4"),
						}),
						NewDictionaryValue(map[string]Value{
							"key2": NewStringValue("value5"),
							"key4": NewStringValue("value6"),
						}),
					}),
				}
				value := NewTupleValue(values)
				key := NewStringValue("key2")
				Expect(value.SelectKey(key)).To(Equal(
					OK(
						NewTupleValue([]Value{
							NewStringValue("value2"),
							NewTupleValue([]Value{
								NewStringValue("value3"),
								NewStringValue("value5"),
							}),
						}),
					),
				))
			})
			Describe("exceptions", func() {
				Specify("non-selectable element", func() {
					values := []Value{NewIntegerValue(0)}
					value := NewTupleValue(values)
					key := NewStringValue("key2")
					Expect(value.SelectKey(key)).To(Equal(
						ERROR("value is not key-selectable"),
					))
				})
			})
		})
		Describe("Select", func() {
			It("should apply selector to elements", func() {
				values := []Value{
					NewListValue([]Value{NewStringValue("value1"), NewStringValue("value2")}),
					NewStringValue("12345"),
				}
				value := NewTupleValue(values)
				index := NewIntegerValue(1)
				Expect(value.Select(NewIndexedSelector(index))).To(Equal(
					OK(NewTupleValue([]Value{NewStringValue("value2"), NewStringValue("2")})),
				))
			})
			It("should recurse into tuples", func() {
				values := []Value{
					NewDictionaryValue(map[string]Value{
						"key1": NewStringValue("value1"),
						"key2": NewStringValue("value2"),
					}),
					NewTupleValue([]Value{
						NewDictionaryValue(map[string]Value{
							"key2": NewStringValue("value3"),
							"key3": NewStringValue("value4"),
						}),
						NewDictionaryValue(map[string]Value{
							"key2": NewStringValue("value5"),
							"key4": NewStringValue("value6"),
						}),
					}),
				}
				value := NewTupleValue(values)
				key := NewStringValue("key2")
				Expect(value.Select(NewKeyedSelector([]Value{key}))).To(Equal(
					OK(
						NewTupleValue([]Value{
							NewStringValue("value2"),
							NewTupleValue([]Value{
								NewStringValue("value3"),
								NewStringValue("value5"),
							}),
						}),
					),
				))
			})
			Describe("exceptions", func() {
				Specify("non-selectable element", func() {
					values := []Value{NewIntegerValue(0)}
					value := NewTupleValue(values)
					index := NewIntegerValue(1)
					Expect(value.Select(NewIndexedSelector(index))).To(Equal(
						ERROR("value is not index-selectable"),
					))
					key := NewStringValue("key2")
					Expect(value.Select(NewKeyedSelector([]Value{key}))).To(Equal(
						ERROR("value is not key-selectable"),
					))
				})
			})
		})
	})

	Describe("ScriptValue", func() {
		Specify("type should be SCRIPT", func() {
			value := NewScriptValue(Script{}, "")
			Expect(value.Type()).To(Equal(ValueType_SCRIPT))
		})
		Describe("should be displayed as a Helena block", func() {
			Specify("empty script", func() {
				value := NewScriptValue(Script{}, "")
				Expect(value.Display(nil)).To(Equal(`{}`))
			})
			Specify("regular script", func() {
				const script = "cmd arg1 arg2"
				value := NewScriptValue(Script{}, script)
				Expect(value.Display(nil)).To(Equal("{" + script + "}"))
			})
			Specify("script with no source", func() {
				value := NewScriptValueWithNoSource(Script{})
				Expect(value.Display(nil)).To(Equal(`{#{undisplayable script}#}`))
			})
			Specify("custom display function", func() {
				value := NewScriptValueWithNoSource(Script{})
				Expect(value.Display(func(_ any) string { return "{}" })).To(Equal(`{}`))
			})
		})
		It("should not be index-selectable", func() {
			value := NewScriptValue(Script{}, "")
			Expect(func() { _ = Value(value).(IndexSelectable) }).To(Panic())
			Expect(NewIndexedSelector(NewIntegerValue(1)).Apply(value)).To(Equal(
				ERROR("value is not index-selectable"),
			))
		})
		It("should not be key-selectable", func() {
			value := NewScriptValue(Script{}, "")
			Expect(func() { _ = Value(value).(KeySelectable) }).To(Panic())
			Expect(NewKeyedSelector([]Value{NewStringValue("key")}).Apply(value)).To(Equal(
				ERROR("value is not key-selectable"),
			))
		})
		It("should not be selectable", func() {
			value := NewScriptValue(Script{}, "")
			Expect(func() { _ = Value(value).(Selectable) }).To(Panic())
			Expect(func() { _ = Value(value).(RulesSelectable) }).To(Panic())
			Expect(
				NewGenericSelector([]Value{NewStringValue("rule")}).Apply(value)).To(Equal(
				ERROR("value is not selectable"),
			))
		})
	})

	Describe("CommandValue", func() {
		Specify("type should be COMMAND", func() {
			value := NewCommandValue(simpleCommand{func(args []Value) Result {
				return OK(NIL)
			}})
			Expect(value.Type()).To(Equal(ValueType_COMMAND))
		})
		It("should not be index-selectable", func() {
			value := NewCommandValue(simpleCommand{func(args []Value) Result {
				return OK(NIL)
			}})
			Expect(func() { _ = Value(TRUE).(IndexSelectable) }).To(Panic())
			Expect(NewIndexedSelector(NewIntegerValue(1)).Apply(value)).To(Equal(
				ERROR("value is not index-selectable"),
			))
		})
		It("should not be key-selectable", func() {
			value := NewCommandValue(simpleCommand{func(args []Value) Result {
				return OK(NIL)
			}})
			Expect(func() { _ = Value(TRUE).(KeySelectable) }).To(Panic())
			Expect(NewKeyedSelector([]Value{NewStringValue("key")}).Apply(value)).To(Equal(
				ERROR("value is not key-selectable"),
			))
		})
		It("should not be selectable", func() {
			value := NewCommandValue(simpleCommand{func(args []Value) Result {
				return OK(NIL)
			}})
			Expect(func() { _ = Value(NIL).(Selectable) }).To(Panic())
			Expect(func() { _ = Value(TRUE).(RulesSelectable) }).To(Panic())
			Expect(
				NewGenericSelector([]Value{NewStringValue("rule")}).Apply(value),
			).To(Equal(ERROR("value is not selectable")))
		})
	})

	Describe("QualifiedValue", func() {
		Specify("type should be QUALIFIED", func() {
			value := NewQualifiedValue(NewStringValue("name"), []Selector{})
			Expect(value.Type()).To(Equal(ValueType_QUALIFIED))
		})
		Describe("should be displayed as a Helena qualified word", func() {
			Specify("indexed selectors", func() {
				value := NewQualifiedValue(NewStringValue("name"), []Selector{
					NewIndexedSelector(NewStringValue("index")),
				})

				Expect(value.Display(nil)).To(Equal("name[index]"))
			})
			Specify("keyed selectors", func() {
				value := NewQualifiedValue(NewStringValue("name"), []Selector{
					NewKeyedSelector([]Value{NewStringValue("key1"), NewStringValue("key2")}),
				})

				Expect(value.Display(nil)).To(Equal(`name(key1 key2)`))
			})
			Specify("generic selector", func() {
				value := NewQualifiedValue(NewStringValue("name"), []Selector{
					NewGenericSelector([]Value{
						NewStringValue("rule1"),
						NewTupleValue([]Value{NewStringValue("rule2"), NewIntegerValue(123)}),
					}),
				})

				Expect(value.Display(nil)).To(Equal(`name{rule1; rule2 123}`))
			})
			Specify("custom selector", func() {
				value := NewQualifiedValue(NewStringValue("name"), []Selector{
					undisplayableSelector{},
					displayableSelector{},
				})

				Expect(value.Display(nil)).To(Equal(
					`name{#{undisplayable value}#}{foo bar}`,
				))
				Expect(value.Display(func(_ any) string { return UndisplayableValueWithLabel("baz sprong") })).To(Equal(
					`name{#{baz sprong}#}{foo bar}`,
				))
			})
			Specify("source with special characters", func() {
				value := NewQualifiedValue(
					NewStringValue(`some # \"$[${$( $string`),
					[]Selector{
						NewKeyedSelector([]Value{
							NewStringValue("key1"),
							NewStringValue("key2"),
						}),
					},
				)

				Expect(value.Display(nil)).To(Equal(
					`{some \# \\\"\$\[\$\{\$\( \$string}(key1 key2)`,
				))
			})
		})
		Describe("indexed selectors", func() {
			It("should return a new qualified value", func() {
				value := NewQualifiedValue(NewStringValue("name"), []Selector{})
				Expect(value.SelectIndex(NewStringValue("index"))).To(Equal(
					OK(
						NewQualifiedValue(NewStringValue("name"), []Selector{
							NewIndexedSelector(NewStringValue("index")),
						}),
					),
				))
			})
		})
		Describe("keyed selectors", func() {
			It("should return a new qualified value", func() {
				value := NewQualifiedValue(NewStringValue("name"), []Selector{})
				Expect(value.SelectKey(NewStringValue("key"))).To(Equal(
					OK(
						NewQualifiedValue(NewStringValue("name"), []Selector{
							NewKeyedSelector([]Value{NewStringValue("key")}),
						}),
					),
				))
			})
			It("should aggregate keys", func() {
				value := NewQualifiedValue(NewStringValue("name"), []Selector{
					NewKeyedSelector([]Value{NewStringValue("key1"), NewStringValue("key2")}),
					NewIndexedSelector(NewStringValue("index")),
					NewKeyedSelector([]Value{NewStringValue("key3"), NewStringValue("key4")}),
				})
				Expect(value.SelectKey(NewStringValue("key5"))).To(Equal(
					OK(
						NewQualifiedValue(NewStringValue("name"), []Selector{
							NewKeyedSelector([]Value{
								NewStringValue("key1"),
								NewStringValue("key2"),
							}),
							NewIndexedSelector(NewStringValue("index")),
							NewKeyedSelector([]Value{
								NewStringValue("key3"),
								NewStringValue("key4"),
								NewStringValue("key5"),
							}),
						}),
					),
				))
			})
		})
		Describe("generic selectors", func() {
			It("should return a new qualified value", func() {
				value := NewQualifiedValue(NewStringValue("name"), []Selector{})
				Expect(value.SelectRules([]Value{NewStringValue("rule")})).To(Equal(
					OK(
						NewQualifiedValue(NewStringValue("name"), []Selector{
							NewGenericSelector([]Value{NewStringValue("rule")}),
						}),
					),
				))
			})
		})
		Describe("select", func() {
			It("should return a new qualified value", func() {
				value := NewQualifiedValue(NewStringValue("name"), []Selector{})
				selector := testSelector{}
				Expect(value.Select(selector)).To(Equal(
					OK(NewQualifiedValue(NewStringValue("name"), []Selector{selector})),
				))
			})
		})
	})
})

type undisplayableSelector struct{}

func (selector undisplayableSelector) Apply(value Value) Result {
	return ERROR("not implemented")
}

type displayableSelector struct{}

func (selector displayableSelector) Apply(_ Value) Result {
	return ERROR("not implemented")
}
func (selector displayableSelector) Display(_ DisplayFunction) string {
	return "{foo bar}"
}

type testSelector struct{}

func (selector testSelector) Apply(_ Value) Result {
	return OK(NewStringValue("value"))
}
