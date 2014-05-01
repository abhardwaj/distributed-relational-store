package main

/**
 * Sample Go Client for Barista
 *
 * @author: Anant Bhardwaj
 * @date: 03/23/2014
 */

import "fmt"
import "barista"
import "git.apache.org/thrift.git/lib/go/thrift"
import "sync"
import crand "crypto/rand"
import "math/big"
import "strconv"

func nrand() int64 {
  max := big.NewInt(int64(1) << 62)
  bigx, _ := crand.Int(crand.Reader, max)
  x := bigx.Int64()
  return x
}

func PrintResultSet(res *barista.ResultSet) {
  if res != nil && res.FieldNames != nil {
    for _, field_name := range *(res.FieldNames) {
      fmt.Printf("%s\t", field_name)
    }
  }

  fmt.Println()

  if res != nil && res.Tuples != nil {
    for _, tuple := range *(res.Tuples) {
      for _, cell := range *(tuple.Cells) {
        fmt.Printf("%s\t", cell)
      }
      fmt.Println()
    }
  }

  fmt.Println()
}

type Clerk struct {
  mu sync.Mutex
  me int64 // passed as clientId
  curRequest int
}


func MakeClerk() *Clerk {
  ck := new(Clerk)
  ck.me = nrand()
  ck.curRequest = 0

  return ck
}

// List of machines running on the server forming a paxos group
// 128.52.161.243:9000
// 128.52.160.104:9000
// 128.52.161.242:9000
// 128.52.160.122:9000
// 128.52.161.24:9000

// to demonstrate external consistency we create three groups
var group_1 = []string {"128.52.161.243:9000", "128.52.160.104:9000"}
var group_2 = []string {"128.52.161.242:9000", "128.52.160.122:9000"}
var group_3 = []string {"128.52.161.24:9000"}


func main() {  
  clerk := MakeClerk()

  var con *barista.Connection
  var err error

  // The clerk should keep retrying to servers in a round-robin function.
  // Ideally the clerk would retry to all the 5 servers but to demonstrate
  // the external consistency we retry only to two machines in the below code

  // open connection to a machine in group 1
  con, err = clerk.OpenConnection(group_1)
  if err != nil {
    fmt.Println(err)
    return
  }

  // create the table on a machine in group 2  
  _, err = clerk.ExecuteSQL(group_2, con, "CREATE TABLE IF NOT EXISTS courses (id text, name text)", nil)
  if err != nil {
    fmt.Println(err)
    return
  }
  
  // delete all the data on a machine in group 3  
  _, err = clerk.ExecuteSQL(group_3, con, "DELETE FROM courses", nil)
  if err != nil {
    fmt.Println(err)
    return
  }

  // insert a record to a machine in group 1  
  _, err = clerk.ExecuteSQL(group_1, con, "INSERT INTO courses values('6.831', 'UID')", nil)
  if err != nil {
    fmt.Println(err)
    return
  }

  // insert a record to a machine in group 2  
  _, err = clerk.ExecuteSQL(group_2, con, "INSERT INTO courses values('6.830', 'Databases')", nil)
  if err != nil {
    fmt.Println(err)
    return
  }

  // insert a record to a machine in group 3 
  _, err = clerk.ExecuteSQL(group_1, con, "INSERT INTO courses values('6.824', 'Distributed Systems')", nil)
  if err != nil {
    fmt.Println(err)
    return
  }

  // print all the records from a machine in group 1
  // all queries should apply in the same order on all the machines
  // all the three records should print regardless of whichever machine you query 
  res, err := clerk.ExecuteSQL(group_1, con, "SELECT * FROM courses", nil)
  if err != nil {
    fmt.Println(err)
    return
  }
  
  PrintResultSet(res)

  
  // close the connection to a machine in group 3
  // it should close this client's connection from all machines  
  err = clerk.CloseConnection(group_3, con)
  if err != nil {
    fmt.Println(err)
    return
  }
  
}

// open database connection
func (ck *Clerk) OpenConnection(addrs []string) (*barista.Connection, error) {
  ck.mu.Lock()
  defer ck.mu.Unlock()

  ck.curRequest++

  clientId := strconv.FormatInt(ck.me, 10)
  seqId := strconv.Itoa(ck.curRequest)

  user, password, database := "postgres", "postgres", "postgres"
  con_params := barista.ConnectionParams {
     ClientId: &clientId,
     SeqId: &seqId,
     User: &user,
     Password: &password,
     Database: &database }

  var err error

  for _, addr := range addrs {  
    con, err := ck.openConnection(addr, &con_params)
    if err == nil {
      return con, nil
    }
  }

  return nil, err

}


// execute SQL query
func (ck *Clerk) ExecuteSQL(addrs []string, con *barista.Connection, query string, query_params [][]byte) (*barista.ResultSet, error) {
  ck.mu.Lock()
  defer ck.mu.Unlock()

  ck.curRequest++

  clientId := strconv.FormatInt(ck.me, 10)
  seqId := strconv.Itoa(ck.curRequest)

  con.ClientId = &clientId
  con.SeqId = &seqId

  var err error

  for _, addr := range addrs {
    res, err :=  ck.executeSQL(addr, query, query_params, con)
    if err == nil {
      return res, err
    }    
  }

  return nil, err
}

// close database connection
func (ck *Clerk) CloseConnection(addrs []string, con *barista.Connection) error {
  ck.mu.Lock()
  defer ck.mu.Unlock()

  ck.curRequest++

  clientId := strconv.FormatInt(ck.me, 10)
  seqId := strconv.Itoa(ck.curRequest)

  con.ClientId = &clientId
  con.SeqId = &seqId

  var err error

  for _, addr := range addrs {
    err = ck.closeConnection(addr, con)
    if err == nil {
      return nil
    }   
  }

  return err
}


func (ck *Clerk) executeSQL(addr string, query string, query_params [][]byte, con *barista.Connection) (*barista.ResultSet, error) {
  protocolFactory := thrift.NewTBinaryProtocolFactoryDefault()
  transport, err := thrift.NewTSocket(addr)

  if err != nil {
     return nil, err
  }

  transport.Open()
  defer transport.Close()

  client := barista.NewBaristaClientFactory(transport, protocolFactory)

  res, err := client.ExecuteSql(con, query, query_params)

  if err != nil {
     return nil, err
  }

  return res, nil
}

func (ck *Clerk) openConnection(addr string, con_params *barista.ConnectionParams) (*barista.Connection, error) {
  protocolFactory := thrift.NewTBinaryProtocolFactoryDefault()
  transport, err := thrift.NewTSocket(addr)

  if err != nil {
     return nil, err
  }

  transport.Open()
  defer transport.Close()

  client := barista.NewBaristaClientFactory(transport, protocolFactory)

  con, err := client.OpenConnection(con_params)

  if err != nil {
     return nil, err
  }

  return con, nil
}

func (ck *Clerk) closeConnection(addr string, con *barista.Connection) error {
  protocolFactory := thrift.NewTBinaryProtocolFactoryDefault()
  transport, err := thrift.NewTSocket(addr)

  if err != nil {
     return err
  }

  transport.Open()
  defer transport.Close()

  client := barista.NewBaristaClientFactory(transport, protocolFactory)

  err = client.CloseConnection(con)

  if err != nil {
     return err
  }

  return nil
}

