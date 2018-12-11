package main

import (
  "fmt"
  "net"
  "log"
  "os"
  "io/ioutil"
  "time"
  "strconv"
  "strings"
  "flag"
  "github.com/hashicorp/consul/api"
  "github.com/geniousphp/autowire/wireguard"
  "github.com/geniousphp/autowire/ifconfig"
  "github.com/geniousphp/autowire/util"
)

func main() {
  flag.Parse()
  if err := initConfig(); err != nil {
    log.Fatal(err.Error())
  }
  log.Print("Starting Autowire")

  physicalIpAddr, err := getPhysicalIpAddr()
  if err != nil {
    log.Fatal(err)
  }
  if physicalIpAddr == "" {
    log.Fatal("Error while detecting network interface or Ip Address")
  }

  privKey, pubKey, err := initWgKeys()
  if err != nil {
    log.Fatal(err)
  }

  conf := api.DefaultConfig()
  ConsulClient, err := api.NewClient(conf)
  if err != nil {
    log.Fatal(err)
  }
  err = initialize(ConsulClient, physicalIpAddr, privKey, pubKey)
  if err != nil {
    log.Fatal(err)
  }

  monitorPeers(ConsulClient, physicalIpAddr)
}


func initWgKeys() (string, string, error) {
  wgInterfaceConfigFolder := config.WGConfigFolder + "/" + config.WGInterfaceName
  if _, err := os.Stat(wgInterfaceConfigFolder + "/private"); os.IsNotExist(err) {
    err := os.MkdirAll(wgInterfaceConfigFolder, 0700)

    if err != nil {
      return "", "", err
    }

    privKey, err := wireguard.Genkey()
    if err != nil {
      return "", "", err
    }

    err = ioutil.WriteFile(wgInterfaceConfigFolder + "/private", privKey, 0600)
    if err != nil {
      return "", "", err
    }

  }

  privKey, err := ioutil.ReadFile(wgInterfaceConfigFolder + "/private")
  if err != nil {
    return "", "", err
  }

  pubKey, err := wireguard.ExtractPubKey(privKey)
  if err != nil {
    return "", "", err
  }

  return strings.TrimSuffix(string(privKey[:]), "\n"), strings.TrimSuffix(string(pubKey[:]), "\n"), nil
}

func getPhysicalIpAddr() (string, error) {
  var myInet string
  if(config.InterfaceName == ""){
    inet, err := ifconfig.GetFirstIpOfFirtIf()
    if err != nil {
      return "", err
    }
    myInet = inet
  } else {
    inet, err := ifconfig.GetIpOfIf(config.InterfaceName)
    if err != nil {
      return "", err
    }
    myInet = inet
  }

  ipAddr, _, err := net.ParseCIDR(myInet)
  if err != nil {
    return "", err
  }
  return ipAddr.String(), nil
}



