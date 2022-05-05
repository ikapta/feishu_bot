package bot_test

import (
	"fmt"
	"testing"

	"github.com/ikapta/bot"
)

var testWh = "95f3e056-56d6-4ede-8f18-30b8107ff69b"
var myBot = bot.NewBot(testWh, "")

func TestPushText(t *testing.T) {
  myBot.SendText("Hello world." + bot.Strikethrough("abc"))
}

func TestPushTextAtAll(t *testing.T) {
  myBot.SendText(fmt.Sprintf("Hello world. %s", bot.AtAllInPost()))
}

func TestDeployDevPushTextLink(t *testing.T) {
  myBot.SendText("测试环境已发布，" + bot.TextLink("点击去测试", "https://kaifa.baidu.com/"))
}

func TestDeployProdPushTextLink(t *testing.T) {
  myBot.SendText(bot.Bold("生产环境") + "已发布，" + bot.TextLink("点击去测试", "https://kaifa.baidu.com/"))
}

func TestAtByEmail(t *testing.T) {
  myBot.SendText("测试环境已发布，" + bot.AtUserInPost("ou_xxxx_open_id"))
}

func TestPushImage(t *testing.T) {
  myBot.SendImage("img_7ea74629-9191-4176-998c-2e603c9c5e8g")
}

func TestSendCardDev(t *testing.T) {
  myBot.SendCard(
    bot.BgColorBlue, nil,
    bot.WithCard(
      bot.LangChinese,
      "[DEV]发布提醒",
      bot.WithCardElementMarkdown("测试环境发布成功，相关研发人员注意检查回归测试。👉[点击去测试](https://kaifa.baidu.com/)"),
    ),
  )
}

func TestSendCardProd(t *testing.T) {
  myBot.SendCard(
    bot.BgColorGreen, nil,
    bot.WithCard(
      bot.LangChinese,
      "[PROD]发布提醒",
      bot.WithCardElementMarkdown(fmt.Sprintf("生产环境发布成功，相关研发人员注意检查回归测试。%s", bot.AtAllInCard())),
      bot.WithCardElementMarkdown("👉[点击去测试](https://kaifa.baidu.com/)"),
    ),
  )
}

