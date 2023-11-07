package go2oapi

import (
	"go/ast"
	"go/types"
)

// Encodes the type of a field.
type DataType string

const (
	Object  DataType = "object"
	Number  DataType = "number"
	Integer DataType = "integer"
	String  DataType = "string"
	Array   DataType = "array"
	Null    DataType = "null"
	Boolean DataType = "boolean"
)

// FunctionDetails describes a Go function.
type FunctionDetails struct {
	// Name of the function.
	Name string `json:"name"`
	// Description of the function.
	Description string `json:"description"`
	// Parameters of the function.
	Parameters *Definition `json:"parameters"`
}

// Definition holds the name and type of a function parameter.
type Definition struct {
	// Type specifies the data type of the schema.
	Type DataType `json:"type,omitempty"`
	// Description is the description of the schema.
	Description string `json:"description,omitempty"`
	// Enum is used to restrict a value to a fixed set of values. It must be an array with at least
	// one element, where each element is unique. You will probably only use this with strings.
	Enum []string `json:"enum,omitempty"`
	// Properties describes the properties of an object, if the schema type is Object.
	Properties map[string]*Definition `json:"properties,omitempty"`
	// Required specifies which properties are required, if the schema type is Object.
	Required []string `json:"required,omitempty"`
	// Items specifies which data type an array contains, if the schema type is Array.
	Items *Definition `json:"items,omitempty"`
}

func exprToType(info *types.Info, expr ast.Expr) DataType {
	typ := info.TypeOf(expr)
	if typ == nil {
		return Null // or some error handling
	}

	switch typ := typ.Underlying().(type) {
	case *types.Basic:
		return basicTypeToDataType(typ)
	case *types.Array, *types.Slice:
		return Array
	case *types.Struct:
		return Object
	case *types.Pointer:
		return exprToType(info, &ast.Ident{Name: typ.Elem().String()})
	default:
		return Null
	}
}

func basicTypeToDataType(basic *types.Basic) DataType {
	switch basic.Kind() {
	case types.Bool:
		return Boolean
	case types.Int, types.Int8, types.Int16, types.Int32, types.Int64,
		types.Uint, types.Uint8, types.Uint16, types.Uint32, types.Uint64:
		return Integer
	case types.Float32, types.Float64:
		return Number
	case types.String:
		return String
	default:
		return Null
	}
}
