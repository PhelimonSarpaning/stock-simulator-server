package items

import (
	"testing"
	"time"

	"github.com/stock-simulator-server/src/effect"

	"github.com/stock-simulator-server/src/wires"

	"github.com/stock-simulator-server/src/utils"

	"github.com/stock-simulator-server/src/session"

	"github.com/stock-simulator-server/src/account"
	"github.com/stock-simulator-server/src/order"
	"github.com/stock-simulator-server/src/portfolio"
	"github.com/stock-simulator-server/src/valuable"
)

func MakeUser(name string) string {
	sessionToken, _ := account.NewUser(name, "test", "password")
	userId, _ := session.GetUserId(sessionToken)
	port := portfolio.Portfolios[account.UserList[userId].PortfolioId]
	port.Wallet = 1000000000
	port.Level = 2
	return port.Uuid
}

func populateItemConfig() {
	validItems["1"] = validItemConfiguration{
		Name:          "Test Item - Double Protifts",
		Type:          TradeItemType,
		Cost:          10,
		RequiredLevel: 0,
		Prams: &TradeEffectItem{
			BuyFeeMultiplier: utils.CreateFloat(2),
			Duration: utils.Duration{
				Duration: time.Second * 1,
			},
		},
	}
}

func MakeStocks() {
	valuable.NewStock("TEST1", "Test 3 Account", 1000, time.Second*10)
	valuable.NewStock("TEST2", "Test 2 Account", 1000, time.Second*10)
}

func TestTrade(t *testing.T) {
	effect.RunEffectCleaner()
	wires.PrintAll()
	populateItemConfig()
	portUuid := MakeUser("testuser")
	itemUuid, _ := BuyItem(portUuid, "1")
	Use(itemUuid, portUuid, nil)
	utils.Spin(time.Second * 4)
}

func TestMail(t *testing.T) {
	order.Run()

	u1 := MakeUser("user1")
	u2 := MakeUser("user2")

	//itemId := BuyItem(u1, mailItemType)
	//_, err := Use(itemId, u1, MailItemParameters{To:u2, ShareCount:1000})
	//t.Log(err)
	t.Log(portfolio.Portfolios[account.UserList[u1].PortfolioId].Wallet)
	t.Log(portfolio.Portfolios[account.UserList[u2].PortfolioId].Wallet)
}
