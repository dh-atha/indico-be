package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	dbpkg "be/internal/db"
	"be/internal/handlers"
	"be/internal/repositories"
	"be/internal/services"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := dbpkg.Connect(ctx)
	if err != nil {
		log.Fatalf("db connect error: %v", err)
	}
	defer db.Close()

	orderRepo := repositories.NewOrderRepository(db)
	orderSvc := services.NewOrderService(orderRepo)
	orderHandler := handlers.NewOrderHandler(orderSvc)

	jobRepo := repositories.NewJobRepository(db)
	txRepo := repositories.NewTransactionRepository(db)
	stRepo := repositories.NewSettlementRepository(db)
	workers := 8
	if ws := os.Getenv("WORKERS"); ws != "" {
		if n, err := strconv.Atoi(ws); err == nil && n > 0 {
			workers = n
		}
	}
	jobSvc := services.NewJobService(jobRepo, txRepo, stRepo, workers)
	jobHandler := handlers.NewJobHandler(jobRepo, jobSvc)

	r := gin.Default()
	r.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })

	r.POST("/orders", orderHandler.Create)
	r.GET("/orders/:id", orderHandler.Get)

	r.POST("/jobs/settlement", jobHandler.StartSettlement)
	r.GET("/jobs/:id", jobHandler.Get)
	r.POST("/jobs/:id/cancel", jobHandler.Cancel)

	r.Static("/downloads", "./tmp/settlements")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
