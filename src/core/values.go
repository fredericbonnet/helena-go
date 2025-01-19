//
// Helena values
//

package core

import (
	"fmt"
	"strconv"
)

// Helena standard value types
type ValueType int8

const (
	ValueType_NIL ValueType = iota
	ValueType_BOOLEAN
	ValueType_INTEGER
	ValueType_REAL
	ValueType_STRING
	ValueType_LIST
	ValueType_DICTIONARY
	ValueType_TUPLE
	ValueType_SCRIPT
	ValueType_COMMAND
	ValueType_QUALIFIED
	ValueType_CUSTOM
)

func (t ValueType) String() string {
	switch t {
	case ValueType_NIL:
		return "NIL"
	case ValueType_BOOLEAN:
		return "BOOLEAN"
	case ValueType_INTEGER:
		return "INTEGER"
	case ValueType_REAL:
		return "REAL"
	case ValueType_STRING:
		return "STRING"
	case ValueType_LIST:
		return "LIST"
	case ValueType_DICTIONARY:
		return "DICTIONARY"
	case ValueType_TUPLE:
		return "TUPLE"
	case ValueType_SCRIPT:
		return "SCRIPT"
	case ValueType_COMMAND:
		return "COMMAND"
	case ValueType_QUALIFIED:
		return "QUALIFIED"
	case ValueType_CUSTOM:
		return "CUSTOM"
	default:
		panic("CANTHAPPEN")
	}
}

// Helena value
type Value interface {
	// Type identifier
	Type() ValueType
}

type IndexSelectable interface {
	// Select value at index
	SelectIndex(index Value) Result
}

type KeySelectable interface {
	// Select value at key
	SelectKey(key Value) Result
}

type Selectable interface {
	// Select value with selector. When present, this takes precedence over
	// RulesSelectable
	//
	// Note: Implementations must not call Selector.Apply else this would
	// result in an infinite loop
	Select(selector Selector) Result
}

type RulesSelectable interface {
	// Select value from rules
	SelectRules(rules []Value) Result
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

//
// Nil value
//

type NilValue struct {
}

func (value NilValue) Type() ValueType {
	return ValueType_NIL
}
func (value NilValue) Display(_ DisplayFunction) string {
	return "[]"
}

// Singleton nil value
var NIL = NilValue{}

//
// Boolean value
//

type BooleanValue struct {
	// Encapsulated value
	Value bool
}

func (value BooleanValue) Type() ValueType {
	return ValueType_BOOLEAN
}

// Constructor with boolean value to encapsulate
func NewBooleanValue(value bool) BooleanValue {
	return BooleanValue{value}
}

// Convert value to BooleanValue
func BooleanValueFromValue(value Value) (Result, BooleanValue) {
	if value.Type() == ValueType_BOOLEAN {
		return OK(value), BooleanValue(value.(BooleanValue))
	}
	result, b := ValueToBoolean(value)
	if result.Code != ResultCode_OK {
		return result, FALSE
	}
	if b {
		return OK(TRUE), TRUE
	} else {
		return OK(FALSE), FALSE
	}
}

// Convert value to boolean:
// - Booleans: use boolean value
// - Strings: true, false
func ValueToBoolean(value Value) (Result, bool) {
	if value.Type() == ValueType_BOOLEAN {
		return OK(NIL), BooleanValue(value.(BooleanValue)).Value
	}
	result, s := ValueToString(value)
	if result.Code != ResultCode_OK {
		return result, false
	}
	if s == "true" {
		return OK(NIL), true
	}
	if s == "false" {
		return OK(NIL), false
	}
	return ERROR(`invalid boolean "` + s + `"`), false
}

func (value BooleanValue) Display(_ DisplayFunction) string {
	if value.Value {
		return "true"
	} else {
		return "false"
	}
}

// Singleton true value
var TRUE = BooleanValue{true}

// Singleton false value
var FALSE = BooleanValue{false}

//
// Integer value
//

type IntegerValue struct {
	// Encapsulated value
	Value int64
}

func (value IntegerValue) Type() ValueType {
	return ValueType_INTEGER
}

// Constructor with integer value to encapsulate
func NewIntegerValue(value int64) IntegerValue {
	return IntegerValue{value}
}

// Convert value to IntegerValue
func IntegerValueFromValue(value Value) (Result, IntegerValue) {
	if value.Type() == ValueType_INTEGER {
		return OK(value), value.(IntegerValue)
	}
	result, i := ValueToInteger(value)
	if result.Code != ResultCode_OK {
		return result, IntegerValue{}
	}
	v := NewIntegerValue(i)
	return OK(v), v
}

// Report whether string value is convertible to integer
func StringIsInteger(value string) bool {
	_, err := strconv.ParseInt(value, 10, 64)
	return err == nil
}

// Convert value to integer:
// - Integers: use integer value
// - Reals: any safe integer number
// - Strings: any integer Number()-accepted string
func ValueToInteger(value Value) (Result, int64) {
	if value.Type() == ValueType_INTEGER {
		return OK(NIL), value.(IntegerValue).Value
	}
	// TODO: is it needed? strconv.ParseInt would work but maybe converting int to float is more efficient
	// if (value.Type() == ValueType_REAL) {
	//   if (!Number.isSafeInteger((value as RealValue).value))
	//     return ERROR(`invalid integer "${(value as RealValue).value}"`);
	//   return OK(NIL, (value as RealValue).value);
	// }
	result, s := ValueToString(value)
	if result.Code != ResultCode_OK {
		return result, 0
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return ERROR(`invalid integer "` + s + `"`), 0
	}
	return OK(NIL), n
}

func (value IntegerValue) Display(_ DisplayFunction) string {
	return fmt.Sprint(value.Value)
}

//
// Real value
//

type RealValue struct {
	// Encapsulated value
	Value float64
}

func (value RealValue) Type() ValueType {
	return ValueType_REAL
}

// Constructor with float value to encapsulate
func NewRealValue(value float64) RealValue {
	return RealValue{value}
}

// Convert value to RealValue
func RealValueFromValue(value Value) (Result, RealValue) {
	if value.Type() == ValueType_REAL {
		return OK(value), value.(RealValue)
	}
	if value.Type() == ValueType_INTEGER {
		v := NewRealValue(float64(value.(IntegerValue).Value))
		return OK(v), v
	}
	result, f := ValueToFloat(value)
	if result.Code != ResultCode_OK {
		return result, RealValue{}
	}
	v := NewRealValue(f)
	return OK(v), v
}

// Report whether string value is convertible to number
func StringIsNumber(value string) bool {
	_, err := strconv.ParseFloat(value, 64)
	return err == nil
}

// Convert value to float:
// - Reals: use float value
// - Integers: use int value
// - Strings: any strconv.ParseFloat()-accepted string
func ValueToFloat(value Value) (Result, float64) {
	if value.Type() == ValueType_REAL {
		return OK(NIL), value.(RealValue).Value
	}
	if value.Type() == ValueType_INTEGER {
		return OK(NIL), float64(value.(IntegerValue).Value)
	}
	result, s := ValueToString(value)
	if result.Code != ResultCode_OK {
		return result, 0
	}
	n, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return ERROR(`invalid number "` + s + `"`), 0
	}
	return OK(NIL), n
}

