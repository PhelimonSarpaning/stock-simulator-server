package database

import (
	"database/sql"
	"fmt"

	"github.com/ThisWillGoWell/stock-simulator-server/src/models"

	"github.com/ThisWillGoWell/stock-simulator-server/src/ledger"
)

var (
	ledgerTableName            = `ledger`
	ledgerTableCreateStatement = `CREATE TABLE IF NOT EXISTS ` + ledgerTableName +
		`( ` +
		`id serial,` +
		`uuid text NOT NULL,` +
		`portfolio_id text NOT NULL,` +
		`stock_id text NOT NULL,` +
		`amount bigint NOT NULL,` +
		`record_id text NOT NULL, ` +
		`PRIMARY KEY(uuid)` +
		`);`

	ledgerTableUpdateInsert = `INSERT into ` + ledgerTableName + `(uuid, portfolio_id, record_id, stock_id, amount ) values($1, $2, $3, $4, $5) ` +
		`ON CONFLICT (uuid) DO UPDATE SET amount=EXCLUDED.amount`

	ledgerTableQueryStatement = "SELECT uuid, portfolio_id, stock_id, record_id,  amount FROM " + ledgerTableName + `;`

	ledgerTableDeleteStatement = "DELETE from " + ledgerTableName + `WHERE uuid = $1`
)

func (d *Database) InitLedger() error {
	return d.Exec("ledgers-init", ledgerTableCreateStatement)
}

func writeLedger(entry models.Ledger, tx *sql.Tx) error {
	_, err := tx.Exec(ledgerTableUpdateInsert, entry.Uuid, entry.PortfolioId, entry.RecordBookId, entry.StockId, entry.Amount)
	return err
}

func deleteLedger(entry models.Ledger, tx *sql.Tx) error {
	_, err := tx.Exec(ledgerTableDeleteStatement, uuid)
	return err
}

func (d *Database) PopulateLedger() (map[string]models.Ledger, error) {
	var uuid, portfolioId, stockId, recordId string
	var amount int64

	var rows *sql.Rows
	var err error
	if rows, err = d.db.Query(ledgerTableQueryStatement); err != nil {
		return nil, fmt.Errorf("failed to query portfolio err=[%v]", err)
	}
	defer func() {
		_ = rows.Close()
	}()
	for rows.Next() {
		if err = rows.Scan(&uuid, &portfolioId, &stockId, &recordId, &amount); err != nil {
			return err
		}
		if _, err = ledger.MakeLedgerEntry(uuid, portfolioId, stockId, recordId, amount); err != nil {
			return err
		}
	}
	return rows.Err()
}
