package server

import (
	"google.golang.org/grpc"
	"grpstest/proto/pb"
	"grpstest/registercenter"
	"grpstest/server"
	"log"
	"net"
	"testing"
	"time"
)

func TestStartClient(t *testing.T) {
	lis, err := net.Listen("tcp", "127.0.0.1:8090")
	if err != nil {
		t.Errorf("failed to listen: %v", err)
		return
	}
	etcdRegister := registercenter.NewEtcdRegisterImpl("8.210.188.38:2379")

	go func() {
		err := etcdRegister.Register(registercenter.ServiceDescInfo{
			ServiceName:  "HelloService",
			Host:         "127.0.0.1",
			Port:         8090,
			IntervalTime: time.Duration(10),
		})
		if err != nil {
			t.Errorf("failed to register: %v", err)
		}
		time.Sleep(time.Second * 1)
	}()

	grpcServer := grpc.NewServer()
	pb.RegisterHelloServer(grpcServer, server.HelloService)
	err = grpcServer.Serve(lis)
	if err != nil {
		t.Errorf("failed to serve:%v", err)
		return
	}
}

func TestStartClient2(t *testing.T) {
	lis, err := net.Listen("tcp", "127.0.0.1:8091")
	if err != nil {
		t.Errorf("failed to listen: %v", err)
		return
	}
	etcdRegister := registercenter.NewEtcdRegisterImpl("8.210.188.38:2379")

	go func() {
		err := etcdRegister.Register(registercenter.ServiceDescInfo{
			ServiceName:  "HelloService",
			Host:         "127.0.0.1",
			Port:         8091,
			IntervalTime: time.Duration(10),
		})
		if err != nil {
			t.Errorf("failed to register: %v", err)
		}
		time.Sleep(time.Second * 1)
	}()

	grpcServer := grpc.NewServer()
	pb.RegisterHelloServer(grpcServer, server.HelloService)
	err = grpcServer.Serve(lis)
	if err != nil {
		t.Errorf("failed to serve:%v", err)
		return
	}
}

func TestStartClient3(t *testing.T) {
	lis, err := net.Listen("tcp", "127.0.0.1:8092")
	if err != nil {
		t.Errorf("failed to listen: %v", err)
		return
	}
	etcdRegister := registercenter.NewEtcdRegisterImpl("8.210.188.38:2379")

	go func() {
		err := etcdRegister.Register(registercenter.ServiceDescInfo{
			ServiceName:  "HelloService",
			Host:         "127.0.0.1",
			Port:         8092,
			IntervalTime: time.Duration(10),
		})
		if err != nil {
			t.Errorf("failed to register: %v", err)
		}
		time.Sleep(time.Second * 1)
	}()

	grpcServer := grpc.NewServer()
	pb.RegisterHelloServer(grpcServer, server.HelloService)
	err = grpcServer.Serve(lis)
	if err != nil {
		t.Errorf("failed to serve:%v", err)
		return
	}
}
func TestUnRegisterServer(t *testing.T) {
	etcdRegister := registercenter.NewEtcdRegisterImpl("8.210.188.38:2379")
	err := etcdRegister.UnRegister(registercenter.ServiceDescInfo{
		ServiceName:  "HelloService",
		Host:         "127.0.0.1",
		Port:         8090,
		IntervalTime: time.Duration(10000),
	})
	if err != nil {
		log.Printf("fail to UnRegister%v", err)
	}
}