func (value RealValue) Display(_ DisplayFunction) string {
	return fmt.Sprint(value.Value)
}

//
// String value
//

type StringValue struct {
	// Encapsulated value
	Value string
}

// Constructor with string value to encapsulate
func NewStringValue(value string) StringValue {
	return StringValue{value}
}

func (value StringValue) Type() ValueType {
	return ValueType_STRING
}

// Convert value to StringValue
func StringValueFromValue(value Value) (Result, StringValue) {
	if value.Type() == ValueType_STRING {
		return OK(value), value.(StringValue)
	}
	result, s := ValueToString(value)
	if result.Code != ResultCode_OK {
		return result, StringValue{}
	}
	v := NewStringValue(s)
	return OK(v), v
}

// Convert value to string
func ValueToString(value Value) (Result, string) {
	return valueToString(value, nil)
}

// Convert value to string, or default value if value has no string representation
func ValueToStringOrDefault(value Value, def string) (Result, string) {
	return valueToString(value, &def)
}

func valueToString(value Value, def *string) (Result, string) {
	switch value.Type() {
	case ValueType_STRING:
		return OK(NIL), value.(StringValue).Value
	case ValueType_BOOLEAN:
		if value.(BooleanValue).Value {
			return OK(NIL), "true"
		} else {
			return OK(NIL), "false"
		}
	case ValueType_INTEGER:
		return OK(NIL), fmt.Sprint(value.(IntegerValue).Value)
	case ValueType_REAL:
		return OK(NIL), fmt.Sprint(value.(RealValue).Value)
	case ValueType_SCRIPT:
		{
			source := value.(ScriptValue).Source
			if source != nil {
				return OK(NIL), *source
			}
		}
	}
	if def != nil {
		return OK(NIL), *def
	}
	return ERROR("value has no string representation"), ""
}

// Return index-th string character as StringValue
func StringAt(value string, index Value) Result {
	return StringAtOrDefault(value, index, nil)
}

// Return index-th string character as StringValue, or default value for
// out-of-range index
func StringAtOrDefault(value string, index Value, def Value) Result {
	result, i := ValueToInteger(index)
	if result.Code != ResultCode_OK {
		return result
	}
	if i < 0 || i >= int64(len(value)) {
		if def != nil {
			return OK(def)
		} else {
			return ERROR(`index out of range "` + fmt.Sprint(i) + `"`)
		}
	}
	return OK(StringValue{string(value[i])})
}

