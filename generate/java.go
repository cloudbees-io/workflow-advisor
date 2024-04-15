package generate

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/calculi-corp/dsl-engine-cli/pkg/dsl"
	"github.com/calculi-corp/workflow-advisor/pkg/utils"
)

type java struct {
	jobName string
}

type javaBuildStep struct {
	files []string
	steps []dsl.Step
}

func init() {
	registerGenerator("java", &java{
		jobName: "java-build",
	})
}

const (
	defaultMavenImage  = "maven:3.9-eclipse-temurin-21-alpine"
	defaultGradleImage = "gradle:8.6-jdk21-alpine"
)

// dsl steps based on repo contents
var javaBuildSteps = []javaBuildStep{
	{
		files: []string{"mvnw"},
		steps: []dsl.Step{
			{
				Name: "mvn install",
				Uses: "docker://" + defaultMavenImage,
				Run:  "./mvnw install",
			},
		},
	},
	{
		files: []string{"pom.xml"},
		steps: []dsl.Step{
			{
				Name: "mvn install",
				Uses: "docker://" + defaultMavenImage,
				Run:  "mvn install",
			},
		},
	},
	{
		files: []string{"gradlew"},
		steps: []dsl.Step{
			{
				Name: "gradle build",
				Uses: "docker://" + defaultGradleImage,
				Run:  "./gradlew build",
			},
			{
				Name: "gradle test",
				Uses: "docker://" + defaultGradleImage,
				Run:  "./gradlew test",
			},
		},
	},
	{
		files: []string{"build.gradle", "build.gradle.kts"},
		steps: []dsl.Step{
			{
				Name: "gradle build",
				Uses: "docker://" + defaultGradleImage,
				Run:  "gradle build",
			},
			{
				Name: "gradle test",
				Uses: "docker://" + defaultGradleImage,
				Run:  "gradle test",
			},
		},
	},
}

func (j *java) Generate(ctx context.Context, workflowContext *WorkflowContext) error {
	dslSteps, err := j.dslSteps(workflowContext.SrcDir)
	if err != nil {
		return err
	}

	if len(dslSteps) == 0 {
		return nil
	}

	return j.addJobIfNotExists(workflowContext.Workflow, dslSteps)
}

func (j *java) addJobIfNotExists(workflow *dsl.Workflow, steps []dsl.Step) error {
	if workflow.Jobs == nil {
		workflow.Jobs = make(map[string]dsl.Job)
	}

	_, ok := workflow.Jobs[j.jobName]

	if ok {
		return fmt.Errorf("error adding job %s, it already exists, java generator is not compatible with workflow provided", j.jobName)
	}

	wfSteps := []dsl.Step{
		{
			Name: "checkout",
			Uses: "cloudbees-io/checkout@v1",
		},
	}
	wfSteps = append(wfSteps, steps...)
	wfSteps = append(wfSteps, dsl.Step{
		Name: "scan",
		Uses: "cloudbees-io/sonarqube-bundled-sast-scan-code@v2",
		With: map[string]string{
			"language": "JAVA",
		},
	})
	workflow.Jobs[j.jobName] = dsl.Job{
		Steps: wfSteps,
	}

	return nil
}
func (j *java) dslSteps(srcDir string) ([]dsl.Step, error) {
	var dslSteps []dsl.Step
	var err error
	var exists bool
	for _, javaBuildStep := range javaBuildSteps {
		for _, f := range javaBuildStep.files {
			if exists, err = utils.Stat(filepath.Join(srcDir, f)); exists && err == nil {
				dslSteps = append(dslSteps, javaBuildStep.steps...)
				return dslSteps, nil
			}
		}
	}

	return dslSteps, err
}
