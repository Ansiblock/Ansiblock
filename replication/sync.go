package replication

import (
	"bytes"
	"errors"
	"fmt"
	"math/rand"
	"net"
	synchro "sync"

	"github.com/Ansiblock/Ansiblock/block"
	"github.com/Ansiblock/Ansiblock/log"
	"go.uber.org/zap"

	"golang.org/x/crypto/ed25519"
)

// Addresses represents network addresses for different pipelines
type Addresses struct {
	Sync        net.UDPAddr
	Replication net.UDPAddr
	Message     net.UDPAddr
	Transaction net.UDPAddr
	Repair      net.UDPAddr
}

// NodeData represents structure for replication through sync
type NodeData struct {
	Self          ed25519.PublicKey
	Version       uint64
	Addresses     Addresses
	Producer      ed25519.PublicKey
	ValidVDFValue block.VDFValue
	NodeType      string
	NodeName      string
}

// Sync struct is responsible for replication
type Sync struct {
	table          map[string]*NodeData
	localVersions  map[string]uint64
	remoteVersions map[string]uint64
	index          uint64
	me             ed25519.PublicKey
	mutex          *synchro.RWMutex
}

// NewNodeData returns new NodeIno struct
func NewNodeData(self ed25519.PublicKey, nodeType string, name string, sync net.UDPAddr, replication net.UDPAddr,
	request net.UDPAddr, transaction net.UDPAddr, repair net.UDPAddr) *NodeData {
	return &NodeData{Self: self, Version: 0,
		Addresses: Addresses{Sync: sync, Replication: replication, Message: request, Transaction: transaction, Repair: repair},
		Producer:  block.VDF([]byte("hello")), ValidVDFValue: block.VDF([]byte("hello")),
		NodeType: nodeType, NodeName: name}
}

// Copy copies NodeData object
// TODO maybe should be changed to deep copy
func (n *NodeData) Copy() *NodeData {
	res := new(NodeData)
	*res = *n
	return res
}

// String method returns identity of NodeData.Self for debuging purposes
func (n *NodeData) String() string {
	return n.NodeName
}

// NewSync returns new Sync struct
func NewSync(me *NodeData) (*Sync, error) {
	if me.Version != 0 {
		return nil, errors.New("Bad NodeData")
	}
	//TODO check that addresses are valid
	sync := Sync{table: make(map[string]*NodeData),
		localVersions:  make(map[string]uint64),
		remoteVersions: make(map[string]uint64),
		index:          1,
		me:             me.Self}
	sync.localVersions[string(me.Self)] = sync.index
	sync.table[string(me.Self)] = me
	sync.mutex = &synchro.RWMutex{}
	return &sync, nil
}

//Equals comapres two Addresses
func (a *Addresses) Equals(other *Addresses) bool {
	return a.Sync.Network() == other.Sync.Network() && a.Sync.String() == other.Sync.String() &&
		a.Repair.Network() == other.Repair.Network() && a.Repair.String() == other.Repair.String() &&
		a.Replication.Network() == other.Replication.Network() && a.Replication.String() == other.Replication.String() &&
		a.Message.Network() == other.Message.Network() && a.Message.String() == other.Message.String() &&
		a.Transaction.Network() == other.Transaction.Network() && a.Transaction.String() == other.Transaction.String()
}

// Equals compares two NodeData
func (n *NodeData) Equals(other *NodeData) bool {
	return bytes.Equal(n.ValidVDFValue, other.ValidVDFValue) &&
		bytes.Equal(n.Producer, other.Producer) &&
		bytes.Equal(n.Self, other.Self) &&
		n.Version == other.Version &&
		n.Addresses.Equals(&other.Addresses)

}

func bytesToString(b []byte) string {
	return string(b[:8])
}

// String method returns identity of me for debuging purposes
func (c *Sync) String() string {
	return c.table[string(c.me)].NodeName
}

// MyNodeData returns my NodeData saved in the table
func (c *Sync) MyNodeData() *NodeData {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.table[string(c.me)]
}

// MyCopy returns a copy of my NodeData saved in the table
func (c *Sync) MyCopy() *NodeData {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.table[string(c.me)].Copy()
}

// TableCopy returns a copy of a Sync table
func (c *Sync) TableCopy() map[string]*NodeData {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	res := make(map[string]*NodeData)
	for k, v := range c.table {
		res[k] = v.Copy()
	}
	return res
}

// RemoteTableCopy returns a copy of the remote nodes
func (c *Sync) RemoteTableCopy() map[string]*NodeData {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	res := make(map[string]*NodeData)
	for k, v := range c.table {
		if _, ok := c.remoteVersions[k]; ok {
			res[k] = v.Copy()
		}
	}
	res[string(c.me)] = c.table[string(c.me)].Copy()
	return res
}

