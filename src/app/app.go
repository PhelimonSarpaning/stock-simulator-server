package app

import (
	"fmt"
	"github.com/stock-simulator-server/src/messages"
	"math/rand"
	"time"

	"github.com/stock-simulator-server/src/portfolio"
	"github.com/stock-simulator-server/src/valuable"

	"github.com/stock-simulator-server/src/account"
	"github.com/stock-simulator-server/src/client"
	"github.com/stock-simulator-server/src/order"
)



func LoadVars() {
	fmt.Println("running app")
	//make the stocks
	type stockConfig struct {
		id       string
		name     string
		price    int
		duration time.Duration
	}

	stockConfigs := append(make([]stockConfig, 0),
		stockConfig{"CHUNT", "Slut Hut",rand.Intn(10000) , time.Second },
		stockConfig{"ABBV", "AbbVie Inc.", rand.Intn(10000), time.Second * 10 },
		stockConfig{"ABT", "Abbott Laboratories", rand.Intn(10000), time.Second * 10 },
		stockConfig{"ACN", "Accenture plc", rand.Intn(10000), time.Second * 10 },
		stockConfig{"AGN", "Allergan plc", rand.Intn(10000), time.Second * 10 },
		stockConfig{"AIG", "American International Group Inc.", rand.Intn(10000), time.Second * 10 },
		stockConfig{"ALL", "Allstate Corp.", rand.Intn(10000), time.Second * 10 },
		stockConfig{"AMGN", "Amgen Inc.", rand.Intn(10000), time.Second * 10 },
		stockConfig{"AMZN", "Amazon.com", rand.Intn(10000), time.Second * 10 },
		stockConfig{"AXP", "American Express Inc.", rand.Intn(10000), time.Second * 10 },
		stockConfig{"BA", "Boeing Co.", rand.Intn(10000), time.Second * 10 },
		stockConfig{"BAC", "Bank of America Corp", rand.Intn(10000), time.Second * 10 },
		stockConfig{"BIIB", "Biogen Idec", rand.Intn(10000), time.Second * 10 },
		stockConfig{"BK", "The Bank of New York Mellon", rand.Intn(10000), time.Second * 10 },
		stockConfig{"BKNG", "Booking Holdings", rand.Intn(10000), time.Second * 10 },
		stockConfig{"BLK", "BlackRock Inc", rand.Intn(10000), time.Second * 10 },
		stockConfig{"BMY", "Bristol-Myers Squibb", rand.Intn(10000), time.Second * 10 },
		stockConfig{"BRK.B", "Berkshire Hathaway", rand.Intn(10000), time.Second * 10 },
		stockConfig{"C", "Citigroup Inc", rand.Intn(10000), time.Second * 10 },
		stockConfig{"CAT", "Caterpillar Inc", rand.Intn(10000), time.Second * 10 },
		stockConfig{"CELG", "Celgene Corp", rand.Intn(10000), time.Second * 10 },
		stockConfig{"CHTR", "Charter Communications", rand.Intn(10000), time.Second * 10 },
		stockConfig{"CL", "Colgate-Palmolive Co.", rand.Intn(10000), time.Second * 10 },
		stockConfig{"CMCSA", "Comcast Corporation", rand.Intn(10000), time.Second * 10 },
		stockConfig{"COF", "Capital One Financial Corp.", rand.Intn(10000), time.Second * 10 },
		stockConfig{"COP", "ConocoPhillips", rand.Intn(10000), time.Second * 10 },
		stockConfig{"COST", "Costco", rand.Intn(10000), time.Second * 10 },
		stockConfig{"CSCO", "Cisco Systems", rand.Intn(10000), time.Second * 10 },
		stockConfig{"CVS", "CVS Health", rand.Intn(10000), time.Second * 10 },
		stockConfig{"CVX", "Chevron", rand.Intn(10000), time.Second * 10 },
		stockConfig{"DHR", "Danaher", rand.Intn(10000), time.Second * 10 },
		stockConfig{"DIS", "The Walt Disney Company", rand.Intn(10000), time.Second * 10 },
		stockConfig{"DUK", "Duke Energy", rand.Intn(10000), time.Second * 10 },
		stockConfig{"DWDP", "DowDuPont", rand.Intn(10000), time.Second * 10 },
		stockConfig{"EMR", "Emerson Electric Co.", rand.Intn(10000), time.Second * 10 },
		stockConfig{"EXC", "Exelon", rand.Intn(10000), time.Second * 10 },
		stockConfig{"F", "Ford Motor", rand.Intn(10000), time.Second * 10 },
		stockConfig{"FB", "Facebook", rand.Intn(10000), time.Second * 10 },
		stockConfig{"FDX", "FedEx", rand.Intn(10000), time.Second * 10 },
		stockConfig{"FOX", "21st Century Fox", rand.Intn(10000), time.Second * 10 },
		stockConfig{"FOXA", "21st Century Fox", rand.Intn(10000), time.Second * 10 },
		stockConfig{"GD", "General Dynamics", rand.Intn(10000), time.Second * 10 },
		stockConfig{"GE", "General Electric Co.", rand.Intn(10000), time.Second * 10 },
		stockConfig{"GILD", "Gilead Sciences", rand.Intn(10000), time.Second * 10 },
		stockConfig{"GM", "General Motors", rand.Intn(10000), time.Second * 10 },
		stockConfig{"GOOG", "Alphabet Inc", rand.Intn(10000), time.Second * 10 },
		stockConfig{"GOOGL", "Alphabet Inc", rand.Intn(10000), time.Second * 10 },
		stockConfig{"GS", "Goldman Sachs", rand.Intn(10000), time.Second * 10 },
		stockConfig{"HAL", "Halliburton", rand.Intn(10000), time.Second * 10 },
		stockConfig{"HD", "Home Depot", rand.Intn(10000), time.Second * 10 },
		stockConfig{"HON", "Honeywell", rand.Intn(10000), time.Second * 10 },
		stockConfig{"IBM", "International Business Machines", rand.Intn(10000), time.Second * 10 },
	)

	//Make an exchange //exchanger, _ := exchange.BuildExchange("US")
	//#exchanger.StartExchange()

	//Register stocks with Exchange
	for _, ele := range stockConfigs {
		valuable.NewStock(ele.id, ele.name, int64(ele.price), ele.duration)
	}
	fmt.Println("done adding stocks")

	//start the builder
	//go client.BroadcastMessageBuilder()
	//build and simulate a client
	account.NewUser("Mike", "pass")
	account.NewUser("Will", "pass")

	token, _ := account.ValidateUser("Will", "pass")
	acc, _ := account.ConnectUser(token)
	portfolio.Portfolios[acc.PortfolioId].Wallet = 1000000


	accs := []string{acc.PortfolioId}
	users := []string{acc.Uuid}

	for id := range valuable.Stocks {
		for i:=0; i<50; i++{
			portId := accs[0]
			po2 := order.MakePurchaseOrder(id, portId, 1)
			client.SendToUser(users[0],messages.BuildResponseMsg( <-po2.ResponseChannel, fmt.Sprintf("backend-trade-%d", i)))
			<-time.After(time.Second * 30)
		}
	}

}
