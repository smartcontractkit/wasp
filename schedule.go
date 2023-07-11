package wasp

import (
	"math"
	"time"
)

/* Different load profile schedules definitions */

const (
	// DefaultStepChangePrecision is default amount of steps in which we split a schedule
	DefaultStepChangePrecision = 10
)

// Plain create a constant workload Segment
func Plain(from int64, duration time.Duration) []*Segment {
	return []*Segment{
		{
			From:                  from,
			Steps:                 1,
			StepDuration:          duration,
			RateLimitUnitDuration: time.Second,
		},
	}
}

func PlainWithTimeUnit(from int64, duration time.Duration, rl time.Duration) []*Segment {
	return []*Segment{
		{
			From:                  from,
			Steps:                 1,
			StepDuration:          duration,
			RateLimitUnitDuration: rl,
		},
	}
}

// Line creates a series of increasing/decreasing Segments
func Line(from, to int64, duration time.Duration) []*Segment {
	var inc int64
	stepDur := duration / DefaultStepChangePrecision
	stepsRange := float64(to) - float64(from)
	incFloat := stepsRange / DefaultStepChangePrecision
	if math.Signbit(incFloat) {
		inc = int64(math.Floor(incFloat))
	} else {
		inc = int64(math.Ceil(incFloat))
	}
	// if Line range is lower than 1..DefaultStepChangePrecision or DefaultStepChangePrecision..1
	// we populate segments using Plain profile and calculate subStep duration accordingly
	if stepsRange >= 0 && stepsRange < DefaultStepChangePrecision {
		return fillLieBelowPrecisionSegments(true, from, to, stepsRange, duration)
	} else if stepsRange > -DefaultStepChangePrecision && stepsRange <= 0 {
		return fillLieBelowPrecisionSegments(false, from, to, stepsRange, duration)
	}
	return []*Segment{
		{
			From:         from,
			Steps:        DefaultStepChangePrecision,
			Increase:     inc,
			StepDuration: stepDur,
		},
	}
}

// fillLieBelowPrecisionSegments generates profile using Plain if line is below 1..DefaultStepChangePrecision or DefaultStepChangePrecision..1
func fillLieBelowPrecisionSegments(positiveRange bool, from, to int64, stepsRange float64, stepDur time.Duration) []*Segment {
	segs := make([]*Segment, 0)
	// inclusive range
	distance := math.Abs(stepsRange) + 1
	// split Line duration over the range
	subStepDur := stepDur / time.Duration(distance)
	switch positiveRange {
	case true:
		for i := from; i <= to; i++ {
			segs = append(segs, Plain(i, subStepDur)...)
		}
	case false:
		for i := from; i >= to; i-- {
			segs = append(segs, Plain(i, subStepDur)...)
		}
	}
	return segs
}

// Combine combines profile segments
func Combine(segs ...[]*Segment) []*Segment {
	acc := make([]*Segment, 0)
	for _, ss := range segs {
		acc = append(acc, ss...)
	}
	return acc
}

// CombineAndRepeat combines and repeats profile segments
func CombineAndRepeat(times int, segs ...[]*Segment) []*Segment {
	if len(segs) == 0 {
		panic(ErrNoSchedule)
	}
	acc := make([]*Segment, 0)
	for i := 0; i < times; i++ {
		for _, ss := range segs {
			acc = append(acc, ss...)
		}
	}
	return acc
}
