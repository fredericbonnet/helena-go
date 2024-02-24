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
)

//
// /** Helena custom value types */
// export interface CustomValueType {
//   /** Custom value name */
//   readonly name: string;
// }
//

// Helena value
type Value interface {
	// Displayable

	// Type identifier
	// readonly type: ValueType | CustomValueType;
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

// Nil value
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

// Boolean value
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
func BooleanValueFromValue(value Value) TypedResult[BooleanValue] {
	if value.Type() == ValueType_BOOLEAN {
		return OK_T(value, BooleanValue(value.(BooleanValue)))
	}
	result := ValueToBoolean(value)
	if result.Code != ResultCode_OK {
		return ResultAs[BooleanValue](result.AsResult())
	}
	if result.Data {
		return OK_T(TRUE, TRUE)
	} else {
		return OK_T(FALSE, FALSE)
	}
}

// Convert value to boolean:
// - Booleans: use boolean value
// - Strings: true, false
func ValueToBoolean(value Value) TypedResult[bool] {
	if value.Type() == ValueType_BOOLEAN {
		return OK_T(NIL, BooleanValue(value.(BooleanValue)).Value)
	}
	result := ValueToString(value)
	if result.Code != ResultCode_OK {
		return ResultAs[bool](result.AsResult())
	}
	s := result.Data
	if s == "true" {
		return OK_T(NIL, true)
	}
	if s == "false" {
		return OK_T(NIL, false)
	}
	return ERROR_T[bool](`invalid boolean "` + s + `"`)
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
func IntegerValueFromValue(value Value) TypedResult[IntegerValue] {
	if value.Type() == ValueType_INTEGER {
		return OK_T(value, value.(IntegerValue))
	}
	result := ValueToInteger(value)
	if result.Code != ResultCode_OK {
		return ResultAs[IntegerValue](result.AsResult())
	}
	v := NewIntegerValue(result.Data)
	return OK_T(v, v)
}

//   /**
//    * Test if value is convertible to number
//    * - Reals
//    * - Integers
//    * - Strings: any Number()-accepted string
//    *
//    * @param value - Value to convert
//    *
//    * @returns       True if value is convertible
//    */
//   static isInteger(value: Value): boolean {
//     switch (value.type) {
//       case ValueType_INTEGER:
//         return true;
//       case ValueType_REAL:
//         return Number.isSafeInteger((value as RealValue).value);
//       default: {
//         const { data, code } = StringValue.toString(value);
//         if (code != ResultCode_OK) return false;
//         const n = Number(data);
//         return !isNaN(n) && Number.isSafeInteger(n);
//       }
//     }
//   }

// Convert value to integer:
// - Integers: use integer value
// - Reals: any safe integer number
// - Strings: any integer Number()-accepted string
func ValueToInteger(value Value) TypedResult[int64] {
	if value.Type() == ValueType_INTEGER {
		return OK_T(NIL, value.(IntegerValue).Value)
	}
	// TODO: is it needed? strconv.ParseInt would work but maybe converting int to float is more efficient
	// if (value.Type() == ValueType_REAL) {
	//   if (!Number.isSafeInteger((value as RealValue).value))
	//     return ERROR(`invalid integer "${(value as RealValue).value}"`);
	//   return OK(NIL, (value as RealValue).value);
	// }
	result := ValueToString(value)
	if result.Code != ResultCode_OK {
		return ResultAs[int64](result.AsResult())
	}
	s := result.Data
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return ERROR_T[int64](`invalid integer "` + s + `"`)
	}
	return OK_T(NIL, n)
}

func (value IntegerValue) Display(_ DisplayFunction) string {
	return fmt.Sprint(value.Value)
}

// Real value
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
func RealValueFromValue(value Value) TypedResult[RealValue] {
	if value.Type() == ValueType_REAL {
		return OK_T(value, value.(RealValue))
	}
	if value.Type() == ValueType_INTEGER {
		v := NewRealValue(float64(value.(IntegerValue).Value))
		return OK_T(v, v)
	}
	result := ValueToFloat(value)
	if result.Code != ResultCode_OK {
		return ResultAs[RealValue](result.AsResult())
	}
	v := NewRealValue(result.Data)
	return OK_T(v, v)
}

//   /**
//    * Test if value is convertible to number
//    * - Reals
//    * - Integers
//    * - Strings: any Number()-accepted string
//    *
//    * @param value - Value to convert
//    *
//    * @returns       True if value is convertible
//    */
//   static isNumber(value: Value): boolean {
//     switch (value.type) {
//       case ValueType_INTEGER:
//       case ValueType_REAL:
//         return true;
//       default: {
//         const { data, code } = StringValue.toString(value);
//         if (code != ResultCode_OK) return false;
//         return !isNaN(Number(data));
//       }
//     }
//   }

// Convert value to float:
// - Reals: use float value
// - Integers: use int value
// - Strings: any strconv.ParseFloat()-accepted string
func ValueToFloat(value Value) TypedResult[float64] {
	if value.Type() == ValueType_REAL {
		return OK_T(NIL, value.(RealValue).Value)
	}
	if value.Type() == ValueType_INTEGER {
		return OK_T(NIL, float64(value.(IntegerValue).Value))
	}
	result := ValueToString(value)
	if result.Code != ResultCode_OK {
		return ResultAs[float64](result.AsResult())
	}
	s := result.Data
	n, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return ERROR_T[float64](`invalid number "` + s + `"`)
	}
	return OK_T(NIL, n)
}

