#!/bin/sh
function check_env_variable() {
    env_var_name=$1
    env_var_value="$2"
    if [ -z "$2" ]; then
        echo "ERROR: $1 is not defined." >&2
        exit 1  # Exit with failure code
    fi
}

check_env_variable MOUNT_POINT $MOUNT_POINT

echo "MOUNT_POINT="$MOUNT_POINT