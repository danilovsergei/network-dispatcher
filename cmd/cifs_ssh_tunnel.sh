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
local_port=4445
remote_cifs_port=445

check_env_variable SSH_PORT $SSH_PORT
check_env_variable SSH_USER $SSH_USER
check_env_variable SSH_HOST $SSH_HOST
check_env_variable PRIVATE_KEY $PRIVATE_KEY

# kill previous ssh tunnel to prevent any stale process
fuser -k 4445/tcp
ssh -fNT -o ExitOnForwardFailure=yes -o ServerAliveInterval=10 -o ServerAliveCountMax=3 -i $PRIVATE_KEY -p $SSH_PORT -L $local_port:127.0.0.1:$remote_cifs_port $SSH_USER@$SSH_HOST