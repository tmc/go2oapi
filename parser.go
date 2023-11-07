package go2oapi

import (
	"errors"
	"fmt"
	"go/ast"
	"regexp"
	"strings"

	"github.com/fatih/structtag"
	"golang.org/x/tools/go/packages"
)

var ErrFunctionNotFound = errors.New("function not found")

// ParseFunction parses the Go source code for a specific function.
func ParseFunction(filePath string, funcName string) (*FunctionDetails, error) {
	cfg := &packages.Config{
		Mode: packages.NeedFiles | packages.NeedSyntax | packages.NeedTypesInfo | packages.NeedTypes,
		Dir:  filePath,
	}
	pkgs, err := packages.Load(cfg, ".")
	if err != nil {
		return nil, fmt.Errorf("failed to load packages: %v", err)
	}
	if packages.PrintErrors(pkgs) > 0 {
		return nil, fmt.Errorf("encountered parse errors")
	}

	var fn *ast.FuncDecl
	var pkg *packages.Package
	for _, p := range pkgs {
		if f, ok := findFunction(p, funcName); ok {
			fn = f
			pkg = p
			break
		}
	}
	if fn == nil {
		return nil, ErrFunctionNotFound
	}

	funcDetail := &FunctionDetails{
		Name:        fn.Name.Name,
		Description: cleanupComment(fn.Doc),
		Parameters:  parseParameters(pkg, fn.Type.Params.List),
	}
	return funcDetail, nil
}

// parseParameters parses function parameters and constructs a Definition.
func parseParameters(pkg *packages.Package, params []*ast.Field) *Definition {
	paramDef := &Definition{
		Type:       Object,
		Properties: make(map[string]*Definition),
	}

	for _, param := range params {
		name := param.Names[0].Name
		pd, err := paramToDetail(pkg, param)
		if err != nil {
			fmt.Printf("issue parsing parameter '%v': %v\n", name, err)
			continue
		}
		paramDef.Properties[name] = pd
		paramDef.Required = append(paramDef.Required, name)
	}

	if len(paramDef.Properties) == 1 && paramDef.Type == Object {
		var singlePropertyName string
		for name := range paramDef.Properties {
			singlePropertyName = name
			break
		}
		paramDef = paramDef.Properties[singlePropertyName]
	}
	return paramDef
}

// Regex for comment prefix.
var slashCommentPrefixRe = regexp.MustCompile("^// ?")

func cleanupComment(commentGroups ...*ast.CommentGroup) string {
	var comments []string
	for _, cg := range commentGroups {
		if cg != nil {
			for _, c := range cg.List {
				comments = append(comments, slashCommentPrefixRe.ReplaceAllString(c.Text, ""))
			}
		}
	}
	return strings.Join(comments, " ")
}

func paramToDetail(pkg *packages.Package, param *ast.Field) (*Definition, error) {
	paramType := exprToType(pkg.TypesInfo, param.Type)
	d := &Definition{
		Type:        paramType,
		Properties:  make(map[string]*Definition),
		Description: cleanupComment(param.Doc, param.Comment),
	}
	if enumOptions, err := parseEnumTag(param.Tag); err != nil {
		return nil, fmt.Errorf("issue parsing 'enum' field tag: %v", err)
	} else {
		d.Enum = enumOptions
	}

	var err error
	switch paramType {
	case Object:
		if st, ok := findStructTypeFromIdent(param.Type.(*ast.Ident)); ok {
			for _, f := range st.Fields.List {
				fieldName := f.Names[0].Name
				d.Properties[fieldName], err = paramToDetail(pkg, f)
				if err != nil {
					return nil, fmt.Errorf("issue parsing field '%v': %v", fieldName, err)
				}

				// Set field as required unless the 'required' tag explicitly marks it otherwise.
				if required, tagErr := parseRequiredTag(f.Tag); tagErr == nil && required {
					d.Required = append(d.Required, fieldName)
				}
			}
		}
	case Array:
		d.Items = &Definition{
			Type: exprToType(pkg.TypesInfo, param.Type.(*ast.ArrayType).Elt),
		}
	}
	return d, nil
}

func parseEnumTag(tag *ast.BasicLit) ([]string, error) {
	if tag == nil {
		return nil, nil
	}
	val := strings.Trim(tag.Value, "`")
	tags, err := structtag.Parse(val)
	if err != nil {
		return nil, err
	}
	enumTag, err := tags.Get("enum")
	if err != nil {
		return nil, nil // No enum tag is not an error.
	}
	return append([]string{enumTag.Name}, enumTag.Options...), nil
}

// parseRequiredTag parses the 'required' tag from a struct field's tag.
func parseRequiredTag(tag *ast.BasicLit) (bool, error) {
	if tag == nil {
		return true, nil
	}
	val := strings.Trim(tag.Value, "`")
	tags, err := structtag.Parse(val)
	if err != nil {
		return false, fmt.Errorf("error parsing tags '%v': %v", tag, err)
	}
	requiredTag, err := tags.Get("required")
	if err != nil { // we presume that the 'required' tag is not present.
		return true, nil
	}

	// We define falsey values that denote a field is not required.
	falseyValues := map[string]bool{
		"false": true,
		"no":    true,
		"0":     true,
	}
	// If the 'required' tag's value is a falsey value, the field is not required.
	return !falseyValues[strings.ToLower(requiredTag.Name)], nil
}

func findFunction(pkg *packages.Package, funcName string) (*ast.FuncDecl, bool) {
	for _, file := range pkg.Syntax {
		for _, d := range file.Decls {
			if fn, ok := d.(*ast.FuncDecl); ok && fn.Name.Name == funcName {
				return fn, true
			}
		}
	}
	return nil, false
}

func findStructTypeFromIdent(expr ast.Expr) (*ast.StructType, bool) {
	if ident, ok := expr.(*ast.Ident); ok && ident.Obj != nil {
		if typeSpec, ok := ident.Obj.Decl.(*ast.TypeSpec); ok {
			if structType, ok := typeSpec.Type.(*ast.StructType); ok {
				return structType, true
			}
		}
	}
	return nil, false
}
