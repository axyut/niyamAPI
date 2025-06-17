---
title: Niyam API
emoji: 🦀
colorFrom: blue
colorTo: pink
sdk: docker
pinned: false
---

# niyamAPI

System for the Niyam project.
https://niyam.onrender.com

# Setup

## env setup

```bash
cp example.env .env
```

then edit `.env` with required variables

## run with go

```bash
go run .
```

## run with air

```bash
air
```

if you do not have air, install with, `go install github.com/air-verse/air@latest`

## with docker compose

```bash
docker compose up --build
```

## with docker for test

```bash
docker build -t niyam:latest -f Dockerfile .
```

```bash
docker run -p 7860:7860 --env-file .env niyam:latest
```
