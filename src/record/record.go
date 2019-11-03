package record

import (
	"fmt"
	"time"

	"github.com/ThisWillGoWell/stock-simulator-server/src/database"

	"github.com/ThisWillGoWell/stock-simulator-server/src/log"

	"github.com/ThisWillGoWell/stock-simulator-server/src/sender"

	"github.com/ThisWillGoWell/stock-simulator-server/src/change"

	"github.com/ThisWillGoWell/stock-simulator-server/src/wires"

	"github.com/ThisWillGoWell/stock-simulator-server/src/lock"

	"github.com/ThisWillGoWell/stock-simulator-server/src/utils"
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
	Uuid             string            `json:"uuid"`
	LedgerUuid       string            `json:"ledger_uuid"`
	PortfolioUuid    string            `json:"portfolio_uuid"`
	ActiveBuyRecords []ActiveBuyRecord `json:"buy_records" change:"-"`
}

type ActiveBuyRecord struct {
	RecordUuid string
	AmountLeft int64
}

type Record struct {
	Uuid           string    `json:"uuid"`
	SharePrice     int64     `json:"share_price"`
	ShareCount     int64     `json:"share_count"`
	Time           time.Time `json:"time"`
	RecordBookUuid string    `json:"book_uuid"`
	Fees           int64     `json:"fee"`
	Taxes          int64     `json:"taxes"`
	Bonus          int64     `json:"bonus"`
	Result         int64     `json:"result"`
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

func NewRecord(recordBookUuid string, amount, sharePrice, taxes, fees, bonus, result int64) error {
	uuid := utils.SerialUuid()
	// need to hold the lock to make sure if it fails, we can delete it before another one gets made, messing up the activeBuyRecords
	recordsLock.Acquire("new-record")
	defer recordsLock.Release()

	r, err := MakeRecord(uuid, recordBookUuid, amount, sharePrice, taxes, fees, bonus, result, time.Now(), true)
	if err != nil {
		return fmt.Errorf("failed to make record err=[%v]", err)
	}
	if err := database.Db.WriteRecord(r); err != nil {
		_ = deleteRecord(r)
		return fmt.Errorf("failed to make record err=[%v]", err)
	}
	wires.RecordsNewObject.Offer(r)
	return nil
}

func deleteRecord(r *Record) error {
	log.Log.Printf("deleting record from book uuid=%s", r.RecordBookUuid)
	// this should only get called if the database write fails
	recordsLock.Acquire("delete-record")
	defer recordsLock.Release()
	book, ok := books[r.RecordBookUuid]
	if !ok {
		log.Log.Errorf("got delete for a record but there was no book %s", r.RecordBookUuid)
		return nil
	}
	//remove the record from the book
	// we know its the last record on the book and that it was a buy so no need to rewalk
	book.ActiveBuyRecords = book.ActiveBuyRecords[:len(book.ActiveBuyRecords)-1]
	// attempt to delete even though we know something failed with the db
	// remove from db first
	dbErr := database.Db.DeleteRecord(r)
	delete(records, r.Uuid)
	utils.RemoveUuid(r.Uuid)
	return dbErr

	//for i, r := range book.ActiveBuyRecords {
	//	if r.RecordUuid == r.RecordUuid {
	//		remove = i
	//		continue
	//	}
	//}
	//if remove != -1 {
	//	// Remove the element at index i from a.
	//	copy(book.ActiveBuyRecords[remove:], book.ActiveBuyRecords[remove+1:])       // Shift a[i+1:] left one index.
	//	book.ActiveBuyRecords[len(book.ActiveBuyRecords)-1] = ActiveBuyRecord{}      // Erase last element (write zero value).
	//	book.ActiveBuyRecords = book.ActiveBuyRecords[:len(book.ActiveBuyRecords)-1] // Truncate slice.
	//
	//	removedRecord := book.ActiveBuyRecords[remove]
	//	book.ActiveBuyRecords[remove] = book.ActiveBuyRecords[len(book.ActiveBuyRecords)-1]
	//	book.ActiveBuyRecords = book.ActiveBuyRecords[len(book.ActiveBuyRecords)-1:]
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
			log.Log.Printf("did not find delete record book=%s for protfolio=%s", uuid, b.PortfolioUuid)
		}
	}
}

func MakeBook(uuid, ledgerUuid, portfolioUuid string) error {

	book := &Book{
		Uuid:             uuid,
		LedgerUuid:       ledgerUuid,
		PortfolioUuid:    portfolioUuid,
		ActiveBuyRecords: make([]ActiveBuyRecord, 0),
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
	utils.RegisterUuid(uuid, books[uuid])
	return nil
}

func MakeRecord(uuid, recordBookUuid string, amount, sharePrice, taxes, fees, bonus, result int64, t time.Time, lockAcquired bool) (*Record, error) {
	if !lockAcquired {
		recordsLock.Acquire("new-record")
		defer recordsLock.Release()
	}

	book, ok := books[recordBookUuid]
	if !ok {
		return nil, fmt.Errorf("record book id %s not found for record how?", recordBookUuid)
	}
	newRecord := &Record{
		Uuid:           uuid,
		SharePrice:     sharePrice,
		Time:           t,
		ShareCount:     amount,
		RecordBookUuid: recordBookUuid,
		Fees:           fees,
		Bonus:          bonus,
		Result:         result,
		Taxes:          taxes,
	}
	records[uuid] = newRecord
	if amount > 0 {
		book.ActiveBuyRecords = append(book.ActiveBuyRecords, ActiveBuyRecord{RecordUuid: uuid, AmountLeft: amount})
	} else {
		walkRecords(book, amount*-1, true)
	}
	utils.RegisterUuid(uuid, newRecord)
	wires.BookUpdate.Offer(book)
	sender.SendNewObject(book.PortfolioUuid, newRecord)
	return newRecord, nil
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
		if amountCleared >= len(book.ActiveBuyRecords) {
			fmt.Println("WRONG")
		}
		lastAmountCleared = sharesLeft
		activeBuyRecord := book.ActiveBuyRecords[amountCleared]
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
		book.ActiveBuyRecords = book.ActiveBuyRecords[amountCleared:] // remove any that we have
		if len(book.ActiveBuyRecords) != 0 {
			book.ActiveBuyRecords[0].AmountLeft -= lastAmountCleared // remove any remainder off the new count
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
		for _, active := range b.ActiveBuyRecords {
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
