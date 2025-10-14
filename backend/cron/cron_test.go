package cron

import (
	"backend/util"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func init() {
	util.SetDebug(true)
}

func TestCronJobDay(t *testing.T) {
	// Clear any existing jobs
	jobs = nil

	now := util.TimeNow()
	var executed int

	testJob := func(now time.Time) error {
		executed++
		return nil
	}

	CronJobDay(testJob, now)

	// Check if a job is scheduled
	if len(jobs) != 1 {
		t.Fatalf("Expected 1 job to be scheduled, found %d", len(jobs))
	}

	// Verify the job is scheduled to run at the correct time
	if !jobs[0].nextExecAt.Equal(now) {
		t.Errorf("Job is scheduled to run at %v, expected %v", jobs[0].nextExecAt, now)
	}

	// Advance time by 24 hours to trigger the job
	util.AddTimeNowSeconds(24 * 60 * 60)
	CheckNow()

	// Assuming there's a way to process or trigger the cron check here,
	// like a public method or by advancing the internal ticker

	// Verify the job was executed twice
	assert.Equal(t, 2, executed, "executed twice")

	// Reset the time manipulation after the test
	util.ResetTimeNow() // Assuming ResetTimeNow resets the time manipulation in your util package

}

func TestCronJobHour(t *testing.T) {
	// Clear any existing jobs
	jobs = nil

	// Set the current time to a known value
	now := util.TimeNow()

	var executed int
	testJob := func(now time.Time) error {
		executed++
		return nil
	}

	CronJobHour(testJob, now)

	// Verify job is scheduled
	if len(jobs) != 1 {
		t.Fatalf("Expected 1 job to be scheduled, found %d", len(jobs))
	}

	// Verify the job is scheduled to run at the correct time
	if !jobs[0].nextExecAt.Equal(now) {
		t.Errorf("Job is scheduled to run at %v, expected %v", jobs[0].nextExecAt, now)
	}

	// Advance time by 1 hour to trigger the job
	util.AddTimeNowSeconds(3600)
	CheckNow()

	// Assuming there's a way to process or trigger the cron check here

	// Verify the job was executed
	assert.Equal(t, 2, executed, "executed twice")

	// Reset the time manipulation after the test
	util.ResetTimeNow() // Reset the mocked time
}

func TestTimeDayPlusOne(t *testing.T) {
	// Test cases for different days and times
	testCases := []struct {
		name     string
		input    time.Time
		expected time.Time
	}{
		{
			name:     "Midnight",
			input:    time.Date(2024, 1, 14, 0, 0, 0, 0, time.UTC),
			expected: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "Midday",
			input:    time.Date(2024, 1, 14, 12, 0, 0, 0, time.UTC),
			expected: time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
		},
		{
			name:     "One minute to midnight",
			input:    time.Date(2024, 1, 14, 23, 59, 0, 0, time.UTC),
			expected: time.Date(2024, 1, 15, 23, 59, 0, 0, time.UTC),
		},
		{
			name:     "Leap Year",
			input:    time.Date(2024, 2, 28, 15, 0, 0, 0, time.UTC),
			expected: time.Date(2024, 2, 29, 15, 0, 0, 0, time.UTC),
		},
		{
			name:     "End of Year",
			input:    time.Date(2024, 12, 31, 10, 0, 0, 0, time.UTC),
			expected: time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := timeDayPlusOne(tc.input)
			if !actual.Equal(tc.expected) {
				t.Errorf("timeDayPlusOne(%v) = %v; want %v", tc.input, actual, tc.expected)
			}
		})
	}
}

func TestTimeHourPlusOne(t *testing.T) {
	// Test cases for different times of the day
	testCases := []struct {
		name     string
		input    time.Time
		expected time.Time
	}{
		{
			name:     "Midnight",
			input:    time.Date(2024, 1, 14, 0, 0, 0, 0, time.UTC),
			expected: time.Date(2024, 1, 14, 1, 0, 0, 0, time.UTC),
		},
		{
			name:     "Midday",
			input:    time.Date(2024, 1, 14, 12, 0, 0, 0, time.UTC),
			expected: time.Date(2024, 1, 14, 13, 0, 0, 0, time.UTC),
		},
		{
			name:     "One hour before midnight",
			input:    time.Date(2024, 1, 14, 23, 0, 0, 0, time.UTC),
			expected: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := timeHourPlusOne(tc.input)
			if !actual.Equal(tc.expected) {
				t.Errorf("timeHourPlusOne(%v) = %v; want %v", tc.input, actual, tc.expected)
			}
		})
	}
}

func TestTwoHourlyJobs(t *testing.T) {
	// Clear existing jobs
	jobs = nil

	// Set the mock time
	now := util.TimeNow()

	// Variables to track job executions
	var executedJob1, executedJob2 int

	// First hourly job
	testJob1 := func(now time.Time) error {
		executedJob1++
		return nil
	}

	// Second hourly job
	testJob2 := func(now time.Time) error {
		executedJob2++
		return nil
	}

	// Schedule both jobs
	CronJobHour(testJob1, now)
	CronJobHour(testJob2, now)

	// Verify both jobs are scheduled
	if len(jobs) != 2 {
		t.Fatalf("Expected 2 jobs to be scheduled, found %d", len(jobs))
	}

	// Advance time by 1 hour to trigger the jobs
	util.AddTimeNowSeconds(2 * 60 * 60)
	CheckNow()

	// Trigger the cron job execution (this step depends on your cron implementation)

	// Verify both jobs were executed
	assert.Equal(t, 3, executedJob1, "needs to be executed twice")
	assert.Equal(t, 3, executedJob2, "needs to be executed twice")

	// Reset the time after the test
	util.ResetTimeNow()
}
