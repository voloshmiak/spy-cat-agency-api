package main

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
	"os/signal"
	"spy-cat-agency/config"
	"spy-cat-agency/internal/cat"
	"spy-cat-agency/internal/db"
	"spy-cat-agency/internal/middleware"
	"spy-cat-agency/internal/mission"
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

	c, err := config.New()
	if err != nil {
		return err
	}

	conn, err := db.Connect(c.Database.User, c.Database.Password,
		c.Database.Host, c.Database.DBName, c.Database.Port)
	if err != nil {
		return err
	}

	err = db.Migrate(conn, "./migrations")
	if err != nil {
		return err
	}

	router := gin.New()

	router.Use(middleware.Logger())

	cr := cat.NewRepository(conn)
	cs := cat.NewService(cr)
	ch := cat.NewHandler(cs)

	v1 := router.Group("/api/v1")

	catRoutes := v1.Group("/cats")
	{
		catRoutes.GET("", ch.ListCats)         // api/v1/cats
		catRoutes.POST("", ch.CreateCat)       // api/v1/cats
		catRoutes.GET("/:id", ch.GetCat)       // api/v1/cats/:id
		catRoutes.PATCH("/:id", ch.UpdateCat)  // api/v1/cats/:id
		catRoutes.DELETE("/:id", ch.DeleteCat) // api/v1/cats/:id
	}

	mr := mission.NewRepository(conn)
	ms := mission.NewService(mr)
	mh := mission.NewHandler(ms, cs)

	missionRoutes := v1.Group("/missions")
	{
		missionRoutes.GET("", mh.ListMissions)         // api/v1/missions
		missionRoutes.POST("", mh.CreateMission)       // api/v1/missions
		missionRoutes.GET("/:id", mh.GetMission)       // api/v1/missions/:id
		missionRoutes.PATCH("/:id", mh.UpdateMission)  // api/v1/missions/:id
		missionRoutes.DELETE("/:id", mh.DeleteMission) // api/v1/missions/:id

		missionRoutes.POST("/:id/targets", mh.AddTarget)                 // api/v1/missions/:id/targets
		missionRoutes.PATCH("/:id/targets/:target_id", mh.UpdateTarget)  // api/v1/missions/:id/targets/:target_id
		missionRoutes.DELETE("/:id/targets/:target_id", mh.DeleteTarget) // api/v1/missions/:id/targets/:target_id
	}

	s := &http.Server{
		Addr:         ":" + c.Server.Port,
		Handler:      router,
		ReadTimeout:  c.Server.ReadTimeout,
		WriteTimeout: c.Server.WriteTimeout,
		IdleTimeout:  c.Server.IdleTimeout,
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
