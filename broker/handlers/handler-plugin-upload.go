package handlers

import (
	"archive/zip"
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

		// unzip (ZipSlip safe)
		if err := unzipSafe(zipPath, pluginDir); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot unzip", "details": err.Error()})
			return
		}
		buildDir, err := findDockerBuildDir(pluginDir)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "dockerfile not found", "details": err.Error()})
			return
		}

		imageName := "hotelhub-plugin-" + slug
		containerName := imageName

		if err := services.DockerBuild(imageName, buildDir); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "docker build failed", "details": err.Error()})
			return
		}

		containerID, err := services.DockerRun(containerName, imageName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "docker run failed", "details": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"slug":          slug,
			"pluginDir":     pluginDir,
			"buildDir":      buildDir,
			"imageName":     imageName,
			"containerName": containerName,
			"containerId":   containerID,
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

		if !strings.HasPrefix(targetAbs, destAbs+string(os.PathSeparator)) && targetAbs != destAbs {
			return fmt.Errorf("illegal zip path: %s", f.Name)
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

func findDockerBuildDir(pluginDir string) (string, error) {
	// 1) Root Dockerfile?
	if _, err := os.Stat(filepath.Join(pluginDir, "Dockerfile")); err == nil {
		return pluginDir, nil
	}

	// 2) Check common folder names first (jouw plugin: backend/Dockerfile)
	common := []string{"backend", "server", "api", "src"}
	for _, name := range common {
		p := filepath.Join(pluginDir, name, "Dockerfile")
		if _, err := os.Stat(p); err == nil {
			return filepath.Join(pluginDir, name), nil
		}
	}

	// 3) Fallback: search 2 levels deep for a Dockerfile
	entries, err := os.ReadDir(pluginDir)
	if err != nil {
		return "", err
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		sub := filepath.Join(pluginDir, e.Name())
		if _, err := os.Stat(filepath.Join(sub, "Dockerfile")); err == nil {
			return sub, nil
		}

		entries2, err := os.ReadDir(sub)
		if err != nil {
			continue
		}
		for _, e2 := range entries2 {
			if !e2.IsDir() {
				continue
			}
			sub2 := filepath.Join(sub, e2.Name())
			if _, err := os.Stat(filepath.Join(sub2, "Dockerfile")); err == nil {
				return sub2, nil
			}
		}
	}

	return "", fmt.Errorf("Dockerfile not found in %s (searched root + common folders + 2 levels deep)", pluginDir)
}
