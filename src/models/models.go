package models

import (
	"time"

	"github.com/ThisWillGoWell/stock-simulator-server/src/utils"
)

type Portfolio struct {
	UserUUID string `json:"user_uuid"`
	Uuid     string `json:"uuid"`
	Wallet   int64  `json:"wallet" change:"-"`
	NetWorth int64  `json:"net_worth" change:"-"`
	Level    int64  `json:"level" change:"-"`
}

type Effect struct {
	PortfolioUuid string         `json:"portfolio_uuid"`
	Uuid          string         `json:"uuid"`
	Title         string         `json:"title" change:"-"`
	Duration      utils.Duration `json:"duration"`
	StartTime     time.Time      `json:"time"`
	Type          string         `json:"type"`
	InnerEffect   interface{}    `json:"-" change:"inner"`
	Tag           string         `json:"tag"`
}

type Stock struct {
	Uuid           string        `json:"uuid"`
	Name           string        `json:"name"`
	TickerId       string        `json:"ticker_id"`
	CurrentPrice   int64         `json:"current_price" change:"-"`
	OpenShares     int64         `json:"open_shares" change:"-"`
	ChangeDuration time.Duration `json:"-"`
}

type User struct {
	UserName      string                 `json:"-"`
	Password      string                 `json:"-"`
	DisplayName   string                 `json:"display_name" change:"-"`
	Uuid          string                 `json:"-"`
	PortfolioId   string                 `json:"portfolio_uuid"`
	Active        bool                   `json:"active" change:"-"`
	Config        map[string]interface{} `json:"-"`
	ConfigStr     string                 `json:"-"`
	ActiveClients int64                  `json:"-"`
}

type Item struct {
	Uuid            string      `json:"uuid"`
	Name            string      `json:"name"`
	ConfigId        string      `json:"config"`
	Type            string      `json:"type"`
	PortfolioUuid   string      `json:"portfolio_uuid"`
	CreateTime      time.Time   `json:"create_time"`
	InnerItemString string      `json:"-"`
	InnerItem       interface{} `json:"-" change:"inner"`
}

type Ledger struct {
	Uuid         string `json:"uuid"`
	PortfolioId  string `json:"portfolio_id"`
	StockId      string `json:"stock_id"`
	Amount       int64  `json:"amount" change:"-"`
	RecordBookId string `json:"record_book"`
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

type Notification struct {
	Uuid          string      `json:"uuid"`
	PortfolioUuid string      `json:"portfolio_uuid"`
	Timestamp     time.Time   `json:"time"`
	Type          string      `json:"type"`
	Notification  interface{} `json:"notification"`
	Seen          bool        `json:"seen"`
}
