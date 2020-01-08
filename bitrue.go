package bitrue

import (
	"encoding/json"
	"github.com/ericlagergren/decimal"
	"github.com/kr/pretty"
	"github.com/spf13/cast"
	"github.com/tidwall/gjson"
	"log"
	"math"
	"strings"
)

type Exchange struct {
	symbols           map[string]string
	SymbolInfos       []*SymbolData
	MinQuoteAmountMap map[string]float64
	MinBaseAmountMap  map[string]float64
}

var accessKey string
var secretKey string

func NewExchange(ak string, sk string) *Exchange {
	accessKey = ak
	secretKey = sk
	ex := &Exchange{
		MinQuoteAmountMap: make(map[string]float64),
		MinBaseAmountMap:  make(map[string]float64),
	}
	ex.getSymbols()
	ex.initMinQuoteAmount()
	return ex
}

//need set yourself
func (ex *Exchange) initMinQuoteAmount() {
	ex.MinQuoteAmountMap["BTRUSDT"] = 10
	ex.MinQuoteAmountMap["BTRXRP"] = 10
	ex.MinQuoteAmountMap["BTRBTC"] = 0.0001
	ex.MinQuoteAmountMap["BTRETH"] = 0.01
}

func (ex *Exchange) Update(symbol string) bool {
	symbol = strings.ToUpper(symbol)
	if a, ok := ex.MinQuoteAmountMap[symbol]; ok {
		if price, ok := ex.GetTickerPrice(symbol).Float64(); ok {
			amount, _ := ex.TruncAmount(symbol, a/price)
			ex.MinBaseAmountMap[symbol] = amount
			return true
		}
	}
	return false
}

func (ex *Exchange) GetMinAmount(symbol string) (float64, bool) {
	symbol = strings.ToUpper(symbol)
	if amount, ok := ex.MinBaseAmountMap[symbol]; ok {
		return amount, true
	}
	return 0, false
}
func (ex *Exchange) GetUpdateMinAmount(symbol string) (float64, bool) {
	symbol = strings.ToUpper(symbol)
	ex.Update(symbol)
	if amount, ok := ex.MinBaseAmountMap[symbol]; ok {
		return amount, true
	}
	return 0, false
}

func (ex *Exchange) getSymbols() {
	body := HttpGetRequest(https+"/api/v1/exchangeInfo", nil)
	//log.Println(body)
	data := gjson.Get(body, "symbols")
	symbolInfos := make([]*SymbolData, 0)
	err := json.Unmarshal([]byte(data.String()), &symbolInfos)
	if err != nil {
		log.Panicln(err)
	}
	ex.SymbolInfos = symbolInfos
}

func (ex *Exchange) getSymbolInfo(symbol string) *SymbolData {
	symbol = strings.ToUpper(symbol)
	for _, symbolInfo := range ex.SymbolInfos {
		if symbolInfo.Symbol == symbol {
			return symbolInfo
		}
	}
	return nil
}

func (ex *Exchange) GetDepth(symbol string) *Depth {
	params := make(map[string]string)
	params["symbol"] = symbol
	body := HttpGetRequest(https+"/api/v1/depth", params)
	depth := &Depth{}
	err := json.Unmarshal([]byte(body), depth)
	if err != nil {
		log.Println(err)
		return nil
	}
	return depth
}

func (ex *Exchange) GetTickerPrice(symbol string) *decimal.Big {
	params := make(map[string]string)
	params["symbol"] = symbol
	body := HttpGetRequest(https+"/api/v1/ticker/price", params)
	priceTicker := &PriceTicker{}
	err := json.Unmarshal([]byte(body), priceTicker)
	if err != nil {
		log.Println(err)
	}
	return priceTicker.Price
}

func (ex *Exchange) GetBookTicker(symbol string) *BookTicker {
	params := make(map[string]string)
	params["symbol"] = symbol
	body := HttpGetRequest(https+"/api/v1/ticker/bookTicker", params)

	bookTicker := &BookTicker{}
	err := json.Unmarshal([]byte(body), bookTicker)
	if err != nil {
		log.Println(err)
	}
	return bookTicker
}

func (ex *Exchange) GetBuyPrice(symbol string) float64 {
	bookTicker := ex.GetBookTicker(symbol)
	buyPrice, _ := bookTicker.BidPrice.Float64()
	return buyPrice
}

func (ex *Exchange) GetSellPrice(symbol string) float64 {
	bookTicker := ex.GetBookTicker(symbol)
	sellPrice, _ := bookTicker.AskPrice.Float64()
	return sellPrice
}

