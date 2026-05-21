package application

import (
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
)

func TestParseAnalysisPeriod(t *testing.T) {
	tests := []struct {
		name     string
		raw      string
		expected AnalysisPeriod
	}{
		{name: "month by default", raw: "", expected: AnalysisPeriodMonth},
		{name: "unknown as month", raw: "week", expected: AnalysisPeriodMonth},
		{name: "quarter", raw: "quarter", expected: AnalysisPeriodQuarter},
		{name: "year", raw: "year", expected: AnalysisPeriodYear},
		{name: "trim upper", raw: " YEAR ", expected: AnalysisPeriodYear},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseAnalysisPeriod(tt.raw)
			if got != tt.expected {
				t.Fatalf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestBuildAnalysisRangeWithCustomStartDate(t *testing.T) {
	location := time.UTC
	start := time.Date(2026, time.May, 13, 0, 0, 0, 0, location)
	now := time.Date(2026, time.May, 20, 12, 0, 0, 0, location)

	tests := []struct {
		name        string
		period      AnalysisPeriod
		expectedEnd time.Time
	}{
		{
			name:        "month from custom day",
			period:      AnalysisPeriodMonth,
			expectedEnd: time.Date(2026, time.June, 12, 23, 59, 59, int(time.Second-time.Nanosecond), location),
		},
		{
			name:        "quarter from custom day",
			period:      AnalysisPeriodQuarter,
			expectedEnd: time.Date(2026, time.August, 12, 23, 59, 59, int(time.Second-time.Nanosecond), location),
		},
		{
			name:        "year from custom day",
			period:      AnalysisPeriodYear,
			expectedEnd: time.Date(2027, time.May, 12, 23, 59, 59, int(time.Second-time.Nanosecond), location),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildAnalysisRange(tt.period, &start, now)
			if !got.Start.Equal(start) {
				t.Fatalf("expected start %s, got %s", start, got.Start)
			}
			if !got.End.Equal(tt.expectedEnd) {
				t.Fatalf("expected end %s, got %s", tt.expectedEnd, got.End)
			}
		})
	}
}

func TestIsBudgetOverlappingRange(t *testing.T) {
	location := time.UTC
	rangeStart := time.Date(2026, time.May, 1, 0, 0, 0, 0, location)
	rangeEnd := time.Date(2026, time.May, 31, 23, 59, 59, int(time.Second-time.Nanosecond), location)

	tests := []struct {
		name     string
		budget   models.BudgetModel
		expected bool
	}{
		{
			name: "fully inside",
			budget: models.BudgetModel{
				StartAt: time.Date(2026, time.May, 10, 0, 0, 0, 0, location),
				EndAt:   time.Date(2026, time.May, 20, 0, 0, 0, 0, location),
			},
			expected: true,
		},
		{
			name: "overlaps from left",
			budget: models.BudgetModel{
				StartAt: time.Date(2026, time.April, 20, 0, 0, 0, 0, location),
				EndAt:   time.Date(2026, time.May, 12, 0, 0, 0, 0, location),
			},
			expected: true,
		},
		{
			name: "overlaps from right",
			budget: models.BudgetModel{
				StartAt: time.Date(2026, time.May, 20, 0, 0, 0, 0, location),
				EndAt:   time.Date(2026, time.June, 12, 0, 0, 0, 0, location),
			},
			expected: true,
		},
		{
			name: "covers whole range",
			budget: models.BudgetModel{
				StartAt: time.Date(2026, time.April, 20, 0, 0, 0, 0, location),
				EndAt:   time.Date(2026, time.June, 12, 0, 0, 0, 0, location),
			},
			expected: true,
		},
		{
			name: "before range",
			budget: models.BudgetModel{
				StartAt: time.Date(2026, time.April, 1, 0, 0, 0, 0, location),
				EndAt:   time.Date(2026, time.April, 30, 0, 0, 0, 0, location),
			},
			expected: false,
		},
		{
			name: "after range",
			budget: models.BudgetModel{
				StartAt: time.Date(2026, time.June, 1, 0, 0, 0, 0, location),
				EndAt:   time.Date(2026, time.June, 30, 0, 0, 0, 0, location),
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsBudgetOverlappingRange(tt.budget, rangeStart, rangeEnd)
			if got != tt.expected {
				t.Fatalf("expected %v, got %v", tt.expected, got)
			}
		})
	}
}

func TestGetBudgetRangeIntersection(t *testing.T) {
	location := time.UTC
	rangeStart := time.Date(2026, time.May, 1, 0, 0, 0, 0, location)
	rangeEnd := time.Date(2026, time.May, 31, 23, 59, 59, int(time.Second-time.Nanosecond), location)
	budget := models.BudgetModel{
		StartAt: time.Date(2026, time.May, 12, 0, 0, 0, 0, location),
		EndAt:   time.Date(2026, time.June, 23, 0, 0, 0, 0, location),
	}

	start, end := GetBudgetRangeIntersection(budget, rangeStart, rangeEnd)
	if !start.Equal(budget.StartAt) {
		t.Fatalf("expected intersection start %s, got %s", budget.StartAt, start)
	}
	if !end.Equal(rangeEnd) {
		t.Fatalf("expected intersection end %s, got %s", rangeEnd, end)
	}
}
