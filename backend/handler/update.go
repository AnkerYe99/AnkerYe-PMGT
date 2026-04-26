package handler

import (
	"ankerye-pmgt/util"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"

	"github.com/gin-gonic/gin"
)

const (
	currentVersion = "1.0.0"
	githubRepo     = "AnkerYe99/AnkerYe-PMGT"
	giteaRepo      = "anker/AnkerYe-PMGT"
	giteaBase      = "http://10.14.6.51:3000"
	binaryName     = "ankerye-pmgt-server"
	serviceName    = "ankerye-pmgt"
	installPath    = "/opt/ankerye-pmgt/ankerye-pmgt-server"
)

type githubRelease struct {
	TagName string `json:"tag_name"`
	Body    string `json:"body"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

func CheckUpdate(c *gin.Context) {
	// Try GitHub first
	releaseURL := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", githubRepo)
	resp, err := http.Get(releaseURL)
	if err != nil || resp.StatusCode != 200 {
		// Try Gitea
		releaseURL = fmt.Sprintf("%s/api/v1/repos/%s/releases?limit=1", giteaBase, giteaRepo)
		resp, err = http.Get(releaseURL)
		if err != nil {
			util.OK(c, gin.H{
				"current_version": currentVersion,
				"latest_version":  currentVersion,
				"has_update":      false,
				"error":           "无法连接更新服务器",
			})
			return
		}
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var release githubRelease
	if err := json.Unmarshal(body, &release); err != nil {
		util.OK(c, gin.H{
			"current_version": currentVersion,
			"latest_version":  currentVersion,
			"has_update":      false,
		})
		return
	}

	hasUpdate := release.TagName != "" && release.TagName != "v"+currentVersion && release.TagName != currentVersion
	util.OK(c, gin.H{
		"current_version": currentVersion,
		"latest_version":  release.TagName,
		"has_update":      hasUpdate,
		"changelog":       release.Body,
	})
}

func ApplyUpdate(c *gin.Context) {
	// Check for update binary in same directory
	execPath, err := os.Executable()
	if err != nil {
		util.Fail(c, http.StatusInternalServerError, "获取执行路径失败")
		return
	}

	// Download latest binary
	arch := runtime.GOARCH
	osName := runtime.GOOS
	assetName := fmt.Sprintf("%s-%s-%s", binaryName, osName, arch)

	downloadURL := fmt.Sprintf("https://github.com/%s/releases/latest/download/%s", githubRepo, assetName)

	resp, err := http.Get(downloadURL)
	if err != nil || resp.StatusCode != 200 {
		util.Fail(c, http.StatusBadGateway, "下载更新失败，请手动更新")
		return
	}
	defer resp.Body.Close()

	tmpPath := execPath + ".new"
	f, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		util.Fail(c, http.StatusInternalServerError, "写入临时文件失败")
		return
	}
	io.Copy(f, resp.Body)
	f.Close()

	// Replace binary and restart service
	if err := os.Rename(tmpPath, installPath); err != nil {
		util.Fail(c, http.StatusInternalServerError, "替换二进制失败")
		return
	}

	util.OKMsg(c, "更新文件已下载，正在重启服务...")
	go func() {
		exec.Command("systemctl", "restart", serviceName).Run()
	}()
}