func (value StringValue) Display(_ DisplayFunction) string {
	return DisplayLiteralOrString(value.Value)
}

func (value StringValue) SelectIndex(index Value) Result {
	return StringAt(value.Value, index)
}

//
// List value
//
// Lists are linear collections of other values
//

type ListValue struct {
	// Encapsulated values
	Values []Value
}

func (value ListValue) Type() ValueType {
	return ValueType_LIST
}

// Constructor with array of values to encapsulate
func NewListValue(values []Value) ListValue {
	return ListValue{values}
}

// Convert value to ListValue
func ListValueFromValue(value Value) (Result, ListValue) {
	if value.Type() == ValueType_LIST {
		return OK(value), value.(ListValue)
	}
	result, l := ValueToValues(value)
	if result.Code != ResultCode_OK {
		return result, ListValue{}
	}
	v := NewListValue(l)
	return OK(v), v
}

// Convert value to array of values:
// - Lists
// - Tuples
func ValueToValues(value Value) (Result, []Value) {
	switch value.Type() {
	case ValueType_LIST:
		return OK(NIL), value.(ListValue).Values
	case ValueType_TUPLE:
		return OK(NIL), value.(TupleValue).Values
	default:
		return ERROR("invalid list"), nil
	}
}

// Return index-th element in list
func ListAt(values []Value, index Value) Result {
	return ListAtOrDefault(values, index, nil)
}

// Return index-th element in list, or default value for out-of-range index
func ListAtOrDefault(values []Value, index Value, def Value) Result {
	result, i := ValueToInteger(index)
	if result.Code != ResultCode_OK {
		return result
	}
	if i < 0 || i >= int64(len(values)) {
		if def != nil {
			return OK(def)
		} else {
			return ERROR(`index out of range "` + fmt.Sprint(i) + `"`)
		}
	}
	return OK(values[i])
}

func (value ListValue) SelectIndex(index Value) Result {
	return ListAt(value.Values, index)
}

//
// Dictionary value
//
// Dictionaries are key-value collections with string keys
//

type DictionaryValue struct {
	// Encapsulated key-value map
	Map map[string]Value
}

func (value DictionaryValue) Type() ValueType {
	return ValueType_DICTIONARY
}

// Constructor with key-value map to encapsulate
func NewDictionaryValue(value map[string]Value) DictionaryValue {
	return DictionaryValue{value}
}

func (value DictionaryValue) SelectKey(key Value) Result {
	result, s := ValueToString(key)
	if result.Code != ResultCode_OK {
		return ERROR("invalid key")
	}
	v, ok := value.Map[s]
	if !ok {
		return ERROR("unknown key")
	}
	return OK(v)
}

//
// Tuple value
//
// Tuples are syntactic constructs in Helena. Selectors apply recursively to
// their elements.
//

type TupleValue struct {
	// Encapsulated values
	Values []Value
}

func (value TupleValue) Type() ValueType {
	return ValueType_TUPLE
}

// Constructor from array of values to encapsulate
func NewTupleValue(values []Value) TupleValue {
	return TupleValue{values}
}

func (value TupleValue) SelectIndex(index Value) Result {
	values := make([]Value, len(value.Values))
	for i, v := range value.Values {
		switch vv := v.(type) {
		default:
			return ERROR("value is not index-selectable")
		case IndexSelectable:
			result := vv.SelectIndex(index)
			if result.Code != ResultCode_OK {
				return result
			}
			values[i] = result.Value
		}
	}
	return OK(NewTupleValue(values))
}

func (value TupleValue) SelectKey(key Value) Result {
	values := make([]Value, len(value.Values))
	for i, v := range value.Values {
		switch vv := v.(type) {
		default:
			return ERROR("value is not key-selectable")
		case KeySelectable:
			result := vv.SelectKey(key)
			if result.Code != ResultCode_OK {
				return result
			}
			values[i] = result.Value
		}
	}
	return OK(NewTupleValue(values))
}

func (value TupleValue) Select(selector Selector) Result {
	values := make([]Value, len(value.Values))
	for i, v := range value.Values {
		result := ApplySelector(v, selector)
		if result.Code != ResultCode_OK {
			return result
		}
		values[i] = result.Value
	}
	return OK(NewTupleValue(values))
}

func (value TupleValue) Display(fn DisplayFunction) string {
	if fn == nil {
		fn = DefaultDisplayFunction
	}
	return "(" + DisplayList(value.Values, fn) + ")"
}

//
// Script value
//
// Script values hold Helena ASTs. They are typically used to represent blocks.
//

type ScriptValue struct {
	// Encapsulated script
	Script Script

	// Script source string
	Source *string

	// Run-time cache
	Cache *ScriptValueCache
}

