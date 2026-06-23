# Deploying to Render (Free Tier)

Your app has two services — a Go backend (WebSocket) and a React frontend.
Deploy them as separate Render services.

---

## Prerequisites

- Code pushed to a GitHub repo (public or private)
- A free account at [render.com](https://render.com)

---

## Step 1 — Backend: Web Service

1. Render dashboard → **New → Web Service**
2. Connect your GitHub repo
3. Set these options:

   | Field | Value |
   |---|---|
   | Root Directory | `backend` |
   | Environment | `Docker` |
   | Dockerfile Path | `./Dockerfile` |
   | Instance Type | Free |

4. Click **Create Web Service**
5. Wait for the build to finish, then copy your backend URL:
   `https://YOUR-BACKEND-NAME.onrender.com`

---

## Step 2 — Frontend: Static Site

1. Render dashboard → **New → Static Site**
2. Connect the same repo
3. Set these options:

   | Field | Value |
   |---|---|
   | Root Directory | `frontend` |
   | Build Command | `npm ci && npm run build` |
   | Publish Directory | `dist` |

4. Add an **Environment Variable**:

   | Key | Value |
   |---|---|
   | `VITE_WS_URL` | `wss://YOUR-BACKEND-NAME.onrender.com/ws` |

   > Use `wss://` (not `https://`) and replace with your actual backend URL from Step 1.

5. Click **Create Static Site**

---

## How it works

```
Browser → Static Site (frontend) → wss://backend.onrender.com/ws → Go server
```

Render auto-deploys both services on every push to your main branch.

---

## Free Tier Gotcha

Backend instances **spin down after 15 min of inactivity** and take ~30s to cold-start.
The frontend will briefly show "Disconnected" — the auto-reconnect in the app handles this automatically.
