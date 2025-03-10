package ver

import (
	"fmt"
	"io"
	"strings"
	"time"

	pretty "github.com/andanhm/go-prettytime"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	colour "github.com/gookit/color"
)

type CalVerTagGroup struct {
	Hash      string
	Tags      []string
	When      time.Time
	Commit    *object.Commit
	Refs      []*plumbing.Reference
	Latest    bool
	LatestTag string
	ChangeLog []*object.Commit
}

func (cvt *CalVerTagGroup) Time() time.Time {
	if cvt.Commit == nil {
		return time.Time{}
	}
	return cvt.Commit.Author.When
}

func (cvt *CalVerTagGroup) printTags() string {
	result := ""
	for i, tag := range cvt.Tags {
		if i == len(cvt.Tags)-1 {
			result += fmt.Sprintf("tag: %s", tag)
			continue
		}
		result += fmt.Sprintf("tag: %s, ", tag)
	}

	return result
}

func (cvt *CalVerTagGroup) Print(w io.Writer, noColour bool, lean bool) {
	if noColour {
		colour.Disable()
	}

	if lean {
		_, _ = w.Write([]byte(cvt.LatestTag + "\n"))
		return
	}

	headline := colour.Yellow.Sprint(cvt.Hash) + " - "
	if cvt.Latest {
		headline += colour.HEX("#af5fff").Sprint(cvt.printTags())
	} else {
		headline += colour.Green.Sprintf(cvt.printTags())
	}

	subtitle := fmt.Sprintf("%s - %s", cvt.Commit.Author.Name, pretty.Format(cvt.Time()))

	changeLog := "CHANGELOG:"

	for i, commit := range cvt.ChangeLog {
		b := colour.Red.Sprint("*")
		when := colour.Gray.Sprint(commit.Author.When.Format("2006-01-02 15:04"))
		hash := colour.Yellow.Sprint(commit.Hash.String()[:7])
		msg := commit.Message
		who := colour.Cyan.Sprintf("[%s]", commit.Author.Name)

		line := fmt.Sprintf("\t%s %s  %s  %s  %s", b, when, hash, msg, who)
		line = strings.Replace(line, "\n", "", -1)
		changeLog += "\n" + line
		if i > 10 {
			changeLog += "\n\t..."
			break
		}
	}

	result := fmt.Sprintf("\n%s\n%s\n%s\n", headline, subtitle, changeLog)
	_, _ = w.Write([]byte(result))
}
