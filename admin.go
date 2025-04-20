package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// AdminMiddleware is middleware that ensures the user has admin privileges
type AdminMiddleware struct {
	db *Database
}

// NewAdminMiddleware creates a new admin middleware
func NewAdminMiddleware(db *Database) *AdminMiddleware {
	return &AdminMiddleware{
		db: db,
	}
}

// RequireAdmin ensures the user has admin role
func (m *AdminMiddleware) RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user from auth middleware
		userInterface, exists := c.Get("user")
		if !exists {
			c.Redirect(http.StatusSeeOther, "/login?error=You must be logged in to access the admin panel")
			c.Abort()
			return
		}
		
		user := userInterface.(User)
		
		// Check if user has admin role
		if user.Role != RoleAdmin {
			c.HTML(http.StatusForbidden, "error.html", gin.H{
				"error": "Forbidden: You don't have permission to access this page",
			})
			c.Abort()
			return
		}
		
		c.Next()
	}
}

// AdminHandlers contains all admin-related request handlers
type AdminHandlers struct {
	db *Database
}

// NewAdminHandlers creates a new AdminHandlers instance
func NewAdminHandlers(db *Database) *AdminHandlers {
	return &AdminHandlers{
		db: db,
	}
}

// Dashboard renders the admin dashboard with statistics
func (h *AdminHandlers) Dashboard(c *gin.Context) {
	userStats, err := h.db.GetUserStats()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": "Error fetching user statistics: " + err.Error(),
		})
		return
	}
	
	tripStats, err := h.db.GetTripStats()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": "Error fetching trip statistics: " + err.Error(),
		})
		return
	}
	
	c.HTML(http.StatusOK, "admin_dashboard.html", gin.H{
		"title":      "Admin Dashboard",
		"userStats":  userStats,
		"tripStats":  tripStats,
	})
}

// UserManagement renders the user management page
func (h *AdminHandlers) UserManagement(c *gin.Context) {
	page := 1
	perPage := 20
	
	users, err := h.db.GetAllUsers(page, perPage)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": "Error fetching users: " + err.Error(),
		})
		return
	}
	
	totalUsers, err := h.db.GetTotalUserCount()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": "Error fetching user count: " + err.Error(),
		})
		return
	}
	
	totalPages := (totalUsers + perPage - 1) / perPage
	
	c.HTML(http.StatusOK, "admin_users.html", gin.H{
		"title":       "User Management",
		"users":       users,
		"currentPage": page,
		"totalPages":  totalPages,
		"totalUsers":  totalUsers,
	})
}

// UpdateUserRole updates a user's role
func (h *AdminHandlers) UpdateUserRole(c *gin.Context) {
	userID := c.Param("id")
	role := c.PostForm("role")
	
	if role != RoleUser && role != RoleAdmin {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role"})
		return
	}
	
	if err := h.db.UpdateUserRole(userID, role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	// Add audit log
	adminUser, _ := c.Get("user")
	h.db.AddAuditLog(adminUser.(User).ID, "update_user_role", 
		"Updated user " + userID + " role to " + role, c.ClientIP())
	
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// DeleteUserHandler deletes a user and all associated data
func (h *AdminHandlers) DeleteUserHandler(c *gin.Context) {
	userID := c.Param("id")
	
	if err := h.db.DeleteUser(userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	// Add audit log
	adminUser, _ := c.Get("user")
	h.db.AddAuditLog(adminUser.(User).ID, "delete_user", 
		"Deleted user " + userID, c.ClientIP())
	
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// AuditLogs renders the audit logs page
func (h *AdminHandlers) AuditLogs(c *gin.Context) {
	page := 1
	perPage := 50
	
	logs, err := h.db.GetAuditLogs(page, perPage)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": "Error fetching audit logs: " + err.Error(),
		})
		return
	}
	
	c.HTML(http.StatusOK, "admin_logs.html", gin.H{
		"title": "System Audit Logs",
		"logs":  logs,
	})
}
