package generator

import (
	"go/ast"
	"go/doc"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"log"
	"os"
	"strings"

	"github.com/machinebox/remoto/generator/definition"
	"github.com/pkg/errors"
)

// Parse parses a package of .remoto.go files.
func Parse(dir string) (definition.Definition, error) {
	var def definition.Definition
	def.Comments = make(map[string]string)
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, dir, func(info os.FileInfo) bool {
		return strings.HasSuffix(info.Name(), ".remoto.go")
	}, parser.ParseComments)
	if err != nil {
		return def, errors.Wrap(err, "parser.ParseDir")
	}
	pkgNames := make([]string, 0, len(pkgs))
	for pkg := range pkgs {
		pkgNames = append(pkgNames, pkg)
	}
	if len(pkgNames) == 0 {
		return def, errors.New("no packages found")
	}
	if len(pkgNames) > 1 {
		return def, errors.New("multiple packages found: " + strings.Join(pkgNames, ", "))
	}
	firstPkg := pkgs[pkgNames[0]]
	files := make([]*ast.File, 0, len(firstPkg.Files))
	for _, file := range firstPkg.Files {
		files = append(files, file)
	}
	docs := doc.New(firstPkg, "./", doc.AllDecls+doc.AllMethods)
	def.PackageName = pkgNames[0]
	def.PackageComment = strings.TrimSpace(docs.Doc)
	info := &types.Info{}
	conf := types.Config{Importer: importer.Default()}
	pkg, err := conf.Check(dir, fset, files, info)
	if err != nil {
		return def, errors.Wrap(err, "conf.Check")
	}
	for _, f := range files {
		for _, comment := range f.Comments {
			pos := comment.Pos()
			trimmedComment := strings.TrimSpace(comment.Text())
			name := strings.Split(trimmedComment, " ")[0]
			inner := pkg.Scope().Innermost(pos)
			if _, obj := inner.LookupParent(name, pos); obj != nil {
				def.Comments[obj.Name()] = trimmedComment
			}
		}
	}
	for _, name := range pkg.Scope().Names() {
		obj := pkg.Scope().Lookup(name)
		switch v := obj.Type().Underlying().(type) {
		case *types.Interface:
			service, err := parseService(fset, docs, pkg, &def, obj, v)
			if err != nil {
				for sub, tip := range tips {
					if strings.Contains(err.Error(), sub) {
						err = errors.New(err.Error() + ": " + tip)
						break
					}
				}
				return def, err
			}
			def.Services = append(def.Services, service)
		}
	}
	return def, nil
}

func parseService(fset *token.FileSet, docs *doc.Package, pkg *types.Package, def *definition.Definition, obj types.Object, v *types.Interface) (definition.Service, error) {
	docstype, comment := commentForType(docs, obj.Name())
	srv := definition.Service{
		Name:    obj.Name(),
		Comment: comment,
	}
	for i := 0; i < v.NumMethods(); i++ {
		m := v.Method(i)
		method, err := parseMethod(fset, docs, docstype, pkg, def, &srv, m)
		if err != nil {
			return srv, err
		}
		srv.Methods = append(srv.Methods, method)
	}
	return srv, nil
}

func parseMethod(fset *token.FileSet, docs *doc.Package, docstype *doc.Type, pkg *types.Package, def *definition.Definition, srv *definition.Service, m *types.Func) (definition.Method, error) {
	method := definition.Method{
		Name: m.Name(),
	}
	if !m.Exported() {
		return method, newErr(fset, m.Pos(), "method "+m.Name()+": must be exported")
	}
	sig := m.Type().(*types.Signature)
	if sig.Variadic() {
		return method, newErr(fset, m.Pos(), "service methods must have signature (*Request) *Response")
	}

	if docstype != nil {
		ast.Inspect(docstype.Decl, func(n ast.Node) bool {
			astfield, ok := n.(*ast.Field)
			if !ok {
				return true // skip
			}
			if len(astfield.Names) < 1 {
				return true // skip
			}
			if astfield.Names[0].Name == m.Name() {
				method.Comment = strings.TrimSpace(astfield.Doc.Text())
			}
			return true
		})
	}

	params := sig.Params()
	// process input arguments
	if params.Len() != 1 {
		return method, newErr(fset, m.Pos(), "service methods must have signature (*Request) *Response")
	}
	requestParam := params.At(0)
	requestStructure, err := parseStructureFromParam(fset, docs, pkg, def, srv, "request", requestParam)
	if err != nil {
		return method, err
	}
	if !strings.HasSuffix(requestStructure.Name, "Request") {
		return method, newErr(fset, m.Pos(), "request object type name should end with \"Request\"")
	}
	requestStructure.IsRequestObject = true
	method.RequestStructure = requestStructure
	srv.EnsureStructure(requestStructure)
	// process return arguments
	returns := sig.Results()
	if returns.Len() != 1 {
		return method, newErr(fset, m.Pos(), "service methods must have signature (*Request) *Response")
	}
	responseParam := returns.At(0)
	responseStructure, err := parseStructureFromParam(fset, docs, pkg, def, srv, "response", responseParam)
	if err != nil {
		return method, err
	}
	if requestStructure.Name == responseStructure.Name {
		return method, newErr(fset, m.Pos(), "service methods must use different types for request and response objects")
	}
	responseStructure.IsResponseObject = true
	if !strings.HasSuffix(responseStructure.Name, "Response") {
		return method, newErr(fset, m.Pos(), "response object type name should end with \"Response\"")
	}
	addDefaultResponseFields(&responseStructure)
	method.ResponseStructure = responseStructure
	srv.EnsureStructure(responseStructure)
	return method, nil
}

