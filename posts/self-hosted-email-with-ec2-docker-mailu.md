---
title: Self-hosted email with EC2, Docker and Mailu
date: 2018-10-27
aliases:
    - /posts/self-hosted-email-with-ec2-docker-mailu
    - /2018/10/27/self-hosted-email-with-ec2-docker-mailu.html
---

These are the steps I followed when setting up my own mail server.

Fire up an Ubuntu EC2 instance, then:
```
sudo apt update && sudo apt upgrade -y
sudo apt-get install -y apt-transport-https ca-certificates curl software-properties-common git vim
```

Make sure these ports are open in your AWS security group:
- 25
- 80
- 110
- 143
- 443
- 465
- 587
- 993
- 995

Install Docker:
```
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -
sudo add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable"
sudo apt-get update
sudo apt-get -y install docker-ce
```

Install docker-compose:
```
sudo curl -L "https://github.com/docker/compose/releases/download/1.22.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose
```

Install [Mailu](https://mailu.io/):
```
sudo mkdir /mailu
cd /mailu
sudo wget https://mailu.io/1.5/_downloads/docker-compose.yml
sudo wget https://mailu.io/1.5/_downloads/.env
```

Configure Mailu with `sudo vim .env` and set these entries:
- `SECRET_KEY` to a random 16 byte string;
- `BIND_ADDRESS4` to the EC2 instance's **private IP**;
- `DOMAIN` to the domain of your mail server;
- `HOSTNAMES` also to the domain of your server;
- `TLS_FLAVOR=tlsencrypt`;
- `DISABLE_STATISTICS=true`;
- `ADMIN=true`;
- `WEBMAIL=rainloop`;
- `WELCOME=true`;

Run it with:
```
sudo docker-compose up -d
```

Add an admin account with, replacing the uppercase strings:
```
sudo docker-compose run --rm admin python manage.py admin USERNAME DOMAIN PASSWORD
```

## Antispam measures
Test it with [mail-tester](https://www.mail-tester.com/).

### SPF
Add a `TXT` record to your nameserver. Name `caioalonso.com.` and content `v=spf1 mx a:mail.caioalonso.com -all`.

### DMARC
Go to your admin panel (https://`DOMAIN`/admin), then `Mail domains > the Details icon` and click `Regenerate keys`.

Copy the `DNS DMARC entry` line and do the same thing you did for the SPF. This time with name `_dmarc.caioalonso.com` and the content is everything between quotes, in my case:
```
v=DMARC1; p=reject; rua=mailto:admin@mail.caioalonso.com; ruf=mailto:admin@mail.caioalonso.com; adkim=s; aspf=s
```

### DKIM
On that same page do the same thing you did with DMARC, but using the `DNS DKIM Entry`.
