// Package toon implements the TOON (Token-Oriented Object Notation) encoder.
// TOON is a compact, LLM-optimized format that minimizes tokens while remaining human-readable.
// See: https://github.com/toon-format/toon
package toon

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

// Encode converts any Go value to TOON format with default indentation (2 spaces).
func Encode(v interface{}) (string, error) {
	return EncodeIndent(v, "", "  ")
}

// EncodeIndent converts any Go value to TOON format with custom indentation.
func EncodeIndent(v interface{}, prefix, indent string) (string, error) {
	var sb strings.Builder
	enc := &encoder{
		prefix: prefix,
		indent: indent,
	}
	if err := enc.encode(&sb, reflect.ValueOf(v), 0, ""); err != nil {
		return "", err
	}
	return sb.String(), nil
}

type encoder struct {
	prefix string
	indent string
}

func (e *encoder) encode(sb *strings.Builder, v reflect.Value, depth int, fieldName string) error {
	// Handle nil/invalid
	if !v.IsValid() {
		return nil
	}

	// Dereference pointers
	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Struct:
		return e.encodeStruct(sb, v, depth, fieldName)
	case reflect.Map:
		return e.encodeMap(sb, v, depth, fieldName)
	case reflect.Slice, reflect.Array:
		return e.encodeSlice(sb, v, depth, fieldName)
	default:
		return e.encodePrimitive(sb, v, depth, fieldName)
	}
}

func (e *encoder) encodePrimitive(sb *strings.Builder, v reflect.Value, depth int, fieldName string) error {
	prefix := e.prefix + strings.Repeat(e.indent, depth)

	if fieldName != "" {
		sb.WriteString(prefix)
		sb.WriteString(fieldName)
		sb.WriteString(": ")
	}

	switch v.Kind() {
	case reflect.String:
		s := v.String()
		if needsQuoting(s) {
			sb.WriteString(quoteString(s))
		} else {
			sb.WriteString(s)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		sb.WriteString(strconv.FormatInt(v.Int(), 10))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		sb.WriteString(strconv.FormatUint(v.Uint(), 10))
	case reflect.Float32:
		sb.WriteString(formatFloat(float64(v.Float()), 32))
	case reflect.Float64:
		sb.WriteString(formatFloat(v.Float(), 64))
	case reflect.Bool:
		if v.Bool() {
			sb.WriteString("true")
		} else {
			sb.WriteString("false")
		}
	default:
		fmt.Fprintf(sb, "%v", v.Interface())
	}

	if fieldName != "" {
		sb.WriteString("\n")
	}
	return nil
}

func (e *encoder) encodeStruct(sb *strings.Builder, v reflect.Value, depth int, fieldName string) error {
	prefix := e.prefix + strings.Repeat(e.indent, depth)
	t := v.Type()

	// If we have a field name, write the header
	if fieldName != "" {
		sb.WriteString(prefix)
		sb.WriteString(fieldName)
		sb.WriteString(":\n")
		depth++
	}

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}

		fv := v.Field(i)
		name := getFieldName(field)
		if name == "-" {
			continue
		}

		// Skip zero values for omitempty
		if hasOmitEmpty(field) && isZero(fv) {
			continue
		}

		if err := e.encode(sb, fv, depth, name); err != nil {
			return err
		}
	}
	return nil
}

func (e *encoder) encodeMap(sb *strings.Builder, v reflect.Value, depth int, fieldName string) error {
	prefix := e.prefix + strings.Repeat(e.indent, depth)

	if v.Len() == 0 {
		if fieldName != "" {
			sb.WriteString(prefix)
			sb.WriteString(fieldName)
			sb.WriteString(": {}\n")
		}
		return nil
	}

	if fieldName != "" {
		sb.WriteString(prefix)
		sb.WriteString(fieldName)
		sb.WriteString(":\n")
		depth++
	}

	// Sort keys for consistent output
	keys := v.MapKeys()
	sort.Slice(keys, func(i, j int) bool {
		return fmt.Sprintf("%v", keys[i].Interface()) < fmt.Sprintf("%v", keys[j].Interface())
	})

	for _, key := range keys {
		keyStr := fmt.Sprintf("%v", key.Interface())
		if err := e.encode(sb, v.MapIndex(key), depth, keyStr); err != nil {
			return err
		}
	}
	return nil
}

