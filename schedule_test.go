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
					From:                  1,
					Increase:              10,
					Steps:                 DefaultStepChangePrecision,
					StepDuration:          100 * time.Millisecond,
					RateLimitUnitDuration: 1 * time.Second,
				},
			},
		},
		{
			name:  "decreasing line",
			input: Line(10, 0, 1*time.Second),
			output: []*Segment{
				{
					From:                  10,
					Increase:              -1,
					Steps:                 DefaultStepChangePrecision,
					StepDuration:          100 * time.Millisecond,
					RateLimitUnitDuration: 1 * time.Second,
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
					From:                  1,
					Increase:              10,
					Steps:                 DefaultStepChangePrecision,
					StepDuration:          100 * time.Millisecond,
					RateLimitUnitDuration: 1 * time.Second,
				},
				{
					From:                  200,
					Steps:                 1,
					StepDuration:          1 * time.Second,
					RateLimitUnitDuration: 1 * time.Second,
				},
				{
					From:                  100,
					Increase:              -10,
					Steps:                 DefaultStepChangePrecision,
					StepDuration:          100 * time.Millisecond,
					RateLimitUnitDuration: 1 * time.Second,
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
					From:                  1,
					Increase:              10,
					Steps:                 DefaultStepChangePrecision,
					StepDuration:          100 * time.Millisecond,
					RateLimitUnitDuration: 1 * time.Second,
				},
				{
					From:                  1,
					Increase:              30,
					Steps:                 DefaultStepChangePrecision,
					StepDuration:          100 * time.Millisecond,
					RateLimitUnitDuration: 1 * time.Second,
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
					From:                  1,
					Increase:              10,
					Steps:                 DefaultStepChangePrecision,
					StepDuration:          100 * time.Millisecond,
					RateLimitUnitDuration: 1 * time.Second,
				},
				{
					From:                  100,
					Increase:              -10,
					Steps:                 DefaultStepChangePrecision,
					StepDuration:          100 * time.Millisecond,
					RateLimitUnitDuration: 1 * time.Second,
				},
				{
					From:                  1,
					Increase:              10,
					Steps:                 DefaultStepChangePrecision,
					StepDuration:          100 * time.Millisecond,
					RateLimitUnitDuration: 1 * time.Second,
				},
				{
					From:                  100,
					Increase:              -10,
					Steps:                 DefaultStepChangePrecision,
					StepDuration:          100 * time.Millisecond,
					RateLimitUnitDuration: 1 * time.Second,
				},
			},
		},
		{
			name: "combine and repeat with diff rate limits",
			input: CombineAndRepeat(
				2,
				PlainWithCustomRateLimit(1, 10, 1*time.Second),
				PlainWithCustomRateLimit(2, 15, 5*time.Second),
			),
			output: []*Segment{
				{
					From:                  1,
					Steps:                 1,
					StepDuration:          10,
					RateLimitUnitDuration: 1 * time.Second,
				},
				{
					From:                  2,
					Steps:                 1,
					StepDuration:          15,
					RateLimitUnitDuration: 5 * time.Second,
				},
				{
					From:                  1,
					Steps:                 1,
					StepDuration:          10,
					RateLimitUnitDuration: 1 * time.Second,
				},
				{
					From:                  2,
					Steps:                 1,
					StepDuration:          15,
					RateLimitUnitDuration: 5 * time.Second,
				},
			},
		},
		{
			name:  "line is below default precision (min interval)",
			input: Line(1, 2, 2*time.Second),
			output: []*Segment{
				{
					From:                  1,
					Steps:                 1,
					StepDuration:          1 * time.Second,
					RateLimitUnitDuration: 1 * time.Second,
				},
				{
					From:                  2,
					Steps:                 1,
					StepDuration:          1 * time.Second,
					RateLimitUnitDuration: 1 * time.Second,
				},
			},
		},
		{
			name:  "line is below default precision (negative)",
			input: Line(10, 8, 3*time.Second),
			output: []*Segment{
				{
					From:                  10,
					Steps:                 1,
					StepDuration:          1 * time.Second,
					RateLimitUnitDuration: 1 * time.Second,
				},
				{
					From:                  9,
					Steps:                 1,
					StepDuration:          1 * time.Second,
					RateLimitUnitDuration: 1 * time.Second,
				},
				{
					From:                  8,
					Steps:                 1,
					StepDuration:          1 * time.Second,
					RateLimitUnitDuration: 1 * time.Second,
				},
			},
		},
		{
			name:  "line is below default precision (no rounding on duration)",
			input: Line(1, 5, 15*time.Second),
			output: []*Segment{
				{
					From:                  1,
					Steps:                 1,
					StepDuration:          3 * time.Second,
					RateLimitUnitDuration: 1 * time.Second,
				},
				{
					From:                  2,
					Steps:                 1,
					StepDuration:          3 * time.Second,
					RateLimitUnitDuration: 1 * time.Second,
				},
				{
					From:                  3,
					Steps:                 1,
					StepDuration:          3 * time.Second,
					RateLimitUnitDuration: 1 * time.Second,
				},
				{
					From:                  4,
					Steps:                 1,
					StepDuration:          3 * time.Second,
					RateLimitUnitDuration: 1 * time.Second,
				},
				{
					From:                  5,
					Steps:                 1,
					StepDuration:          3 * time.Second,
					RateLimitUnitDuration: 1 * time.Second,
				},
			},
		},
		{
			name:  "line is below default precision (max interval)",
			input: Line(1, 9, 9*time.Millisecond),
			output: []*Segment{
				{
					From:                  1,
					Steps:                 1,
					StepDuration:          1 * time.Millisecond,
					RateLimitUnitDuration: 1 * time.Second,
				},
				{
					From:                  2,
					Steps:                 1,
					StepDuration:          1 * time.Millisecond,
					RateLimitUnitDuration: 1 * time.Second,
				},
				{
					From:                  3,
					Steps:                 1,
					StepDuration:          1 * time.Millisecond,
					RateLimitUnitDuration: 1 * time.Second,
				},
				{
					From:                  4,
					Steps:                 1,
					StepDuration:          1 * time.Millisecond,
					RateLimitUnitDuration: 1 * time.Second,
				},
				{
					From:                  5,
					Steps:                 1,
					StepDuration:          1 * time.Millisecond,
					RateLimitUnitDuration: 1 * time.Second,
				},
				{
					From:                  6,
					Steps:                 1,
					StepDuration:          1 * time.Millisecond,
					RateLimitUnitDuration: 1 * time.Second,
				},
				{
					From:                  7,
					Steps:                 1,
					StepDuration:          1 * time.Millisecond,
					RateLimitUnitDuration: 1 * time.Second,
				},
				{
					From:                  8,
					Steps:                 1,
					StepDuration:          1 * time.Millisecond,
					RateLimitUnitDuration: 1 * time.Second,
				},
				{
					From:                  9,
					Steps:                 1,
					StepDuration:          1 * time.Millisecond,
					RateLimitUnitDuration: 1 * time.Second,
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			for i := range tc.output {
				require.NoError(t, tc.input[i].Validate())
			}
			require.Equal(t, tc.input, tc.output)
		})
	}
}
