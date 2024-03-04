package ver

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"sort"
	"time"
)

func GetRepoFormat() (*Format, error) {
	r, err := git.PlainOpen(".")
	if err != nil {
		return nil, fmt.Errorf("could not init repo at .: %w", err)
	}

	conf, err := r.Config()
	if err != nil {
		return nil, fmt.Errorf("could not retrieve config: %w", err)
	}
	if err != nil {
		return nil, err
	}

	if !conf.Raw.HasSection("calver") {
		return nil, fmt.Errorf("[calver] not set")
	}

	val := conf.Raw.Section("calver").Option("format")
	if val == "" {
		return nil, fmt.Errorf("[calver].format not set")
	}

	f, err := NewFormat(val)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func SetRepoFormat(f *Format) error {
	r, err := git.PlainOpen(".")
	if err != nil {
		return fmt.Errorf("could not init repo at .: %w", err)
	}

	conf, err := r.Config()
	if err != nil {
		return fmt.Errorf("could not retrieve config: %w", err)
	}
	if err != nil {
		return err
	}

	conf.Raw.SetOption("calver", "", "format", f.String())
	err = r.SetConfig(conf)
	if err != nil {
		return err
	}

	return nil
}

func ListTags(format *Format) ([]Tag, error) {
	r, err := git.PlainOpen(".")
	if err != nil {
		return nil, fmt.Errorf("could not init repo at .: %w", err)
	}

	refs, err := r.Tags()
	if err != nil {
		return nil, fmt.Errorf("could not find tags: %w", err)
	}
	tags := make([]Tag, 0)
	regex := format.Regex()
	err = refs.ForEach(func(ref *plumbing.Reference) error {
		if ref.Name().IsTag() {
			short := ref.Name().Short()
			if !regex.Match([]byte(short)) {
				return nil
			}

			co, _ := r.CommitObject(ref.Hash())
			li, _ := r.Log(&git.LogOptions{From: ref.Hash()})
			log, _ := li.Next()
			tags = append(tags, Tag{
				Short:       ref.Name().Short(),
				Ref:         ref.String(),
				Hash:        ref.Hash().String(),
				IsBranch:    ref.Name().IsBranch(),
				Time:        co.Committer.When,
				FullMessage: log.String(),
				Commit: Commit{
					Message: co.Message,
					Author:  co.Author.String(),
				},
			})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(tags, func(i, j int) bool {
		return tags[j].Time.Before(tags[i].Time)
	})
	return tags, nil
}

type Tag struct {
	Short       string
	Ref         string
	Hash        string
	IsBranch    bool
	Commit      Commit
	Time        time.Time
	FullMessage string
}

type Commit struct {
	Message string
	Author  string
}
