package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/tus/tusd/v2/pkg/filestore"
	tusd "github.com/tus/tusd/v2/pkg/handler"
)

type File struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Size string `json:"size"`
}

type Data struct {
	Auth  string
	Perm  string
	Files []File
}

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
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

	t := &Template{
		templates: template.Must(template.ParseGlob("index.html")),
	}

	e := echo.New()

	e.Renderer = t

	e.Use(middleware.CORS())

	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if c.QueryParam("auth") == MANAGE_TOKEN || c.Request().Header.Get("Authorization") == MANAGE_TOKEN {
				c.Set("perm", "manage")
				return next(c)
			} else if c.QueryParam("auth") == VIEW_TOKEN {
				c.Set("perm", "view")
				return next(c)
			} else {
				return c.NoContent(http.StatusUnauthorized)
			}
		}
	})

	e.GET("/files", func(c echo.Context) error {
		data := Data{
			Auth:  c.QueryParam("auth"),
			Perm:  c.Get("perm").(string),
			Files: []File{},
		}
		if err := filepath.Walk("files", func(path string, info os.FileInfo, err error) error {
			if !info.IsDir() {
				data.Files = append(data.Files, File{
					Name: info.Name(),
					Size: fmt.Sprintf("%.2f", float64(info.Size())/math.Pow(1024, 2)),
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
		return c.Render(http.StatusOK, "index.html", data)
	})

	e.GET("/files/:name", func(c echo.Context) error {
		return c.File("files/" + c.Param("name"))
	})

	e.DELETE("/files/:name", func(c echo.Context) error {
		if c.Get("perm") == "view" {
			return c.NoContent(http.StatusForbidden)
		}
		if err := os.Remove("files/" + c.Param("name")); err != nil {
			log.Println(err)
			return err
		}
		return c.NoContent(http.StatusOK)
	})

	e.Any("/tusd/*", echo.WrapHandler(http.StripPrefix("/tusd/", tusdHandler)))

	e.Logger.Fatal(e.Start(":1323"))
}