func (value RealValue) Display(_ DisplayFunction) string {
	return fmt.Sprint(value.Value)
}

// String value
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
func StringValueFromValue(value Value) TypedResult[StringValue] {
	if value.Type() == ValueType_STRING {
		return OK_T(value, value.(StringValue))
	}
	result := ValueToString(value)
	if result.Code != ResultCode_OK {
		return ResultAs[StringValue](result.AsResult())
	}
	v := NewStringValue(result.Data)
	return OK_T(v, v)
}

// Convert value to string
func ValueToString(value Value) TypedResult[string] {
	switch value.Type() {
	case ValueType_STRING:
		return OK_T(NIL, value.(StringValue).Value)
	case ValueType_BOOLEAN:
		if value.(BooleanValue).Value {
			return OK_T(NIL, "true")
		} else {
			return OK_T(NIL, "false")
		}
	case ValueType_INTEGER:
		return OK_T(NIL, fmt.Sprint(value.(IntegerValue).Value))
	case ValueType_REAL:
		return OK_T(NIL, fmt.Sprint(value.(RealValue).Value))
	case ValueType_SCRIPT:
		{
			source := value.(ScriptValue).Source
			if source != nil {
				return OK_T(NIL, *source)
			}
		}
	}
	return ERROR_T[string]("value has no string representation")
}

// Convert value to string, or default value if value has no string representation
func ValueToStringOrDefault(value Value, def Value) TypedResult[string] {
	return ERROR_T[string]("TODO")
}

// Return index-th string character as StringValue
func StringAt(value string, index Value) Result {
	result := ValueToInteger(index)
	if result.Code != ResultCode_OK {
		return result.AsResult()
	}
	i := result.Data
	if i < 0 || i >= int64(len(value)) {
		return ERROR(`index out of range "` + fmt.Sprint(i) + `"`)
	}
	return OK(StringValue{string(value[i])})
}

// Return index-th string character as StringValue, or default value for
// out-of-range index
func StringAtOrDefault(value string, index Value, def Value) Result {
	//     const result = IntegerValue.toInteger(index);
	//     if (result.Code != ResultCode_OK) return result;
	//     const i = result.data;)
	//     if (i < 0 || i >= value.length)
	//       return def ? OK(def) : ERROR(`index out of range "${i}"`);
	//     return OK(new StringValue(value[i]));
	return ERROR("TODO")
}

func (value StringValue) Display(_ DisplayFunction) string {
	return DisplayLiteralOrString(value.Value)
}

func (value StringValue) SelectIndex(index Value) Result {
	return StringAt(value.Value, index)
}

// List value
//
// Lists are linear collections of other values
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
func ListValueFromValue(value Value) TypedResult[ListValue] {
	if value.Type() == ValueType_LIST {
		return OK_T(value, value.(ListValue))
	}
	result := ValueToValues(value)
	if result.Code != ResultCode_OK {
		return ResultAs[ListValue](result.AsResult())
	}
	v := NewListValue(result.Data)
	return OK_T(v, v)
}

