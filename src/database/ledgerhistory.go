package database

import (
	"errors"

	"github.com/stock-simulator-server/src/ledger"
)

var (
	ledgerHistoryTableName            = `ledger_history`
	ledgerHistoryTableCreateStatement = `CREATE TABLE IF NOT EXISTS ` + ledgerHistoryTableName +
		`( ` +
		`time TIMESTAMPTZ NOT NULL,` +
		`uuid text NOT NULL,` +
		`portfolio_id text NOT NULL, ` +
		`stock_id text NOT NULL, ` +
		`amount int NULL` +
		`);`
	ledgerHistoryTSInit = `CREATE EXTENSION IF NOT EXISTS timescaledb CASCADE; SELECT create_hypertable('` + ledgerHistoryTableName + `', 'time');`

	ledgerHistoryTableUpdateInsert = `INSERT INTO ` + ledgerHistoryTableName + `(time, uuid, portfolio_id, stock_id, amount) values (NOW(),$1, $2, $3, $4)`

	//getCurrentPrice()
	validLedgerFields = map[string]bool{
		"amount": true,
	}
)

func initLedgerHistory() {
	tx, err := ts.Begin()
	if err != nil {
		ts.Close()
		panic("could not begin portfolio init: " + err.Error())
	}
	_, err = tx.Exec(ledgerHistoryTableCreateStatement)
	if err != nil {

	}
	tx.Commit()
	tx, err = ts.Begin()
	_, err = tx.Exec(ledgerHistoryTSInit)
	if err != nil {

	}
	tx.Commit()
}

func writeLedgerHistory(entry *ledger.Entry) {
	tx, err := ts.Begin()
	if err != nil {
		ts.Close()
		panic("could not begin portfolio init: " + err.Error())
	}
	_, err = tx.Exec(ledgerHistoryTableUpdateInsert, entry.Uuid, entry.PortfolioId, entry.StockId, entry.Amount)
	if err != nil {
		tx.Rollback()
		panic("error occurred while insert ledger in table " + err.Error())
	}
	tx.Commit()
}
func MakeLedgerHistoryTimeQuery(uuid, timeLength, field, intervalLength string) ([][]interface{}, error) {
	if _, valid := validLedgerFields[field]; !valid {
		return nil, errors.New("not valid choice")
	}
	return MakeHistoryTimeQuery(ledgerHistoryTableName, uuid, timeLength, field, intervalLength)

}

func MakeLedgerHistoryLimitQuery(uuid, field string, limit int) ([][]interface{}, error) {
	if _, valid := validLedgerFields[field]; !valid {
		return nil, errors.New("not valid choice")
	}
	return MakeHistoryLimitQuery(ledgerHistoryTableName, uuid, field, limit)
}
