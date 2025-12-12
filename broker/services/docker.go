package services

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func DockerBuild(imageName, buildDir string) error {
	cmd := exec.Command("docker", "build", "-t", imageName, buildDir)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("docker build failed: %v | %s", err, string(out))
	}
	return nil
}

func DockerRun(containerName, imageName string) (string, error) {
	// Cleanup if it already exists (keeps it simple)
	_ = exec.Command("docker", "rm", "-f", containerName).Run()

	args := []string{"run", "-d", "--name", containerName}

	// Put plugin on same network as broker/redis (recommended)
	if net := strings.TrimSpace(os.Getenv("PLUGIN_DOCKER_NETWORK")); net != "" {
		args = append(args, "--network", net)
	}

	// Expose plugin HTTP to host.
	// Use dynamic host port to avoid conflicts (docker chooses a free port).
	// Assumes the plugin listens on 8080 inside the container.
	args = append(args, "-p", "0:8080")

	// Provide redis address consistent with your stack (eventbus container)
	// Only matters if the plugin uses it; harmless otherwise.
	args = append(args, "-e", "REDIS_ADDR=hotelhub-bus:6379")

	args = append(args, imageName)

	cmd := exec.Command("docker", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("docker run failed: %v | %s", err, string(out))
	}

	return strings.TrimSpace(string(out)), nil
}