// addDefaultResponseFields adds the built-in remoto fields to the
// response structure.
func addDefaultResponseFields(structure *definition.Structure) {
	if structure.HasField("Error") {
		return
	}
	structure.Fields = append(structure.Fields, definition.Field{
		Comment: "Error is an error message if one occurred.",
		Name:    "Error",
		Type: definition.Type{
			Name: "string",
		},
	})
}

func parseStructureFromParam(fset *token.FileSet, docs *doc.Package, pkg *types.Package, def *definition.Definition, srv *definition.Service, structureKind string, v *types.Var) (definition.Structure, error) {
	resolver := func(other *types.Package) string {
		if other.Name() != def.PackageName {
			return other.Name()
		}
		return ""
	}
	var structure definition.Structure
	p, ok := v.Type().(*types.Pointer)
	if !ok {
		return structure, newErr(fset, v.Pos(), structureKind+" object must be a pointer to a struct")
	}
	st, ok := p.Elem().Underlying().(*types.Struct)
	if !ok {
		return structure, newErr(fset, v.Pos(), structureKind+" object must be a pointer to a struct")
	}
	structure.Name = types.TypeString(v.Type(), resolver)[1:]
	var docstype *doc.Type
	docstype, structure.Comment = commentForType(docs, structure.Name)
	structure.IsImported = strings.Contains(structure.Name, ".")
	for i := 0; i < st.NumFields(); i++ {
		field, err := parseField(fset, docs, docstype, pkg, def, srv, st.Field(i))
		if err != nil {
			return structure, err
		}
		structure.Fields = append(structure.Fields, field)
	}
	return structure, nil
}

func parseStructure(fset *token.FileSet, docs *doc.Package, pkg *types.Package, def *definition.Definition, srv *definition.Service, obj types.Object) (definition.Structure, error) {
	structure := definition.Structure{
		Name: obj.Name(),
	}
	var docstype *doc.Type
	docstype, structure.Comment = commentForType(docs, structure.Name)
	st, ok := obj.Type().Underlying().(*types.Struct)
	if !ok {
		return structure, newErr(fset, obj.Pos(), obj.Type().String()+" field must be a pointer to a struct")
	}
	for i := 0; i < st.NumFields(); i++ {
		field, err := parseField(fset, docs, docstype, pkg, def, srv, st.Field(i))
		if err != nil {
			return structure, err
		}
		structure.Fields = append(structure.Fields, field)
	}
	return structure, nil
}

func parseField(fset *token.FileSet, docs *doc.Package, docstype *doc.Type, pkg *types.Package, def *definition.Definition, srv *definition.Service, v *types.Var) (definition.Field, error) {
	var field definition.Field
	if !v.IsField() {
		return field, newErr(fset, v.Pos(), v.Name()+" not a field")
	}
	if !v.Exported() {
		return field, newErr(fset, v.Pos(), "field "+v.Name()+": must be exported")
	}
	if docstype != nil {
		ast.Inspect(docstype.Decl, func(n ast.Node) bool {
			astfield, ok := n.(*ast.Field)
			if !ok {
				return true // skip
			}
			if len(astfield.Names) < 1 {
				return true // skip
			}
			if astfield.Names[0].Name == v.Name() {
				field.Comment = strings.TrimSpace(astfield.Doc.Text())
			}
			return true
		})
	}
	typ, err := parseType(def, v.Type())
	if err != nil {
		return field, newErr(fset, v.Pos(), err.Error())
	}
	field.Name = v.Name()
	field.Type = typ
	if typ.IsStruct && !typ.IsImported {
		obj := pkg.Scope().Lookup(typ.Name)
		if obj == nil {
			return field, newErr(fset, v.Pos(), typ.Name+" not found")
		}
		structure, err := parseStructure(fset, docs, pkg, def, srv, obj)
		if err != nil {
			return field, err
		}
		srv.EnsureStructure(structure)
	}
	return field, nil
}

func newErr(fset *token.FileSet, pos token.Pos, err string) error {
	position := fset.Position(pos)
	return errors.New(position.String() + ": " + err)
}

func parseType(def *definition.Definition, typ types.Type) (definition.Type, error) {
	resolver := func(other *types.Package) string {
		if other.Name() != def.PackageName {
			return other.Name()
		}
		return ""
	}
	var ty definition.Type
	slice, ok := typ.(*types.Slice)
	if ok {
		ty.IsMultiple = true
		typ = slice.Elem()
	}
	ty.Name = types.TypeString(typ, resolver)
	ty.IsImported = strings.Contains(ty.Name, ".")
	if _, ok := typ.Underlying().(*types.Struct); ok {
		ty.IsStruct = true
		return ty, nil
	}
	if pointer, ok := typ.Underlying().(*types.Pointer); ok {
		log.Printf("Is pointer %v %T \n", pointer, pointer.Elem().Underlying())
		// trim off *
		ty.Name = ty.Name[1:]
		if _, ok := pointer.Elem().Underlying().(*types.Struct); ok {
			ty.IsStruct = true
			return ty, nil
		}
	}
	switch ty.Name {
	case "string", "float64", "int", "bool", "io.Reader",
		"remototypes.File":
		return ty, nil
	}
	return ty, errors.New("type " + ty.Name + " not supported")
}

// commentForType gets the comment for the specified type name.
func commentForType(docs *doc.Package, typename string) (*doc.Type, string) {
	for _, typ := range docs.Types {
		if typ.Name == typename {
			return typ, strings.TrimSpace(typ.Doc)
		}
	}
	return nil, ""
}

// tips are simple error string matches (keys) which if found,
// will have the tip information (value) appended to the error.
var tips = map[string]string{
	" int32 ":     "use int",
	" int64 ":     "use int",
	" float32 ":   "use float64",
	" time.Time ": "use string",
}