// Run-time caching structure for script values
type ScriptValueCache struct {
	// Cached compiled program
	Program *Program
}

func (value ScriptValue) Type() ValueType {
	return ValueType_SCRIPT
}

// Constructor with script and source to encapsulate
func NewScriptValue(script Script, source string) ScriptValue {
	return ScriptValue{script, &source, &ScriptValueCache{}}
}

// Constructor with script to encapsulate
func NewScriptValueWithNoSource(script Script) ScriptValue {
	return ScriptValue{script, nil, &ScriptValueCache{}}
}

func (value ScriptValue) Display(fn DisplayFunction) string {
	if fn == nil {
		fn = func(_ any) string { return UndisplayableValueWithLabel("undisplayable script") }
	}
	if value.Source == nil {
		return fn(value)
	}
	return "{" + *value.Source + "}"
}

//
// Command value
//
// Command values encapsulate commands. They cannot be created directly from
// source.
//

type CommandValue interface {
	Value

	// Encapsulated command
	Command() Command
}
type commandValue struct {
	command Command
}

func (value commandValue) Type() ValueType {
	return ValueType_COMMAND
}

// Constructor with command to encapsulate
func NewCommandValue(command Command) CommandValue {
	return commandValue{command}
}

func (value commandValue) Command() Command {
	return value.command
}

//
// Qualified value
//
// Qualified values are syntactic constructs in Helena. Selectors build a new
// qualified value with the selector appended.
//

type QualifiedValue struct {
	// Source
	Source Value

	// Selectors
	Selectors []Selector
}

func (value QualifiedValue) Type() ValueType {
	return ValueType_QUALIFIED
}

// Constructor with source and selectors to encapsulate
func NewQualifiedValue(source Value, selectors []Selector) QualifiedValue {
	return QualifiedValue{source, selectors}
}

func (value QualifiedValue) Display(fn DisplayFunction) string {
	if fn == nil {
		fn = DefaultDisplayFunction
	}
	var source string
	if value.Source.Type() == ValueType_TUPLE {
		source = Display(value.Source, fn)
	} else {
		result, s := ValueToString(value.Source)
		if result.Code == ResultCode_OK {
			source = DisplayLiteralOrBlock(s)
		} else {
			source = UndisplayableValueWithLabel("source")
		}
	}
	var sels = ""
	for _, selector := range value.Selectors {
		sels += Display(selector, fn)
	}
	return source + sels
}

func (value QualifiedValue) SelectIndex(index Value) Result {
	return value.Select(NewIndexedSelector(index))
}
func (value QualifiedValue) SelectKey(key Value) Result {
	if len(value.Selectors) > 0 {
		if last, ok := value.Selectors[len(value.Selectors)-1].(KeyedSelector); ok {
			// Merge successive keys
			keys := make([]Value, len(last.Keys)+1)
			copy(keys, last.Keys)
			keys[len(last.Keys)] = key
			selectors := make([]Selector, len(value.Selectors))
			copy(selectors, value.Selectors)
			selectors[len(selectors)-1] = NewKeyedSelector(keys)
			return OK(NewQualifiedValue(value.Source, selectors))
		}
	}
	return value.Select(NewKeyedSelector([]Value{key}))
}
func (value QualifiedValue) SelectRules(rules []Value) Result {
	return value.Select(GenericSelector{rules})
}
func (value QualifiedValue) Select(selector Selector) Result {
	selectors := make([]Selector, len(value.Selectors)+1)
	copy(selectors, value.Selectors)
	selectors[len(value.Selectors)] = selector
	return OK(NewQualifiedValue(value.Source, selectors))
}

// Custom value type
type CustomValueType struct {
	// Custom value name
	Name string
}

//
// Custom values
//

type CustomValue interface {
	Value

	// Custom type info
	CustomType() CustomValueType
}

//
// Type predicates
//

// Report whether value is a Value
func IsValue(value any) bool {
	_, ok := value.(Value)
	return ok
}

// Report whether value is a custom value of the given type
func IsCustomValue(
	value Value,
	customType CustomValueType,
) bool {
	return value.Type() == ValueType_CUSTOM && value.(CustomValue).CustomType() == customType
}

//
// Convenience functions for primitive value creation
//

func BOOL(v bool) BooleanValue                { return NewBooleanValue(v) }
func INT(v int64) IntegerValue                { return NewIntegerValue(v) }
func REAL(v float64) RealValue                { return NewRealValue(v) }
func STR(v string) StringValue                { return NewStringValue(v) }
func LIST(v []Value) ListValue                { return NewListValue(v) }
func DICT(v map[string]Value) DictionaryValue { return NewDictionaryValue(v) }
func TUPLE(v []Value) TupleValue              { return NewTupleValue(v) }
