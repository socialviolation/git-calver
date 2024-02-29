package git

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/socialviolation/git-calver/ver"
)

type TagRef struct {
	Short    string
	Ref      string
	Hash     string
	IsBranch bool
}

func List(format *ver.Format) ([]TagRef, error) {
	r, err := git.PlainOpen(".")
	if err != nil {
		return nil, fmt.Errorf("could not init repo at .: %w", err)
	}
	refs, err := r.Tags()
	if err != nil {
		return nil, fmt.Errorf("could not find tags: %w", err)
	}
	tags := make([]TagRef, 0)
	err = refs.ForEach(func(ref *plumbing.Reference) error {
		if ref.Name().IsTag() {
			tags = append(tags, TagRef{
				Short:    ref.Name().Short(),
				Ref:      ref.String(),
				Hash:     ref.Hash().String(),
				IsBranch: ref.Name().IsBranch(),
			})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return tags, nil
}
