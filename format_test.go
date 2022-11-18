package desire

import (
	"testing"
)

func TestFormatRejections(t *testing.T) {
	rs := []Rejection{
		{Path{"a", "b", "c"}, "test 1"},
		{Path{"a", "b", "d"}, "test 2"},
		{Path{"a", "e"}, "test 3"},
		{Path{"f"}, "test 4"},
	}
	got := FormatRejections(rs)
	equal(t, got, `{
    a: {
        b: {
            c: test 1,
            d: test 2,
        },
        e: test 3,
    },
    f: test 4,
}`)
}