func (ex *Exchange) get24hr(symbol string) {
	params := make(map[string]string)
	params["symbol"] = symbol

	body := HttpGetRequest(https+"/api/v1/ticker/24hr", params)
	println(body)
}

func println(str string) {
	m := make([]interface{}, 0)
	err := json.Unmarshal([]byte(str), &m)
	if err != nil {
		log.Println(str)
	}
	log.Println(pretty.Formatter(m))
}

//return orderId
func (ex *Exchange) BuyLimit(symbol string, price float64, amount float64) int64 {
	params := make(map[string]string)
	params["type"] = "LIMIT"
	params["symbol"] = symbol
	params["side"] = "BUY"
	params["price"] = cast.ToString(price)
	params["quantity"] = cast.ToString(amount)
	data := SignedRequest(POST, https+"/api/v1/order", params)

	orderId := gjson.Get(data, "orderId").Int()
	if orderId == 0 {
		log.Println(data)
	}
	return orderId
}

func (ex *Exchange) SellLimit(symbol string, price float64, amount float64) int64 {
	params := make(map[string]string)
	params["type"] = "LIMIT"
	params["symbol"] = symbol
	params["side"] = "SELL"
	params["price"] = cast.ToString(price)
	params["quantity"] = cast.ToString(amount)
	data := SignedRequest(POST, https+"/api/v1/order", params)
	orderId := gjson.Get(data, "orderId").Int()
	if orderId == 0 {
		log.Println(data)
	}
	return orderId
}

func (ex *Exchange) QueryOrder(symbol string, orderId int64) *OrderData {
	params := make(map[string]string)
	params["symbol"] = symbol
	params["orderId"] = cast.ToString(orderId)

	body := SignedRequest(GET, https+"/api/v1/order", params)
	order := &OrderData{}
	err := json.Unmarshal([]byte(body), order)
	if err != nil {
		log.Println(err, order, body)
		return nil
	}
	return order
}

func (ex *Exchange) QueryOpenOrders(symbol string) []*OrderData {
	params := make(map[string]string)
	params["symbol"] = symbol
	body := SignedRequest(GET, https+"/api/v1/openOrders", params)
	var orders []*OrderData
	err := json.Unmarshal([]byte(body), &orders)
	if err != nil {
		log.Println(err, body)
	}
	return orders
}

func (ex *Exchange) GetBalance(currency string) *BalanceData {
	params := make(map[string]string)
	body := SignedRequest(GET, https+"/api/v1/account", params)
	balance := &Balance{}
	err := json.Unmarshal([]byte(body), balance)
	if err != nil {
		log.Println(err, body)
	}

	for _, balanceData := range balance.Balances {
		if balanceData.Currency == currency {
			return balanceData
		}
	}
	return nil
}

func (ex *Exchange) Cancel(symbol string, orderId int64) bool {
	params := make(map[string]string)
	params["symbol"] = symbol
	params["orderId"] = cast.ToString(orderId)
	body := SignedRequest(DELETE, https+"/api/v1/order", params)
	if body[:7] == `{"code"` {
		log.Println(body)
		return false
	}
	return true
}

func (ex *Exchange) TruncPrice(symbol string, price float64) (float64, bool) {
	symbolInfo := ex.getSymbolInfo(symbol)
	if symbolInfo != nil {
		pre := math.Pow10(symbolInfo.QuotePrecision)
		tPrice := math.Round(price*pre) / pre
		return tPrice, true
	}
	return 0, false
}

func (ex *Exchange) TruncAmount(symbol string, amount float64) (float64, bool) {
	symbolInfo := ex.getSymbolInfo(symbol)
	if symbolInfo != nil {
		pre := math.Pow10(symbolInfo.BasePrecision)
		tAmount := math.Round(amount*pre) / pre
		return tAmount, true
	}
	return 0, false
}

func (ex *Exchange) GetTiny(symbol string) float64 {
	symbolInfo := ex.getSymbolInfo(symbol)
	if symbolInfo != nil {
		return 1 / math.Pow10(symbolInfo.QuotePrecision)
	}
	return 0
}

func (ex *Exchange) GetOrderMap(symbol string) map[int64]*OrderData {
	orders := ex.QueryOpenOrders(symbol)
	if orders == nil {
		return nil
	}
	orderMap := make(map[int64]*OrderData)
	for _, order := range orders {
		orderMap[order.OrderId] = order
	}
	return orderMap
}
