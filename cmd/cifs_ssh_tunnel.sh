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

if ss -tln | grep -qw "127.0.0.1:$local_port" ; then
  # kill previous ssh tunnel. autossh will automatically start new tunnel
  fuser -k $local_port/tcp 2>/dev/null
else
  # start new ssh tunnel if nothing is listening
  AUTOSSH_PORT=0
  AUTOSSH_GATETIME=0
  autossh -fNT -o ExitOnForwardFailure=yes -o ServerAliveInterval=10 -o ServerAliveCountMax=3 -i $PRIVATE_KEY -p $SSH_PORT -L $local_port:127.0.0.1:$remote_cifs_port $SSH_USER@$SSH_HOST
fi