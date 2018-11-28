package pipelines_test

import (
	"fmt"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/Ansiblock/Ansiblock/pipelines"
	"github.com/Ansiblock/Ansiblock/replication"
)

func converge(producer *replication.NodeData, numNodes uint64) (bool, []*replication.NodeData) {
	spy := replication.NewNode("spy", "spy")
	someAddr := net.UDPAddr{
		Port: 0,
		IP:   net.ParseIP("0.0.0.0"),
	}
	me := spy.Data.Self
	spy.Data.Addresses.Replication = someAddr
	spy.Data.Addresses.Message = someAddr
	spySync, _ := replication.NewSync(spy.Data)
	spySync.Insert(producer)
	spySync.ChangeProducer(producer.Self)
	go pipelines.Synchronization(spySync, spy.Sockets.Sync, spy.Sockets.SyncSend)

	converged := false
	for i := 0; i < 100; i++ {
		num := spySync.ConnectedNodes()
		fmt.Println(num)

		if num == numNodes {
			converged = true
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	return converged, spySync.AllNodesExceptMe(me)
}

func TestSynchronization(t *testing.T) {
	producer := replication.NewNode("producer", "test1")
	producer.Data.Producer = producer.Data.Self
	syncL, _ := replication.NewSync(producer.Data)
	fmt.Printf("Producer : %v\n", producer.Data.String())
	go pipelines.Synchronization(syncL, producer.Sockets.Sync, producer.Sockets.SyncSend)
	for i := 0; i < 5; i++ {
		val := replication.NewNode("signer", "test"+strconv.Itoa(i))
		val.Data.Producer = producer.Data.Self
		fmt.Printf("Val : %v\n", val.Data.String())
		syncV, _ := replication.NewSync(val.Data)
		syncV.Insert(producer.Data)
		go pipelines.Synchronization(syncV, val.Sockets.Sync, val.Sockets.SyncSend)
	}

	con, data := converge(producer.Data, 6)
	if !con {
		t.Errorf("Sync did not converge \n\n\n%v\n\n\n", data)
	}

}
