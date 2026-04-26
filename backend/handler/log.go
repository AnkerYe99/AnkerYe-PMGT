package handler

import (
	"ankerye-pmgt/db"
	"ankerye-pmgt/util"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ListProjectLogs GET /api/v1/projects/:id/logs
func ListProjectLogs(c *gin.Context) {
	pid, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		util.Fail(c, 400, "参数错误")
		return
	}

	rows, err := db.DB.Query(`
		SELECT l.id, l.content, l.client_ip, l.created_at,
		       u.username, COALESCE(u.display_name, u.username)
		FROM project_logs l
		JOIN users u ON u.id = l.created_by
		WHERE l.project_id = ?
		ORDER BY l.created_at DESC
	`, pid)
	if err != nil {
		util.Fail(c, 500, err.Error())
		return
	}
	defer rows.Close()

	type LogItem struct {
		ID          int64  `json:"id"`
		Content     string `json:"content"`
		ClientIP    string `json:"client_ip"`
		CreatedAt   string `json:"created_at"`
		Username    string `json:"username"`
		DisplayName string `json:"display_name"`
	}

	list := []LogItem{}
	for rows.Next() {
		var item LogItem
		rows.Scan(&item.ID, &item.Content, &item.ClientIP, &item.CreatedAt, &item.Username, &item.DisplayName)
		list = append(list, item)
	}
	util.OK(c, list)
}

// AddProjectLog POST /api/v1/projects/:id/logs
// 三种用户都可以写进度日志
func AddProjectLog(c *gin.Context) {
	pid, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		util.Fail(c, 400, "参数错误")
		return
	}
	user := util.CurrentUser(c)

	var req struct {
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, 400, "内容不能为空")
		return
	}

	clientIP := c.ClientIP()

	_, err = db.DB.Exec(`
		INSERT INTO project_logs(project_id, content, created_by, client_ip)
		VALUES(?, ?, ?, ?)
	`, pid, req.Content, user.ID, clientIP)
	if err != nil {
		util.Fail(c, 500, err.Error())
		return
	}

	// 自动更新项目 updated_at
	db.DB.Exec(`UPDATE projects SET updated_at=CURRENT_TIMESTAMP WHERE id=?`, pid)

	util.OK(c, nil)
}

// DeleteProjectLog DELETE /api/v1/projects/:id/logs/:lid
// 只有日志创建者本人或管理员可删除
func DeleteProjectLog(c *gin.Context) {
	pid, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	lid, err := strconv.ParseInt(c.Param("lid"), 10, 64)
	if err != nil {
		util.Fail(c, 400, "参数错误")
		return
	}
	user := util.CurrentUser(c)

	var createdBy int64
	err = db.DB.QueryRow(`SELECT created_by FROM project_logs WHERE id=? AND project_id=?`, lid, pid).Scan(&createdBy)
	if err != nil {
		util.Fail(c, 404, "日志不存在")
		return
	}

	if user.Role != "admin" && user.ID != createdBy {
		util.Fail(c, 403, "无权删除他人的日志")
		return
	}

	db.DB.Exec(`DELETE FROM project_logs WHERE id=?`, lid)
	util.OK(c, nil)
}
