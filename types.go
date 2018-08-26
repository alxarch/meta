package meta

import (
	"go/token"
	"go/types"
)

func TypeImports(imports []*types.Package, t ...interface{}) []*types.Package {
	for _, t := range t {
		switch t := t.(type) {
		case *types.Named:
			if t != nil {
				imports = TypeImports(imports, t.Obj())
			}
		case *types.Scope:
			if t != nil {
				for _, name := range t.Names() {
					imports = TypeImports(imports, t.Lookup(name))
				}
			}
		case *types.Pointer:
			if t != nil {
				imports = TypeImports(imports, t.Elem())
			}
		case *types.Map:
			if t != nil {
				imports = TypeImports(imports, t.Elem())
				imports = TypeImports(imports, t.Elem())
			}
		case *types.Slice:
			if t != nil {
				imports = TypeImports(imports, t.Elem())
			}
		case *types.Chan:
			if t != nil {
				imports = TypeImports(imports, t.Elem())
			}
		case *types.Signature:
			if t != nil {
				imports = TypeImports(imports, t.Recv(), t.Params(), t.Results())
			}
		case *types.Var:
			if t != nil {
				imports = TypeImports(imports, t.Type())
			}
		case *types.Tuple:
			if t != nil {
				for i := 0; i < t.Len(); i++ {
					imports = TypeImports(imports, t.At(i))
				}
			}
		case *types.Package:
			if t != nil {
				imports = append(imports, t)
			}
		case *types.TypeName:
			if t != nil {
				imports = TypeImports(imports, t.Pkg())
			}
		}

	}
	return imports

}

func Embedded(field *types.Var) (*types.Struct, bool) {
	if field != nil && field.IsField() && field.Anonymous() {
		return Struct(field.Type())
	}
	return nil, false
}

func Base(t types.Type) string {
	return types.TypeString(t, func(*types.Package) string {
		return ""
	})

}

func String(t types.Type) (*types.Basic, bool) {
	if t != nil {
		if s, ok := t.Underlying().(*types.Basic); ok && s.Kind() == types.String {
			return s, true
		}
	}
	return nil, false

}
func Struct(t types.Type) (*types.Struct, bool) {
	if t != nil {
		if s, ok := t.Underlying().(*types.Struct); ok {
			return s, true
		}
	}
	return nil, false
}

func Pointer(t types.Type) (*types.Pointer, bool) {
	if t != nil {
		if s, ok := t.Underlying().(*types.Pointer); ok {
			return s, true
		}
	}
	return nil, false
}

func Slice(t types.Type) (*types.Slice, bool) {
	if t != nil {
		if s, ok := t.Underlying().(*types.Slice); ok {
			return s, true
		}
	}
	return nil, false
}

func Sized(t types.Type) bool {
	if t == nil {
		return false
	}
	switch t := t.Underlying().(type) {
	case *types.Pointer:
		return Sized(t.Elem())
	case *types.Array:
		return true
	case *types.Slice:
		return true
	case *types.Map:
		return true
	default:
		return false
	}
}

func Nilable(t types.Type) bool {
	if t == nil {
		return false
	}
	switch t.Underlying().(type) {
	case *types.Pointer:
		return true
	case *types.Array:
		return true
	case *types.Interface:
		return true
	case *types.Map:
		return true
	default:
		return false
	}
}

func Basic(t types.Type) (*types.Basic, bool) {
	if t != nil {
		if t, ok := t.Underlying().(*types.Basic); ok {
			return t, ok
		}
	}
	return nil, false
}

func BasicKind(t types.Type, k types.BasicKind) (*types.Basic, bool) {
	if t, ok := Basic(t); ok {
		return t, k == t.Kind()
	}
	return nil, false
}

func BasicInfo(t types.Type, info types.BasicInfo) (*types.Basic, bool) {
	if t, ok := Basic(t); ok {
		return t, t.Info()&info != 0
	}
	return nil, false
}

type TypeFilter func(t types.Type) bool

func AssignableTo(t types.Type) TypeFilter {
	return func(v types.Type) bool {
		return t != nil && v != nil && types.AssignableTo(v, t)
	}
}

func ConvertibleTo(t types.Type) TypeFilter {
	return func(v types.Type) bool {
		return t != nil && v != nil && types.ConvertibleTo(v, t)
	}
}

func IsString(t types.Type) (ok bool) {
	_, ok = String(t)
	return
}

func IsStruct(t types.Type) (ok bool) {
	_, ok = Struct(t)
	return
}

func Vars(pkg *types.Package, typ ...types.Type) (v []*types.Var) {
	v = make([]*types.Var, len(typ))
	for i, typ := range typ {
		v[i] = types.NewVar(token.NoPos, pkg, "", typ)
	}
	return
}

func MakeInterface(name string, params []types.Type, results []types.Type, v bool) *types.Interface {
	vparams := Vars(nil, params...)
	vresults := Vars(nil, results...)
	sig := types.NewSignature(nil, types.NewTuple(vparams...), types.NewTuple(vresults...), v)
	fn := types.NewFunc(token.NoPos, nil, name, sig)
	iface := types.NewInterface([]*types.Func{fn}, []*types.Named{})
	return iface.Complete()

}
