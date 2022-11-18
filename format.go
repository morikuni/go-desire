package desire

import (
	"bytes"
	"sort"
	"strings"
)

func FormatRejections(rs []Rejection) string {
	if len(rs) == 0 {
		return ""
	}
	root := make(map[string]any)
	for _, r := range rs {
		if len(r.Path) == 0 {
			root[""] = r.Reason
			continue
		}
		m := root
		for _, p := range r.Path {
			field, ok := m[p].(map[string]any)
			if !ok {
				field = make(map[string]any)
				m[p] = field
			}
			m = field
		}
		m[""] = r.Reason
	}
	buf := &bytes.Buffer{}
	print(buf, root, 0)
	return buf.String()
}

func print(buf *bytes.Buffer, m map[string]any, depth int) {
	if v, ok := m[""]; ok {
		buf.WriteString(v.(string))
		delete(m, "")
	}
	if len(m) == 0 {
		return
	}
	buf.WriteString("{\n")
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)

	}
	sort.Strings(keys)
	for _, k := range keys {
		v := m[k]
		buf.WriteString(strings.Repeat("    ", depth+1))
		buf.WriteString(k)
		buf.WriteString(": ")
		print(buf, v.(map[string]any), depth+1)
		buf.WriteString(",\n")
	}
	buf.WriteString(strings.Repeat("    ", depth))
	buf.WriteString("}")
}
