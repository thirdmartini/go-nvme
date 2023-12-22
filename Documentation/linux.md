# Linux Usage
Howto use the target with a Linux Host

### Enabling nvme/tcp on linux
You may need to install extra kernel modules first:
```
apt-get install linux-modules-extra-$(uname -r)
```

And you also need nvme-cli
```
apt install nvme-cli
```

Load up the driver:
```
modprobe nvme-tcp
```

## Using the Target

### Discover available Subsystems  on the target
```
root@ubuntu:/home/root#  nvme discover  -t tcp -a 10.0.0.168 -s 4420

Discovery Log Number of Records 1, Generation counter 0
=====Discovery Log Entry 0======
trtype:  tcp
adrfam:  ipv4
subtype: nvme subsystem
treq:    not specified, sq flow control disable supported
portid:  1
trsvcid: 4420
subnqn:  nqn.2020-20.com.thirdmartini:uuid:f81d4fae-7dec-11d0-a765-00a0c91e6bf6
traddr:  10.0.0.168
sectype: none
```

### Connect to the target:

Connect to the discovered target
```
root@ubuntu:/home/root#  nvme connect  -t tcp -a 10.0.0.168 -s 4420 -n nqn.2020-20.com.thirdmartini:uuid:f81d4fae-7dec-11d0-a765-00a0c91e6bf6
```

And see if the device showed up:

```
root@ubuntu:/home/root#  nvme list
Node             SN                   Model                                    Namespace Usage                      Format           FW Rev
---------------- -------------------- ---------------------------------------- --------- -------------------------- ---------------- --------
/dev/nvme0n1     Volume0              ThirdMartini NVME                        1           1.10  TB /   1.10  TB    512   B +  0 B   1.3.0tes
```

### Lets do some io:
Install fio
```
root@ubuntu:/home/root# apt-get install fio
..
..
```


And run some io
```
root@ubuntu:/home/root# fio --filename=/dev/nvme0n1 --direct=1 --rw=randread --ioengine=libaio --bs=4k --iodepth=128  --name=fill --time_based=1 --runtime=30s
fill: (g=0): rw=randread, bs=(R) 4096B-4096B, (W) 4096B-4096B, (T) 4096B-4096B, ioengine=libaio, iodepth=128
fio-3.16
Starting 1 process
fio: cache invalidation of /dev/nvme0n1 failed: Resource temporarily unavailable
Jobs: 1 (f=1): [r(1)][100.0%][r=105MiB/s][r=26.8k IOPS][eta 00m:00s]
fill: (groupid=0, jobs=1): err= 0: pid=2045: Thu Oct 14 23:55:05 2021
  read: IOPS=24.4k, BW=95.5MiB/s (100MB/s)(2864MiB/30005msec)
    slat (nsec): min=1206, max=9010.3k, avg=36294.87, stdev=94638.83
    clat (usec): min=1063, max=2426.2k, avg=5156.40, stdev=31976.48
     lat (usec): min=1065, max=2426.2k, avg=5192.75, stdev=31976.17
    clat percentiles (usec):
     |  1.00th=[   1844],  5.00th=[   2737], 10.00th=[   2802],
     | 20.00th=[   3392], 30.00th=[   3720], 40.00th=[   4146],
     | 50.00th=[   4621], 60.00th=[   4948], 70.00th=[   5538],
     | 80.00th=[   5866], 90.00th=[   6521], 95.00th=[   6849],
     | 99.00th=[  10814], 99.50th=[  13960], 99.90th=[  21365],
     | 99.95th=[  25297], 99.99th=[2432697]
   bw (  KiB/s): min=69616, max=108984, per=100.00%, avg=106626.95, stdev=5098.22, samples=55
   iops        : min=17404, max=27246, avg=26656.73, stdev=1274.55, samples=55
  lat (msec)   : 2=1.77%, 4=32.02%, 10=65.03%, 20=1.05%, 50=0.12%
  lat (msec)   : 100=0.01%, >=2000=0.02%
  cpu          : usr=2.44%, sys=5.03%, ctx=88737, majf=0, minf=147
  IO depths    : 1=0.1%, 2=0.1%, 4=0.1%, 8=0.1%, 16=0.1%, 32=0.1%, >=64=100.0%
     submit    : 0=0.0%, 4=100.0%, 8=0.0%, 16=0.0%, 32=0.0%, 64=0.0%, >=64=0.0%
     complete  : 0=0.0%, 4=100.0%, 8=0.0%, 16=0.0%, 32=0.0%, 64=0.0%, >=64=0.1%
     issued rwts: total=733241,0,0,0 short=0,0,0,0 dropped=0,0,0,0
     latency   : target=0, window=0, percentile=100.00%, depth=128

Run status group 0 (all jobs):
   READ: bw=95.5MiB/s (100MB/s), 95.5MiB/s-95.5MiB/s (100MB/s-100MB/s), io=2864MiB (3003MB), run=30005-30005msec

Disk stats (read/write):
  nvme0n1: ios=0/0, merge=0/0, ticks=0/0, in_queue=0, util=0.00%
```

### Disconnecting from the target 

```
root@ubuntu:/home/root#  nvme disconnect -n nqn.2020-20.com.thirdmartini:uuid:f81d4fae-7dec-11d0-a765-00a0c91e6bf6
NQN:nqn.2020-20.com.thirdmartini:uuid:f81d4fae-7dec-11d0-a765-00a0c91e6bf6 disconnected 1 controller(s)

```