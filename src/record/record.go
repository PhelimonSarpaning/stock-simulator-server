package record

import (
	"fmt"
	"time"

	"github.com/ThisWillGoWell/stock-simulator-server/src/models"

	"github.com/ThisWillGoWell/stock-simulator-server/src/id"

	"github.com/ThisWillGoWell/stock-simulator-server/src/log"

	"github.com/ThisWillGoWell/stock-simulator-server/src/sender"

	"github.com/ThisWillGoWell/stock-simulator-server/src/change"

	"github.com/ThisWillGoWell/stock-simulator-server/src/wires"

	"github.com/ThisWillGoWell/stock-simulator-server/src/lock"
)

var recordsLock = lock.NewLock("records")
var books = make(map[string]*Book)
var records = make(map[string]*Record)
var portfolioBooks = make(map[string][]*Book)

const EntryIdentifiableType = "record_entry"
const BookIdentifiableType = "record_book"
const BuyRecordType = "buy"
const SellRecordType = "sell"

//type Record interface {
//	GetType() string
//	GetId() string
//	GetTime() time.Time
//	GetRecordType() string
//}

type Book struct {
	Uuid          string            `json:"uuid"`
	LedgerUuid    string            `json:"ledger_uuid"`
	PortfolioUuid string            `json:"portfolio_uuid"`
	ActiveRecords []ActiveBuyRecord `json:"active_records" change:"-"`
}

type ActiveBuyRecord struct {
	RecordUuid string
	AmountLeft int64
}

type Record struct {
	models.Record
}

//func (br *BuyRecord) GetTime() time.Time {
//	return br.Time
//}
//func (*BuyRecord) GetRecordType() string {
//	return BuyRecordType
//}

//type SellRecord struct {
//	Uuid       string `json:"uuid"`
//	SharePrice int64  `json:"share_price"`
//	ShareCount     int64  `json:"amount"`
//}

func NewRecord(recordBookUuid string, amount, sharePrice, taxes, fees, bonus, result int64) (*Record, *Book) {
	uuid := id.SerialUuid()
	return MakeRecord(uuid, recordBookUuid, amount, sharePrice, taxes, fees, bonus, result, time.Now(), true), books[recordBookUuid]
}

func DeleteRecord(uuid string, lockAcquired bool) {
	if !lockAcquired {
		recordsLock.Acquire("delete-record")
		defer recordsLock.Release()
	}
	r, ok := records[uuid]
	if !ok {
		log.Log.Warnf("got delete for reord but cant find uuid=%s", uuid)
		return
	}
	// remove from db first
	delete(records, r.Uuid)

	book, ok := books[r.RecordBookUuid]
	if !ok {
		log.Log.Errorf("got delete for a record but there was no book %s", r.RecordBookUuid)
		return
	}
	//remove the record from the book
	book.ActiveRecords = book.ActiveRecords[:len(book.ActiveRecords)-1]
	if r.ShareCount < 0 { // we have a sell, need to readd those those

	}
	id.RemoveUuid(r.Uuid)

	//for i, r := range book.ActiveRecords {
	//	if r.RecordUuid == r.RecordUuid {
	//		remove = i
	//		continue
	//	}
	//}
	//if remove != -1 {
	//	// Remove the element at index i from a.
	//	copy(book.ActiveRecords[remove:], book.ActiveRecords[remove+1:])       // Shift a[i+1:] left one index.
	//	book.ActiveRecords[len(book.ActiveRecords)-1] = ActiveBuyRecord{}      // Erase last element (write zero value).
	//	book.ActiveRecords = book.ActiveRecords[:len(book.ActiveRecords)-1] // Truncate slice.
	//
	//	removedRecord := book.ActiveRecords[remove]
	//	book.ActiveRecords[remove] = book.ActiveRecords[len(book.ActiveRecords)-1]
	//	book.ActiveRecords = book.ActiveRecords[len(book.ActiveRecords)-1:]
	//} else {
	//	log.Log.Printf("did not find buy record=%s for book=%s", r.Uuid, r.RecordBookUuid)

}

func DeleteRecordBook(uuid string) {
	// is called when a ledger fails to make, must delete the record book
	recordsLock.Acquire("delete-record-book")
	defer recordsLock.Release()
	b, ok := books[uuid]
	if !ok {
		log.Log.Warnf("got delete for record book that we dont know uuid=%s", uuid)
	}
	delete(books, uuid)
	if _, ok := portfolioBooks[b.PortfolioUuid]; ok {
		remove := -1
		for i, l := range portfolioBooks[b.PortfolioUuid] {
			if l.Uuid == uuid {
				remove = i
				break
			}
		}
		if remove != -1 {
			portfolioBooks[b.PortfolioUuid][remove] = portfolioBooks[b.PortfolioUuid][len(portfolioBooks[b.PortfolioUuid])-1]
			portfolioBooks[b.PortfolioUuid] = portfolioBooks[b.PortfolioUuid][:len(portfolioBooks[b.PortfolioUuid])-1]
		} else {
			log.Log.Printf("did not find delete record book=%s for portfolio=%s", uuid, b.PortfolioUuid)
		}
	}
}

