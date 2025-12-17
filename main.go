package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"order-book-manager/client"
	"order-book-manager/cmd"
	"order-book-manager/grpc_orderbook"
	"order-book-manager/orderbook"
	"order-book-manager/server"
	"order-book-manager/users"

	_ "net/http/pprof"

	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type grpcServer struct {
	grpc_orderbook.UnimplementedOrderBookServer
}

func main() {
	cmd.Execute()

	fmt.Println("Welcome to Order Book Manager!")

	users.InitUserList()
	orderbook.InitOrderBook()
	go client.ConnectToWebSocket()
	fmt.Println("Websocket Client Initialized")

	srv := server.CreateServer()
	go server.RunWSServer(srv)
	fmt.Println("Websocket Server Initialized")

	grpcSvr := grpc.NewServer()
	go startGrpcServer(grpcSvr)
	fmt.Println("grpc Server Initialized")

	gracefulShutdown(srv, grpcSvr)
}

func startGrpcServer(grpcSvr *grpc.Server) {
	lis, err := net.Listen("tcp", ":8081")
	if err != nil {
		log.Fatal("Error on starting grpc server", err)
	}

	grpc_orderbook.RegisterOrderBookServer(grpcSvr, &grpcServer{})
	reflection.Register(grpcSvr)
	fmt.Println("grpc server is listening at ", lis.Addr())
	err = grpcSvr.Serve(lis)
	if err != nil {
		log.Fatal("Error on Listening on grpc server", err)
	}
}

func (s *grpcServer) SubscribeOrderBookForSymbol(ctx context.Context, request *grpc_orderbook.SubscribeOrderBookRequest) (response *grpc_orderbook.SubscribeOrderBookResponse, err error) {
	fmt.Println("Subscribe Order Book for ", request.CurrPair)
	client.SubscribeToCurrPair(request.CurrPair)
	return &grpc_orderbook.SubscribeOrderBookResponse{
		Message: string(fmt.Sprintf("Subscribed to %s", request.CurrPair)),
	}, nil
}

func (s *grpcServer) UnsubscribeOrderBookForSymbol(ctx context.Context, request *grpc_orderbook.UnsubscribeOrderBookRequest) (response *grpc_orderbook.UnsubscribeOrderBookResponse, err error) {
	fmt.Println("UnSubscribe Order Book for ", request.CurrPair)
	client.UnsubscribeToCurrPair(request.CurrPair)
	return &grpc_orderbook.UnsubscribeOrderBookResponse{
		Message: string(fmt.Sprintf("UnSubscribed to %s", request.CurrPair)),
	}, nil
}

func (s *grpcServer) GetOrderBookForSymbol(ctx context.Context, request *grpc_orderbook.GetOrderBookRequest) (response *grpc_orderbook.GetOrderBookResponse, err error) {
	fmt.Println("Get Order Book for ", request.CurrPair)
	return &grpc_orderbook.GetOrderBookResponse{
		Orderbook: string(orderbook.GetOrderBook(request.CurrPair)),
	}, nil
}

func (s *grpcServer) ListSubscriptions(ctx context.Context, request *grpc_orderbook.ListSubscriptionsRequest) (response *grpc_orderbook.ListSubscriptionsResponse, err error) {
	fmt.Println("Listing Subscriptions")
	subsList := client.ListSubscriptions()
	fmt.Println(subsList)

	subsPairList := []*grpc_orderbook.SubscriptionPair{}
	for _, curr := range subsList {
		subsPair := &grpc_orderbook.SubscriptionPair{
			Subscription: curr,
		}
		subsPairList = append(subsPairList, subsPair)
	}
	return &grpc_orderbook.ListSubscriptionsResponse{
		Subscriptions: subsPairList,
	}, nil
}

func gracefulShutdown(server *http.Server, grpcSvr *grpc.Server) {
	fmt.Println("Gradeful Shutdown is monitoring")
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("Shutdown Signal Received")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := server.Shutdown(ctx)
	if err != nil {
		log.Fatal("Forced Shutdown: ", err)
	}
	log.Println("Websocket Server Closed")

	client.CloseConnection()
	log.Println("Websocket Client Closed")

	grpcSvr.Stop()
	log.Println("grpc server Closed")

	log.Println("Server Exited Gracefully")
}
