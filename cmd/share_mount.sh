#!/bin/bash
function do_log() {
  logger -t network-dispatcher "$@"
}

function check_env_variable() {
    env_var_name=$1
    env_var_value="$2"
    if [ -z "$2" ]; then
        echo "ERROR: $1 is not defined." >&2
        exit 1  # Exit with failure code
    fi
}

function create_mount_dir_link() {
  mountpoint=$(findmnt -S "$MOUNT_POINT" -n -o TARGET)
  if [ ! -e "$MOUNT_LINK" ]; then
    echo "Create new $MOUNT_LINK to $mountpoint"
    ln -s "$mountpoint" "$MOUNT_LINK"
    return
  fi
  if [ -L "$MOUNT_LINK" ]; then
    echo "Repoint $MOUNT_LINK to $mountpoint"
    unlink "$MOUNT_LINK"
    ln -s "$mountpoint" "$MOUNT_LINK"
    else
      echo "$MOUNT_LINK is regular file or directory. Provide another path"
  fi
}

check_env_variable MOUNT_POINT $MOUNT_POINT

# do_log "Restart cifs tunnel"
# systemctl --user restart cifs_tunnel

if grep -qw $MOUNT_POINT /etc/mtab; then
  do_log "mtab contains $MOUNT_POINT. Already mounted. Do nothing";
else
  # match only whole word match using w to avoid substring matches
  while ! grep -qw $MOUNT_POINT "/etc/mtab"
    do
        mount_output=$(mount $MOUNT_POINT 2>&1)
        if [[ $? -ne 0 ]]; then
          do_log "Retry mount $MOUNT_POINT due to error : $mount_output"
          sleep 1
        fi
    done
    do_log "Mounted $MOUNT_POINT"
    # create mount link if MOUNT_LINK is provided
    if ! [ -z "${MOUNT_LINK}" ]; then
          create_mount_dir_link
    else
      echo "Link to mounted folder is not provided. Skipping link creation"
    fi
    exit 0
fi

exit $?
