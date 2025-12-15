package services

import (
	"fmt"
	"log"
	"os/exec"
)

// DockerComposeUp runs docker compose up -d --build for the given compose file.
// It first tries "docker compose" (v2 plugin), then falls back to "docker-compose" (v1 standalone).
// The projectName is used to namespace all resources (containers, networks, etc.).
func DockerComposeUp(composeDir string, composeFile string, projectName string) error {
	log.Printf("[DockerComposeUp] Starting compose up in dir=%s file=%s project=%s", composeDir, composeFile, projectName)

	// Try docker compose (v2) first
	cmd := exec.Command("docker", "compose", "-f", composeFile, "-p", projectName, "up", "-d", "--build")
	cmd.Dir = composeDir
	out, err := cmd.CombinedOutput()
	if err == nil {
		log.Printf("[DockerComposeUp] docker compose v2 succeeded: %s", string(out))
		return nil
	}
	log.Printf("[DockerComposeUp] docker compose v2 failed: %v | %s", err, string(out))

	// Fallback to docker-compose (v1) if docker compose fails
	cmd2 := exec.Command("docker-compose", "-f", composeFile, "-p", projectName, "up", "-d", "--build")
	cmd2.Dir = composeDir
	out2, err2 := cmd2.CombinedOutput()
	if err2 != nil {
		return fmt.Errorf("docker compose (v2) failed: %v | %s\ndocker-compose (v1) failed: %v | %s", err, string(out), err2, string(out2))
	}
	log.Printf("[DockerComposeUp] docker-compose v1 succeeded: %s", string(out2))
	return nil
}
