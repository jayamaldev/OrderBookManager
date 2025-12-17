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

	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type grpcServer struct {
	grpc_orderbook.UnimplementedOrderBookServer
}

func main() {
	cmd.Execute()

	fmt.Println("Welcome to Order Book Manager!")

	g, ctx := errgroup.WithContext(context.Background())

	users.InitUserList()
	orderbook.InitOrderBook()
	srv := server.CreateServer()
	grpcSvr := grpc.NewServer()

	g.Go(func() error {
		if err := client.ConnectToWebSocket(); err != nil {
			return fmt.Errorf("error on WS Client %s", err)
		}
		fmt.Println("Websocket Client Initialized")
		return nil
	})

	g.Go(func() error {
		go func() {
			<-ctx.Done()
			fmt.Println("Shutting Down Websocket Server")
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			err := srv.Shutdown(shutdownCtx)
			if err != nil {
				fmt.Println("error on shutting down WS server.", err)
			}
		}()

		if err := server.RunWSServer(srv); err != nil {
			return fmt.Errorf("error on WS server %s", err)
		}
		fmt.Println("Websocket Server Initialized")
		return nil
	})

	g.Go(func() error {
		go func() {
			<-ctx.Done()
			fmt.Println("Shutting Down gRPC server")
			grpcSvr.GracefulStop()
		}()

		if err := startGrpcServer(grpcSvr); err != nil {
			return fmt.Errorf("error on gRPC server %s", err)
		}
		fmt.Println("grpc Server Initialized")
		return nil
	})

	go gracefulShutdown(ctx, srv, grpcSvr)

	if err := g.Wait(); err != nil {
		fmt.Println("Received error from group:", err)
	}
}

func startGrpcServer(grpcSvr *grpc.Server) error {
	lis, err := net.Listen("tcp", ":8081")
	if err != nil {
		log.Fatal("Error on starting grpc server", err)
		return err
	}

	grpc_orderbook.RegisterOrderBookServer(grpcSvr, &grpcServer{})
	reflection.Register(grpcSvr)
	fmt.Println("grpc server is listening at ", lis.Addr())
	err = grpcSvr.Serve(lis)
	if err != nil {
		log.Fatal("Error on Listening on grpc server", err)
		return err
	}

	return nil
}

func (s *grpcServer) SubscribeOrderBookForSymbol(ctx context.Context, request *grpc_orderbook.SubscribeOrderBookRequest) (response *grpc_orderbook.SubscribeOrderBookResponse, err error) {
	fmt.Println("Subscribe Order Book for ", request.CurrPair)

	subErr := client.SubscribeToCurrPair(ctx, request.CurrPair)
	if subErr != nil {
		return &grpc_orderbook.SubscribeOrderBookResponse{
			Message: string(fmt.Sprintf("Subscription Failed to %s", request.CurrPair)),
		}, subErr
	} else {
		return &grpc_orderbook.SubscribeOrderBookResponse{
			Message: string(fmt.Sprintf("Subscribed to %s", request.CurrPair)),
		}, nil
	}
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

func gracefulShutdown(ctx context.Context, server *http.Server, grpcSvr *grpc.Server) {
	fmt.Println("Gradeful Shutdown is monitoring")
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	<-quit

	log.Println("Shutdown Signal Received")

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
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
