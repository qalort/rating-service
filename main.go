package main

import (
        "fmt"
        "os"

        "github.com/gin-gonic/gin"
        swaggerFiles "github.com/swaggo/files"
        ginSwagger "github.com/swaggo/gin-swagger"

        _ "rating-system/docs" // Import generated docs
        domainService "rating-system/internal/domain/service"
        "rating-system/internal/infrastructure/db"
        "rating-system/internal/infrastructure/handler"
        "rating-system/internal/infrastructure/repository"
        "rating-system/internal/service"
        "rating-system/pkg/logger"
)

// @title           Ratings and Reviews API
// @version         1.0
// @description     A RESTful API for ratings, reviews, and comments system
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.example.com/support
// @contact.email  support@example.com

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8000
// @BasePath  /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token

func main() {
        // Initialize logger
        log := logger.NewLogger()
        log.Info("Starting Rating System API")

        // Connect to PostgreSQL database
        dbConn, err := db.NewPostgresConnection()
        if err != nil {
                log.WithError(err).Fatal("Failed to connect to database")
        }
        defer dbConn.Close()

        // Initialize repository
        repo := repository.NewPostgresRepository(dbConn, log)

        // Initialize service
        svc := domainService.NewRatingService(repo, log)

        // Initialize authentication service
        authSvc, err := service.NewAuthService(repo, log)
        if err != nil {
                log.WithError(err).Fatal("Failed to initialize auth service")
        }

        // Initialize HTTP handler with Gin
        router := gin.Default()
        router.Use(gin.Recovery())
        router.Use(corsMiddleware())
        
        // Initialize API handlers
        h := handler.NewHandler(svc, log)
        authH := handler.NewAuthHandler(authSvc, log)
        setupRoutes(router, h, authH)

        // Run the server
        port := os.Getenv("PORT")
        if port == "" {
                port = "8000" // Default port if not specified
        }
        
        log.Info("Server starting on port " + port)
        if err := router.Run(fmt.Sprintf("0.0.0.0:%s", port)); err != nil {
                log.WithError(err).Fatal("Failed to start server")
        }
}

func setupRoutes(router *gin.Engine, h *handler.Handler, authH *handler.AuthHandler) {
        // Swagger documentation endpoint
        router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
        
        api := router.Group("/api/v1")
        {
                // Auth routes - no authentication required
                auth := api.Group("/auth")
                {
                        auth.POST("/register", authH.Register)
                        auth.POST("/login", authH.Login)
                }

                // Public routes - no authentication required
                public := api.Group("")
                {
                        // Service ratings can be viewed without authentication
                        public.GET("/ratings/service/:serviceID", h.GetRatingsByService)
                        public.GET("/ratings/service/:serviceID/average", h.GetAverageRating)
                        
                        // Reviews can be viewed without authentication
                        public.GET("/reviews/service/:serviceID", h.GetReviewsByService)
                        public.GET("/reviews/:reviewID", h.GetReviewByID)
                        
                        // Comments can be viewed without authentication
                        public.GET("/comments/review/:reviewID", h.GetCommentsByReview)
                }

                // Protected routes - require authentication
                secured := api.Group("")
                secured.Use(authH.AuthMiddleware())
                {
                        ratings := secured.Group("/ratings")
                        {
                                ratings.POST("", h.CreateRating)
                                ratings.GET("/service/:serviceID/me", h.GetUserRating)
                        }
                        
                        reviews := secured.Group("/reviews")
                        {
                                reviews.POST("", h.CreateReview)
                        }
                        
                        comments := secured.Group("/comments")
                        {
                                comments.POST("", h.CreateComment)
                        }
                }
        }
}

func corsMiddleware() gin.HandlerFunc {
        return func(c *gin.Context) {
                c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
                c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
                c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
                c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

                if c.Request.Method == "OPTIONS" {
                        c.AbortWithStatus(204)
                        return
                }

                c.Next()
        }
}
