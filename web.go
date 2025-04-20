// Update the handleProfile method to include OAuth providers:

// handleProfile handles the profile page
func (s *WebServer) handleProfile(c *gin.Context) {
	// Get current user
	user := GetCurrentUser(c)
	userID := user.ID

	// Get OAuth providers
	oauthProviders, err := s.db.GetOAuthProviders(userID)
	if err != nil {
		oauthProviders = []string{}
	}

	// Get trip statistics
	tripCount := 0
	var firstTrip, lastTrip string

	// Get all trips
	trips, err := s.db.GetAllTrips(userID)
	if err == nil {
		tripCount = len(trips)
		if tripCount > 0 {
			firstTrip = trips[tripCount-1].Date
			lastTrip = trips[0].Date
		}
	}

	c.HTML(http.StatusOK, "profile.html", gin.H{
		"title": "Profile",
		"active": "profile",
		"user": user,
		"oauthProviders": oauthProviders,
		"tripCount": tripCount,
		"firstTrip": firstTrip,
		"lastTrip": lastTrip,
		"year": time.Now().Year(),
	})
}

// Update the WebServer struct to include the admin middleware and handlers
type WebServer struct {
	db     *Database
	router *gin.Engine
	port   string
	auth   *AuthMiddleware
	admin  *AdminMiddleware
	adminHandlers *AdminHandlers
}

// Update the NewWebServer function to initialize admin middleware and handlers
func NewWebServer(db *Database, port string) *WebServer {
	// Set Gin to release mode in production
	// gin.SetMode(gin.ReleaseMode)

	router := gin.Default()

	// Create auth middleware
	auth := NewAuthMiddleware(db)
	
	// Create admin middleware and handlers
	admin := NewAdminMiddleware(db)
	adminHandlers := NewAdminHandlers(db)

	server := &WebServer{
		db:           db,
		router:       router,
		port:         port,
		auth:         auth,
		admin:        admin,
		adminHandlers: adminHandlers,
	}

	// Load HTML templates
	router.SetFuncMap(template.FuncMap{
		"contains": func(slice []string, item string) bool {
			for _, s := range slice {
				if s == item {
					return true
				}
			}
			return false
		},
	})
	router.LoadHTMLGlob("templates/*")

	// Serve static files
	router.Static("/static", "./static")

	// Setup routes
	server.setupRoutes()

	return server
}

// Add this to the setupRoutes method
func (s *WebServer) setupRoutes() {
	// Existing routes...
	
	// Admin routes
	adminGroup := s.router.Group("/admin")
	adminGroup.Use(s.auth.RequireAuth())
	adminGroup.Use(s.admin.RequireAdmin())
	{
		adminGroup.GET("/", s.adminHandlers.Dashboard)
		adminGroup.GET("/users", s.adminHandlers.UserManagement)
		adminGroup.POST("/users/:id/role", s.adminHandlers.UpdateUserRole)
		adminGroup.DELETE("/users/:id", s.adminHandlers.DeleteUserHandler)
		adminGroup.GET("/logs", s.adminHandlers.AuditLogs)
	}
}