func (e *encoder) encodeSlice(sb *strings.Builder, v reflect.Value, depth int, fieldName string) error {
	prefix := e.prefix + strings.Repeat(e.indent, depth)
	length := v.Len()

	if length == 0 {
		if fieldName != "" {
			sb.WriteString(prefix)
			sb.WriteString(fieldName)
			sb.WriteString("[0]:\n")
		}
		return nil
	}

	// Check if this is an array of uniform structs (tabular format candidate)
	if length > 0 && isUniformStructSlice(v) {
		return e.encodeTabular(sb, v, depth, fieldName)
	}

	// Check if this is an array of simple primitives
	if length > 0 && isSimplePrimitiveSlice(v) {
		return e.encodeSimpleArray(sb, v, depth, fieldName)
	}

	// Mixed or complex array - encode each element
	if fieldName != "" {
		sb.WriteString(prefix)
		sb.WriteString(fieldName)
		fmt.Fprintf(sb, "[%d]:\n", length)
		depth++
	}

	for i := 0; i < length; i++ {
		elem := v.Index(i)
		if err := e.encode(sb, elem, depth, "-"); err != nil {
			return err
		}
	}
	return nil
}

// encodeTabular encodes a uniform struct slice in TOON tabular format
func (e *encoder) encodeTabular(sb *strings.Builder, v reflect.Value, depth int, fieldName string) error {
	prefix := e.prefix + strings.Repeat(e.indent, depth)
	length := v.Len()

	if length == 0 {
		return nil
	}

	// Get the element type and extract field info
	elemType := getSliceElementType(v)
	fields := getStructFields(elemType)

	if len(fields) == 0 {
		// Fall back to regular encoding
		return e.encodeSlice(sb, v, depth, fieldName)
	}

	// Build header: fieldName[count]{field1,field2,...}:
	var fieldNames []string
	for _, f := range fields {
		fieldNames = append(fieldNames, f.name)
	}

	sb.WriteString(prefix)
	if fieldName != "" {
		sb.WriteString(fieldName)
	}
	fmt.Fprintf(sb, "[%d]{%s}:\n", length, strings.Join(fieldNames, ","))

	// Write rows
	rowPrefix := e.prefix + strings.Repeat(e.indent, depth+1)
	for i := 0; i < length; i++ {
		elem := v.Index(i)
		// Dereference if pointer
		for elem.Kind() == reflect.Ptr || elem.Kind() == reflect.Interface {
			if elem.IsNil() {
				continue
			}
			elem = elem.Elem()
		}

		sb.WriteString(rowPrefix)
		var values []string
		for _, f := range fields {
			fv := elem.FieldByIndex(f.index)
			values = append(values, formatValue(fv))
		}
		sb.WriteString(strings.Join(values, ","))
		sb.WriteString("\n")
	}
	return nil
}

// encodeSimpleArray encodes a slice of primitives as comma-separated values
func (e *encoder) encodeSimpleArray(sb *strings.Builder, v reflect.Value, depth int, fieldName string) error {
	prefix := e.prefix + strings.Repeat(e.indent, depth)
	length := v.Len()

	var values []string
	for i := 0; i < length; i++ {
		elem := v.Index(i)
		values = append(values, formatValue(elem))
	}

	sb.WriteString(prefix)
	if fieldName != "" {
		sb.WriteString(fieldName)
	}
	fmt.Fprintf(sb, "[%d]: %s\n", length, strings.Join(values, ","))
	return nil
}

type fieldInfo struct {
	name  string
	index []int
}

