//go:build !windows

package usecase

func localYTDLPFileVersionString(_ string) string {
	return ""
}
