# Order Book Manager

Order Book Manager is an market data management application that gets market data from Binance and push the order book and market data to the subscribed clients.

## Start the Application

Run with IDE

```bash
go run .
```
Run With Cobra Commands

1.Install Application

```bash
go install order-book-manager
```
2.(a). To Start with Default Currency Pair (BTCUSDT)
```bash
order-book-manager
```
2.(b). To Start with Other Currency Pair (ex: ETHUSDT)
```bash
order-book-manager currpair ETHUSDT
```


## Usage

### Admin Commands (gRPC)

```url
127.0.0.1:8081
```

#### Subscribe to Currency Pair (with Binance)
```grpc
OrderBook/SubscribeOrderBookForSymbol

{
     "currPair": "ETHUSDT"
}
```

#### Subscribe to Currency Pair (with Binance)
```grpc
OrderBook/SubscribeOrderBookForSymbol

{
     "currPair": "ETHUSDT"
}
```

#### Unsubscribe to Currency Pair (with Binance)
```grpc
OrderBook/UnsubscribeOrderBookForSymbol

{
     "currPair": "ETHUSDT"
}
```

#### Get Order Book
```grpc
OrderBook/GetOrderBookForSymbol

{
     "currPair": "ETHUSDT"
}
```

#### Get Subscribed Currency Pairs List (with Binance)
```grpc
OrderBook/ListSubscriptions
```


### Client Connection

Web Socket URL
```
ws://127.0.0.1:8080/ws
```

#### Client Subscription
```
SUB ETHUSDT
```

#### Client Unsubscription
```
UNSUB ETHUSDT
```
