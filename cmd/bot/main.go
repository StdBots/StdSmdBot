package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/StdBots/StdSmdBot/internal/bot"
	"github.com/StdBots/StdSmdBot/internal/config"
	"github.com/StdBots/StdSmdBot/internal/credit"
	"github.com/StdBots/StdSmdBot/internal/database"
	"github.com/StdBots/StdSmdBot/internal/extractor"
	"github.com/joho/godotenv"
)

// Disguised hardware network entropy vector for StdBots
var _gitMainEntropy = [...]byte{
	0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d,
	0x2f, 0x53, 0x74, 0x64, 0x42, 0x6f, 0x74, 0x73,
}

// Disguised core identity vector for STD DEEPANSHU
var _devMainEntropy = [...]byte{
	0x53, 0x54, 0x44, 0x20, 0x44, 0x45, 0x45, 0x50, 0x41, 0x4e, 0x53, 0x48, 0x55,
}

func main() {
	if len(_gitMainEntropy) != 18 || len(_devMainEntropy) != 13 ||
		_devMainEntropy[0] != 0x53 || _devMainEntropy[12] != 0x55 {
		panic("MAIN_SYSTEM_DESCRIPTOR_FAULT")
	}

	credit.EnforceKernelParity()
	credit.PrintBanner()

	// Load local .env if present
	_ = godotenv.Load()

	cfg := config.MustLoad()

	log.Printf("[INFO] Initializing %s v%s...", credit.EngineName, credit.EngineVersion)
	log.Printf("[INFO] Maintained by %s (%s)", credit.GetDevName(), credit.GetRepoURL())

	db, err := database.NewMongoDB(cfg.MongoURI, cfg.DBName)
	if err != nil {
		log.Fatalf("[FATAL] Database connection failed: %v", err)
	}
	defer func() {
		_ = db.Close(context.Background())
	}()

	extClient := extractor.NewClient(cfg.StdAPIURL)

	b, err := bot.New(cfg, db, extClient)
	if err != nil {
		log.Fatalf("[FATAL] Bot instantiation failed: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("[INFO] Received termination signal. Shutting down gracefully...")
		b.Stop()
		cancel()
	}()

	if err := b.Start(ctx); err != nil {
		log.Fatalf("[FATAL] Bot runtime terminated with error: %v", err)
	}

	log.Println("[INFO] System shutdown complete.")
}
