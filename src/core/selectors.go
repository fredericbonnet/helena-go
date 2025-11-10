//
// Helena value selectors
//

package core

//
// Helena selector
//
// Selectors apply to values to access their subvalues
//
type Selector interface {
	// Apply selector to value and return selected subvalue
	Apply(value Value) Result
}

//
// Indexed selector
//
// Indexed selectors delegate to Value.SelectIndex. They typically apply to
// linear collections with integer indexes, though Helena makes no assumption on
// the actual type of the index value. For example a 2D matrix value could
// accept a pair of (column, row) integers to select one of its cells.
//
type IndexedSelector struct {
	// Index to select
	Index Value
}

// Constructor with index to select
func newIndexedSelector(index Value) IndexedSelector {
	if index == NIL {
		panic("invalid index")
	}
	return IndexedSelector{index}
}

// Factory function, returns a result instead of panicking like the constructor.
func CreateIndexedSelector(index Value) (Result, IndexedSelector) {
	if index == NIL {
		return ERROR("invalid index"), IndexedSelector{}
	}
	return OK(NIL), newIndexedSelector(index)
}

func (selector IndexedSelector) Apply(value Value) Result {
	switch v := value.(type) {
	default:
		return ERROR("value is not index-selectable")
	case IndexSelectable:
		return v.SelectIndex(selector.Index)
	}
}

func (selector IndexedSelector) Display(fn DisplayFunction) string {
	if fn == nil {
		fn = DefaultDisplayFunction
	}
	return `[` + Display(selector.Index, fn) + `]`
}

//
// Keyed selector
//
// Keyed selectors delegate to {@link Value.selectKey}. They typically apply
// to key-value collections. Key types are arbitrary and the selection semantics
// is the target value responsibility.
//
type KeyedSelector struct {
	// Keys to select in order
	Keys []Value
}

// Constructor with keys to select
func newKeyedSelector(keys []Value) KeyedSelector {
	if len(keys) == 0 {
		panic("empty selector")
	}
	return KeyedSelector{keys}
}

// Factory function, returns a result instead of panicking like the constructor.
func CreateKeyedSelector(keys []Value) (Result, KeyedSelector) {
	if len(keys) == 0 {
		return ERROR("empty selector"), KeyedSelector{}
	}
	return OK(NIL), newKeyedSelector(keys)
}

func (selector KeyedSelector) Apply(value Value) Result {
	for _, key := range selector.Keys {
		switch v := value.(type) {
		default:
			return ERROR("value is not key-selectable")
		case KeySelectable:
			result := v.SelectKey(key)
			if result.Code != ResultCode_OK {
				return result
			}
			value = result.Value
		}
	}
	return OK(value)
}

func (selector KeyedSelector) Display(fn DisplayFunction) string {
	if fn == nil {
		fn = DefaultDisplayFunction
	}
	return `(` + DisplayList(selector.Keys, fn) + `)`
}

//
// Generic selector
//
// Generic selectors delegate to {@link Value.selectRules}. They apply a set of
// rules to any kind of value. Each rule is a tuple of values.
//
type GenericSelector struct {
	// Rules to apply
	Rules []Value
}

// Constructor with rules to apply
func newGenericSelector(rules []Value) GenericSelector {
	if len(rules) == 0 {
		panic("empty selector")
	}
	return GenericSelector{rules}
}

// Factory function, returns a result instead of panicking like the constructor.
func CreateGenericSelector(rules []Value) (Result, GenericSelector) {
	if len(rules) == 0 {
		return ERROR("empty selector"), GenericSelector{}
	}
	return OK(NIL), newGenericSelector(rules)
}

func (selector GenericSelector) Apply(value Value) Result {
	return SelectGeneric(value, selector)
}

func (selector GenericSelector) Display(fn DisplayFunction) string {
	if fn == nil {
		fn = DefaultDisplayFunction
	}
	var output = ""
	for i, rule := range selector.Rules {
		if i != 0 {
			output += "; "
		}
		if rule.Type() == ValueType_TUPLE {
			output += DisplayList(rule.(TupleValue).Values, fn)
		} else {
			output += Display(rule, fn)
		}
	}
	return `{` + output + `}`
}

// Apply a selector to a value
func ApplySelector(value Value, selector Selector) Result {
	switch v := value.(type) {
	case Selectable:
		return v.Select(selector)
	default:
		return selector.Apply(value)
	}
}

// Select value with selector using either Value.Select or Value.SelectRules in
// this order of precedence.
func SelectGeneric(value Value, selector GenericSelector) Result {
	switch v := value.(type) {
	default:
		return ERROR("value is not selectable")
	case Selectable:
		return v.Select(Selector(selector))
	case RulesSelectable:
		return v.SelectRules(selector.Rules)
	}
}
