module github.com/go-hypervisor/virtio

go 1.17

require golang.org/x/sys v0.0.0-20211004093028-2c5d950f24ef

// until merged https://golang.org/cl/354269 and https://golang.org/cl/354271
replace golang.org/x/sys => github.com/zchee/golang-sys v0.0.0-20211006210402-cad64370c6dc
