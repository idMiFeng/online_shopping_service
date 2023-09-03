package main

import (
	"github.com/idMiFeng/order_service/proto"

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

func main() {
	defer conn.Close()

}
