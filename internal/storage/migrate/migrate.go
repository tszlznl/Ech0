package migrate

import (
	"context"
	"errors"
	"fmt"

	stgx "github.com/lin-snow/ech0/pkg/storagex"
)

var errSkipped = errors.New("skipped")

type ConflictPolicy int

const (
	ConflictSkip      ConflictPolicy = iota
	ConflictOverwrite
)

type Options struct {
	Conflict   ConflictPolicy
	DryRun     bool
	OnProgress func(path string, err error)
}

type Result struct {
	Copied  int
	Skipped int
	Errors  int
}

// Copy copies all files under prefix from src to dst.
// Directories are traversed recursively.
func Copy(ctx context.Context, src, dst stgx.FS, prefix string, opts Options) (Result, error) {
	entries, err := src.List(ctx, prefix)
	if err != nil {
		return Result{}, fmt.Errorf("list source %q: %w", prefix, err)
	}

	var result Result
	for _, entry := range entries {
		if entry.IsDir {
			sub, err := Copy(ctx, src, dst, entry.Path, opts)
			if err != nil {
				return result, err
			}
			result.Copied += sub.Copied
			result.Skipped += sub.Skipped
			result.Errors += sub.Errors
			continue
		}

		err := copyFile(ctx, src, dst, entry.Path, opts)
		switch {
		case errors.Is(err, errSkipped):
			result.Skipped++
		case err != nil:
			result.Errors++
			if opts.OnProgress != nil {
				opts.OnProgress(entry.Path, err)
			}
		default:
			result.Copied++
			if opts.OnProgress != nil {
				opts.OnProgress(entry.Path, nil)
			}
		}
	}
	return result, nil
}

func copyFile(ctx context.Context, src, dst stgx.FS, path string, opts Options) error {
	if opts.Conflict == ConflictSkip {
		exists, err := dst.Exists(ctx, path)
		if err != nil {
			return err
		}
		if exists {
			return errSkipped
		}
	}

	if opts.DryRun {
		return nil
	}

	rc, err := src.Open(ctx, path)
	if err != nil {
		return fmt.Errorf("open %s: %w", path, err)
	}
	defer rc.Close()

	info, _ := src.Stat(ctx, path)
	wopts := stgx.WriteOptions{}
	if info != nil {
		wopts.ContentType = info.ContentType
	}
	return dst.Write(ctx, path, rc, wopts)
}
