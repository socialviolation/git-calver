package ver

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type CalVer struct {
	Format        *Format
	Minor         uint
	Micro         uint
	AutoIncrement bool
	Increment     uint
	Modifier      string
	microSet      bool
	minorSet      bool
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
	if f.Minor != segmentEmpty && f.Minor != segmentMinor {
		bits = append(bits, f.Minor.conv(t))
	}
	if f.Micro != segmentEmpty && f.Micro != segmentMicro {
		bits = append(bits, f.Micro.conv(t))
	}

	return strings.Join(bits, ".")
}

func (c *CalVer) Regex() *regexp.Regexp {
	mod := ""
	if c.AutoIncrement {
		mod = `-\d+`
	} else if c.Modifier != "" {
		mod = c.Modifier
	}

	if c.Format.Minor == segmentEmpty {
		r, _ := regexp.Compile(fmt.Sprintf(`^%s(-(\w+)){0,1}%s$`, c.Format.Major.Regex(), mod))
		return r
	}
	if c.Format.Micro == segmentEmpty {
		r, _ := regexp.Compile(fmt.Sprintf(`^%s\.%s(-\w+){0,1}%s$`, c.Format.Major.Regex(), c.Format.Minor.Regex(), mod))
		return r
	}
	r, _ := regexp.Compile(fmt.Sprintf(`^%s\.%s\.%s(-\w+){0,1}%s$`, c.Format.Major.Regex(), c.Format.Minor.Regex(), c.Format.Micro.Regex(), mod))
	return r
}

func (f *Format) NeedsMinor() bool {
	return f.Minor == segmentMinor
}

func (f *Format) NeedsMicro() bool {
	return f.Micro == segmentMicro
}

const (
	// FullYear notation - 2006, 2016, 2106
	FullYear = "YYYY"
	// ShortYear notation - 6, 16, 106
	ShortYear = "YY"
	// PaddedYear notation - 06, 16, 106
	PaddedYear = "0Y"
	// ShortMonth notation - 1, 2 ... 11, 12
	ShortMonth = "MM"
	// PaddedMonth notation - 01, 02 ... 11, 12
	PaddedMonth = "0M"
	// ShortWeek notation - 1, 2, 33, 52
	ShortWeek = "WW"
	// PaddedWeek notation - 01, 02, 33, 52
	PaddedWeek = "0W"
	// ShortDay notation - 1, 2 ... 30, 31
	ShortDay = "DD"
	// PaddedDay notation - 01, 02 ... 30, 31
	PaddedDay = "0D"

	Minor = "MINOR"
	Micro = "MICRO"
	Auto  = "AUTO"
)

var ValidSegments = [12]string{
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
	Auto,
}

type segment uint

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
	segmentAuto
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
	case segmentAuto:
		return Auto
	case segmentEmpty:
		return ""
	default:
		panic("invalid format segment")
	}
}

func (s segment) Regex() string {
	switch s {
	case segmentFullYear:
		return "20[0-9]{2}"
	case segmentShortYear:
		return "[0-9]{2}"
	case segmentPaddedYear:
		return "[0-9]{1,2}"
	case segmentShortMonth:
		return "[0-9]{1,2}"
	case segmentPaddedMonth:
		return "[0-9]{2}"
	case segmentShortWeek:
		return "[0-9]{1,2}"
	case segmentPaddedWeek:
		return "[0-9]{2}"
	case segmentShortDay:
		return "[0-9]{1,2}"
	case segmentPaddedDay:
		return "[0-9]{2}"
	case segmentMinor:
		return Minor
	case segmentMicro:
		return Micro
	case segmentAuto:
		return "\b(AUTO)|\\d+)\b"
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
	case Auto:
		return segmentAuto, nil
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
	case segmentMinor:
		return ""
	case segmentMicro:
		return ""
	case segmentAuto:
		return ""
	default:
		return t.Format(s.pattern())
	}
}

func NewFormat(raw string) (*Format, error) {
	calBit := strings.Split(raw, "-")
	if len(calBit) < 1 {
		return nil, fmt.Errorf("requires min 2 segments in format: %s", raw)
	}

	segs := strings.Split(calBit[0], ".")
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
	Format        *Format
	RawFormat     string
	Minor         *uint
	Micro         *uint
	Modifier      string
	DryRun        bool
	AutoIncrement bool
	Hash          string
}

func (c *CalVerArgs) String() string {
	return fmt.Sprintf("%s-%s", c.RawFormat, c.Modifier)
}

func NewCalVer(a CalVerArgs) (*CalVer, error) {
	c := &CalVer{
		Format:        a.Format,
		Modifier:      a.Modifier,
		AutoIncrement: a.AutoIncrement,
	}

	if c.Format == nil {
		cf, err := NewFormat(a.RawFormat)
		if err != nil {
			return nil, err
		}
		c.Format = cf
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

func NextCalVer(a CalVerArgs) (*CalVer, error) {
	c, err := NewCalVer(a)
	if err != nil {
		return nil, err
	}
	if c.AutoIncrement {
		nextInc, err := GetLatestAutoInc(c)
		if err != nil {
			return nil, fmt.Errorf("could not find next increment: %s\n", err.Error())
		}
		c.Modifier = c.Modifier + strconv.Itoa(nextInc)
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
