package meta

import "go/types"

// Resolve resolves to an underlying non named type.
func Resolve(typ types.Type) types.Type {
	for typ != nil {
		if _, ok := typ.(*types.Named); ok {
			typ = typ.Underlying()
		} else {
			break
		}
	}
	return typ
}

func QName(t types.Type) (string, string, bool) {
	switch t := t.(type) {
	case *types.Named:
		obj := t.Obj()
		if obj == nil {
			return "", t.String(), false
		}
		pkg := obj.Pkg()
		if pkg != nil {
			return pkg.Name(), obj.Name(), true
		}
		return "", obj.Name(), true
	case *types.Pointer:
		return QName(t.Elem())
	default:
		return "", t.String(), false
	}
}

func Embedded(field *types.Var) (*types.Struct, bool) {
	if field != nil && field.IsField() && field.Anonymous() {
		return Struct(field.Type())
	}
	return nil, false
}

func Struct(t types.Type) (*types.Struct, bool) {
	if t := Resolve(t); t != nil {
		if s, ok := t.(*types.Struct); ok {
			return s, true
		}
	}
	return nil, false
}

func Pointer(t types.Type) (*types.Pointer, bool) {
	if t := Resolve(t); t != nil {
		if s, ok := t.(*types.Pointer); ok {
			return s, true
		}
	}
	return nil, false
}

func Slice(t types.Type) (*types.Slice, bool) {
	if t := Resolve(t); t != nil {
		if s, ok := t.(*types.Slice); ok {
			return s, true
		}
	}
	return nil, false
}

func Sized(t types.Type) bool {
	if t = Resolve(t); t == nil {
		return false
	}
	switch t := t.(type) {
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
	if t = Resolve(t); t == nil {
		return false
	}
	switch t.(type) {
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
	if t = Resolve(t); t != nil {
		if t, ok := t.(*types.Basic); ok {
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

func IsStruct(t types.Type) (ok bool) {
	_, ok = Struct(t)
	return
}
