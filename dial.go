package captain

import (
  "github.com/open-nebula/captain/dockercntrl"
  "github.com/open-nebula/comms"
  "log"
)

// Dial a socket connection to a given url. Listen for reads and writes
func (c *Captain) Dial(dailurl string) error {
  socket, err := comms.EstablishSocket(dailurl)
  if err != nil {return err}
  var config dockercntrl.Config
  socket.Start(config)
  go c.connect(socket.Reader(), socket.Writer())
  return nil
}
// no concurrency on captain side: execute task one by one
func (c *Captain) connect(read chan interface{}, write chan interface{}) {
  for {
    select {
    case data, ok := <- read:
      if !ok {break}
      config, ok := data.(*dockercntrl.Config)
      if !ok {break}
      log.Println(config)
      write <- c.ExecuteConfig(config)
    }
  }
}
