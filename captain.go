// Package captain leads and manages the containers on a single machine
package captain

import (
  "log"
  "github.com/armadanet/captain/dockercntrl"
  "github.com/armadanet/spinner/spinresp"
  "github.com/armadanet/comms"
  "fmt"
)

// Captain holds state information and an exit mechanism.
type Captain struct {
  state   *dockercntrl.State
  exit    chan interface{}
  storage bool
  name    string
}

// Constructs a new captain.
func New(name string) (*Captain, error) {
  state, err := dockercntrl.New()
  if err != nil {return nil, err}
  return &Captain{
    state: state,
    storage: false,
    name: name,
  }, nil
}

// Connects to a given spinner and runs an infinite loop.
// This loop is because the dial runs a goroutine, which
// stops if the main thread closes.
func (c *Captain) Run(beaconURL string, selfSpin bool) {
  // create local bridge network
  bridge, err := c.state.GetNetwork()
  if err != nil {
    log.Println(err)
    return
  }
  // attach self to bridge network
  err = c.state.AttachNetwork(c.name, bridge.ID)
  if err != nil {
    log.Println(err)
    return
  }
  // start cargo container
  c.ConnectStorage()
  // query beacon for a spinner
  spinner_name, err := c.QueryBeacon(beaconURL, selfSpin)
  if err != nil {
    log.Println(err)
    return
  }
  // Register to selected spinner and start acting as a worker
  err = c.Dial("ws://"+spinner_name+":5912/join")
  if err != nil {
    log.Println(err)
    return
  }
  // exit
  select {
  case <- c.exit:
  }
}

type BeaconResponse struct {
  Valid         bool    `json:"Valid"`  // true if find a spinner
  Token         string  `json:"Token"`
  Ip            string  `json:"Ip"`
  OverlayName   string  `json:"OverlayName"`
  ContainerName string  `json:"ContainerName"`
}

func (c *Captain) QueryBeacon(beaconURL string, selfSpin bool) (string, error) {
  var res BeaconResponse
  // query beacon for spinner
  err := comms.SendGetRequest(beaconURL, &res)
  if err != nil {return "",err}

  // selfSpin
  if selfSpin || !res.Valid {
    log.Println("Self-spining... Building up connection to spinner...")
    res.OverlayName, res.ContainerName= c.SelfSpin()
    // just attach the overlay since local spinner already joined swarm
    err = c.state.JoinOverlay(c.name, res.OverlayName)
    if err != nil {return "",err}
  } else {
    // join swarm and connect the selected spinner
    err = c.state.JoinSwarmAndOverlay(res.Token, res.Ip, c.name, res.OverlayName)
    if err != nil {return "",err}
  }

  // return the selected spinner id (name)
  return res.ContainerName, nil
}

// Executes a given config, waiting to print output.
// Should be changed to logging or a logging system.
// Kubeedge uses Mosquito for example.
func (c *Captain) ExecuteConfig(config *dockercntrl.Config, write chan interface{}) {
  container, err := c.state.Create(config)
  if err != nil {
    log.Println(err)
    return
  }
  // // For debugging
  // config.Storage = true
  // // ^^ Remove
  // if config.Storage {
  //   log.Println("Storage in Config")
  //   if !c.storage {
  //     log.Println("Establishing Storage")
  //     c.storage = true
  //     c.ConnectStorage()
  //   } else {
  //     log.Println("Storage already exists")
  //   }
  //   err = c.state.NetworkConnect(container)
  //   if err != nil {
  //     log.Println(err)
  //     return
  //   }
  // } else {
  //   log.Println("No storage in config")
  // }

  // connect all new containers under captain to bridge network
  err = c.state.NetworkConnect(container)
  if err != nil {
    log.Println(err)
    return
  }
  // start and wait this container
  s, err := c.state.Run(container)
  if err != nil {
    log.Println(err)
    return
  }
  log.Println("Task Container Output: ")
  fmt.Println(*s)
  // for system containers: write = nil
  if write != nil {
    write <- &spinresp.Response{
      Id: config.Id,
      Code: spinresp.Success,
      Data: *s,
    }
  }
}
