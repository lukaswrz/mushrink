package planning

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"

	"github.com/hashicorp/go-multierror"
	"hack.moontide.ink/lukas/mushrink/internal/pathutil"
	"hack.moontide.ink/lukas/mushrink/internal/transcoding"
)

type TaskAct int

const (
	TranscodeAct TaskAct = iota
	CleanAct
)

type Task struct {
	act            TaskAct
	source, target string
}

func (t Task) String() string {
	switch t.act {
	case TranscodeAct:
		return "transcode " + t.source + " → " + t.target
	case CleanAct:
		return "clean " + t.target
	}

	return ""
}

func (t Task) Run() error {
	switch t.act {
	case TranscodeAct:
		return transcoding.Transcode(t.source, t.target)
	case CleanAct:
		return os.Remove(t.target)
	}

	return errors.New("unknown action")
}

func Plan(from, to string, re *regexp.Regexp, repl string) ([]Task, error) {
	tasks := []Task{}
	var merr *multierror.Error

	var walkErr error

	var seen []string
	walkErr = filepath.WalkDir(from, func(source string, d fs.DirEntry, err error) error {
		if err != nil {
			merr = multierror.Append(merr, fmt.Errorf("walk: %w", err))
			return nil
		}

		if d.IsDir() {
			return nil
		}

		base := filepath.Base(source)
		if !re.MatchString(filepath.Base(source)) {
			return nil
		}
		base = re.ReplaceAllString(base, repl)

		sourceDir := filepath.Dir(source)
		var targetDir string
		targetDir, err = pathutil.Reparent(from, to, sourceDir)
		if err != nil {
			merr = multierror.Append(merr, err)
			return nil
		}

		target := filepath.Join(targetDir, base)
		if !pathutil.InDir(targetDir, target) {
			return fmt.Errorf(
				"target file %s is not in target directory %s, this is likely because of an erroneous match/replace combination",
				target,
				targetDir,
			)
		}

		_, err = os.Lstat(target)
		if err == nil {
			// If the target file already exists, leave it alone.
			seen = append(seen, target)
			return nil
		} else if !os.IsNotExist(err) {
			return err
		}

		tasks = append(tasks, Task{
			act:    TranscodeAct,
			source: source,
			target: target,
		})

		return nil
	})
	if walkErr != nil {
		merr = multierror.Append(merr, walkErr)
	}

	walkErr = filepath.WalkDir(to, func(target string, d fs.DirEntry, err error) error {
		if err != nil {
			merr = multierror.Append(merr, fmt.Errorf("walk: %w", err))
			return nil
		}

		if d.IsDir() {
			return nil
		}

		wanted := false
		for _, task := range tasks {
			if task.act != TranscodeAct {
				continue
			}

			if task.target == target {
				wanted = true
			}
		}
		for _, seenTarget := range seen {
			if seenTarget == target {
				wanted = true
			}
		}

		if !wanted {
			tasks = append(tasks, Task{
				act:    CleanAct,
				target: target,
			})
		}

		return nil
	})
	if walkErr != nil {
		merr = multierror.Append(merr, walkErr)
	}

	return tasks, merr.ErrorOrNil()
}