func initialize(ConsulClient *api.Client, physicalIpAddr string, privKey string, pubKey string) error{
  _, wgIpNet, err := net.ParseCIDR(config.WGRange)
  if err != nil {
    return err
  }

  var ConsulKV *api.KV
  ConsulKV = ConsulClient.KV()

  kvpairsWgRange, _, err := ConsulKV.Get(config.LongKVPrefix + "range", nil)
  if err != nil {
    return err
  }
  if kvpairsWgRange == nil || string(kvpairsWgRange.Value[:]) != config.WGRange {
    log.Println("The wireguard IP range doesn't exist, willing to create it right now")
    _, err := ConsulKV.Put(&api.KVPair{Key: config.LongKVPrefix + "range", Value: []byte(config.WGRange)}, nil)
    if err != nil {
      return err
    }

  }


  myWgConfigMap := make(map[string]string)
  myWgConfigKVPairs, _, err := ConsulKV.List(config.LongKVPrefix + "nodes/" + physicalIpAddr, nil)
  if err != nil {
    return err
  }
  for _, myWgConfigKVPair := range myWgConfigKVPairs {
    myWgConfigMap[myWgConfigKVPair.Key[strings.LastIndex(myWgConfigKVPair.Key, "/") + 1:]] = string(myWgConfigKVPair.Value[:])
  }

  if myWgConfigKVPairs != nil {
    fmt.Println("I already picked wg ip and registred it into KV", myWgConfigMap["ip"])

    if(wgIpNet.Contains(net.ParseIP(myWgConfigMap["ip"])) &&
      myWgConfigMap["port"] == strconv.Itoa(config.WGPort) &&
      myWgConfigMap["pubkey"] == pubKey &&
      myWgConfigMap["allowedips"] == myWgConfigMap["ip"] + "/32" + config.WGAllowedIPs) {

      fmt.Println("My registred configurations are consistent")

      started, err := ifconfig.IsInterfaceStarted(config.WGInterfaceName)
      if err != nil {
        return err
      }
      maskBits, _ := wgIpNet.Mask.Size()
      newWgInterface := wireguard.Interface{config.WGInterfaceName, fmt.Sprintf("%s/%d", myWgConfigMap["ip"], maskBits), config.WGPort, privKey}
      if(started){
        fmt.Println("I already started my wg interface")

        if(wireguard.IsWgInterfaceWellConfigured(newWgInterface)){
          fmt.Println("My interface is well configured")
          return nil
        } else {
          fmt.Println("My interface is not well configured")
          _, err = ifconfig.StopWGInterface(config.WGInterfaceName)
          if(err != nil){
            return err
          }
          return initialize(ConsulClient, physicalIpAddr, privKey, pubKey)
        }


      } else {
        fmt.Println("Will bring up my wg interface")
        _, err = ifconfig.StartWGInterface(newWgInterface.Name, newWgInterface.Address)
        if(err != nil){
          return err
        }
        _, err = wireguard.ConfigureInterface(newWgInterface)
        if(err != nil){
          return err
        }
        return initialize(ConsulClient, physicalIpAddr, privKey, pubKey)
      }


    } else {
      fmt.Println("My registred configurations are not consistent")
      _, err := ConsulKV.DeleteTree(config.LongKVPrefix + "nodes/" + physicalIpAddr, nil)
      if err != nil {
        return err
      }
      return initialize(ConsulClient, physicalIpAddr, privKey, pubKey)
    }
    

  } else {
    fmt.Println("I didn't yet picked an IP from RANGE")

    opts := &api.LockOptions{
      Key:        config.LongKVPrefix + "pick-ip-lock",
      Value:      []byte(physicalIpAddr),
      SessionOpts: &api.SessionEntry{
        Behavior: "release",
        TTL: "10s",
      },
    }
    lock, err := ConsulClient.LockOpts(opts)
    if err != nil {
      return err
    }

    stopCh := make(chan struct{})
    _, err = lock.Lock(stopCh)
    if err != nil {
      return err
    }
    //Resource locked


    //get all the picked ip
    kvpairsNodes, _, err := ConsulKV.List(config.LongKVPrefix + "nodes", &api.QueryOptions{AllowStale: false, RequireConsistent: true, UseCache: false})
    if err != nil {
      return err
    }
    var usedWgIps []string
    for _, kvpNode := range kvpairsNodes {
      usedWgIps = append(usedWgIps, string(kvpNode.Value[:]))
    }


    wgIpStart := wgIpNet.IP
    //let's pick spare ip
    util.IncIp(wgIpStart) //Skip IP Network

    //The loop goes over all ips in the network
    for myFutureWgIp := wgIpStart; wgIpNet.Contains(myFutureWgIp); util.IncIp(myFutureWgIp) {
      if(util.SliceContains(usedWgIps, myFutureWgIp.String())){
        fmt.Println(myFutureWgIp.String(), "exist, skipping...")
      } else {
        fmt.Println("Found IP", myFutureWgIp)
        //save it to /wireguard/wg0/nodes/physicalIpAddr
        
        nodeKVTxnOps := api.KVTxnOps{
          &api.KVTxnOp{
            Verb:    api.KVSet,
            Key:     config.LongKVPrefix + "nodes/" + physicalIpAddr + "/ip",
            Value:   []byte(myFutureWgIp.String()),
          },
          &api.KVTxnOp{
            Verb:    api.KVSet,
            Key:     config.LongKVPrefix + "nodes/" + physicalIpAddr + "/pubkey",
            Value:   []byte(pubKey),
          },
          &api.KVTxnOp{
            Verb:    api.KVSet,
            Key:     config.LongKVPrefix + "nodes/" + physicalIpAddr + "/port",
            Value:   []byte(strconv.Itoa(config.WGPort)),
          },
          &api.KVTxnOp{
            Verb:    api.KVSet,
            Key:     config.LongKVPrefix + "nodes/" + physicalIpAddr + "/allowedips",
            Value:   []byte(myFutureWgIp.String() + "/32" + config.WGAllowedIPs),
          },
        }
        ok, _, _, err := ConsulKV.Txn(nodeKVTxnOps, nil)
        lock.Unlock()  //Unlock Resource
        if err != nil {
          return err
        }
        if !ok {
          return fmt.Errorf("Transaction was rolled back")
        }

        // TODO: Check that ip we didn't pick broadcast IP
        // Check if there is no free ip left
        // if(util.SliceContains(usedWgIps, myFutureWgIp.String())){
        //   return fmt.Errorf("There is no spare IP left in %s CIDR", config.WGRange)
        // }

        return initialize(ConsulClient, physicalIpAddr, privKey, pubKey)

        break;
      }
    }
  }


  return nil
}


