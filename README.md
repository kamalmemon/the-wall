# The Wall

A minimalist guestbook where visitors pick a color and leave a message. Each entry becomes a tile in a colorful mosaic.

![Screenshot](screenshot.png)

## Features

- **Visitor counter** — Each visitor gets a unique number
- **Color picker** — Drag to select from a gradient
- **Hero mosaic** — The visitor count tile shows colors picked by visitors
- **Mobile responsive**

## Stack

- Go (stdlib `net/http`)
- SQLite
- Vanilla HTML/CSS/JS
- JetBrains Mono + Inter fonts

## Run locally

```bash
go run .
```

Then open http://localhost:8000

## Environment variables

- `PORT` — Server port (default: 8000)
- `DB_PATH` — SQLite database path (default: wall.db)

## Deploy

Build and run:

```bash
go build -o the-wall .
./the-wall
```

Or use the provided systemd service file:

```bash
sudo cp the-wall.service /etc/systemd/system/
sudo systemctl enable the-wall
sudo systemctl start the-wall
```

## Structure

```
the-wall/
├── main.go              # Server + handlers
├── db/migrations/       # SQL migrations
├── static/              # CSS + JS
├── templates/           # HTML templates
├── go.mod
└── README.md
```

## License

MIT

---

Vibecoded with Claude Opus 4.5 on [exe.dev](https://exe.dev)
