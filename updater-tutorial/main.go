package main

import (
	"context"
	"embed"
	"os"

	"log"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
	"github.com/wailsapp/wails/v3/pkg/updater"
	"github.com/wailsapp/wails/v3/pkg/updater/providers/github"
)

var (
	currentVersion = "1.0.0" // overridden by -ldflags at release time
)

// Wails uses Go's `embed` package to embed the frontend files into the binary.
// Any files in the frontend/dist folder will be embedded into the binary and
// made available to the frontend.
// See https://pkg.go.dev/embed for more information.

//go:embed updater-window.html
var updaterHTML string

//go:embed all:frontend/dist
var assets embed.FS

//go:embed updater.key.pub
var updaterPublicKey []byte

type fixedSizeUpdaterWindow struct{ h updater.WindowHandle }

func (f fixedSizeUpdaterWindow) Show()                             { f.h.Show() }
func (f fixedSizeUpdaterWindow) Close()                            { f.h.Close() }
func (f fixedSizeUpdaterWindow) EmitEvent(n string, d ...any) bool { return f.h.EmitEvent(n, d...) }

func init() {
	// Register a custom event whose associated data type is string.
	// This is not required, but the binding generator will pick up registered events
	// and provide a strongly typed JS/TS API for them.
	application.RegisterEvent[string]("time")
}

// main function serves as the application's entry point. It initializes the application,
// creates a window, and starts a goroutine that emits a time-based event every second.
func main() {

	app := application.New(application.Options{
		Name:        "updater-tutorial",
		Description: "A demo of using raw HTML & CSS",
		Services: []application.Service{
			application.NewService(&GreetService{}),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	gh, err := github.New(github.Config{
		Repository:    "davasorus/go-lectron",
		ChecksumAsset: "SHA256SUMS",
		Token:         os.Getenv("GITHUB_TOKEN"),
	})
	if err != nil {
		log.Fatalf("github.New: %v", err)
	}

	updaterWin := app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:                "Updater",
		Width:                480,
		Height:               460,
		Hidden:               true,
		HTML:                 updaterHTML,
		AllowSimpleEventEmit: true,
	})

	// beta.16 BYO contract (from source, guide is wrong): the updater never
	// Shows a BYO window but does Close it at session end — and closing
	// destroys a Wails window. Hook the close into a hide so the handle
	// survives across flows.
	updaterWin.RegisterHook(events.Common.WindowClosing, func(e *application.WindowEvent) {
		e.Cancel()
		updaterWin.Hide()
	})

	if err := app.Updater.Init(updater.Config{
		CurrentVersion: currentVersion,
		PublicKey:      updaterPublicKey,
		Providers:      []updater.Provider{gh},
		CheckInterval:  6 * time.Hour,
		Window:         updater.BYOWindow(fixedSizeUpdaterWindow{h: updaterWin.AsUpdaterWindow()}),
	}); err != nil {
		log.Fatalf("Updater.Init: %v", err)
	}

	// The updater won't Show a BYO window — do it ourselves on flow start.
	app.Event.On(updater.EventCheckStarted, func(*application.CustomEvent) {
		updaterWin.Show()
	})

	menu := app.Menu.New()
	app.Menu.SetApplicationMenu(menu)
	appMenu := menu.AddSubmenu("App")
	appMenu.Add("Check for Updates…").OnClick(func(*application.Context) {
		app.Logger.Info("menu: check clicked",
			"state", app.Updater.State(),
			"skipped", app.Updater.SkippedVersion(),
			"version", app.Updater.CurrentVersion())
		go func() {
			if err := app.Updater.CheckAndInstall(context.Background()); err != nil {
				app.Logger.Error("update", "error", err)
			}
		}()
	})

	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:              "Window 1",
		UseApplicationMenu: true,
		// Window sized to the golden ratio (1000 / 618 ≈ 1.618).
		Width:  1000,
		Height: 618,
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 50,
			Backdrop:                application.MacBackdropTranslucent,
			TitleBar:                application.MacTitleBarHiddenInset,
		},
		BackgroundColour: application.NewRGB(6, 7, 15),
		URL:              "/",
	})

	// Emit the current time every second; the frontend listens for this event.
	go func() {
		for {
			now := time.Now().Format(time.RFC1123)
			app.Event.Emit("time", now)
			time.Sleep(time.Second)
		}
	}()

	// Run the application. This blocks until the application has been exited.
	err = app.Run()
	if err != nil {
		log.Fatal(err)
	}
}
