package main

import "os/exec"

func openExternalURL(url string) error {
	return exec.Command("rundll32.exe", "url.dll,FileProtocolHandler", url).Start()
}
