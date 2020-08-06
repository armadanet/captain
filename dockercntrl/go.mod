module github.com/armadanet/captain/dockercntrl

go 1.13

replace github.com/armadanet/spinner => ../../spinner

require (
	github.com/Microsoft/go-winio v0.4.14 // indirect
	github.com/armadanet/spinner v0.0.0-00010101000000-000000000000
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/docker/docker v1.13.1
	github.com/docker/go-connections v0.4.0
	github.com/docker/go-units v0.4.0 // indirect
	github.com/opencontainers/go-digest v1.0.0-rc1 // indirect
	github.com/phayes/freeport v0.0.0-20180830031419-95f893ade6f2
	github.com/pkg/errors v0.9.1 // indirect
	golang.org/x/net v0.0.0-20200114155413-6afb5195e5aa
)
