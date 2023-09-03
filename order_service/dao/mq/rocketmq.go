package mq

import (
	"fmt"
	"github.com/idMiFeng/order_service/config"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
)

var (
	Producer rocketmq.Producer
)

func Init() (err error) {
	Producer, err = rocketmq.NewProducer(
		producer.WithNsResolver(primitive.NewPassthroughResolver([]string{config.Conf.RocketMqConfig.Addr})),
		//producer.WithNsResolver(primitive.NewPassthroughResolver(endPoint)),
		producer.WithRetry(2),
		producer.WithGroupName(config.Conf.RocketMqConfig.GroupId),
	)
	if err != nil {
		fmt.Println(err)
		return err
	}
	err = Producer.Start()
	if err != nil {
		fmt.Println(err)
		return
	}

	// for i := 0; i < 10; i++ {
	// 	msg := &primitive.Message{
	// 		Topic: topic,
	// 		Body:  []byte("Hello RocketMQ Go Client! " + strconv.Itoa(i)),
	// 	}
	// 	res, err := p.SendSync(context.Background(), msg)

	// 	if err != nil {
	// 		fmt.Printf("send message error: %s\n", err)
	// 	} else {
	// 		fmt.Printf("send message success: result=%s\n", res.String())
	// 	}
	// }

	// err = p.Shutdown()
	// if err != nil {
	// 	fmt.Printf("shutdown producer error: %s", err.Error())
	// }
	return nil
}

func Exit() error {
	err := Producer.Shutdown()
	if err != nil {
		fmt.Printf("shutdown producer error: %s", err.Error())
	}
	return err
}
