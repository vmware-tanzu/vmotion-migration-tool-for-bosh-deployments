# Tanzu Application Service Migration Instructions
This documents the manual and semi-automated steps to vMotion or migrate a running TAS foundation to a new vCenter
environment. This is primarily useful to move running workloads without downtime to new hardware.

## Prerequisites

### Ensure all VMs are Healthy
Before attempting a migration it's important to verify that all BOSH managed VMs are in a consistent good state. Check
the running foundation by running `bosh vms` and ensuring all agents report as healthy.

### Back-up the BOSH Director with BBR
Do not skip this step! If you accidentally lose the BOSH director disk you will need to ensure you have a recent backup to avoid
a complete foundation re-install. See [BBR](https://docs.pivotal.io/ops-manager/2-10/install/backup-restore/backup-pcf-bbr.html#bbr-backup-director)


## Migrate Operations Manager
Find the Operations Manager VM in vCenter and note it's DC, cluster, resource pool etc. and then use `vmotion4bosh migrate-vm`
to move it to its new vCenter. Here's an example:

```shell
./vmotion4bosh migrate-vm \
    --source-vcenter-host 'sc3-vc-01.example.com' \
    --source-datacenter 'Datacenter' \
    --source-username 'administrator@vsphere.local' \
    --source-password 'vspheresecret' \
    --source-vmname 'operations-manager' \
    --target-vcenter-host 'sc3-vc-02.example.com' \
    --target-datacenter 'Datacenter' \
    --target-cluster 'Cluster' \
    --target-resourcepool 'tas1' \
    --target-datastore 'NFS-Datastore' \
    --network-mapping 'PCF-Infra:TAS-Infra-01' \
    --target-username 'administrator@vsphere.local' \
    --target-password 'vspheresecret'
```

If your BOSH director is multi-homed (i.e. attached to more than one network) then include multiple `--network-mapping`
flags, one for each network. For example:

```
--network-mapping 'PAS-Infra:TAS-Infra-01' --network-mapping 'MGMT-01:Mgmt-01'
```

### Backup and Upgrade Operations Manager (*only* required for versions < 2.10.17)
This step is optional and only required if your Operations Manager version is less than 2.10.17. If you have an older
version and skip this step you'll run into problems when you apply-changes. [Upgrade Operations Manager](https://docs.pivotal.io/ops-manager/2-10/install/upgrading-pcf.html):

* Export installation
* Upgrade Operations Manager to the latest version.
* Import installation

### Modify the Operations Manager installation

Once the BOSH Director VM has been migrated, you'll need to perform the following steps on the Operations Manager VM.
Move to the `/var/tempest` directory:

```shell
cd /var/tempest
```     

Run the decrypt script to decrypt the installation.yml and actual-installation.yml to temp files we can edit:
```shell
sudo -u tempest-web \
  SECRET_KEY_BASE="s" \
  RAILS_ENV=production \
  /home/tempest-web/tempest/web/scripts/decrypt \
  /var/tempest/workspaces/default/installation.yml \
  /tmp/installation.yml

sudo -u tempest-web \
  SECRET_KEY_BASE="s" \
  RAILS_ENV=production \
  /home/tempest-web/tempest/web/scripts/decrypt \
  /var/tempest/workspaces/default/actual-installation.yml \
  /tmp/actual-installation.yml         
```

Edit the decrypted `/tmp/installation.yml` and `/tmp/actual-installation.yml`, making sure you make the same edits to
both files. You will need to change the following to match the new vCenter infrastructure:

- Target cluster
- Target resource pool
- Target datastore
- Target network names
- Target vCenter address, username, and password

```shell
sudo vim /tmp/installation.yml 
sudo vim /tmp/actual-installation.yml
```
For example the sections you're looking for look like this:
```yaml
...
availability_zones:
- guid: ad74fca95f8c6d1dfc4b
 iaas_configuration_guid: 8579eb2201e847d160d3
 name: az3
 clusters:
 - guid: 152f649c586bd406e5e1
   cluster: Cluster
   resource_pool: tas-az3
   drs_rule: MUST
- guid: a67f8628c07b04294053
 iaas_configuration_guid: 8579eb2201e847d160d3
 name: az2
 clusters:
 - guid: ec2f2e73724f1fccb6bc
   cluster: Cluster
   resource_pool: tas-az2
   drs_rule: MUST
- guid: a0ed3d5c00bddbecffb8
 iaas_configuration_guid: 8579eb2201e847d160d3
 name: az1
 clusters:
 - guid: 7493a1424fa9c0d4cefd
   cluster: Cluster
   resource_pool: tas-az1
   drs_rule: MUST
...
iaas_configurations:
- guid: 8579eb2201e847d160d3
 name: default
 vcenter_host: vcenter.example.com
 vcenter_username: administrator@vsphere.local
 vcenter_password: hunter2
 datacenter: Datacenter
 persistent_datastores:
 - NFS-Datastore1
 ephemeral_datastores:
 - NFS-Datastore1
 disk_type: thin
```

Re-encrypt installation.yml and actual-installation.yml
```shell
sudo -u tempest-web \
   SECRET_KEY_BASE="s" \
   RAILS_ENV=production \
   /home/tempest-web/tempest/web/scripts/encrypt \
   /tmp/installation.yml \
   /var/tempest/workspaces/default/installation.yml

sudo -u tempest-web \
   SECRET_KEY_BASE="s" \
   RAILS_ENV=production \
   /home/tempest-web/tempest/web/scripts/encrypt \
   /tmp/actual-installation.yml \
   /var/tempest/workspaces/default/actual-installation.yml
```

### Validate Operations Manager Changes

Validate the edits by running `om staged-director-config` to verify changes have been made; ensure the cluster and 
datastore values are updated etc.
```shell
om staged-director-config --no-redact
```

If everything looks correct, validate that the vSphere service account can talk to the IaaS and has the
correct permissions required. This runs the same validations as at the beginning of an apply-changes.
```shell
om pre-deploy-check
```

## Migrate BOSH
Migrate the BOSH director to the new vCenter.

### Turn off the BOSH Resurrector
We don't want the bosh director to attempt any VM creates while we're in a half migrated state, so disable the resurrector.
```shell
bosh update-resurrection off
```

### Migrate the BOSH VM
Find the BOSH director VM in vCenter. The BOSH VM has a custom attribute key of "director" with a value of "bosh-init".
The VM ID have a format similar to `vm-c86ed50e-0bc7-4681-9c95-39c21984c5b8`. Migrate the VM to the new cluster using
the migrate-vm command.

> **NOTE** - It's _highly_ recommended to use environment variables for passwords/secrets instead of command line flags. Run `./vmotion4bosh migrate-vm -h` for more info.

Here's an example:

```shell
./vmotion4bosh migrate-vm \
    --source-vcenter-host 'sc3-vc-01.example.com' \
    --source-datacenter 'Datacenter' \
    --source-username 'administrator@vsphere.local' \
    --source-password 'vspheresecret' \
    --source-vmname 'vm-c86ed50e-0bc7-4681-9c95-39c21984c5b8' \
    --target-vcenter-host 'sc3-vc-02.example.com' \
    --target-datacenter 'Datacenter' \
    --target-cluster 'Cluster' \
    --target-resourcepool 'tas1' \
    --target-datastore 'NFS-Datastore' \
    --network-mapping 'PCF-Infra:TAS-Infra-01' \
    --target-username 'administrator@vsphere.local' \
    --target-password 'vspheresecret'
```

### Move & Rename the BOSH Persistent Disk
With the BOSH VM migrated to the new vCenter instance, SSH to Operations Manager and view the
`/var/tempest/workspaces/default/deployments/bosh-state.json` file; make note of the disk CID (copy the GUID somewhere),
it will look something like
`disk-43edf7b2-467b-4913-8142-91b24896b482.eyJ0YXJnZXRfZGF0YXN0b3JlX3BhdHRlcm4iOiJeKE5GU1xcLURhdGFzdG9yZTEpJCJ9`.
Record the disk GUID, i.e. `disk-43edf7b2-467b-4913-8142-91b24896b482`

Login to the target vSphere UI and find the migrated BOSH director VM. Shut the Director VM. Edit the BOSH VM settings
in vCenter, find the persistent disk (typically the 3rd disk in the list) and detach it
(click the X that appears to the right of the disk on hover).

Navigate to the datastore browser, find the bosh director VM (same <vm cid> as above) and copy the persistent disk to
the `pcf_disk` folder (or whatever folder name that is specifed in the bosh director tile) in the datastore
(should be named something like vm-GUID_3.vmdk*). Rename it to the disk CID that bosh is expecting
(i.e. vm-GUID_3.vmdk becomes disk-GUID.vmdk).


From the Operations Manager VM, edit the bosh-state.json
```shell
sudo vim /var/tempest/workspaces/default/deployments/bosh-state.json
```

Edit the datastore section to point to new datastore. Remove the base64 suffix on disk name (part between period and .vmdk) if it exists ie:
`disk-1983a793-2c33-474d-ad7f-8e24586ccc13.eyJ0YXJnZXRfZGF0YXN0b3JlX3BhdHRlcm4iOiJeKE5GU1xcLURhdGFzdG9yZTIpJCJ9` would be changed to `disk-1983a793-2c33-474d-ad7f-8e24586ccc13`

### Deploy the BOSH Director VM

Apply changes to director tile _only_
```shell
om apply-changes --skip-deploy-products
```

This will recreate the bosh director, create a new disk, migrate contents of the old disk and delete the old disk. The
output should look something like this:

```
attempting to apply changes to the targeted Ops Manager
{"type":"step_started","id":"bosh_product.deploying","description":"Installing BOSH"}
===== 2022-01-07 21:02:19 UTC Running "/usr/local/bin/bosh --no-color --non-interactive --tty create-env /var/tempest/workspaces/default/deployments/bosh.yml"
Deployment manifest: '/var/tempest/workspaces/default/deployments/bosh.yml'
Deployment state: '/var/tempest/workspaces/default/deployments/bosh-state.json'

Started validating
  Validating release 'bosh'... Finished (00:00:01)
  Validating release 'bosh-vsphere-cpi'... Finished (00:00:04)
  Validating release 'uaa'... Finished (00:00:02)
  Validating release 'credhub'... Finished (00:00:00)
  Validating release 'bosh-system-metrics-server'... Finished (00:00:02)
  Validating release 'os-conf'... Finished (00:00:00)
  Validating release 'backup-and-restore-sdk'... Finished (00:00:03)
  Validating release 'bpm'... Finished (00:00:01)
  Validating release 'system-metrics'... Finished (00:00:04)
  Validating cpi release... Finished (00:00:00)
  Validating deployment manifest... Finished (00:00:00)
  Validating stemcell... Finished (00:00:15)
Finished validating (00:00:34)

Started installing CPI
  Compiling package 'golang-1.16-darwin/b603604c44ff8c0cf216aa96da5702c78b165b986c296d58aee5b0db3a63275d'... Finished (00:00:00)
  Compiling package 'golang-1.16-linux/a2f62d570ee65d73efbad06ac60bf20349ed4ec301bb28c7ac24c25efc23a923'... Finished (00:00:00)
  Compiling package 'ruby-2.6.5-r0.29.0/269dc54d5306119b0e4f89be04f6c470b4876f552753815586fd1ab8ebeaa70d'... Finished (00:00:00)
  Compiling package 'iso9660wrap/f119680c1a13eb9a18822f5ac62c5657bd5ba4bcfb915955800e0398d3cb9ea7'... Finished (00:00:00)
  Compiling package 'vsphere_cpi/57272624c83bf52ffa150f14e48dd403d1f05c09e2663549da80759a0abe0bbd'... Finished (00:00:00)
  Installing packages... Finished (00:00:13)
  Rendering job templates... Finished (00:00:01)
  Installing job 'vsphere_cpi'... Finished (00:00:00)
Finished installing CPI (00:00:14)

Uploading stemcell 'bosh-vsphere-esxi-ubuntu-xenial-go_agent/621.151'... Skipped [Stemcell already uploaded] (00:00:00)

Started deploying
  Waiting for the agent on VM 'vm-b0eae799-9adf-4fb0-b74b-b157d6001548'... Finished (00:00:00)
  Running the pre-stop scripts 'unknown/0'... Finished (00:00:00)
  Draining jobs on instance 'unknown/0'... Finished (00:00:08)
  Stopping jobs on instance 'unknown/0'... Finished (00:00:01)
  Running the post-stop scripts 'unknown/0'... Finished (00:00:01)
  Unmounting disk 'disk-8b406418-3d4b-469e-927e-6e42e48a0269.eyJ0YXJnZXRfZGF0YXN0b3JlX3BhdHRlcm4iOiJeKE5GU1xcLURhdGFzdG9yZTEpJCJ9'... Finished (00:00:01)
  Deleting VM 'vm-b0eae799-9adf-4fb0-b74b-b157d6001548'... Finished (00:02:41)
  Creating VM for instance 'bosh/0' from stemcell 'sc-07ea4178-a505-4eae-b776-59392e125129'... Finished (00:00:51)
  Waiting for the agent on VM 'vm-6716052f-2cad-4549-8966-b4808a2a5505' to be ready... Finished (00:00:24)
  Attaching disk 'disk-8b406418-3d4b-469e-927e-6e42e48a0269' to VM 'vm-6716052f-2cad-4549-8966-b4808a2a5505'... Finished (00:00:49)
  Creating disk... Finished (00:00:06)
  Attaching disk 'disk-1983a793-2c33-474d-ad7f-8e24586ccc13.eyJ0YXJnZXRfZGF0YXN0b3JlX3BhdHRlcm4iOiJeKE5GU1xcLURhdGFzdG9yZTIpJCJ9' to VM 'vm-6716052f-2cad-4549-8966-b4808a2a5505'... Finished (00:00:33)
  Migrating disk content from 'disk-8b406418-3d4b-469e-927e-6e42e48a0269' to 'disk-1983a793-2c33-474d-ad7f-8e24586ccc13.eyJ0YXJnZXRfZGF0YXN0b3JlX3BhdHRlcm4iOiJeKE5GU1xcLURhdGFzdG9yZTIpJCJ9'... Finished (00:00:08)
  Detaching disk 'disk-8b406418-3d4b-469e-927e-6e42e48a0269'... Finished (00:00:20)
  Deleting disk 'disk-8b406418-3d4b-469e-927e-6e42e48a0269'... Finished (00:00:05)
```


## Migrate all BOSH Managed VMs
Before you can execute a migrate command, you will need to create a configuration file for the vcenter_migrate
tool to use. Use the following template and change it to match your source and target vCenter environments.

```yaml
---
datastores:
  old_ds1: ssd_ds1
  old_ds2: ssd_ds2
  
networks:
  old-deployment-net: TAS-Deployment
  old-services-net: TAS-Services

resource_pools:
  pas-az1: tas-az1
  pas-az2: tas-az2
  pas-az3: tas-az3

source:
  vcenter:
    host: sc3-vc-01.example.com
    username: administrator@vsphere.local
    insecure: true
  bosh:
    host: 10.212.41.141
    client_id: ops_manager
  datacenter: Datacenter

target:
  vcenter:
    host: sc3-vc-02.example.com
    username: administrator@vsphere.local
    insecure: true
  datacenter: Datacenter
  cluster: Cluster
  datastore: VSAN-Datastore1
```
The networks and resource_pools sections map the old vCenter objects on the left to the new vCenter objects on the right.
The migration requires the same number of resource pools and networks. The target networks must be in the 
same broadcast domain so the VMs can use the same IP addresses.

The `migrate` command currently requires network access to BOSH either directly via a routable network or via a local SOCKS proxy.
Once started the process can be stopped via CTRL-C and restarted later, however that will leave your foundation
in a partially migrated state with BOSH inoperable. Either restart the migration or start the migration process in the reverse direction.

Use the `vmotion4bosh` `migrate` command to move all BOSH managed VMs to another vCenter instance:

```shell
./vmotion4bosh migrate \
   --bosh-client-secret 'boshsecret' \
   --source-password 'vspheresecret' \
   --target-password 'vspheresecret' \
   --debug true 2>debug.log
```

> **NOTE** - It's _highly_ recommended to use the `--dry-run true` flag first to ensure there aren't any obvious
problems trying to migrate any of the VMs, like a missing network mapping etc.

## Finishing Up
After all VMs have been migrated you should make sure all BOSH managed VMs report they're in a running state.

```shell
bosh vms
```

If some are not, attempt to restart the VMs. Sometimes the BOSH agent's BBS gets into a bad state.

```shell
bosh -d <deployment> restart <instance>
```

The last step is an apply-changes of every deployment and VM. If there are on-demand service instances with an upgrade
errand you'll want to run those too, finally turn the resurrector back on.
```shell
om apply-changes && bosh update-resurrection on
```
