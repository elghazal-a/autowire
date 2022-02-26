package main

import (
  "flag"
  "fmt"
  backendModule "github.com/geniousphp/autowire/backends"
  "github.com/geniousphp/autowire/wireguard"
  "github.com/geniousphp/autowire/util"
  "log"
  "net"
  "sort"
  "strings"
)

func main() {
  flag.Parse()
  if err := initConfig(); err != nil {
    log.Fatal("FATAL: ", err.Error())
  }

  // Backend init
  b, err := backendFactory()
  if err != nil {
    log.Fatal("FATAL: ", err)
  }

  newWgInterface, err := initialize(b)
  if err != nil {
    log.Fatal("FATAL: ", err)
  }

  b.MonitorKv(config.CS_FullKVPrefix, *newWgInterface)
  if config.GC_Enable {
    b.MonitorNodes(config.CS_FullKVPrefix, *newWgInterface)
  }

}


func initialize(backend backendModule.Backend) (*wireguard.Interface, error) {
  if err := backend.Lock(config.CS_FullKVPrefix, config.WG_EndpointIP); err != nil {
    return nil, err
  }
  defer backend.Unlock()

  var shouldConfigureWG bool = false 

  peers, err := backend.GetPeers(config.CS_FullKVPrefix)
  if err != nil {
    return nil, err
  }


  if config.WG_IP != "" { //WG_IP is set, this is like a pet server
    if indexPeer := util.IsIPUsed(peers, config.WG_IP); indexPeer >= 0 { //WG_IP is already used
      if !peers[indexPeer].IsEndpointIPDefined() { //WG_IP is used by another device
        return nil, fmt.Errorf("The WG_IP=%s is already used by another client", config.WG_IP)
      } else{
        if !peers[indexPeer].HasSameEndpointIP(config.WG_EndpointIP) { //WG_IP is used by an other server
          return nil, fmt.Errorf("The WG_IP=%s is already used by another server %s", peers[indexPeer].GetIP(), peers[indexPeer].GetEndpointIP())
        } else { //WG_IP is used by the current server
          shouldConfigureWG = true
        }
      }
    } else { //The WG_IP is not used
      if indexPeer := util.IsEndpointIPExist(peers, config.WG_EndpointIP); indexPeer >= 0 { //WG_EndpointIP already registred
        return nil, fmt.Errorf("The current WG_EndpointIP=%s is already registred with another WG_IP=%s", config.WG_EndpointIP, peers[indexPeer].GetIP())
      } else{ //WG_EndpointIP is not registred
        shouldConfigureWG = true
      }
    }
  } else {//WG_IP is not set, this is like a cattle server
    if indexPeer := util.IsEndpointIPExist(peers, config.WG_EndpointIP); indexPeer >= 0 && peers[indexPeer].IP != nil { //WG_EndpointIP already registred
      config.WG_IP = peers[indexPeer].IP.String()
      shouldConfigureWG = true
    } else{ //pick unused IP
      log.Print("INFO: Picking unused IP from the range=", config.WG_Range)
      _, wgRangeIPNet, err := net.ParseCIDR(config.WG_Range)
      if err != nil {
        return nil, err
      }
      wgIPStart := wgRangeIPNet.IP
      util.IncIP(wgIPStart) //Skip IP Network

      //The loop goes over all ips in the network
      for myFutureWG_IP := wgIPStart; wgRangeIPNet.Contains(myFutureWG_IP); util.IncIP(myFutureWG_IP) {
        if indexPeer := util.IsIPUsed(peers, myFutureWG_IP.String()); indexPeer < 0 {
          config.WG_IP = myFutureWG_IP.String()
          break;
        } 
      }
      if config.WG_IP == "" {
        return nil, fmt.Errorf("All IPs are used")
      } else{
        shouldConfigureWG = true
      }

    }
  }


  wireguardConfig := wireguard.Configuration{}

  if(shouldConfigureWG){
    // configure wireguard 
    privKey, pubKey, err := wireguard.InitWgKeys(config.WG_InterfaceConfigFolder)
    if err != nil {
      return nil, err
    }

    var mask string
    if(config.WG_Relay){
      mask = strings.Split(config.WG_Range, "/")[1]
    } else{
      mask = "32"
    }



    wireguardConfig = wireguard.Configuration{
      Interface: wireguard.Interface{
        Name          : config.WG_InterfaceName,
        Address       : fmt.Sprintf("%s/%s", config.WG_IP, mask), 
        ListenPort    : config.WG_Port, 
        PublicKey     : pubKey,
        PrivateKey    : privKey,
        PostUp        : config.WG_PostUp,
        PostDown      : config.WG_PostDown,
      },
      Peers: peers,
    }

    if _, err := wireguard.ConfigureWireguard(wireguardConfig); err != nil {
      return nil, err
    }

    allowedips := []string{}
    if config.WG_AllowedIPs != "" {
      allowedips = strings.Split(config.WG_AllowedIPs, ",")
    }
    allowedips = append(allowedips, fmt.Sprintf("%s/%s", config.WG_IP, "32"))
    if(config.WG_Relay){
      _, rangeIPNet, err := net.ParseCIDR(config.WG_Range)
      if err != nil {
        return nil, err
      }
      allowedips = append(allowedips, rangeIPNet.String())
    }
    sort.Strings(allowedips)

    newPeer := wireguard.Peer{
      PublicKey     : pubKey,
      IP            : net.ParseIP(config.WG_IP),
      EndpointIP    : net.ParseIP(config.WG_EndpointIP), 
      EndpointPort  : config.WG_Port, 
      AllowedIPs    : strings.Join(allowedips, ","),
    }


    //Add my interface as Peer in the Backend
    if err := backend.AddPeer(config.CS_FullKVPrefix, wireguardConfig.Interface, newPeer); err != nil {
      return nil, err
    }



  }


  return &wireguardConfig.Interface, nil
}



func backendFactory() (backendModule.Backend, error) {
  if config.AW_Backend == "consul" {
    b, err := backendModule.NewConsulBackend(config.CS_Address)
    if err != nil {
      return nil, err
    }
    return b, nil
  } else{
    return nil, fmt.Errorf("Unsupported Backend. Available backends: [consul]")
  }
}



