package desire

import (
	"fmt"
	"strconv"

	"github.com/google/go-cmp/cmp"
)

type cmpReporter struct {
	path       cmp.Path
	rejections []Rejection
}

func (r *cmpReporter) PushStep(ps cmp.PathStep) {
	r.path = append(r.path, ps)
}

func (r *cmpReporter) Report(rs cmp.Result) {
	if !rs.Equal() {
		desired, got := r.path.Last().Values()
		if !desired.IsValid() {
			r.rejections = append(r.rejections, Rejection{
				r.Path(),
				fmt.Sprintf("expected undefined but exists with value %v", got),
			})
		} else if !got.IsValid() {
			r.rejections = append(r.rejections, Rejection{
				r.Path(),
				fmt.Sprintf("expected %v but undefined", desired),
			})
		} else {
			r.rejections = append(r.rejections, Rejection{
				r.Path(),
				fmt.Sprintf("expected %v but got %v", desired, got),
			})
		}
	}
}

func (r *cmpReporter) PopStep() {
	r.path = r.path[:len(r.path)-1]
}

func (r *cmpReporter) Path() Path {
	result := make(Path, 0, len(r.path))
	for _, path := range r.path {
		switch p := path.(type) {
		case cmp.MapIndex:
			result = append(result, p.Key().String())
		case cmp.SliceIndex:
			iDesire, iGot := p.SplitKeys()
			if iDesire == -1 {
				iDesire = iGot
			}
			result = append(result, strconv.Itoa(iDesire))
		case cmp.StructField:
			result = append(result, p.Name())
		}
	}
	return result
}
