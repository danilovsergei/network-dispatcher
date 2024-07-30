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

function kill_pid() {
pid="$1"
if [ "$pid" -gt 0 ]; then
  if ps -p "$pid" > /dev/null; then
    kill -- "$pid" 2>/dev/null
    do_log "Umount process $pid killed after share umounted"
  else
    do_log "Umount process $pid sucessfully finished"
  fi
elif [ "$pid" -lt 0 ]; then
  if ps -o pgid= | grep -q "^ ${$pid/-}$"; then
    kill -- "-$pid" 2>/dev/null
    do_log "Umount process $pid killed after share umounted"
  else
    do_log "Umount process $pid sucessfully finished"
  fi
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
  # don't care about returned error code since check shared umounted directly
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
    kill_pid ${umount_pid}
    exit 0
else
  do_log  "$MOUNT_POINT absent in mtab. Do nothing";
fi