package replication

import (
	"time"

	"github.com/Ansiblock/Ansiblock/network"
)

// SyncGenerator thread sends sync request to the random node in every 100 millisecond
func SyncGenerator(sync *Sync, sleep time.Duration) <-chan *network.Blob {
	out := make(chan *network.Blob, 10)
	go func(out chan<- *network.Blob) {
		for {
			syncAddr, req := sync.requestSync()
			if req != nil {

				blob := req.Serialize()
				blob.Addr = &syncAddr
				// fmt.Printf("SyncGenerator: %v\n", syncAddr)

				out <- blob
			}
			time.Sleep(sleep)
		}
	}(out)
	return out
}

func handleBlob(sync *Sync, blob *network.Blob) *network.Blob {
	if blob.Data[0] == getUpdatesType {
		req := new(GetUpdates)
		req.Deserialize(blob)
		addr := req.MyInfo.Addresses.Sync
		from, index, updates := sync.updatesSince(req.LastUpdateIndex)
		// fmt.Printf("getUpdates: %v\n", string(from[:8]))

		if len(updates) < 1 {
			// log.Debug(fmt.Sprintf("No updates for {%v} since %v", sync.me, sync.index))
			// fmt.Printf("No updates for {%v} since %v", sync.me, sync.index)
			return nil
		}
		response := new(Updates)
		response.From = from
		response.LastUpdateIndex = index
		response.Updates = updates
		resBlob := response.Serialize()
		resBlob.Addr = &addr
		sync.Insert(req.MyInfo)
		return resBlob

	} else if blob.Data[0] == updatesType {
		rec := new(Updates)
		rec.Deserialize(blob)
		// fmt.Printf("updates: %v\n", string(rec.From[:8]))
		sync.update(rec)
	}
	return nil
}

// SyncListener is a thread responsible for receiving updates from sync channel
// TODO optimize: there should be pointers in Blobs
func SyncListener(sync *Sync, input <-chan *network.Blobs) <-chan *network.Blobs {
	out := make(chan *network.Blobs, 10)
	go func(out chan<- *network.Blobs) {
		for {
			// fmt.Println("SyncListener trying to read")
			blobs := <-input
			// fmt.Println("SyncListener read!!!")
			responses := make([]network.Blob, 0, len(blobs.Bs))
			for _, blob := range blobs.Bs {
				resp := handleBlob(sync, &blob)
				if resp != nil {
					responses = append(responses, *resp)
				}
			}
			if len(responses) > 0 {
				out <- &network.Blobs{Bs: responses}
			}
		}
	}(out)
	return out
}
