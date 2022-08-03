# The Atlas host

The Atlas host is used to

* populate the Atlas database
* generate data files used by the Census Map front end application
* upload and publish data files through the Upload Service

A specific host is used for this work because it is too heavyweight for our laptops,
and because it allows us to do all processing within the AWS environment.
No data needs to downloaded to our laptops.

The Atlas host itself mainly runs Docker, and most of the processing is done
from within containers.

The container image includes all the tools needed, and users shell in to do their work.
Persistent data is held on an EFS mount.

Although all users shell into the Atlas host itself as `ubuntu`, there is a notion of separate
users inside containers.
Unix accounts are set up within the container, and homes live on EFS.
To use this mechanism, `sshd` listens on a host-local docker network.
There are no passwords on user accounts; only ssh keys are allowed.

The user mechanism is designed as a convenience to keep user work separate;
it's not really secure since `ubuntu` is effectively superuser.

## Developer Prerequisites

To use the Atlas host, you must have a few things set up locally:

* AWS account in Sandbox -- talk to a tech lead
* aws cli and session manager installed and working [Supporting AWSB](https://docs.google.com/document/d/1N8k1HnI7R1f9KgFPLAu37bGDLGPPWF9Gk-eiAQlsxm4)
* dp cli installed and working [dp-cli](https://github.com/ONSdigital/dp-cli)


## atlas.env

Certain environment variables must be set when working with any of the atlas utilities:

	AWS_PROFILE
	ATLAS_USER
	COMPOSE_PROJECT_NAME

`AWS_PROFILE` is `dp-sandbox` currently.
None of this has been tested in any other environment, but we expect it to be used in prod eventually.

`ATLAS_USER` is your username in the Atlas container.
These are hardcoded in the Dockerfile used to create the image.
See the Dockerfile for yours.

`COMPOSE_PROJECT_NAME` is used to distinguish your Docker objects from other users'.
The convention is `atlas-$ATLAS_USER`.

You can copy `atlas.env.example` to `atlas.env` and modify to suit.

Do this before working with any atlas utilities:

	. ./atlas.env

You can clear the environment variables with:

	. ./atlas.env -d


## ssh.cfg

ssh is used to get shells on the Atlas host and its containers, and to create tunnels to Docker running on the Atlas host.
Once you have set the required environment variables, you can Generate a custom `ssh.cfg` file like this:

	make ssh.cfg

There are `Makefile` targets for basic ssh operations, but for anything else you can call `ssh`, `scp`, or `sftp` with `-F ssh.cfg`.

## Logging in to the Atlas host

To log in to the Atlas host itself, do this:

	make ssh-atlas

This will give you a shell as `ubuntu`.

It is equivalent to

	dp ssh sandbox atlas 1

except that `dp` won't work until the ansible inventory in `dp-setup` is updated with an `[atlas]` host group.


## Create docker tunnel

When you need to access docker on the Atlas host, set up a tunnel like this:

	make tunnel

Then you can point the docker cli to `tcp://localhost:2375`.


## Setting up Remote Docker Context

As an easy way to get the docker cli to talk to the tunnel, you can create a docker context.
Do something like this locally:

	docker context create --description 'docker on atlas host' --docker tcp://localhost:2375 atlas

And when you want to access docker on the Atlas host, do this first:

	docker context use atlas


## Building the Atlas Image

The Atlas image is a "fat" image based on a ubuntu base with all of the necessary utilities installed.
To avoid stomping on each others' work, each developer builds and runs their own images.

The convention for image names is `atlas-$ATLAS_USER`, where `ATLAS_USER is set in `atlas.env`.

If the ssh tunnel is up, and you are using the `atlas` context, you can do this:

	make image


## Starting your Atlas Container

You can start a detached container running your image:

	make up

This will automatically update `ssh.cfg` to allow you to ssh to your container.


## Logging in to your Container

You can ssh to your running container like this:

	make ssh-container


## Stopping your Atlas Container

Stop your container like this:

	make down


## scp and sftp

You can transfer files to and from your container with `scp` and `sftp`.
Just add `-F ssh.cfg` to the command line:

	scp -F ssh.cfg example.txt container:

or

	sftp -F ssh.cfg container


## Provisioning Atlas users

If you need to add another shell user to the Atlas image:

1. Edit `Dockerfile` and add a new group and user in the "Create user accounts" section.
   Note the uid and gid.
2. Become superuser on the Atlas host itself and create a home directory for the new user under `/data`.
   Change user and group of the home directory to the new user's uid and gid.
   Set the home directory to mode `0700`.
3. Create a `.ssh` directory for the user, setting user, group and mode like the home directory.
4. Create an `authorized_keys` file for the new user with the user's ssh public key.
   Set user, group and mode to uid, gid, `0644`.
5. Build a new image with `make image`.

It's up to the new user to log in and set up their own environment.


## Credits

The `aws-ssm-ec2-proxy-command.sh` script was stolen from https://github.com/qoomon/aws-ssm-ec2-proxy-command.
