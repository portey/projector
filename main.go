package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"

	"github.com/alecthomas/kingpin/v2"
	"github.com/portey/projector/internal"
	"github.com/portey/projector/internal/git"
	"github.com/portey/projector/internal/output"
	"github.com/portey/projector/internal/sh"
	"github.com/portey/projector/internal/types"
	"gopkg.in/yaml.v3"
)

func main() {
	app := kingpin.New(filepath.Base(os.Args[0]), "A tool to control projects").UsageWriter(os.Stdout)
	app.HelpFlag.Short('h')

	dirname, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	configPath := app.Flag("config-file", "Service configuration file").
		Short('c').
		Default(path.Join(dirname, ".projector.yaml")).
		ExistingFile()

	syncProject := app.Flag("sync", "Sync projects before operation").Default("true").Bool()

	var operator *internal.Operator
	app.PreAction(func(_ *kingpin.ParseContext) error {
		f, err := os.Open(*configPath)
		if err != nil {
			return fmt.Errorf("failed to read config file: %w", err)
		}

		config := types.Config{
			HomeDir: dirname,
		}

		err = yaml.NewDecoder(f).Decode(&config)
		if err != nil {
			return fmt.Errorf("failed to parse Config: %w", err)
		}

		auth, err := config.Auth()
		if err != nil {
			return fmt.Errorf("failed to create public key auth: %w", err)
		}

		progressWriter := io.Discard
		if config.DebugMode {
			progressWriter = os.Stdout
		}

		operator = internal.NewOperator(
			config,
			git.NewGit("main", progressWriter, auth),
			sh.NewShellExecutor(progressWriter),
			output.NewColoredStdOut(),
		)

		return nil
	})

	versionsCMD := app.Command("versions", "Show versions for each of projects").Alias("v")
	versionsCMDProject := versionsCMD.Flag("project", "Project name").Short('p').String()
	versionsCMD.Action(func(_ *kingpin.ParseContext) error {
		if *syncProject {
			if versionsCMDProject != nil && *versionsCMDProject != "" {
				if err := operator.SyncProject(context.Background(), *versionsCMDProject); err != nil {
					return err
				}
			} else {
				if err := operator.SyncAllProjects(context.Background()); err != nil {
					return err
				}
			}
		}

		return operator.ListVersions(context.Background(), versionsCMDProject)
	})

	deployCMD := app.Command("deploy", "Deploy a project version to specific env").Alias("d")
	deployCMDProject := deployCMD.Arg("project", "Project name").Required().String()
	deployCMDVersion := deployCMD.Arg("version", "Project version").Required().String()
	deployCMDEnv := deployCMD.Arg("env", "Environment").Required().Enum(types.DeployTargetEnvs...)
	deployCMD.Action(func(_ *kingpin.ParseContext) error {
		version, err := types.VersionFromString(*deployCMDVersion)
		if err != nil {
			return err
		}

		env, err := types.EnvFromString(*deployCMDEnv)
		if err != nil {
			return err
		}

		if *syncProject {
			if err = operator.SyncProject(context.Background(), *deployCMDProject); err != nil {
				return err
			}
		}

		return operator.Deploy(context.Background(), *deployCMDProject, version, env)
	})

	deployLatestCMD := app.Command("deploy-latest", "Deploy a latest project version to specific env").Alias("dl")
	deployLatestCMDProject := deployLatestCMD.Arg("project", "Project name").Required().String()
	deployLatestCMDEnv := deployLatestCMD.Arg("env", "Environment").Required().Enum(types.DeployTargetEnvs...)
	deployLatestCMD.Action(func(_ *kingpin.ParseContext) error {
		env, err := types.EnvFromString(*deployLatestCMDEnv)
		if err != nil {
			return err
		}

		if *syncProject {
			if err = operator.SyncProject(context.Background(), *deployLatestCMDProject); err != nil {
				return err
			}
		}

		return operator.DeployLatest(context.Background(), *deployLatestCMDProject, env)
	})

	_, err = app.Parse(os.Args[1:])
	if err != nil {
		os.Exit(1)
	}
}
