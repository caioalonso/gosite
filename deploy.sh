#!/bin/sh
ssh caioalonso.com "cd /home/caio/dev/gosite && git pull && go build && systemctl --user restart gosite"
