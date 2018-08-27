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
	fset    *token.FileSet
	info    types.Info
	ast     *ast.Package
	files   []*ast.File
	stat    os.FileInfo
	qual    types.Qualifier
	typdefs map[string]*types.Named
}

func (p *Package) Qualifier() types.Qualifier {
	if p.qual == nil {
		p.qual = types.RelativeTo(p.Package)
	}
	return p.qual
}

func (p *Package) Code(format string, args ...interface{}) (c Code) {
	return c.QPrintf(p.Qualifier(), format, args...)
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
		qual:    types.RelativeTo(pkg),
	}
	return

}
func (p *Package) NamedTypes() map[string]*types.Named {
	named := map[string]*types.Named{}
	p.DefinedTypes(func(t types.Type) bool {
		if t, ok := t.(*types.Named); ok {
			named[t.Obj().Name()] = t
		}
		return false
	})
	return named
}

type Var struct {
	*ast.Ident
	Type  types.TypeAndValue
	Value ast.Expr
}

func (p *Package) DefinedTypes(filter TypeFilter) (typs []types.Type) {
	return p.DefinedTypesN(-1, filter)
}
func (p *Package) DefinedTypesN(maxResults int, filter TypeFilter) (typs []types.Type) {
	if p == nil || p.info.Defs == nil {
		return nil
	}
	if maxResults == 0 {
		maxResults = -1
	}
	typs = make([]types.Type, 0, 64)
	for _, f := range p.files {
		if maxResults == 0 {
			return
		}
		ForEachTypeSpec(f, func(t *ast.TypeSpec) {
			if maxResults == 0 {
				return
			}
			if def := p.info.Defs[t.Name]; def != nil {
				if typ := def.Type(); typ != nil {
					if filter == nil || filter(typ) {
						typs = append(typs, typ)
					}
					maxResults--
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

func (p *Package) TypeString(t types.Type) string {
	return types.TypeString(t, p.qual)
}

// LookupType looks up a named type in the package's definitions.
func (p *Package) LookupType(name string) (T *types.Named) {
	p.DefinedTypesN(-1, func(t types.Type) bool {
		if t != nil {
			if t, ok := t.(*types.Named); ok && t.Obj() != nil {
				if t.Obj().Name() == name {
					T = t
					return true
				}
			}
		}
		return false
	})
	return
}

func (p *Package) Fprint(w io.Writer, node interface{}) error {
	return printer.Fprint(w, p.fset, node)
}

func MustImport(path string) *types.Package {
	imp := importer.Default()
	pkg, err := imp.Import(path)
	if err != nil {
		panic(err)
	}
	return pkg
}
