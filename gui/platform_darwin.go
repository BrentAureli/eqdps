package main

import "os/exec"

func openExternalURL(url string) error {
	return exec.Command("open", url).Start()
}
