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
Before you can execute a migrate command, you will need to create a configuration file for the vmotion4bosh migrate
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
  datacenter: Irvine

target:
  vcenter:
    host: sc3-vc-02.example.com
    username: administrator@vsphere.local
    insecure: true
  datacenter: Irvine
  cluster: Cluster01
```

The datastores, networks, and resource_pools sections map the old vCenter objects on the left (the yaml key) to the new
vCenter objects on the right. The migration requires the same number of resource pools and networks. The target networks must be in the
same broadcast domain so the VMs can use the same IP addresses.

The `worker_pool_size` controls how many VMs are migrated in parallel. This optional config setting defaults to 3 but
must be 1 or higher. Depending on your hardware and network higher values may decrease the total migration time. It's
generally best to keep this value at 6 or lower to avoid overwhelming the infrastructure.

The `additional_vms` section is optional, however it's recommended that you use it to migrate your BOSH director
and your Operations Manager VM (if using TAS). Just add each VM's name to the list to have vmotion4bosh migrate the
listed VMs.

The `vcenter` section under `target` is only required if you're migrating the VMs to a vCenter that isn't the same as
the source vCenter.

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

The last step is an apply-changes of every deployment and VM. If there are on-demand service instances with an upgrade
errand you'll want to run those too, finally turn the resurrector back on.
```shell
om apply-changes && bosh update-resurrection on
```

