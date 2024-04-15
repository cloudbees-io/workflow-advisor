package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/calculi-corp/dsl-engine-cli/pkg/dsl"
	"gopkg.in/yaml.v3"
)

const (
	ApiVersionField   = "apiVersion"
	KindField         = "kind"
	CurrentApiVersion = "automation.cloudbees.io/v1alpha1"
	WorkflowKind      = "workflow"
)

func Stat(file string) (bool, error) {
	_, err := os.Stat(file)

	if err != nil && !os.IsNotExist(err) {
		return false, err
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	return true, nil
}

func MarshalWorkflowToFile(path string, workflow *dsl.Workflow) error {
	b, err := MarshalWorkflow(workflow)
	if err != nil {
		return err
	}

	err = os.WriteFile(path, b, 0640)
	return err
}

func MarshalWorkflow(workflow *dsl.Workflow) ([]byte, error) {
	header := fmt.Sprintf("apiVersion: %s\nkind: %s\nname: %s\n", workflow.APIVersion, workflow.Kind, workflow.Name)

	trigger, err := toYAML(workflow.On)
	if err != nil {
		return nil, err
	}
	trigger = "  " + strings.ReplaceAll(trigger, "\n", "\n  ")

	var jobs strings.Builder

	if len(workflow.Jobs) > 0 {
		jobs.WriteString("\njobs:\n")

		keys := make([]string, 0, len(workflow.Jobs))
		for k, job := range workflow.Jobs {
			if len(job.Steps) > 0 {
				keys = append(keys, k)
			}
		}
		sort.Strings(keys)

		for _, k := range keys {
			_, _ = jobs.WriteString(fmt.Sprintf("  %s:\n    steps:\n", k))
			job := workflow.Jobs[k]
			for _, step := range job.Steps {
				stepYAML := ""
				if step.ID != "" {
					stepYAML = fmt.Sprintf("id: %s\n", step.ID)
				}
				if step.Name != "" {
					stepYAML += fmt.Sprintf("name: %s\n", step.Name)
				}
				if step.If != "" {
					ifYAML, err := toYAML(step.If)
					if err != nil {
						return nil, err
					}
					stepYAML += fmt.Sprintf("if: %s\n", strings.TrimSpace(ifYAML))
				}
				stepYAML += fmt.Sprintf("uses: %s\n", step.Uses)
				if step.Run != "" {
					run, err := toYAML(step.Run)
					if err != nil {
						return nil, err
					}
					stepYAML += fmt.Sprintf("run: %s\n", strings.TrimSpace(run))
				}
				if len(step.With) > 0 {
					with, err := toYAML(step.With)
					if err != nil {
						return nil, err
					}
					with = strings.TrimSpace(with)
					with = "  " + strings.ReplaceAll(with, "\n", "\n  ")
					stepYAML += fmt.Sprintf("with:\n%s\n", with)
				}
				stepYAML = strings.TrimSpace(stepYAML)
				stepYAML = strings.ReplaceAll(stepYAML, "\n", "\n        ")
				_, _ = jobs.WriteString(fmt.Sprintf("      - %s\n", stepYAML))
			}
		}
	}

	jobsYAML := jobs.String()
	y := fmt.Sprintf("%s\non:\n%s\n%s", header, trigger, jobsYAML)

	return []byte(y), nil
}

func toYAML(o any) (string, error) {
	b, err := json.Marshal(o)
	if err != nil {
		return "", err
	}

	var data any

	err = json.Unmarshal(b, &data)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	err = enc.Encode(data)
	if err != nil {
		return "", err
	}
	err = enc.Close()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(buf.String()), nil
}

func UnmarshalWorkflow(path string) (*dsl.Workflow, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	m := map[string]interface{}{}
	err = yaml.Unmarshal(raw, m)
	if err != nil {
		return nil, err
	}

	apiVersionStr, ok := m[ApiVersionField].(string)
	if !ok {
		return nil, fmt.Errorf("descriptor does not specify field %q", ApiVersionField)
	}
	if apiVersionStr != CurrentApiVersion {
		return nil, fmt.Errorf("unsupported api version %s, expected %s", apiVersionStr, CurrentApiVersion)
	}

	kindStr, ok := m[KindField].(string)
	if !ok {
		return nil, fmt.Errorf("descriptor does not specify field %q", kindStr)
	}
	if kindStr != WorkflowKind {
		return nil, fmt.Errorf("unsupported api version %s, expected %s", kindStr, WorkflowKind)
	}

	raw, err = json.Marshal(m)
	if err != nil {
		return nil, err
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()

	workflow := &dsl.Workflow{}

	err = d.Decode(workflow)
	if err != nil {
		return nil, err
	}
	return workflow, nil
}
