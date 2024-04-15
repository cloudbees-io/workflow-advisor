package generate

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/calculi-corp/dsl-engine-cli/pkg/dsl"
	"github.com/calculi-corp/workflow-advisor/pkg/utils"
)

const (
	requirementsTxt    = "requirements.txt"
	setupPy            = "setup.py"
	defaultPythonImage = "docker://python:3.13.0a4-alpine3.19"
)

type python struct {
	jobName string
}

func init() {
	registerGenerator("python", &python{
		jobName: "python-build",
	})
}

func (p *python) Generate(ctx context.Context, wc *WorkflowContext) error {
	isPython, err := p.containsSource(wc.SrcDir)
	if err != nil {
		return err
	}
	if !isPython {
		return nil
	}
	return p.addJobIfNotExists(wc.Workflow, wc.SrcDir)
}

func (p *python) containsSource(srcDir string) (bool, error) {
	containsPySource := func(path string) (bool, error) {
		name := filepath.Base(path)
		return strings.HasSuffix(name, ".py"), nil
	}
	hasPySource, err := findInDir(srcDir, containsPySource)
	if err != nil {
		return false, err
	}
	return hasPySource, nil
}

func (p *python) addJobIfNotExists(wf *dsl.Workflow, srcDir string) error {
	if wf.Jobs == nil {
		wf.Jobs = make(map[string]dsl.Job)
	}

	if _, ok := wf.Jobs[p.jobName]; ok {
		return fmt.Errorf("error adding job: job %s already exists", p.jobName)
	}

	steps, err := buildSteps(srcDir)
	if err != nil {
		return err
	}

	wf.Jobs[p.jobName] = dsl.Job{
		Steps: steps,
	}
	return nil
}

func buildSteps(srcDir string) ([]dsl.Step, error) {
	steps := []dsl.Step{
		{
			Name: "checkout",
			Uses: "cloudbees-io/checkout@v1",
		},
	}

	type assembleFunc func(srcDir string) (*dsl.Step, error)

	for _, assemble := range []assembleFunc{installStep, buildStep, testStep, scanStep} {
		step, err := assemble(srcDir)
		if err != nil {
			return steps, err
		}
		if step != nil {
			steps = append(steps, *step)
		}
	}

	return steps, nil
}

func scanStep(_ string) (*dsl.Step, error) {
	return &dsl.Step{
		Name: "scan",
		Uses: "cloudbees-io/sonarqube-bundled-sast-scan-code@v2",
		With: map[string]string{
			"language": "LANGUAGE_PYTHON",
		},
	}, nil
}

func installStep(srcDir string) (*dsl.Step, error) {
	hasRequirements, err := utils.Stat(filepath.Join(srcDir, requirementsTxt))
	if err != nil {
		return nil, err
	}
	if hasRequirements {
		return &dsl.Step{
			Name: "install packages",
			Uses: defaultPythonImage,
			Run: `python -m pip install --upgrade pip
pip install -r requirements.txt`,
		}, nil
	}
	return nil, nil
}

func buildStep(srcDir string) (*dsl.Step, error) {
	hasSetup, err := utils.Stat(filepath.Join(srcDir, setupPy))
	if err != nil {
		return nil, err
	}
	if hasSetup {
		return &dsl.Step{
			Name: "build",
			Uses: defaultPythonImage,
			Run: `python -m pip install build
python -m build --sdist
python -m build --wheel`,
		}, nil
	}
	return nil, nil
}

func testStep(srcDir string) (*dsl.Step, error) {
	requiresPytest, err := fileContains(filepath.Join(srcDir, requirementsTxt), func(t string) bool {
		return strings.Contains(t, "pytest==") && !strings.HasPrefix("#", t)
	})
	if err != nil {
		return nil, err
	}

	installPytest := `python -m pip install pytest
`
	if requiresPytest {
		installPytest = ""
	}

	hasTests, err := hasTests(srcDir)
	if err != nil {
		return nil, err
	}

	if requiresPytest || hasTests {
		return &dsl.Step{
			Name: "test",
			Uses: defaultPythonImage,
			Run:  installPytest + "python -m unittest",
		}, nil
	}
	return nil, nil
}

func hasTests(srcDir string) (bool, error) {
	importsPytest := func(text string) bool {
		return !strings.HasPrefix("#", text) && strings.Contains("import pytest", text)
	}
	isTest := func(path string) (bool, error) {
		name := filepath.Base(path)
		if strings.HasSuffix(name, ".py") && strings.Contains(name, "test") {
			ok, err := fileContains(path, importsPytest)
			if err != nil {
				return false, err
			}
			if ok {
				return true, nil
			}
		}
		return false, nil
	}

	if found, err := findInDir(filepath.Join(srcDir, "tests"), isTest); err != nil {
		return false, err
	} else {
		return found, nil
	}
}

// findInDir walks a file directory until the provided filter returns true, if filter does not match returns false
func findInDir(dir string, filter func(path string) (bool, error)) (bool, error) {
	found := false
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.Mode().IsRegular() {
			if ok, err := filter(path); err != nil {
				return err
			} else if ok {
				found = true
				return filepath.SkipAll
			}
		}
		return nil
	})

	if os.IsNotExist(err) {
		return false, nil
	}
	return found, err
}

// fileContains will read a file until the provided filter returns true, if filter does not match returns false
func fileContains(file string, filter func(text string) bool) (bool, error) {
	f, err := os.Open(file)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	for s.Scan() {
		if filter(s.Text()) {
			return true, nil
		}
	}
	return false, nil
}
