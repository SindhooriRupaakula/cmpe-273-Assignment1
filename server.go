package main

 import (
        "errors"
        "encoding/json"
        "fmt"
        "net"
        "net/rpc"
        "net/http"
        "os"
        "net/rpc/jsonrpc"
        "strings"
        "strconv"
        "io/ioutil"
        "math"
        "math/rand"
        "log"
 )

var tradeID int

type JSONObj struct {
    List struct {
        Resources []struct {

        Resource struct {
            Fields struct {
                Name    string `json:"name"`
                Price   string `json:"price"`
                Symbol  string `json:"symbol"`
                Ts      string `json:"ts"`
                Type    string `json:"type"`
                UTCTime string `json:"utctime"`
                Volume  string `json:"volume"`
            } `json:"fields"`
        } `json:"resource"`
    } `json:"resources"`
} `json:"list"`
}

type StockRequest struct {
    Stocks string
    Budget float32
 }

type StockResponse struct {
    TradeID int
    Stocks []string
    UnvestedAmount float32
}

type PortfolioResponse struct {
  Stocks []string
  Cmv float32
  UnvestedAmount float32
}

type Share struct {
  Price float32
  Count int
}

 type Portfolio struct {
   Stocks map[string](*Share)
   UnvestedAmount float32
 }

 type PortfolioCollection struct {
   Portfolios map[int](*Portfolio)
 }

 func (p *PortfolioCollection) RequestParser(sreq *StockRequest, sresp *StockResponse) error {

     tradeID++
     sresp.TradeID = tradeID
     if p.Portfolios == nil {
             p.Portfolios = make(map[int](*Portfolio))
             p.Portfolios[tradeID] = new(Portfolio)
             p.Portfolios[tradeID].Stocks = make(map[string]*Share)
     }


     stocks := strings.Split(sreq.Stocks, ",")
     budget := float32(sreq.Budget)
     var totalSpent float32
     for _, stock := range stocks {

             splitString := strings.Split(stock, ":")
             stockSymbol := splitString[0]
             stockPercent := splitString[1]
             stockPercent = strings.TrimSuffix(stockPercent, "%")
             fPercent64, _ := strconv.ParseFloat(stockPercent, 32)
             fPercent := float32(fPercent64 / 100.00)

             fmt.Println("The Stock Symbol is",stockSymbol)

             fmt.Println("It's Percentage is", fPercent)

             financeAPIPrice := YahooAPI(stockSymbol)

             sharesCount := int(math.Floor(float64(budget * fPercent / financeAPIPrice)))
             sharesCountFloat := float32(sharesCount)
             totalSpent += sharesCountFloat * financeAPIPrice

             endResult := stockSymbol + ":" + strconv.Itoa(sharesCount) + ":$" + strconv.FormatFloat(float64(financeAPIPrice), 'f', 2, 32)

             sresp.Stocks = append(sresp.Stocks, endResult)

             if _, ok := p.Portfolios[tradeID]; !ok {

                     pfObj := new(Portfolio)
                     pfObj.Stocks = make(map[string]*Share)
                     p.Portfolios[tradeID] = pfObj
             }
             if _, ok := p.Portfolios[tradeID].Stocks[stockSymbol]; !ok {

                     shareObj := new(Share)
                     shareObj.Price = financeAPIPrice
                     shareObj.Count = sharesCount
                     p.Portfolios[tradeID].Stocks[stockSymbol] = shareObj
             } else {

                     total := float32(sharesCountFloat*financeAPIPrice) + float32(p.Portfolios[tradeID].Stocks[stockSymbol].Count)*p.Portfolios[tradeID].Stocks[stockSymbol].Price
                     p.Portfolios[tradeID].Stocks[stockSymbol].Price = total / float32(sharesCount+p.Portfolios[tradeID].Stocks[stockSymbol].Count)
                     p.Portfolios[tradeID].Stocks[stockSymbol].Count += sharesCount
             }

     }

     unvestedAmount := budget - totalSpent
     sresp.UnvestedAmount = unvestedAmount
     p.Portfolios[tradeID].UnvestedAmount += unvestedAmount
     return nil
 }

 func (p* PortfolioCollection) CheckPortfolio(tradeID int ,  presp *PortfolioResponse) error {


                                      if objValues, ok := p.Portfolios[tradeID]; ok {

                                     var currentMarketValue float32
                                     for stockSymbol, p := range objValues.Stocks {

                                             financeAPIPrice := YahooAPI(stockSymbol)

                                             var result string
                                             if p.Price < financeAPIPrice {
                                                     result = "+$" + strconv.FormatFloat(float64(financeAPIPrice), 'f', 2, 32)
                                             } else if p.Price > financeAPIPrice {
                                                     result = "-$" + strconv.FormatFloat(float64(financeAPIPrice), 'f', 2, 32)
                                             } else {
                                                     result = "$" + strconv.FormatFloat(float64(financeAPIPrice), 'f', 2, 32)
                                             }
                                             stock := stockSymbol + ":" + strconv.Itoa(p.Count) + ":" + result

                                             presp.Stocks = append(presp.Stocks, stock)

                                             currentMarketValue += float32(p.Count) * financeAPIPrice
                                     }
                                     fmt.Print("Unvested amount: ", objValues.UnvestedAmount)

                                     presp.UnvestedAmount = objValues.UnvestedAmount
                                     presp.Cmv = currentMarketValue
                             }else {
                                     return errors.New("Trade ID doesnt exist")
                             }

                             return nil
     }

func YahooAPI(stockSymbol string) float32 {
  url := fmt.Sprintf("http://finance.yahoo.com/webservice/v1/symbols/%s/quote?format=json",stockSymbol)
  urlRes,err := http.Get(url)

  if err != nil {
    log.Fatal(err)
  }

  body, err := ioutil.ReadAll(urlRes.Body)
  urlRes.Body.Close()

  if err != nil {
    log.Fatal(err)
  }

  var jsonObj JSONObj
  err = json.Unmarshal(body, &jsonObj)

  if err != nil {
    panic(err)
  }

  fmt.Println(jsonObj.List.Resources[0].Resource.Fields.Name)
  fmt.Println(jsonObj.List.Resources[0].Resource.Fields.Symbol)
  fmt.Println(jsonObj.List.Resources[0].Resource.Fields.Price)

  floatFinalPrice, err := strconv.ParseFloat(jsonObj.List.Resources[0].Resource.Fields.Price, 32)
  return float32(floatFinalPrice)
}

func main(){

  tradeID = rand.Intn(100) + 1000
  portfolios := new(PortfolioCollection)
  rpc.Register(portfolios)

  tcpAddr, err := net.ResolveTCPAddr("tcp", ":1238")
  checkError(err)

  listener, err := net.ListenTCP("tcp", tcpAddr)
  checkError(err)

    for {
        conn, err := listener.Accept()
        if err != nil {
            continue
        }
        jsonrpc.ServeConn(conn)
    }

}

func checkError(err error) {
    if err != nil {
        fmt.Println("Fatal error ", err.Error())
        os.Exit(1)
    }

}
