package handler

import (
	"ankerye-pmgt/db"
	"ankerye-pmgt/middleware"
	"ankerye-pmgt/util"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type loginReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func Login(c *gin.Context) {
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, http.StatusBadRequest, "参数错误")
		return
	}

	var id int64
	var hash, role, displayName string
	var status string
	err := db.DB.QueryRow(`SELECT id, password_hash, role, display_name, status FROM users WHERE username=?`, req.Username).
		Scan(&id, &hash, &role, &displayName, &status)
	if err != nil {
		util.Fail(c, http.StatusUnauthorized, "用户名或密码错误")
		return
	}
	if status != "active" {
		util.Fail(c, http.StatusForbidden, "账号已被禁用")
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)); err != nil {
		util.Fail(c, http.StatusUnauthorized, "用户名或密码错误")
		return
	}

	claims := jwt.MapClaims{
		"uid":      id,
		"username": req.Username,
		"role":     role,
		"exp":      time.Now().Add(72 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(middleware.GetJWTSecret())
	if err != nil {
		util.Fail(c, http.StatusInternalServerError, "Token 生成失败")
		return
	}

	db.DB.Exec("UPDATE users SET last_login_at=CURRENT_TIMESTAMP WHERE id=?", id)

	util.OK(c, gin.H{
		"token":        tokenStr,
		"id":           id,
		"username":     req.Username,
		"display_name": displayName,
		"role":         role,
	})
}

func GetProfile(c *gin.Context) {
	user := util.CurrentUser(c)
	row := db.DB.QueryRow(`SELECT id,username,email,display_name,role,status,created_at,last_login_at FROM users WHERE id=?`, user.ID)
	var u struct {
		ID          int64   `json:"id"`
		Username    string  `json:"username"`
		Email       *string `json:"email"`
		DisplayName *string `json:"display_name"`
		Role        string  `json:"role"`
		Status      string  `json:"status"`
		CreatedAt   string  `json:"created_at"`
		LastLoginAt *string `json:"last_login_at"`
	}
	if err := row.Scan(&u.ID, &u.Username, &u.Email, &u.DisplayName, &u.Role, &u.Status, &u.CreatedAt, &u.LastLoginAt); err != nil {
		util.Fail(c, http.StatusInternalServerError, "查询失败")
		return
	}
	util.OK(c, u)
}

type changePwdReq struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

func ChangePassword(c *gin.Context) {
	user := util.CurrentUser(c)
	var req changePwdReq
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, http.StatusBadRequest, "参数错误")
		return
	}
	var hash string
	db.DB.QueryRow("SELECT password_hash FROM users WHERE id=?", user.ID).Scan(&hash)
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.OldPassword)); err != nil {
		util.Fail(c, http.StatusBadRequest, "原密码错误")
		return
	}
	newHash, _ := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	db.DB.Exec("UPDATE users SET password_hash=? WHERE id=?", string(newHash), user.ID)
	util.OKMsg(c, "密码修改成功")
}
