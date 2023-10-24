package main

import (
	"context"
	"os"

	_ "k8s.io/client-go/plugin/pkg/client/auth"

	flag "github.com/spf13/pflag"
	"k8s.io/cli-runtime/pkg/genericiooptions"

	"github.com/kubectl-plugin/kubectl-crd_extractor/pkg/cmd"
)

func main() {
	flag.CommandLine = flag.NewFlagSet("kubectl-crd_extractor", flag.ExitOnError)
	ioStream := genericiooptions.IOStreams{
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	root := cmd.NewCmdCRDExtractor(ioStream)
	if err := root.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}
