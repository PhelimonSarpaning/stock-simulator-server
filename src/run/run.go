package run

import (
	"io"
	"io/ioutil"
	"os"

	"github.com/stock-simulator-server/src/config"
	"github.com/stock-simulator-server/src/effect"

	"github.com/stock-simulator-server/src/metics"

	"github.com/stock-simulator-server/src/alert"
	"github.com/stock-simulator-server/src/change"
	"github.com/stock-simulator-server/src/database"
	"github.com/stock-simulator-server/src/histroy"
	"github.com/stock-simulator-server/src/log"
	"github.com/stock-simulator-server/src/order"
	"github.com/stock-simulator-server/src/sender"
	"github.com/stock-simulator-server/src/session"
	"github.com/stock-simulator-server/src/valuable"
	"github.com/stock-simulator-server/src/web"
	"github.com/stock-simulator-server/src/wires"
)

func App() {
	disableDb := os.Getenv("DISABLE_DB") == "True"
	disableDbWrite := os.Getenv("DISABLE_DB_WRITE") == "True"
	serveLog := os.Getenv("SERVE_FRONTEND") == "True"
	seedDb := os.Getenv("SEED_ON_START") == "True"

	metrics.RunMetrics()

	config.LoadConfigs()

	//start DB
	if !disableDb {
		database.InitDatabase(disableDbWrite)
	}
	if serveLog {
		filepath := os.Getenv("FILE_SERVE")
		web.ServePath(filepath)
	}
	//valuable.ValuablesLock.EnableDebug()
	//ledger.EntriesLock.EnableDebug()
	discordAlertToken := os.Getenv("DISCORD_TOKEN")
	var alertWriter io.Writer
	if discordAlertToken != "" {
		alertWriter = alert.Init(discordAlertToken, "504397270075179029")
	} else {
		// if there is discord token, discard all alerts
		alertWriter = ioutil.Discard
	}
	log.Init(alertWriter)
	log.Alerts.Info("Starting App")
	log.Log.Info("Starting App")

	//Wiring of system
	wires.ConnectWires()
	//this takes the subscribe output and converts it to a message
	change.StartDetectChanges()
	session.StartSessionCleaner()
	sender.RunGlobalSender()
	histroy.RunCacheUpdater()
	effect.RunEffectCleaner()

	order.Run()
	valuable.StartStockStimulation()

	if seedDb {
		go config.Seed()
	}
	//go app.LoadVars()
	go web.StartHandlers()
}