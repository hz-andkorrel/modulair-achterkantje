package handlers

import (
	"archive/zip"
	"bufio"
	"fmt"
	"hotelhub/broker/services"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

type PluginUploadHandler struct {
	config *services.Configuration
}

func NewPluginUploadHandler(config *services.Configuration) BaseHandler {
	return &PluginUploadHandler{config: config}
}

func (h *PluginUploadHandler) RegisterRoutes(engine *gin.Engine, jwt *services.JwtService) {
	log.Println("[PluginUploadHandler] RegisterRoutes called")

	engine.GET("/plugin/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	engine.POST("/plugin/upload", h.uploadAndDeploy())
}

func (h *PluginUploadHandler) uploadAndDeploy() gin.HandlerFunc {
	return func(c *gin.Context) {
		file, header, err := c.Request.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing form file 'file'", "details": err.Error()})
			return
		}
		defer file.Close()

		if !strings.HasSuffix(strings.ToLower(header.Filename), ".zip") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "only .zip files allowed"})
			return
		}

		rawSlug := strings.TrimSuffix(header.Filename, filepath.Ext(header.Filename))
		slug := sanitizeSlug(rawSlug)
		if slug == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid filename -> empty slug"})
			return
		}

		pluginRoot := h.config.PluginDirectory
		if strings.TrimSpace(pluginRoot) == "" {
			pluginRoot = "./plugins"
		}

		pluginDir := filepath.Join(pluginRoot, slug)
		if err := os.MkdirAll(pluginDir, 0755); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot create plugin dir", "details": err.Error()})
			return
		}

		zipPath := filepath.Join(pluginDir, header.Filename)
		out, err := os.Create(zipPath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot save zip", "details": err.Error()})
			return
		}
		if _, err := io.Copy(out, file); err != nil {
			out.Close()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot write zip", "details": err.Error()})
			return
		}
		out.Close()

		// Unzip safely (ZipSlip protection)
		if err := unzipSafe(zipPath, pluginDir); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot unzip", "details": err.Error()})
			return
		}

		// Find docker-compose.yml or docker-compose.yaml (root or 1-level deep)
		composeDir, composeFile, err := findComposeFile(pluginDir)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "docker-compose file not found", "details": err.Error()})
			return
		}

		// Create patched runtime compose file for broker network integration
		runtimeComposeFile := "docker-compose.runtime.yml"
		runtimeComposePath := filepath.Join(composeDir, runtimeComposeFile)
		if err := writeRuntimeCompose(filepath.Join(composeDir, composeFile), runtimeComposePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot create runtime compose file", "details": err.Error()})
			return
		}

		// Project name for docker compose: hotelhub-plugin-<slug>
		projectName := "hotelhub-plugin-" + slug

		// Run docker compose up -d --build (v2 with fallback to v1)
		if err := services.DockerComposeUp(composeDir, runtimeComposeFile, projectName); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "docker compose up failed", "details": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"slug":               slug,
			"pluginDir":          pluginDir,
			"composeDir":         composeDir,
			"runtimeComposeFile": runtimeComposeFile,
			"projectName":        projectName,
		})
	}
}

func sanitizeSlug(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, " ", "-")
	re := regexp.MustCompile(`[^a-z0-9\-_]+`)
	s = re.ReplaceAllString(s, "")
	return s
}

func unzipSafe(srcZip, dest string) error {
	r, err := zip.OpenReader(srcZip)
	if err != nil {
		return err
	}
	defer r.Close()

	destAbs, err := filepath.Abs(dest)
	if err != nil {
		return err
	}

	for _, f := range r.File {
		targetPath := filepath.Join(dest, f.Name)
		targetAbs, err := filepath.Abs(targetPath)
		if err != nil {
			return err
		}

		// ZipSlip protection: ensure extracted path is within destination
		if !strings.HasPrefix(targetAbs, destAbs+string(os.PathSeparator)) && targetAbs != destAbs {
			return fmt.Errorf("illegal zip path (ZipSlip): %s", f.Name)
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(targetAbs, 0755); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(targetAbs), 0755); err != nil {
			return err
		}

		in, err := f.Open()
		if err != nil {
			return err
		}
		out, err := os.OpenFile(targetAbs, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, f.Mode())
		if err != nil {
			in.Close()
			return err
		}

		if _, err := io.Copy(out, in); err != nil {
			out.Close()
			in.Close()
			return err
		}

		out.Close()
		in.Close()
	}

	return nil
}

// findComposeFile searches for docker-compose.yml or docker-compose.yaml in pluginDir (root or 1 level deep).
// Returns (composeDir, composeFileName, error).
func findComposeFile(pluginDir string) (string, string, error) {
	names := []string{"docker-compose.yml", "docker-compose.yaml"}

	// 1. Check root of pluginDir
	for _, name := range names {
		p := filepath.Join(pluginDir, name)
		if _, err := os.Stat(p); err == nil {
			return pluginDir, name, nil
		}
	}

	// 2. Check 1 level deep (e.g., mews-plugin/backend/docker-compose.yml)
	entries, err := os.ReadDir(pluginDir)
	if err != nil {
		return "", "", err
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		subDir := filepath.Join(pluginDir, e.Name())
		for _, name := range names {
			p := filepath.Join(subDir, name)
			if _, err := os.Stat(p); err == nil {
				return subDir, name, nil
			}
		}
	}

	return "", "", fmt.Errorf("docker-compose.yml/yaml not found in %s (searched root + 1 level deep)", pluginDir)
}

