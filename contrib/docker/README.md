Exchangecoin Full Node for Docker
===========================

[![ISC License](https://img.shields.io/badge/license-ISC-blue.svg)](http://copyfree.org)

## Overview

This provides all of the necessary files to build your own lightweight non-root
container image based on `scratch` that provides `exccd`, `exccctl`,
`promptsecret` and `gencerts`.

The approach used by the primary `Dockerfile` is to employ a multi-stage build
that downloads and builds the latest source code, compresses the resulting
binaries, and then produces a final image based on `scratch` that only includes
the Exchangecoin-specific binaries.

### Container Image Security Properties

The provided `Dockerfile` places a strong focus on security as follows:

- Runs as a non-root user
- Uses a static UID:GID of 10000:10000
  - Note that using UIDs/GIDs below 10000 for container users is a security
    risk on several systems since a hypothetical attack which allows escalation
    outside of the container might otherwise coincide with an existing user's
    UID or existing group's GID which has additional permissions
- The image is based on `scratch` image (aka completely empty) and only includes
  the Exchangecoin-specific binaries which means there is no shell or any other
  binaries available if an attacker were to somehow manage to find a remote
  execution vulnerability exploit in a Exchangecoin binary

### Container Environment Variables

- `EXCC_DATA` (Default: `/home/excc`):  
  The directory where data is stored inside the container.  This typically does
  not need to be changed.

- `EXCCD_NO_FILE_LOGGING` (Default: `true`):  
  Controls whether or not dcrd additionally logs to files under `EXCC_DATA`.
  Logging is only done via stdout by default in the container since that is
  standard practice for containers.

- `EXCCD_ALT_DNSNAMES`: (Default: None)
  Adds alternate server DNS names to the server certificate that is automatically
  generated for the RPC server.  This is important when attempting to access the
  RPC from external sources because TLS is required and clients verify the
  server name matches the certificate.

## Usage

### Quick Start

The following are typical commands to get up and going quickly.  The remaining
sections describe things more in depth.

**Note:** These series of commands have you define and use environment variables
in order to help make it clear exactly what every command line argument refers
to.  However, this means that if you close the shell, the commands will no
longer work as written because those environment variables will no longer exist.
You may wish to replace all instances of `"${...}"` with the associated concrete
value.

1. Build the base image with a tag to make it easy to reference later.  These
   commands all use `yourusername/exccd` for the image tag, but you should
   replace `yourusername` with your username or something else unique to you so
   you can easily identify it as being one of your images:

   ```sh
   $ EXCCD_IMAGE_NAME="yourusername/exccd"
   $ docker build -t "${EXCCD_IMAGE_NAME}" -f contrib/docker/Dockerfile .
   ```

   **NOTE: This MUST be run from the main directory of the exccd code repo.**

2. Create a data volume and change its ownership to the user id of the user
   inside of the container so it has the necessary permissions to write to it:

   ```sh
   $ docker volume create excc-data
   $ EXCC_DATA_VOLUME=$(docker volume inspect excc-data -f '{{.Mountpoint}}')
   $ sudo chown -R 10000:10000 "${EXCC_DATA_VOLUME}"
   ```

   **NOTE: The data volume only needs to be created once.**

3. Run `exccd` on `testnet` in the background using the aforementioned data
   volume to store the blockchain and configuration data along with a name to
   make it easy to reference later:

   ```sh
   $ EXCCD_CONTAINER_NAME="exccd-testnet"
   $ docker run -d --read-only \
     --name "${EXCCD_CONTAINER_NAME}" \
     -v decred-data:/home/excc \
     "${EXCCD}" --testnet --altdnsnames "${EXCCD_CONTAINER_NAME}"
   ```

4. View the output logs of `exccd` with the docker logs command:

   ```sh
   $ docker logs "${EXCCD_CONTAINER_NAME}"
   ```

### Starting and Stopping the Container

   ```sh
   $ docker stop "${EXCCD_CONTAINER_NAME}"
   $ docker start "${EXCCD_CONTAINER_NAME}"
   ```

### Preliminaries

TODO: Explain about non-root permissions, network, etc

### Basics

TODO: Finish documenting all the details here

- [Dockerfile](./Dockerfile)  
  Provides a user-contributed configuration file for building a container image
  that consists of `exccd`, `exccctl`, `gencerts`, and `promptsecret` along with
  exposed ports for exccd's RPC server.  It is based on `scratch` and runs as a
  non-root container.

TODO: It would probably be nice to provide some variants such as:
- `Dockerfile.release` that either grabs the latest release code or checks out the
  latest release tag instead of building the master branch
- `Dockerfile.local` that builds an image using the code in the build context
  instead of cloning and building the latest master branch

### Interacting via RPC

#### TODO: With shared network...

Assuming the environment variables and configuration matches what was outlined
in the quick start section:

```sh
$ docker run --rm --network container:"${EXCCD_CONTAINER_NAME}" --read-only \
  -v excc-data:/home/excc \
  "${EXCCD_IMAGE_NAME}" exccctl --testnet getblockchaininfo
```

#### TODO: With user-defined network...

Assuming the environment variables and configuration matches what was outlined
in the quick start section:

TODO: Would need to remove existing container and start new one with `--network decred` as follows...

**NOTE: The network volume only needs to be created once.**

```sh
$ docker network create excc
$ docker stop "${EXCCD_CONTAINER_NAME}"
$ docker rm "${EXCCD_CONTAINER_NAME}"
$ docker run -d --read-only \
  --network excc \
  --name "${EXCCD_CONTAINER_NAME}" \
  -v decred-data:/home/excc \
  "${EXCCD_IMAGE_NAME}" --testnet --altdnsnames "${EXCCD_CONTAINER_NAME}"
$ docker run --rm --read-only \
  --network excc \
  -v decred-data:/home/excc \
  "${EXCCD_IMAGE_NAME}" exccctl --testnet --rpcserver "${EXCCD_CONTAINER_NAME}" getblockchaininfo
```

#### TODO: Accessing the RPC server from remote services outside of a docker network

TODO: Needs to be running with port mapped...aka:

```sh
$ docker run -d --read-only \
  --name "${EXCCD_CONTAINER_NAME}" \
  -v excc-data:/home/excc \
  -p 127.0.0.1:19109:19109 \
  "${EXCCD_IMAGE_NAME}" --testnet --altdnsnames "${EXCCD_CONTAINER_NAME}"
```

TODO: From other machine such as the host machine (not inside a docker container)...
TODO: sudo required here because the data volume needs the perms of the uid/gid
      inside the container which the local user on the host won't have access too..
      could alternatively use `docker cp` to copy the cert and conf file out of the container...

```sh
$ exccdrpcuser=$(sudo cat "${EXCC_DATA_VOLUME}/.exccd/exccd.conf" | grep rpcuser= | cut -c9-)
$ exccdrpcpass=$(sudo cat "${EXCC_DATA_VOLUME}/.exccd/exccd.conf" | grep rpcpass= | cut -c9-)
$ sudo curl --cacert "${EXCC_DATA_VOLUME}/.exccd/rpc.cert" --user "${exccdrpcuser}:${exccdrpcpass}" \
  --data-binary '{"jsonrpc":"1.0","id":"1","method":"getbestblock","params":[]}' \
  https://127.0.0.1:19109
```

## Troubleshooting / Common Issues

TODO

### TODO: Write permission errors

TODO

### TODO: Remote access certificate errors

TODO

