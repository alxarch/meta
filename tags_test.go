package meta_test

import (
	"testing"

	"github.com/alxarch/meta"
)

func TestParseTag(t *testing.T) {
	tag, ok := meta.ParseTag(`json:"Foo,omitempty"`, "json")
	if !ok {
		t.Errorf("Tag not found")
		return
	}
	if tag.Name != "Foo" {
		t.Errorf("Invalid tag name %s", tag.Name)
	}
	if tag.Params.Get("omitempty") != "omitempty" {
		t.Errorf("Invalid tag params %s", tag.Params.Values())

	}
	if !tag.Params.Has("omitempty") {
		t.Errorf("Invalid tag params %s", tag.Params.Values())
	}

}
