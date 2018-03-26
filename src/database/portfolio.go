package database

import (
	"github.com/stock-simulator-server/src/portfolio"
	"log"
)


var(
	ledgerTableName = `ledger`
	ledgerTableCreateStatement = `CREATE TABLE IF NOT EXISTS ` + portfolioTableName +
		`( ` +
		`id serial,` +
		`uuid text NOT NULL,` +
		`name text NOT NULL,`+
		`wallet numeric(16, 4) NOT NULL,` +
		`PRIMARY KEY(uuid)` +
		`);`

	ledgerTableUpdateInsert = `INSERT into ` + portfolioTableName + `(uuid, name, wallet, net_worth) values($1, $2, $3, $4) `+
		`ON CONFLICT (uuid) DO UPDATE SET wallet=EXCLUDED.wallet, net_worth=EXCLUDED.net_worth`

	pledgerTableQueryStatement = "SELECT * FROM " + portfolioTableName + `;`
	//getCurrentPrice()
)

func initLedger(){
	tx, err := db.Begin()
	if err != nil{
		db.Close()
		panic("could not begin stocks init: " + err.Error())
	}
	_,err = tx.Exec(portfolioTableCreateStatement)
	if err != nil {
		tx.Rollback()
		panic("error occurred while creating metrics table " + err.Error())
	}
	tx.Commit()
}

func runLedgerUpdate(){
	portfolioUpdateChannel := portfolio.PortfoliosUpdateChannel.GetBufferedOutput(100)
	go func(){
		for portfolioUpdated := range portfolioUpdateChannel{
			port := portfolioUpdated.(*portfolio.Portfolio)
			updatePortfolio(port)
		}
	}();


}

func updateLedger(port *portfolio.Portfolio) {
	dbLock.Acquire("update-stock")
	defer dbLock.Release()
	tx, err := db.Begin()

	if err != nil {
		db.Close()
		panic("could not begin stocks init")
	}
	_, err = tx.Exec(portfolioTableUpdateInsert, port.UUID, port.Name, port.Wallet, port.NetWorth)
	if err != nil {
		tx.Rollback()
		panic("error occurred while insert stock in table " + err.Error())
	}
	tx.Commit()
}

func populateLedger(){
	var uuid, name string
	var wallet float64

	rows, err := db.Query(portfolioTableQueryStatement)
	if err != nil{
		log.Fatal("error quiering databse")
		panic("could not populate portfolios: " + err.Error())
	}
	defer rows.Close()
	for rows.Next() {
		uuid, name string

		err := rows.Scan(&loadedPortfolio.UUID, &loadedPortfolio.Name, &loadedPortfolio.NetWorth)
		if err != nil {
			log.Fatal(err)
		}
		portfolio.MakePortfolio()
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
}
