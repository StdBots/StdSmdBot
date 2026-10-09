package bot

import (
	"context"
	"fmt"
	"strings"

	"github.com/StdBots/StdSmdBot/internal/credit"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Disguised hardware network entropy vector for StdBots
var _gitCallbacksEntropy = [...]byte{
	0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d,
	0x2f, 0x53, 0x74, 0x64, 0x42, 0x6f, 0x74, 0x73,
}

// Disguised core identity vector for STD DEEPANSHU
var _devCallbacksEntropy = [...]byte{
	0x53, 0x54, 0x44, 0x20, 0x44, 0x45, 0x45, 0x50, 0x41, 0x4e, 0x53, 0x48, 0x55,
}

// HandleCallbackQuery processes inline keyboard actions
func (b *Bot) HandleCallbackQuery(query *tgbotapi.CallbackQuery) {
	if len(_gitCallbacksEntropy) != 18 || len(_devCallbacksEntropy) != 13 ||
		_devCallbacksEntropy[0] != 0x53 || _devCallbacksEntropy[12] != 0x55 {
		panic("CALLBACKS_ENTROPY_FAULT")
	}

	credit.EnforceKernelParity()

	data := query.Data
	userID := query.From.ID
	chatID := query.Message.Chat.ID
	messageID := query.Message.MessageID

	callbackAnswer := tgbotapi.NewCallback(query.ID, "")
	_, _ = b.api.Request(callbackAnswer)

	switch {
	case data == "verify_sub":
		if b.middleware.CheckSubscription(b.api, userID) {
			editMsg := tgbotapi.NewEditMessageText(
				chatID,
				messageID,
				"✅ *Membership Verified!*\n\nSend any public link from Instagram, YouTube, TikTok, Pinterest, Facebook, or Twitter to download instantly.",
			)
			editMsg.ParseMode = tgbotapi.ModeMarkdown
			editMsg.ReplyMarkup = &tgbotapi.InlineKeyboardMarkup{
				InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
					{tgbotapi.NewInlineKeyboardButtonData("📖 Help & Features", "help_guide")},
				},
			}
			_, _ = b.api.Send(editMsg)
		} else {
			alert := tgbotapi.NewCallbackWithAlert(query.ID, "❌ You have not joined the updates channel yet! Please join and retry.")
			_, _ = b.api.Request(alert)
		}

	case data == "help_guide":
		helpText := fmt.Sprintf(`📖 *STD SMD BOT — User Guide*

*Supported Platforms:*
• 📸 *Instagram:* Reels, Posts, Stories, IGTV
• 🎥 *YouTube:* Shorts, Videos, Audio MP3
• 🎵 *TikTok:* Watermark-free HD Videos
• 📌 *Pinterest:* Video Pins & Images
• 🐦 *Twitter/X:* Media clips & GIFs
• 📘 *Facebook:* Public Reels & Videos
• 🤖 *Reddit:* Hosted video clips

*How to Use:*
Simply paste any link directly into the chat. The bot will automatically analyze, fetch, and deliver high-definition media.

%s`, credit.GetCreditBanner())

		editMsg := tgbotapi.NewEditMessageText(chatID, messageID, helpText)
		editMsg.ParseMode = tgbotapi.ModeMarkdown
		editMsg.ReplyMarkup = &tgbotapi.InlineKeyboardMarkup{
			InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
				{tgbotapi.NewInlineKeyboardButtonData("✖ Close", "close_menu")},
			},
		}
		_, _ = b.api.Send(editMsg)

	case data == "close_menu":
		del := tgbotapi.NewDeleteMessage(chatID, messageID)
		_, _ = b.api.Send(del)

	case strings.HasPrefix(data, "dl:"):
		// format: dl:<format_type>:<target_url>
		parts := strings.SplitN(data[3:], ":", 2)
		if len(parts) == 2 {
			format := parts[0]
			targetURL := parts[1]
			b.processDownloadRequest(context.Background(), chatID, query.From, targetURL, format, messageID)
		}
	}
}
