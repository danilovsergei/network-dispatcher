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

check_env_variable MOUNT_POINT $MOUNT_POINT

# do_log "Restart cifs tunnel"
# systemctl --user restart cifs_tunnel

if grep -qw $MOUNT_POINT /etc/mtab; then
  do_log "mtab contains $MOUNT_POINT. Already mounted. Do nothing";
else
  # match only whole word match using w to avoid substring matches
  while ! grep -qw $MOUNT_POINT "/etc/mtab"
    do
        do_log "Mounting attempt for $MOUNT_POINT"
        mount $MOUNT_POINT
        sleep 1

    done
    do_log "Mounted $MOUNT_POINT"
    exit 0
fi

exit $?
