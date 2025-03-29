package main

import (
        "fmt"
        "os"

        "github.com/gin-gonic/gin"

        domainService "rating-system/internal/domain/service"
        "rating-system/internal/infrastructure/db"
        "rating-system/internal/infrastructure/handler"
        "rating-system/internal/infrastructure/repository"
        "rating-system/internal/service"
        "rating-system/pkg/logger"
)

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
        api := router.Group("/api/v1")
        {
                // Auth routes - no authentication required
                auth := api.Group("/auth")
                {
                        auth.POST("/register", authH.Register)
                        auth.POST("/login", authH.Login)
                }

                // Protected routes - require authentication
                secured := api.Group("")
                secured.Use(authH.AuthMiddleware())
                {
                        ratings := secured.Group("/ratings")
                        {
                                ratings.POST("", h.CreateRating)
                                ratings.GET("/service/:serviceID", h.GetRatingsByService)
                                ratings.GET("/service/:serviceID/average", h.GetAverageRating)
                                ratings.GET("/user/:userID/service/:serviceID", h.GetUserRating)
                        }
                        
                        reviews := secured.Group("/reviews")
                        {
                                reviews.POST("", h.CreateReview)
                                reviews.GET("/service/:serviceID", h.GetReviewsByService)
                                reviews.GET("/:reviewID", h.GetReviewByID)
                        }
                        
                        comments := secured.Group("/comments")
                        {
                                comments.POST("", h.CreateComment)
                                comments.GET("/review/:reviewID", h.GetCommentsByReview)
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
