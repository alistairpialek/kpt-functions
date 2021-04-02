// Copyright 2019 Google LLC
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

package getioreader

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/GoogleContainerTools/kpt/pkg/kptfile"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/kio/filters"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// Get reads a package from input and applies a pattern for generating filenames.
func Get(path, pattern string, input io.Reader) error {
	if err := os.MkdirAll(path, 0700); err != nil {
		return err
	}

	b := &bytes.Buffer{}
	fs := &filters.FileSetter{FilenamePattern: pattern, Mode: fmt.Sprintf("%d", 0600)}
	err := kio.Pipeline{
		Inputs:  []kio.Reader{&kio.ByteReader{Reader: input}},
		Filters: []kio.Filter{fs, filters.FormatFilter{}},
		Outputs: []kio.Writer{
			kio.ByteWriter{Writer: b, KeepReaderAnnotations: true},
			kio.LocalPackageWriter{PackagePath: path},
		},
	}.Execute()
	if err != nil {
		return err
	}

	k := kptfile.KptFile{
		ResourceMeta: yaml.ResourceMeta{
			ObjectMeta: yaml.ObjectMeta{
				NameMeta: yaml.NameMeta{
					Name: filepath.Base(path)},
			},
			TypeMeta: yaml.TypeMeta{
				Kind: "Kptfile",
			},
		},
		Upstream: kptfile.Upstream{
			Type:  kptfile.StdinOrigin,
			Stdin: kptfile.Stdin{Original: b.String(), FilenamePattern: fs.FilenamePattern},
		},
	}
	f, err := os.OpenFile(filepath.Join(path, kptfile.KptFileName),
		os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	e := yaml.NewEncoder(f)
	defer e.Close()
	return e.Encode(k)
}
