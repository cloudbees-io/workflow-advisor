package generate

import (
	"context"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/calculi-corp/dsl-engine-cli/pkg/dsl"
)

var (
	csharpExtension    = ".csproj"
	solutionExtension  = ".sln"
	solutionFileHeader = "Microsoft Visual Studio Solution File"
	defaultVersion     = "net8.0"
	supportedVersions  = map[string]string{
		"net8.0": "docker://mcr.microsoft.com/dotnet/sdk:8.0",
		"net7.0": "docker://mcr.microsoft.com/dotnet/sdk:7.0",
		"net6.0": "docker://mcr.microsoft.com/dotnet/sdk:6.0",
		"net5.0": "docker://mcr.microsoft.com/dotnet/sdk:5.0",
	}
)

type csharp struct {
	jobName string
}

type CsProj struct {
	XMLName        xml.Name        `xml:"Project"`
	Sdk            string          `xml:"Sdk,attr"`
	PropertyGroups []PropertyGroup `xml:"PropertyGroup"`
}

type PropertyGroup struct {
	TargetFramework string `xml:"TargetFramework"`
}

type csharpContext struct {
	isCSharpRepo bool
	Version      string
	solutions    []string
}

func init() {
	registerGenerator("csharp", &csharp{
		jobName: "csharp-build",
	})
}

func (g *csharp) Generate(ctx context.Context, workflowContext *WorkflowContext) error {
	csContext, err := g.detectTech(workflowContext.SrcDir)
	if err != nil {
		return err
	}

	if !csContext.isCSharpRepo {
		return nil
	}

	err = g.generateJob(workflowContext.Workflow, csContext)

	return err
}

func (g *csharp) generateJob(workflow *dsl.Workflow, csContext csharpContext) error {
	if workflow.Jobs == nil {
		workflow.Jobs = make(map[string]dsl.Job)
	}

	_, ok := workflow.Jobs[g.jobName]

	if ok {
		return fmt.Errorf("error adding job %s, it already exists, csharp generator is not compatible with workflow provided", g.jobName)
	}

	job := dsl.Job{
		Steps: []dsl.Step{
			{
				Name: "checkout",
				Uses: checkoutAction,
			},
		},
	}

	image := g.getImage(csContext.Version)

	if len(csContext.solutions) > 0 {
		for _, solution := range csContext.solutions {
			solutionName := filepath.Base(solution)

			job.Steps = append(job.Steps, dsl.Step{
				Name: fmt.Sprintf("build %s", solutionName),
				Uses: image,
				Run:  fmt.Sprintf("dotnet build %s", solution),
			})

			job.Steps = append(job.Steps, dsl.Step{
				Name: fmt.Sprintf("Test %s", solutionName),
				Uses: image,
				Run:  fmt.Sprintf("dotnet test %s", solution),
			})
		}
	} else {
		job.Steps = append(job.Steps, dsl.Step{
			Name: "Create solution",
			Uses: image,
			Run: `dotnet new sln -n all-projects
find . -name "*.csproj" -print0 | xargs -0 dotnet sln add`,
		})

		job.Steps = append(job.Steps, dsl.Step{
			Name: "Build",
			Uses: image,
			Run:  "dotnet build ./all-projects.sln",
		})

		job.Steps = append(job.Steps, dsl.Step{
			Name: "Test",
			Uses: image,
			Run:  "dotnet test ./all-projects.sln",
		})
	}

	job.Steps = append(job.Steps, dsl.Step{
		Name: "Scan",
		Uses: "cloudbees-io/sonarqube-bundled-sast-scan-code@v2",
		With: map[string]string{
			"language": "LANGUAGE_DOTNET"},
	})

	workflow.Jobs[g.jobName] = job

	return nil
}

func (g *csharp) getImage(version string) string {
	val, ok := supportedVersions[version]
	if ok {
		return val
	}

	return supportedVersions[defaultVersion]
}

func (g *csharp) detectTech(folder string) (csharpContext, error) {
	var files []string
	res := csharpContext{
		isCSharpRepo: false,
	}

	err := filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(info.Name(), csharpExtension) {
			files = append(files, path)
		}

		return nil
	})

	if err != nil {
		return res, nil
	}

	for _, file := range files {
		iscs, version := g.isSdkProj(file)
		if iscs {
			res.isCSharpRepo = true
			_, ok := supportedVersions[version]
			if ok && version > res.Version {
				res.Version = version
			}
		}
	}

	if len(res.Version) == 0 {
		res.Version = defaultVersion
	}

	solutions, err := g.findSolutions(folder)
	if err != nil {
		return res, nil
	}

	res.solutions = solutions

	return res, nil
}

func (g *csharp) findSolutions(folder string) ([]string, error) {
	var files []string

	err := filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(info.Name(), solutionExtension) {
			projBytes, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			if strings.Contains(string(projBytes), solutionFileHeader) {
				files = append(files, path)
			}
		}

		return nil
	})

	return files, err
}

func (g *csharp) isSdkProj(path string) (bool, string) {
	var proj CsProj
	projBytes, err := os.ReadFile(path)
	if err != nil {
		// ignore project if can not read it
		return false, ""
	}
	if err := xml.Unmarshal(projBytes, &proj); err != nil {
		// ignore project if it has unknown format
		return false, ""
	}
	if len(proj.Sdk) > 0 {
		version := defaultVersion
		for _, prop := range proj.PropertyGroups {
			if len(prop.TargetFramework) > 0 {
				version = prop.TargetFramework
				break
			}
		}
		return true, version
	}

	return false, ""
}
