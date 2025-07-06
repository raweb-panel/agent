# raweb panel agent. ![](https://jenkins.julio.al/job/raweb-agent/badge/icon)

---

 This version of agent, is a free copy, the features are limited without limiting the functionality of app deployments, management, or other core features. 
 The premium version is build for multi-node support, so any PR trying to tap into that field is respectfully declined.

---

## Installation 

- Debian 12/11.
- Ubuntu 24.04/22.04.
```bash
apt update; apt install -y wget apt-transport-https ca-certificates gnupg2 sudo
echo "deb [trusted=yes] https://repo.julio.al/ $(cat /etc/os-release | grep VERSION_CODENAME= | cut -d= -f2) main" | sudo tee /etc/apt/sources.list.d/raweb.list
sudo apt update; sudo apt install -y raweb-agent
```

---

## Features :
 - Container controller (create, start, stop, kill, info).
 - Container stats (cpu, ram, bandwidth).
 - Images (pull, delete).
 - System user handler. (create, delete).
 - Nginx Config handler. (create, edit, delete).
 - Uses docker socket without need of exposing tcp for api usage.

 --- 

## License

- By downloading, installing, or using the software, the Licensee agrees to be bound by the terms of this License. [Custom Deployment License (CDL)](./LICENSE).  
- Free for commercial use on up to 100 nodes and 100 websites per node.
- "Developer Statistics", Must remain enabled for the simply purpose of above numbers tracking.
- Redistribution, resale, or large-scale deployment requires a commercial license.

Contact me@julio.al licensing or additional rights/questions.
