# Description
It's a network dispatcher which allows to execute scrips on wifi connect/disconnect events.

It mitigates big flow in [NetworkManager](https://networkmanager.dev)'s dispatcher which ignores lost wifi events. \
That's it, [NetworkManager](https://networkmanager.dev) only reacts on proper connect/disconnect events when WIFI was disconnect button pressed.\
But with laptops most common case is when laptop is often moved across different WIFI access points and connection lost

This project solves this problem by properly reacting on all WIFI connect/disconnect events.

Motivation to write it was my NAS network each time made dolphin hang since share was still mounted when wifi was lost.
Network dispatcher provides mount and umount script and handles case when the same network share needs to be mounted at home \
and remotely via ssh tunnel

# Features
This project main use case was to mount network share directly when on home network \
and mount the same share through ssh tunnel when everywhere else. \

* Listens to WIFI connect/disconnect events
* Provides convenient way to automatically mount network share when at home or remote location
* Allows to specify configurable scripts per event. See the [Usage section](#usage) for more details
* Provides filters by mac address to run scripts in the specific locations.
* Includes tested scripts to properly mount and umount network shares on connects/disconnects. See the [Usage section](#usage) for more details


# Requirements
* [NetworkManager](https://networkmanager.dev). My network dispatcher listens[NetworkManager](https://networkmanager.dev) low level dbus events
* [Dbus](https://www.freedesktop.org/wiki/Software/dbus/). See above
* [Systemd](https://systemd.io/). My network dispatcher provides systemd service to run itself
* Ssh tunnel script relies on console autossh and ssh to maintain tunnel

# Install
* Install by executing
```
bash -c "$(curl -L https://raw.githubusercontent.com/danilovsergei/network-dispatcher/main/install.sh)" -- "$HOME/bin"

```
It will download and unpack latest [release](https://github.com/danilovsergei/network-dispatcher/releases/latest/download/network-dispatcher.zip) into `"$HOME/bin"` directory and install network-dispatcher systemd service into `"$HOME/.config/systemd/user/network-dispatcher.service"` \
Everything runs under current user.

After that network disatcher is ready to react on events. However it's necessary to define config with scripts to react. See the [Usage section](#usage) for examples

# Usage
To react on events create network-dispatcher config in `$HOME/.config/network-dispatcher/config.json`

Here is example of config which mounts and umounts CIFS network share.
This config is using scripts provided in [release](https://github.com/danilovsergei/network-dispatcher/releases/latest/download/network-dispatcher.zip) and installed into `$HOME/network-dispatcher` 

## Script mounts/umount local CIFS share
It's a basic example how to mount and unmount share.\
Note that no filters specified which means given scripts will be triggered in ANY wifi network , at home or outside.\
It's not optimal since mounted local share will hang the system when outside.

See [Script mounts/umount using filter](#cifs-mount-filter)

```
{
  "Entities": [
    {
      "Script": "$HOME/network-dispatcher/share_mount.sh",
      "Event": "connected",
      "EnvVariables": {
        "MOUNT_POINT": "//192.168.1.1/Storage"
      }
    },
    {
      "Script": "$HOME/network-dispatcher/share_umount.sh",
      "Event": "disconnected",
      "EnvVariables": {
        "MOUNT_POINT": "//192.168.1.1/Storage"
      }
    }
  ]
}
```
`share_mount.sh` and `share_umount.sh` relies on `/etc/fstab` record to be able to work as user.  So please add mount line to your `/etc/fstab`

```
//192.168.1.1/Storage  /home/Storage  cifs   noauto,rw,users,nodev,relatime 0 0
```
`noauto` parameter is crucial because it instruct systemd that share will be mounted manually by our scripts

<!----><a name="cifs-mount-filter"></a>
## Script mounts/umount local CIFS share using filter
Given option provides ability to run the scripts only when connected to certain gateway.\
Which is typically home router at home. It allows to specify gateway mac address in the `Included_MacAddresses`

```
{
  "Entities": [
    {
      "Script": "$HOME/network-dispatcher/share_mount.sh",
      "Event": "connected",
      "EnvVariables": {
        "MOUNT_POINT": "//192.168.1.1/Storage"
      },
      "Included_MacAddresses": [
        "cc:ce:cc:ce:ce:cc"
      ]
    },
    {
      "Script": "$HOME/network-dispatcher/share_umount.sh",
      "Event": "disconnected",
      "EnvVariables": {
        "MOUNT_POINT": "//192.168.1.1/Storage"
      },
      "Included_MacAddresses": [
        "cc:ce:cc:ce:ce:c"
      ]
    }
  ]
}
```

There is opposite option called `Excluded_MacAddresses` which skips script execution on specified 

## Script mounts/umount local and remote CIFS share based on location
This script is a most common scenario:
* when at home it mounts share as `//192.168.1.1/Storage` directly via cifs using home gateway mac address in `Included_MacAddresses`.\
* when outside share is mounted as `//127.0.0.1/Storage` via ssh tunnel using `Excluded_MacAddresses` to exclude home network
* mount script generates a symlink `$HOME/Storage` for both mount at home and outside cases. So share is always accessible by the same path.
```
{
  "Entities": [
   {
      "Script": "$HOME/network-dispatcher/cifs_ssh_tunnel.sh",
      "Event": "connected",
      "EnvVariables": {
        "SSH_PORT": "2222",
        "SSH_USER": "homeuser",
        "SSH_HOST": "my-external-address.dyndns.com",
        "PRIVATE_KEY": "$HOME/.ssh/cifs_id_rsa"
      },
      "Excluded_MacAddresses": [
        "cc:ce:cc:ce:ce:cc"
      ]
    },
    {
      "Script": "$HOME/network-dispatcher/share_mount.sh",
      "Event": "connected",
      "EnvVariables": {
        "MOUNT_POINT": "//127.0.0.1/Storage",
        "MOUNT_LINK": "$HOME/Storage"
      },
      "Excluded_MacAddresses": [
        "cc:ce:cc:ce:ce:cc"
      ]
    },
    {
      "Script": "$HOME/network-dispatcher/share_umount.sh",
      "Event": "disconnected",
      "EnvVariables": {
        "MOUNT_POINT": "//127.0.0.1/Storage"
      },
      "Included_MacAddresses": [
        "cc:ce:cc:ce:ce:cc"
      ]
    },
        {
      "Script": "$HOME/network-dispatcher/share_mount.sh",
      "Event": "connected",
      "EnvVariables": {
        "MOUNT_POINT": "//192.168.1.1/Storage",
        "MOUNT_LINK": "$HOME/Storage"
      },
      "Included_MacAddresses": [
        "cc:ce:cc:ce:ce:cc"
      ]
    },
    {
      "Script": "$HOME/network-dispatcher/share_umount.sh",
      "Event": "disconnected",
      "EnvVariables": {
        "MOUNT_POINT": "//192.168.1.1/Storage"
      },
      "Included_MacAddresses": [
        "cc:ce:cc:ce:ce:c"
      ]
    }

  ]
}
```
### Specifying correct fstab options
Network dispatcher relies on mount points specified in `/etc/fstab` because it allows to mount and umount as user.

One limitation is `/etc/fstab` does not allow to specify two different network shares pointing to the same mount directory. 
Following does not work and breaks mounting as user:
```
#/etc/fstab
//127.0.0.1/Storage    /home/storage  cifs   noauto,port=4445,rw,users,nodev,relatime 0 0
//192.168.1.1/Storage  /home/storage  cifs   noauto,port=4445,rw,users,nodev,relatime 0 0
```
So correct `/etc/fstab` looks like this.

```
#/etc/fstab
# noauto parameter is crucial because it instruct systemd that share will be mounted manually by our scripts
# port=4445 is local port to establish the ssh tunnel to the NAS. Its hardcoded in cifs_ssh_tunnel.sh
#
//127.0.0.1/Storage  /home/storage_remote  cifs   noauto,port=4445,rw,users,nodev,relatime 0 0
//192.168.1.1/Storage  /home/storage_local  cifs   noauto,port=4445,rw,users,nodev,relatime 0 0
```
But that add inconvenience of having two different paths when at home and outside\
which is solved by specifying `MOUNT_LINK`

* `MOUNT_LINK` - when specified  mount script will create a symbolic link from mounted share to specified `MOUNT_LINK`. 
Provided in the script `"MOUNT_LINK": "$HOME/Storage"` will automatically point /home/Storage to /home/storage_local when at home and
/home/storage_remote when outside


`noauto` parameter is crucial because it instruct systemd that share will be mounted manually by our scripts
`port=4445` is local port to establish the ssh tunnel to the NAS. It's hardcoded in `cifs_ssh_tunnel.sh` 

### Understanding ssh tunnel options
This config implements ssh tunnel to the local cifs share to make it securely accessible through internet
* `SSH_PORT` - port on external hostname to forward your ssh server local port. Local 22-> `SSH_PORT` port forwarding rule must be configured on the router to make it work 
* `SSH_USER` - user to login to your ssh server
* `SSH_HOST` - your external hostname, I use dynamic DNS provider to get my hostname
* `PRIVATE_KEY` - private key to login to your ssh server\

It requires `autossh` package to correctly manage ssh tunnel errors and restart it



