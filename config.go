package main

import (
  "flag"

)


type Config struct {
  KVPrefix string
  LongKVPrefix string
  WGInterfaceName string
  WGConfigFolder string
  WGRange string
  WGPort int
  WGAllowedIPs string
  InterfaceName string
}

var config Config

func init() {
  flag.StringVar(&config.KVPrefix, "kv-prefix", "autowire", "Prefix in KV store where configurations will be stored")
  flag.StringVar(&config.WGInterfaceName, "wg-if-name", "wg0", "Wireguard interface name")
  flag.StringVar(&config.WGConfigFolder, "wg-config-folder", "/etc/wireguard", "Wireguard config folder")
  flag.StringVar(&config.WGRange, "wg-range", "192.168.10.0/24", "Wireguard CIDR")
  flag.IntVar(&config.WGPort, "wg-port", 51820, "Wireguard Port")
  flag.StringVar(&config.WGAllowedIPs, "allowed-ips", "", "Wireguard Allowed IPs (comma separated cidr)")
  flag.StringVar(&config.InterfaceName, "if-name", "", "The network ip address of this interface will be used for wireguard endpoint")
}


func initConfig() error {

  // ToDo: Update config from environment variables.
  // processEnv() 

  config.LongKVPrefix = config.KVPrefix + "/" + config.WGInterfaceName + "/"

  if config.WGAllowedIPs != "" {
    config.WGAllowedIPs = "," + config.WGAllowedIPs
  }

  // if config.LogLevel != "" {
  //   log.SetLevel(config.LogLevel)
  // }


  return nil
}