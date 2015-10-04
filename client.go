package main

import (
  "fmt"
  "log"
  "net/rpc/jsonrpc"
  "os"
  "strconv"
)

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

func main() {
  if len(os.Args) == 3 {
        buyStocks()
    }else if len(os.Args) == 2{
        checkPortfolio()
  }else {
        fmt.Println("Error: ", os.Args[0], "127.0.0.1:1238")
        log.Fatal(1)
  }
}

func buyStocks() {

    var reqObj StockRequest
    reqObj.Stocks = os.Args[1]
    budget64, _  := strconv.ParseFloat(os.Args[2], 32)
    reqObj.Budget = float32(budget64)
    client, err := jsonrpc.Dial("tcp", "127.0.0.1:1238")
    if err != nil {
        log.Fatal("dialing:", err)
    }

        respObj := new(StockResponse)

        err = client.Call("PortfolioCollection.RequestParser", reqObj, &respObj)

        if err != nil {
                log.Fatal("Error: ", err)
        }

                fmt.Print("Stocks:")
                fmt.Println(respObj.Stocks)
                fmt.Print("TradeID:")
                fmt.Println(respObj.TradeID)
                fmt.Print("UnvestedAmount:")
                fmt.Println(respObj.UnvestedAmount)
        }


        func checkPortfolio() {

            var TradeID int

            tradeID,_ := strconv.ParseInt(os.Args[1], 10, 32)
            TradeID = int(tradeID)

            client, err := jsonrpc.Dial("tcp", "localhost:1238")
            if err != nil {
                log.Fatal("dialing:", err)
            }

            var pfResponseObj PortfolioResponse

                err = client.Call("PortfolioCollection.CheckPortfolio", TradeID, &pfResponseObj)
                if err != nil {
                    log.Fatal("Input Error: ", err)
                }

                fmt.Print("Stocks:")
                fmt.Println(pfResponseObj.Stocks)
                fmt.Print("Current Market Value:")
                fmt.Println(pfResponseObj.Cmv)
                fmt.Print("Unvested Amount:")
                fmt.Println(pfResponseObj.UnvestedAmount)
        }
