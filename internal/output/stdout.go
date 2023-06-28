package output

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/portey/projector/internal/types"
)

type ColoredStdOut struct {
}

func NewColoredStdOut() *ColoredStdOut {
	return &ColoredStdOut{}
}

func (c *ColoredStdOut) Success(projectName, message string) {
	color.New(color.FgGreen).Println(fmt.Sprintf("> %s: %s", projectName, message))
}

func (c *ColoredStdOut) Error(projectName, message string) {
	color.New(color.FgRed).Println(fmt.Sprintf("> %s: %s", projectName, message))
}

func (c *ColoredStdOut) PrintProjectStates(states []types.DeployState) {
	red := color.New(color.FgRed).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Project", "DEV", "QA", "UAT"})

	for _, state := range states {
		qaColor, uatColor := green, green
		if state.Tags[types.EnvDEV].Version.SemVer().Core().GreaterThan(state.Tags[types.EnvQA].Version.SemVer().Core()) {
			qaColor = red
		}
		if state.Tags[types.EnvQA].Version.SemVer().Core().GreaterThan(state.Tags[types.EnvUAT].Version.SemVer().Core()) {
			uatColor = red
		}

		t.AppendRow(table.Row{
			state.Project.Name,
			state.Tags[types.EnvDEV].Version.Tag(),
			qaColor(state.Tags[types.EnvQA].Version.Tag()),
			uatColor(state.Tags[types.EnvUAT].Version.Tag()),
		})
	}
	t.Render()
}
