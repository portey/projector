package git

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/protocol/packp/sideband"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/portey/projector/internal/types"
)

type Git struct {
	mainBranch string
	progress   sideband.Progress
	auth       transport.AuthMethod
}

func NewGit(
	mainBranch string,
	progress sideband.Progress,
	auth transport.AuthMethod,
) *Git {
	return &Git{
		mainBranch: mainBranch,
		auth:       auth,
		progress:   progress,
	}
}

func (g *Git) Sync(ctx context.Context, project types.Project) error {
	repo, tree, err := g.openRepo(project, true)
	if err != nil {
		return err
	}

	err = repo.FetchContext(ctx, &git.FetchOptions{
		Tags:     git.AllTags,
		Auth:     g.auth,
		Progress: g.progress,
	})
	if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		return fmt.Errorf("failed fetch project %q: %w", project.Name, err)
	}

	err = tree.PullContext(ctx, &git.PullOptions{
		ReferenceName: plumbing.NewBranchReferenceName(g.mainBranch),
		Auth:          g.auth,
		Progress:      g.progress,
	})
	if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		return fmt.Errorf("failed pull project %q: %w", project.Name, err)
	}

	return nil
}

func (g *Git) Tags(_ context.Context, project types.Project) (types.Tags, error) {
	repo, _, err := g.openRepo(project, false)
	if err != nil {
		return nil, err
	}

	return g.listVersionTags(project, repo)
}

func (g *Git) openRepo(project types.Project, withCheckout bool) (*git.Repository, *git.Worktree, error) {
	repo, err := git.PlainOpen(project.Path)
	if err != nil {
		return nil, nil, fmt.Errorf("failed load project repo %q: %w", project.Name, err)
	}

	tree, err := repo.Worktree()
	if err != nil {
		return nil, nil, fmt.Errorf("failed load project repo tree %q: %w", project.Name, err)
	}

	if !withCheckout {
		return repo, tree, nil
	}

	err = tree.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(g.mainBranch),
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed checkout project %q to main branch: %w", project.Name, err)
	}

	return repo, tree, nil
}

func (g *Git) listVersionTags(project types.Project, repo *git.Repository) (types.Tags, error) {
	tags, err := repo.Tags()
	if err != nil {
		return nil, fmt.Errorf("failed get tags for project %q: %w", project.Name, err)
	}

	res := make([]types.Tag, 0)
	err = tags.ForEach(func(reference *plumbing.Reference) error {
		if !reference.Name().IsTag() {
			return nil
		}

		version, err := types.VersionFromString(reference.Name().Short())
		if err == nil {
			res = append(res, types.Tag{
				Hash:    reference.Hash(),
				Name:    reference.Name().Short(),
				Version: version,
			})
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed itterate tags for project %q: %w", project.Name, err)
	}

	return res, nil
}
