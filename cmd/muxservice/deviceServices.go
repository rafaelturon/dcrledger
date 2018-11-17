package muxservice

import (
	"os/exec"
	"time"
)

func executeBashCommand(argCmd string) (string, error) {
	var timer *time.Timer
	bashCmd := exec.Command("bash", "-c", argCmd)
	logger.Debugf("Executing bash ~$ %s", argCmd)
	timer = time.AfterFunc(3*time.Second, func() {
		timer.Stop()
		bashCmd.Process.Kill()
	})
	bashOut, err := bashCmd.Output()

	if err != nil {
		return "", err
	}
	return string(bashOut), nil
}

// TurnOffDevice enable an immediate shut down command execution
func TurnOffDevice() (string, error) {
	shutdownArg := "sudo shutdown -h 1"

	return executeBashCommand(shutdownArg)
}
