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

package commands

import (
	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/cmd/config/configcobra"

	"github.com/GoogleContainerTools/kpt/internal/cmdexport"
	"github.com/GoogleContainerTools/kpt/internal/docs/generated/fndocs"
)

func GetFnCommand(name string) *cobra.Command {
	functions := &cobra.Command{
		Use:     "fn",
		Short:   fndocs.FnShort,
		Long:    fndocs.FnLong,
		Example: fndocs.FnExamples,
		Aliases: []string{"functions"},
		RunE: func(cmd *cobra.Command, args []string) error {
			h, err := cmd.Flags().GetBool("help")
			if err != nil {
				return err
			}
			if h {
				return cmd.Help()
			}
			return cmd.Usage()
		},
	}

	run := configcobra.RunFn(name)
	run.Short = fndocs.RunShort
	run.Long = fndocs.RunShort + "\n" + fndocs.RunLong
	run.Example = fndocs.RunExamples

	source := configcobra.Source(name)
	source.Short = fndocs.SourceShort
	source.Long = fndocs.SourceShort + "\n" + fndocs.SourceLong
	source.Example = fndocs.SourceExamples

	sink := configcobra.Sink(name)
	sink.Short = fndocs.SinkShort
	sink.Long = fndocs.SinkShort + "\n" + fndocs.SinkLong
	sink.Example = fndocs.SinkExamples

	functions.AddCommand(run, source, sink, cmdexport.ExportCommand())
	return functions
}
