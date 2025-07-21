# raweb panel agent. ![](https://jenkins.julio.al/job/raweb-agent/badge/icon) [![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=raweb-panel_agent&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=raweb-panel_agent)

---

 This version of agent, is a free copy, the features are limited without limiting the functionality of app deployments, management, or other core features. 
 The premium version is build for multi-node support, so any PR trying to tap into that field is respectfully declined.

---

## Requirements 

- Debian 12/11.
- Ubuntu 24.04/22.04.
- AlmaLinux 9/8.
- Docker, should be present on the same server. (free version limit).
- Raweb Panel, should be present on the same server. (free version limit).

This is part of the panel, it cannot run standalone, so you can follow installation steps on the [panel](https://github.com/raweb-panel/raweb) repository.

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

- By downloading, installing, or using the software, the Licensee agrees to be bound by the terms of this License. [Custom Deployment License (CDL)](./LICENSE.md).  
- Free for commercial use on up to unlimited nodes and 100 websites per node.
- "Developer Statistics", Must remain enabled for the simply purpose of above numbers tracking.
- Redistribution, resale, or large-scale deployment requires a commercial license.

Contact me@julio.al licensing or additional rights/questions.
