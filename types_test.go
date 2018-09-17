package meta_test

import (
	"go/types"
	"testing"

	"github.com/alxarch/meta"
)

func TestMakeInterface(t *testing.T) {
	iface := meta.MakeInterface("Foo", []types.Type{}, []types.Type{}, false)
	if iface == nil {
		t.Errorf("Nil interface %v", iface)
		return
	}
	if iface.Empty() {
		t.Errorf("Empty interface %v", iface)
		return
	}

}

func TestIsString(t *testing.T) {
	name := types.NewTypeName(0, nil, "Foo", nil)
	str := types.NewNamed(name, types.Typ[types.String], nil)
	if !meta.IsString(str) {
		t.Errorf("IsString not")
	}
	if !meta.IsString(types.Typ[types.String]) {
		t.Errorf("IsString not")
	}

}
