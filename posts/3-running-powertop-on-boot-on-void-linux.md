---
id: 3
title: Running PowerTOP on boot on Void Linux
date: 2018-11-05
aliases:
    - /posts/running-powertop-on-boot-on-void-linux
    - /2018/11/05/running-powertop-on-boot-on-void-linux.html
---

[PowerTOP](https://01.org/powertop/) doesn't remember settings between restarts, so it needs to be executed on every boot. Here's how I did it on Void Linux with runit.

Runit executes every script in `/etc/runit/core-services` once on boot. So I created a `/etc/runit/core-services/97-powertop.sh` script:
```
msg "Powertop autotune..."
powertop --auto-tune
```

And made sure that it was executable with `sudo chmod +x /etc/runit/core-services/97-powertop.sh`.

Since PowerTOP's auto tune enables power management for every available device, my USB mouse kept turning off after 2 seconds of standby. In order to increase that delay to 10 minutes I changed the service file to this:
```
msg "Powertop autotune..."
echo $((10 * 60 * 1000)) > /sys/bus/usb/devices/1-1.6/power/autosuspend_delay_ms
powertop --auto-tune
```

The `/sys/bus/.../autosuspend_delay_ms` path will change depending on the device, so to discover it run `sudo powertop` and manually switch the flag in the Tunables tab.
