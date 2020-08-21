package server

import (
	"context"
	"fmt"
	"grpstest/proto/pb"
)

type helloService struct{}

var HelloService = helloService{}

func (h helloService) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloResponse, error) {
	resp := new(pb.HelloResponse)
	resp.Message = fmt.Sprintf("v1 Hello %s.", in.Name)
	return resp, nil
}
