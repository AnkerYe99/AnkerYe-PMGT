package handler

import (
	"ankerye-pmgt/db"
	"ankerye-pmgt/util"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type apiKeyRow struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	KeyPrefix   string  `json:"key_prefix"`
	Permissions string  `json:"permissions"`
	CreatedBy   int64   `json:"created_by"`
	ExpiresAt   *string `json:"expires_at"`
	LastUsedAt  *string `json:"last_used_at"`
	Status      string  `json:"status"`
	CreatedAt   string  `json:"created_at"`
}

func ListAPIKeys(c *gin.Context) {
	rows, err := db.DB.Query(`SELECT id,name,key_prefix,permissions,created_by,expires_at,last_used_at,status,created_at FROM api_keys ORDER BY id DESC`)
	if err != nil {
		util.Fail(c, http.StatusInternalServerError, "查询失败")
		return
	}
	defer rows.Close()
	var keys []apiKeyRow
	for rows.Next() {
		var k apiKeyRow
		rows.Scan(&k.ID, &k.Name, &k.KeyPrefix, &k.Permissions, &k.CreatedBy, &k.ExpiresAt, &k.LastUsedAt, &k.Status, &k.CreatedAt)
		keys = append(keys, k)
	}
	if keys == nil {
		keys = []apiKeyRow{}
	}
	util.OK(c, keys)
}

type createAPIKeyReq struct {
	Name        string `json:"name" binding:"required"`
	Permissions string `json:"permissions"`
	ExpiresAt   string `json:"expires_at"`
}

func CreateAPIKey(c *gin.Context) {
	user := util.CurrentUser(c)
	var req createAPIKeyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, http.StatusBadRequest, "参数错误")
		return
	}
	if req.Permissions == "" {
		req.Permissions = "read"
	}

	// Generate key: ak_<32 random hex>
	buf := make([]byte, 16)
	rand.Read(buf)
	rawKey := "ak_" + hex.EncodeToString(buf)
	keyPrefix := rawKey[:8]
	keyHash := fmt.Sprintf("%x", sha256.Sum256([]byte(rawKey)))

	var expiresAt interface{}
	if req.ExpiresAt != "" {
		expiresAt = req.ExpiresAt
	}

	res, err := db.DB.Exec(`INSERT INTO api_keys(name,key_hash,key_prefix,permissions,created_by,expires_at) VALUES(?,?,?,?,?,?)`,
		req.Name, keyHash, keyPrefix, req.Permissions, user.ID, expiresAt)
	if err != nil {
		util.Fail(c, http.StatusInternalServerError, "创建失败")
		return
	}
	id, _ := res.LastInsertId()
	// Return full key only once
	util.OK(c, gin.H{
		"id":          id,
		"key":         rawKey,
		"key_prefix":  keyPrefix,
		"name":        req.Name,
		"permissions": req.Permissions,
	})
}

func DeleteAPIKey(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseInt(idStr, 10, 64)
	db.DB.Exec("UPDATE api_keys SET status='revoked' WHERE id=?", id)
	util.OKMsg(c, "已撤销")
}
