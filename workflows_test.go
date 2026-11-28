package main

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"go.yaml.in/yaml/v3"
)

// workflowFile is the part of a GitHub Actions workflow that matters for the token permissions.
type workflowFile struct {
	Permissions any `yaml:"permissions"`
	Jobs        map[string]struct {
		Permissions any `yaml:"permissions"`
	} `yaml:"jobs"`
}

// jobsWithDefaultToken returns the jobs of a workflow that would run with the repository's default token:
// no top-level permissions and no job-level permissions.
func jobsWithDefaultToken(t *testing.T, content []byte) []string {
	t.Helper()

	var wf workflowFile
	if err := yaml.Unmarshal(content, &wf); err != nil {
		t.Fatalf("Couldn't parse workflow: %v", err)
	}

	if wf.Permissions != nil {
		return nil
	}

	var jobs []string

	for name, job := range wf.Jobs {
		if job.Permissions == nil {
			jobs = append(jobs, name)
		}
	}

	slices.Sort(jobs)

	return jobs
}

// TestWorkflowsDeclarePermissions makes sure no workflow silently gets the repository's default (write) token.
func TestWorkflowsDeclarePermissions(t *testing.T) {
	files, err := filepath.Glob(filepath.Join(".github", "workflows", "*.yml"))
	if err != nil {
		t.Fatal(err)
	}

	if len(files) == 0 {
		t.Fatal("No workflow found")
	}

	for _, file := range files {
		content, errRead := os.ReadFile(file)
		if errRead != nil {
			t.Fatal(errRead)
		}

		if jobs := jobsWithDefaultToken(t, content); len(jobs) > 0 {
			t.Errorf("%s: no top-level permissions and none on jobs %v", file, jobs)
		}
	}
}

func TestJobsWithDefaultToken(t *testing.T) {
	cases := map[string]struct {
		content string
		jobs    []string
	}{
		"top-level": {
			content: "permissions:\n  contents: read\njobs:\n  a:\n    runs-on: x\n",
		},
		"empty top-level": {
			content: "permissions: {}\njobs:\n  a:\n    runs-on: x\n",
		},
		"job-level only": {
			content: "jobs:\n  a:\n    permissions:\n      contents: write\n  b:\n    permissions: read-all\n",
		},
		"missing": {
			content: "jobs:\n  b:\n    runs-on: x\n  a:\n    permissions:\n      contents: read\n  c:\n    runs-on: x\n",
			jobs:    []string{"b", "c"},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if got := jobsWithDefaultToken(t, []byte(tc.content)); !slices.Equal(got, tc.jobs) {
				t.Errorf("got %v, want %v", got, tc.jobs)
			}
		})
	}
}