func monitorPeers(ConsulClient *api.Client, physicalIpAddr string) {
  stopMonitorKvPrefixchan := make(chan bool)
  stopMonitorNodesChan := make(chan bool)
  newPeerschan := make(chan map[string]map[string]string)
  newNodesChan := make(chan map[string]string)
  go monitorKvPrefix(ConsulClient, newPeerschan, stopMonitorKvPrefixchan)
  go monitorNodes(ConsulClient, physicalIpAddr, newNodesChan, stopMonitorNodesChan)

  for {
    select {
      case <-stopMonitorKvPrefixchan:
        fmt.Println("monitorKvPrefix goroutine stopped")
      case newPeers := <-newPeerschan:
        fmt.Println("received new peers from monitorKvPrefix goroutine")
        configureWgPeers(physicalIpAddr, newPeers)
      case <-stopMonitorNodesChan:
        fmt.Println("monitorNodes goroutine stopped")
      case nodesPhysicalIpAddr := <-newNodesChan:
        fmt.Println("received new nodes")
        removeLeftNodes(ConsulClient, nodesPhysicalIpAddr)

    }
  }
  
}

func monitorKvPrefix(ConsulClient *api.Client, newPeerschan chan map[string]map[string]string, stopMonitorKvPrefixChan chan bool) {
  var ConsulKV *api.KV
  ConsulKV = ConsulClient.KV()

  var newPeers map[string]map[string]string
  var waitIndex uint64
  waitIndex = 0

  for {
    newPeers = make(map[string]map[string]string)
    opts := api.QueryOptions{
      AllowStale: false, 
      RequireConsistent: true, 
      UseCache: false,
      WaitIndex: waitIndex,
    }
    fmt.Println("Will watch consul kv prefix in blocking query now", waitIndex)
    kvpairsNodes, meta, err := ConsulKV.List(config.LongKVPrefix + "nodes", &opts)
    if err != nil {
      // Prevent backend errors from consuming all resources.
      log.Fatal(err)
      time.Sleep(time.Second * 2)
      continue
    }
    for _, kvpNode := range kvpairsNodes {
      physicalIpAddr := strings.Split(kvpNode.Key, "/")[3]
      field := strings.Split(kvpNode.Key, "/")[4]
      value := string(kvpNode.Value[:])
      if _, ok := newPeers[physicalIpAddr]; !ok {
        newPeers[physicalIpAddr] = make(map[string]string)
        newPeers[physicalIpAddr]["endpoint"] = physicalIpAddr
      }
      newPeers[physicalIpAddr][field] = value
    }
    newPeerschan <- newPeers

    waitIndex = meta.LastIndex
  }
  stopMonitorKvPrefixChan <- true
}

