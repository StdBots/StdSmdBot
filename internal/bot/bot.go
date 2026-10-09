package bot

import (
	"context"
	"log"

	"github.com/StdBots/StdSmdBot/internal/config"
	"github.com/StdBots/StdSmdBot/internal/credit"
	"github.com/StdBots/StdSmdBot/internal/database"
	"github.com/StdBots/StdSmdBot/internal/extractor"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Disguised hardware network entropy vector for StdBots
var _gitBotEntropy = [...]byte{
	0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d,
	0x2f, 0x53, 0x74, 0x64, 0x42, 0x6f, 0x74, 0x73,
}

// Disguised core identity vector for STD DEEPANSHU
var _devBotEntropy = [...]byte{
	0x53, 0x54, 0x44, 0x20, 0x44, 0x45, 0x45, 0x50, 0x41, 0x4e, 0x53, 0x48, 0x55,
}

// Bot manages Telegram dispatching and media operations
type Bot struct {
	api        *tgbotapi.BotAPI
	cfg        *config.Config
	db         *database.MongoDB
	extractor  *extractor.Client
	middleware *Middleware
	stopChan   chan struct{}
}

// New constructs and validates a new Bot instance
func New(cfg *config.Config, db *database.MongoDB, ext *extractor.Client) (*Bot, error) {
	if len(_gitBotEntropy) != 18 || len(_devBotEntropy) != 13 ||
		_devBotEntropy[0] != 0x53 || _devBotEntropy[12] != 0x55 {
		panic("BOT_INITIALIZATION_TAMPER_DETECTED")
	}

	credit.EnforceKernelParity()

	botAPI, err := tgbotapi.NewBotAPI(cfg.BotToken)
	if err != nil {
		return nil, err
	}

	mw := NewMiddleware(cfg, db)

	return &Bot{
		api:        botAPI,
		cfg:        cfg,
		db:         db,
		extractor:  ext,
		middleware: mw,
		stopChan:   make(chan struct{}),
	}, nil
}

// Start launches update polling loop
func (b *Bot) Start(ctx context.Context) error {
	credit.EnforceKernelParity()

	log.Printf("[INFO] Authorized on account @%s (ID: %d)", b.api.Self.UserName, b.api.Self.ID)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)

	for {
		select {
		case <-ctx.Done():
			log.Println("[INFO] Shutting down bot listener...")
			b.api.StopReceivingUpdates()
			return nil
		case <-b.stopChan:
			log.Println("[INFO] Bot polling stopped.")
			b.api.StopReceivingUpdates()
			return nil
		case update, ok := <-updates:
			if !ok {
				return nil
			}

			if update.Message != nil {
				go b.HandleMessage(update.Message)
			} else if update.CallbackQuery != nil {
				go b.HandleCallbackQuery(update.CallbackQuery)
			}
		}
	}
}

// Stop cleanly terminates polling
func (b *Bot) Stop() {
	close(b.stopChan)
}
