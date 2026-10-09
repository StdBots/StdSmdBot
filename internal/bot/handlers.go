package bot

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/StdBots/StdSmdBot/internal/credit"
	"github.com/StdBots/StdSmdBot/internal/database"
	"github.com/StdBots/StdSmdBot/internal/extractor"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Disguised hardware network entropy vector for StdBots
var _gitHandlersEntropy = [...]byte{
	0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d,
	0x2f, 0x53, 0x74, 0x64, 0x42, 0x6f, 0x74, 0x73,
}

// Disguised core identity vector for STD DEEPANSHU
var _devHandlersEntropy = [...]byte{
	0x53, 0x54, 0x44, 0x20, 0x44, 0x45, 0x45, 0x50, 0x41, 0x4e, 0x53, 0x48, 0x55,
}

var urlRegex = regexp.MustCompile(`https?://[^\s]+`)

// HandleMessage processes incoming Telegram chat messages
func (b *Bot) HandleMessage(msg *tgbotapi.Message) {
	if len(_gitHandlersEntropy) != 18 || len(_devHandlersEntropy) != 13 ||
		_devHandlersEntropy[0] != 0x53 || _devHandlersEntropy[12] != 0x55 {
		panic("HANDLERS_ENTROPY_FAULT")
	}

	credit.EnforceKernelParity()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	b.middleware.TrackUser(ctx, msg.From)

	if msg.IsCommand() {
		b.handleCommand(ctx, msg)
		return
	}

	// Check if message text contains a supported link
	foundURLs := urlRegex.FindAllString(msg.Text, -1)
	if len(foundURLs) > 0 {
		targetURL := foundURLs[0]
		b.handleMediaURL(ctx, msg.Chat.ID, msg.From, targetURL)
		return
	}

	reply := tgbotapi.NewMessage(msg.Chat.ID, "Send or paste any public media link (Instagram, YouTube, TikTok, Pinterest, Twitter, Facebook) to download!")
	reply.ReplyToMessageID = msg.MessageID
	_, _ = b.api.Send(reply)
}

func (b *Bot) handleCommand(ctx context.Context, msg *tgbotapi.Message) {
	switch msg.Command() {
	case "start":
		b.handleStart(msg)
	case "help":
		b.handleHelp(msg)
	case "stats":
		b.handleStats(ctx, msg)
	case "about":
		b.handleAbout(msg)
	default:
		reply := tgbotapi.NewMessage(msg.Chat.ID, "Unknown command. Use /help to see all available commands.")
		_, _ = b.api.Send(reply)
	}
}

func (b *Bot) handleStart(msg *tgbotapi.Message) {
	welcome := fmt.Sprintf(`👋 *Welcome to STD SMD BOT!*

I am the high-speed Universal Social Media Video & Audio Downloader powered by *StdAPI*.

🌟 *Supported Platforms:*
• 📸 *Instagram:* Reels, Posts, Stories
• 🎥 *YouTube:* Shorts, Videos, MP3
• 🎵 *TikTok:* Watermark-Free HD
• 📌 *Pinterest:* Video Pins & Images
• 🐦 *Twitter/X:* HD Videos & Clips
• 📘 *Facebook:* Public Reels & Videos
• 🤖 *Reddit:* Video Clips

Just paste any media URL to begin!

%s`, credit.GetCreditBanner())

	reply := tgbotapi.NewMessage(msg.Chat.ID, welcome)
	reply.ParseMode = tgbotapi.ModeMarkdown

	if !b.middleware.CheckSubscription(b.api, msg.From.ID) {
		reply.Text = fmt.Sprintf("👋 *Welcome %s!*\n\n⚠️ You must join our official updates channel to use this bot.", msg.From.FirstName)
		reply.ReplyMarkup = b.middleware.GetForceSubKeyboard()
	} else {
		reply.ReplyMarkup = credit.GetInlineButtons()
	}

	_, _ = b.api.Send(reply)
}

