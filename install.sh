#!/bin/bash

# location network-dispatcher binary will be installed
bin_dir="$1/network-dispatcher"

bin_name="network-dispatcher"
service_file=network-dispatcher.service
service_file_path=$bin_dir/systemd/$service_file
systemd_dir=$HOME/.config/systemd/user
release_name="network-dispatcher.zip"
latest_release="https://github.com/danilovsergei/network-dispatcher/releases/latest/download/$release_name"

# creates directory if it does not exist
function create_dir() {
  if [[ ! -d "$1" ]]; then
    mkdir -p "$1"
  fi
}

if [ $# -eq 0 ]; then
  echo "Error: No installation dir is provided"
  echo "Example: install.sh ~/local/kde-turn-off-screen"
  exit 1
fi

create_dir $bin_dir

# Copy and unzip release
echo -e "Download latest release from github $latest_release\n"
temp_dir=$(mktemp -d)
cd "$temp_dir"
curl -s -L "$latest_release" -o "$temp_dir/$release_name"
unzip -o "$temp_dir/$release_name" -d "$1"

chmod +x $bin_dir/$bin_name
chmod +x $bin_dir/*.sh

# prepare systemd service
echo -e "Replace <bin_dir> with $bin_dir in $service_file_path\n"  
sed -i "s|=<bin_dir>|=$bin_dir|g" $service_file_path

echo -e "Replace <username> with $USER in $service_file_path\n"
sed -i "s|<username>|$USER|g" $service_file_path

if ! ( grep -q "$bin_dir/$bin_name" "$service_file_path" ); then
  echo "Failed to replace ExecStart in $service_file_path to $bin_dir/$bin_name"
fi

echo -e "Remove TODO line from $service_file_path \n"
sed -i '/^# TODO/d' "$service_file_path"

#Install systemd service
echo -e "Copy $service_file_path to $systemd_dir\n"
create_dir $systemd_dir
cp -f $service_file_path "$systemd_dir/"

echo -e "Start $service_file_path service\n"
systemctl --user enable $service_file
systemctl --user start $service_file

service_status=$(systemctl --user status $service_file)
if ! ( echo $service_status | grep -q "active" ); then
  echo -e "\nFailed to start $service_file\n"
  echo "Run systemctl --user status $service_file for details"
fi

echo -e "Service $service_file started\n"
