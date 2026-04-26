package handler

import (
	"ankerye-pmgt/config"
	"ankerye-pmgt/db"
	"ankerye-pmgt/util"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func UploadFile(c *gin.Context) {
	user := util.CurrentUser(c)
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		util.Fail(c, http.StatusBadRequest, "文件获取失败")
		return
	}
	defer file.Close()

	now := time.Now()
	dateDir := fmt.Sprintf("%d/%02d/%02d", now.Year(), now.Month(), now.Day())
	uploadDir := filepath.Join(config.Cfg.Upload.Dir, dateDir)
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		util.Fail(c, http.StatusInternalServerError, "目录创建失败")
		return
	}

	ext := filepath.Ext(header.Filename)
	storedName := uuid.New().String() + ext
	savePath := filepath.Join(uploadDir, storedName)

	if err := c.SaveUploadedFile(header, savePath); err != nil {
		util.Fail(c, http.StatusInternalServerError, "文件保存失败")
		return
	}

	projectIDStr := c.PostForm("project_id")
	var projectID *int64
	if projectIDStr != "" {
		pid, _ := strconv.ParseInt(projectIDStr, 10, 64)
		projectID = &pid
	}

	mime := header.Header.Get("Content-Type")
	if mime == "" {
		mime = "application/octet-stream"
	}

	filePath := "/" + dateDir + "/" + storedName

	var res interface{}
	if projectID != nil {
		r, err := db.DB.Exec(`INSERT INTO files(project_id,original_name,stored_name,file_path,file_size,mime_type,uploaded_by) VALUES(?,?,?,?,?,?,?)`,
			*projectID, header.Filename, storedName, filePath, header.Size, mime, user.ID)
		if err == nil {
			id, _ := r.LastInsertId()
			res = gin.H{"id": id, "url": "/api/v1/files/" + strconv.FormatInt(id, 10), "name": header.Filename}
		}
	} else {
		r, err := db.DB.Exec(`INSERT INTO files(original_name,stored_name,file_path,file_size,mime_type,uploaded_by) VALUES(?,?,?,?,?,?)`,
			header.Filename, storedName, filePath, header.Size, mime, user.ID)
		if err == nil {
			id, _ := r.LastInsertId()
			res = gin.H{"id": id, "url": "/api/v1/files/" + strconv.FormatInt(id, 10), "name": header.Filename}
		}
	}

	util.OK(c, res)
}

func ServeFile(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseInt(idStr, 10, 64)

	var filePath, originalName, mimeType string
	err := db.DB.QueryRow("SELECT file_path, original_name, mime_type FROM files WHERE id=?", id).
		Scan(&filePath, &originalName, &mimeType)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "文件不存在"})
		return
	}

	fullPath := filepath.Join(config.Cfg.Upload.Dir, strings.TrimPrefix(filePath, "/"))
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"message": "文件已删除"})
		return
	}

	c.Header("Content-Disposition", "inline; filename="+originalName)
	c.File(fullPath)
}
