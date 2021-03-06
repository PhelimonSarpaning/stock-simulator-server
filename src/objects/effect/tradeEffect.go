package effect

import (
	"fmt"
	"time"

	"github.com/ThisWillGoWell/stock-simulator-server/src/id"
	"github.com/ThisWillGoWell/stock-simulator-server/src/objects"

	"github.com/ThisWillGoWell/stock-simulator-server/src/app/log"

	"github.com/ThisWillGoWell/stock-simulator-server/src/utils"

	"github.com/ThisWillGoWell/stock-simulator-server/src/game/money"
)

const TradeEffectType = "trade"
const baseTradeEffectTag = "base_trade"

const BaseTaxRate = 0.15
const BaseSellFell = 20 * money.One
const BaseBuyFee = 20 * money.One

type TradingType struct {
}

func (TradingType) Name() string {
	return TradeEffectType
}

type TradeEffect struct {
	parentEffect *Effect `json:"-"`

	BuyFeeAmount     *int64   `json:"buy_fee_amount"`     // fee on all trades ex: base fee
	BuyFeeMultiplier *float64 `json:"buy_fee_multiplier"` // fee % of the total fees, ex: double fees

	SellFeeAmount     *int64   `json:"sell_fee_amount"`     // fee on all sales trades ex: base fee
	SellFeeMultiplier *float64 `json:"sell_fee_multiplier"` // fee % of the total fees, ex: double fees on trades

	BonusProfitMultiplier *float64 `json:"profit_multiplier"` // current profit multiplier, ex: bonus

	TaxPercent    *float64 `json:"tax_percent"`    // tax payed on profits, ex:  base tax
	TaxMultiplier *float64 `json:"tax_multiplier"` // percent multiplier, ex: taxless sales

	TradeBlocked *bool `json:"trade_blocked"` // if trade is blocked
}

func NewBaseTradeEffect(portfolioUuid string) (*Effect, error) {
	baseTradeEffect := &TradeEffect{
		BuyFeeAmount:  utils.CreateInt(BaseBuyFee),
		SellFeeAmount: utils.CreateInt(BaseSellFell),
		TaxPercent:    utils.CreateFloat(BaseTaxRate),
	}
	var err error
	// should be no current base effect
	baseTradeEffect.parentEffect, _, err = newEffect(portfolioUuid, "Base Effect", TradeEffectType, baseTradeEffectTag, baseTradeEffect, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to make base effect err=[%v]", err)
	}
	return baseTradeEffect.parentEffect, nil
}

func UpdateBaseProfit(portfolioUuid string, profitMultiplier float64) (*Effect, *Effect, error) {
	effect := getTaggedEffect(portfolioUuid, baseTradeEffectTag)

	if effect == nil {
		return nil, nil, fmt.Errorf("there was no base effect?? port=[%v]", portfolioUuid)
	}

	newEffectObject := utils.Copy(effect.Effect).(objects.Effect)
	newEffectObject.Uuid = id.SerialUuid()
	newEffectObject.InnerEffect.(*TradeEffect).BonusProfitMultiplier = utils.CreateFloat(profitMultiplier)

	newEffect, err := MakeEffect(newEffectObject, true)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to make new effect err=[%v]", err)
	}

	return newEffect, effect, nil
}

func NewTradeEffect(portfolioUuid, title, tag string, effect *TradeEffect, duration time.Duration) (*Effect, *Effect, error) {
	var err error
	var deletedEffect *Effect
	effect.parentEffect, deletedEffect, err = newEffect(portfolioUuid, title, TradeEffectType, tag, effect, duration)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to make trade effect err=[%v]", err)
	}
	return effect.parentEffect, deletedEffect, nil
}

func NewTaxModifier(portfolioUuid, title string, duration time.Duration, taxMultiplier float64) (*Effect, *Effect, error) {
	var err error
	newTradeEffect := &TradeEffect{
		TaxMultiplier: &taxMultiplier,
	}
	var deleteEffect *Effect
	newTradeEffect.parentEffect, deleteEffect, err = newEffect(portfolioUuid, title, "", TradeEffectType, newTradeEffect, duration)
	if err != nil {
		log.Log.Errorf("failed to make tax effect err=[%v]", err)
		return nil, nil, fmt.Errorf("failed to make effect")
	}

	return newTradeEffect.parentEffect, deleteEffect, nil
}

// Calculate the total bonus  for a portfolio
func TotalTradeEffect(portfolioUuid string) (*TradeEffect, []string) {
	EffectLock.Acquire("TotalBonus")
	defer EffectLock.Release()
	effect, ok := portfolioEffects[portfolioUuid]
	totalEffect := &TradeEffect{
		BuyFeeAmount:     utils.CreateInt(0),
		BuyFeeMultiplier: utils.CreateFloat(1),

		SellFeeAmount:     utils.CreateInt(0),
		SellFeeMultiplier: utils.CreateFloat(1),

		BonusProfitMultiplier: utils.CreateFloat(0),

		TaxPercent:    utils.CreateFloat(0),
		TaxMultiplier: utils.CreateFloat(1),

		TradeBlocked: utils.CreateBool(false),
	}
	uuids := make([]string, 0)
	if !ok {
		return totalEffect, uuids
	}
	for uuid, e := range effect {
		switch e.Type {
		case TradeEffectType:
			totalEffect.Add(e.InnerEffect.(*TradeEffect))
			uuids = append(uuids, uuid)
		}
	}
	return totalEffect, uuids
}

func (t *TradeEffect) Add(effect *TradeEffect) {

	if effect.BuyFeeAmount != nil {
		if t.BuyFeeAmount == nil {
			t.BuyFeeAmount = utils.CreateInt(*effect.BuyFeeAmount)
		} else {
			*t.BuyFeeAmount += *effect.BuyFeeAmount
		}
	}

	if effect.BuyFeeMultiplier != nil {
		if t.BuyFeeMultiplier == nil {
			t.BuyFeeMultiplier = utils.CreateFloat(*effect.BuyFeeMultiplier)
		} else {
			*t.BuyFeeMultiplier += *effect.BuyFeeMultiplier
		}
	}

	if effect.SellFeeAmount != nil {
		if t.SellFeeAmount == nil {
			t.SellFeeAmount = utils.CreateInt(*effect.SellFeeAmount)
		} else {
			*t.SellFeeAmount += *effect.SellFeeAmount
		}
	}

	if effect.BonusProfitMultiplier != nil {
		if t.BonusProfitMultiplier == nil {
			t.BonusProfitMultiplier = utils.CreateFloat(*effect.BonusProfitMultiplier)
		} else {
			*t.BonusProfitMultiplier += *effect.BonusProfitMultiplier
		}
	}
	if effect.SellFeeMultiplier != nil {
		if t.SellFeeMultiplier == nil {
			t.SellFeeMultiplier = utils.CreateFloat(*effect.SellFeeMultiplier)
		} else {
			*t.SellFeeMultiplier += *effect.SellFeeMultiplier
		}
	}
	if effect.TaxMultiplier != nil {
		if t.TaxMultiplier == nil {
			t.TaxMultiplier = utils.CreateFloat(*effect.TaxMultiplier)
		} else {
			*t.TaxMultiplier += *effect.TaxMultiplier
		}
	}
	if effect.TaxPercent != nil {
		if t.TaxPercent == nil {
			t.TaxPercent = utils.CreateFloat(*effect.TaxPercent)
		} else {
			*t.TaxPercent += *effect.TaxPercent
		}
	}

	if effect.TradeBlocked != nil {
		if *effect.TradeBlocked == true {
			t.TradeBlocked = utils.CreateBool(true)
		}
	}
}
