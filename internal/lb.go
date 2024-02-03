package internal

import (
	"AIPainter-Dispatcher/internal/ws"
	"cmp"
	"encoding/json"
	"github.com/samber/lo"
	"io"
	"log"
	"net/http"
	"net/url"
	"slices"
	"sync"
	"time"
)

type QueueInf struct {
	Running int `json:"running"`
	Pending int `json:"pending"`
}

type Instance struct {
	target *url.URL
	inf    QueueInf
}

type LoadBalancer struct {
	instances []*url.URL
	queues    map[string]QueueInf
	queuesCh  chan string
	mutex     *sync.Mutex
	index     int
}

// NewLoadBalancer 创建目标代理
func NewLoadBalancer(targets ...string) *LoadBalancer {

	urls := lo.Map(targets, func(item string, index int) *url.URL {
		target, err := url.Parse(item)
		if err != nil {
			log.Panic(err)
		}
		return target
	})

	lb := &LoadBalancer{
		instances: urls,
		queues:    make(map[string]QueueInf),
		queuesCh:  make(chan string, 10),
		mutex:     &sync.Mutex{},
		index:     0,
	}

	//异步检查队列信息
	go lb.check()

	//异步刷新队列信息
	go lb.refresh()

	return lb
}

func (lb *LoadBalancer) refresh() {
	for {
		select {
		case host := <-lb.queuesCh:
			inf := lb.queues[host]
			inf.Pending++
			inf.Running++
		case <-time.After(time.Second * 5):

		}
	}
}

// Next 权重算法
func (lb *LoadBalancer) Next() (target *url.URL) {

	//轮询
	//target := lb.targets[lb.index]
	//lb.index = (lb.index + 1) % len(lb.targets)

	//异步更新
	defer func() {
		lb.queuesCh <- target.String()
	}()

	//根据节点队列数，权重分配
	if len(lb.instances) == 1 {
		target = lb.instances[0]
		return
	}

	sortTargets := make([]*url.URL, len(lb.instances))
	copy(sortTargets, lb.instances)

	//根据 pending，running 排序
	slices.SortFunc(sortTargets, func(a, b *url.URL) int {
		qa := lb.queues[a.String()]
		qb := lb.queues[b.String()]
		rs := cmp.Compare(qa.Pending, qb.Pending)
		if rs == 0 {
			return cmp.Compare(qa.Running, qb.Running)
		}
		return rs
	})

	target = sortTargets[0]
	return
}

type ComfyPacket struct {
	*ws.UpgradePacket
	PromptId string
	UserId   string
}

func (lb *LoadBalancer) check() {

	//刷新队列信息
	refreshQueueInf := func(target *url.URL) {

		//查询队列信息
		queueURL, _ := url.JoinPath(target.String(), "/queue")
		resp, err := http.Get(queueURL)
		if err != nil {
			return
		}

		defer resp.Body.Close()
		respByte, _ := io.ReadAll(resp.Body)
		var inf QueueInf
		_ = json.Unmarshal(respByte, &inf)

		lb.mutex.Lock()
		defer lb.mutex.Unlock()

		lb.queues[target.String()] = inf
	}

	for {
		<-time.After(time.Second * 5)

		//查询每个节点队列信息
		for i := 0; i < len(lb.instances); i++ {
			go refreshQueueInf(lb.instances[i])
		}
	}
}
