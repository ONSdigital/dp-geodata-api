# filescli

## Running the cli

filescli is a command line interface to Static File services.
See [man page](filescli.md).

To use the cli, you must compile it, then create a config file and set some environment variables.

Build the cli like this:

    $ make filescli

The config file is a yaml file setting the URLs of the required Static File services.
See [config-example.yml](config-example.yml).

Environment variables hold authentication tokens or credentials that can be used to generate tokens.
See [man page](filescli.md) for specifics.

## Generating docs

The man page is written in portable mandoc [filescli.mandoc](filescli.mandoc).
The markdown and pdf versions are generated with the [mandoc](https://mandoc.bsd.lv) utility.

To generate markdown and pdf versions, you need a local `mandoc` or a docker image with mandoc.
(`mandoc.sh` detects a local version.)
To create the docker image:

    $ make mandoc-image

Then generate the documents:

    $ make filescli.md filescli.pdf
