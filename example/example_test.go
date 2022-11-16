package main

import (
	"fmt"

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
	},
	)

	for _, r := range rejections {
		fmt.Println(r)
	}

	// Output:
	// name: expected one of [alice bob] but got david
	// friends.1.id: expected 3 but got 4
}
