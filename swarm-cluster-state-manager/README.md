# Docker Swarm Cluster State Manager

This is not an official Google product.

## Overview

The Swarm cluster state manager is a simple prototype that demonstrates how cluster state could be managed for Docker Swarm. The ideas for this prototype are based on the Kubernetes pod replication controller.

## Usage

### Start the swarm cluster state manager

Get the swarm-master address using the `docker-machine env` command:

```
$ docker-machine env swarm-master --swarm
```

Output:

```
export DOCKER_TLS_VERIFY="1"
export DOCKER_HOST="tcp://203.0.113.13:3376"
export DOCKER_CERT_PATH="/Users/kelseyhightower/.docker/machine/machines/swarm-master"
export DOCKER_MACHINE_NAME="swarm-master"
# Run this command to configure your shell: 
# eval "$(docker-machine env swarm-master --swarm)"
```

Edit the `docker-compose.yml` file and replace `SWARM_MASTER_ADDR` with your swarm-master address.

```
swarm-cluster-state-manager:
  image: kelseyhightower/swarm-cluster-state-manager
  volumes:
    - /etc/docker:/etc/docker
  command: |
    --addr=0.0.0.0:3476
    --swarm-manager="tcp://203.0.113.13:3376"
    --tlscacert="/etc/docker/ca.pem"
    --tlscert="/etc/docker/server.pem"
    --tlskey="/etc/docker/server-key.pem"
  external_links:
    - swarm-agent-master
  ports:
    - "3476:3476"
```

Launch swarm-cluster-state-manager on the same Docker host as the swarm-master:

Use the `eval` command to ensure your are pointing to the same Docker host where the swarm-master
is running.

```
$ eval $(docker-machine env swarm-master)
```

Use the `docker-compose up` command to start the `swarm-cluster-state-manager` service:

```
$ docker-compose up -d
```

### Reusing the Docker machine client certs

#### OS X

On OS X you'll need to create a pkcs12 bundle of the PEM encoded client certs in order
to use them with cURL.

```
$ openssl pkcs12 -export \
  -in ~/.docker/machine/certs/cert.pem \
  -inkey ~/.docker/machine/certs/key.pem \
  -out ~/.docker/machine/certs/cert.pfx 
```

At the follow prompt set the export password to protect the cert and key:

```
Enter Export Password:
Verifying - Enter Export Password:
```

Later examples assume the word `swarm` was used for the password.

### Submit a new cluster state object

The following command submits a cluster state object named nginx and will ensure 5
containers are started from the `nginx:1.9.6` Docker image.

```
$ curl https://<swarm-master-ip>:2476/submit \
  -d '{"Name": "nginx", "Image": "nginx:1.9.6", "Count": 5}' \
  --cacert ~/.docker/machine/certs/ca.pem \
  --cert ~/.docker/machine/certs/cert.pfx \
  --pass swarm
```

### Get the current status

The following command retrieves the cluster status from the Swarm cluster state manager.

```
$ curl https://<swarm-master-ip>:2476/status \
  --cacert ~/.docker/machine/certs/ca.pem \
  --cert ~/.docker/machine/certs/cert.pfx \
  --pass swarm
```
