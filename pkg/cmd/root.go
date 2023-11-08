// Copyright 2023 The kubectl-plugin Authors
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	json "github.com/goccy/go-json"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"sigs.k8s.io/yaml"

	"github.com/kubectl-plugin/kubectl-crd_extractor/pkg/extractor"
	"github.com/kubectl-plugin/kubectl-crd_extractor/pkg/version"
)

const extractorExample = `
  # extracts CRD to JSON schema
  %[1]s crd-extractor [CRD GVK]
`

type CRDExtractorOptions struct {
	genericiooptions.IOStreams

	kubeConfigFlags *genericclioptions.ConfigFlags
	args            []string
	output          string
}

func NewCRDExtractorOptions(streams genericiooptions.IOStreams) *CRDExtractorOptions {
	return &CRDExtractorOptions{
		IOStreams:       streams,
		kubeConfigFlags: genericclioptions.NewConfigFlags(true),
	}
}

func NewCmdCRDExtractor(streams genericiooptions.IOStreams) *cobra.Command {
	o := NewCRDExtractorOptions(streams)

	cmd := &cobra.Command{
		Use:          "crd-extractor [CRD GVK] [flags]",
		Version:      version.Version,
		Short:        "Extracts CRD to JSON Schema",
		Example:      fmt.Sprintf(extractorExample, "kubectl"),
		SilenceUsage: true,
		Run: func(c *cobra.Command, args []string) {
			o.args = args
			cmdutil.CheckErr(o.Run(c.Context()))
		},
	}

	cmd.Flags().StringVar(&o.output, "output", "", `output base directory`)
	o.kubeConfigFlags.AddFlags(cmd.Flags())

	return cmd
}

func (o *CRDExtractorOptions) Run(ctx context.Context) error {
	g, v, ok := strings.Cut(o.args[0], "/")
	if !ok {
		return fmt.Errorf("get ")
	}
	k := o.args[1]

	gvk := schema.GroupVersionKind{
		Group:   g,
		Version: v,
		Kind:    k,
	}
	cr, err := extractor.GetCR(ctx, o.kubeConfigFlags, gvk)
	if err != nil {
		return fmt.Errorf("GetCRDs: %w", err)
	}

	b, err := yaml.Marshal(cr.OpenAPIV3Schema)
	if err != nil {
		return fmt.Errorf("marshal yaml: %w", err)
	}
	data, err := yaml.YAMLToJSON(b)
	if err != nil {
		return fmt.Errorf("YAMLToJSON: %w", err)
	}
	var buf bytes.Buffer
	if err := json.Indent(&buf, data, "", "  "); err != nil {
		return fmt.Errorf("indent json: %w", err)
	}
	// write end of newline
	if err := buf.WriteByte('\n'); err != nil {
		return fmt.Errorf("write byte: %w", err)
	}

	dir := filepath.Join(strings.ToLower(gvk.Group), strings.ToLower(gvk.Version))
	if o.output != "" {
		dir = filepath.Join(o.output, dir)
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create %s directory: %w", dir, err)
	}
	f, err := os.Create(filepath.Join(dir, strings.ToLower(fmt.Sprintf("%s.json", gvk.Kind))))
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer f.Close()

	if _, err := buf.WriteTo(f); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	return nil
}
