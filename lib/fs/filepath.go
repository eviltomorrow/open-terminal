package fs

import "path/filepath"

func ResetPath(baseDir, path string) string {
	if path != "" {
		if !filepath.IsAbs(path) {
			path = filepath.Join(baseDir, path)
		}
	}
	return path
}
