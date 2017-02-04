// Command.go
package main



func handleHi (bot tgapi.BotAPI) {
    msg := tgbotapi.NewMessage(cmd.Chat.ID, "Hello to you too!")
    bot.Send(msg)
}
