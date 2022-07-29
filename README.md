# vmotion-migration-tool-for-bosh-deployments

![Build workflow](https://github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/actions/workflows/build.yml/badge.svg)

Tool and instructions to seamlessly move a BOSH deployed Cloud Foundry installation to new hardware.

See [docs/migrate.md](docs/migrate.md) for detailed Tanzu Application Service migration instructions.

## Basic Usage

The `vmotion4bosh` binary has the following commands:

- `migrate` which supports migrating all BOSH managed VMs
- `migrate-vm` which migrates a single VM like the BOSH Directory or Operations Manager
- `version` displays the git SHA the binary was built with and optionally a version number

`--help` will display the main help or if after a command the command specific help.

## Building

Git clone this repository and run:
```shell
make build
```

This will cross compile and compress the Mac, Linux, and Windows versions:
```shell
make release
```

To make an official release on GitHub, make a git tag with the new version and push it. The GitHub Release Action
will run on any new pushed tags. Once complete the new release artifacts will show up on the GitHub releases page.
For example to release version v0.3.2 use the following commands:
```shell
git tag v0.3.2
git push origin --tags
```

## Testing
Git clone this repository and run
```shell
make test
```

## Contributing

The vmotion-migration-tool-for-bosh-deployments project team welcomes contributions from the community. Before you start working with vmotion-migration-tool-for-bosh-deployments, please
read our [Developer Certificate of Origin](https://cla.vmware.com/dco). All contributions to this repository must be
signed as described on that page. Your signature certifies that you wrote the patch or have the right to pass it on
as an open-source patch. For more detailed information, refer to [CONTRIBUTING.md](CONTRIBUTING.md).

