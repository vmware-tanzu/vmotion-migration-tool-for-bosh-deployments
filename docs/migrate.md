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

datastores:
  irvine-ds1: ssd_ds1
  irvine-ds2: ssd_ds2

networks:
  PAS-Deployment-01: TAS-Deployment
  PAS-Services-01: TAS-Services

bosh:
  host: 10.212.41.141
  client_id: ops_manager
  client_secret: ${BOSH_CLIENT_SECRET}

vcenters:
  - vcenter: &vcenter1
      host: vc01.example.com
      username: administrator@vsphere.local
      password: ${VCENTER1_PASSWORD}
      insecure: true
      datacenter: Datacenter1
  - vcenter: &vcenter2
      host: vc02.example.com
      username: administrator2@vsphere.local
      password: ${VCENTER2_PASSWORD}
      insecure: true
      datacenter: Datacenter2

compute:
  source:
    - name: az1
      vcenter: *vcenter1
      clusters:
        - name: cf1
          resource_pool: pas-az1
    - name: az2
      vcenter: *vcenter1
      clusters:
        - name: cf2
          resource_pool: pas-az2
    - name: az3
      vcenter: *vcenter1
      clusters:
        - name: cf3
          resource_pool: pas-az3
  target:
    - name: az1
      vcenter: *vcenter2
      clusters:
        - name: tanzu-1
          resource_pool: tas-az1
    - name: az2
      vcenter: *vcenter2
      clusters:
        - name: tanzu-2
          resource_pool: tas-az2
    - name: az3
      vcenter: *vcenter2
      clusters:
        - name: tanzu-3
          resource_pool: tas-az3

additional_vms:
  az1:
    - vm-2b8bc4a2-90c8-4715-9bc7-ddf64560fdd5
    - ops-manager-2.10.27
```

The datastores and networks sections map the old vCenter objects on the left (the yaml key) to the new
vCenter objects on the right. The migration requires the same number of resource pools and networks. The target 
networks must be in the same broadcast domain so the VMs can use the same IP addresses.

#### worker_pool_size
The optional `worker_pool_size` section controls how many VMs are migrated in parallel. This setting defaults to 3 but
must be 1 or higher. Depending on your hardware and network higher values may decrease the total migration time. It's
generally best to keep this value at 6 or lower to avoid overwhelming the infrastructure.

#### additional_vms
The optional `additional_vms` section is used to explicitly migrate any VM in vCenter that BOSH doesn't know about. it's
recommended that you use it to migrate your BOSH director and your Operations Manager VM (if using TAS/TKGI).
Just add each VM's name to the list to have vmotion4bosh migrate the listed VMs. The AZ key(s) you use should match
one of the AZs listed in compute section so vmotion4bosh knows where which vcenter connections to use.

Each VM specified can be a fully qualified inventory path like
`/vc01/vm/tas_vms/96212534-4543-41d4-892a-b534e9469ba3/vm-03b5517d-8fcc-4640-8b88-0b6f5b6d2adb`
or just the VM name `vm-03b5517d-8fcc-4640-8b88-0b6f5b6d2adb`. The short VM name is usually fine unless there are
2 VMs with the same name in the same datacenter.

Stemcells (i.e. VM templates) can also be added to the `additiona_vms` section. This can be useful to copy over the BOSH
director stemcell as vmotion4bosh doesn't currently do this automatically. You can find the required stemcell in the
Operations Manager bosh-state.json file.

#### datastores
The required `datastores` section maps the source datastores to the destination datastores. Each yaml key on the left
is the name of the source datastore and the value on the right is the destination datastore name. All datastores 
used by any migrated VM must be present. If migrating to the same storage on the destination you will still need to
include the datastore mapping, for example `ds1: ds1`.

#### networks
The required `networks` section maps the source networks to the destination networks. Each yaml key on the left is the
name of the source network and the value on the right is the destination network name. All networks used by any 
migrated VM must be present. If migrating to the same network on the destination you will still need to include the
network mapping, for example `net1: net1`.  Network name comes from network in vCenter(not from bosh config).

If migrating TKGI you will need to include a mapping for each `pks-<GUID>` NCP auto-generated cluster network segment.

#### compute
The required `compute` section maps the source AZ/cluster/resource pool to the destination AZ/cluster/resource pool.
Generally the structure of the compute section should follow the same structure as the BOSH director CPI configuration.
The AZ name is arbitrary but each source AZ _must_ have a corresponding target AZ with the _same_ name. While you can
inline the vCenter connection details it is better to use a yaml reference from the `vcenters` section. Each AZ 
supports multiple clusters/resource pools - resource pools are optional. All AZs/clusters used by any migrated VM must
be present in the compute section.

While you must have the _same_ number of source and target AZs, it is possible to map the vSphere primitives
differently. vCenters, Clusters, and Resource Pools can be reused between AZs. For example map 3 AZs with different
vCenters/clusters/resource pools to 1 vCenter with 3 clusters:
```yaml
compute:
  source:
    - name: az1
      vcenter: *vcenter1
      clusters:
        - name: vc01cl01
          resource_pool: pas-az1
    - name: az2
      vcenter: *vcenter2
      clusters:
        - name: vc01cl01
          resource_pool: pas-az2
    - name: az3
      vcenter: *vcenter3
      clusters:
        - name: vc01cl01
          resource_pool: pas-az3
  target:
    - name: az1
      vcenter: *new_vcenter
      clusters:
        - name: tanzu1
    - name: az2
      vcenter: *new_vcenter
      clusters:
        - name: tanzu2
    - name: az3
      vcenter: *new_vcenter
      clusters:
        - name: tanzu3
