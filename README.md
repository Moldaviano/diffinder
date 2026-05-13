# Diffinder

Tracking del processo di rilascio: dev → cert → prod, con guardrail che verifica
che il commit HEAD di una PR verso produzione sia stato effettivamente certificato.

## Stack

- **Backend**: Go 1.22, Chi, pgx/v5, sqlc, JWT HS256, slog
- **Frontend**: Angular (ultima stabile), Standalone components, Angular Material, Signals
- **DB**: PostgreSQL 16
- **Migrations**: golang-migrate

## Quick start (Docker Compose)

```bash
cp .env.example .env
# editare JWT_SECRET e GITHUB_WEBHOOK_SECRET
docker compose up -d --build
# Migrations vengono applicate dal servizio `migrate`
# Backend:  http://localhost:8080
# Frontend: http://localhost:4200
```

Per popolare il DB con dati demo (3 progetti, 10 release, eventi, PR):

```bash
docker compose exec backend /app/seed
```

## Variabili d'ambiente

Vedi `.env.example`. Le più rilevanti:

| Variabile | Descrizione |
|-----------|-------------|
| `JWT_SECRET` | Chiave HS256 (≥16 char). `openssl rand -base64 48` |
| `JWT_ACCESS_TTL` | TTL access token (default Go `15m`, compose override `2h`) |
| `JWT_REFRESH_TTL` | TTL refresh token (es. `168h`) |
| `GITHUB_WEBHOOK_SECRET` | Secret HMAC per `X-Hub-Signature-256` |
| `CORS_ALLOWED_ORIGINS` | CSV di origini consentite |
| `LOG_LEVEL` / `LOG_FORMAT` | `debug\|info\|warn\|error` / `json\|text` |

## Flusso di cert check

1. Lo sviluppatore crea una **Release** su un branch e la sposta in **dev** registrando un `DeploymentEvent`.
2. Quando la release entra in **cert**, oltre al `DeploymentEvent` viene catturato un set di `CommitSnapshot` con tutti i commit presenti nel branch in quel momento.
3. Allo `commit_sha` del deploy in cert ci si riferisce come "**cert HEAD**" per quella release.
4. Quando si apre/aggiorna una PR verso uno dei branch tracciati (`master`, `main`, `test/dev`, `test/staging` — configurabili nel workflow), GitHub Actions invia un webhook `POST /api/webhooks/github/pr` firmato con `X-Hub-Signature-256`. Il payload contiene `repo`, `pr_number`, `head_sha`, `base_branch`, `head_branch`, `pr_url`.
5. Il backend cerca la release corrispondente al repo + branch, recupera il cert HEAD e verifica che `head_sha` sia **discendente o uguale** al cert HEAD (cioè che tutti i commit fra cert HEAD e head_sha siano stati testati). Tecnicamente: `head_sha` deve essere presente nello snapshot **oppure** il cert HEAD deve essere ancestor di `head_sha`. La semplificazione che adottiamo (poiché non abbiamo l'accesso git diretto) è: il check passa se `head_sha == cert_head_sha`. Diversamente l'esito è `passed=false` con dettagli.
6. Il risultato viene salvato in `CertificationCheck` e ritornato a GitHub Actions come `{ passed, reason }`. Se `passed=false` la PR viene marcata `blocked`.

> Nota implementativa: per check più sofisticati (ancestor check reale) si può integrare l'API GitHub `repos/:owner/:repo/compare/{base}...{head}` oppure mantenere lato server una lista ordinata di commit conosciuti in cert.

## Struttura

```
cmd/
  server/     # entrypoint API
  seed/       # popolatore dati demo
internal/
  handler/    # HTTP handlers (chi)
  service/    # logica di dominio
  repository/ # accesso DB (sqlc-generated)
  middleware/ # auth, logging, request id
  model/      # entità di dominio
  config/     # caricamento env
  logger/     # slog setup
  httpx/      # helpers JSON / errori / paginazione
  auth/       # JWT issue/verify
migrations/   # golang-migrate
frontend/     # Angular workspace
```

## Endpoints

Vedi specifica in [`docs/API.md`](docs/API.md) (TODO) — anteprima:

```
POST /api/auth/login
GET  /api/projects
POST /api/releases/:id/deploy
POST /api/webhooks/github/pr
GET  /api/dashboard/summary
```

## Sviluppo locale (senza Docker)

```bash
# 1. Avvia Postgres
docker compose up -d postgres
# 2. Applica migrations
migrate -path migrations -database "postgres://diffinder:diffinder@localhost:5432/diffinder?sslmode=disable" up
# 3. Esporta env
export $(grep -v '^#' .env | xargs)
# 4. Server
go run ./cmd/server
# 5. Frontend
cd frontend && npm install && npm start
```

## Webhook GitHub

Per attivare la notifica delle PR da una repo GitHub vedi [`docs/github-actions/README.md`](docs/github-actions/README.md). Sintesi:

- Copia `docs/github-actions/diffinder-notify.yml` in `.github/workflows/` **su tutti i branch che possono essere base di una PR** (es. `main`, `master`, `test/dev`, `test/staging`).
- Aggiungi i due secret repo `DIFFINDER_URL` e `DIFFINDER_WEBHOOK_SECRET` (quest'ultimo identico a `GITHUB_WEBHOOK_SECRET` del backend).
- Registra il progetto in Diffinder con `repository_url` esatto e crea la Release con `branch_name = head_branch` della PR.
- In dev locale esponi il backend con un tunnel (`cloudflared tunnel --url http://localhost:8080`) e usa l'URL come `DIFFINDER_URL`.
