# ML-KEM Web Application — Deployment Guide

**For network administrators.** This guide covers deploying the ML-KEM Post-Quantum
Key Encapsulation demo on any Linux server so users on the local network can access it.

---

## What this application is

A browser-based interactive learning tool that demonstrates the ML-KEM (FIPS 203)
post-quantum key encapsulation algorithm. It has two parts:

| Component | Technology | Purpose |
|-----------|-----------|---------|
| Frontend  | React + Nginx | Serves the web UI on port 3000 |
| Backend   | Go (WebSocket) | Runs the cryptographic steps on port 8080 |

Both run as Docker containers. The frontend proxies `/ws` traffic to the backend
internally — only **port 3000** needs to be exposed to the network.

---

## Requirements

- A Linux server (Ubuntu 20.04+ or similar)
- **Docker Engine** 24+ installed
- **Docker Compose plugin** (the `docker compose` command, not `docker-compose`)
- At least 512 MB RAM and 2 GB disk space
- The project source folder copied to the server

### Install Docker (if not already installed)

```bash
# Ubuntu / Debian
sudo apt-get update
sudo apt-get install -y ca-certificates curl
sudo install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg \
  | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] \
  https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" \
  | sudo tee /etc/apt/sources.list.d/docker.list
sudo apt-get update
sudo apt-get install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin
sudo systemctl enable --now docker
```

Verify:
```bash
docker --version        # should show Docker 24+
docker compose version  # should show v2.x
```

---

## Step 1 — Copy the project to the server

From your machine, create a compressed archive and transfer it:

```bash
# On your machine (run from the project parent folder)
tar -czf mlkem-webapp.tar.gz mlkem-webapp/

# Transfer to server (replace SERVER_IP with the actual address)
scp mlkem-webapp.tar.gz admin@SERVER_IP:/opt/
```

On the server, extract it:

```bash
cd /opt
sudo tar -xzf mlkem-webapp.tar.gz
cd mlkem-webapp
```

---

## Step 2 — Configure the port (optional)

By default the app runs on port **3000**. To change it, open `docker-compose.yml`
and edit the frontend ports line:

```yaml
ports:
  - "80:80"      # use port 80 (standard HTTP) — requires root or port forwarding
  - "8888:80"    # or any other port you prefer
```

No other file needs to change — the backend stays internal.

---

## Step 3 — Build and start

```bash
cd /opt/mlkem-webapp
sudo docker compose up --build -d
```

- `--build` compiles the Go backend and the React frontend from source
- `-d` runs everything in the background

First build takes 2-5 minutes depending on internet speed (it downloads base images).
Subsequent starts (without `--build`) take under 10 seconds.

---

## Step 4 — Verify it is running

```bash
sudo docker compose ps
```

You should see both containers with status **Up**:

```
NAME                      STATUS
mlkem-webapp-backend-1    Up
mlkem-webapp-frontend-1   Up
```

Test locally on the server:
```bash
curl -I http://localhost:3000
# Expected: HTTP/1.1 200 OK
```

---

## Step 5 — Find the server IP and share the URL

```bash
ip addr show | grep "inet " | grep -v 127.0.0.1
```

Share this URL with users on your network:

```
http://<SERVER_IP>:3000
```

Example: `http://192.168.1.50:3000`

Users open this in any modern browser (Chrome, Firefox, Edge, Safari).
No installation needed on their devices.

---

## Firewall (if applicable)

If the server has a firewall enabled, allow port 3000:

```bash
# Ubuntu UFW
sudo ufw allow 3000/tcp
sudo ufw reload

# Or for firewalld (RHEL/CentOS)
sudo firewall-cmd --permanent --add-port=3000/tcp
sudo firewall-cmd --reload
```

---

## Managing the application

| Task | Command |
|------|---------|
| Start | `sudo docker compose up -d` |
| Stop | `sudo docker compose down` |
| Restart | `sudo docker compose restart` |
| View logs | `sudo docker compose logs -f` |
| Update (after code change) | `sudo docker compose up --build -d` |
| Check status | `sudo docker compose ps` |

Run all commands from inside `/opt/mlkem-webapp/`.

---

## Auto-start on server reboot

The containers already have `restart: unless-stopped` in `docker-compose.yml`,
so they will start automatically after a reboot as long as the Docker service itself
starts on boot:

```bash
sudo systemctl enable docker
```

---

## Troubleshooting

**Port 3000 already in use:**
```bash
sudo lsof -i :3000   # find what is using it
# Then either stop that process or change the port in docker-compose.yml
```

**Frontend loads but shows "Disconnected":**
The backend container is not reachable. Check:
```bash
sudo docker compose logs backend
```

**Build fails with "no space left on device":**
```bash
sudo docker system prune -f   # remove unused images and containers
```

**See live application logs:**
```bash
sudo docker compose logs -f frontend   # nginx access logs
sudo docker compose logs -f backend    # Go server logs
```

---

## Architecture summary

```
Browser (user)
     │
     │  HTTP   port 3000
     ▼
[Nginx / Frontend container]
     │
     │  WebSocket  /ws  (internal Docker network)
     ▼
[Go Backend container]  port 8080 (not exposed externally)
```

The backend port 8080 is **not** exposed to the network — only Nginx on port 3000
is reachable from outside, which proxies WebSocket connections internally.

---

*For issues contact the project maintainer.*
