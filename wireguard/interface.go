package wireguard

import (
  "strings"
)

type Interface struct {
  Name string
  Address string //can be /32 IP or /24 subnet in case of Relay server
  ListenPort int
  PublicKey string
  PrivateKey string
  PostUp string
  PostDown string
}

func (i Interface) IsRelay() bool {
  return !strings.HasSuffix(i.Address, "/32")
}