func configureWgPeers(myPhysicalIpAddr string, newPeers map[string]map[string]string) {
  peers, err := wireguard.GetPeers(config.WGInterfaceName)
  if err != nil {
    log.Fatal(err)
    return
  }
  util.PrintPeersMap(peers)
  util.PrintPeersMap(newPeers)

  for physicalIpAddrKey, peer := range peers {
    
    if _, ok := newPeers[physicalIpAddrKey]; !ok { // peer doesn't exit in newPeers anymore
      fmt.Println("Removing a peer that doesn't exist anymore")
      _, err = wireguard.RemovePeer(config.WGInterfaceName, peer["pubkey"])
      if(err != nil){
        log.Fatal(err)
      }
    } else {
      if(peer["pubkey"] != newPeers[physicalIpAddrKey]["pubkey"]){
        fmt.Println("Reconfiguring a Peer that has same endpoint and different public key")
        _, err = wireguard.RemovePeer(config.WGInterfaceName, peer["pubkey"])
        if(err != nil){
          log.Fatal(err)
        }
        _, err = wireguard.ConfigurePeer(config.WGInterfaceName, peer)
        if(err != nil){
          log.Fatal(err)
        }
      } else {
        if(!util.IsTheSameAllowedips(peer["allowedips"], newPeers[physicalIpAddrKey]["allowedips"]) || 
          peer["port"] != newPeers[physicalIpAddrKey]["port"] || 
          peer["endpoint"] != newPeers[physicalIpAddrKey]["endpoint"]){

          fmt.Println("Reconfiguring a Peer that changes its params")
          _, err = wireguard.ConfigurePeer(config.WGInterfaceName, peer)
          if(err != nil){
            log.Fatal(err)
          }
        }
      }
    }
  }

  for physicalIpAddrKey, peer := range newPeers {
    if(myPhysicalIpAddr == physicalIpAddrKey){ // If physicalIpAddrKey is my ip, skip it
      continue
    }

    if _, ok := peers[physicalIpAddrKey]; !ok { // new peer doesn't exist in peers
      fmt.Println("Adding New Peer")
      _, err = wireguard.ConfigurePeer(config.WGInterfaceName, peer)
      if(err != nil){
        log.Fatal(err)
      }
    }
  }

}



func monitorNodes(ConsulClient *api.Client, physicalIpAddr string, newNodesChan chan map[string]string, stopMonitorNodesChan chan bool) {
  opts := &api.LockOptions{
    Key:        config.LongKVPrefix + "monitor-nodes-lock",
    Value:      []byte(physicalIpAddr),
    SessionOpts: &api.SessionEntry{
      Behavior: "release",
      TTL: "10s",
    },
  }
  lock, err := ConsulClient.LockOpts(opts)
  if err != nil {
    log.Fatal(err)
  }
  stopCh := make(chan struct{})
  _, err = lock.Lock(stopCh)
  if err != nil {
    log.Fatal(err)
  }

  var ConsulCatalog *api.Catalog
  ConsulCatalog = ConsulClient.Catalog()
  var waitIndex uint64
  waitIndex = 0
  for {
    opts := api.QueryOptions{
      AllowStale: false, 
      RequireConsistent: true, 
      UseCache: false,
      WaitIndex: waitIndex,
    }
    fmt.Println("Will watch consul nodes", waitIndex)
    listNodes, meta, err := ConsulCatalog.Nodes(&opts)
    if err != nil {
      // Prevent backend errors from consuming all resources.
      log.Fatal(err)
      time.Sleep(time.Second * 2)
      continue
    }

    newNodes := make(map[string]string)
    for _, node := range listNodes {
      newNodes[node.Address] = node.ID
    }

    newNodesChan <- newNodes

    waitIndex = meta.LastIndex
  }
  stopMonitorNodesChan <- true
}

func removeLeftNodes(ConsulClient *api.Client, nodesPhysicalIpAddr map[string]string) {
  var ConsulKV *api.KV
  ConsulKV = ConsulClient.KV()

  peers, err := wireguard.GetPeers(config.WGInterfaceName)
  if err != nil {
    log.Fatal(err)
    return
  }
  for physicalIpAddrKey, _ := range peers {
    if _, ok := nodesPhysicalIpAddr[physicalIpAddrKey]; !ok { //Peer doesn't exist in Consul Catalog anymore
      fmt.Println("Release node IP from the pool")
      _, err := ConsulKV.DeleteTree(config.LongKVPrefix + "nodes/" + physicalIpAddrKey, nil)
      if err != nil {
        log.Fatal(err)
      }
    }
  }
}


