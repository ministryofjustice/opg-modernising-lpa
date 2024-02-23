package search

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPagination(t *testing.T) {
	testCases := map[string]struct {
		Total, CurrentPage, TotalPages, PageSize int
		Start, End, Previous, Next               int
		HasPrevious, HasNext                     bool
		Pages                                    []int
	}{
		"empty": {
			Total:       0,
			CurrentPage: 1,
			TotalPages:  0,
			PageSize:    25,
			Start:       1,
			End:         0,
			HasPrevious: false,
			HasNext:     false,
			Pages:       []int{},
		},
		"one-item": {
			Total:       1,
			CurrentPage: 1,
			TotalPages:  1,
			PageSize:    25,
			Start:       1,
			End:         1,
			HasPrevious: false,
			HasNext:     false,
			Pages:       []int{1},
		},
		"one-page": {
			Total:       25,
			CurrentPage: 1,
			TotalPages:  1,
			PageSize:    25,
			Start:       1,
			End:         25,
			HasPrevious: false,
			HasNext:     false,
			Pages:       []int{1},
		},
		"many-pages": {
			Total:       76,
			CurrentPage: 2,
			TotalPages:  4,
			PageSize:    25,
			Start:       26,
			End:         50,
			HasPrevious: true,
			Previous:    1,
			HasNext:     true,
			Next:        3,
			Pages:       []int{1, 2, 3, 4},
		},
		"first-of-many-pages": {
			Total:       76,
			CurrentPage: 1,
			TotalPages:  4,
			PageSize:    25,
			Start:       1,
			End:         25,
			HasPrevious: false,
			HasNext:     true,
			Next:        2,
			Pages:       []int{1, 2, 3, 4},
		},
		"last-of-many-pages": {
			Total:       76,
			CurrentPage: 4,
			TotalPages:  4,
			PageSize:    25,
			Start:       76,
			End:         76,
			HasPrevious: true,
			Previous:    3,
			HasNext:     false,
			Pages:       []int{1, 2, 3, 4},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			pagination := newPagination(tc.Total, tc.CurrentPage, tc.PageSize)

			assert.Equal(tc.Start, pagination.Start())
			assert.Equal(tc.End, pagination.End())
			assert.Equal(tc.HasPrevious, pagination.HasPrevious())
			if tc.HasPrevious {
				assert.Equal(tc.Previous, pagination.Previous())
			}
			assert.Equal(tc.HasNext, pagination.HasNext())
			if tc.HasNext {
				assert.Equal(tc.Next, pagination.Next())
			}
			assert.Equal(tc.Pages, pagination.Pages())
		})
	}
}

func TestPaginationPagesWhenOverflow(t *testing.T) {
	testCases := map[int][]int{
		1:  {1, 2, -1, 10},
		2:  {1, 2, 3, -1, 10},
		3:  {1, 2, 3, 4, -1, 10},
		4:  {1, -1, 3, 4, 5, -1, 10},
		5:  {1, -1, 4, 5, 6, -1, 10},
		6:  {1, -1, 5, 6, 7, -1, 10},
		7:  {1, -1, 6, 7, 8, -1, 10},
		8:  {1, -1, 7, 8, 9, 10},
		9:  {1, -1, 8, 9, 10},
		10: {1, -1, 9, 10},
	}

	for current, pages := range testCases {
		t.Run(fmt.Sprintf("Page%d", current), func(t *testing.T) {
			pagination := newPagination(250, current, 25)

			assert.Equal(t, pages, pagination.Pages())
		})
	}
}
