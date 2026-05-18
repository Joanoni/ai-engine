package main

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/signal"
	"syscall"

	aiengine "github.com/swarmit/ai-engine"
	"github.com/swarmit/ai-engine/internal/config"
	"github.com/swarmit/ai-engine/internal/events"
	anthropicprovider "github.com/swarmit/ai-engine/internal/llm/anthropic"
	"github.com/swarmit/ai-engine/internal/registry"
	"github.com/swarmit/ai-engine/internal/sandbox"
	"github.com/swarmit/ai-engine/internal/scaffold"
	"github.com/swarmit/ai-engine/internal/server"
	"github.com/swarmit/ai-engine/internal/session"
	"github.com/swarmit/ai-engine/internal/sessionstore"
	"github.com/swarmit/ai-engine/internal/tokenstore"
)

// Version is set at build time via -ldflags "-X main.Version=x.y.z"
var Version = "dev"

func main() {
	// Workspace is the current working directory.
	workspacePath, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get working directory: %v", err)
	}

	// Handle subcommands.
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "init":
			runInit(workspacePath)
			return
		case "help", "--help", "-h":
			printUsage()
			return
		default:
			fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", os.Args[1])
			printUsage()
			os.Exit(1)
		}
	}

	runServer(workspacePath)
}

func runInit(workspacePath string) {
	fmt.Printf("Initialising AI Engine workspace at: %s\n", workspacePath)

	if err := scaffold.Init(workspacePath); err != nil {
		log.Fatalf("Init failed: %v", err)
	}

	fmt.Println()
	fmt.Println("Workspace initialised successfully.")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Open .ai-engine/.env and set your ANTHROPIC_API_KEY.")
	fmt.Println("  2. Open .ai-engine/config.json and set provider and default_model.")
	fmt.Println("  3. Add your agents under .ai-engine/agents/.")
	fmt.Println("  4. Run ai-engine to start the server.")
}

func runServer(workspacePath string) {
	// Load workspace .ai-engine/.env (optional — shell variables take precedence).
	if err := config.LoadEnv(workspacePath); err != nil {
		log.Fatalf("Failed to load workspace .env: %v", err)
	}

	// 1. Create Sandbox.
	sb, err := sandbox.New(workspacePath)
	if err != nil {
		log.Fatalf("Failed to create sandbox: %v", err)
	}

	// 2. Create LLM provider (Anthropic).
	provider, err := anthropicprovider.New()
	if err != nil {
		log.Fatalf("Failed to create LLM provider: %v", err)
	}

	// 3. Create agent Registry.
	reg := registry.New(workspacePath)

	// 4. Create session Manager.
	sessionMgr := session.NewManager()

	// 5. Create event Bus.
	bus := events.NewBus()

	// Config is hot-reloaded on each session start inside the server handler.
	// We load it once here only to print startup info and get the port.
	cfg, err := config.Load(workspacePath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	fmt.Printf("AI Engine starting\n")
	fmt.Printf("  Workspace : %s\n", workspacePath)
	fmt.Printf("  Provider  : %s\n", cfg.Provider)
	fmt.Printf("  Model     : %s\n", cfg.DefaultModel)
	fmt.Printf("  Root agent: %s\n", cfg.RootAgent)
	fmt.Printf("  Port      : %d\n", cfg.Port)
	fmt.Printf("  Version   : %s\n", Version)

	// 6. Strip the "frontend/dist" prefix so files are served at /.
	distFS, err := fs.Sub(aiengine.Files, "frontend/dist")
	if err != nil {
		log.Fatalf("Failed to create sub FS: %v", err)
	}

	// 7. Create session store for disk persistence.
	store := sessionstore.New(workspacePath)

	// 8. Create token store for token usage tracking.
	tokenStore := tokenstore.New(workspacePath)

	// 9. Create and start the WebSocket server with all dependencies.
	// tools.Registry and agent.Runner are created per-session inside the server handler.
	srv := server.New(cfg.Port, sessionMgr, bus, reg, sb, provider, distFS, store, tokenStore, Version)

	// Set up context that cancels on OS signal.
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := srv.Start(ctx); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func printUsage() {
	fmt.Println("Usage: ai-engine [command]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  init    Initialise a new AI Engine workspace in the current directory")
	fmt.Println("  help    Show this help message")
	fmt.Println()
	fmt.Println("With no command, starts the AI Engine server.")
}
