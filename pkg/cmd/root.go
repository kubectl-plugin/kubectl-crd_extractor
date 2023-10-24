// Copyright 2023 The kubectl-plugin Authors
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/client-go/tools/clientcmd/api"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"

	"github.com/kubectl-plugin/kubectl-crd_extractor/pkg/extractor"
)

const extractorExample = `
	# extracts CRD to JSON schema
	%[1]s crd-extractor [CRD GVK]
`

type ExtractorOptions struct {
	genericiooptions.IOStreams

	configFlags *genericclioptions.ConfigFlags
	factory     cmdutil.Factory

	args      []string
	rawConfig api.Config

	cluster   string
	context   string
	authInfo  string
	namespace string

	resultingContext     *api.Context
	resultingContextName string
	listNamespaces       bool
}

func NewExtractorOptions(streams genericiooptions.IOStreams) *ExtractorOptions {
	return &ExtractorOptions{
		IOStreams: streams,

		configFlags: genericclioptions.NewConfigFlags(true),
	}
}

func defaultConfigFlags() *genericclioptions.ConfigFlags {
	return genericclioptions.NewConfigFlags(true).
		WithDeprecatedPasswordFlag().
		WithDiscoveryBurst(300).
		WithDiscoveryQPS(50.0)
}

func NewCmdCRDExtractor(streams genericiooptions.IOStreams) *cobra.Command {
	o := NewExtractorOptions(streams)

	kubeConfigFlags := o.configFlags
	if kubeConfigFlags == nil {
		kubeConfigFlags = defaultConfigFlags()
	}
	matchVersionFlags := cmdutil.NewMatchVersionFlags(kubeConfigFlags)
	o.factory = cmdutil.NewFactory(matchVersionFlags)

	cmd := &cobra.Command{
		Use:          "crd-extractor [CRD GVK] [flags]",
		Short:        "Extracts CRD to JSON Schema",
		Example:      fmt.Sprintf(extractorExample, "kubectl"),
		SilenceUsage: true,
		Run: func(c *cobra.Command, args []string) {
			o.args = args
			cmdutil.CheckErr(o.Run(c.Context()))
		},
	}

	return cmd
}

func (o *ExtractorOptions) Run(ctx context.Context) error {
	g, v, _ := strings.Cut(o.args[0], "/")
	k := o.args[1]
	if err := extractor.GetCRDs(ctx, o.factory, g, v, k); err != nil {
		return fmt.Errorf("GetCRDs: %w", err)
	}

	return nil
}
