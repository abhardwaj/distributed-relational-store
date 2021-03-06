package main

/**
 * Barista Server
 *
 * @author: Anant Bhardwaj
 * @date: 03/23/2014
 */

import "fmt"
import "barista"
import "git.apache.org/thrift.git/lib/go/thrift"
import "net"
import "strings"
import "handler"

const PORT_BINARY = ":9000"
const PORT_JSON = ":9090"
var ADDRS = []string {"128.52.161.243", "128.52.160.104", "128.52.161.242", "128.52.160.122", "128.52.161.24"}
var PG_PORTS = []string {"5432", "5432", "5432", "5432", "5432"}
var SP_PORTS = []string {":9001", ":9001", ":9001", ":9001", ":9001"}
var use_zookeeper = false

func main() {  
  binary_protocol_factory := thrift.NewTBinaryProtocolFactoryDefault()
  json_protocol_factory := thrift.NewTJSONProtocolFactory()
  transport_factory := thrift.NewTTransportFactory()

  addrs, err := net.InterfaceAddrs()
  addr := ""

  if err != nil || len(addrs) < 2 {
     fmt.Println("Error getting ip: ", err)
     addr = "localhost"
  } else {
     addr = strings.Split(addrs[1].String(), "/")[0]
  }

  binary_transport, err := thrift.NewTServerSocket(addr + PORT_BINARY)
  json_transport, err := thrift.NewTServerSocket(addr + PORT_JSON)
 
  if err != nil {
    fmt.Println("Error: ", err)
    return
  }

  me := -1
  for i, server := range ADDRS {
     if addr == server {
        me = i
     }
  }

  if me == -1 {
     fmt.Println("Error: I am not listed in the servers")
     return
  }

  handler1 := handler.NewBaristaHandler(ADDRS, me, PG_PORTS, SP_PORTS, false, false, use_zookeeper)
  processor := barista.NewBaristaProcessor(handler1)
  binary_server := thrift.NewTSimpleServer4(processor, binary_transport, transport_factory, binary_protocol_factory)
  json_server := thrift.NewTSimpleServer4(processor, json_transport, transport_factory, json_protocol_factory)

  fmt.Println("Starting the Barista server (Binary Mode) on ", addr + PORT_BINARY)
  go binary_server.Serve() 

  fmt.Println("Starting the Barista server (JSON Mode) on ", addr + PORT_JSON)
  json_server.Serve() 
}