func getStructFields(t reflect.Type) []fieldInfo {
	var fields []fieldInfo
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}
		name := getFieldName(f)
		if name == "-" {
			continue
		}
		// Only include primitive fields for tabular format
		fType := f.Type
		for fType.Kind() == reflect.Ptr {
			fType = fType.Elem()
		}
		if isPrimitiveKind(fType.Kind()) {
			fields = append(fields, fieldInfo{name: name, index: f.Index})
		}
	}
	return fields
}

func getFieldName(f reflect.StructField) string {
	tag := f.Tag.Get("json")
	if tag == "" {
		return f.Name
	}
	parts := strings.Split(tag, ",")
	if parts[0] == "" {
		return f.Name
	}
	return parts[0]
}

func hasOmitEmpty(f reflect.StructField) bool {
	tag := f.Tag.Get("json")
	return strings.Contains(tag, "omitempty")
}

func isZero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Ptr, reflect.Interface:
		return v.IsNil()
	case reflect.Slice, reflect.Map:
		return v.IsNil() || v.Len() == 0
	case reflect.String:
		return v.String() == ""
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Bool:
		return !v.Bool()
	default:
		return false
	}
}

func isUniformStructSlice(v reflect.Value) bool {
	if v.Len() == 0 {
		return false
	}
	elemType := getSliceElementType(v)
	return elemType.Kind() == reflect.Struct
}

func getSliceElementType(v reflect.Value) reflect.Type {
	elemType := v.Type().Elem()
	for elemType.Kind() == reflect.Ptr {
		elemType = elemType.Elem()
	}
	return elemType
}

func isSimplePrimitiveSlice(v reflect.Value) bool {
	if v.Len() == 0 {
		return false
	}
	elemType := v.Type().Elem()
	for elemType.Kind() == reflect.Ptr {
		elemType = elemType.Elem()
	}
	return isPrimitiveKind(elemType.Kind())
}

func isPrimitiveKind(k reflect.Kind) bool {
	switch k {
	case reflect.String, reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return true
	}
	return false
}

func formatValue(v reflect.Value) string {
	// Dereference pointers
	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		if v.IsNil() {
			return "-"
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.String:
		s := v.String()
		if needsQuotingInTable(s) {
			return quoteString(s)
		}
		return s
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(v.Uint(), 10)
	case reflect.Float32:
		return formatFloat(float64(v.Float()), 32)
	case reflect.Float64:
		return formatFloat(v.Float(), 64)
	case reflect.Bool:
		if v.Bool() {
			return "true"
		}
		return "false"
	default:
		return fmt.Sprintf("%v", v.Interface())
	}
}

func formatFloat(f float64, bitSize int) string {
	// Use shortest representation that doesn't lose precision
	s := strconv.FormatFloat(f, 'f', -1, bitSize)
	// Remove trailing zeros after decimal point
	if strings.Contains(s, ".") {
		s = strings.TrimRight(s, "0")
		s = strings.TrimRight(s, ".")
	}
	return s
}

func needsQuoting(s string) bool {
	if s == "" {
		return true
	}
	// Quote if contains special characters or looks like a number/boolean
	for _, c := range s {
		if c == ':' || c == '\n' || c == '\r' || c == '\t' || c == '"' {
			return true
		}
	}
	// Check if it looks like a reserved word or number
	if s == "true" || s == "false" || s == "null" || s == "-" {
		return true
	}
	if _, err := strconv.ParseFloat(s, 64); err == nil {
		return true
	}
	return false
}

func needsQuotingInTable(s string) bool {
	if s == "" {
		return false // Empty string in table is fine
	}
	for _, c := range s {
		if c == ',' || c == '\n' || c == '\r' || c == '"' {
			return true
		}
	}
	return false
}

func quoteString(s string) string {
	// Escape double quotes and backslashes
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	s = strings.ReplaceAll(s, "\n", `\n`)
	s = strings.ReplaceAll(s, "\r", `\r`)
	s = strings.ReplaceAll(s, "\t", `\t`)
	return `"` + s + `"`
}
