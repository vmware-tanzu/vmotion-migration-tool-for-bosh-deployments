# FAQ

## Can I cold migrate VMs?
Yes, however if you're migrating BOSH managed VMS you will need to leave the BOSH director running while you execute
the migration so vmotion4bosh can query the list of VMs. Once that is complete you can vMotion the BOSH director
separately.

If you need to power off all BOSH managed VMs quickly, you can use the following command:

```shell
govc vm.power -off $(bosh vms --column=vm_cid --json | jq -r '.Tables[].Rows[].vm_cid')
```

## Can I migrate from a newer version of vSphere to an older version?
Yes, but it requires that each VM have set an individual EVC mode that is compatible with the target cluster. To do
this requires that you have vSphere 6.7 or higher and the VM is machine version 14 or higher. By default the
machine version for all BOSH deployed VMs is version 9. To have BOSH automatically upgrade the VM machine version
anytime it deploys a VM, set the [upgrade_hw_version](https://bosh.io/docs/vsphere-cpi/#resource-pools) property to
true.

## Can I migrate TKGI clusters one at a time?
Not today, all migrations using vmotion4bosh assume an all or nothing approach.

## Can I migrate one AZ at a time?
Not today, all migrations using vmotion4bosh assume an all or nothing approach.

## Can I migrate from one AZ to three?
Not really, vMotion4bosh will migrate the existing foundation and requires that the source and target mappings
have the same number of AZs. Once migrated though, you could add additional AZs.

## Can I migrate from three AZs to one?
vMotion4bosh requires that the source and target mappings have the same number of AZs. You cannot remove an AZ
from the BOSH configuration.

## Do I have to use environment variables for passwords and secrets?
No, but it's highly recommended. You can set environment variables without echoing to the terminal (i.e. history):
```shell
read -s SECRET
```
This will read the secret into the SECRET environment variable without the actual contents of the secret showing up
in any files in the file system.

## Can I use environment variables for other config values besides passwords?
Yes. Just use the `${VAR_NAME}` syntax in the migrate.yml.
