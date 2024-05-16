Copy service to
~/.config/systemd/user folder

Enable service
systemctl --user enable network-dispatcher.service

Start service
systemctl --user start network-dispatcher.service
