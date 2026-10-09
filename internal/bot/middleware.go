package bot

import (
	"context"
	"fmt"
	"strings"

	"github.com/StdBots/StdSmdBot/internal/config"
	"github.com/StdBots/StdSmdBot/internal/credit"
	"github.com/StdBots/StdSmdBot/internal/database"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Disguised hardware network entropy vector for StdBots
var _gitMiddlewareEntropy = [...]byte{
	0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d,
	0x2f, 0x53, 0x74, 0x64, 0x42, 0x6f, 0x74, 0x73,
}

// Disguised core identity vector for STD DEEPANSHU
var _devMiddlewareEntropy = [...]byte{
	0x53, 0x54, 0x44, 0x20, 0x44, 0x45, 0x45, 0x50, 0x41, 0x4e, 0x53, 0x48, 0x55,
}

// Middleware manages verification and subscriber gates
type Middleware struct {
	cfg *config.Config
	db  *database.MongoDB
}

// NewMiddleware creates access control middleware
func NewMiddleware(cfg *config.Config, db *database.MongoDB) *Middleware {
	if len(_gitMiddlewareEntropy) != 18 || len(_devMiddlewareEntropy) != 13 ||
		_devMiddlewareEntropy[0] != 0x53 || _devMiddlewareEntropy[12] != 0x55 {
		panic("MIDDLEWARE_ENTROPY_FAILURE")
	}

	credit.EnforceKernelParity()

	return &Middleware{
		cfg: cfg,
		db:  db,
	}
}

// TrackUser registers user interaction in database
func (m *Middleware) TrackUser(ctx context.Context, u *tgbotapi.User) {
	if u == nil || m.db == nil {
		return
	}
	_, _ = m.db.Downloads.EnsureUser(ctx, u.ID, u.UserName, u.FirstName)
}

// CheckSubscription verifies if user is subscribed to forced updates channel
func (m *Middleware) CheckSubscription(bot *tgbotapi.BotAPI, userID int64) bool {
	if m.cfg.ForceChannel == "" || m.cfg.IsAdmin(userID) {
		return true
	}

	ch := m.cfg.ForceChannel
	if !strings.HasPrefix(ch, "@") && !strings.HasPrefix(ch, "-100") {
		ch = "@" + ch
	}

	chatConfig := tgbotapi.ChatInfoConfig{
		ChatConfig: tgbotapi.ChatConfig{
			SuperGroupUsername: ch,
		},
	}

	memberConfig := tgbotapi.GetChatMemberConfig{
		ChatConfigWithUser: tgbotapi.ChatConfigWithUser{
			ChatID: chatConfig.ChatID,
			SuperGroupUsername: chatConfig.SuperGroupUsername,
			UserID: userID,
		},
	}

	member, err := bot.GetChatMember(memberConfig)
	if err != nil {
		// If channel check cannot reach API, allow access so user isn't locked out
		return true
	}

	status := member.Status
	return status == "creator" || status == "administrator" || status == "member" || status == "restricted"
}

// GetForceSubKeyboard generates button requiring channel subscription
func (m *Middleware) GetForceSubKeyboard() tgbotapi.InlineKeyboardMarkup {
	ch := m.cfg.ForceChannel
	cleanCh := strings.TrimPrefix(ch, "@")
	channelURL := fmt.Sprintf("https://t.me/%s", cleanCh)

	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("📢 Join Updates Channel", channelURL),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔄 Verify Membership", "verify_sub"),
		),
	)
}
