# Test cases for the clustertest 

## Requirements
* [clustertest](https://github.com/yuuki0xff/clustertest)
* [Proxmox VE](https://www.proxmox.com/en/proxmox-ve)


## Setup
1. Setup the Proxmox VE cluster.
2. Setup clustertest server.
3. Create templates on local storage.
    1. Import the base image.  See [import-template.sh](./import-template.sh).
    2. Run the `node-setup.sh` on the VM.
    3. Shutdown.
    4. Convert to template.
4. Copy templates to all nodes.  


## Execute tests
1. Start test cases.  
   `clustertest task start ./clustertest/ltp-fail.generated.*.yaml`
2. Wait for test cases.  
   `clustertest task wait`
3. Show results.  
   `clustertest task output`
