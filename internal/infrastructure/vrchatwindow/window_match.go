package vrchatwindow

const unityWndClassName = "UnityWndClass"

// classOrTitleLooksLikeVRChat reports whether a top-level window is likely the game client.
func classOrTitleLooksLikeVRChat(className, title string) bool {
	if className == unityWndClassName {
		return true
	}
	return len(title) >= 6 && title[:6] == "VRChat"
}
