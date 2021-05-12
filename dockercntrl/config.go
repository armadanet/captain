package dockercntrl

import (
  "github.com/docker/docker/api/types/container"
  "github.com/docker/docker/api/types/mount"
  "github.com/docker/go-connections/nat"
  "strconv"

  // "github.com/phayes/freeport"
  // "github.com/google/uuid"
  // "strconv"
  "github.com/armadanet/spinner/spincomm"
)

// Limits hold the set of limits for a given container.
type Limits struct {
  CPUShares int64     `json:"cpushares"`
  Memory    int64     `json:"Memory limit (in bytes)"`
}

// Config represents the configuration to build a new container.
type Config struct {
  Id        string      `json:"nebula_id,omitempty"`
  Image     string      `json:"image"`
  Cmd       []string    `json:"command"`
  Tty       bool        `json:"tty"`
  Name      string      `json:"name"`
  Limits    *Limits     `json:"limits"`
  Env       []string    `json:"env"`
  Port      int64       `json:"port"`
  Storage   bool        `json:"storage"`
  mounts    []mount.Mount
}

const (
  LABEL = "nebula-id"
)

var (
  HostPort = 8080
)

func TaskRequestLimits(req map[string]*spincomm.ResourceRequirement) *Limits {
  limits := Limits{}
  if val, ok := req["CPU"]; ok {
    limits.CPUShares = val.Requested
  }
  if val, ok := req["Memory"]; ok {
    limits.Memory = val.Requested
  }

  return &limits
}

func TaskRequestConfig(task *spincomm.TaskRequest) (*Config, error) {
  config := &Config{
    Id: task.GetTaskId().GetValue(),
    Image: task.GetImage(),
    Cmd: task.GetCommand(),
    Tty: task.GetTty(),
    Name: task.GetTaskId().GetValue(),
    Limits: TaskRequestLimits(task.GetTaskspec().GetResourceMap()),
    Env: task.GetEnv(),
    Port: task.GetPort(),
    Storage: false,
  }
  return config, nil
}

func (c *Config) AddMount(name string) {
  c.mounts = []mount.Mount{
    {
      Type: mount.TypeVolume,
      Source: name,
      Target: "/data",
    },
  }
}

// Converts a dockercntrl.Config into the necessary docker-go-sdk configs
func (c *Config) convert() (*container.Config, *container.HostConfig, error) {
  var id string
  if c.Id != "" {id = c.Id}
  portS := strconv.FormatInt(c.Port, 10)
  port, _ := nat.NewPort("tcp", portS)
  config := &container.Config{
    Image: c.Image,
    Cmd: c.Cmd,
    Tty: c.Tty,
    Env: c.Env,
    Labels: map[string]string{
      LABEL: id, // To identify as belonging to nebula
    },
    ExposedPorts: nat.PortSet{
        port: struct{}{},
    },
  }

  //hostPort, _ := freeport.GetFreePort()
  //hostPortS := strconv.Itoa(hostPort)
  hostPortS := strconv.Itoa(HostPort)
  HostPort = HostPort + 1

  hostConfig := &container.HostConfig{
    Resources: container.Resources{
      NanoCPUs: c.Limits.CPUShares * 1000000000,
    },
    Mounts: c.mounts,
    PortBindings: nat.PortMap{
      port: []nat.PortBinding{{HostIP: "", HostPort: hostPortS}},
    },
    //NetworkMode: "spinner-local-network",
    AutoRemove: true,
  }

  // If port is supplied, open that port on the container thru
  // a random open port on the host machine.
  // if c.Port != 0 {
  //   port, err := nat.NewPort("tcp", strconv.FormatInt(c.Port,10))
  //   if err != nil {return config, hostConfig, err}
  //   config.ExposedPorts = nat.PortSet{port: struct{}{}}
  //   openPort, err := freeport.GetFreePort()
  //   if err != nil {return config, hostConfig, err}
  //   hostConfig.PortBindings = nat.PortMap{
  //     port: []nat.PortBinding{{HostIP: "0.0.0.0", HostPort: strconv.FormatInt(c.Port, 10)}},
  //   }
  // }

  return config, hostConfig, nil
}
