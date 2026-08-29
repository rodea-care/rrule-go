// 2017-2022, Teambition. All rights reserved.

package rrule

import (
	"testing"
	"time"
)

func TestBasicRecurrencePatterns(t *testing.T) {
	tests := []struct {
		name   string
		option ROption
		want   []time.Time
	}{
		{
			name: "daily",
			option: ROption{Freq: DAILY, Count: 3,
				Dtstart: time.Date(2026, 1, 5, 9, 0, 0, 0, time.UTC)},
			want: []time.Time{
				time.Date(2026, 1, 5, 9, 0, 0, 0, time.UTC),
				time.Date(2026, 1, 6, 9, 0, 0, 0, time.UTC),
				time.Date(2026, 1, 7, 9, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "daily interval two",
			option: ROption{Freq: DAILY, Interval: 2, Count: 3,
				Dtstart: time.Date(2026, 1, 5, 9, 0, 0, 0, time.UTC)},
			want: []time.Time{
				time.Date(2026, 1, 5, 9, 0, 0, 0, time.UTC),
				time.Date(2026, 1, 7, 9, 0, 0, 0, time.UTC),
				time.Date(2026, 1, 9, 9, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "weekly",
			option: ROption{Freq: WEEKLY, Count: 3,
				Dtstart: time.Date(2026, 1, 5, 9, 0, 0, 0, time.UTC)},
			want: []time.Time{
				time.Date(2026, 1, 5, 9, 0, 0, 0, time.UTC),
				time.Date(2026, 1, 12, 9, 0, 0, 0, time.UTC),
				time.Date(2026, 1, 19, 9, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "weekly Monday Wednesday Friday",
			option: ROption{Freq: WEEKLY, Count: 6, Byweekday: []Weekday{MO, WE, FR},
				Dtstart: time.Date(2026, 1, 5, 9, 0, 0, 0, time.UTC)},
			want: []time.Time{
				time.Date(2026, 1, 5, 9, 0, 0, 0, time.UTC),
				time.Date(2026, 1, 7, 9, 0, 0, 0, time.UTC),
				time.Date(2026, 1, 9, 9, 0, 0, 0, time.UTC),
				time.Date(2026, 1, 12, 9, 0, 0, 0, time.UTC),
				time.Date(2026, 1, 14, 9, 0, 0, 0, time.UTC),
				time.Date(2026, 1, 16, 9, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "until inclusive",
			option: ROption{Freq: DAILY,
				Dtstart: time.Date(2026, 1, 5, 9, 0, 0, 0, time.UTC),
				Until:   time.Date(2026, 1, 7, 9, 0, 0, 0, time.UTC)},
			want: []time.Time{
				time.Date(2026, 1, 5, 9, 0, 0, 0, time.UTC),
				time.Date(2026, 1, 6, 9, 0, 0, 0, time.UTC),
				time.Date(2026, 1, 7, 9, 0, 0, 0, time.UTC),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rule, err := NewRRule(test.option)
			if err != nil {
				t.Fatal(err)
			}
			if got := rule.All(); !timesEqual(got, test.want) {
				t.Errorf("got %v, want %v", got, test.want)
			}
		})
	}
}

func TestDailyIntervalsRemainAtLocalTimeAcrossTimezoneTransitions(t *testing.T) {
	tests := []struct {
		name             string
		zone             string
		start            [3]int
		interval         int
		wantOffsetChange bool
	}{
		{name: "Bogota daily control", zone: "America/Bogota", start: [3]int{2024, 3, 9}, interval: 1},
		{name: "Bogota interval control", zone: "America/Bogota", start: [3]int{2024, 11, 1}, interval: 2},
		{name: "New York DST start daily", zone: "America/New_York", start: [3]int{2024, 3, 9}, interval: 1, wantOffsetChange: true},
		{name: "New York DST end interval", zone: "America/New_York", start: [3]int{2024, 11, 1}, interval: 2, wantOffsetChange: true},
		{name: "London DST start interval", zone: "Europe/London", start: [3]int{2024, 3, 29}, interval: 2, wantOffsetChange: true},
		{name: "London DST end daily", zone: "Europe/London", start: [3]int{2024, 10, 26}, interval: 1, wantOffsetChange: true},
		{name: "Sydney DST end interval", zone: "Australia/Sydney", start: [3]int{2024, 4, 5}, interval: 2, wantOffsetChange: true},
		{name: "Sydney DST start daily", zone: "Australia/Sydney", start: [3]int{2024, 10, 5}, interval: 1, wantOffsetChange: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			location := mustLoadLocation(t, test.zone)
			start := time.Date(test.start[0], time.Month(test.start[1]), test.start[2], 9, 0, 0, 0, location)
			rule, err := NewRRule(ROption{
				Freq: DAILY, Interval: test.interval, Count: 4, Dtstart: start,
			})
			if err != nil {
				t.Fatal(err)
			}

			got := rule.All()
			if len(got) != 4 {
				t.Fatalf("got %d occurrences, want 4", len(got))
			}
			for i, occurrence := range got {
				want := start.AddDate(0, 0, i*test.interval)
				if !occurrence.Equal(want) || occurrence.Location().String() != test.zone {
					t.Errorf("occurrence %d: got %v, want %v", i, occurrence, want)
				}
				if occurrence.Hour() != 9 {
					t.Errorf("occurrence %d moved from 09:00 local time: %v", i, occurrence)
				}
			}

			_, firstOffset := got[0].Zone()
			_, lastOffset := got[len(got)-1].Zone()
			if changed := firstOffset != lastOffset; changed != test.wantOffsetChange {
				t.Errorf("timezone offset changed=%v, want %v (%d -> %d)", changed, test.wantOffsetChange, firstOffset, lastOffset)
			}
		})
	}
}