// ProducerNodeData returns producer's NodeData saved in the table
func (c *Sync) ProducerNodeData() *NodeData {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.table[string(c.MyNodeData().Producer)]
}

// ChangeProducer updates producer with new one
func (c *Sync) ChangeProducer(key ed25519.PublicKey) {
	my := c.MyNodeData().Copy()
	log.Info(fmt.Sprintf("Update Producer: {%v} -> {%v}", bytesToString(my.Producer), bytesToString(key)))
	my.Producer = key
	my.Version++
	c.Insert(my)
}

// Insert method will insert new NodeData into the table
func (c *Sync) Insert(info *NodeData) {
	c.mutex.Lock()
	c.insert(info)
	c.mutex.Unlock()
}

func (c *Sync) insert(info *NodeData) {
	if val, ok := c.table[string(info.Self)]; !ok || info.Version > val.Version {
		log.Debug(fmt.Sprintf("{%v}: Insert info.Self {%v} version {%v}",
			c.String(), info.String(), info.Version))
		// fmt.Printf("{%v}: Insert info.Self {%v} version {%v}\n\n",
		// 	c.String(), info.String(), info.Version)
		c.index++
		c.table[string(info.Self)] = info.Copy()
		c.localVersions[string(info.Self)] = c.index

	} else {
		log.Debug(fmt.Sprintf("{%v}: Insert Failed data: {%v} new.version {%v}  me.version {%v}",
			c.String(), info.String(), info.Version, c.table[string(info.Self)].Version))
		// fmt.Printf("{%v}: Insert Failed data: {%v} new.version {%v}  me.version {%v}\n",
		// 	c.String(), info.String(), info.Version, c.table[string(info.Self)].Version)
	}
}

func (c *Sync) updatesSince(index uint64) (ed25519.PublicKey, uint64, []*NodeData) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	res := make([]*NodeData, 0, len(c.table))
	for _, val := range c.table {
		if c.localVersions[string(val.Self)] > index {
			res = append(res, val)
		}
	}
	return c.me, c.index, res
}

func (c *Sync) RandomNode() (*NodeData, error) {
	var key string
	if len(c.table) < 2 {
		return nil, errors.New("Sync too small")
	}
	for {
		index := rand.Intn(len(c.table))
		for key = range c.table {
			if index == 0 {
				break
			}
			index--
		}
		if !bytes.Equal(c.table[key].Self, c.me) {
			break
		}
	}
	return c.table[key], nil
}

func (c *Sync) requestSync() (net.UDPAddr, *GetUpdates) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	node, err := c.RandomNode()
	if err != nil {
		return net.UDPAddr{}, nil
	}
	remoteIndex := uint64(0)
	if val, ok := c.remoteVersions[string(node.Self)]; ok {
		remoteIndex = val
	}
	request := GetUpdates{LastUpdateIndex: remoteIndex, MyInfo: c.table[string(c.me)].Copy()}
	// fmt.Printf("requestSync {%v}: remote count - %v\n", c.String(), len(c.remoteVersions))
	return node.Addresses.Sync, &request
}

func (c *Sync) update(rec *Updates) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	log.Info("got updates", zap.Int("", len(rec.Updates)))
	// fmt.Printf("{%v} got %v updates\n", c.String(), len(rec.Updates))
	for _, up := range rec.Updates {
		c.insert(up)
		// fmt.Printf("update inserted in {%v}\n", c.String())

	}
	c.remoteVersions[string(rec.From)] = rec.LastUpdateIndex
	// fmt.Printf("remote count %v in {%v}\n", len(c.remoteVersions), c.String())
}

// ConnectedNodes is responsible to calculate number of remote nodes
func (c *Sync) ConnectedNodes() uint64 {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	res := uint64(len(c.remoteVersions))
	for _, v := range c.remoteVersions {
		if res > v {
			res = v
		}
	}
	return res
}

// AllNodesExceptMe returns all nodes from sync except me node
func (c *Sync) AllNodesExceptMe(me ed25519.PublicKey) []*NodeData {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	res := make([]*NodeData, 0, len(c.table))
	for _, v := range c.table {
		if !bytes.Equal(v.Self, me) {
			res = append(res, v)
		}
	}
	return res
}

func (c *Sync) transitNodes() []*NodeData {
	me := c.MyCopy()
	table := c.TableCopy()
	nodes := make([]*NodeData, 0)
	for _, v := range table {
		if !bytes.Equal(me.Self, v.Self) && !bytes.Equal(me.Producer, v.Self) &&
			v.Addresses.Replication.String() != "0.0.0.0:0" {
			nodes = append(nodes, v)
		}
	}
	return nodes
}
