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
	"strings"
)

type Package struct {
	pkg   *types.Package
	fset  *token.FileSet
	info  types.Info
	files []*ast.File
	qual  types.Qualifier
}

func (p *Package) Name() string {
	return p.pkg.Name()
}
func (p *Package) Path() string {
	return p.pkg.Path()
}
func (p *Package) Qualifier() types.Qualifier {
	if p.qual == nil {
		p.qual = types.RelativeTo(p.pkg)
	}
	return p.qual
}

func (p *Package) Code(format string, args ...interface{}) (c Code) {
	return c.QPrintf(p.Qualifier(), format, args...)
}

func (p *Package) FindImport(path string) *types.Package {
	if p == nil || p.pkg == nil {
		return nil
	}
	for _, pkg := range p.pkg.Imports() {
		if pkg.Path() == path {
			return pkg
		}
	}
	return nil
}

func IgnoreTestFiles(f os.FileInfo) bool {
	return !strings.HasSuffix(f.Name(), "_test.go")
}

type Parser struct {
	fset  *token.FileSet
	mode  parser.Mode
	files map[string][]*ast.File
}

func NewParser(mode parser.Mode) *Parser {
	p := Parser{
		fset:  token.NewFileSet(),
		mode:  mode,
		files: make(map[string][]*ast.File),
	}
	return &p
}

func (p *Parser) ParseFile(filename string, src interface{}) (string, error) {
	f, err := parser.ParseFile(p.fset, filename, src, p.mode)
	if err != nil {
		return "", err
	}
	pkgName := f.Name.String()
	p.files[pkgName] = append(p.files[pkgName], f)
	return pkgName, nil
}

func (p *Parser) ParseDir(path string, filter func(os.FileInfo) bool) error {
	packages, err := parser.ParseDir(p.fset, path, filter, p.mode)
	if err != nil {
		return err
	}
	for name, pkg := range packages {
		for _, f := range pkg.Files {
			p.files[name] = append(p.files[name], f)
		}
	}
	return nil
}

func (p *Parser) Package(name, path string, filter func(*ast.File) bool) (*Package, error) {
	files, ok := p.files[name]
	if !ok {
		return nil, fmt.Errorf("Package %s not parsed", name)
	}
	if filter != nil {
		filtered := make([]*ast.File, 0, len(files))
		for _, f := range files {
			if filter(f) {
				filtered = append(filtered, f)
			}
		}
		files = filtered
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
	pkg, err := config.Check(path, p.fset, files, &info)
	if err != nil {
		return nil, err
	}
	return &Package{
		pkg:   pkg,
		fset:  p.fset,
		info:  info,
		files: files,
		qual:  types.RelativeTo(pkg),
	}, nil

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
