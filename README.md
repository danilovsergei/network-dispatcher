# Description
It's a network dispatcher which allows to execute scrips on wifi connect/disconnect events.

It mitigates big flow in [NetworkManager](https://networkmanager.dev)'s dispatcher which ignores lost wifi events. \
That's it, [NetworkManager](https://networkmanager.dev) only reacts on proper connect/disconnect events when WIFI was disconnect button pressed.\
But with laptops most common case is when laptop is often moved across different WIFI access points and connection lost

This project solves this problem by properly reacting on all WIFI connect/disconnect events.

Motivation to write it was my NAS network each time made dolphin hang when wifi was lost.

# Features
This project main use case was to mount network share directly when on home network \
and mount the same share through ssh tunnel when everywhere else. \

* Listens to WIFI connect/disconnect events
* Allows to specify configurable scripts per event. See the [Usage section](#usage) for more details
* Includes tested scripts to properly mount and umount network shares on connects/disconnects. See the [Usage section](#usage) for more details


# Requirements
* [NetworkManager](https://networkmanager.dev). My network dispatcher listens[NetworkManager](https://networkmanager.dev) low level dbus events
* [Dbus](https://www.freedesktop.org/wiki/Software/dbus/). See above
* [Systemd](https://systemd.io/). My network dispatcher provides systemd service to run itself
* Scripts are using standard mount, umount, fuser, ssh commands to perform operations.

# Install
* Install by executing
```
bash -c "$(curl -L https://raw.githubusercontent.com/danilovsergei/network-dispatcher/main/install.sh)" -- "$HOME/bin"

```
It will download and unpack latest [release](https://github.com/danilovsergei/network-dispatcher/releases/latest/download/network-dispatcher.zip) into `"$HOME/bin"` directory and install network-dispatcher systemd service into `"$HOME/.config/systemd/user/network-dispatcher.service"` \
Everything runs under current user.

After that network disatcher is ready to react on events. However it's necessary to define config with scripts to react. See the [Usage section](#usage)  for examples


# Usage
To react on events create network-dispatcher config in `$HOME/.config/network-dispatcher/config.json`

Here is example of config which mounts and umounts CIFS network share.
This config is using scripts provided in [release](https://github.com/danilovsergei/network-dispatcher/releases/latest/download/network-dispatcher.zip) and installed into `$HOME/network-dispatcher` 

## Script mounts/umount local CIFS share
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
//192.168.1.1/Storage  /home/<your_user>/storage  cifs   noauto,rw,users,nodev,relatime 0 0
```
`noauto` parameter is crucial because it instruct systemd that share will be mounted manually by our scripts


## Script mounts/umount local CIFS share through SSH tunnel
This config instructs to mount CIFS network share through SSH tunnel which is created each time WIFI is connected

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
      }
    },
    {
      "Script": "$HOME/network-dispatcher/share_mount.sh",
      "Event": "connected",
      "EnvVariables": {
        "MOUNT_POINT": "//127.0.0.1/Storage"
      }
    },
    {
      "Script": "$HOME/network-dispatcher/share_umount.sh",
      "Event": "disconnected",
      "EnvVariables": {
        "MOUNT_POINT": "//127.0.0.1/Storage"
      }
    }
  ]
}
```

`share_mount.sh` and `share_umount.sh` relies on `/etc/fstab` record to be able to work as user.  So please add mount line to your `/etc/fstab`

```
//127.0.0.1/Storage  /home/<your_user>/storage  cifs   noauto,port=4445,rw,users,nodev,relatime 0 0
```

`noauto` parameter is crucial because it instruct systemd that share will be mounted manually by our scripts
`port=4445` is local port to establish the ssh tunnel to the NAS. It's hardcoded in `cifs_ssh_tunnel.sh` 

