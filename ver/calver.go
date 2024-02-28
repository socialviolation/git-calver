package ver

import (
	"fmt"
	"strings"
	"time"
)

type CalVer struct {
	Format   *Format
	Minor    int16
	Micro    int16
	Modifier string
	microSet bool
	minorSet bool
}

type Format struct {
	Major segment
	Minor segment
	Micro segment
}

func (f *Format) String() string {
	if f.Minor == segmentEmpty {
		return fmt.Sprintf("%s", f.Major)
	}
	if f.Micro == segmentEmpty {
		return fmt.Sprintf("%s.%s", f.Major, f.Minor)
	}
	return fmt.Sprintf("%s.%s.%s", f.Major, f.Minor, f.Micro)
}

func (f *Format) Version(t time.Time) string {
	bits := make([]string, 0, 1)
	if f.Major != segmentEmpty {
		bits = append(bits, f.Major.conv(t))
	}
	if f.Minor != segmentEmpty {
		bits = append(bits, f.Minor.conv(t))
	}
	if f.Micro != segmentEmpty {
		bits = append(bits, f.Micro.conv(t))
	}

	return strings.Join(bits, ".")
}

func (f *Format) NeedsMinor() bool {
	return f.Minor != segmentMinor
}

func (f *Format) NeedsMicro() bool {
	return f.Micro != segmentMicro
}

const (
	// FullYear notation for CalVerOld - 2006, 2016, 2106
	FullYear = "YYYY"
	// ShortYear notation for CalVerOld - 6, 16, 106
	ShortYear = "YY"
	// PaddedYear notation for CalVerOld - 06, 16, 106
	PaddedYear = "0Y"
	// ShortMonth notation for CalVerOld - 1, 2 ... 11, 12
	ShortMonth = "MM"
	// PaddedMonth notation for CalVerOld - 01, 02 ... 11, 12
	PaddedMonth = "0M"
	// ShortWeek notation for CalVerOld - 1, 2, 33, 52
	ShortWeek = "WW"
	// PaddedWeek notation for CalVerOld - 01, 02, 33, 52
	PaddedWeek = "0W"
	// ShortDay notation for CalVerOld - 1, 2 ... 30, 31
	ShortDay = "DD"
	// PaddedDay notation for CalVerOld - 01, 02 ... 30, 31
	PaddedDay = "0D"

	Minor = "MINOR"
	Micro = "MICRO"
)

var ValidSegments = [11]string{
	FullYear,
	ShortYear,
	PaddedYear,
	ShortMonth,
	PaddedMonth,
	ShortWeek,
	PaddedWeek,
	ShortDay,
	PaddedDay,
	Minor,
	Micro,
}

type segment int

const (
	segmentEmpty segment = iota
	segmentFullYear
	segmentShortYear
	segmentPaddedYear
	segmentShortMonth
	segmentPaddedMonth
	segmentShortWeek
	segmentPaddedWeek
	segmentShortDay
	segmentPaddedDay
	segmentMinor
	segmentMicro
)

func (s segment) String() string {
	switch s {
	case segmentFullYear:
		return FullYear
	case segmentShortYear:
		return ShortYear
	case segmentPaddedYear:
		return PaddedYear
	case segmentShortMonth:
		return ShortMonth
	case segmentPaddedMonth:
		return PaddedMonth
	case segmentShortWeek:
		return ShortWeek
	case segmentPaddedWeek:
		return PaddedWeek
	case segmentShortDay:
		return ShortDay
	case segmentPaddedDay:
		return PaddedDay
	case segmentMinor:
		return Minor
	case segmentMicro:
		return Micro
	case segmentEmpty:
		return ""
	default:
		panic("invalid format segment")
	}
}

func fmtToSegment(format string) (segment, error) {
	switch format {
	case FullYear:
		return segmentFullYear, nil
	case ShortYear:
		return segmentShortYear, nil
	case PaddedYear:
		return segmentPaddedYear, nil
	case ShortMonth:
		return segmentShortMonth, nil
	case PaddedMonth:
		return segmentPaddedMonth, nil
	case ShortWeek:
		return segmentShortWeek, nil
	case PaddedWeek:
		return segmentPaddedWeek, nil
	case ShortDay:
		return segmentShortDay, nil
	case PaddedDay:
		return segmentPaddedDay, nil
	case Minor:
		return segmentMinor, nil
	case Micro:
		return segmentMicro, nil
	default:
		return segmentEmpty, fmt.Errorf("invalid format segment: %s", format)
	}
}

func (s segment) pattern() string {
	switch s {
	case segmentFullYear:
		return "2006"
	case segmentPaddedYear:
		return "06"
	case segmentShortMonth:
		return "1"
	case segmentPaddedMonth:
		return "01"
	case segmentShortDay:
		return "2"
	case segmentPaddedDay:
		return "02"
	default:
		panic("unsupported format segment")
	}
}

func (s segment) conv(t time.Time) string {
	switch s {
	case segmentEmpty:
		return ""
	case segmentShortWeek:
		_, w := t.ISOWeek()
		return fmt.Sprintf("%d", w)
	case segmentPaddedWeek:
		_, w := t.ISOWeek()
		return fmt.Sprintf("%02d", w)
	case segmentShortYear:
		y := t.Format("06")
		if strings.HasPrefix(y, "0") {
			return strings.TrimPrefix(y, "0")
		}
		return y
	default:
		return t.Format(s.pattern())
	}
}

func NewFormat(raw string) (*Format, error) {
	segs := strings.Split(raw, ".")
	if len(segs) < 2 {
		return nil, fmt.Errorf("requires min 2 segments in format: %s", raw)
	}

	major, err := fmtToSegment(segs[0])
	if err != nil {
		return nil, err
	}

	minor, err := fmtToSegment(segs[1])
	if err != nil {
		return nil, err
	}

	var micro segment
	if len(segs) > 2 {
		micro, err = fmtToSegment(segs[2])
		if err != nil {
			return nil, err
		}
	}

	return &Format{Major: major, Minor: minor, Micro: micro}, nil
}

type CalVerArgs struct {
	Format   string
	Minor    *int16
	Micro    *int16
	Modifier string
}

func (c *CalVerArgs) String() string {
	return fmt.Sprintf("%s.%d.%d-%s", c.Format, c.Minor, c.Micro, c.Modifier)
}

func NewCalVer(a CalVerArgs) (*CalVer, error) {
	f, err := NewFormat(a.Format)
	if err != nil {
		return nil, err
	}

	c := &CalVer{
		Format:   f,
		Modifier: a.Modifier,
	}

	if a.Micro != nil {
		c.Micro = *a.Micro
		c.minorSet = true
	}
	if a.Minor != nil {
		c.Minor = *a.Minor
		c.microSet = true
	}

	return c, nil
}

func (c *CalVer) Version(t time.Time) (string, error) {
	ver := c.Format.Version(t)
	if c.Format.NeedsMinor() {
		if !c.minorSet {
			return "", fmt.Errorf("minor version required for format: %s", c.Format.String())
		}
		ver = fmt.Sprintf("%s.%d", ver, c.Minor)
	}

	if c.Format.NeedsMicro() {
		if !c.microSet {
			return "", fmt.Errorf("micro version required for format: %s", c.Format.String())
		}
		ver = fmt.Sprintf("%s.%d", ver, c.Micro)
	}

	if c.Modifier != "" {
		ver = fmt.Sprintf("%s-%s", ver, c.Modifier)
	}

	return ver, nil
}
