package desire

import (
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestDesire(t *testing.T) {
	for name, tt := range map[string]struct {
		got    any
		desire any
		want   []Rejection
	}{
		"primitive ok": {
			1,
			1,
			nil,
		},
		"primitive reject": {
			1,
			2,
			[]Rejection{
				{nil, "expected 2 but got 1"},
			},
		},
		"map ok": {
			map[string]int{"a": 1, "b": 2},
			map[string]int{"a": 1, "b": 2},
			nil,
		},
		"map ng value": {
			map[string]int{"a": 1, "b": 2},
			map[string]int{"a": 1, "b": 3},
			[]Rejection{
				{Path{"b"}, "expected 3 but got 2"},
			},
		},
		"map ng key": {
			map[string]int{"a": 1, "b": 2},
			map[string]int{"a": 1, "c": 2},
			[]Rejection{
				{Path{"b"}, "expected undefined but exists with value 2"},
				{Path{"c"}, "expected 2 but undefined"},
			},
		},
		"slice ok": {
			[]int{1, 2, 3},
			[]int{1, 2, 3},
			nil,
		},
		"slice ng value": {
			[]int{1, 2, 3},
			[]int{1, 2, 4},
			[]Rejection{
				{Path{"2"}, "expected 4 but got 3"},
			},
		},
		"slice ng size": {
			[]int{1, 2, 3},
			[]int{1, 2},
			[]Rejection{
				{Path{"2"}, "expected undefined but exists with value 3"},
			},
		},
		"map with slice ng": {
			map[string]any{
				"a": 1,
				"b": []int{1, 2, 3},
			},
			map[string]any{
				"a": 1,
				"b": []int{1, 2, 4},
			},
			[]Rejection{
				{Path{"b", "2"}, "expected 4 but got 3"},
			},
		},
		"Partial ok": {
			map[string]any{
				"a": 1,
				"b": []int{1, 2, 3},
				"c": 4,
			},
			Partial{
				"a": 1,
				"b": Partial{
					1: 2,
				},
			},
			nil,
		},
		"Partial ng value": {
			map[string]any{
				"a": 1,
				"b": []int{1, 2, 3},
			},
			Partial{
				"b": []int{0, 2, 3},
			},
			[]Rejection{
				{Path{"b", "0"}, "expected 0 but got 1"},
			},
		},
		"Partial ng undefined key": {
			map[any]any{
				"a":  1,
				"b":  []int{1, 2, 3},
				"10": 4,
			},
			Partial{
				"b": []int{1, 2, 3},
				"c": 2,
				10:  4,
			},
			[]Rejection{
				{Path{"c"}, "expected 2 but undefined"},
				{Path{"10"}, "expected 4 but undefined"},
			},
		},
		"Partial ng different key type": {
			map[int]any{
				1: 1,
				2: []int{1, 2, 3},
			},
			Partial{
				"c": 2,
				2: Partial{
					"a": 1,
				},
			},
			[]Rejection{
				{Path{"c"}, "expected key type string but got int"},
				{Path{"2"}, "key type of Partial must be int for slice or array"},
			},
		},
		"Partial ng index out of range": {
			[]int{1, 2, 3},
			Partial{
				-1: 0,
				3:  4,
			},
			[]Rejection{
				{Path{"-1"}, "index out of range for size 3"},
				{Path{"3"}, "index out of range for size 3"},
			},
		},
		"List ok": {
			[]int{1, 2, 3},
			List{1, 2, 3},
			nil,
		},
		"List ng value": {
			[]int{1, 2, 3},
			List{1, 2, 4},
			[]Rejection{
				{Path{"2"}, "expected 4 but got 3"},
			},
		},
		"List ng size larger": {
			[]int{1, 2, 3},
			List{1, 2},
			[]Rejection{
				{Path{"2"}, "expected undefined but exists with value 3"},
			},
		},
		"List ng size smaller": {
			[]int{1, 2, 3},
			List{1, 2, 3, 4},
			[]Rejection{
				{Path{"3"}, "expected 4 but undefined"},
			},
		},
		"NotZero ok": {
			1,
			NotZero(),
			nil,
		},
		"NotZero ng zero": {
			0,
			NotZero(),
			[]Rejection{
				{nil, "expected non-zero value but got 0"},
			},
		},
		"NotZeroT ok": {
			1,
			NotZeroT[int](),
			nil,
		},
		"NotZeroT ng zero": {
			0,
			NotZeroT[int](),
			[]Rejection{
				{nil, "expected non-zero value but got 0"},
			},
		},
		"NotZeroT ng different type": {
			"aa",
			NotZeroT[int](),
			[]Rejection{
				{nil, "expected type int but got string"},
			},
		},
		"OneOf ok": {
			1,
			OneOf(1, 2, 3),
			nil,
		},
		"OneOf ng": {
			0,
			OneOf(1, 2, 3),
			[]Rejection{
				{nil, "expected one of [1 2 3] but got 0"},
			},
		},
		"All ok": {
			1,
			All(NotZero(), OneOf(1, 2, 3)),
			nil,
		},
		"All ng": {
			0,
			All(OneOf(0, 1, 2), NotZero()),
			[]Rejection{
				{nil, "expected non-zero value but got 0"},
			},
		},
		"Any ok": {
			[]any{1, "a", []int{}, ([]int)(nil), struct{}{}},
			List{Any(), Any(), Any(), Any(), Any()},
			nil,
		},
		"complex ok": {
			map[string]any{
				"id":   1,
				"name": "alice",
				"age":  20,
				"friends": []map[string]any{
					{
						"id":   2,
						"name": "bob",
					},
					{
						"id":   3,
						"name": "charlie",
					},
				},
			},
			Partial{
				"id":   NotZeroT[int](),
				"name": OneOf("alice", "bob"),
				// not validate "age"
				"friends": List{
					NotZero(),
					map[string]any{
						"id":   3,
						"name": "charlie",
					},
				},
			},
			nil,
		},
	} {
		t.Run(name, func(t *testing.T) {
			got := Desire(tt.got, tt.desire)
			sort.Slice(got, func(i, j int) bool { return got[i].Path.String() < got[j].Path.String() })
			sort.Slice(tt.want, func(i, j int) bool { return tt.want[i].Path.String() < tt.want[j].Path.String() })
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("(-want, +got)\n%s", diff)
			}
		})
	}
}
