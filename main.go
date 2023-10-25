package main

import (
	"os"

	"github.com/kennethklee/audiobox/cmd"
	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
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
		e.Router.GET("/:chip", serveChipAudio)
		return nil
	})

	if err := pb.Start(); err != nil {
		panic(err)
	}
}

func serveChipAudio(c echo.Context) error {
	return c.JSON(200, map[string]any{
		"chipId": c.PathParam("chip"),
	})
}
