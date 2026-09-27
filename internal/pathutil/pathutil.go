package pathutil

import (
	"os"
	"path/filepath"
)

func InDir(dir, path string) bool {
	var err error

	dir, err = filepath.Abs(dir)
	if err != nil {
		return false
	}

	path, err = filepath.Abs(path)
	if err != nil {
		return false
	}

	rel, err := filepath.Rel(dir, path)
	if err != nil {
		return false
	}

	return filepath.IsLocal(rel)
}

func Reparent(from, to, path string) (result string, err error) {
	from, err = filepath.Abs(from)
	if err != nil {
		return
	}

	path, err = filepath.Abs(path)
	if err != nil {
		return
	}

	rel, err := filepath.Rel(from, path)
	if err != nil {
		return
	}

	result = filepath.Join(to, rel)
	return
}

func Resolve(path string) (string, error) {
	var (
		err          error
		wd, abs, rel string
	)

	path, err = filepath.EvalSymlinks(path)
	if err != nil {
		return "", err
	}

	abs, err = filepath.Abs(path)
	if err != nil {
		return "", err
	}

	wd, err = os.Getwd()
	if err != nil {
		return "", err
	}

	rel, err = filepath.Rel(wd, abs)
	if err != nil {
		return "", err
	}

	if len([]rune(rel)) > len([]rune(abs)) {
		return abs, nil
	}

	return rel, nil
}
