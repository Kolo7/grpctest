package client

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/grpclog"
	"grpstest/proto/pb"
	"grpstest/registercenter"
	"log"
	"testing"
	"time"
)

func TestHelloClient(t *testing.T) {
	schema, err := registercenter.GenerateAndRegisterEtcdResolver("8.210.188.38:2379", "HelloService")
	if err != nil {
		log.Fatal("init etcd resolver err:", err.Error())
		return
	}

	conn, err := grpc.Dial(fmt.Sprintf("%s:///HelloService", schema), grpc.WithInsecure(), grpc.WithBalancerName(roundrobin.Name))
	if err != nil {
		grpclog.Fatalln(err)
		return
	}
	defer conn.Close()
	c := pb.NewHelloClient(conn)
	resp, err := c.SayHello(context.Background(), &pb.HelloRequest{
		Name: "kuangle",
	})
	if err != nil {
		grpclog.Fatal(err)
		return
	}
	log.Printf("from grpc service:%s", resp.Message)
}

func TestTime(t *testing.T) {
	timer := time.NewTicker(time.Second * 3)
	defer timer.Stop()

	for {
		<-timer.C
		fmt.Println("timeout")
	}
}
