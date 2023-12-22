# ESXi Usage

## Enabling nvme-tcp on esx
This Requires ESXi 7.0U3 or newer (https://tinkertry.com/easy-update-to-latest-esxi)
ESX7.0U3 includes  a functioning NVME/TCP driver

### Configure NQN name support
By default ESXi does not configure a hostnqn.  
```
[root@localhost:~] esxcfg-module -s 'vmknvme_hostnqn_format=0' vmknvme
[root@localhost:~] reboot
```

After the reboot verify that a new NQN was created and assigned properly/

```
[root@localhost:~] esxcli nvme info get
Host NQN:
nqn.2014-08.org.nvmexpress:uuid:5ffefdd7-0a51-3700-0b16-001e6792303a
```

### Enable and Configure NVME/TCP vmHBA

Load the nvmetcp driver
```
[root@localhost:~] esxcli system module load -m nvmetcp  
```

And configure the vmnic (Physical Nic) for use. This will create a new vmhba that is bound to this physical network adapter. The nick you choose here should an attached vmk nick (VM Kernel Netowkr Interface)
```
[root@localhost:~] esxcli nvme fabrics enable -p TCP -d vmnic0
true
```

We should now be able to see the new NVME/TC hba:
```
[root@localhost:~] esxcli storage adapter list
HBA Name  Driver    Link State  UID                           Capabilities  Description
--------  --------  ----------  ----------------------------  ------------  -----------
vmhba0    vmw_ahci  link-n/a    sata.vmhba0                                 (0000:02:01.0) VMware Inc Vmware Virtual SATA Controller
vmhba65   nvmetcp   link-n/a    tcp.vmnic0:00:0c:29:34:2a:0c                VMware NVMe over TCP Storage Adapter
```

To finish the setup, we also need to tell the associated vmk NIC that it can be used for NVME/TCP traffic
```
esxcli network ip interface tag add -i vmk0 -t NVMeTCP 
```

At this point everything is setup and we can start using the target software.

## Attaching to a remote NVME/TCP Target

### Discovering Available Target Subsystems
```
[root@localhost:~] esxcli nvme fabrics discover -a vmhba65 -i 10.0.0.168 -p 4420
Transport Type  Address Family  Subsystem Type  Controller ID  Admin Queue Max Size  Transport Address  Transport Service ID  Subsystem NQN                                                           Connected
--------------  --------------  --------------  -------------  --------------------  -----------------  --------------------  ----------------------------------------------------------------------  ---------
TCP             IPv4            NVM                     65535                  8192  10.0.0.168         4420                  nqn.2020-20.com.thirdmartini:uuid:f81d4fae-7dec-11d0-a765-00a0c91e6bf6      false
```

### Connecting to the Subsystem
```
[root@localhost:~] esxcli nvme fabrics  connect -a vmhba65 -i 10.0.0.168 -p 4420 -s  nqn.2020-20.com.thirdmartini.nvme:demo-volume
```

And verify that everything went as expected:
```
[root@localhost:~]  esxcli storage core device list
uuid.2b6522e97f86428c91e193efa93c864f
   Display Name: NVMe TCP Disk (uuid.2b6522e97f86428c91e193efa93c864f)
   Has Settable Display Name: false
   Size: 1048576
   Device Type: Direct-Access
   Multipath Plugin: HPP
   Devfs Path: /vmfs/devices/disks/uuid.2b6522e97f86428c91e193efa93c864f
   Vendor: NVMe
   Model: ThirdMartini NVME
   Revision: 1.3.
   SCSI Level: 7
   Is Pseudo: false
   Status: degraded
   Is RDM Capable: false
   Is Local: false
   Is Removable: false
   Is SSD: true
   Is VVOL PE: false
   Is Offline: false
   Is Perennially Reserved: false
   Queue Full Sample Size: 0
   Queue Full Threshold: 0
   Thin Provisioning Status: yes
   Attached Filters:
   VAAI Status: supported
   Other UIDs: vml.072b6522e97f86428c91e193efa93c864f
   Is Shared Clusterwide: true
   Is SAS: false
   Is USB: false
   Is Boot Device: false
   Device Max Queue Depth: 126
   No of outstanding IOs with competing worlds: 32
   Drive Type: unknown
   RAID Level: unknown
   Number of Physical Drives: unknown
   Protection Enabled: false
   PI Activated: false
   PI Type: 0
   PI Protection Mask: NO PROTECTION
   Supported Guard Types: NO GUARD SUPPORT
   DIX Enabled: false
   DIX Guard Type: NO GUARD SUPPORT
   Emulated DIX/DIF Enabled: false
```

### Disconnecting 
```
[root@localhost:~]  esxcli nvme fabrics  disconnect -a vmhba65 -s nqn.2020-20.com.thirdmartini.nvme:demo-volume
```

### Other Useful commands
Listing Connected namespaces
```
[root@localhost:~] esxcli nvme controller list
Name                                                                                            Controller Number  Adapter  Transport Type  Is Online
----------------------------------------------------------------------------------------------  -----------------  -------  --------------  ---------
nqn.2020-20.com.thirdmartini:uuid:f81d4fae-7dec-11d0-a765-00a0c91e6bf6#vmhba65#10.0.0.168:4420                266  vmhba65  TCP                  true

[root@localhost:~] esxcli nvme namespace list
Name                                   Controller Number  Namespace ID  Block Size  Capacity in MB
-------------------------------------  -----------------  ------------  ----------  --------------
uuid.2b6522e97f86428c91e193efa93c864f                266             1         512         1048576
```


### Deleting retained connections
By default ESX will remeber past targets and try to reconnect to them on reboot. You can wipe one or more of them using:

```
[root@localhost:~] configstorecli config current delete -c esx -g storage_nvmeof -k nvme_connections --all
```

