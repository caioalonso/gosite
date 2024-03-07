---
id: 1
title: Restarting fancontrol after resume from sleep/suspend
date: 2019-01-23
hidden: true
aliases:
    - /posts/restarting-fancontrol-after-resume
    - /2019/01/23/restarting-fancontrol-after-resume.html
---

Whenever I set up a new Linux system I install [lm-sensors](https://github.com/lm-sensors/lm-sensors) and enable the fancontrol utility. Most of the time fancontrol doesn't get restarted when resuming from sleep/suspend. This is how I fix this with systemd:

/etc/systemd/system/fancontrol-resume.service
```
[Unit]
Description=Restart fancontrol after resume from sleep/suspend
After=suspend.target

[Service]
Type=oneshot
ExecStart=/bin/systemctl restart fancontrol.service

[Install]
WantedBy=suspend.target
```

Then:
```
sudo systemctl daemon-reload
sudo systemctl enable fancontrol-resume.service
```