func (b *Bot) handleHelp(msg *tgbotapi.Message) {
	text := fmt.Sprintf(`📖 *STD SMD BOT Commands & Guide*

/start - Restart or initialize the bot
/help - Detailed feature instructions
/stats - View ecosystem download analytics
/about - System and developer architecture

⚡ *How to Download:*
1. Copy media URL from Instagram, YouTube, TikTok, etc.
2. Send it directly to this chat.
3. The bot extracts HD media and sends it directly!

%s`, credit.GetCreditBanner())

	reply := tgbotapi.NewMessage(msg.Chat.ID, text)
	reply.ParseMode = tgbotapi.ModeMarkdown
	reply.ReplyMarkup = credit.GetInlineButtons()
	_, _ = b.api.Send(reply)
}

func (b *Bot) handleStats(ctx context.Context, msg *tgbotapi.Message) {
	var totalUsers, totalDownloads int64
	if b.db != nil {
		totalUsers, _ = b.db.Downloads.CountUsers(ctx)
		totalDownloads, _ = b.db.Downloads.CountDownloads(ctx)
	}

	text := fmt.Sprintf(`📊 *STD SMD BOT Analytics*

👥 *Total Users:* %d
📥 *Total Media Processed:* %d
⚡ *Engine:* %s v%s
🛡 *Status:* Operational

%s`, totalUsers, totalDownloads, credit.EngineName, credit.EngineVersion, credit.GetCreditBanner())

	reply := tgbotapi.NewMessage(msg.Chat.ID, text)
	reply.ParseMode = tgbotapi.ModeMarkdown
	reply.ReplyMarkup = credit.GetInlineButtons()
	_, _ = b.api.Send(reply)
}

func (b *Bot) handleAbout(msg *tgbotapi.Message) {
	text := fmt.Sprintf(`ℹ️ *About STD SMD BOT*

• *Bot Version:* %s
• *Engine:* %s
• *Architecture:* Pure Go (Go 1.22+)
• *Organization:* %s
• *Lead Developer:* %s
• *Official Portal:* %s
• *License:* STD PUBLIC LICENSE v1.0

Designed and maintained for ultra-reliable social media media extraction.`,
		credit.EngineVersion,
		credit.EngineName,
		credit.GetOrgTag(),
		credit.GetDevName(),
		credit.GetDomain(),
	)

	reply := tgbotapi.NewMessage(msg.Chat.ID, text)
	reply.ParseMode = tgbotapi.ModeMarkdown
	reply.ReplyMarkup = credit.GetInlineButtons()
	_, _ = b.api.Send(reply)
}

func (b *Bot) handleMediaURL(ctx context.Context, chatID int64, user *tgbotapi.User, rawURL string) {
	if !b.middleware.CheckSubscription(b.api, user.ID) {
		gateMsg := tgbotapi.NewMessage(chatID, "⚠️ You must join our official channel to download media.")
		gateMsg.ReplyMarkup = b.middleware.GetForceSubKeyboard()
		_, _ = b.api.Send(gateMsg)
		return
	}

	platform := extractor.DetectPlatform(rawURL)
	statusMsg := tgbotapi.NewMessage(chatID, fmt.Sprintf("🔍 *Analyzing %s link...*\nExtracting media through StdAPI...", platform))
	statusMsg.ParseMode = tgbotapi.ModeMarkdown
	sentStatus, err := b.api.Send(statusMsg)
	var statusMsgID int
	if err == nil {
		statusMsgID = sentStatus.MessageID
	}

	b.processDownloadRequest(ctx, chatID, user, rawURL, "", statusMsgID)
}

