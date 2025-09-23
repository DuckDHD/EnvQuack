# 🦆 EnvQuack

**EnvQuack** is a CLI tool that quacks when your environment variables drift.  
No more silent hours wasted debugging missing `.env` keys or mismatched `docker-compose.yml` configs.  
Let the gopher-in-a-duck-suit yell at you *before* production does.  

---

## ✨ Features

- ✅ Validate `.env` against `.env.example`
- ✅ Detect missing or extra variables
- ✅ Sync `.env` with placeholders for missing vars
- 🐳 Parse `docker-compose.yml` for env drift (coming soon)
- 🐋 Parse `Dockerfile` for `ARG`/`ENV` usage (coming soon)
- ☸️ Kubernetes `ConfigMap` / `Secret` support (planned)

---

## 🚀 Install

```bash
go install github.com/duckdhd/envquack/cmd/envquack@latest
