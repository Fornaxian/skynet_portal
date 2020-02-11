# Fornaxian Portal

This is the portal software for Skynet by Fornaxian Technologies. You can see
how it works [here](https://sky.pixeldrain.com).

## Compiling

You can compile the portal by running this command with Go installed:

```
go build -o portal main.go proxy.go
```

## Running

The portal uses a few commandline flags for configuration:

 * `--listen` the address this server will listen on. Defaults to `:8082`
 * `--res` the directory where the resouces are stored. This should point at the
   `res` directory in this repository. Defaults to `res`
 * `--siad-url` the URL of the `siad` API to use. Defaults to `http://127.0.0.1:9980`

The API password for the `siad` API will be read from `~/.sia/apipassword`. This
is the location where `siad` installs the password by default.

## Configuring siad

To make the portal software work `siad` needs to be configured as a portal node.
To do this you need to set an allowance. More info
[here](https://support.sia.tech/article/thvymhf1ff-about-renting)
