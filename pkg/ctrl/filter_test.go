package ctrl

import (
	"testing"

	"github.com/ftl/tci/client"
	"github.com/stretchr/testify/assert"
)

func TestFilterShape_BoundsAndSliderValue(t *testing.T) {
	tt := []struct {
		desc        string
		shape       filterShape
		value       int
		expectedMin int
		expectedMax int
	}{
		{
			desc:        "CW min width",
			shape:       shapeByMode[client.ModeCW],
			value:       0,
			expectedMin: -25,
			expectedMax: 25,
		},
		{
			desc:        "CW max width",
			shape:       shapeByMode[client.ModeCW],
			value:       127,
			expectedMin: -500,
			expectedMax: 500,
		},
		{
			desc:        "CW middle width",
			shape:       shapeByMode[client.ModeCW],
			value:       63,
			expectedMin: -260,
			expectedMax: 260,
		},
		{
			desc:        "LSB min width",
			shape:       shapeByMode[client.ModeLSB],
			value:       0,
			expectedMin: -1070,
			expectedMax: -70,
		},
		{
			desc:        "LSB max width",
			shape:       shapeByMode[client.ModeLSB],
			value:       127,
			expectedMin: -3070,
			expectedMax: -70,
		},
		{
			desc:        "LSB middle width",
			shape:       shapeByMode[client.ModeLSB],
			value:       63,
			expectedMin: -2062,
			expectedMax: -70,
		},
		{
			desc:        "DIGU min width",
			shape:       shapeByMode[client.ModeDIGU],
			value:       0,
			expectedMin: 1450,
			expectedMax: 1550,
		},
		{
			desc:        "DIGU max width",
			shape:       shapeByMode[client.ModeDIGU],
			value:       127,
			expectedMin: 0,
			expectedMax: 3000,
		},
		{
			desc:        "DIGU middle width",
			shape:       shapeByMode[client.ModeDIGU],
			value:       63,
			expectedMin: 731,
			expectedMax: 2269,
		},
	}
	for _, tc := range tt {
		t.Run(tc.desc, func(t *testing.T) {
			actualMin, actualMax := tc.shape.Bounds(tc.shape.Width(tc.value))

			assert.Equal(t, tc.expectedMin, actualMin)
			assert.Equal(t, tc.expectedMax, actualMax)
		})
	}
}
