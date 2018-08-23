package meta

import (
	"go/ast"
	"go/token"
)

func TypeSpec(i *ast.Ident) *ast.TypeSpec {
	if i == nil || i.Obj == nil || i.Obj.Kind != ast.Typ || i.Obj.Decl == nil {
		return nil
	}
	t, _ := i.Obj.Decl.(*ast.TypeSpec)
	return t
}

func ForEachTypeSpec(f *ast.File, fn func(t *ast.TypeSpec)) {
	if fn == nil {
		return
	}
	for _, decl := range f.Decls {
		if decl, ok := decl.(*ast.GenDecl); ok && decl.Tok == token.TYPE {
			for _, spec := range decl.Specs {
				spec, _ := spec.(*ast.TypeSpec)
				fn(spec)
			}
		}
	}
}

func ForEachVar(spec *ast.ValueSpec, fn func(id *ast.Ident, v ast.Expr)) {
	if spec == nil || fn == nil {
		return
	}
	for i, j := 0, 0; i < len(spec.Names) && j < len(spec.Values); i++ {
		id := spec.Names[i]
		if id.Name == "_" {
			continue
		}
		expr := spec.Values[j]
		j++
		fn(id, expr)
	}
}

func ForEachStruct(f *ast.File, fn func(s *ast.StructType, t *ast.TypeSpec)) {
	if f == nil || fn == nil {
		return
	}
	ForEachTypeSpec(f, func(t *ast.TypeSpec) {
		if t.Type != nil {
			if s, ok := t.Type.(*ast.StructType); ok {
				fn(s, t)
			}
		}
	})
}

func ForEachValueSpec(f *ast.File, fn func(spec *ast.ValueSpec)) {
	if fn == nil {
		return
	}
	for _, decl := range f.Decls {
		if decl, ok := decl.(*ast.GenDecl); ok && decl.Tok == token.VAR {
			for _, spec := range decl.Specs {
				spec, _ := spec.(*ast.ValueSpec)
				fn(spec)
			}
		}
	}

}

func AppendStructs(specs []*ast.TypeSpec, f *ast.File) []*ast.TypeSpec {
	for _, d := range f.Decls {
		if d, ok := d.(*ast.GenDecl); ok && d.Tok == token.TYPE {
			for _, spec := range d.Specs {
				spec := spec.(*ast.TypeSpec)
				if spec.Type == nil {
					continue
				}
				if _, ok := spec.Type.(*ast.StructType); ok {
					specs = append(specs, spec)
				}
			}
		}
	}
	return specs
}
