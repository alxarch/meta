package meta

import (
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/printer"
	"go/token"
	"go/types"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Package struct {
	*types.Package
	fset  *token.FileSet
	info  types.Info
	ast   *ast.Package
	files []*ast.File
	stat  os.FileInfo
}

func (p *Package) FindImport(pkgName string) *types.Package {
	if p == nil || p.Package == nil {
		return nil
	}
	for _, pkg := range p.Package.Imports() {
		if pkg.Name() == pkgName {
			return pkg
		}
	}
	return nil
}

func IgnoreTestFiles(f os.FileInfo) bool {
	return !strings.HasSuffix(f.Name(), "_test.go")
}

func ParsePackage(name, path string, fileFilter func(os.FileInfo) bool, mode parser.Mode, astFilter func(f *ast.File) bool) (p *Package, err error) {
	stat, err := os.Stat(path)
	if err != nil {
		return
	}
	if stat.IsDir() {
		path = filepath.Clean(path)
	} else {
		path = filepath.Dir(path)
	}
	if name == "" {
		name = filepath.Base(path)
	}
	fset := token.NewFileSet()
	astPkgs, err := parser.ParseDir(fset, path, fileFilter, mode)
	if err != nil {
		return
	}
	astPkg := astPkgs[name]
	if astPkg == nil {
		err = fmt.Errorf("Target package %q not found in path %s", name, path)
		return
	}
	config := types.Config{
		IgnoreFuncBodies: true,
		FakeImportC:      true,
		Importer:         importer.Default(),
	}
	info := types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
		Defs:  make(map[*ast.Ident]types.Object),
	}
	files := make([]*ast.File, 0, len(astPkg.Files))
	for _, file := range astPkg.Files {
		if astFilter == nil || astFilter(file) {
			files = append(files, file)
		}
	}
	pkg, err := config.Check(path, fset, files, &info)
	if err != nil {
		return
	}
	p = new(Package)
	*p = Package{
		fset:    fset,
		stat:    stat,
		ast:     astPkg,
		Package: pkg,
		files:   files,
		info:    info,
	}
	return

}

type Var struct {
	*ast.Ident
	Type  types.TypeAndValue
	Value ast.Expr
}

func (p *Package) DefinedTypes(filter TypeFilter) (typs []types.Type) {
	if p == nil || p.info.Defs == nil {
		return nil
	}
	typs = make([]types.Type, 0, 64)
	for _, f := range p.files {
		ForEachTypeSpec(f, func(t *ast.TypeSpec) {
			if def := p.info.Defs[t.Name]; def != nil {
				if typ := def.Type(); typ != nil {
					if filter == nil || filter(typ) {
						typs = append(typs, typ)
					}
				}
			}
		})
	}
	return
}

func (p *Package) DefinedVars(filter TypeFilter) (vars []Var) {
	if p == nil || p.info.Defs == nil {
		return nil
	}
	for _, f := range p.files {
		ForEachValueSpec(f, func(spec *ast.ValueSpec) {
			ForEachVar(spec, func(id *ast.Ident, v ast.Expr) {
				if t, ok := p.info.Types[v]; ok && t.Type != nil {

					if filter == nil || filter(t.Type) {
						vars = append(vars, Var{id, t, v})
					}
				}
			})
		})
	}
	return
}

func (p *Package) QName(t types.Type) (pkgName, typName string, ok bool) {
	pkgName, typName, ok = QName(t)
	if !ok {
		return
	}
	if pkgName == p.Package.Name() {
		pkgName = ""
	}
	return
}

// LookupType looks up a named type in the package's definitions.
func (p *Package) LookupType(name string) (t *types.Named, spec *ast.TypeSpec) {
	for i, def := range p.info.Defs {
		if def == nil {
			continue
		}
		if spec = TypeSpec(i); spec == nil {
			continue
		}

		typ := def.Type()
		if typ == nil {
			continue
		}
		if t, ok := typ.(*types.Named); ok {
			if obj := t.Obj(); obj != nil && obj.Name() == name {
				return t, spec
			}
		}
	}
	return nil, nil
}

func (p *Package) Fprint(w io.Writer, node interface{}) error {
	return printer.Fprint(w, p.fset, node)
}

// func inject(fset *token.FileSet, target, pkg string) (*ast.File, error) {
// 	src := fmt.Sprintf(`package %s
// 	import _ %q
// 	`, target, pkg)
// 	filename := fmt.Sprintf("njson/inject/%s.go", pkg)
// 	return parser.ParseFile(fset, filename, src, 0)
// }
