package main

import (
	"io"
	"log/slog"
	"net/http"
	"os"

	_ "github.com/breml/rootcerts"

	"github.com/kennethklee/audiobox/cmd"
	"github.com/labstack/echo/v5"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"
)

var Version = "(untracked)"
var pb *pocketbase.PocketBase
var dev = os.Getenv("APP_ENV") == "development"

func main() {
	pb = pocketbase.New()
	pb.RootCmd.Use = os.Args[0]
	pb.RootCmd.Short = "Audiobox Web server"
	pb.RootCmd.Version = Version
	if dev {
		migratecmd.MustRegister(pb, pb.RootCmd, migratecmd.Config{Automigrate: dev})
	}
	pb.RootCmd.AddCommand(cmd.NewHealthCheckCmd(pb))

	// Create tmpdir if not exists
	if err := os.MkdirAll(os.TempDir(), 0755); err != nil {
		panic(err)
	}

	pb.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		e.Router.GET("/nfc/:chip", serveChipAudio(pb))
		return nil
	})

	if err := pb.Start(); err != nil {
		panic(err)
	}
}

func serveChipAudio(app core.App) echo.HandlerFunc {
	return func(c echo.Context) error {
		chips, err := app.Dao().FindRecordsByExpr("chips", &dbx.HashExp{"chip": c.PathParam("chip")})
		if err != nil {
			return echo.ErrNotFound
		}

		if len(chips) == 0 {
			return echo.ErrNotFound
		}

		audio, err := app.Dao().FindRecordById("audio", chips[0].GetString("audio"))
		if err != nil {
			slog.WarnContext(c.Request().Context(), "Couldn't find associated audio with chip", "chip", c.PathParam("chip"), "err", err)
			return echo.ErrNotFound
		}

		slog.InfoContext(c.Request().Context(), "Found audio for chip", "chip", c.PathParam("chip"), "audio", audio.PublicExport())
		if audio.GetString("type") == "url" {
			return streamAudioUrl(c, audio.GetString("url"))
		} else if audio.GetString("type") == "file" {
			return streamAudioFile(app, c, audio)
		}

		return echo.ErrNotFound
	}
}

func streamAudioUrl(c echo.Context, audioUrl string) error {
	resp, err := http.Get(audioUrl)
	if err != nil {
		slog.ErrorContext(c.Request().Context(), "Couldn't stream audio from url", "url", audioUrl, "err", err)
		return err
	}
	defer resp.Body.Close()

	// Set the content type and length headers
	c.Response().Header().Set(echo.HeaderContentType, resp.Header.Get(echo.HeaderContentType))
	c.Response().Header().Set(echo.HeaderContentLength, resp.Header.Get(echo.HeaderContentLength))

	// Stream the response body to the client
	_, err = io.Copy(c.Response(), resp.Body)
	if err != nil {
		slog.ErrorContext(c.Request().Context(), "Couldn't stream audio from url", "url", audioUrl, "err", err)
	}
	return err
}

func streamAudioFile(app core.App, c echo.Context, audio *models.Record) error {
	baseFilesPath := audio.BaseFilesPath()
	originalPath := baseFilesPath + "/" + audio.GetString("file")
	fs, err := app.NewFilesystem()
	if err != nil {
		slog.ErrorContext(c.Request().Context(), "Couldn't create filesystem", "err", err)
		return err
	}
	defer fs.Close()

	if err := fs.Serve(c.Response(), c.Request(), originalPath, audio.GetString("file")); err != nil {
		slog.ErrorContext(c.Request().Context(), "Couldn't serve audio file", "err", err)
		return err
	}
	return nil
}
