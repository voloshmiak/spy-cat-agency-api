package main

import (
	"context"
	"database/sql"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"os"
	"os/signal"
	"spy-cat-agency/internal/cat"
	"syscall"
	"time"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	gin.SetMode(gin.ReleaseMode)

	conn, err := Connect()
	if err != nil {
		return err
	}

	router := gin.Default()

	cr := cat.NewRepository(conn)
	cs := cat.NewService(cr)
	ch := cat.NewHandler(cs)

	{
		v1 := router.Group("/api/v1")
		{
			// Cat routes
			v1.GET("/cats", ch.ListCats)
			v1.POST("/cats", ch.CreateCat)
			v1.GET("/cats/:id", ch.GetCat)
			v1.PATCH("/cats/:id", ch.UpdateCat)
			v1.DELETE("/cats/:id", ch.DeleteCat)
		}
	}

	s := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	done := make(chan bool)

	go func() {
		quit := make(chan os.Signal, 1)

		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit

		log.Println("Shutting down server...")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := s.Shutdown(ctx); err != nil {
			log.Println("Server forced to shutdown:", err)
		}

		done <- true
	}()

	log.Println("Starting server on :8080")

	if err := s.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	<-done
	log.Println("Graceful shutdown complete")

	return nil
}

func Connect() (*sql.DB, error) {
	databaseURL := "postgres://postgres:12345@localhost:5432/cat-db?sslmode=disable"

	conn, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, err
	}

	driver, err := postgres.WithInstance(conn, &postgres.Config{})
	if err != nil {
		return nil, err
	}
	m, err := migrate.NewWithDatabaseInstance("file://migrations", "postgres", driver)
	if err != nil {
		return nil, err
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return nil, err
	} else if errors.Is(err, migrate.ErrNoChange) {
		log.Println("No new migrations to apply.")
	} else {
		log.Println("Migrations applied successfully!")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	err = conn.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return conn, nil
}
