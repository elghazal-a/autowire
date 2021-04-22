package wireguard

import (
  "fmt"
  "net"
)

type Peer struct {
  PublicKey      string
  IP             net.IP
  EndpointIP     net.IP
  EndpointPort   int
  AllowedIPs     string
}


func (p Peer) IsEndpointIPDefined() bool {
	return p.EndpointIP != nil
}


func (p Peer) HasSameEndpointIP(wgEndpointIp string) bool {
	return p.EndpointIP.Equal(net.ParseIP(wgEndpointIp))
}

func (p Peer) GetEndpointIP() string {
	return p.EndpointIP.String()
}

func (p Peer) GetIP() string {
  return p.IP.String()
}

func (p Peer) IsClient() bool {
  return p.EndpointIP == nil
}

func (p Peer) String() string {
  return fmt.Sprintf("PublicKey=%s|IP=%s|EndpointIP=%s|EndpointPort=%d|AllowedIPs=%s", 
    p.PublicKey, p.IP.String(), p.EndpointIP.String(), p.EndpointPort, p.AllowedIPs)
}