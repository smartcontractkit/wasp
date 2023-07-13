package wasp

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSmokeSchedules(t *testing.T) {
	type test struct {
		name   string
		input  []*Segment
		output []*Segment
	}

	tests := []test{
		{
			name:  "increasing line",
			input: Line(1, 100, 1*time.Second),
			output: []*Segment{
				{
					From:         1,
					Increase:     10,
					Steps:        DefaultStepChangePrecision,
					StepDuration: 100 * time.Millisecond,
				},
			},
		},
		{
			name:  "decreasing line",
			input: Line(10, 0, 1*time.Second),
			output: []*Segment{
				{
					From:         10,
					Increase:     -1,
					Steps:        DefaultStepChangePrecision,
					StepDuration: 100 * time.Millisecond,
				},
			},
		},
		{
			name: "combine lines",
			input: Combine(
				Line(1, 100, 1*time.Second),
				Plain(200, 1*time.Second),
				Line(100, 1, 1*time.Second),
			),
			output: []*Segment{
				{
					From:         1,
					Increase:     10,
					Steps:        DefaultStepChangePrecision,
					StepDuration: 100 * time.Millisecond,
				},
				{
					From:         200,
					Steps:        1,
					StepDuration: 1 * time.Second,
				},
				{
					From:         100,
					Increase:     -10,
					Steps:        DefaultStepChangePrecision,
					StepDuration: 100 * time.Millisecond,
				},
			},
		},
		{
			name: "combine disjointed lines",
			input: Combine(
				Line(1, 100, 1*time.Second),
				Line(1, 300, 1*time.Second),
			),
			output: []*Segment{
				{
					From:         1,
					Increase:     10,
					Steps:        DefaultStepChangePrecision,
					StepDuration: 100 * time.Millisecond,
				},
				{
					From:         1,
					Increase:     30,
					Steps:        DefaultStepChangePrecision,
					StepDuration: 100 * time.Millisecond,
				},
			},
		},
		{
			name: "combine and repeat",
			input: CombineAndRepeat(
				2,
				Line(1, 100, 1*time.Second),
				Line(100, 1, 1*time.Second),
			),
			output: []*Segment{
				{
					From:         1,
					Increase:     10,
					Steps:        DefaultStepChangePrecision,
					StepDuration: 100 * time.Millisecond,
				},
				{
					From:         100,
					Increase:     -10,
					Steps:        DefaultStepChangePrecision,
					StepDuration: 100 * time.Millisecond,
				},
				{
					From:         1,
					Increase:     10,
					Steps:        DefaultStepChangePrecision,
					StepDuration: 100 * time.Millisecond,
				},
				{
					From:         100,
					Increase:     -10,
					Steps:        DefaultStepChangePrecision,
					StepDuration: 100 * time.Millisecond,
				},
			},
		},
		{
			name:  "line is below default precision (min interval)",
			input: Line(1, 2, 2*time.Second),
			output: []*Segment{
				{
					From:         1,
					Steps:        1,
					StepDuration: 1 * time.Second,
				},
				{
					From:         2,
					Steps:        1,
					StepDuration: 1 * time.Second,
				},
			},
		},
		{
			name:  "line is below default precision (negative)",
			input: Line(10, 8, 3*time.Second),
			output: []*Segment{
				{
					From:         10,
					Steps:        1,
					StepDuration: 1 * time.Second,
				},
				{
					From:         9,
					Steps:        1,
					StepDuration: 1 * time.Second,
				},
				{
					From:         8,
					Steps:        1,
					StepDuration: 1 * time.Second,
				},
			},
		},
		{
			name:  "line is below default precision (no rounding on duration)",
			input: Line(1, 5, 15*time.Second),
			output: []*Segment{
				{
					From:         1,
					Steps:        1,
					StepDuration: 3 * time.Second,
				},
				{
					From:         2,
					Steps:        1,
					StepDuration: 3 * time.Second,
				},
				{
					From:         3,
					Steps:        1,
					StepDuration: 3 * time.Second,
				},
				{
					From:         4,
					Steps:        1,
					StepDuration: 3 * time.Second,
				},
				{
					From:         5,
					Steps:        1,
					StepDuration: 3 * time.Second,
				},
			},
		},
		{
			name:  "line is below default precision (max interval)",
			input: Line(1, 9, 9*time.Millisecond),
			output: []*Segment{
				{
					From:         1,
					Steps:        1,
					StepDuration: 1 * time.Millisecond,
				},
				{
					From:         2,
					Steps:        1,
					StepDuration: 1 * time.Millisecond,
				},
				{
					From:         3,
					Steps:        1,
					StepDuration: 1 * time.Millisecond,
				},
				{
					From:         4,
					Steps:        1,
					StepDuration: 1 * time.Millisecond,
				},
				{
					From:         5,
					Steps:        1,
					StepDuration: 1 * time.Millisecond,
				},
				{
					From:         6,
					Steps:        1,
					StepDuration: 1 * time.Millisecond,
				},
				{
					From:         7,
					Steps:        1,
					StepDuration: 1 * time.Millisecond,
				},
				{
					From:         8,
					Steps:        1,
					StepDuration: 1 * time.Millisecond,
				},
				{
					From:         9,
					Steps:        1,
					StepDuration: 1 * time.Millisecond,
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.input, tc.output)
		})
	}
}
