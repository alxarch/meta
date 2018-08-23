package meta

import (
	"go/types"
	"sort"
)

type FieldIndex struct {
	Index int
	Type  types.Type
	Name  string
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
		buf = append(buf, p[i].Name...)
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
	Tags []Tag
	Path FieldPath
}

type FieldMap map[string][]Field

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

func (f *Field) Tag(key string) *Tag {
	for i := range f.Tags {
		if f.Tags[i].Key == key {
			return &f.Tags[i]
		}
	}
	return nil
}

func (fields FieldMap) Get(name string) *Field {
	if fields, ok := fields[name]; ok && len(fields) > 0 {
		return &fields[0]
	}
	return nil
}
func (f *Field) AddTag(tag ...Tag) {
	if len(f.Tags) == 0 {
		f.Tags = append(f.Tags, tag...)
		return
	}
	for _, t := range tag {
		oldTag := f.Tag(t.Key)
		if oldTag == nil {
			f.Tags = append(f.Tags, t)
			continue
		}
		*oldTag = t
	}
	return
}

// Add adds a field to a field map handling duplicates.
func (fields FieldMap) Add(field Field) FieldMap {
	name := field.Name()
	if fields == nil {
		fields = make(map[string][]Field)
		fields[name] = []Field{field}
		return fields
	}
	existing := fields.Get(name)
	if existing == nil {
		fields[name] = []Field{field}
		return fields
	}

	switch ShortestPath(existing.Path, field.Path) {
	case 0:
		existing.AddTag(field.Tags...)
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

// Merge merges a struct's fields to a field map.
func NewFields(s *types.Struct, tag string, embed bool) FieldMap {
	if s == nil {
		return nil
	}
	fields := FieldMap(make(map[string][]Field))
	fields = fields.Merge(s, tag, embed, nil)
	return fields

}
func (fields FieldMap) Merge(s *types.Struct, tagKey string, embed bool, path FieldPath) FieldMap {
	if fields == nil || s == nil {
		return nil
	}

	depth := len(path)

	for i := 0; i < s.NumFields(); i++ {
		field := s.Field(i)
		tag, tagged := ParseTag(s.Tag(i), tagKey)
		path = append(path[:depth], FieldIndex{i, field.Type(), field.Name()})
		if !tagged && field.Anonymous() {
			t := Resolve(field.Type())
			if ptr, isPointer := t.(*types.Pointer); isPointer {
				t = ptr.Elem()
			}
			tt := Resolve(t)
			if tt, ok := tt.(*types.Struct); ok && embed {
				// embedded struct
				fields = fields.Merge(tt, tagKey, embed, path)
				continue
			}
		}

		f := Field{Var: field, Path: path.Copy()}
		if tagged {
			f.Tags = []Tag{tag}
		}
		fields = fields.Add(f)

	}
	return fields
}
