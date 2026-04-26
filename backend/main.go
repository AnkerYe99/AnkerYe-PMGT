package main

import (
	"ankerye-pmgt/config"
	"ankerye-pmgt/db"
	"ankerye-pmgt/handler"
	"ankerye-pmgt/middleware"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

//go:embed all:frontend/dist
var frontendFS embed.FS

func main() {
	// Load config
	if err := config.Load("/opt/ankerye-pmgt/config.yaml"); err != nil {
		log.Printf("Config load warning: %v", err)
	}

	// Init DB
	if err := db.Init(config.Cfg.DB.Path); err != nil {
		log.Fatalf("DB init failed: %v", err)
	}

	middleware.SetJWTSecret(config.Cfg.JWT.Secret)

	r := gin.Default()

	// CORS
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type,Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	api := r.Group("/api/v1")
	{
		// Public
		api.POST("/auth/login", handler.Login)
		api.GET("/version", handler.GetVersion)
		api.GET("/files/:id", handler.ServeFile)

		// JWT protected
		auth := api.Group("")
		auth.Use(middleware.JWTAuth())
		{
			auth.GET("/auth/profile", handler.GetProfile)
			auth.PUT("/auth/password", handler.ChangePassword)

			auth.GET("/projects", handler.ListProjects)
			auth.GET("/projects/:id", handler.GetProject)
			auth.GET("/projects/:id/members", handler.ListMembers)
			auth.GET("/projects/:id/docs", handler.ListDocs)
			auth.GET("/projects/:id/docs/:did", handler.GetDoc)

			auth.GET("/search", handler.Search)
		}

		// Developer or admin
		dev := api.Group("")
		dev.Use(middleware.JWTAuth(), middleware.RequireDevOrAbove())
		{
			dev.POST("/projects", handler.CreateProject)
			dev.PUT("/projects/:id", handler.UpdateProject)
			dev.DELETE("/projects/:id", handler.DeleteProject)
			dev.POST("/projects/:id/members", handler.AddMember)
			dev.DELETE("/projects/:id/members/:uid", handler.RemoveMember)

			dev.POST("/projects/:id/docs", handler.CreateDoc)
			dev.PUT("/projects/:id/docs/:did", handler.UpdateDoc)
			dev.DELETE("/projects/:id/docs/:did", handler.DeleteDoc)

			dev.POST("/upload", handler.UploadFile)
		}

		// Admin only
		admin := api.Group("")
		admin.Use(middleware.JWTAuth(), middleware.RequireRole("admin"))
		{
			admin.GET("/users", handler.ListUsers)
			admin.POST("/users", handler.CreateUser)
			admin.PUT("/users/:id", handler.UpdateUser)
			admin.DELETE("/users/:id", handler.DeleteUser)

			admin.POST("/projects/:id/restore", handler.RestoreProject)
			admin.GET("/projects/deleted", handler.GetDeletedProjects)

			admin.GET("/settings", handler.GetSettings)
			admin.PUT("/settings", handler.UpdateSettings)
			admin.GET("/settings/backup", handler.BackupDB)
			admin.POST("/settings/restore", handler.RestoreDB)

			admin.GET("/apikeys", handler.ListAPIKeys)
			admin.POST("/apikeys", handler.CreateAPIKey)
			admin.DELETE("/apikeys/:id", handler.DeleteAPIKey)

			admin.GET("/update/check", handler.CheckUpdate)
			admin.POST("/update/apply", handler.ApplyUpdate)
		}
	}

	// Serve frontend
	distFS, err := fs.Sub(frontendFS, "frontend/dist")
	if err != nil {
		log.Fatalf("Frontend embed failed: %v", err)
	}

	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		// Try to serve static file
		if !strings.HasPrefix(path, "/api/") {
			f, err := distFS.Open(strings.TrimPrefix(path, "/"))
			if err == nil {
				f.Close()
				c.FileFromFS(strings.TrimPrefix(path, "/"), http.FS(distFS))
				return
			}
			// SPA fallback
			c.FileFromFS("index.html", http.FS(distFS))
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"message": "Not Found"})
	})

	addr := fmt.Sprintf("%s:%d", config.Cfg.Server.Host, config.Cfg.Server.Port)
	log.Printf("AnkerYe-PMGT Server starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
