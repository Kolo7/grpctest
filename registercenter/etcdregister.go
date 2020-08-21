package registercenter

import (
	"context"
	"fmt"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/etcdserver/api/v3rpc/rpctypes"
	"log"
	"strings"
	"time"
)

type etcdRegisterImpl struct {
	Address string
}

//创建etcd的注册结构体
func NewEtcdRegisterImpl(target string) *etcdRegisterImpl {
	return &etcdRegisterImpl{Address: target}
}

func (etcd *etcdRegisterImpl) Register(info ServiceDescInfo) error {
	client, err := clientv3.NewFromURL(etcd.Address)
	if err != nil {
		log.Println("连接etcd失败:", err)
		return err
	}
	defer client.Close()

	kv := clientv3.NewKV(client)
	lease := clientv3.NewLease(client)
	var curLeaseId clientv3.LeaseID = 0
	for {
		if curLeaseId == 0 {
			leaseResp, err := lease.Grant(context.Background(), int64(info.IntervalTime))
			if err != nil {
				panic(err)
			}
			key := fmt.Sprintf("%s/%d", info.ServiceName, leaseResp.ID)
			value := fmt.Sprintf("%s:%d", info.Host, info.Port)
			if _, err := kv.Put(context.Background(), key, value, clientv3.WithLease(leaseResp.ID)); err != nil {
				panic(err)
			}
			curLeaseId = leaseResp.ID
		} else {
			if _, err := lease.KeepAliveOnce(context.Background(), curLeaseId); err == rpctypes.ErrLeaseNotFound {
				curLeaseId = 0
				continue
			}
		}
		time.Sleep(time.Duration(5) * time.Second)
	}
}

//etcd 实现下线接口
func (etcd *etcdRegisterImpl) UnRegister(info ServiceDescInfo) error {
	client, err := clientv3.NewFromURL(etcd.Address)
	defer client.Close()
	if err != nil {
		return err
	}
	serviceValue := fmt.Sprintf("%s:%d", info.Host, info.Port)
	resp, err := client.Grant(context.TODO(), int64(info.IntervalTime))
	if err != nil {
		log.Println("创建租约失败:", err)
		return err
	}
	err = etcd.removeFormList(info.ServiceName, serviceValue, client, resp)
	if err != nil {
		return err
	}
	return err
}

func (etcd *etcdRegisterImpl) update(k, v string, client *clientv3.Client, LeaseResp *clientv3.LeaseGrantResponse) (err error) {
	// TODO 此处需要注意并发事务问题,可以采用watch 乐观锁 + channel 抢占方式
	resp, err := client.Get(context.Background(), k)
	if err != nil {
		log.Printf("etcd get service fail %s", err.Error())
		return
	}
	if len(resp.Kvs) != 1 {
		log.Printf("etcd update service fail")
		return
	}
	value := string(resp.Kvs[0].Value)
	newValue := fmt.Sprintf("%s,%s", value, v)
	_, err = client.Put(context.Background(), k, newValue, clientv3.WithLease(LeaseResp.ID))
	if err != nil {
		log.Printf("etcd: refresh service '%s' with ttl to etcd3 failed: %s", k, err.Error())
		return
	}
	return
}

func (etcd *etcdRegisterImpl) removeFormList(k, v string, client *clientv3.Client, LeaseResp *clientv3.LeaseGrantResponse) (err error) {
	resp, err := client.Get(context.Background(), k)
	if err != nil {
		log.Printf("etcd get service fail %s", err.Error())
		return
	}
	if len(resp.Kvs) != 1 {
		log.Printf("etcd remove fail")
		return
	}
	values := string(resp.Kvs[0].Value)
	vArray := strings.Split(values, ",")
	newValues := ""
	for _, e := range vArray {
		if newValues != "" && e != v {
			newValues = fmt.Sprintf("%s,", newValues)
		}
		if e != v {
			newValues = fmt.Sprintf("%s%s", newValues, e)
		}
	}
	if newValues == "" {
		_, err = client.Delete(context.Background(), k)
		if err != nil {
			log.Printf("etcd delete empty kvs fail[%s]", err.Error())
			return
		}
		return
	}
	_, err = client.Put(context.Background(), k, newValues, clientv3.WithLease(LeaseResp.ID))
	if err != nil {
		log.Printf("etcd update service[%s] to new service[%s] fail.", values, newValues)
		return
	}
	return
}
