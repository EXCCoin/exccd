contrib
=======

## Overview

This consists of extra optional tools which may be useful when working with exccd
and related software.

## Contents

### Example Service Configurations

- [OpenBSD rc.d](services/rc.d/exccd)  
  Provides an example `rc.d` script for configuring exccd as a background service
  on OpenBSD.  It also serves as a good starting point for other operating
  systems that use the rc.d system for service management.

- [Service Management Facility](services/smf/exccd.xml)  
  Provides an example XML file for configuring exccd as a background service on
  illumos.  It also serves as a good starting point for other operating systems
  that use use SMF for service management.

- [systemd](services/systemd/exccd.service)  
  Provides an example service file for configuring exccd as a background service
  on operating systems that use systemd for service management.

### Simulation Network (--simnet) Preconfigured Environment Setup Script

The [excc_tmux_simnet_setup.sh](./excc_tmux_simnet_setup.sh) script provides a
preconfigured `simnet` environment which facilitates testing with a private test
network where the developer has full control since the difficulty levels are low
enough to generate blocks on demand and the developer owns all of the tickets
and votes on the private network.

The environment will be housed in the `$HOME/exccdsimnetnodes` directory by
default.  This can be overridden with the `EXCC_SIMNET_ROOT` environment variable
if desired.

See the full [Simulation Network Reference](../docs/simnet_environment.mediawiki)
for more details.

### Building and Running OCI Containers (aka Docker/Podman)

The project does not officially provide container images.  However, all of the
necessary files to build your own lightweight non-root container image based on
`scratch` from the latest source code are available in the docker directory.
See [docker/README.md](./docker/README.md) for more details.
