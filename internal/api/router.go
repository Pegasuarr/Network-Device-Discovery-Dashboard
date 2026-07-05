package api

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/user/network-monitoring/internal/middleware"
	"github.com/user/network-monitoring/internal/websocket"
)

func SetupRouter(h *Handlers, wsHub *websocket.Hub) *gin.Engine {
	r := gin.New()

	// Global Middlewares
	r.Use(gin.Recovery())
	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.RateLimiter(120)) // 120 reqs/min

	// Prometheus Metrics endpoint
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// WebSocket Live stream connection
	r.GET("/ws", func(c *gin.Context) {
		websocket.HandleWS(wsHub, c)
	})

	api := r.Group("/api/v1")
	{
		// Public Auth Routes
		authGroup := api.Group("/auth")
		{
			authGroup.POST("/register", h.Register)
			authGroup.POST("/login", h.Login)
			authGroup.POST("/refresh", h.Refresh)
			authGroup.POST("/logout", h.Logout)
		}

		// Protected Routes
		protected := api.Group("/")
		protected.Use(middleware.AuthRequired())
		protected.Use(middleware.AuditLogger())
		{
			// Dashboard Endpoints
			dash := protected.Group("/dashboard")
			{
				dash.GET("/stats", h.GetDashboardStats)
				dash.GET("/latency", h.GetDashboardLatency)
			}

			// Devices Endpoints
			devices := protected.Group("/devices")
			{
				devices.GET("", h.ListDevices)
				devices.POST("", middleware.RoleRequired("Operator", "Admin"), h.CreateDevice)
				devices.PUT("/:id", middleware.RoleRequired("Operator", "Admin"), h.UpdateDevice)
				devices.DELETE("/:id", middleware.RoleRequired("Admin"), h.DeleteDevice)
				devices.GET("/:id/history", h.GetDeviceHistory)

				// CSV Bulk actions
				devices.POST("/import", middleware.RoleRequired("Operator", "Admin"), h.ImportDevices)
				devices.GET("/export", h.ExportDevices)
			}

			// Discovery Sweepers
			discovery := protected.Group("/discovery")
			{
				discovery.POST("/scan", middleware.RoleRequired("Operator", "Admin"), h.TriggerDiscovery)
			}

			// Alerts Endpoints
			alerts := protected.Group("/alerts")
			{
				alerts.GET("", h.ListAlerts)
				alerts.PUT("/:id/resolve", middleware.RoleRequired("Operator", "Admin"), h.ResolveAlert)
			}

			// Rules Endpoints
			rules := protected.Group("/rules")
			{
				rules.GET("", h.ListAlertRules)
				rules.POST("", middleware.RoleRequired("Operator", "Admin"), h.CreateAlertRule)
				rules.DELETE("/:id", middleware.RoleRequired("Admin"), h.DeleteAlertRule)
			}

			// Users Endpoints (Admin only)
			users := protected.Group("/users")
			users.Use(middleware.RoleRequired("Admin"))
			{
				users.GET("", h.ListUsers)
				users.POST("", h.CreateUser)
				users.PUT("/:id", h.UpdateUser)
				users.DELETE("/:id", h.DeleteUser)
			}

			// Settings Endpoints (Admin only)
			settings := protected.Group("/settings")
			settings.Use(middleware.RoleRequired("Admin"))
			{
				settings.GET("", h.GetSettings)
				settings.PUT("", h.SaveSettings)
			}

			// Monitoring Pause/Resume Hooks
			mon := protected.Group("/monitoring")
			{
				mon.GET("/status", h.GetMonitoringStatus)
				mon.POST("/start", middleware.RoleRequired("Admin"), h.StartMonitoring)
				mon.POST("/stop", middleware.RoleRequired("Admin"), h.StopMonitoring)
			}
		}
	}

	return r
}
