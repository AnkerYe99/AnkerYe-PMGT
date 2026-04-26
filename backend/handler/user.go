package handler

import (
	"ankerye-pmgt/db"
	"ankerye-pmgt/util"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type userRow struct {
	ID          int64   `json:"id"`
	Username    string  `json:"username"`
	Email       *string `json:"email"`
	DisplayName *string `json:"display_name"`
	Role        string  `json:"role"`
	Status      string  `json:"status"`
	CreatedAt   string  `json:"created_at"`
	LastLoginAt *string `json:"last_login_at"`
}

func ListUsers(c *gin.Context) {
	rows, err := db.DB.Query(`SELECT id,username,email,display_name,role,status,created_at,last_login_at FROM users ORDER BY id`)
	if err != nil {
		util.Fail(c, http.StatusInternalServerError, "查询失败")
		return
	}
	defer rows.Close()
	var users []userRow
	for rows.Next() {
		var u userRow
		rows.Scan(&u.ID, &u.Username, &u.Email, &u.DisplayName, &u.Role, &u.Status, &u.CreatedAt, &u.LastLoginAt)
		users = append(users, u)
	}
	if users == nil {
		users = []userRow{}
	}
	util.OK(c, users)
}

type createUserReq struct {
	Username    string `json:"username" binding:"required"`
	Password    string `json:"password" binding:"required,min=6"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	Role        string `json:"role"`
}

func CreateUser(c *gin.Context) {
	var req createUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}
	if req.Role == "" {
		req.Role = "executor"
	}
	if req.Role != "admin" && req.Role != "developer" && req.Role != "executor" {
		util.Fail(c, http.StatusBadRequest, "无效角色")
		return
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	res, err := db.DB.Exec(`INSERT INTO users(username,password_hash,email,display_name,role,status) VALUES(?,?,?,?,?,?)`,
		req.Username, string(hash), req.Email, req.DisplayName, req.Role, "active")
	if err != nil {
		util.Fail(c, http.StatusConflict, "用户名已存在")
		return
	}
	id, _ := res.LastInsertId()
	util.OK(c, gin.H{"id": id})
}

type updateUserReq struct {
	Email       *string `json:"email"`
	DisplayName *string `json:"display_name"`
	Role        *string `json:"role"`
	Status      *string `json:"status"`
	Password    *string `json:"password"`
}

func UpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseInt(idStr, 10, 64)
	var req updateUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, http.StatusBadRequest, "参数错误")
		return
	}
	if req.Email != nil {
		db.DB.Exec("UPDATE users SET email=? WHERE id=?", *req.Email, id)
	}
	if req.DisplayName != nil {
		db.DB.Exec("UPDATE users SET display_name=? WHERE id=?", *req.DisplayName, id)
	}
	if req.Role != nil {
		db.DB.Exec("UPDATE users SET role=? WHERE id=?", *req.Role, id)
	}
	if req.Status != nil {
		db.DB.Exec("UPDATE users SET status=? WHERE id=?", *req.Status, id)
	}
	if req.Password != nil && len(*req.Password) >= 6 {
		hash, _ := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
		db.DB.Exec("UPDATE users SET password_hash=? WHERE id=?", string(hash), id)
	}
	util.OKMsg(c, "更新成功")
}

func DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseInt(idStr, 10, 64)
	cur := util.CurrentUser(c)
	if cur.ID == id {
		util.Fail(c, http.StatusBadRequest, "不能删除自己")
		return
	}
	db.DB.Exec("UPDATE users SET status='deleted' WHERE id=?", id)
	util.OKMsg(c, "已删除")
}
