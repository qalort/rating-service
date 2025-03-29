package main

import (
        "fmt"
        "os"

        "github.com/gin-gonic/gin"

        "rating-system/internal/domain/service"
        "rating-system/internal/infrastructure/db"
        "rating-system/internal/infrastructure/handler"
        "rating-system/internal/infrastructure/repository"
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
        svc := service.NewRatingService(repo, log)

        // Initialize HTTP handler with Gin
        router := gin.Default()
        router.Use(gin.Recovery())
        router.Use(corsMiddleware())
        
        // Initialize API handler
        h := handler.NewHandler(svc, log)
        setupRoutes(router, h)

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

func setupRoutes(router *gin.Engine, h *handler.Handler) {
        api := router.Group("/api/v1")
        {
                ratings := api.Group("/ratings")
                {
                        ratings.POST("", h.CreateRating)
                        ratings.GET("/service/:serviceID", h.GetRatingsByService)
                        ratings.GET("/service/:serviceID/average", h.GetAverageRating)
                        ratings.GET("/user/:userID/service/:serviceID", h.GetUserRating)
                }
                
                reviews := api.Group("/reviews")
                {
                        reviews.POST("", h.CreateReview)
                        reviews.GET("/service/:serviceID", h.GetReviewsByService)
                        reviews.GET("/:reviewID", h.GetReviewByID)
                }
                
                comments := api.Group("/comments")
                {
                        comments.POST("", h.CreateComment)
                        comments.GET("/review/:reviewID", h.GetCommentsByReview)
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
