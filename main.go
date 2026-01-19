package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/docker/go-units"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/tus/tusd/v2/pkg/filestore"
	tusd "github.com/tus/tusd/v2/pkg/handler"
)

type File struct {
	Name     string `json:"name"`
	Size     string `json:"size"`
	Modified string `json:"modified"`
}

type Data struct {
	Auth  string
	Perm  string
	Files []File
}

func main() {
	MANAGE_TOKEN := os.Getenv("MANAGE_TOKEN")
	VIEW_TOKEN := os.Getenv("VIEW_TOKEN")

	if err := os.MkdirAll("files", os.ModePerm); err != nil {
		log.Fatal(err)
	}

	store := filestore.FileStore{
		Path: "files",
	}
	composer := tusd.NewStoreComposer()
	store.UseIn(composer)

	tusdHandler, err := tusd.NewHandler(tusd.Config{
		BasePath:              "/tusd/",
		StoreComposer:         composer,
		NotifyCompleteUploads: true,
	})
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for {
			event := <-tusdHandler.CompleteUploads
			if err := os.Rename("files/"+event.Upload.ID, "files/"+event.Upload.MetaData["filename"]); err != nil {
				log.Println(err)
			}
			if err := os.Remove("files/" + event.Upload.ID + ".info"); err != nil {
				log.Println(err)
			}
			log.Println(event.Upload.MetaData["filename"] + " Uploaded")
		}
	}()

	e := echo.New()

	e.Renderer = &echo.TemplateRenderer{
		Template: template.Must(template.ParseGlob("index.html")),
	}

	e.Use(middleware.CORS("*"))

	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			if c.QueryParam("auth") == MANAGE_TOKEN || c.Request().Header.Get("Authorization") == MANAGE_TOKEN {
				c.Set("perm", "manage")
				return next(c)
			} else if c.QueryParam("auth") == VIEW_TOKEN && c.Request().Method == "GET" {
				c.Set("perm", "view")
				return next(c)
			} else {
				return c.NoContent(http.StatusUnauthorized)
			}
		}
	})

	e.File("/favicon.ico", "favicon.ico")

	e.GET("/", func(c *echo.Context) error {
		data := Data{
			Auth:  c.QueryParam("auth"),
			Perm:  c.Get("perm").(string),
			Files: []File{},
		}
		if err := filepath.Walk("files", func(path string, info os.FileInfo, err error) error {
			if !info.IsDir() {
				data.Files = append(data.Files, File{
					Name:     info.Name(),
					Size:     units.HumanSize(float64(info.Size())),
					Modified: units.HumanDuration(time.Since(info.ModTime())) + " ago",
				})
			}
			if err != nil {
				log.Println(err)
				return err
			}
			return nil
		}); err != nil {
			log.Println(err)
			return err
		}
		if c.Request().Header.Get("Accept") == "application/json" {
			return c.JSON(http.StatusOK, data.Files)
		} else {
			return c.Render(http.StatusOK, "index.html", data)
		}
	})

	e.GET("/file/:name", func(c *echo.Context) error {
		return c.File("files/" + c.Param("name"))
	})

	e.DELETE("/file/:name", func(c *echo.Context) error {
		if err := os.Remove("files/" + strings.ReplaceAll(c.Param("name"), "%20", " ")); err != nil {
			log.Println(err)
			return err
		}
		return c.NoContent(http.StatusNoContent)
	})

	e.Any("/tusd/*", echo.WrapHandler(http.StripPrefix("/tusd/", tusdHandler)))

	if err := e.Start(":1323"); err != nil {
		e.Logger.Error("failed to start server", "error", err)
	}
}
