package backend

import (
  "github.com/geniousphp/autowire/wireguard"
)

type Backend interface {
	//location can be a key for K/V stores or table for databases,....
	Lock(location string, value string) (error) 
	Unlock()
	GetPeers(location string) ([]wireguard.Peer, error)
	AddPeer(location string, wgInterface wireguard.Interface, peer wireguard.Peer) (error)
	MonitorKv(location string, wgInterface wireguard.Interface) ()
	MonitorNodes(location string, wgInterface wireguard.Interface) ()
}
