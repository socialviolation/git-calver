package ver

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNewFormat(t *testing.T) {
	tests := []struct {
		fmt       string
		out       Format
		wantError bool
	}{
		{
			fmt:       "YYYY.MM.DD",
			out:       Format{Major: segmentFullYear, Minor: segmentShortMonth, Micro: segmentShortDay},
			wantError: false,
		},
		{
			fmt:       "YYYY.0M.DD",
			out:       Format{Major: segmentFullYear, Minor: segmentPaddedMonth, Micro: segmentShortDay},
			wantError: false,
		},
		{
			fmt:       "YYYY.0M.0D",
			out:       Format{Major: segmentFullYear, Minor: segmentPaddedMonth, Micro: segmentPaddedDay},
			wantError: false,
		},
		{
			fmt:       "YY.MM.DD",
			out:       Format{Major: segmentShortYear, Minor: segmentShortMonth, Micro: segmentShortDay},
			wantError: false,
		},
		{
			fmt:       "YY.WW",
			out:       Format{Major: segmentShortYear, Minor: segmentShortWeek, Micro: segmentEmpty},
			wantError: false,
		},
		{
			fmt:       "YY.0W",
			out:       Format{Major: segmentShortYear, Minor: segmentPaddedWeek, Micro: segmentEmpty},
			wantError: false,
		},
		{
			fmt:       "YY.MINOR.MICRO",
			out:       Format{Major: segmentShortYear, Minor: segmentMinor, Micro: segmentMicro},
			wantError: false,
		},
		{
			fmt:       "YY.MINOR",
			out:       Format{Major: segmentShortYear, Minor: segmentMinor, Micro: segmentEmpty},
			wantError: false,
		},
		{
			fmt:       "YY",
			out:       Format{Major: segmentShortYear, Minor: segmentEmpty, Micro: segmentEmpty},
			wantError: true,
		},
		{
			fmt:       "WW",
			out:       Format{Major: segmentShortWeek, Minor: segmentEmpty, Micro: segmentEmpty},
			wantError: true,
		},
		{
			fmt:       "YYYY.MM.DD",
			out:       Format{Major: segmentFullYear, Minor: segmentShortMonth, Micro: segmentShortDay},
			wantError: false,
		},
	}

	for _, test := range tests {
		t.Run(test.fmt, func(t *testing.T) {
			out, err := NewFormat(test.fmt)
			if test.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.out.Major, out.Major)
				assert.Equal(t, test.out.Minor, out.Minor)
				assert.Equal(t, test.out.Micro, out.Micro)
			}

		})
	}
}

func TestFormatVersion(t *testing.T) {
	tests := []struct {
		fmt       Format
		timestamp time.Time
		out       string
	}{
		{
			fmt:       Format{Major: segmentFullYear, Minor: segmentShortMonth, Micro: segmentShortDay},
			timestamp: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			out:       "2020.1.1",
		},
		{
			fmt:       Format{Major: segmentShortYear, Minor: segmentShortMonth, Micro: segmentShortDay},
			timestamp: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			out:       "20.1.1",
		},
		{
			fmt:       Format{Major: segmentShortYear, Minor: segmentPaddedMonth, Micro: segmentPaddedDay},
			timestamp: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			out:       "20.01.01",
		},
		{
			fmt:       Format{Major: segmentShortYear, Minor: segmentPaddedMonth, Micro: segmentEmpty},
			timestamp: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			out:       "20.01",
		},
		{
			fmt:       Format{Major: segmentShortYear, Minor: segmentShortMonth, Micro: segmentShortDay},
			timestamp: time.Date(2020, 11, 11, 0, 0, 0, 0, time.UTC),
			out:       "20.11.11",
		},
		{
			fmt:       Format{Major: segmentPaddedYear, Minor: segmentShortMonth, Micro: segmentShortDay},
			timestamp: time.Date(2001, 11, 11, 0, 0, 0, 0, time.UTC),
			out:       "01.11.11",
		},
		{
			fmt:       Format{Major: segmentShortYear, Minor: segmentShortMonth, Micro: segmentShortDay},
			timestamp: time.Date(2001, 11, 11, 0, 0, 0, 0, time.UTC),
			out:       "1.11.11",
		},
		{
			fmt:       Format{Major: segmentShortYear, Minor: segmentShortWeek, Micro: segmentEmpty},
			timestamp: time.Date(2001, 11, 11, 0, 0, 0, 0, time.UTC),
			out:       "1.45",
		},
		{
			fmt:       Format{Major: segmentShortYear, Minor: segmentShortWeek, Micro: segmentEmpty},
			timestamp: time.Date(2001, 1, 11, 0, 0, 0, 0, time.UTC),
			out:       "1.2",
		},
		{
			fmt:       Format{Major: segmentShortYear, Minor: segmentPaddedWeek, Micro: segmentEmpty},
			timestamp: time.Date(2001, 1, 11, 0, 0, 0, 0, time.UTC),
			out:       "1.02",
		},
		{
			fmt:       Format{Major: segmentFullYear, Minor: segmentEmpty, Micro: segmentEmpty},
			timestamp: time.Date(2001, 1, 11, 0, 0, 0, 0, time.UTC),
			out:       "2001",
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%s -> %s", test.fmt.String(), test.out), func(t *testing.T) {
			out := test.fmt.Version(test.timestamp)
			assert.Equal(t, test.out, out)
		})
	}
}

func TestCalVerVersion(t *testing.T) {
	one := int16(1)
	twelve := int16(12)
	random := int16(420)

	tests := []struct {
		args      CalVerArgs
		out       string
		timestamp time.Time
		errMsg    string
	}{
		{
			args: CalVerArgs{
				Format: "YYYY.MM.DD",
			},
			out:       "2020.1.1",
			timestamp: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			args: CalVerArgs{
				Format:   "YY.MM.DD",
				Modifier: "TEST",
			},
			out:       "2020.1.1-TEST",
			timestamp: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			args: CalVerArgs{
				Format:   "YY.MM",
				Modifier: "TEST",
			},
			out:       "2020.1-TEST",
			timestamp: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			args: CalVerArgs{
				Format: "0Y.0M",
			},
			out:       "01.01-TEST",
			timestamp: time.Date(2001, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			args: CalVerArgs{
				Format:   "0Y.0M",
				Modifier: "RC",
			},
			out:       "01.01-TEST",
			timestamp: time.Date(2001, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			args: CalVerArgs{
				Format: "YYYY.MINOR.MICRO",
				Minor:  &one,
				Micro:  &twelve,
			},
			out:       "2024.1.12",
			timestamp: time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			args: CalVerArgs{
				Format: "YYYY.MINOR",
				Minor:  &random,
				Micro:  &twelve,
			},
			out:       "2024.420",
			timestamp: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%s -> %s", test.args.String(), test.out), func(t *testing.T) {
			cv, err := NewCalVer(test.args)
			if err != nil {
				assert.Error(t, err, test.errMsg)
				return
			}

			out, err := cv.Version(test.timestamp)
			if err != nil {
				assert.Error(t, err, test.errMsg)
			} else {
				assert.Equal(t, test.out, out)
			}
		})
	}
}
