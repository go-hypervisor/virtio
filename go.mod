module github.com/go-hypervisor/virtio

go 1.17

require golang.org/x/sys v0.0.0-20211003122950-b1ebd4e1001c

// until https://golang.org/cl/352209 merged
replace golang.org/x/sys => github.com/zchee/golang-sys v0.0.0-20211002173438-7d04f3b7eb76