// writeRuntimeCompose reads the original compose file and writes a patched version for broker integration.
// The patched file:
// - Removes container_name: lines (avoids conflicts when multiple plugins run)
// - Removes ports: blocks entirely (plugins should not bind host ports; broker is on 8080)
// - Rewrites REDIS_ADDR=host.docker.internal:6379 -> REDIS_ADDR=hotelhub-bus:6379 (use broker's redis)
// - Adds networks: [broker_net] to each service so plugin containers join the broker's private network
// - Adds a top-level networks section for the external broker network
func writeRuntimeCompose(srcPath, dstPath string) error {
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("cannot open source compose file: %w", err)
	}
	defer srcFile.Close()

	var lines []string
	scanner := bufio.NewScanner(srcFile)
	inPorts := false
	portsIndent := 0

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Detect start of ports: block and skip it entirely
		// ports: can appear at service level, so we track indentation
		if strings.HasPrefix(trimmed, "ports:") {
			inPorts = true
			portsIndent = countLeadingSpaces(line)
			continue // skip this line
		}

		// If inside ports block, skip until we reach a line with same or less indentation (new key)
		if inPorts {
			if trimmed == "" {
				// blank line inside ports block, skip
				continue
			}
			currentIndent := countLeadingSpaces(line)
			if currentIndent <= portsIndent && !strings.HasPrefix(trimmed, "-") {
				// We've exited the ports block (new key at same or higher level)
				inPorts = false
			} else {
				// Still inside ports block (list items or nested)
				continue
			}
		}

		// Remove container_name: lines to avoid conflicts between plugins
		if strings.HasPrefix(trimmed, "container_name:") {
			continue
		}

		// Rewrite REDIS_ADDR from host.docker.internal to hotelhub-bus (broker's redis container)
		if strings.Contains(line, "REDIS_ADDR=host.docker.internal:6379") {
			line = strings.ReplaceAll(line, "REDIS_ADDR=host.docker.internal:6379", "REDIS_ADDR=hotelhub-bus:6379")
		}
		if strings.Contains(line, "REDIS_ADDR=localhost:6379") {
			line = strings.ReplaceAll(line, "REDIS_ADDR=localhost:6379", "REDIS_ADDR=hotelhub-bus:6379")
		}

		lines = append(lines, line)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading compose file: %w", err)
	}

	// Add networks: [broker_net] to each service (simple approach: add after each line that looks like a service definition)
	// Also add the top-level networks block at the end
	patchedLines := addNetworkToServices(lines)

	// Append top-level networks block for external broker network
	patchedLines = append(patchedLines, "")
	patchedLines = append(patchedLines, "# Added by broker: join the broker's private network")
	patchedLines = append(patchedLines, "networks:")
	patchedLines = append(patchedLines, "  broker_net:")
	patchedLines = append(patchedLines, "    external: true")
	patchedLines = append(patchedLines, "    name: modulair-achterkantje_private_network")

	// Write patched file
	dstFile, err := os.Create(dstPath)
	if err != nil {
		return fmt.Errorf("cannot create runtime compose file: %w", err)
	}
	defer dstFile.Close()

	for _, l := range patchedLines {
		if _, err := dstFile.WriteString(l + "\n"); err != nil {
			return fmt.Errorf("cannot write to runtime compose file: %w", err)
		}
	}

	return nil
}

// countLeadingSpaces returns the number of leading spaces in a line
func countLeadingSpaces(s string) int {
	return len(s) - len(strings.TrimLeft(s, " "))
}

// addNetworkToServices injects networks: [broker_net] into each service definition.
// A simple heuristic: after finding the last property of a service (before next service or top-level key),
// add the networks line. This is a simplified approach that works for most compose files.
func addNetworkToServices(lines []string) []string {
	var result []string
	inServices := false
	serviceLineIndices := []int{} // indices where services start

	// First pass: identify service blocks
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		indent := countLeadingSpaces(line)

		if trimmed == "services:" {
			inServices = true
			continue
		}

		// Top-level keys end the services block
		if inServices && indent == 0 && trimmed != "" && !strings.HasPrefix(trimmed, "#") {
			inServices = false
		}

		// Service definition: 2-space indent under services:, ends with ':'
		if inServices && indent == 2 && strings.HasSuffix(trimmed, ":") && !strings.HasPrefix(trimmed, "#") {
			serviceLineIndices = append(serviceLineIndices, i)
		}
	}

	// Second pass: insert networks: [broker_net] at the end of each service
	insertAfter := make(map[int]bool)

	for idx, serviceStart := range serviceLineIndices {
		// Find where this service ends (next service or end of services block)
		var serviceEnd int
		if idx < len(serviceLineIndices)-1 {
			serviceEnd = serviceLineIndices[idx+1] - 1
		} else {
			// Last service: find where services block ends
			serviceEnd = len(lines) - 1
			for i := serviceStart + 1; i < len(lines); i++ {
				trimmed := strings.TrimSpace(lines[i])
				indent := countLeadingSpaces(lines[i])
				// Top-level key or networks: block
				if indent == 0 && trimmed != "" && !strings.HasPrefix(trimmed, "#") {
					serviceEnd = i - 1
					break
				}
			}
		}

		// Find last non-empty line of this service to insert after
		for i := serviceEnd; i > serviceStart; i-- {
			if strings.TrimSpace(lines[i]) != "" {
				insertAfter[i] = true
				break
			}
		}
	}

	for i, line := range lines {
		result = append(result, line)
		if insertAfter[i] {
			// Insert networks line with proper indentation (4 spaces for service properties)
			result = append(result, "    networks:")
			result = append(result, "      - broker_net")
		}
	}

	return result
}
