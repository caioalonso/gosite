#!/bin/sh
ssh -p 29266 caio@caioalonso.com "cd /home/caio/dev/gosite && git pull && /usr/local/go/bin/go build && systemctl --user restart gosite"
