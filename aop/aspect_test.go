package aop

import (
	"testing"
)

func TestSortAspectsByOrder_NilElements(t *testing.T) {
	a1 := &AspectMeta{Order: 1}
	a2 := &AspectMeta{Order: 2}
	aspects := []*AspectMeta{nil, a2, nil, a1}

	SortAspectsByOrder(aspects)

	if aspects[0] != nil && aspects[0].Order != 1 {
		t.Errorf("expected first non-nil to have Order 1 after sort")
	}
}

func TestSortAspectsByOrder_NilSlice(t *testing.T) {
	SortAspectsByOrder(nil)
}

func TestSortAspectsByOrder_AllNil(t *testing.T) {
	SortAspectsByOrder([]*AspectMeta{nil, nil, nil})
}
