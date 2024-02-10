package core_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "helena/core"
)

type mockValue struct {
	store *mockValueStore
}
type mockValueStore struct {
	selectedIndex *Value
	selectedKeys  []Value
	selectedRules []Value
}

func newMockValue() mockValue {
	return mockValue{&mockValueStore{}}
}

func (value mockValue) Type() ValueType {
	//TODO
	//   type = { name: "mock" };
	return ValueType_NIL
}

func (value mockValue) SelectIndex(index Value) Result {
	value.store.selectedIndex = &index
	return OK(value)
}

func (value mockValue) SelectKey(key Value) Result {
	value.store.selectedKeys = append(value.store.selectedKeys, key)
	return OK(value)
}

func (value mockValue) SelectRules(rules []Value) Result {
	value.store.selectedRules = rules
	return OK(value)
}

type unselectableValue struct{}

func (value unselectableValue) Type() ValueType {
	//TODO
	//   type = { name: "unselectable" };
	return ValueType_NIL
}

var _ = Describe("selectors", func() {
	Describe("IndexedSelector", func() {
		Specify("literal index", func() {
			index := STR("index")
			selector := NewIndexedSelector(index)
			value := newMockValue()
			Expect(selector.Apply(value)).To(Equal(OK(value)))
			Expect(*value.store.selectedIndex).To(Equal(index))
		})
		Describe("display", func() {
			Specify("simple index", func() {
				index := STR("index")
				selector := NewIndexedSelector(index)
				Expect(selector.Display(nil)).To(Equal("[index]"))
			})
			Specify("index with special characters", func() {
				index := STR(`index with spaces and \"$[${$( $special characters`)
				selector := NewIndexedSelector(index)
				Expect(selector.Display(nil)).To(Equal(
					`["index with spaces and \\\"\$\[\$\{\$\( \$special characters"]`,
				))
			})
		})
		Describe("exceptions", func() {
			Specify("invalid index", func() {
				Expect(func() { _ = NewIndexedSelector(NIL) }).To(PanicWith("invalid index"))
				Expect(CreateIndexedSelector(NIL)).To(Equal(ERROR_T[Selector]("invalid index")))
			})
			Specify("non-selectable value", func() {
				selector := NewIndexedSelector(INT(1))
				value := unselectableValue{}
				Expect(selector.Apply(value)).To(Equal(
					ERROR("value is not index-selectable"),
				))
			})
		})
	})

	Describe("KeyedSelector", func() {
		Specify("one key", func() {
			keys := []Value{STR("key")}
			selector := NewKeyedSelector(keys)
			value := newMockValue()
			Expect(selector.Apply(value)).To(Equal(OK(value)))
			Expect(value.store.selectedKeys).To(Equal(keys))
		})
		Specify("multiple keys", func() {
			keys := []Value{STR("key1"), STR("key2")}
			selector := NewKeyedSelector(keys)
			value := newMockValue()
			Expect(selector.Apply(value)).To(Equal(OK(value)))
			Expect(value.store.selectedKeys).To(Equal(keys))
		})
		Describe("display", func() {
			Specify("simple key", func() {
				keys := []Value{STR("key")}
				selector := NewKeyedSelector(keys)
				Expect(selector.Display(nil)).To(Equal("(key)"))
			})
			Specify("multiple keys", func() {
				keys := []Value{STR("key1"), STR("key2")}
				selector := NewKeyedSelector(keys)
				Expect(selector.Display(nil)).To(Equal("(key1 key2)"))
			})
			Specify("key with special characters", func() {
				keys := []Value{STR(`key with spaces and \"$[${$( $special characters`)}
				selector := NewKeyedSelector(keys)
				Expect(selector.Display(nil)).To(Equal(
					`("key with spaces and \\\"\$\[\$\{\$\( \$special characters")`,
				))
			})
		})
		Describe("exceptions", func() {
			Specify("empty key list", func() {
				Expect(func() { _ = NewKeyedSelector([]Value{}) }).To(PanicWith("empty selector"))
				Expect(CreateKeyedSelector([]Value{})).To(Equal(ERROR_T[Selector]("empty selector")))
			})
			Specify("non-selectable value", func() {
				selector := NewKeyedSelector([]Value{INT(1)})
				value := unselectableValue{}
				Expect(selector.Apply(value)).To(Equal(
					ERROR("value is not key-selectable"),
				))
			})
		})
	})

	Describe("GenericSelector", func() {
		Specify("string rule", func() {
			rules := []Value{STR("rule")}
			selector := NewGenericSelector(rules)
			value := newMockValue()
			Expect(selector.Apply(value)).To(Equal(OK(value)))
			Expect(value.store.selectedRules).To(Equal(rules))
		})
		Specify("tuple rule", func() {
			rules := []Value{TUPLE([]Value{STR("rule"), INT(1)})}
			selector := NewGenericSelector(rules)
			value := newMockValue()
			Expect(selector.Apply(value)).To(Equal(OK(value)))
			Expect(value.store.selectedRules).To(Equal(rules))
		})
		Specify("multiple rules", func() {
			rules := []Value{STR("rule1"), TUPLE([]Value{STR("rule2")})}
			selector := NewGenericSelector(rules)
			value := newMockValue()
			Expect(selector.Apply(value)).To(Equal(OK(value)))
			Expect(value.store.selectedRules).To(Equal(rules))
		})
		Describe("display", func() {
			Specify("string rule", func() {
				rules := []Value{STR("rule")}
				selector := NewGenericSelector(rules)
				Expect(selector.Display(nil)).To(Equal("{rule}"))
			})
			Specify("tuple rule", func() {
				rules := []Value{TUPLE([]Value{STR("rule"), INT(1)})}
				selector := NewGenericSelector(rules)
				Expect(selector.Display(nil)).To(Equal("{rule 1}"))
			})
			Specify("multiple keys", func() {
				rules := []Value{
					TUPLE([]Value{
						STR("rule1"),
						STR(`arg1 with spaces and \"$[${$( $special ; characters`),
					}),
					STR("rule2 with spaces"),
					TUPLE([]Value{STR("rule3")}),
				}
				selector := NewGenericSelector(rules)
				Expect(selector.Display(nil)).To(Equal(
					`{rule1 "arg1 with spaces and \\\"\$\[\$\{\$\( \$special ; characters"; "rule2 with spaces"; rule3}`,
				))
			})
		})
		Describe("exceptions", func() {
			Specify("empty rules", func() {
				Expect(func() { _ = NewGenericSelector([]Value{}) }).To(PanicWith("empty selector"))
				Expect(CreateGenericSelector([]Value{})).To(Equal(ERROR_T[Selector]("empty selector")))
			})
			Specify("non-selectable value", func() {
				selector := NewGenericSelector([]Value{STR("rule")})
				value := unselectableValue{}
				Expect(selector.Apply(value)).To(Equal(ERROR("value is not selectable")))
			})
		})
	})

	Specify("custom selectors", func() {
		selector := newCustomSelector("custom")
		value := newMockValue()
		Expect(selector.Apply(value)).To(Equal(OK(TUPLE([]Value{STR("custom"), value}))))
	})
})

type customSelector struct {
	name string
}

func newCustomSelector(name string) customSelector {
	return customSelector{name}
}
func (selector customSelector) Apply(value Value) Result {
	return OK(TUPLE([]Value{STR(selector.name), value}))
}