func MakeBook(uuid, ledgerUuid, portfolioUuid string) error {

	book := &Book{
		Uuid:          uuid,
		LedgerUuid:    ledgerUuid,
		PortfolioUuid: portfolioUuid,
		ActiveRecords: make([]ActiveBuyRecord, 0),
	}
	bookChange := make(chan interface{})
	if err := change.RegisterPrivateChangeDetect(book, bookChange); err != nil {
		return err
	}
	books[uuid] = book
	if _, ok := portfolioBooks[portfolioUuid]; !ok {
		portfolioBooks[portfolioUuid] = make([]*Book, 0)
	}
	portfolioBooks[portfolioUuid] = append(portfolioBooks[portfolioUuid], books[uuid])

	sender.RegisterChangeUpdate(portfolioUuid, bookChange)
	sender.SendNewObject(portfolioUuid, books[uuid])
	id.RegisterUuid(uuid, books[uuid])
	return nil
}

func MakeRecord(uuid, recordBookUuid string, amount, sharePrice, taxes, fees, bonus, result int64, t time.Time, lockAcquired bool) *Record {
	if !lockAcquired {
		recordsLock.Acquire("new-record")
		defer recordsLock.Release()
	}

	book, ok := books[recordBookUuid]
	if !ok {
		panic("record book not found for a record, NO!")
	}
	newRecord := &Record{
		Record: models.Record{
			Uuid:           uuid,
			SharePrice:     sharePrice,
			Time:           t,
			ShareCount:     amount,
			RecordBookUuid: recordBookUuid,
			Fees:           fees,
			Bonus:          bonus,
			Result:         result,
			Taxes:          taxes,
		},
	}
	records[uuid] = newRecord
	if amount > 0 {
		book.ActiveRecords = append(book.ActiveRecords, ActiveBuyRecord{RecordUuid: uuid, AmountLeft: amount})
	} else {
		walkRecords(book, amount*-1, true)
	}
	id.RegisterUuid(uuid, newRecord)
	wires.BookUpdate.Offer(book)
	sender.SendNewObject(book.PortfolioUuid, newRecord)
	return newRecord
}

// ok so how does this work?
// start with book: current booke
// shares: the number of shares is the (-) of the total shares
// mark: do you actually commit the wirte to the data
// this returns the total amount of $$ we have for all the stocks we have bought
// so we can ask "If I were to sell my 10 shares, that came from 5 different purchases, we can see how much
func walkRecords(book *Book, shares int64, mark bool) int64 {
	amountCleared := 0
	lastAmountCleared := int64(0)
	sharesLeft := shares
	totalCost := int64(0)
	for sharesLeft != 0 {
		if amountCleared >= len(book.ActiveRecords) {
			fmt.Println("WRONG")
		}
		lastAmountCleared = sharesLeft
		activeBuyRecord := book.ActiveRecords[amountCleared]
		record := records[activeBuyRecord.RecordUuid]
		removedShares := activeBuyRecord.AmountLeft

		if activeBuyRecord.AmountLeft > sharesLeft {
			removedShares = sharesLeft
			sharesLeft = 0
		} else if activeBuyRecord.AmountLeft == sharesLeft {
			lastAmountCleared = 0
			amountCleared += 1
			sharesLeft = 0
		} else {
			sharesLeft = sharesLeft - activeBuyRecord.AmountLeft
			amountCleared += 1
		}
		totalCost += removedShares * record.SharePrice

	}

	if mark {
		book.ActiveRecords = book.ActiveRecords[amountCleared:] // remove any that we have
		if len(book.ActiveRecords) != 0 {
			book.ActiveRecords[0].AmountLeft -= lastAmountCleared // remove any remainder off the new count
		}
	}
	return totalCost
}

func GetPrinciple(recordUuid string, shares int64) int64 {
	book := books[recordUuid]
	recordsLock.Acquire("get-principle")
	defer recordsLock.Release()
	return walkRecords(book, shares, false)
}

func GetRecordsForPortfolio(portfolioUuid string) ([]*Book, []*Record) {
	recordsLock.Acquire("get-records")
	defer recordsLock.Release()
	books := portfolioBooks[portfolioUuid]
	portRecord := make([]*Record, 0)

	for _, b := range books {
		for _, active := range b.ActiveRecords {
			portRecord = append(portRecord, records[active.RecordUuid])
		}
	}
	return books, portRecord
}

func GetAllBooks() []*Book {
	recordsLock.Acquire("get-all-books")
	defer recordsLock.Release()
	bookList := make([]*Book, len(books))
	i := 0
	for _, book := range books {
		bookList[i] = book
		i += 1
	}
	return bookList
}
func GetAllRecords() []*Record {
	recordsLock.Acquire("get-all-books")
	defer recordsLock.Release()
	recordList := make([]*Record, len(records))
	i := 0
	for _, record := range records {
		recordList[i] = record
		i += 1
	}
	return recordList
}

func (*Record) GetType() string {
	return EntryIdentifiableType
}
func (br *Record) GetId() string {
	return br.Uuid
}

func (*Book) GetType() string {
	return BookIdentifiableType
}

func (b *Book) GetId() string {
	return b.Uuid
}
