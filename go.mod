module grpstest

go 1.14

require (
	//github.com/coreos/etcd v3.3.22+incompatible
	github.com/coreos/go-semver v0.3.0 // indirect
	github.com/coreos/pkg v0.0.0-20180928190104-399ea9e2e55f // indirect
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/golang/protobuf v1.4.1
	github.com/google/uuid v1.1.1 // indirect
	go.etcd.io/etcd v0.0.0-20191023171146-3cf2f69b5738
	go.uber.org/zap v1.15.0 // indirect
	golang.org/x/net v0.0.0-20190813141303-74dc4d7220e7
	google.golang.org/grpc v1.29.1
	google.golang.org/protobuf v1.25.0
)

replace go.etcd.io/etcd v0.0.0-20191023171146-3cf2f69b5738 =>  ../../golangWork/pkg/mod/go.etcd.io/etcd
