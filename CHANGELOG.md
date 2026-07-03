## [unreleased]

### ⚙️ Miscellaneous Tasks

- [[a68987d…](https://github.com/danilovsergei/remote-torrent-adder/commit/a68987dc48d97c32ee24322d39fba731e1a9e290)] Add github release workflow

### 📚 Documentation

- [[bcd875f…](https://github.com/danilovsergei/remote-torrent-adder/commit/bcd875f849971b002e5dd5dda7cecf2394832b7d)] Update Install instructions at readme

### 🚀 Features

- [[5a3ad31…](https://github.com/danilovsergei/remote-torrent-adder/commit/5a3ad31a9a05178f41313fdcea57686c03662cc0)] Update install.sh script to install systemd service into the /etc/systemd/system
- [[b860daf…](https://github.com/danilovsergei/remote-torrent-adder/commit/b860daf85f5830f903c79ff5115fdf0939de9d1c)] Improve share_mount.sh script
## [2025-12-20] - 2025-12-21

### 🐛 Bug Fixes

- [[d95c1a6…](https://github.com/danilovsergei/remote-torrent-adder/commit/d95c1a673251f0ec42fe062966525d49b27d865d)] Fix gateway not found error from previous retries printing in the end when gateway was already found

### 🚀 Features

- [[6dee7ff…](https://github.com/danilovsergei/remote-torrent-adder/commit/6dee7fff39562bddffa1ac14a304c41d492e5ee1)] Fix the  100% cpu usage error
- [[c28fd54…](https://github.com/danilovsergei/remote-torrent-adder/commit/c28fd54c578cc78dc8f74b24ff4652cd7aefc8c1)] Use local dir name instead of cifs share name to allow mount as user
- [[29506d9…](https://github.com/danilovsergei/remote-torrent-adder/commit/29506d92d4bf8fd63a90e9ef4fcd99bbe7584774)] Run network-dispatcher service before NetworkManager to catch onConnected events during boot
## [2025-10-12] - 2025-10-12

### ⚙️ Miscellaneous Tasks

- [[879a535…](https://github.com/danilovsergei/remote-torrent-adder/commit/879a53571f95157c0d3a436c6e6e21091916c0ea)] Update gitcliff to v4

### 📚 Documentation

- [[472c843…](https://github.com/danilovsergei/remote-torrent-adder/commit/472c843dc48b774c2052277e916e7d3ef0f0431f)] Add default cifs port to the cifs tunnel example

### 🚀 Features

- [[a02c215…](https://github.com/danilovsergei/remote-torrent-adder/commit/a02c2151a0bb7d01e8a24dc08af7d94c0bd2f452)] Do not continue other scripts execution if ContinueOnFail is false
- [[cef253b…](https://github.com/danilovsergei/remote-torrent-adder/commit/cef253b0b2b341340fa6882f1f8c1b7f93f37279)] Expose local cifs port to be configurable
## [2024-08-01] - 2024-08-01

### ⚙️ Miscellaneous Tasks

- [[d05e303…](https://github.com/danilovsergei/remote-torrent-adder/commit/d05e303cf30659f199f9508e218a92dfe2508d3f)] Generate automated changelog using git-cliff

### 🐛 Bug Fixes

- [[248ff51…](https://github.com/danilovsergei/remote-torrent-adder/commit/248ff51cc18504f5ba9eb4c1ebbcedae3b13afd1)] Fix error message in the log about umount is not able to find umount process pid to kill
- [[5891052…](https://github.com/danilovsergei/remote-torrent-adder/commit/5891052fa945f15836dae4c06f6a341ba29d4442)] Make establishing cifs ssh tunnel more reliable
- [[047eada…](https://github.com/danilovsergei/remote-torrent-adder/commit/047eada34e3ff985d2cbb5302b2ca8a722b330a2)] Fixed error message on killing umount process group if it still exists after successful amount
- [[d16187d…](https://github.com/danilovsergei/remote-torrent-adder/commit/d16187d0b9e18dd423b9f8b7e8d39ccea8d5c90f)] Kill umount process only if it exists with checks for both process pid and group pid
- [[bf6a234…](https://github.com/danilovsergei/remote-torrent-adder/commit/bf6a23491d41df4758337d18152f5b05fc42f3c1)] Avoid failing shell script on non empty error output

### 🚀 Features

- [[b24895a…](https://github.com/danilovsergei/remote-torrent-adder/commit/b24895af60996ef3e89a8dea68eac20996697bb3)] Rework dbus api to connect/disconnect events
## [2024-05-21] - 2024-05-22

### ⚙️ Miscellaneous Tasks

- [[7a5971a…](https://github.com/danilovsergei/remote-torrent-adder/commit/7a5971ae13313ff761a870f6a6b705bad7b66751)] Migrate shell scripts to print only to systemd journal
- [[1b0175b…](https://github.com/danilovsergei/remote-torrent-adder/commit/1b0175b530e5e1eff5c32240dfac9836fcd8ff58)] Reduce noice from autossh

### 📚 Documentation

- [[facf1f0…](https://github.com/danilovsergei/remote-torrent-adder/commit/facf1f0ef0a906cf0a93494635f1d4b8112f9d90)] Small defails to ssh tunnel section
- [[5a72d17…](https://github.com/danilovsergei/remote-torrent-adder/commit/5a72d17545138ed6eebba54c865b87ca30c41a53)] Updated SSH_PORT info in ssh tunnel section
- [[4e26f9c…](https://github.com/danilovsergei/remote-torrent-adder/commit/4e26f9c21cad0a30b8a4b704f0d7b2541078ab06)] Updated readme with filter options and MOUNT_LINK
- [[a8596a9…](https://github.com/danilovsergei/remote-torrent-adder/commit/a8596a9b25246335ac59118dab25478d45997fb7)] Updated readme with troubleshooting

### 🚀 Features

- [[414f763…](https://github.com/danilovsergei/remote-torrent-adder/commit/414f76388d14eb8a1f01ae4e647a7b613b870fa9)] Add --config flag to specify custom config file location
- [[f87aa16…](https://github.com/danilovsergei/remote-torrent-adder/commit/f87aa16f938f3b6956c257d994ea76c76001ec9d)] Added IncludedMacAddresses and ExcludedMacAddresses config option to define filters
- [[e7358a7…](https://github.com/danilovsergei/remote-torrent-adder/commit/e7358a7edce94dd588df5f02562e75caec8199ad)] Share_mount.sh script can create a symbolic link to the mounted folder
## [2024-05-19] - 2024-05-20

### 🐛 Bug Fixes

- [[880fede…](https://github.com/danilovsergei/remote-torrent-adder/commit/880feded02937b34a02726bd7b12ec831fc94cbf)] Don't wait for gateway IP address if Wifi is disconnected or gateway already received

### 💼 Other

- [[53d5aff…](https://github.com/danilovsergei/remote-torrent-adder/commit/53d5aff42079c8fe395fef129a58c23eb9e4d58d)] Update changelong workflow
- [[3a187de…](https://github.com/danilovsergei/remote-torrent-adder/commit/3a187de4f3e310c1d9498327f8529edcb231a69f)] Update generate github changelog to comply to nodejs 20

### 📚 Documentation

- [[3ff6247…](https://github.com/danilovsergei/remote-torrent-adder/commit/3ff6247704ec0f6660357eb29a8ee5b7e5f08b4b)] Update README with more details how ssh tunnel works
## [05/19/2024] - 2024-05-20

### 🚀 Features

- [[db73ddf…](https://github.com/danilovsergei/remote-torrent-adder/commit/db73ddfef4b176726125f1ea6e0701fb51686c34)] Forcefully kill the previous running script if new network event happen and previous script is still running
- [[92ee056…](https://github.com/danilovsergei/remote-torrent-adder/commit/92ee056807d7ed75d90c30414f0d1397935fe853)] Prevent mount script to write mount attempts to stderr since it makes dispatcher to report script failed
## [05/15/2024] - 2024-05-16
