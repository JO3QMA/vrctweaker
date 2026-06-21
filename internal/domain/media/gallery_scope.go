package media

import (
	"path/filepath"
	"strings"
)

// PictureFolderPathPrefix returns the FilePathPrefix value for listing screenshots
// under pictureFolderRoot without matching sibling paths (e.g. VRChat vs VRChat_old).
func PictureFolderPathPrefix(pictureFolderRoot string) string {
	pictureFolderRoot = strings.TrimSpace(pictureFolderRoot)
	if pictureFolderRoot == "" {
		return ""
	}
	pictureFolderRoot = filepath.Clean(pictureFolderRoot)
	if pictureFolderRoot == "." {
		return ""
	}
	return pictureFolderRoot + string(filepath.Separator)
}
