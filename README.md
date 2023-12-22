# Disclaimer 
**Use at your own risk, do not store any data you care about on this target.**  

This was a proof of concept project to play with the NVME/TCP support in ESX 6.x and is sufficient to run VMs from it.
But I would not trust it with my data.  There are a lot of mode and log pages that are not supported, and NVME 
client drivers have become more and more strict about that,  so your mileage may vary. 

This code base also grew pretty organically as I worked through the NVME specification, so its quite a mess. 
I'll refactor it one day... 

That said feel free to experiment, find bugs, fix issues etc.

# What is This?
An implementation fo NVME over TCP target and client in pure golang.  Why?  Reasons.... 

## Building

```
$ go build github.com/thirdmartini/go-nvme/cmd/nvmed 
```

or  use the make file
```
$ make 
```

To build with CEPH/RBD bridge support:
```
$ go build --tags ceph,luminous github.com/thirdmartini/go-nvme/cmd/nvmed
```

## Running 

Create a configuration file named target.yaml. see config/example.yml
```
cat targets.yaml:

targets:
 - name: "nqn.2020-20.com.thirdmartini.nvme:null"
   uuid: "2eff04dd-745a-4fc8-9f5f-10432b13a04f"
   type: "null"
   modelname: "ThirdMartini NVME"
   firmwareversion: "1.3.0test"
   serialnumber: "Volume0"
   options:

 - name: "nqn.2020-20.com.thirdmartini.nvme:ceph:volume0"
   uuid: "c5f83782-6bb5-47bc-a78a-b68a5af85974"
   modelname: "cephvolume"
   firmwareversion: "1.0"
   serialnumber: "0000000001"
   type: "rbd"
   options:
     cluster: "ceph"
     user: "admin"
     pool: "rbd"
     image: "volume0"
     
 - name: "nqn.2020-20.com.thirdmartini.nvme:file"
   uuid: "39e92c9c-f486-41e2-812b-4ebbc56665ee"
   type: "file"
   modelname: "file"
   firmwareversion: "1.0"
   type: "file"
   options:
     image: "/data/image.raw"        
```

Run the target 

```
$ ./nvmed
Local Address: 10.0.0.111
NVME Server: 10.0.0.111:4420
Profiler started at http://10.0.0.111:6060/debug/pprof
    go tool pprof -http :8080 'http://10.0.0.111:6060/debug/pprof/profile?seconds=10'
Registering Target: nqn.2020-20.com.thirdmartini.nvme:ceph:volume0 (rbd)
Registering Target: nqn.2020-20.com.thirdmartini.nvme:null (null)
Registering Target: nqn.2020-20.com.thirdmartini.nvme:file (file)
WEB UI: http://137.184.39.71:8090/
```

## Configuring Clients

To configure and use the target with an initiator, follow the OS/Network card guide that you are using. 
You can also use the built in NVME/TCP support in linux/vmware documented here:

* [VMWare ESXi](Documentation/esx.md)
* [Linux](Documentation/linux.md)


## Debuging
You can enable verbose PDU tracing by running with debug enabled
```
./nvmed --debug=0xffffffffffffffff
```

See source code for levels
