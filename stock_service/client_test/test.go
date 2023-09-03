package main

import (
	"context"
	"fmt"
	"github.com/idMiFeng/stock_service/proto"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	conn   *grpc.ClientConn
	client proto.StockClient
)

func init() {
	var err error
	conn, err = grpc.Dial("127.0.0.1:8382", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	client = proto.NewStockClient(conn)
}

func TestGetStock() {
	param := &proto.GoodsStockInfo{
		GoodsId: 1,
		Num:     1,
	}
	resp, err := client.GetStock(context.Background(), param)
	fmt.Printf("resp:%v err:%v\n", resp, err)
}

func TestReduceStock(wg *sync.WaitGroup) {
	defer wg.Done()
	param := &proto.GoodsStockInfo{
		GoodsId: 1,
		Num:     1,
	}
	resp, err := client.ReduceStock(context.Background(), param)
	fmt.Printf("resp:%v err:%v\n", resp, err)
}

func main() {
	defer conn.Close()
	TestGetStock()
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go TestReduceStock(&wg)
	}
	wg.Wait()
}
