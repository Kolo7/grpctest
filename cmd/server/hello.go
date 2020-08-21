package main

import (
	"fmt"
	"golang.org/x/net/context"
	"golang.org/x/net/trace"
	pb "grpstest/proto/pb"
	"net/http"
)

const (
	// Address gRPC服务地址
	Address = "127.0.0.1:50052"
)

type helloService struct{}

var HelloService = helloService{}

func (h helloService) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloResponse, error) {
	resp := new(pb.HelloResponse)
	resp.Message = fmt.Sprintf("v1 Hello %s.", in.Name)
	return resp, nil
}

/*func main() {
	listen, err := net.Listen("tcp", Address)
	if err != nil {
		grpclog.Fatalf("Failed to listen: %v", err)
	}


	// 实例化grpc Server
	s := grpc.NewServer()

	// 注册HelloService
	pb.RegisterHelloServer(s, HelloService)

	go startTrace()

	fmt.Println("Listen on " + Address)
	s.Serve(listen)
}*/

func startTrace() {
	trace.AuthRequest = func(req *http.Request) (any, sensitive bool) {
		return true, true
	}
	go http.ListenAndServe(":50051", nil)
	fmt.Println("Trace listen on 50051")
}
