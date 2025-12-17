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

// DockerComposeStart starts existing containers for a project (without rebuilding).
// Used to resume a stopped plugin.
func DockerComposeStart(composeDir string, composeFile string, projectName string) error {
	log.Printf("[DockerComposeStart] Starting containers in dir=%s file=%s project=%s", composeDir, composeFile, projectName)

	cmd := exec.Command("docker", "compose", "-f", composeFile, "-p", projectName, "start")
	cmd.Dir = composeDir
	out, err := cmd.CombinedOutput()
	if err == nil {
		log.Printf("[DockerComposeStart] docker compose v2 start succeeded: %s", string(out))
		return nil
	}
	log.Printf("[DockerComposeStart] docker compose v2 start failed: %v | %s", err, string(out))

	// Fallback to docker-compose (v1)
	cmd2 := exec.Command("docker-compose", "-f", composeFile, "-p", projectName, "start")
	cmd2.Dir = composeDir
	out2, err2 := cmd2.CombinedOutput()
	if err2 != nil {
		return fmt.Errorf("docker compose start (v2) failed: %v | %s\ndocker-compose start (v1) failed: %v | %s", err, string(out), err2, string(out2))
	}
	log.Printf("[DockerComposeStart] docker-compose v1 start succeeded: %s", string(out2))
	return nil
}

// DockerComposeStop stops running containers for a project (without removing them).
// Used to pause a plugin without uninstalling.
func DockerComposeStop(composeDir string, composeFile string, projectName string) error {
	log.Printf("[DockerComposeStop] Stopping containers in dir=%s file=%s project=%s", composeDir, composeFile, projectName)

	cmd := exec.Command("docker", "compose", "-f", composeFile, "-p", projectName, "stop")
	cmd.Dir = composeDir
	out, err := cmd.CombinedOutput()
	if err == nil {
		log.Printf("[DockerComposeStop] docker compose v2 stop succeeded: %s", string(out))
		return nil
	}
	log.Printf("[DockerComposeStop] docker compose v2 stop failed: %v | %s", err, string(out))

	// Fallback to docker-compose (v1)
	cmd2 := exec.Command("docker-compose", "-f", composeFile, "-p", projectName, "stop")
	cmd2.Dir = composeDir
	out2, err2 := cmd2.CombinedOutput()
	if err2 != nil {
		return fmt.Errorf("docker compose stop (v2) failed: %v | %s\ndocker-compose stop (v1) failed: %v | %s", err, string(out), err2, string(out2))
	}
	log.Printf("[DockerComposeStop] docker-compose v1 stop succeeded: %s", string(out2))
	return nil
}

// DockerComposeDown stops and removes containers, networks, and volumes for a project.
// Used to fully uninstall a plugin.
func DockerComposeDown(composeDir string, composeFile string, projectName string) error {
	log.Printf("[DockerComposeDown] Removing containers in dir=%s file=%s project=%s", composeDir, composeFile, projectName)

	cmd := exec.Command("docker", "compose", "-f", composeFile, "-p", projectName, "down", "--volumes", "--remove-orphans")
	cmd.Dir = composeDir
	out, err := cmd.CombinedOutput()
	if err == nil {
		log.Printf("[DockerComposeDown] docker compose v2 down succeeded: %s", string(out))
		return nil
	}
	log.Printf("[DockerComposeDown] docker compose v2 down failed: %v | %s", err, string(out))

	// Fallback to docker-compose (v1)
	cmd2 := exec.Command("docker-compose", "-f", composeFile, "-p", projectName, "down", "--volumes", "--remove-orphans")
	cmd2.Dir = composeDir
	out2, err2 := cmd2.CombinedOutput()
	if err2 != nil {
		return fmt.Errorf("docker compose down (v2) failed: %v | %s\ndocker-compose down (v1) failed: %v | %s", err, string(out), err2, string(out2))
	}
	log.Printf("[DockerComposeDown] docker-compose v1 down succeeded: %s", string(out2))
	return nil
}
