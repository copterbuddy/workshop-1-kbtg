package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/copterbuddy/assessment/expense"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	_ "github.com/lib/pq"
)

var db *sql.DB

func main() {
	db, err := InitDB()
	if err != nil {
		log.Fatal(err)
	}

	e := echo.New()
	e.Logger.SetLevel(log.INFO)
	h := expense.NewExpenseHandler(db)

	g := e.Group("/expenses")
	{
		g.Use(middleware.Logger())
		g.Use(middleware.Recover())
		g.Use(Auth)

		g.POST("/", h.CreateExpenseHandler)
		g.GET("/:id", h.GetExpenseByIdHandler)
		g.PUT("/:id", h.UpdateExpenseHandler)
		g.GET("/", h.ListExpenseHandler)
	}

	e.Logger.Fatal(e.Start(":2565"))

	// e.Logger.Fatal(e.Start(":2565"))
	go func() {
		if err := e.Start(":2565"); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal("shutting down the server: ", err)
		}
	}()

	e.GET("/", func(c echo.Context) error {
		time.Sleep(8 * time.Second)
		return c.JSON(http.StatusOK, "OK")
	})

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)
	<-shutdown
	fmt.Println("shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
	fmt.Println("bye bye")
}

func InitDB() (*sql.DB, error) {
	url := os.Getenv("DATABASE_URL")
	var err error
	db, err = sql.Open("postgres", url)
	if err != nil {
		log.Fatal("connection to database error ", url)
	}

	createTb := `
	CREATE TABLE IF NOT EXISTS expenses (
		id SERIAL PRIMARY KEY,
		title TEXT,
		amount FLOAT,
		note TEXT,
		tags TEXT[]
	);
	`
	_, err = db.Exec(createTb)
	if err != nil {
		log.Fatal("can't create database", err)
	}

	return db, nil
}

func Auth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if len(c.Request().Header["Authorization"]) > 0 {
			if c.Request().Header["Authorization"][0] == "November 10, 2009" {
				c.Response().Header().Set(echo.HeaderServer, "Echo/3.0")
				return next(c)
			}
		}
		return c.JSON(http.StatusUnauthorized, "You are not authorized!")
	}
}
