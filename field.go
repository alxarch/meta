package meta

import (
	"go/types"
	"sort"
)

type FieldIndex struct {
	Index int
	*types.Var
	Tag string
}

func (f FieldIndex) Name() string {
	return FieldName(f.Var)
}

type FieldPath []FieldIndex

// Copy creates a copy of a path.
func (p FieldPath) Copy() FieldPath {
	cp := make([]FieldIndex, len(p))
	copy(cp, p)
	return cp
}

func (p FieldPath) String() string {
	buf := make([]byte, 0, len(p)*16)
	for i := range p {
		buf = append(buf, '.')
		buf = append(buf, FieldName(p[i].Var)...)
	}
	return string(buf)
}

// ShortestPath compares the paths of two fields.
func ShortestPath(a, b FieldPath) int {
	if len(a) < len(b) {
		return -1
	}
	if len(b) < len(a) {
		return 1
	}
	for i := range a {
		if a[i].Index < b[i].Index {
			return -1
		}
		if b[i].Index < a[i].Index {
			return 1
		}
	}
	return 0
}

type Field struct {
	*types.Var
	Tag  string
	Path FieldPath
}

func (f Field) WithTag(key string) Field {
	for i, ff := range f.Path {
		if HasTag(ff.Tag, key) {
			return Field{
				Var:  ff.Var,
				Tag:  ff.Tag,
				Path: f.Path[:i+1],
			}
		}
	}
	return f
}

type Fields map[string][]Field

func FieldName(field *types.Var) string {
	if field == nil || !field.IsField() {
		return ""
	}
	if field.Anonymous() {
		return fieldTypeName(field.Type())
	}
	return field.Name()
}
func fieldTypeName(t types.Type) string {
	switch t := t.(type) {
	case *types.Named:
		return t.Obj().Name()
	case *types.Pointer:
		return fieldTypeName(t.Elem())
	default:
		return t.String()
	}
}

func (fields Fields) get(name string) *Field {
	if fields, ok := fields[name]; ok && len(fields) > 0 {
		return &fields[0]
	}
	return nil
}

// Add adds a field to a field map handling duplicates.
func (fields Fields) Add(field Field) Fields {
	name := field.Name()
	if fields == nil {
		fields = make(map[string][]Field)
		fields[name] = []Field{field}
		return fields
	}
	existing := fields.get(name)
	if existing == nil {
		fields[name] = []Field{field}
		return fields
	}

	switch ShortestPath(existing.Path, field.Path) {
	case 0:
		*existing = field
	case 1:
		fields[name] = append([]Field{field}, fields[name]...)
	default:
		sorted := append(fields[name], field)
		sort.SliceStable(sorted, func(i, j int) bool {
			return ShortestPath(sorted[i].Path, sorted[j].Path) == -1
		})
		fields[name] = sorted
	}
	return fields
}

// NewFields creates a field map for a struct.
func NewFields(s *types.Struct, embed bool) Fields {
	if s == nil {
		return nil
	}
	fields := Fields(make(map[string][]Field))
	fields = fields.Merge(s, embed, nil)
	return fields

}
func (fields Fields) Merge(s *types.Struct, embed bool, path FieldPath) Fields {
	if fields == nil || s == nil {
		return nil
	}

	depth := len(path)

	for i := 0; i < s.NumFields(); i++ {
		field := s.Field(i)
		tag := s.Tag(i)
		path = append(path[:depth], FieldIndex{i, field, tag})
		if embed && field.Anonymous() {
			t := field.Type().Underlying()
			if ptr, isPointer := t.(*types.Pointer); isPointer {
				t = ptr.Elem()
			}
			if tt, ok := t.Underlying().(*types.Struct); ok {
				// embedded struct
				fields = fields.Merge(tt, embed, path)
				continue
			}
		}

		fields = fields.Add(Field{field, tag, path.Copy()})

	}
	return fields
}
