package sh

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"

	"github.com/portey/projector/internal/types"
)

type ShellExecutor struct {
	stdout io.Writer
}

func NewShellExecutor(stdout io.Writer) *ShellExecutor {
	return &ShellExecutor{stdout: stdout}
}

const (
	runFileLocation = "deploy"
	qaDeployFile    = "qa-deploy.sh"
	uatDeployFile   = "uat-deploy.sh"
)

func (s *ShellExecutor) Deploy(ctx context.Context, project types.Project, version types.Version, env types.Env) error {
	var runFile string
	switch env {
	case types.EnvQA:
		runFile = str(project.QADeployRunFile, qaDeployFile)
	case types.EnvUAT:
		runFile = str(project.UATDeployRunFile, uatDeployFile)
	default:
		return fmt.Errorf("env %s not supported to be deployed", env.String())
	}

	script := path.Join(project.Path, str(project.RunFileLocation, runFileLocation), runFile)
	command := version.Tag()

	cmd := exec.CommandContext(ctx, script, command)
	cmd.Stdout = s.stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = path.Join(project.Path, str(project.RunFileLocation, runFileLocation))

	err := cmd.Start()
	if err != nil {
		return err
	}

	err = cmd.Wait()
	if err != nil {
		return err
	}

	return nil
}

func str(v, defaultV string) string {
	if v == "" {
		return defaultV
	}

	return v
}
