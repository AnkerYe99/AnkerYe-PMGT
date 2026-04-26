package handler

import (
	"ankerye-pmgt/db"
	"ankerye-pmgt/util"
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type projectRow struct {
	ID          int64   `json:"id"`
	Title       string  `json:"title"`
	Summary     *string `json:"summary"`
	Type        string  `json:"type"`
	Status      string  `json:"status"`
	IsPrivate   int     `json:"is_private"`
	Tags        string  `json:"tags"`
	CoverColor  string  `json:"cover_color"`
	CreatedBy   int64   `json:"created_by"`
	CreatorName *string `json:"creator_name"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
	DeletedAt   *string `json:"deleted_at,omitempty"`
	MemberCount int     `json:"member_count"`
}

func canViewProject(userID int64, role string, projectID int64) bool {
	var isPrivate int
	var createdBy int64
	err := db.DB.QueryRow("SELECT is_private, created_by FROM projects WHERE id=? AND deleted_at IS NULL", projectID).Scan(&isPrivate, &createdBy)
	if err != nil {
		return false
	}
	if isPrivate == 0 || role == "admin" || createdBy == userID {
		return true
	}
	var cnt int
	db.DB.QueryRow("SELECT COUNT(*) FROM project_members WHERE project_id=? AND user_id=?", projectID, userID).Scan(&cnt)
	return cnt > 0
}

func ListProjects(c *gin.Context) {
	user := util.CurrentUser(c)
	statusFilter := c.Query("status")
	typeFilter := c.Query("type")
	q := c.Query("q")

	query := `SELECT p.id,p.title,p.summary,p.type,p.status,p.is_private,p.tags,p.cover_color,
		p.created_by, u.display_name, p.created_at,p.updated_at,p.deleted_at,
		(SELECT COUNT(*) FROM project_members pm WHERE pm.project_id=p.id) as member_count
		FROM projects p LEFT JOIN users u ON u.id=p.created_by
		WHERE p.deleted_at IS NULL`
	args := []interface{}{}

	if user.Role != "admin" {
		query += ` AND (p.is_private=0 OR p.created_by=? OR EXISTS(SELECT 1 FROM project_members pm WHERE pm.project_id=p.id AND pm.user_id=?))`
		args = append(args, user.ID, user.ID)
	}
	if statusFilter != "" {
		query += " AND p.status=?"
		args = append(args, statusFilter)
	}
	if typeFilter != "" {
		query += " AND p.type=?"
		args = append(args, typeFilter)
	}
	if q != "" {
		query += " AND (p.title LIKE ? OR p.summary LIKE ?)"
		args = append(args, "%"+q+"%", "%"+q+"%")
	}
	query += " ORDER BY p.updated_at DESC"

	rows, err := db.DB.Query(query, args...)
	if err != nil {
		util.Fail(c, http.StatusInternalServerError, "查询失败")
		return
	}
	defer rows.Close()
	var projects []projectRow
	for rows.Next() {
		var p projectRow
		rows.Scan(&p.ID, &p.Title, &p.Summary, &p.Type, &p.Status, &p.IsPrivate, &p.Tags, &p.CoverColor,
			&p.CreatedBy, &p.CreatorName, &p.CreatedAt, &p.UpdatedAt, &p.DeletedAt, &p.MemberCount)
		projects = append(projects, p)
	}
	if projects == nil {
		projects = []projectRow{}
	}
	util.OK(c, projects)
}

type createProjectReq struct {
	Title      string `json:"title" binding:"required"`
	Summary    string `json:"summary"`
	Type       string `json:"type"`
	Status     string `json:"status"`
	Tags       string `json:"tags"`
	CoverColor string `json:"cover_color"`
}

func CreateProject(c *gin.Context) {
	user := util.CurrentUser(c)
	var req createProjectReq
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, http.StatusBadRequest, "参数错误")
		return
	}
	if req.Type == "" {
		req.Type = "general"
	}
	if req.Status == "" {
		req.Status = "draft"
	}
	if req.Tags == "" {
		req.Tags = "[]"
	}
	if req.CoverColor == "" {
		req.CoverColor = "#6366F1"
	}
	res, err := db.DB.Exec(`INSERT INTO projects(title,summary,type,status,tags,cover_color,created_by) VALUES(?,?,?,?,?,?,?)`,
		req.Title, req.Summary, req.Type, req.Status, req.Tags, req.CoverColor, user.ID)
	if err != nil {
		util.Fail(c, http.StatusInternalServerError, "创建失败")
		return
	}
	id, _ := res.LastInsertId()
	util.OK(c, gin.H{"id": id})
}

func GetProject(c *gin.Context) {
	user := util.CurrentUser(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if !canViewProject(user.ID, user.Role, id) {
		util.Fail(c, http.StatusForbidden, "无权访问")
		return
	}
	var p projectRow
	err := db.DB.QueryRow(`SELECT p.id,p.title,p.summary,p.type,p.status,p.is_private,p.tags,p.cover_color,
		p.created_by, u.display_name, p.created_at,p.updated_at,p.deleted_at,
		(SELECT COUNT(*) FROM project_members pm WHERE pm.project_id=p.id) as member_count
		FROM projects p LEFT JOIN users u ON u.id=p.created_by WHERE p.id=?`, id).
		Scan(&p.ID, &p.Title, &p.Summary, &p.Type, &p.Status, &p.IsPrivate, &p.Tags, &p.CoverColor,
			&p.CreatedBy, &p.CreatorName, &p.CreatedAt, &p.UpdatedAt, &p.DeletedAt, &p.MemberCount)
	if err != nil {
		util.Fail(c, http.StatusNotFound, "项目不存在")
		return
	}
	util.OK(c, p)
}

type updateProjectReq struct {
	Title      *string `json:"title"`
	Summary    *string `json:"summary"`
	Type       *string `json:"type"`
	Status     *string `json:"status"`
	Tags       *string `json:"tags"`
	CoverColor *string `json:"cover_color"`
}

func UpdateProject(c *gin.Context) {
	user := util.CurrentUser(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var createdBy int64
	db.DB.QueryRow("SELECT created_by FROM projects WHERE id=? AND deleted_at IS NULL", id).Scan(&createdBy)
	if user.Role != "admin" && createdBy != user.ID {
		util.Fail(c, http.StatusForbidden, "只有创建者或管理员可编辑")
		return
	}

	var req updateProjectReq
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, http.StatusBadRequest, "参数错误")
		return
	}
	if req.Title != nil {
		db.DB.Exec("UPDATE projects SET title=?,updated_at=CURRENT_TIMESTAMP WHERE id=?", *req.Title, id)
	}
	if req.Summary != nil {
		db.DB.Exec("UPDATE projects SET summary=?,updated_at=CURRENT_TIMESTAMP WHERE id=?", *req.Summary, id)
	}
	if req.Type != nil {
		db.DB.Exec("UPDATE projects SET type=?,updated_at=CURRENT_TIMESTAMP WHERE id=?", *req.Type, id)
	}
	if req.Status != nil {
		db.DB.Exec("UPDATE projects SET status=?,updated_at=CURRENT_TIMESTAMP WHERE id=?", *req.Status, id)
	}
	if req.Tags != nil {
		db.DB.Exec("UPDATE projects SET tags=?,updated_at=CURRENT_TIMESTAMP WHERE id=?", *req.Tags, id)
	}
	if req.CoverColor != nil {
		db.DB.Exec("UPDATE projects SET cover_color=?,updated_at=CURRENT_TIMESTAMP WHERE id=?", *req.CoverColor, id)
	}
	util.OKMsg(c, "更新成功")
}

func DeleteProject(c *gin.Context) {
	user := util.CurrentUser(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var createdBy int64
	db.DB.QueryRow("SELECT created_by FROM projects WHERE id=? AND deleted_at IS NULL", id).Scan(&createdBy)
	if user.Role != "admin" && createdBy != user.ID {
		util.Fail(c, http.StatusForbidden, "只有创建者或管理员可删除")
		return
	}
	db.DB.Exec("UPDATE projects SET deleted_at=CURRENT_TIMESTAMP WHERE id=?", id)
	util.OKMsg(c, "已删除")
}

func RestoreProject(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	db.DB.Exec("UPDATE projects SET deleted_at=NULL WHERE id=?", id)
	util.OKMsg(c, "已恢复")
}

// --- Members ---

type memberRow struct {
	ID          int64   `json:"id"`
	ProjectID   int64   `json:"project_id"`
	UserID      int64   `json:"user_id"`
	Username    string  `json:"username"`
	DisplayName *string `json:"display_name"`
	MemberRole  string  `json:"member_role"`
	AssignedAt  string  `json:"assigned_at"`
}

func ListMembers(c *gin.Context) {
	user := util.CurrentUser(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if !canViewProject(user.ID, user.Role, id) {
		util.Fail(c, http.StatusForbidden, "无权访问")
		return
	}
	rows, err := db.DB.Query(`SELECT pm.id,pm.project_id,pm.user_id,u.username,u.display_name,pm.member_role,pm.assigned_at
		FROM project_members pm JOIN users u ON u.id=pm.user_id WHERE pm.project_id=? ORDER BY pm.assigned_at`, id)
	if err != nil {
		util.Fail(c, http.StatusInternalServerError, "查询失败")
		return
	}
	defer rows.Close()
	var members []memberRow
	for rows.Next() {
		var m memberRow
		rows.Scan(&m.ID, &m.ProjectID, &m.UserID, &m.Username, &m.DisplayName, &m.MemberRole, &m.AssignedAt)
		members = append(members, m)
	}
	if members == nil {
		members = []memberRow{}
	}
	util.OK(c, members)
}

type addMemberReq struct {
	UserID     int64  `json:"user_id" binding:"required"`
	MemberRole string `json:"member_role"`
}

func AddMember(c *gin.Context) {
	user := util.CurrentUser(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var createdBy int64
	db.DB.QueryRow("SELECT created_by FROM projects WHERE id=?", id).Scan(&createdBy)
	if user.Role != "admin" && createdBy != user.ID {
		util.Fail(c, http.StatusForbidden, "权限不足")
		return
	}

	var req addMemberReq
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, http.StatusBadRequest, "参数错误")
		return
	}
	if req.MemberRole == "" {
		req.MemberRole = "executor"
	}
	_, err := db.DB.Exec(`INSERT OR REPLACE INTO project_members(project_id,user_id,member_role,assigned_by) VALUES(?,?,?,?)`,
		id, req.UserID, req.MemberRole, user.ID)
	if err != nil {
		util.Fail(c, http.StatusInternalServerError, "添加失败")
		return
	}
	// Make project private when members added
	db.DB.Exec("UPDATE projects SET is_private=1,updated_at=CURRENT_TIMESTAMP WHERE id=?", id)
	util.OKMsg(c, "已添加")
}

func RemoveMember(c *gin.Context) {
	user := util.CurrentUser(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	uid, _ := strconv.ParseInt(c.Param("uid"), 10, 64)

	var createdBy int64
	db.DB.QueryRow("SELECT created_by FROM projects WHERE id=?", id).Scan(&createdBy)
	if user.Role != "admin" && createdBy != user.ID {
		util.Fail(c, http.StatusForbidden, "权限不足")
		return
	}

	db.DB.Exec("DELETE FROM project_members WHERE project_id=? AND user_id=?", id, uid)

	// If no more members, make project public
	var cnt int
	db.DB.QueryRow("SELECT COUNT(*) FROM project_members WHERE project_id=?", id).Scan(&cnt)
	if cnt == 0 {
		db.DB.Exec("UPDATE projects SET is_private=0,updated_at=CURRENT_TIMESTAMP WHERE id=?", id)
	}
	util.OKMsg(c, "已移除")
}

// GetDeletedProjects returns soft-deleted projects (admin only)
func GetDeletedProjects(c *gin.Context) {
	rows, err := db.DB.Query(`SELECT p.id,p.title,p.summary,p.type,p.status,p.is_private,p.tags,p.cover_color,
		p.created_by, u.display_name, p.created_at,p.updated_at,p.deleted_at,
		(SELECT COUNT(*) FROM project_members pm WHERE pm.project_id=p.id) as member_count
		FROM projects p LEFT JOIN users u ON u.id=p.created_by
		WHERE p.deleted_at IS NOT NULL ORDER BY p.deleted_at DESC`)
	if err != nil {
		util.Fail(c, http.StatusInternalServerError, "查询失败")
		return
	}
	defer rows.Close()
	var projects []projectRow
	for rows.Next() {
		var p projectRow
		var deletedAt sql.NullString
		rows.Scan(&p.ID, &p.Title, &p.Summary, &p.Type, &p.Status, &p.IsPrivate, &p.Tags, &p.CoverColor,
			&p.CreatedBy, &p.CreatorName, &p.CreatedAt, &p.UpdatedAt, &deletedAt, &p.MemberCount)
		if deletedAt.Valid {
			p.DeletedAt = &deletedAt.String
		}
		projects = append(projects, p)
	}
	if projects == nil {
		projects = []projectRow{}
	}
	util.OK(c, projects)
}

// UpdateProjectStatus POST /api/v1/projects/:id/status
// 三种用户均可将项目置为 active(进行中) 或 completed(已完成)
// 已删除的项目不可操作
func UpdateProjectStatus(c *gin.Context) {
	user := util.CurrentUser(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	if !canViewProject(user.ID, user.Role, id) {
		util.Fail(c, http.StatusForbidden, "无权访问此项目")
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, 400, "参数错误")
		return
	}

	// 只允许设置这两个状态（其他状态如 draft/archived/deleted 需要更高权限的接口）
	if req.Status != "active" && req.Status != "completed" {
		util.Fail(c, 400, "只能设置为：active(进行中) 或 completed(已完成)")
		return
	}

	// 已软删除的项目不可操作
	var deletedAt sql.NullString
	db.DB.QueryRow("SELECT deleted_at FROM projects WHERE id=?", id).Scan(&deletedAt)
	if deletedAt.Valid {
		util.Fail(c, 400, "已删除的项目无法修改状态")
		return
	}

	db.DB.Exec("UPDATE projects SET status=?, updated_at=CURRENT_TIMESTAMP WHERE id=?", req.Status, id)
	util.OKMsg(c, "状态已更新")
}
