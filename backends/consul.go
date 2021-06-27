package backend

import (
	"encoding/json"
	"fmt"
  "github.com/geniousphp/autowire/wireguard"
	"github.com/hashicorp/consul/api"
	"log"
	"time"
	"github.com/geniousphp/autowire/util"
)


type ConsulBackend struct {
	client 		*api.Client
	lock 			*api.Lock
	lockChan  <-chan struct{}
}


func NewConsulBackend(endpoint string) (*ConsulBackend, error) {

	config := api.DefaultConfig()
	config.Address = endpoint


	log.Printf("INFO: Connecting to consul %s", config.Address)

	ConsulClient, err := api.NewClient(config)
	if err != nil {
		return nil, err
	}

	// check health to ensure communication with consul are working
	if _, _, err := ConsulClient.Health().State(api.HealthAny, nil); err != nil {
		log.Printf("ERROR: health check failed for %v", config.Address)
		return nil, err
	}

	return &ConsulBackend{client: ConsulClient}, nil
}



func (cb *ConsulBackend) Lock(location string, value string) (error) {
  opts := &api.LockOptions{
    Key 					: fmt.Sprintf("%s/%s", location, "pick-ip-lock"),
    Value					: []byte(value),
    SessionOpts 	: &api.SessionEntry{
      Behavior		: "release",
      TTL					: "10s",
    },
    MonitorRetries: 20,
  }
  var err error
  cb.lock, err = cb.client.LockOpts(opts)
  if err != nil {
    return err
  }

  stopCh := make(chan struct{})
  cb.lockChan, err = cb.lock.Lock(stopCh)
  if err != nil {
    return err
  }
  return nil
}
func (cb *ConsulBackend) Unlock()  {
	cb.lock.Unlock()
}

func (cb *ConsulBackend) GetPeers(location string) ([]wireguard.Peer, error) {

  ConsulKV := cb.client.KV()

  consulPeers, _, err := ConsulKV.List(fmt.Sprintf("%s/%s", location, "peers"), nil)
	if err != nil {
		return nil, err
	}

	peers := []wireguard.Peer{}

	if consulPeers == nil {
		return peers, nil
	}


	for _, consulPeer := range consulPeers {
		peer := wireguard.Peer{}

		err = json.Unmarshal(consulPeer.Value, &peer)
		if err != nil {
			return nil, err
		}


		peers = append(peers, peer)

	}


	return peers, nil
}


func (cb *ConsulBackend) AddPeer(location string, wgInterface wireguard.Interface, peer wireguard.Peer) (error) {
  ConsulKV := cb.client.KV()


	peerJson, err := json.Marshal(peer)
	if err != nil {
		return err
	}
	_, err = ConsulKV.Put(
		&api.KVPair{
			Key:   fmt.Sprintf("%s/%s/%s", location, "peers", peer.IP.String()),
			Value: peerJson,
		},
		nil,
	)

  if err != nil {
    return err
  }

	return nil
}


func (cb *ConsulBackend) MonitorKv(location string, wgInterface wireguard.Interface) {
  stopMonitorKvPrefixChan := make(chan bool)
  newPeersChan            := make(chan []wireguard.Peer)

  go cb.monitorKvPrefix(location, newPeersChan, stopMonitorKvPrefixChan)

  for {
    select {
      case <-stopMonitorKvPrefixChan:
        log.Print("INFO: monitorKvPrefix goroutine stopped")
      case newPeers := <-newPeersChan:
        log.Print("INFO: Received new peers from monitorKvPrefix goroutine")
        wireguardConfig := wireguard.Configuration{
          Interface: wgInterface,
          Peers: newPeers,
        }
        if _, err := wireguard.ConfigureWireguard(wireguardConfig); err != nil {
          log.Fatal(err)
        }
    }
  }
}




