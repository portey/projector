package internal

import (
	"context"
	"errors"
	"fmt"

	"github.com/portey/projector/internal/git"
	"github.com/portey/projector/internal/sh"
	"github.com/portey/projector/internal/types"
	"golang.org/x/sync/errgroup"
)

type outputWriter interface {
	Success(projectName, message string)
	Error(projectName, message string)
	PrintProjectStates(states []types.DeployState)
}

type Operator struct {
	config types.Config
	git    *git.Git
	sh     *sh.ShellExecutor
	output outputWriter
}

func NewOperator(
	config types.Config,
	git *git.Git,
	sh *sh.ShellExecutor,
	output outputWriter,
) *Operator {
	return &Operator{
		config: config,
		git:    git,
		output: output,
		sh:     sh,
	}
}

func (o *Operator) SyncAllProjects(ctx context.Context) error {
	g, gctx := errgroup.WithContext(ctx)

	for _, project := range o.config.Projects() {
		p := project
		g.Go(func() error {
			err := o.SyncProject(gctx, p.Name)
			if errors.Is(err, context.Canceled) {
				return nil
			}

			return err
		})
	}

	return g.Wait()
}

func (o *Operator) SyncProject(ctx context.Context, projectName string) (err error) {
	defer func() {
		if err != nil {
			o.output.Error(projectName, fmt.Sprintf("synced with error: %s", err.Error()))
		} else {
			o.output.Success(projectName, "successfully synced!")
		}
	}()

	project, err := o.config.Project(projectName)
	if err != nil {
		return err
	}

	return o.git.Sync(ctx, project)
}

func (o *Operator) ListVersions(_ context.Context, projectName *string) (err error) {
	projects := o.config.Projects()
	if projectName != nil && *projectName != "" {
		p, err := o.config.Project(*projectName)
		if err != nil {
			o.output.Error(*projectName, fmt.Sprintf("failed to list versions: %s", err.Error()))
			return err
		}

		projects = []types.Project{p}
	}

	states := make([]types.DeployState, 0, len(projects))
	for _, project := range projects {
		tags, err := o.git.Tags(context.Background(), project)
		if err != nil {
			o.output.Error(project.Name, fmt.Sprintf("failed to list versions: %s", err.Error()))
			return err
		}

		states = append(states, types.DeployState{
			Project: project,
			Tags: map[types.Env]types.Tag{
				types.EnvDEV: tags.LatestInEnv(types.EnvDEV),
				types.EnvQA:  tags.LatestInEnv(types.EnvQA),
				types.EnvUAT: tags.LatestInEnv(types.EnvUAT),
			},
		})
	}

	o.output.PrintProjectStates(states)

	return nil
}

func (o *Operator) Deploy(ctx context.Context, projectName string, version types.Version, env types.Env) (err error) {
	defer func() {
		if err != nil {
			o.output.Error(projectName, fmt.Sprintf("failed to deploy version %s to %s: %s", version.Tag(), env.String(), err.Error()))
		} else {
			o.output.Success(projectName, fmt.Sprintf("successfully deployed version %s to %s", version.Tag(), env.String()))
		}
	}()

	if env == types.EnvDEV {
		return nil
	}
	if !version.IsEnv(types.EnvDEV) {
		return fmt.Errorf("only dev version tag can be deployed")
	}

	return o.doDeploy(ctx, projectName, version, env)
}

func (o *Operator) DeployLatest(ctx context.Context, projectName string, env types.Env) (err error) {
	defer func() {
		if err != nil {
			o.output.Error(projectName, fmt.Sprintf("failed to deploy the latest version to %s: %s", env.String(), err.Error()))
		}
	}()

	if env == types.EnvDEV {
		return nil
	}

	project, err := o.config.Project(projectName)
	if err != nil {
		return err
	}

	tags, err := o.git.Tags(ctx, project)
	if err != nil {
		return err
	}

	tag := tags.LatestInEnv(types.EnvDEV)
	if tag.IsNil() {
		return fmt.Errorf("looks like the repo doesn't have dev tags yet")
	}

	if err = o.doDeploy(ctx, projectName, tag.Version, env); err != nil {
		return err
	}

	o.output.Success(projectName, fmt.Sprintf("successfully deployed version %s to %s", tag.Version.Tag(), env.String()))

	return nil
}

func (o *Operator) doDeploy(ctx context.Context, projectName string, version types.Version, env types.Env) error {
	project, err := o.config.Project(projectName)
	if err != nil {
		return err
	}

	tags, err := o.git.Tags(ctx, project)
	if err != nil {
		return err
	}

	targetVersion := version.ToEnv(env)
	for _, tag := range tags {
		if tag.Version.Equal(targetVersion) {
			return fmt.Errorf("version %s already exists", targetVersion.Tag())
		}
	}

	return o.sh.Deploy(ctx, project, version, env)
}
