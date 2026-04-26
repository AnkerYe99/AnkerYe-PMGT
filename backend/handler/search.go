package handler

import (
	"ankerye-pmgt/db"
	"ankerye-pmgt/util"

	"github.com/gin-gonic/gin"
)

type searchResult struct {
	Type      string `json:"type"`
	ID        int64  `json:"id"`
	ProjectID int64  `json:"project_id"`
	Title     string `json:"title"`
	Snippet   string `json:"snippet"`
}

func Search(c *gin.Context) {
	user := util.CurrentUser(c)
	q := c.Query("q")
	if q == "" {
		util.OK(c, []searchResult{})
		return
	}

	var results []searchResult

	// Search projects
	var projectQuery string
	var projectArgs []interface{}
	if user.Role == "admin" {
		projectQuery = `SELECT id, title, COALESCE(summary,'') FROM projects WHERE deleted_at IS NULL AND (title LIKE ? OR summary LIKE ?) ORDER BY updated_at DESC LIMIT 10`
		projectArgs = []interface{}{"%" + q + "%", "%" + q + "%"}
	} else {
		projectQuery = `SELECT id, title, COALESCE(summary,'') FROM projects WHERE deleted_at IS NULL
			AND (is_private=0 OR created_by=? OR EXISTS(SELECT 1 FROM project_members pm WHERE pm.project_id=projects.id AND pm.user_id=?))
			AND (title LIKE ? OR summary LIKE ?) ORDER BY updated_at DESC LIMIT 10`
		projectArgs = []interface{}{user.ID, user.ID, "%" + q + "%", "%" + q + "%"}
	}

	rows, err := db.DB.Query(projectQuery, projectArgs...)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var r searchResult
			r.Type = "project"
			rows.Scan(&r.ID, &r.Title, &r.Snippet)
			r.ProjectID = r.ID
			results = append(results, r)
		}
	}

	// Search documents via FTS
	ftsRows, err := db.DB.Query(`SELECT doc_id, project_id, title, snippet(fts_docs, 1, '<mark>', '</mark>', '...', 20)
		FROM fts_docs WHERE fts_docs MATCH ? ORDER BY rank LIMIT 20`, q)
	if err == nil {
		defer ftsRows.Close()
		for ftsRows.Next() {
			var r searchResult
			r.Type = "document"
			ftsRows.Scan(&r.ID, &r.ProjectID, &r.Title, &r.Snippet)
			// Check visibility
			if canViewProject(user.ID, user.Role, r.ProjectID) {
				results = append(results, r)
			}
		}
	}

	if results == nil {
		results = []searchResult{}
	}
	util.OK(c, gin.H{
		"query":   q,
		"results": results,
		"total":   len(results),
	})
}

func GetVersion(c *gin.Context) {
	util.OK(c, gin.H{
		"version":    "1.0.0",
		"build_time": "2026-04-26",
		"go_version": "1.22",
	})
}
