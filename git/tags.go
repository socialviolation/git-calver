package git

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/socialviolation/git-calver/ver"
)

type Tag struct {
	Short    string
	Ref      string
	Hash     string
	IsBranch bool
	Commit   Commit
}

type Commit struct {
	Message string
	Author  string
}

func List(format *ver.Format) ([]Tag, error) {
	r, err := git.PlainOpen(".")
	if err != nil {
		return nil, fmt.Errorf("could not init repo at .: %w", err)
	}
	refs, err := r.Tags()
	if err != nil {
		return nil, fmt.Errorf("could not find tags: %w", err)
	}
	tags := make([]Tag, 0)
	err = refs.ForEach(func(ref *plumbing.Reference) error {
		if ref.Name().IsTag() {
			co, _ := r.CommitObject(ref.Hash())

			tags = append(tags, Tag{
				Short:    ref.Name().Short(),
				Ref:      ref.String(),
				Hash:     ref.Hash().String(),
				IsBranch: ref.Name().IsBranch(),
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

	return tags, nil
}