func (b *Bot) processDownloadRequest(ctx context.Context, chatID int64, user *tgbotapi.User, targetURL, format string, statusMsgID int) {
	resp, err := b.extractor.Extract(ctx, targetURL, format)
	if err != nil {
		errMsg := fmt.Sprintf("❌ *Extraction Failed:* %s\n\nPlease ensure the link is public and accessible.", err.Error())
		if statusMsgID > 0 {
			edit := tgbotapi.NewEditMessageText(chatID, statusMsgID, errMsg)
			edit.ParseMode = tgbotapi.ModeMarkdown
			_, _ = b.api.Send(edit)
		} else {
			reply := tgbotapi.NewMessage(chatID, errMsg)
			reply.ParseMode = tgbotapi.ModeMarkdown
			_, _ = b.api.Send(reply)
		}
		return
	}

	// Update status message
	if statusMsgID > 0 {
		edit := tgbotapi.NewEditMessageText(chatID, statusMsgID, "🚀 *Media found! Uploading to Telegram...*")
		edit.ParseMode = tgbotapi.ModeMarkdown
		_, _ = b.api.Send(edit)
	}

	mediaURL := resp.DownloadURL
	if mediaURL == "" && len(resp.Formats) > 0 {
		mediaURL = resp.Formats[0].URL
	}

	if mediaURL == "" {
		if statusMsgID > 0 {
			edit := tgbotapi.NewEditMessageText(chatID, statusMsgID, "❌ Unable to resolve a direct download stream for this link.")
			_, _ = b.api.Send(edit)
		}
		return
	}

	caption := fmt.Sprintf("🎬 *%s*\n\n📌 *Platform:* %s\n%s",
		resp.Title,
		resp.Platform,
		credit.GetCreditBanner(),
	)
	if len(caption) > 1024 {
		caption = caption[:1020] + "..."
	}

	// Dispatch media (send as video if video platform, or audio if audio requested)
	var sendErr error
	if format == "audio" || format == "mp3" {
		audioMsg := tgbotapi.NewAudio(chatID, tgbotapi.FileURL(mediaURL))
		audioMsg.Caption = caption
		audioMsg.ParseMode = tgbotapi.ModeMarkdown
		audioMsg.ReplyMarkup = credit.GetInlineButtons()
		_, sendErr = b.api.Send(audioMsg)
	} else {
		videoMsg := tgbotapi.NewVideo(chatID, tgbotapi.FileURL(mediaURL))
		videoMsg.Caption = caption
		videoMsg.ParseMode = tgbotapi.ModeMarkdown
		videoMsg.ReplyMarkup = credit.GetInlineButtons()
		if resp.Thumbnail != "" {
			videoMsg.Thumb = tgbotapi.FileURL(resp.Thumbnail)
		}
		if resp.Duration > 0 {
			videoMsg.Duration = resp.Duration
		}
		_, sendErr = b.api.Send(videoMsg)
	}

	// If direct send by URL failed (e.g. file size over 20MB limit for URL upload), send direct download button
	if sendErr != nil {
		fallbackMsg := tgbotapi.NewMessage(chatID, fmt.Sprintf("📥 *Download Ready:*\n\n🎬 *Title:* %s\n📌 *Platform:* %s\n\nTelegram API restricted direct upload (file may be large). Click below to download directly:", resp.Title, resp.Platform))
		fallbackMsg.ParseMode = tgbotapi.ModeMarkdown
		fallbackMsg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonURL("⬇️ Download Media File", mediaURL),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonURL("👨‍💻 Developer", credit.GetDomain()),
				tgbotapi.NewInlineKeyboardButtonURL("📢 Updates", "https://t.me/STDBOTS"),
			),
		)
		_, _ = b.api.Send(fallbackMsg)
	}

	// Remove status indicator message
	if statusMsgID > 0 {
		del := tgbotapi.NewDeleteMessage(chatID, statusMsgID)
		_, _ = b.api.Send(del)
	}

	// Record database log
	if b.db != nil && user != nil {
		_ = b.db.Downloads.LogDownload(ctx, database.DownloadLog{
			UserID:   user.ID,
			Platform: resp.Platform,
			URL:      targetURL,
			Title:    resp.Title,
			Format:   resp.Quality,
		})
	}
}

// CleanURL cleans trailing query parameters if needed
func CleanURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	return u.String()
}
