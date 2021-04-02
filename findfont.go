package gfx

import (
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

type FontKind int

const (
	Unknown FontKind = iota
	TrueType
	OpenType
	TrueTypeCollection
	OpenTypeCollection
)

type SystemFontRecord struct {
	Kind FontKind
	Path string
	Name string
}

func GetSystemFonts() (records []SystemFontRecord) {
	for _, dir := range getFontDirectories() {
		filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}

			if kind, ok := isFontFile(path); ok {
				_, filename := filepath.Split(path)
				name := strings.TrimSuffix(filename, filepath.Ext(filename))
				records = append(records, SystemFontRecord{Kind: kind, Path: path, Name: name})
			}

			return nil
		})
	}
	return
}

func isFontFile(fileName string) (FontKind, bool) {
	lower := strings.ToLower(fileName)
	switch {
	case strings.HasSuffix(lower, ".ttf"):
		return TrueType, true
	case strings.HasSuffix(lower, ".ttc"):
		return TrueTypeCollection, true
	case strings.HasSuffix(lower, ".otf"):
		return OpenType, true
	case strings.HasSuffix(lower, ".otc"):
		return OpenTypeCollection, true
	}

	return Unknown, false
}

func expandUser(path string) (expandedPath string) {
	if strings.HasPrefix(path, "~") {
		if u, err := user.Current(); err == nil {
			return strings.Replace(path, "~", u.HomeDir, -1)
		}
	}
	return path
}
