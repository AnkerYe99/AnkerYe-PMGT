package handler

import (
	"ankerye-pmgt/db"
	"ankerye-pmgt/util"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

var backupEncKey = []byte("ankerye-pmgt-backup-aes-key-32by") // 32 bytes

func deriveKey(key []byte) []byte {
	h := sha256.Sum256(key)
	return h[:]
}

func GetSettings(c *gin.Context) {
	rows, err := db.DB.Query("SELECT k,v FROM system_settings")
	if err != nil {
		util.Fail(c, http.StatusInternalServerError, "查询失败")
		return
	}
	defer rows.Close()
	settings := map[string]string{}
	for rows.Next() {
		var k, v string
		rows.Scan(&k, &v)
		settings[k] = v
	}
	util.OK(c, settings)
}

func UpdateSettings(c *gin.Context) {
	var data map[string]string
	if err := c.ShouldBindJSON(&data); err != nil {
		util.Fail(c, http.StatusBadRequest, "参数错误")
		return
	}
	for k, v := range data {
		db.DB.Exec("INSERT OR REPLACE INTO system_settings(k,v) VALUES(?,?)", k, v)
	}
	util.OKMsg(c, "设置已保存")
}

type backupData struct {
	Version   string          `json:"version"`
	Timestamp string          `json:"timestamp"`
	Users     []map[string]interface{} `json:"users"`
	Projects  []map[string]interface{} `json:"projects"`
	Documents []map[string]interface{} `json:"documents"`
	Settings  map[string]string        `json:"settings"`
}

func BackupDB(c *gin.Context) {
	backup := backupData{
		Version:   "1.0.0",
		Timestamp: time.Now().Format(time.RFC3339),
		Settings:  map[string]string{},
	}

	// Export users
	rows, _ := db.DB.Query("SELECT id,username,password_hash,email,display_name,role,status,created_at FROM users")
	if rows != nil {
		defer rows.Close()
		cols, _ := rows.Columns()
		for rows.Next() {
			vals := make([]interface{}, len(cols))
			ptrs := make([]interface{}, len(cols))
			for i := range vals {
				ptrs[i] = &vals[i]
			}
			rows.Scan(ptrs...)
			m := map[string]interface{}{}
			for i, col := range cols {
				m[col] = vals[i]
			}
			backup.Users = append(backup.Users, m)
		}
	}

	// Export projects
	prows, _ := db.DB.Query("SELECT id,title,summary,type,status,is_private,tags,cover_color,created_by,created_at,updated_at,deleted_at FROM projects")
	if prows != nil {
		defer prows.Close()
		cols, _ := prows.Columns()
		for prows.Next() {
			vals := make([]interface{}, len(cols))
			ptrs := make([]interface{}, len(cols))
			for i := range vals {
				ptrs[i] = new(sql.NullString)
				ptrs[i] = &vals[i]
			}
			prows.Scan(ptrs...)
			m := map[string]interface{}{}
			for i, col := range cols {
				m[col] = vals[i]
			}
			backup.Projects = append(backup.Projects, m)
		}
	}

	// Export documents
	drows, _ := db.DB.Query("SELECT id,project_id,title,content,sort_order,status,created_by,created_at FROM documents")
	if drows != nil {
		defer drows.Close()
		cols, _ := drows.Columns()
		for drows.Next() {
			vals := make([]interface{}, len(cols))
			ptrs := make([]interface{}, len(cols))
			for i := range vals {
				ptrs[i] = &vals[i]
			}
			drows.Scan(ptrs...)
			m := map[string]interface{}{}
			for i, col := range cols {
				m[col] = vals[i]
			}
			backup.Documents = append(backup.Documents, m)
		}
	}

	// Export settings
	srows, _ := db.DB.Query("SELECT k,v FROM system_settings")
	if srows != nil {
		defer srows.Close()
		for srows.Next() {
			var k, v string
			srows.Scan(&k, &v)
			backup.Settings[k] = v
		}
	}

	plaintext, _ := json.Marshal(backup)

	// Encrypt
	key := deriveKey(backupEncKey)
	block, _ := aes.NewCipher(key)
	gcm, _ := cipher.NewGCM(block)
	nonce := make([]byte, gcm.NonceSize())
	io.ReadFull(rand.Reader, nonce)
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)

	fname := "ankerye-pmgt-backup-" + time.Now().Format("20060102-150405") + ".bak"
	c.Header("Content-Disposition", "attachment; filename="+fname)
	c.Header("Content-Type", "application/octet-stream")
	c.Data(http.StatusOK, "application/octet-stream", ciphertext)
}

func RestoreDB(c *gin.Context) {
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		util.Fail(c, http.StatusBadRequest, "文件获取失败")
		return
	}
	defer file.Close()

	ciphertext, err := io.ReadAll(file)
	if err != nil {
		util.Fail(c, http.StatusBadRequest, "读取失败")
		return
	}

	key := deriveKey(backupEncKey)
	block, err := aes.NewCipher(key)
	if err != nil {
		util.Fail(c, http.StatusInternalServerError, "解密初始化失败")
		return
	}
	gcm, _ := cipher.NewGCM(block)
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		util.Fail(c, http.StatusBadRequest, "备份文件损坏")
		return
	}
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		util.Fail(c, http.StatusBadRequest, "解密失败，备份文件可能损坏")
		return
	}

	var backup backupData
	if err := json.Unmarshal(plaintext, &backup); err != nil {
		util.Fail(c, http.StatusBadRequest, "备份格式错误")
		return
	}

	util.OK(c, gin.H{"message": "备份解析成功，恢复功能需手动操作数据库", "timestamp": backup.Timestamp, "version": backup.Version})
}
