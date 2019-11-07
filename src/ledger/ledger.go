package ledger

import (
	"fmt"

	"github.com/ThisWillGoWell/stock-simulator-server/src/id"

	"github.com/ThisWillGoWell/stock-simulator-server/src/models"

	"github.com/ThisWillGoWell/stock-simulator-server/src/record"

	"github.com/ThisWillGoWell/stock-simulator-server/src/change"

	"github.com/ThisWillGoWell/stock-simulator-server/src/wires"

	"github.com/ThisWillGoWell/stock-simulator-server/src/duplicator"
	"github.com/ThisWillGoWell/stock-simulator-server/src/lock"
)

const objectType = "ledger"

//map of uuid -> entry
var Entries = make(map[string]*Entry)
var EntriesStockPortfolio = make(map[string]map[string]*Entry)
var EntriesPortfolioStock = make(map[string]map[string]*Entry)

// map of stock_uuid -> open shares
var EntriesLock = lock.NewLock("ledger-entries-lock")
var NewEntriesLock = lock.NewLock("new-entries-lock")

/**
Ledgers store who owns what stock
They are all done though uuid strings since that's all that's required
They are stored in two maps
1) given a stock uuid, get all portfolios that own it
2) given a portfolio uuid, get all stocks it owns
*/
type Entry struct {
	models.Ledger
	Lock          *lock.Lock                    `json:"-"`
	UpdateChannel *duplicator.ChannelDuplicator `json:"-"`
}

/**
build a new ledger entry and generate a new uuid for it
takes in the lock acquired since trade already owns the lock for the entries
*/
func NewLedgerEntry(portfolioId, stockId string) (*Entry, error) {
	uuid := id.SerialUuid()
	recordId := id.SerialUuid()
	e, err := MakeLedgerEntry(uuid, portfolioId, stockId, recordId, 0, true)
	if err != nil {
		id.RemoveUuid(uuid)
		id.RemoveUuid(recordId)
		return nil, err
	}

	return e, nil
}

func DeleteLedger(l *Entry, lockAcquired bool) {
	if !lockAcquired {
		NewEntriesLock.Acquire("delete-ledger")
		defer NewEntriesLock.Release()
		EntriesLock.Acquire("delete-ledger")
		defer EntriesLock.Release()
		l.Lock.Acquire("delete-ledger")
		l.Lock.Release()
	}
	delete(Entries, l.Uuid)
	if _, ok := EntriesPortfolioStock[l.PortfolioId]; ok {
		delete(EntriesPortfolioStock[l.PortfolioId], l.Uuid)
	}
	if _, ok := EntriesStockPortfolio[l.PortfolioId]; ok {
		delete(EntriesStockPortfolio[l.PortfolioId], l.Uuid)
	}
	record.DeleteRecordBook(l.Uuid)
	change.UnregisterChangeDetect(l)
	l.UpdateChannel.StopDuplicator()
	id.RemoveUuid(l.Uuid)
}

/**
Make a Ledger
*/
func MakeLedgerEntry(uuid, portfolioId, stockId, recordId string, amount int64, lockAquired bool) (*Entry, error) {
	if !lockAquired {
		EntriesLock.Acquire("make-ledger")
		defer EntriesLock.Release()
	}
	entry := &Entry{
		Ledger: models.Ledger{
			Uuid:         uuid,
			PortfolioId:  portfolioId,
			StockId:      stockId,
			Amount:       amount,
			RecordBookId: recordId,
		},
		UpdateChannel: duplicator.MakeDuplicator(fmt.Sprintf("LedgerEntry-%s", uuid)),
	}
	if err := record.MakeBook(recordId, uuid, portfolioId); err != nil {
		return nil, fmt.Errorf("failed to make recored book for ledger err=[%v]", err)
	}

	if err := change.RegisterPublicChangeDetect(entry); err != nil {
		return nil, err
	}

	if EntriesPortfolioStock[portfolioId] == nil {
		EntriesPortfolioStock[portfolioId] = make(map[string]*Entry)
	}
	EntriesPortfolioStock[portfolioId][stockId] = entry

	if EntriesStockPortfolio[stockId] == nil {
		EntriesStockPortfolio[stockId] = make(map[string]*Entry)
	}

	Entries[uuid] = entry
	EntriesStockPortfolio[stockId][portfolioId] = entry
	entry.UpdateChannel.EnableCopyMode()
	wires.LedgerUpdate.RegisterInput(entry.UpdateChannel.GetOutput())
	id.RegisterUuid(uuid, entry)
	return entry, nil
}

/*
/**
get All ledgers so they can be sent on connection
*/
func GetAllLedgers() []*Entry {
	EntriesLock.Acquire("get-all-ledgers")
	defer EntriesLock.Release()
	lst := make([]*Entry, len(Entries))
	i := 0
	for _, val := range Entries {
		lst[i] = val
		i += 1
	}
	return lst
}

func (ledger *Entry) GetId() string {
	return ledger.Uuid
}

func (ledger *Entry) GetType() string {
	return objectType
}
