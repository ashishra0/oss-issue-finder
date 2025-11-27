package output

import (
	"fmt"
	"os/exec"
	"runtime"
)

// SendNotification sends a desktop notification
func SendNotification(title, message string) error {
	switch runtime.GOOS {
	case "darwin":
		return sendMacOSNotification(title, message)
	case "linux":
		return sendLinuxNotification(title, message)
	case "windows":
		return sendWindowsNotification(title, message)
	default:
		return fmt.Errorf("notifications not supported on %s", runtime.GOOS)
	}
}

func sendMacOSNotification(title, message string) error {
	script := fmt.Sprintf(`display notification "%s" with title "%s" sound name "Glass"`, message, title)
	cmd := exec.Command("osascript", "-e", script)
	return cmd.Run()
}

func sendLinuxNotification(title, message string) error {
	cmd := exec.Command("notify-send", title, message)
	return cmd.Run()
}

func sendWindowsNotification(title, message string) error {
	script := fmt.Sprintf(`
		[Windows.UI.Notifications.ToastNotificationManager, Windows.UI.Notifications, ContentType = WindowsRuntime] | Out-Null
		[Windows.UI.Notifications.ToastNotification, Windows.UI.Notifications, ContentType = WindowsRuntime] | Out-Null
		[Windows.Data.Xml.Dom.XmlDocument, Windows.Data.Xml.Dom.XmlDocument, ContentType = WindowsRuntime] | Out-Null

		$template = @"
		<toast>
			<visual>
				<binding template="ToastText02">
					<text id="1">%s</text>
					<text id="2">%s</text>
				</binding>
			</visual>
		</toast>
"@

		$xml = New-Object Windows.Data.Xml.Dom.XmlDocument
		$xml.LoadXml($template)
		$toast = New-Object Windows.UI.Notifications.ToastNotification $xml
		[Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier("Issue Finder").Show($toast)
	`, title, message)

	cmd := exec.Command("powershell", "-Command", script)
	return cmd.Run()
}
