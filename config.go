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
  InterfaceName string
}

var config Config

func init() {
  flag.StringVar(&config.KVPrefix, "kv-prefix", "autowire", "Auth bearer token to use")
  flag.StringVar(&config.WGInterfaceName, "wg-if-name", "wg0", "Wireguard interface name")
  flag.StringVar(&config.WGConfigFolder, "wg-config-folder", "/etc/wireguard", "Wireguard interface name")
  flag.StringVar(&config.WGRange, "wg-range", "192.168.10.0/24", "Wireguard interface name")
  flag.IntVar(&config.WGPort, "wg-port", 51820, "Wireguard interface name")
  flag.StringVar(&config.InterfaceName, "if-name", "", "Wireguard interface name")
}


func initConfig() error {

  // ToDo: Update config from environment variables.
  // processEnv() 

  config.LongKVPrefix = config.KVPrefix + "/" + config.WGInterfaceName + "/"
  // if config.LogLevel != "" {
  //   log.SetLevel(config.LogLevel)
  // }


  return nil
}