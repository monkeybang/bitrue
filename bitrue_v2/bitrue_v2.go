package bitrue

import (
	"encoding/json"
	"github.com/ericlagergren/decimal"
	"github.com/kr/pretty"
	"github.com/monkeybang/bitrue"
	"github.com/spf13/cast"
	"github.com/tidwall/gjson"
	"log"
	"math"
	"strconv"
	"strings"
	"time"
)

type Exchange struct {
	AppKey            string `json:"app_key"`
	SecretKey         string `json:"secret_key"`
	Host              string `json:"host"`
	symbols           map[string]string
	SymbolInfos       []*bitrue.SymbolData
	MinQuoteAmountMap map[string]float64
}

func NewExchange(ak, sk, host string) *Exchange {
	ex := &Exchange{
		AppKey:            ak,
		SecretKey:         sk,
		Host:              host,
		MinQuoteAmountMap: make(map[string]float64),
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

// Current exchange trading rules and symbol information
func (ex *Exchange) getSymbols() {
	body := bitrue.HttpGetRequest(ex.Host+"/api/v1/exchangeInfo", nil)
	//log.Println(body)
	data := gjson.Get(body, "symbols")
	symbolInfos := make([]*bitrue.SymbolData, 0)
	err := json.Unmarshal([]byte(data.String()), &symbolInfos)
	if err != nil {
		log.Panicln("getSymbols error", err)
	}
	ex.SymbolInfos = symbolInfos
}

func (ex *Exchange) GetQuoteAmount(symbol string) (float64, bool) {
	symbol = strings.ToUpper(symbol)
	if a, ok := ex.MinQuoteAmountMap[symbol]; ok {
		return a, true
	}
	return 0, false
}

func (ex *Exchange) GetSymbolInfo(symbol string) *bitrue.SymbolData {
	symbol = strings.ToUpper(symbol)
	for _, symbolInfo := range ex.SymbolInfos {
		if symbolInfo.Symbol == symbol {
			return symbolInfo
		}
	}
	return nil
}

// 获取深度数据
func (ex *Exchange) GetDepth(symbol string) *bitrue.Depth {
	params := make(map[string]string)
	params["symbol"] = symbol
	body := bitrue.HttpGetRequest(ex.Host+"/api/v1/depth", params)
	depth := &bitrue.Depth{}
	err := json.Unmarshal([]byte(body), depth)
	if err != nil {
		log.Println(err, body)
		return nil
	}
	return depth
}

// 获取交易对最新价
func (ex *Exchange) GetTickerPrice(symbol string) *decimal.Big {
	params := make(map[string]string)
	params["symbol"] = symbol
	body := bitrue.HttpGetRequest(ex.Host+"/api/v1/ticker/price", params)
	priceTicker := &bitrue.PriceTicker{}
	err := json.Unmarshal([]byte(body), priceTicker)
	if err != nil {
		log.Println(err, body)
	}
	return priceTicker.Price
}

// Best price/qty on the order book for a symbol or symbols.
func (ex *Exchange) GetBookTicker(symbol string) *bitrue.BookTicker {
	params := make(map[string]string)
	params["symbol"] = symbol
	body := bitrue.HttpGetRequest(ex.Host+"/api/v1/ticker/bookTicker", params)

	bookTicker := &bitrue.BookTicker{}
	err := json.Unmarshal([]byte(body), bookTicker)
	if err != nil {
		log.Println(err, body)
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

// 24小时内的价格变化
func (ex *Exchange) get24hr(symbol string) {
	params := make(map[string]string)
	params["symbol"] = symbol

	body := bitrue.HttpGetRequest(ex.Host+"/api/v1/ticker/24hr", params)
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

// return orderId
func (ex *Exchange) BuyLimit(symbol string, price float64, amount float64) int64 {
	params := make(map[string]string)
	params["type"] = "LIMIT"
	params["symbol"] = symbol
	params["side"] = "BUY"
	params["price"] = cast.ToString(price)
	params["quantity"] = cast.ToString(amount)

	data := bitrue.SignedRequestWithKey(bitrue.POST, ex.Host+"/api/v1/order", params, ex.AppKey, ex.SecretKey)

	orderId := gjson.Get(data, "orderId").Int()
	if orderId == 0 {
		log.Println(data, symbol, price, amount)
	}
	return orderId
}

//
func (ex *Exchange) SellLimit(symbol string, price float64, amount float64) int64 {
	params := make(map[string]string)
	params["type"] = "LIMIT"
	params["symbol"] = symbol
	params["side"] = "SELL"
	params["price"] = cast.ToString(price)
	params["quantity"] = cast.ToString(amount)
	data := bitrue.SignedRequestWithKey(bitrue.POST, ex.Host+"/api/v1/order", params, ex.AppKey, ex.SecretKey)
	orderId := gjson.Get(data, "orderId").Int()
	if orderId == 0 {
		log.Println(data, symbol, price, amount)
	}
	return orderId
}

func (ex *Exchange) QueryOrder(symbol string, orderId int64) *bitrue.OrderData {
	params := make(map[string]string)
	params["symbol"] = symbol
	params["orderId"] = cast.ToString(orderId)

	body := bitrue.SignedRequestWithKey(bitrue.GET, ex.Host+"/api/v1/order", params, ex.AppKey, ex.SecretKey)
	order := &bitrue.OrderData{}
	err := json.Unmarshal([]byte(body), order)
	if err != nil {
		log.Println(err, body, params)
		return nil
	}
	return order
}

func (ex *Exchange) QueryOpenOrders(symbol string) []*bitrue.OrderData {
	params := make(map[string]string)
	params["symbol"] = symbol
	body := bitrue.SignedRequestWithKey(bitrue.GET, ex.Host+"/api/v1/openOrders", params, ex.AppKey, ex.SecretKey)
	var orders []*bitrue.OrderData
	err := json.Unmarshal([]byte(body), &orders)
	if err != nil {
		log.Println(err, body)
	}
	return orders
}

func (ex *Exchange) QueryAllOrders(symbol string, orderId int64) []*bitrue.OrderData {
	params := make(map[string]string)
	params["symbol"] = symbol
	if orderId > 0 {
		params["orderId"] = strconv.Itoa(int(orderId))
	}
	body := bitrue.SignedRequestWithKey(bitrue.GET, ex.Host+"/api/v1/allOrders", params, ex.AppKey, ex.SecretKey)
	var orders []*bitrue.OrderData
	err := json.Unmarshal([]byte(body), &orders)
	if err != nil {
		log.Println(err, body)
	}
	return orders
}

func (ex *Exchange) GetBalance(currency string) *bitrue.BalanceData {
	params := make(map[string]string)
	body := bitrue.SignedRequestWithKey(bitrue.GET, ex.Host+"/api/v1/account", params, ex.AppKey, ex.SecretKey)
	balance := &bitrue.Balance{}
	err := json.Unmarshal([]byte(body), balance)
	if err != nil {
		log.Println(err, body)
		return nil
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
	body := bitrue.SignedRequestWithKey(bitrue.DELETE, ex.Host+"/api/v1/order", params, ex.AppKey, ex.SecretKey)
	if body[:7] == `{"code"` {
		log.Println(body)
		return false
	}
	return true
}

func (ex *Exchange) TruncPrice(symbol string, price float64) (float64, bool) {
	symbolInfo := ex.GetSymbolInfo(symbol)
	if symbolInfo != nil {
		pre := math.Pow10(symbolInfo.QuotePrecision)
		tPrice := math.Round(price*pre) / pre
		return tPrice, true
	}
	return 0, false
}

func (ex *Exchange) TruncAmount(symbol string, amount float64) (float64, bool) {
	symbolInfo := ex.GetSymbolInfo(symbol)
	if symbolInfo != nil {
		pre := math.Pow10(symbolInfo.BasePrecision)
		tAmount := math.Round(amount*pre) / pre
		return tAmount, true
	}
	return 0, false
}

func (ex *Exchange) GetTiny(symbol string) float64 {
	symbolInfo := ex.GetSymbolInfo(symbol)
	if symbolInfo != nil {
		return 1 / math.Pow10(symbolInfo.QuotePrecision)
	}
	return 0
}

func (ex *Exchange) GetOrderMap(symbol string) map[int64]*bitrue.OrderData {
	orders := ex.QueryOpenOrders(symbol)
	if orders == nil {
		return nil
	}
	orderMap := make(map[int64]*bitrue.OrderData)
	for _, order := range orders {
		orderMap[order.OrderId] = order
	}
	return orderMap
}

// 获取本地当前时间
func GetCurrentLocalTime() string {
	return cast.ToString(time.Now().UnixNano() / 1000000)
}

// TODO 获取Bitrue服务器时间
func GetCurrentServerTime() string {
	return ""
}
