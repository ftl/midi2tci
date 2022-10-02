package ctrl

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTranslate(t *testing.T) {
	tt := []struct {
		desc     string
		r        ValueRange
		value    uint8
		expected int
	}{
		{
			desc:     "begin1",
			r:        StaticRange{-50, 50},
			value:    0,
			expected: -50,
		},
		{
			desc:     "center1",
			r:        StaticRange{-50, 50},
			value:    0x40,
			expected: 0,
		},
		{
			desc:     "end1",
			r:        StaticRange{-50, 50},
			value:    0x7f,
			expected: 50,
		},
		{
			desc:     "begin2",
			r:        StaticRange{-50, 49},
			value:    0,
			expected: -50,
		},
		{
			desc:     "center2",
			r:        StaticRange{-50, 49},
			value:    0x40,
			expected: 0,
		},
		{
			desc:     "end2",
			r:        StaticRange{-50, 49},
			value:    0x7f,
			expected: 49,
		},
		{
			desc:     "infinite",
			r:        InfiniteRange{},
			value:    123,
			expected: 123,
		},
	}
	for _, tc := range tt {
		t.Run(tc.desc, func(t *testing.T) {
			actual := Translate(tc.r, tc.value)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestProject(t *testing.T) {
	tt := []struct {
		desc     string
		r        ValueRange
		value    int
		expected uint8
	}{
		{
			desc:     "begin1",
			r:        StaticRange{-50, 50},
			value:    -50,
			expected: 0,
		},
		{
			desc:     "center1",
			r:        StaticRange{-50, 50},
			value:    0,
			expected: 0x40,
		},
		{
			desc:     "end1",
			r:        StaticRange{-50, 50},
			value:    50,
			expected: 0x7f,
		},
		{
			desc:     "begin2",
			r:        StaticRange{-50, 49},
			value:    -50,
			expected: 0,
		},
		{
			desc:     "center2",
			r:        StaticRange{-50, 49},
			value:    0,
			expected: 0x40,
		},
		{
			desc:     "end2",
			r:        StaticRange{-50, 49},
			value:    49,
			expected: 0x7f,
		},
		{
			desc:     "below",
			r:        StaticRange{-50, 50},
			value:    -51,
			expected: 0,
		},
		{
			desc:     "above",
			r:        StaticRange{-50, 50},
			value:    51,
			expected: 0x7f,
		},
		{
			desc:     "infinite",
			r:        InfiniteRange{},
			value:    123,
			expected: 123,
		},
	}
	for _, tc := range tt {
		t.Run(tc.desc, func(t *testing.T) {
			actual := Project(tc.r, tc.value)
			assert.Equal(t, tc.expected, actual)
		})
	}
}
