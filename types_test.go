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
