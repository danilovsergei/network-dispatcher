#!/bin/sh

project_dir=$(dirname `realpath "$0"`)
bin_dir=$(echo $project_dir"/bin/network-dispatcher")
timestamp=$(date -d "@$(date +%s)" +"%y-%m-%d")
cd "$project_dir"

go build -ldflags="-s -w" -o $bin_dir/network-dispatcher

cp cmd/* $bin_dir/
chmod +x $bin_dir/*.sh

cp -fr systemd $bin_dir/

cd  $project_dir"/bin"
zip -r $timestamp-release.zip network-dispatcher