// +build linux

package gfx

func getFontDirectories() []string {
	directories := getUserFontDirs()
	directories = append(directories, getSystemFontDirs()...)
	return directories
}

func getUserFontDirs() (paths []string) {
	if dataPath := os.Getenv("XDG_DATA_HOME"); dataPath != "" {
		return []string{expandUser("~/.fonts/"), filepath.Join(expandUser(dataPath), "fonts")}
	}
	return []string{expandUser("~/.fonts/"), expandUser("~/.local/share/fonts/")}
}

func getSystemFontDirs() (paths []string) {
	if dataPaths := os.Getenv("XDG_DATA_DIRS"); dataPaths != "" {
		for _, dataPath := range filepath.SplitList(dataPaths) {
			paths = append(paths, filepath.Join(expandUser(dataPath), "fonts"))
		}
	}
	paths = append(paths, "/usr/local/share/fonts/", "/usr/share/fonts/")
	return
}
