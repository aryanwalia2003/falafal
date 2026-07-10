package main

import (
	"reflect"
	"testing"
)

func TestReorderArgs(t *testing.T) {
	cases := []struct {
		name string
		in   []string
		want []string
	}{
		{
			name: "flags after path",
			in:   []string{"/some/path", "--format", "text", "--clean"},
			want: []string{"--format", "text", "--clean", "/some/path"},
		},
		{
			name: "flags before path",
			in:   []string{"--format", "text", "/some/path"},
			want: []string{"--format", "text", "/some/path"},
		},
		{
			name: "equals form",
			in:   []string{"/some/path", "--format=json"},
			want: []string{"--format=json", "/some/path"},
		},
		{
			name: "bool flag not consuming next token",
			in:   []string{"/some/path", "--all", "--top", "5"},
			want: []string{"--all", "--top", "5", "/some/path"},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := reorderArgs(c.in)
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("reorderArgs(%v) = %v, want %v", c.in, got, c.want)
			}
		})
	}
}
