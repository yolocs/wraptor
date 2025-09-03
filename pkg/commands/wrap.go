package commands

import (
	"context"
	"fmt"

	"github.com/abcxyz/pkg/cli"
	"github.com/yolocs/wraptor/pkg/download/huggingface"
	"github.com/yolocs/wraptor/pkg/wrap"
)

type WrapCommand struct {
	cli.BaseCommand

	flagSource     string
	flagImage      string
	flagBaseImage  string
	flagFilePrefix string
	flagCacheDir   string
}

func (c *WrapCommand) Desc() string {
	return "Wrap the files from the source into a container image."
}

func (c *WrapCommand) Help() string {
	return `
Usage: {{ COMMAND }} [options]

Wrap the files from the source into a container image.

	{{ COMMAND }} --source /path/to/source --image my-image:latest
`
}

func (c *WrapCommand) Flags() *cli.FlagSet {
	set := c.NewFlagSet()

	sec := set.NewSection("OPTIONS")
	sec.StringVar(&cli.StringVar{
		Name:   "source",
		Usage:  "Path to the source files (local directory or HuggingFace repo).",
		Target: &c.flagSource,
		EnvVar: "WRAPTOR_WRAP_SOURCE",
	})
	sec.StringVar(&cli.StringVar{
		Name:   "image",
		Usage:  "Name of the container image to create.",
		Target: &c.flagImage,
		EnvVar: "WRAPTOR_WRAP_IMAGE",
	})
	sec.StringVar(&cli.StringVar{
		Name:   "base",
		Usage:  "Name of the base container image.",
		Target: &c.flagBaseImage,
		EnvVar: "WRAPTOR_WRAP_BASE_IMAGE",
	})
	sec.StringVar(&cli.StringVar{
		Name:   "file-prefix",
		Usage:  "Prefix to add to all files in the container image.",
		Target: &c.flagFilePrefix,
		EnvVar: "WRAPTOR_WRAP_FILE_PREFIX",
	})
	sec.StringVar(&cli.StringVar{
		Name:   "cache-dir",
		Usage:  "Path to the cache directory.",
		Target: &c.flagCacheDir,
		EnvVar: "WRAPTOR_WRAP_CACHE_DIR",
	})

	return set
}

func (c *WrapCommand) Run(ctx context.Context, args []string) error {
	f := c.Flags()
	if err := f.Parse(args); err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}

	if c.flagSource == "" {
		return fmt.Errorf("source flag is required")
	}
	if c.flagImage == "" {
		return fmt.Errorf("image flag is required")
	}

	hd := huggingface.New(c.flagSource, huggingface.WithCacheDir(c.flagCacheDir))
	// defer hd.Cleanup()
	fs, err := hd.Load(ctx)
	if err != nil {
		return fmt.Errorf("failed to load files from source: %w", err)
	}

	w := wrap.NewWrapper(
		wrap.WithBaseImage(c.flagBaseImage),
		wrap.WithFilePrefix(c.flagFilePrefix),
	)

	for _, f := range fs {
		if err := w.AppendFiles(f); err != nil {
			return fmt.Errorf("failed to append files: %w", err)
		}
	}

	if err := w.ToRemote(c.flagImage); err != nil {
		return fmt.Errorf("failed to push image to remote: %w", err)
	}

	return nil
}
