package handler

import (
	"ankerye-pmgt/db"
	"ankerye-pmgt/util"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type docRow struct {
	ID         int64   `json:"id"`
	ProjectID  int64   `json:"project_id"`
	Title      string  `json:"title"`
	Content    string  `json:"content"`
	SortOrder  int     `json:"sort_order"`
	Status     string  `json:"status"`
	CreatedBy  int64   `json:"created_by"`
	UpdatedBy  *int64  `json:"updated_by"`
	CreatedAt  string  `json:"created_at"`
	UpdatedAt  string  `json:"updated_at"`
	AuthorName *string `json:"author_name"`
}

func ListDocs(c *gin.Context) {
	user := util.CurrentUser(c)
	pid, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if !canViewProject(user.ID, user.Role, pid) {
		util.Fail(c, http.StatusForbidden, "无权访问")
		return
	}
	rows, err := db.DB.Query(`SELECT d.id,d.project_id,d.title,d.content,d.sort_order,d.status,
		d.created_by,d.updated_by,d.created_at,d.updated_at,u.display_name
		FROM documents d LEFT JOIN users u ON u.id=d.created_by
		WHERE d.project_id=? AND d.status='active' ORDER BY d.sort_order,d.id`, pid)
	if err != nil {
		util.Fail(c, http.StatusInternalServerError, "查询失败")
		return
	}
	defer rows.Close()
	var docs []docRow
	for rows.Next() {
		var d docRow
		rows.Scan(&d.ID, &d.ProjectID, &d.Title, &d.Content, &d.SortOrder, &d.Status,
			&d.CreatedBy, &d.UpdatedBy, &d.CreatedAt, &d.UpdatedAt, &d.AuthorName)
		docs = append(docs, d)
	}
	if docs == nil {
		docs = []docRow{}
	}
	util.OK(c, docs)
}

type createDocReq struct {
	Title     string `json:"title" binding:"required"`
	Content   string `json:"content"`
	SortOrder int    `json:"sort_order"`
}

func CreateDoc(c *gin.Context) {
	user := util.CurrentUser(c)
	pid, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req createDocReq
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, http.StatusBadRequest, "参数错误")
		return
	}

	res, err := db.DB.Exec(`INSERT INTO documents(project_id,title,content,sort_order,created_by,updated_by) VALUES(?,?,?,?,?,?)`,
		pid, req.Title, req.Content, req.SortOrder, user.ID, user.ID)
	if err != nil {
		util.Fail(c, http.StatusInternalServerError, "创建失败")
		return
	}
	docID, _ := res.LastInsertId()

	// Update FTS
	db.DB.Exec(`INSERT INTO fts_docs(title,content,project_id,doc_id) VALUES(?,?,?,?)`,
		req.Title, req.Content, pid, docID)

	// Update project updated_at
	db.DB.Exec("UPDATE projects SET updated_at=CURRENT_TIMESTAMP WHERE id=?", pid)

	util.OK(c, gin.H{"id": docID})
}

func GetDoc(c *gin.Context) {
	user := util.CurrentUser(c)
	pid, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	did, _ := strconv.ParseInt(c.Param("did"), 10, 64)
	if !canViewProject(user.ID, user.Role, pid) {
		util.Fail(c, http.StatusForbidden, "无权访问")
		return
	}
	var d docRow
	err := db.DB.QueryRow(`SELECT d.id,d.project_id,d.title,d.content,d.sort_order,d.status,
		d.created_by,d.updated_by,d.created_at,d.updated_at,u.display_name
		FROM documents d LEFT JOIN users u ON u.id=d.created_by
		WHERE d.id=? AND d.project_id=?`, did, pid).
		Scan(&d.ID, &d.ProjectID, &d.Title, &d.Content, &d.SortOrder, &d.Status,
			&d.CreatedBy, &d.UpdatedBy, &d.CreatedAt, &d.UpdatedAt, &d.AuthorName)
	if err != nil {
		util.Fail(c, http.StatusNotFound, "文档不存在")
		return
	}
	util.OK(c, d)
}

type updateDocReq struct {
	Title     *string `json:"title"`
	Content   *string `json:"content"`
	SortOrder *int    `json:"sort_order"`
}

func UpdateDoc(c *gin.Context) {
	user := util.CurrentUser(c)
	pid, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	did, _ := strconv.ParseInt(c.Param("did"), 10, 64)

	var req updateDocReq
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, http.StatusBadRequest, "参数错误")
		return
	}
	if req.Title != nil {
		db.DB.Exec("UPDATE documents SET title=?,updated_by=?,updated_at=CURRENT_TIMESTAMP WHERE id=? AND project_id=?",
			*req.Title, user.ID, did, pid)
	}
	if req.Content != nil {
		db.DB.Exec("UPDATE documents SET content=?,updated_by=?,updated_at=CURRENT_TIMESTAMP WHERE id=? AND project_id=?",
			*req.Content, user.ID, did, pid)
	}
	if req.SortOrder != nil {
		db.DB.Exec("UPDATE documents SET sort_order=?,updated_at=CURRENT_TIMESTAMP WHERE id=? AND project_id=?",
			*req.SortOrder, did, pid)
	}
	// Update FTS
	var title, content string
	db.DB.QueryRow("SELECT title,content FROM documents WHERE id=?", did).Scan(&title, &content)
	db.DB.Exec("DELETE FROM fts_docs WHERE doc_id=?", did)
	db.DB.Exec(`INSERT INTO fts_docs(title,content,project_id,doc_id) VALUES(?,?,?,?)`, title, content, pid, did)

	db.DB.Exec("UPDATE projects SET updated_at=CURRENT_TIMESTAMP WHERE id=?", pid)
	util.OKMsg(c, "更新成功")
}

func DeleteDoc(c *gin.Context) {
	pid, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	did, _ := strconv.ParseInt(c.Param("did"), 10, 64)
	db.DB.Exec("UPDATE documents SET status='deleted' WHERE id=? AND project_id=?", did, pid)
	db.DB.Exec("DELETE FROM fts_docs WHERE doc_id=?", did)
	util.OKMsg(c, "已删除")
}
