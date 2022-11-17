package main

import (
	"fmt"
	"sort"

	"github.com/morikuni/go-desire"
)

func Example() {
	val := map[string]any{
		"id":   1,
		"name": "david",
		"age":  20,
		"friends": []map[string]any{
			{
				"id":   2,
				"name": "bob",
			},
			{
				"id":   4,
				"name": "charlie",
			},
		},
	}

	rejections := desire.Desire(val, desire.Partial{
		"id":   desire.NotZeroT[int](),
		"name": desire.OneOf("alice", "bob"),
		// not validate "age"
		"friends": desire.List{
			desire.NotZero(),
			map[string]any{
				"id":   3,
				"name": "charlie",
			},
		},
	})

	sort.Slice(rejections, func(i, j int) bool {
		return rejections[i].Path.String() < rejections[j].Path.String()
	})
	for _, r := range rejections {
		fmt.Println(r)
	}

	// Output:
	// friends.1.id: expected 3 but got 4
	// name: expected one of [alice bob] but got david
}
