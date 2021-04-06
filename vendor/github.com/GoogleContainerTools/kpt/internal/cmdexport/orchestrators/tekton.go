// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package orchestrators

import (
	"bytes"
	"path"

	"sigs.k8s.io/kustomize/kyaml/yaml"

	"github.com/GoogleContainerTools/kpt/internal/cmdexport/types"
)

// TektonPipeline wraps a Pipeline object in Tekton.
// @see https://github.com/tektoncd/pipeline/blob/master/docs/pipelines.md
type TektonPipeline struct {
	APIVersion string              `yaml:"apiVersion,omitempty"`
	Kind       string              `yaml:",omitempty"`
	Metadata   *TektonMetadata     `yaml:",omitempty"`
	Spec       *TektonPipelineSpec `yaml:",omitempty"`
	task       *TektonTask
}

func (p *TektonPipeline) Init(config *types.PipelineConfig) Pipeline {
	taskName := "run-kpt-functions"

	p.APIVersion = "tekton.dev/v1beta1"
	p.Kind = "Pipeline"
	p.Metadata = &TektonMetadata{Name: taskName}

	workspaceName := "shared-workspace"
	pipelineWorkspace := &TektonWorkspace{
		Name: workspaceName,
	}

	taskConfig := &TektonTaskConfig{PipelineConfig: config, Name: taskName}
	task := new(TektonTask).Init(taskConfig)

	pipelineTaskWorkspace := &TektonWorkspace{
		Name:      "source",
		Workspace: workspaceName,
	}
	pipelineTask := &TektonPipelineTask{
		Name:       "kpt",
		TaskRef:    &TektonPipelineTaskRef{Name: taskName},
		Workspaces: []*TektonWorkspace{pipelineTaskWorkspace},
	}

	p.task = task
	p.Spec = &TektonPipelineSpec{
		Workspaces: []*TektonWorkspace{pipelineWorkspace},
		Tasks:      []*TektonPipelineTask{pipelineTask},
	}

	return p
}

// Generate outputs a multi-doc yaml that contains Tekton Pipeline Object and a Tekton Task object.
func (p *TektonPipeline) Generate() (out []byte, err error) {
	task, err := p.task.Generate()

	if err != nil {
		return
	}

	pipeline, err := yaml.Marshal(p)

	if err != nil {
		return
	}

	// Concat the two yaml docs.
	b := &bytes.Buffer{}

	b.Write(task)
	b.WriteString("---\n")
	b.Write(pipeline)

	return b.Bytes(), nil
}

// TektonPipelineSpec describes the spec of a Pipeline.
type TektonPipelineSpec struct {
	Workspaces []*TektonWorkspace    `yaml:",omitempty"`
	Tasks      []*TektonPipelineTask `yaml:",omitempty"`
}

// // TektonPipelineTaskRef represents a task in a Tekton Pipeline object.
type TektonPipelineTask struct {
	Name       string                 `yaml:",omitempty"`
	TaskRef    *TektonPipelineTaskRef `yaml:"taskRef,omitempty"`
	RunAfter   []string               `yaml:"runAfter,omitempty"`
	Workspaces []*TektonWorkspace     `yaml:",omitempty"`
}

// TektonPipelineTaskRef represents a taskRef field in a task of a Tekton Pipeline object.
type TektonPipelineTaskRef struct {
	Name string `yaml:",omitempty"`
}

// TektonMetadata contains metadata to describe a resource object.
type TektonMetadata struct {
	Name string `yaml:",omitempty"`
}

// TektonWorkspace represents a shared workspace.
type TektonWorkspace struct {
	Name      string `yaml:",omitempty"`
	MountPath string `yaml:"mountPath,omitempty"`
	// Workspace references another workspace declared elsewhere.
	Workspace string `yaml:",omitempty"`
}

// TektonTaskConfig contains necessary configurations of the TektonTask class.
type TektonTaskConfig struct {
	*types.PipelineConfig
	// Name specifies the name of the task.
	Name string
}

// TektonTask represents a Task object in Tekton.
// @see https://github.com/tektoncd/pipeline/blob/master/docs/tasks.md
type TektonTask struct {
	APIVersion string          `yaml:"apiVersion,omitempty"`
	Kind       string          `yaml:",omitempty"`
	Metadata   *TektonMetadata `yaml:",omitempty"`
	Spec       *TektonTaskSpec `yaml:",omitempty"`
}

func (task *TektonTask) Init(config *TektonTaskConfig) *TektonTask {
	task.APIVersion = "tekton.dev/v1beta1"
	task.Kind = "Task"
	task.Metadata = &TektonMetadata{Name: config.Name}

	volumeName := "docker-socket"
	volumeMount := &TektonVolumeMount{
		Name:      volumeName,
		MountPath: "/var/run/docker.sock",
	}

	workspaceRoot := "$(workspaces.source.path)"
	args := []string{
		"fn",
		"run",
		path.Join(workspaceRoot, config.Dir),
	}
	if len(config.FnPaths) > 0 {
		for _, fnPath := range config.FnPaths {
			args = append(
				args,
				"--fn-path",
				path.Join(workspaceRoot, fnPath),
			)
		}
	}

	step := &TektonTaskStep{
		Name:         config.Name,
		Image:        KptImage,
		Args:         args,
		VolumeMounts: []*TektonVolumeMount{volumeMount},
	}

	volume := &TektonVolume{
		Name: volumeName,
		HostPath: &TektonVolumeHostPath{
			Path: "/var/run/docker.sock",
			Type: "Socket",
		},
	}

	workspace := &TektonWorkspace{
		Name:      "source",
		MountPath: "/source",
	}

	task.Spec = &TektonTaskSpec{
		Workspaces: []*TektonWorkspace{workspace},
		Steps:      []*TektonTaskStep{step},
		Volumes:    []*TektonVolume{volume},
	}

	return task
}

func (task *TektonTask) Generate() (out []byte, err error) {
	return yaml.Marshal(task)
}

// TektonTaskSpec describes the spec of a Task object.
type TektonTaskSpec struct {
	Workspaces []*TektonWorkspace `yaml:",omitempty"`
	Steps      []*TektonTaskStep  `yaml:",omitempty"`
	Volumes    []*TektonVolume    `yaml:",omitempty"`
}

// TektonTaskStep is a step in the Task spec.
type TektonTaskStep struct {
	Name         string               `yaml:",omitempty"`
	Image        string               `yaml:",omitempty"`
	Args         []string             `yaml:",omitempty"`
	VolumeMounts []*TektonVolumeMount `yaml:"volumeMounts,omitempty"`
}

// TektonVolumeMount mounts a volume to a path.
type TektonVolumeMount struct {
	Name      string `yaml:",omitempty"`
	MountPath string `yaml:"mountPath,omitempty"`
}

// TektonVolume describes a mountable volume on the host.
type TektonVolume struct {
	Name     string                `yaml:",omitempty"`
	HostPath *TektonVolumeHostPath `yaml:"hostPath,omitempty"`
}

// TektonVolumeHostPath indicates the path and its file type of a file on the host.
type TektonVolumeHostPath struct {
	Path string `yaml:",omitempty"`
	Type string `yaml:",omitempty"`
}
