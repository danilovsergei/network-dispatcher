#!/bin/bash

function do_log() {
  logger -t network-dispatcher "$@"
}

function check_env_variable() {
    env_var_name=$1
    env_var_value="$2"
    if [ -z "$2" ]; then
        do_log "ERROR: $1 is not defined."
        exit 1  # Exit with failure code
    fi
}

check_env_variable MOUNT_POINT $MOUNT_POINT

# match only whole word match using w to avoid substring matches
if grep -qw $MOUNT_POINT /etc/mtab; then
  do_log "mtab contains $MOUNT_POINT";
  set -m
  # use mount source to get mount destination for umount command
  mount_dest=$(awk -v mount_point="$MOUNT_POINT" '$1 == mount_point {print $2}' /etc/mtab)
  check_env_variable mount_dest $mount_dest

  umount -l $mount_dest &
  umount_pid=$!
  
  # match only whole word match using w to avoid substring matches
  while grep -qw $MOUNT_POINT "/etc/mtab"
    do
        sleep 0.1
        do_log "umounting attempt for $MOUNT_POINT"
    done
    do_log "umounted $MOUNT_POINT"
    # For some reason umount -l command process could stuck and block script exit.
    # Despite that it umounts storage successfully.
    # Kiling using command from the guide https://copyprogramming.com/howto/killing-background-processes-when-script-exists-duplicate
    if [[ -n "${umount_pid}" ]]; then
      kill -- -${umount_pid}  || true
    fi

else
  do_log  "$MOUNT_POINT absent in mtab. Do nothing";
fi