// Convert value to array of values:
// - Lists
// - Tuples
func ValueToValues(value Value) TypedResult[[]Value] {
	switch value.Type() {
	case ValueType_LIST:
		return OK_T(NIL, value.(ListValue).Values)
	case ValueType_TUPLE:
		return OK_T(NIL, value.(TupleValue).Values)
	default:
		return ERROR_T[[]Value]("invalid list")
	}
}

// Return index-th element in list
func ListAt(values []Value, index Value) Result {
	result := ValueToInteger(index)
	if result.Code != ResultCode_OK {
		return result.AsResult()
	}
	i := result.Data
	if i < 0 || i >= int64(len(values)) {
		return ERROR(`index out of range "` + fmt.Sprint(i) + `"`)
	}
	return OK(values[i])
}

// Return index-th element in list, or default value for out-of-range index
func ListAtOrDefault(value string, index Value, def Value) Result {
	//     const result = IntegerValue.toInteger(index);
	//     if (result.Code != ResultCode_OK) return result;
	//     const i = result.data;)
	//     if (i < 0 || i >= value.length)
	//       return def ? OK(def) : ERROR(`index out of range "${i}"`);
	//     return OK(new StringValue(value[i]));
	return ERROR("TODO")
}

func (value ListValue) SelectIndex(index Value) Result {
	return ListAt(value.Values, index)
}

// Dictionary value
//
// Dictionaries are key-value collections with string keys
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
	result := ValueToString(key)
	if result.Code != ResultCode_OK {
		return ERROR("invalid key")
	}
	k := result.Data
	v, ok := value.Map[k]
	if !ok {
		return ERROR("unknown key")
	}
	return OK(v)
}

// Tuple value
//
// Tuples are syntactic constructs in Helena. Selectors apply recursively to
// their elements.
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

// Script value
//
// Script values hold Helena ASTs. They are typically used to represent blocks.
type ScriptValue struct {
	// Encapsulated script
	Script Script

	// Script source string
	Source *string
}

func (value ScriptValue) Type() ValueType {
	return ValueType_SCRIPT
}

// Constructor with script and source to encapsulate
func NewScriptValue(script Script, source string) ScriptValue {
	return ScriptValue{script, &source}
}

// Constructor with script to encapsulate
func NewScriptValueWithNoSource(script Script) ScriptValue {
	return ScriptValue{script, nil}
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

// /**
//  * Command value
//  *
//  * Command values encapsulate commands. They cannot be created directly from
//  * source.
//  */
// export class CommandValue implements Value {
//   /** @override */
//   readonly type = ValueType_COMMAND;

//   /** Encapsulated command */
//   readonly command: Command;

//   /**
//    * @param command - Command to encapsulate
//    */
//   constructor(command: Command) {
//     this.command = command;
//   }
// }

// Qualified value
//
// Qualified values are syntactic constructs in Helena. Selectors build a new
// qualified value with the selector appended.
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
		result := ValueToString(value.Source)
		if result.Code == ResultCode_OK {
			source = DisplayLiteralOrBlock(result.Data)
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

// /*
//  * Type predicates
//  *
//  * See https://www.typescriptlang.org/docs/handbook/2/narrowing.html#using-type-predicates
//  */

// /**
//  * Type predicate for Value
//  *
//  * @param value - Object to test
//  * @returns       Whether value is a Value
//  */
// export function isValue(value: Displayable): value is Value {
//   return !!value["type"];
// }

// /**
//  * Type predicate for CustomValueType
//  *
//  * @param type - Type to test
//  * @returns      Whether type is a CustomValueType
//  */
// export function isCustomValueType(
//   type: ValueType | CustomValueType
// ): type is CustomValueType {
//   return !!type["name"];
// }

//
// Convenience functions for primitive value creation
//

func BOOL(v bool) Value             { return NewBooleanValue(v) }
func INT(v int64) Value             { return NewIntegerValue(v) }
func REAL(v float64) Value          { return NewRealValue(v) }
func STR(v string) Value            { return NewStringValue(v) }
func LIST(v []Value) Value          { return NewListValue(v) }
func DICT(v map[string]Value) Value { return NewDictionaryValue(v) }
func TUPLE(v []Value) Value         { return NewTupleValue(v) }