```

#### bosh
the optional `bosh` section is used to login to bosh to get a list of all BOSH managed VMs to migrate. This will
migrate all BOSH managed VMs and doesn't yet allow you to choose VMs by deployment or other criteria (at least yet).
If this section is left out then the tool will only migrate the VMs listed in the `additional_vms` section. It's 
recommended to use an environment variable in the format of `${BOSH_CLIENT_SECRET}` for the bosh client secret that
will be expanded during runtime.

The optional `vcenters` section can be used to declare vCenter connections that can be reused via a yaml reference for
each AZ section under `compute`. At a minimum you should have once vcenter list item, with as many entries as required.
It's recommended to use an environment variable in the format of `${VCENTER_PASSWORD}` for the vcenter password that
will be expanded during runtime. In the above example config it is expected there is an environment variable with the
vcenter password named `VCENTER1_PASSWORD`.

If you have the same vcenter but a different datacenter needed for another AZ you'll need to redeclare another
vcenter entry as there is a 1:1 relationship between vcenter entry and datacenter.

### Execute Migrate Command
The `migrate` command currently requires network access to BOSH either directly via a routable network or via a local SOCKS proxy.
Once started the process can be stopped via CTRL-C and restarted later, however that will leave your foundation
in a partially migrated state with BOSH inoperable. Either restart the migration or start the migration process in the
reverse direction via the `revert` command.

Use the `vmotion4bosh migrate` command to move all BOSH managed VMs to another vCenter instance and/or cluster:
```shell
vmotion4bosh migrate --debug 2>debug.log
```

> **NOTE** - It's _highly_ recommended to use `--dry-run` flag first to ensure there aren't any obvious
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

If you didn't copy the BOSH director stemcell over in the `additional_vms` section, then you will need to delete the
stemcell from the stemcells section of the bosh-state.json file, otherwise you will receive an error during apply
changes about the stemcell not being found.

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

If the deployment is outside of Operations Manager or individual service instances or if you see following CPI error(`No valid placement found for VM compute and storage requirement`) in the followup bosh command, it means that the deployment still has old CPI information. run following commands to apply the changes.
```shell
bosh -d DEPLOYMENT_NAME manifest > ./DEPLOYMENT_NAME_manifest.yml
```
```shell
bosh -d DEPLOYMENT_NAME deploy ./DEPLOYMENT_NAME_manifest.yml
```
It fetches deployment manifest and then applies the new CPI to the deployment.

