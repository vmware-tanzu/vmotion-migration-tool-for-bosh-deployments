# Tanzu Application Service Migration Instructions
This documents the manual and semi-automated steps to vMotion or migrate a running TAS foundation to a new vCenter
environment. This is primarily useful to move running workloads without downtime to new hardware.

## Prerequisites
### Back-up the BOSH Director with BBR
Do not skip this step! If you accidentally lose the BOSH director disk you will need to ensure you have a recent backup to avoid
a complete foundation re-install. See [BBR](https://docs.pivotal.io/ops-manager/2-10/install/backup-restore/backup-pcf-bbr.html#bbr-backup-director)

### Turn off the BOSH Resurrector
We don't want the bosh director to attempt any VM creates while we're in a half migrated state, so disable the resurrector.
```shell
bosh update-resurrection off
```

### Disable Platform Automation Pipelines
We don't want the bosh director to attempt deployment changes while we're migrating, so disable your platform automation pipelines.

### Ensure all VMs are Healthy
Before attempting a migration it's important to verify that all BOSH managed VMs are in a consistent good state. Check
the running foundation by running `bosh vms` and ensuring all agents report as healthy.

## Migrate All VMs
### Configuration - migrate.yml
Before you can execute a migrate command, you will need to create a configuration file for the `vmotion4bosh migrate`
command to use. Use the following template and change it to match your source and target vCenter environments.

```yaml
---
worker_pool_size: 3

additional_vms:
  - vm-2b8bc4a2-90c8-4715-9bc7-ddf64560fdd5 #BOSH director
  - ops-manager-2.10.27

datastores:
  irvine-ds1: ssd_ds1
  irvine-ds2: ssd_ds2

networks:
  PAS-Deployment-01: TAS-Deployment
  PAS-Services-01: TAS-Services
  
clusters:
  PAS-Cluster-AZ1: TAS-AZ1
  PAS-Cluster-AZ2: TAS-AZ2
  PAS-Cluster-AZ3: TAS-AZ3

resource_pools:
  pas-az1: tas-az1
  pas-az2: tas-az2
  pas-az3: tas-az3

bosh:
  host: 10.212.41.141
  client_id: ops_manager

source:
  vcenter:
    host: sc3-vc-01.example.com
    username: administrator@vsphere.local
    insecure: true
  datacenter: Irvine

target:
  vcenter:
    host: sc3-vc-02.example.com
    username: administrator@vsphere.local
    insecure: true
  datacenter: Irvine
```

The datastores, networks, clusters, and resource_pools sections map the old vCenter objects on the left (the yaml key) to the new
vCenter objects on the right. The migration requires the same number of resource pools and networks. The target networks must be in the
same broadcast domain so the VMs can use the same IP addresses.

The `worker_pool_size` controls how many VMs are migrated in parallel. This optional config setting defaults to 3 but
must be 1 or higher. Depending on your hardware and network higher values may decrease the total migration time. It's
generally best to keep this value at 6 or lower to avoid overwhelming the infrastructure.

The `additional_vms` section is optional, however it's recommended that you use it to migrate your BOSH director
and your Operations Manager VM (if using TAS). Just add each VM's name to the list to have vmotion4bosh migrate the
listed VMs.

The required `datastores` section maps the source datastores to the destination 
datastores. Each yaml key on the left is the name of the source datastore 
and the value on the right is the destination datastore name. All datastores 
used by any migrated VM must be present. If migrating to the same storage on 
the destination you will still need to include the datastore mapping, for 
example `ds1: ds1`

The required `networks` section maps the source networks to the destination networks.
Each yaml key on the left is the name of the source network and the value
on the right is the destination network name. All networks used by any 
migrated VM must be present. If migrating to the same network on the
destination you will still need to include the network mapping, for example
`net1: net1`. If migrating TKGI you will need to include a mapping for
each `pks-<GUID>` NCP auto-generated cluster network segment.

The required `clusters` section maps the source cluster to the destination cluster.
Each yaml key on the left is the name of the source cluster and the value
on the right is the destination cluster name. All clusters used by any
migrated VM must be present.

The optional `resource_pools` section maps the source resource pools to the destination
resource pools. Each yaml key on the left is the name of the source resource
pool and the value on the right is the destination resource pool name. All 
resource pools used by any migrated VM must be present. The tool currently
assumes the source and destination both use resource pool or don't.

the optional `bosh` section is used to login to bosh to get a list of all
BOSH managed VMs to migrate. This will migrate all BOSH managed
VMs and doesn't yet allow you to choose VMs by deployment or other criteria.
If this section is left out then the tool will only migrate the VMs listed
in the `additional_vms` section.

The `vcenter` section under `target` is only required if you're migrating the VMs to a vCenter that isn't the same as
the source vCenter.

### Execute Migrate Command
The `migrate` command currently requires network access to BOSH either directly via a routable network or via a local SOCKS proxy.
Once started the process can be stopped via CTRL-C and restarted later, however that will leave your foundation
in a partially migrated state with BOSH inoperable. Either restart the migration or start the migration process in the
reverse direction.

Before running any commands set a few required environment variables for sensitive secrets/passwords. If you're
migrating only to a new cluster within the same vCenter then the TARGET_PASSWORD can be skipped
```shell
export BOSH_CLIENT_SECRET='secret'
export SOURCE_PASSWORD='pwd'
export TARGET_PASSWORD='pwd'
```

Use the `vmotion4bosh` `migrate` command to move all BOSH managed VMs to another vCenter instance and/or cluster:
```shell
vmotion4bosh migrate --debug true 2>debug.log
```

> **NOTE** - It's _highly_ recommended to use `--dry-run true` flag first to ensure there aren't any obvious
problems trying to migrate any of the VMs, like a missing network mapping etc.

## Update Operations Manager & BOSH Configuration
### Backup and Upgrade Operations Manager (*only* required for versions < 2.10.17)
This step is optional and only required if your Operations Manager version is less than 2.10.17. If you have an older
version and skip this step you'll run into problems when you apply-changes. [Upgrade Operations Manager](https://docs.pivotal.io/ops-manager/2-10/install/upgrading-pcf.html):

* Export installation
* Upgrade Operations Manager to the latest version.
* Import installation

### Modify the Operations Manager installation
Once all VMs have been migrated including the BOSH director, you'll need to perform the following steps on the Operations Manager VM.
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
```
```shell
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
- Target network names
- Target vCenter address, username, and password (if moving to a new vCenter)

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

If you're migrating a TKGI deployment you will also need to update the vSphere details for TKGI as well. The TKGI
properties section looks like:
```yaml
- properties:
  - deployed: true
    identifier: vcenter_master_creds
    value:
      identity: administrator@vsphere.local
      password: hunter2
  - deployed: true
    identifier: vcenter_ip
    value: vcenter.example.com
  - deployed: true
    identifier: vcenter_dc
    value: Datacenter
  - deployed: true
    identifier: vcenter_ds
    value: NFS-Datastore1
  - deployed: true
    identifier: vcenter_vms
    value: tkgi_vms
```

Re-encrypt installation.yml and actual-installation.yml
```shell
sudo -u tempest-web \
   SECRET_KEY_BASE="s" \
   RAILS_ENV=production \
   /home/tempest-web/tempest/web/scripts/encrypt \
   /tmp/installation.yml \
   /var/tempest/workspaces/default/installation.yml
```
```shell
sudo -u tempest-web \
   SECRET_KEY_BASE="s" \
   RAILS_ENV=production \
   /home/tempest-web/tempest/web/scripts/encrypt \
   /tmp/actual-installation.yml \
   /var/tempest/workspaces/default/actual-installation.yml
```

### Validate Operations Manager Changes
From your workstation or jumpbox validate the edits by running `om staged-director-config` to verify changes have been
made; ensure the cluster and datastore values are updated etc.
```shell
om staged-director-config --no-redact
```

If everything looks correct, validate that the vSphere service account can talk to the IaaS and has the
correct permissions required. This runs the same validations as at the beginning of an apply-changes.
```shell
om pre-deploy-check
```

### Move & Rename the BOSH Persistent Disk

> **NOTE** - This step only needs to be completed if you migrated storage. If you only migrated compute then skip to the
> next step and deploy the updated BOSH director.

With the BOSH VM migrated to the new vCenter instance, SSH to Operations Manager and view the
`/var/tempest/workspaces/default/deployments/bosh-state.json` file; make note of the disk CID (copy the GUID somewhere),
it will look something like
`disk-43edf7b2-467b-4913-8142-91b24896b482.eyJ0YXJnZXRfZGF0YXN0b3JlX3BhdHRlcm4iOiJeKE5GU1xcLURhdGFzdG9yZTEpJCJ9`.
Record the disk GUID, i.e. `disk-43edf7b2-467b-4913-8142-91b24896b482`

Login to the target vSphere UI and find the migrated BOSH director VM. Shut the Director VM. Edit the BOSH VM settings
in vCenter, find the persistent disk (typically the 3rd disk in the list) and detach it
(click the X that appears to the right of the disk on hover).

Navigate to the datastore browser, find the bosh director VM (same <vm cid> as above) and copy the persistent disk to
the `pcf_disk` folder (or whatever folder name that is specified in the bosh director tile) in the datastore
(should be named something like vm-GUID_3.vmdk*). You may need to rename it to the disk CID that bosh is expecting
(i.e. vm-GUID_3.vmdk becomes disk-GUID.vmdk).


From the Operations Manager VM, edit the bosh-state.json
```shell
sudo vim /var/tempest/workspaces/default/deployments/bosh-state.json
```

Edit the datastore section to point to new datastore. Remove the base64 suffix on disk name (part between period and .vmdk) if it exists i.e.:
`disk-1983a793-2c33-474d-ad7f-8e24586ccc13.eyJ0YXJnZXRfZGF0YXN0b3JlX3BhdHRlcm4iOiJeKE5GU1xcLURhdGFzdG9yZTIpJCJ9` would be changed to `disk-1983a793-2c33-474d-ad7f-8e24586ccc13`


### Deploy the Updated BOSH Director
Apply changes to director tile _only_
```shell
om apply-changes --skip-deploy-products
```

This will recreate the bosh director and ensure the CPI is working on the new vSphere cluster.

## Finishing Up
After all VMs have been migrated you should make sure all BOSH managed VMs report they're in a running state.

```shell
bosh vms
```

If some are not, attempt to restart the VMs. Sometimes the BOSH agent's BBS gets into a bad state.

```shell
bosh -d <deployment> restart <instance>
```

> **NOTE** - Never attempt to run bosh cck or recreate until you've successfully applied changes. Attempting to run
> those commands will fail and/or incorrectly create replacement VMs on the old vSphere cluster.

The last step is an apply-changes of every deployment and VM. If there are on-demand service instances you will need to
enable the on demand tile's upgrade-all-service-instances or recreate-all-instances errand. Finally turn the resurrector back on.
```shell
om apply-changes && bosh update-resurrection on
```

