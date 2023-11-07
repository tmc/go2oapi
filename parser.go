package go2oapi

import (
	"errors"
	"fmt"
	"go/ast"
	"go/types"
	"os"
	"regexp"
	"strings"

	"github.com/fatih/structtag"
	"golang.org/x/tools/go/packages"
)

var ErrFunctionNotFound = errors.New("function not found")

// ParseFunction parses the Go source code for a specific function.
func ParseFunction(filePath string, funcName string) (*FunctionDetails, error) {
	// Many tools pass their command-line arguments (after any flags)
	// uninterpreted to packages.Load so that it can interpret them
	// according to the conventions of the underlying build system.
	cfg := &packages.Config{
		Mode: packages.NeedFiles | packages.NeedSyntax | packages.NeedTypesInfo | packages.NeedTypes,
		Dir:  filePath,
	}
	pkgs, err := packages.Load(cfg, ".")
	if err != nil {
		fmt.Fprintf(os.Stderr, "load: %v\n", err)
		os.Exit(1)
	}
	if packages.PrintErrors(pkgs) > 0 {
		return nil, fmt.Errorf("parse error")
	}

	var fn *ast.FuncDecl
	var pkg *packages.Package
	for _, p := range pkgs {
		f, ok := findFunction(p, funcName)
		if ok {
			fn = f
			pkg = p
		}
	}
	if fn == nil {
		return nil, ErrFunctionNotFound
	}

	funcDetail := &FunctionDetails{
		Name:        fn.Name.Name,
		Description: cleanupComment(fn.Doc),
		Parameters: &Definition{
			Type:       "object",
			Properties: map[string]*Definition{},
		},
	}
	if fn.Type.Params != nil {
		for _, param := range fn.Type.Params.List {
			name := identsToName(param.Names)
			pd, err := paramToDetail(pkg, param)
			if err != nil {
				return nil, fmt.Errorf("issue parsing parameter '%v': %w", name, err)
			}
			funcDetail.Parameters.Properties[name] = pd
			funcDetail.Parameters.Required = append(funcDetail.Parameters.Required, name)
		}
	}
	return funcDetail, nil
}

var slashCommentPrefixRe = regexp.MustCompile("^// ?")

func trimCommentPrefix(c string) string {
	return slashCommentPrefixRe.ReplaceAllString(c, "")
}

func cleanupComment(commentGroups ...*ast.CommentGroup) string {
	var commentParts []string
	for _, cg := range commentGroups {
		if cg == nil {
			continue
		}
		for _, c := range cg.List {
			commentParts = append(commentParts, trimCommentPrefix(c.Text))
		}
	}
	return strings.Join(commentParts, " ")
}

func paramToDetail(pkg *packages.Package, param *ast.Field) (*Definition, error) {
	paramType := exprToType(pkg, param.Type)
	d := &Definition{
		Type:        paramType,
		Properties:  map[string]*Definition{},
		Description: cleanupComment(param.Doc, param.Comment),
	}
	if param.Tag != nil {
		enumOptions, err := parseEnumTag(param.Tag.Value)
		if err != nil {
			return nil, fmt.Errorf("issue parsing 'enum' field tag: %w", err)
		}
		d.Enum = enumOptions
	}

	if paramType == "object" {
		var err error
		var st *ast.StructType
		switch pt := param.Type.(type) {
		case *ast.StructType:
			st = pt
		case *ast.Ident:
			st, _ = findStructTypeFromIdent(pt)
		}
		for _, f := range st.Fields.List {
			d.Properties[identsToName(f.Names)], err = paramToDetail(pkg, f)
			if err != nil {
				return nil, err
			}
		}
	}
	if paramType == "array" {
		d.Items = &Definition{
			Type: exprToType(pkg, param.Type.(*ast.ArrayType).Elt),
		}
	}
	return d, nil
}

func findStructTypeFromIdent(ident *ast.Ident) (*ast.StructType, bool) {
	// Check if the ident has an associated object (it should if the parser had type info).
	if ident.Obj == nil {
		return nil, false
	}

	// Check if the declaration of the object is a type specification.
	typeSpec, ok := ident.Obj.Decl.(*ast.TypeSpec)
	if !ok {
		return nil, false
	}

	// Finally, assert that the type specification is indeed a struct type.
	structType, ok := typeSpec.Type.(*ast.StructType)
	if !ok {
		return nil, false
	}

	return structType, true
}

func parseEnumTag(tag string) ([]string, error) {
	tag = strings.Trim(tag, "`")
	tags, err := structtag.Parse(tag)
	if err != nil {
		return nil, fmt.Errorf("parse('%v'): %w", tag, err)
	}
	value, err := tags.Get("enum")
	if err != nil {
		return nil, nil
	}
	var options []string
	options = append(options, value.Name)
	options = append(options, value.Options...)
	return options, nil
}

func identsToName(idents []*ast.Ident) string {
	for _, i := range idents {
		if i.Name != "" {
			return i.Name
		}
	}
	return ""
}

func findFunction(pkg *packages.Package, funcName string) (*ast.FuncDecl, bool) {
	for _, file := range pkg.Syntax {
		f, ok := findFunctionFile(file, funcName)
		if ok {
			return f, true
		}
	}
	return nil, false
}
func findFunctionFile(f *ast.File, funcName string) (*ast.FuncDecl, bool) {
	for _, d := range f.Decls {
		fn, ok := d.(*ast.FuncDecl)
		if !ok {
			continue
		}
		if fn.Name.Name == funcName {
			return fn, true
		}
	}
	return nil, false
}

var goTypesToDataType = map[string]DataType{
	"int":    Integer,
	"int32":  Integer,
	"int64":  Integer,
	"string": String,
	"float":  Number,
	"bool":   Boolean,
}

// exprToType takes an expression and returns its string representation.
func exprToType(pkg *packages.Package, expr ast.Expr) DataType {
	switch t := expr.(type) {
	case *ast.Ident:
		typ := goTypesToDataType[t.Name]
		ut := pkg.TypesInfo.Types[t].Type.Underlying()
		switch ut.(type) {
		case *types.Struct:
			return Object
		}
		return typ
	case *ast.ArrayType:
		return Array
	case *ast.StarExpr:
		return exprToType(pkg, t.X)
	case *ast.StructType:
		return Object
	// Add more cases as needed for other types.
	default:
		fmt.Printf("uhandled type %T\n", t)
		return Null
	}
}
