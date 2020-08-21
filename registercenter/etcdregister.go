package registercenter

import (
	"context"
	"fmt"
	"go.etcd.io/etcd/clientv3"
	"log"
	"strings"
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
	// minimum lease TTL is ttl-second
	resp, err := client.Grant(context.TODO(), int64(info.IntervalTime))
	if err != nil {
		log.Println("创建租约失败:", err)
		return err
	}
	// should get first, if not exist, set it
	kv, err := client.Get(context.Background(), info.ServiceName)
	log.Printf("%v", kv)
	serviceValue := fmt.Sprintf("%s:%d", info.Host, info.Port)
	if err != nil {
		log.Printf("etcd: service '%s' connect to etcd3 failed: %s", info.ServiceName, err.Error())
		return err
	}
	if kv.Kvs == nil {
		if _, err := client.Put(context.TODO(), info.ServiceName, serviceValue, clientv3.WithLease(resp.ID)); err != nil {
			log.Printf("etcd: set service '%s' with ttl to etcd3 failed: %s", info.ServiceName, err.Error())
		}
	} else {
		// refresh set to true for not notifying the watcher
		err := etcd.update(info.ServiceName, serviceValue, client, resp)
		if err != nil {
			return err
		}
	}
	log.Println("register successful")
	return nil
}

//etcd 实现下线接口
func (etcd *etcdRegisterImpl) UnRegister(info ServiceDescInfo) error {
	client, err := clientv3.NewFromURL(etcd.Address)
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
