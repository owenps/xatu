package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	xaws "github.com/owenps/xatu/internal/aws"
	"github.com/owenps/xatu/internal/config"
	"github.com/owenps/xatu/internal/ui"
)

func main() {
	forceSetup := flag.Bool("setup", false, "Run the setup wizard")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	client, err := xaws.NewClient(context.Background(), cfg.General.Region)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error connecting to AWS: %v\n", err)
		os.Exit(1)
	}

	needsSetup := !config.Exists() || *forceSetup
	app := ui.NewApp(cfg, client, needsSetup)

	p := tea.NewProgram(app, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
