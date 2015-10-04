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
    stocks string
    budget float32
 }

type StockResponse struct {
    tradeID int
    stocks []string
    unvestedAmount float32
}

type PortfolioResponse struct {
  stocks []string
  cmv float32
  unvestedAmount float32
}

type Share struct {
  price float32
  count int
}

 type Portfolio struct {
   stocks map[string](*Share)
   unvestedAmount float32
 }

 type PortfolioCollection struct {
    portfolios map[int](*Portfolio)
 }

 func (p *PortfolioCollection) RequestParser(sreq *StockRequest, sresp *StockResponse) error {

     tradeID++
     sresp.tradeID = tradeID
     if p.portfolios == nil {
             p.portfolios = make(map[int](*Portfolio))
             p.portfolios[tradeID] = new(Portfolio)
             p.portfolios[tradeID].stocks = make(map[string]*Share)
     }


     stocks := strings.Split(sreq.stocks, ",")
     budget := float32(sreq.budget)
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

             sresp.stocks = append(sresp.stocks, endResult)

             if _, ok := p.portfolios[tradeID]; !ok {

                     pfObj := new(Portfolio)
                     pfObj.stocks = make(map[string]*Share)
                     p.portfolios[tradeID] = pfObj
             }
             if _, ok := p.portfolios[tradeID].stocks[stockSymbol]; !ok {

                     shareObj := new(Share)
                     shareObj.price = financeAPIPrice
                     shareObj.count = sharesCount
                     p.portfolios[tradeID].stocks[stockSymbol] = shareObj
             } else {

                     total := float32(sharesCountFloat*financeAPIPrice) + float32(p.portfolios[tradeID].stocks[stockSymbol].count)*p.portfolios[tradeID].stocks[stockSymbol].price
                     p.portfolios[tradeID].stocks[stockSymbol].price = total / float32(sharesCount+p.portfolios[tradeID].stocks[stockSymbol].count)
                     p.portfolios[tradeID].stocks[stockSymbol].count += sharesCount
             }

     }

     unvestedAmount := budget - totalSpent
     sresp.unvestedAmount = unvestedAmount
     p.portfolios[tradeID].unvestedAmount += unvestedAmount
     return nil
 }

 func (p* PortfolioCollection) CheckPortfolio(tradeID int ,  presp *PortfolioResponse) error {


                                      if objValues, ok := p.portfolios[tradeID]; ok {

                                     var currentMarketValue float32
                                     for stockSymbol, p := range objValues.stocks {

                                             financeAPIPrice := YahooAPI(stockSymbol)

                                             var result string
                                             if p.price < financeAPIPrice {
                                                     result = "+$" + strconv.FormatFloat(float64(financeAPIPrice), 'f', 2, 32)
                                             } else if p.price > financeAPIPrice {
                                                     result = "-$" + strconv.FormatFloat(float64(financeAPIPrice), 'f', 2, 32)
                                             } else {
                                                     result = "$" + strconv.FormatFloat(float64(financeAPIPrice), 'f', 2, 32)
                                             }
                                             stock := stockSymbol + ":" + strconv.Itoa(p.count) + ":" + result

                                             presp.stocks = append(presp.stocks, stock)

                                             currentMarketValue += float32(p.count) * financeAPIPrice
                                     }
                                     fmt.Print("Unvested amount: ", objValues.unvestedAmount)

                                     presp.unvestedAmount = objValues.unvestedAmount
                                     presp.cmv = currentMarketValue
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