func (cb *ConsulBackend) monitorKvPrefix(location string, newPeersChan chan []wireguard.Peer, stopMonitorKvPrefixChan chan bool) {
  ConsulKV := cb.client.KV()

  var shouldSyncPeers bool = false
  var waitIndex uint64
  waitIndex = 0

  for {
    opts := api.QueryOptions{
      AllowStale: false, 
      RequireConsistent: true, 
      UseCache: false,
      WaitIndex: waitIndex,
    }
    log.Print("INFO: Watching Consul KV Prefix in blocking query index=", waitIndex)
    consulPeers, meta, err := ConsulKV.List(fmt.Sprintf("%s/%s", location, "peers"), &opts)
    if err != nil {
      // Prevent backend errors from consuming all resources.
      fmt.Errorf("Error watching Peers in KV %s", err)
      time.Sleep(time.Second * 2)
      continue
    }

    if meta.LastIndex < waitIndex {
    	//index went backward, should reset the index to 0
    	log.Print("INFO: the waitIndex went backward, it will be reset to 0")
  		waitIndex = 0
  		shouldSyncPeers = true
    } else {
    	if waitIndex != meta.LastIndex {
  			shouldSyncPeers = true
    	} else {
  			shouldSyncPeers = false
    	}
    	waitIndex = meta.LastIndex
    }

    if shouldSyncPeers {
			newPeers := []wireguard.Peer{}
			if consulPeers == nil {
	    	newPeersChan <- newPeers
			} else {
				for _, consulPeer := range consulPeers {
					peer := wireguard.Peer{}
					err = json.Unmarshal(consulPeer.Value, &peer)
					if err != nil {
			      fmt.Errorf("Error Parsing Peers in KV %s", err)
			      time.Sleep(time.Second * 2)
			      continue
					}
					newPeers = append(newPeers, peer)
				}
		    newPeersChan <- newPeers
			}
    }

  }
  stopMonitorKvPrefixChan <- true
}


func (cb *ConsulBackend) MonitorNodes(location string, wgInterface wireguard.Interface) {
  stopMonitorNodesChan    := make(chan bool)
  newNodesChan            := make(chan []string)

  go cb.monitorNodes(location, wgInterface, newNodesChan, stopMonitorNodesChan)

  for {
    select {
      case <-stopMonitorNodesChan:
        log.Print("INFO: monitorNodes goroutine stopped")
      case newNodes := <-newNodesChan:
        log.Print("INFO: received new nodes from monitorNodes goroutine")
        cb.removeLeftPeers(location, newNodes)
    }
  }
}


func (cb *ConsulBackend) monitorNodes(location string, wgInterface wireguard.Interface, newNodesChan chan []string, stopMonitorNodesChan chan bool) {
  opts := &api.LockOptions{
    Key 					: fmt.Sprintf("%s/%s", location, "monitor-nodes-lock"),
    Value					: []byte(wgInterface.Address),
    SessionOpts 	: &api.SessionEntry{
      Behavior		: "release",
      TTL					: "10s",
    },
    MonitorRetries: 20,
  }


  lock, err := cb.client.LockOpts(opts)
  if err != nil {
    log.Fatal(err)
  }
  stopCh := make(chan struct{})
  _, err = lock.Lock(stopCh)
  if err != nil {
    log.Fatal(err)
  }

  var shouldSyncPeers bool = false


  var ConsulCatalog *api.Catalog
  ConsulCatalog = cb.client.Catalog()
  var waitIndex uint64
  waitIndex = 0
  for {
    opts := api.QueryOptions{
      AllowStale: false, 
      RequireConsistent: true, 
      UseCache: false,
      WaitIndex: waitIndex,
    }

    log.Print("INFO: Watching Consul Nodes in blocking query index=", waitIndex)
    listNodes, meta, err := ConsulCatalog.Nodes(&opts)
    if err != nil {
      // Prevent backend errors from consuming all resources.
      fmt.Errorf("Error watching Nodes in Consul %s", err)
      time.Sleep(time.Second * 2)
      continue
    }

    if meta.LastIndex < waitIndex {
    	//index went backward, should reset the index to 0
    	log.Print("INFO: the waitIndex went backward, it will be reset to 0")
  		waitIndex = 0
  		shouldSyncPeers = true
    } else {
    	if waitIndex != meta.LastIndex {
  			shouldSyncPeers = true
    	} else {
  			shouldSyncPeers = false
    	}
    	waitIndex = meta.LastIndex
    }

    if shouldSyncPeers {
	    newNodes := []string{}
	    for _, node := range listNodes {
				newNodes = append(newNodes, node.Address)

	    }

	    newNodesChan <- newNodes
    }


  }


  stopMonitorNodesChan <- true
}

func (cb *ConsulBackend) removeLeftPeers(location string, newNodes []string) {
  ConsulKV := cb.client.KV()

  peers, err := cb.GetPeers(location)
  if err != nil {
    fmt.Errorf("Error Retriving Peers from Consul KV %s", err)
  }
  for _, peer := range peers {
		if !peer.IsClient() && !util.SliceContains(newNodes, peer.GetEndpointIP()) {
      log.Print("INFO: Removing Peer from Consul KV: ", peer.GetEndpointIP())
      _, err := ConsulKV.DeleteTree(fmt.Sprintf("%s/%s/%s", location, "peers", peer.GetIP()), nil)
      if err != nil {
        fmt.Errorf("Error Removing Peer from Consul KV %s", err)
      }
		}
  }
}