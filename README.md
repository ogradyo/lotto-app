# Lotto App (Powerball & Mega Millions)

A small Go + Postgres backend to store lottery tickets (Powerball and Mega Millions) and prepare for later grading against official draw results.

This version (v0.1) provides:

- A Go HTTP API with:
  - `GET /healthz` – health check
  - `POST /api/tickets` – store a single ticket (5 white numbers + 1 special)
- A Postgres schema with `users`, `draws`, and `tickets` tables
- A development setup where:
  - Postgres runs on a Proxmox VM
  - The Go API runs locally on a Mac and connects to that VM

Later versions will add a mobile‑friendly UI and ticket grading logic.

---

## Architecture

- **Backend:** Go HTTP service (net/http + chi + pgxpool)
- **Database:** Postgres on a Proxmox VM
- **Current topology:**
  - Postgres: Ubuntu VM at `10.83.91.246:5432`
  - Go API: running on the Mac (localhost, port `8080`)

### Data model (current)

- `users`
  - basic owner record for tickets
- `draws`
  - official draw results (per game + date), not yet fully populated/used
- `tickets`
  - one row per ticket line:
    - game (`POWERBALL` or `MEGAMILLIONS`)
    - draw date
    - 5 white ball numbers
    - 1 special ball (Powerball or Mega Ball)
    - optional multiplier
    - optional image URL

---

## Prerequisites

- Go (1.22+ recommended) installed on the Mac
- `psql` client installed (e.g. via Homebrew)
- A Linux VM (e.g. Ubuntu) on Proxmox with:
  - Static IP `10.83.91.246`
  - Postgres installed and running

---

## Postgres Setup (on the VM)

All commands in this section run on the VM at `10.83.91.246`.

### 1. Install Postgres

