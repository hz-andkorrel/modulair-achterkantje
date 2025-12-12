package services

import (
	"fmt"
	"os/exec"
)

func DockerComposeUp(composeDir string, composeFile string, projectName string) error {
	cmd := exec.Command(
		"docker", "compose",
		"-f", composeFile,
		"-p", projectName,
		"up", "-d", "--build",
	)
	cmd.Dir = composeDir

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("docker compose up failed: %v | %s", err, string(out))
	}
	return nil
}
