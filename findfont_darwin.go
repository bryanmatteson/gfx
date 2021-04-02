// +build darwin

package gfx

func getFontDirectories() (paths []string) {
	return append(paths, expandUser("~/Library/Fonts/"), "/Library/Fonts/", "/System/Library/Fonts/", "/System/Library/Fonts/Supplemental")
}
