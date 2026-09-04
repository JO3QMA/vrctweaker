package locale

// TrayMenuLabels returns localized tray menu strings for the given app language code.
func TrayMenuLabels(lang string) (showWindow, quit string) {
	switch lang {
	case "ja":
		return "ウィンドウを表示", "終了"
	case "ko":
		return "창 표시", "종료"
	case "zh-TW":
		return "顯示視窗", "結束"
	case "zh-CN":
		return "显示窗口", "退出"
	default:
		return "Show window", "Quit"
	}
}
