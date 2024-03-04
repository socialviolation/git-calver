package ver

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
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

	refs, err := r.TagObjects()
	if err != nil {
		return nil, fmt.Errorf("could not find tags: %w", err)
	}
	tags := make([]Tag, 0)
	regex := format.Regex()
	ot := []*object.Tag{}
	err = refs.ForEach(func(t *object.Tag) error {
		short := t.Name
		if !regex.Match([]byte(short)) {
			return nil
		}
		ot = append(ot, t)
		return nil
	})

	//err = refs.ForEach(func(ref *plumbing.Reference) error {
	//	if ref.Name().IsTag() {
	//		short := ref.Name().Short()
	//		if !regex.Match([]byte(short)) {
	//			return nil
	//		}
	//
	//		co, _ := r.CommitObject(ref.Hash())
	//		li, _ := r.Log(&git.LogOptions{From: ref.Hash()})
	//		log, _ := li.Next()
	//		tags = append(tags, Tag{
	//			Short:       ref.Name().Short(),
	//			Ref:         ref.String(),
	//			Hash:        ref.Hash().String(),
	//			IsBranch:    ref.Name().IsBranch(),
	//			Time:        co.Committer.When,
	//			FullMessage: log.String(),
	//			Commit: Commit{
	//				Message: co.Message,
	//				Author:  co.Author.String(),
	//			},
	//		})
	//	}
	//	return nil
	//})
	if err != nil {
		return nil, err
	}

	sort.Slice(tags, func(i, j int) bool {
		return tags[j].Time.Before(tags[i].Time)
	})
	return tags, nil
}

func TagNext(ver CalVer, hash string) error {
	r, err := git.PlainOpen(".")
	if err != nil {
		return fmt.Errorf("could not init repo at .: %w", err)
	}

	v, err := ver.Version(time.Now())
	if err != nil {
		return err
	}

	if hash == "" {
		h, err := r.ResolveRevision("HEAD")
		if err != nil {
			return fmt.Errorf("could not resolve HEAD: %w", err)
		}
		hash = h.String()
	} else {
		_, err := r.ResolveRevision(plumbing.Revision(hash))
		if err != nil {
			return fmt.Errorf("could not resolve hash %s: %w", hash, err)
		}
	}
	_, err = r.CreateTag(v, plumbing.NewHash(hash), nil)
	if err != nil {
		return fmt.Errorf("could not create tag: %w", err)
	}
	return nil
}

func Retag(ver CalVer, hash string) error {
	r, err := git.PlainOpen(".")
	if err != nil {
		return fmt.Errorf("could not init repo at .: %w", err)
	}

	v, err := ver.Version(time.Now())
	if err != nil {
		return err
	}

	if hash == "" {
		hash = "HEAD"
	}
	_, err = r.CreateTag(v, plumbing.NewHash(hash), &git.CreateTagOptions{
		//Force: true,
	})
	if err != nil {
		return fmt.Errorf("could not create tag: %w", err)
	}
	return nil
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
