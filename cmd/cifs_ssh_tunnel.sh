#!/bin/bash
function do_log() {
  logger -t network-dispatcher "$@"
}

function check_env_variable() {
    env_var_name=$1
    env_var_value="$2"
    if [ -z "$2" ]; then
        do_log "ERROR: $1 is not defined." >&2
        exit 1  # Exit with failure code
    fi
}
remote_cifs_port=4445

check_env_variable SSH_PORT $SSH_PORT
check_env_variable SSH_USER $SSH_USER
check_env_variable SSH_HOST $SSH_HOST
check_env_variable PRIVATE_KEY $PRIVATE_KEY
check_env_variable LOCAL_CIFS_PORT $LOCAL_CIFS_PORT

# restart both autossh and ssh because sometimes autossh does not restart ssh process when its killed
autossh_pids=$(ps aux | grep "autossh" | grep "$remote_cifs_port:127.0.0.1:" |  grep -v grep | awk '{print $2}')
IFS=$'\n'
for autossh_pid in $autossh_pids 
do
  autossh_pid=$(echo $autossh_pid | tr -d ' ')
  group_id=$(ps -o pgid= $autossh_pid | tr -d ' ')
  if [ -n "$group_id" ]; then
    do_log "kill autossh $group_id and child ssh process"
    kill -9 -$group_id
  fi
done

AUTOSSH_PORT=0
AUTOSSH_GATETIME=0
# Reduce noise from ssh tunnel errors/restarts etc 
AUTOSSH_LOGLEVEL=0
autossh -fNT -o ExitOnForwardFailure=yes -o ServerAliveInterval=10 -o ServerAliveCountMax=3 -i $PRIVATE_KEY -p $SSH_PORT -L $remote_cifs_port:127.0.0.1:$LOCAL_CIFS_PORT $SSH_USER@$SSH_HOST

# check ssh is listening on port. tunnel takes some time to come up
start_time=$(date +%s)
max_wait_time_sec=20
while true; do
  ssh_tunnel_pid=$(fuser -n tcp $remote_cifs_port 2>/dev/null| awk -F":" '{print $1}')
  if [ -n "$ssh_tunnel_pid" ]; then
    do_log "ssh tunnel is listening on port $remote_cifs_port after $elapsed_time seconds"
    exit 0
  else
    current_time=$(date +%s)
    elapsed_time=$((current_time - start_time))
    if [ $elapsed_time -ge $max_wait_time_sec ]; then
      do_log "Error: ssh tunnel failed to start after $elapsed_time seconds" >&2
      exit 1
    fi
    sleep 0.5
  fi
done
