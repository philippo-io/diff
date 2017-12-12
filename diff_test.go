package diff

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type tistruct struct {
	Name  string `diff:"name,identifier"`
	Value int    `diff:"value"`
}

type tstruct struct {
	Name          string            `diff:"name"`
	Value         int               `diff:"value"`
	Bool          bool              `diff:"bool"`
	Values        []string          `diff:"values"`
	Map           map[string]string `diff:"map"`
	Pointer       *string           `diff:"pointer"`
	Ignored       bool              `diff:"-"`
	Identifiables []tistruct        `diff:"identifiables"`
}

func sptr(s string) *string {
	return &s
}

func TestDiff(t *testing.T) {
	cases := []struct {
		Name      string
		A, B      interface{}
		Changelog Changelog
		Error     error
	}{
		{
			"int-slice-insert", []int{1, 2, 3}, []int{1, 2, 3, 4},
			Changelog{
				Change{Type: CREATE, Path: []string{"3"}, To: 4},
			},
			nil,
		},
		{
			"int-slice-delete", []int{1, 2, 3}, []int{1, 3},
			Changelog{
				Change{Type: DELETE, Path: []string{"1"}, From: 2},
			},
			nil,
		},
		{
			"int-slice-insert-delete", []int{1, 2, 3}, []int{1, 3, 4},
			Changelog{
				Change{Type: DELETE, Path: []string{"1"}, From: 2},
				Change{Type: CREATE, Path: []string{"2"}, To: 4},
			},
			nil,
		},
		{
			"string-slice-insert", []string{"1", "2", "3"}, []string{"1", "2", "3", "4"},
			Changelog{
				Change{Type: CREATE, Path: []string{"3"}, To: "4"},
			},
			nil,
		},
		{
			"string-slice-delete", []string{"1", "2", "3"}, []string{"1", "3"},
			Changelog{
				Change{Type: DELETE, Path: []string{"1"}, From: "2"},
			},
			nil,
		},
		{
			"string-slice-insert-delete", []string{"1", "2", "3"}, []string{"1", "3", "4"},
			Changelog{
				Change{Type: DELETE, Path: []string{"1"}, From: "2"},
				Change{Type: CREATE, Path: []string{"2"}, To: "4"},
			},
			nil,
		},
		{
			"comparable-slice-insert", []tistruct{{"one", 1}}, []tistruct{{"one", 1}, {"two", 2}},
			Changelog{
				Change{Type: CREATE, Path: []string{"two"}, To: tistruct{"two", 2}},
			},
			nil,
		},
		{
			"comparable-slice-delete", []tistruct{{"one", 1}, {"two", 2}}, []tistruct{{"one", 1}},
			Changelog{
				Change{Type: DELETE, Path: []string{"two"}, From: tistruct{"two", 2}},
			},
			nil,
		},
		{
			"comparable-slice-update", []tistruct{{"one", 1}}, []tistruct{{"one", 50}},
			Changelog{
				Change{Type: UPDATE, Path: []string{"one", "value"}, From: 1, To: 50},
			},
			nil,
		},
		{
			"struct-string-update", tstruct{Name: "one"}, tstruct{Name: "two"},
			Changelog{
				Change{Type: UPDATE, Path: []string{"name"}, From: "one", To: "two"},
			},
			nil,
		},
		{
			"struct-int-update", tstruct{Value: 1}, tstruct{Value: 50},
			Changelog{
				Change{Type: UPDATE, Path: []string{"value"}, From: 1, To: 50},
			},
			nil,
		},
		{
			"struct-bool-update", tstruct{Bool: true}, tstruct{Bool: false},
			Changelog{
				Change{Type: UPDATE, Path: []string{"bool"}, From: true, To: false},
			},
			nil,
		},
		{
			"struct-map-update", tstruct{Map: map[string]string{"test": "123"}}, tstruct{Map: map[string]string{"test": "456"}},
			Changelog{
				Change{Type: UPDATE, Path: []string{"map"}, From: map[string]string{"test": "123"}, To: map[string]string{"test": "456"}},
			},
			nil,
		},
		{
			"struct-string-pointer-update", tstruct{Pointer: sptr("test")}, tstruct{Pointer: sptr("test2")},
			Changelog{
				Change{Type: UPDATE, Path: []string{"pointer"}, From: "test", To: "test2"},
			},
			nil,
		},
		{
			"struct-nil-string-pointer-update", tstruct{Pointer: nil}, tstruct{Pointer: sptr("test")},
			Changelog{
				Change{Type: UPDATE, Path: []string{"pointer"}, From: nil, To: sptr("test")},
			},
			nil,
		},
		{
			"struct-generic-slice-insert", tstruct{Values: []string{"one"}}, tstruct{Values: []string{"one", "two"}},
			Changelog{
				Change{Type: CREATE, Path: []string{"values", "1"}, From: nil, To: "two"},
			},
			nil,
		},
		{
			"struct-identifiable-slice-insert", tstruct{Identifiables: []tistruct{{"one", 1}}}, tstruct{Identifiables: []tistruct{{"one", 1}, {"two", 2}}},
			Changelog{
				Change{Type: CREATE, Path: []string{"identifiables", "two"}, From: nil, To: tistruct{"two", 2}},
			},
			nil,
		},
		{
			"struct-generic-slice-delete", tstruct{Values: []string{"one", "two"}}, tstruct{Values: []string{"one"}},
			Changelog{
				Change{Type: DELETE, Path: []string{"values", "1"}, From: "two", To: nil},
			},
			nil,
		},
		{
			"struct-identifiable-slice-delete", tstruct{Identifiables: []tistruct{{"one", 1}, {"two", 2}}}, tstruct{Identifiables: []tistruct{{"one", 1}}},
			Changelog{
				Change{Type: DELETE, Path: []string{"identifiables", "two"}, From: tistruct{"two", 2}, To: nil},
			},
			nil,
		},
		{
			"omittable", tstruct{Ignored: false}, tstruct{Ignored: true},
			Changelog{},
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			cl, err := Diff(tc.A, tc.B)

			assert.Equal(t, tc.Error, err)
			assert.Equal(t, len(tc.Changelog), len(cl))

			for i, c := range cl {
				assert.Equal(t, tc.Changelog[i].Type, c.Type)
				assert.Equal(t, tc.Changelog[i].Path, c.Path)
				assert.Equal(t, tc.Changelog[i].From, c.From)
				assert.Equal(t, tc.Changelog[i].To, c.To)
			}
		})
	}

}