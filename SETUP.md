# Diffinder — Setup & Avvio

Guida operativa per portare l'applicazione su un ambiente di sviluppo locale, eseguire i test del flusso e configurare il webhook GitHub.

## Prerequisiti

| Strumento | Versione minima | Note |
|-----------|-----------------|------|
| Docker Engine | 24+ | con Compose v2 incluso (`docker compose`) |
| Go | 1.22 | solo se vuoi eseguire il backend fuori da Docker |
| Node.js | 20+ | solo per `ng serve` in dev |
| Angular CLI | 18+ | `npm i -g @angular/cli` (facoltativo, c'è già `npx`) |
| `openssl` | qualunque | per generare `JWT_SECRET` / `GITHUB_WEBHOOK_SECRET` |
| `golang-migrate` | v4.17+ | solo se vuoi applicare le migrations a mano |

## 1. Clona e prepara le variabili d'ambiente

```bash
git clone <repo-url> diffinder
cd diffinder
cp .env.example .env
```

Apri `.env` e **modifica almeno questi due valori**:

```bash
# Genera un secret robusto:
JWT_SECRET=$(openssl rand -base64 48)

# Genera un secret per il webhook (lo stesso che metterai nel GitHub Action):
GITHUB_WEBHOOK_SECRET=$(openssl rand -hex 32)
```

Tutte le altre variabili hanno default sensati per Docker. La descrizione completa è nella sezione "Variabili d'ambiente" in fondo a questo file.

## 2. Avvio rapido con Docker Compose (raccomandato)

```bash
docker compose up -d --build
```

L'orchestrazione fa partire **4 servizi**:

1. **postgres** — DB con healthcheck. Volume persistente `diffinder-pg`.
2. **migrate** — applica le migrations e termina. Dipende dal healthcheck di postgres.
3. **backend** — il server Go (`:8080`). Parte solo dopo che `migrate` ha completato.
4. **frontend** — nginx che serve l'Angular buildato e proxiia `/api` verso `backend:8080` (`:4200`).

Verifica lo stato:

```bash
docker compose ps
docker compose logs -f backend
```

Test di salute:

```bash
curl http://localhost:8080/healthz
# → {"status":"ok"}
```

## 3. Popola il database con dati demo (seed)

```bash
docker compose exec backend /app/seed
```

Cosa crea:

- **4 utenti**: `admin` (admin), `alice` e `bob` (developer), `viewer` (viewer)
- **3 progetti**: `payments-api`, `web-dashboard`, `notifications`
- **10 release** in stati diversi (`draft`, `in_dev`, `in_cert`, `approved`, `in_prod`, `rejected`)
- Eventi di deploy coerenti con lo stato di ciascuna release
- Snapshot dei commit per le release passate per `cert`
- PR con esiti cert-check misti (alcune passate, alcune **fallite** → status `blocked`)

Il seed è **idempotente**: se rilanciato non duplica i dati.

## 4. Login e prima interazione

Apri il frontend: <http://localhost:4200>

Credenziali demo (definite in `.env`):

| Username | Email | Password | Ruolo |
|----------|-------|----------|-------|
| admin | `admin@diffinder.local` | `admin123` | admin |
| alice | `alice@diffinder.local` | `alice123` | developer |
| bob | `bob@diffinder.local` | `bob123` | developer |
| viewer | `viewer@diffinder.local` | `viewer123` | viewer |

> Cambia `SEED_ADMIN_PASSWORD` in `.env` prima di rilanciare il seed in produzione.

Le viste disponibili dopo il login:

- **Overview** (`/dashboard`) — card metriche + lista attività con polling 30s
- **Release** (`/releases`) — tabella filtrabile, semaforo dev/cert/prod
- **Pull Requests** (`/pull-requests`) — colonna cert-check, toggle "solo bloccate"
- **Progetti** (`/projects`) — CRUD con statistiche per progetto
- **Impostazioni** (`/settings`) — solo admin: utenti + webhook token

## 5. Sviluppo locale senza Docker

Utile quando vuoi un ciclo di compilazione/restart veloce su backend o frontend.

### Solo Postgres in Docker, backend a mano

```bash
docker compose up -d postgres

# Applica le migrations
migrate -path migrations \
  -database "postgres://diffinder:diffinder@localhost:5432/diffinder?sslmode=disable" up

# Esporta le env del .env nella shell corrente
set -a; source .env; set +a
# DB_HOST nel .env punta a "postgres" (nome container): in modalità locale override:
export DB_HOST=localhost

# Avvia il server
go run ./cmd/server

# Popola
go run ./cmd/seed
```

### Frontend in dev mode

```bash
cd frontend
npm install
npm start   # ng serve --host 0.0.0.0 --port 4200
```

Il `proxy.conf.json` di Angular CLI fa il forward di `/api/*` su `http://localhost:8080`, quindi non hai problemi di CORS in dev.

Apri <http://localhost:4200>.

## 6. Migrations

Sono in `migrations/`, numerate con prefisso a 4 cifre, formato `golang-migrate`.

### Applicare in avanti

```bash
migrate -path migrations \
  -database "postgres://diffinder:diffinder@localhost:5432/diffinder?sslmode=disable" up
```

In Docker viene fatto automaticamente dal servizio `migrate` allo startup.

### Rollback dell'ultima migration

```bash
migrate -path migrations -database "..." down 1
```

### Aggiungere una nuova migration

```bash
migrate create -ext sql -dir migrations -seq nome_modifica
# crea: migrations/0002_nome_modifica.{up,down}.sql
```

## 7. Configurare il webhook GitHub

Sul **repository GitHub** del progetto che vuoi tracciare:

1. Settings → Secrets and variables → Actions → New repository secret
   - `DIFFINDER_WEBHOOK_SECRET` = stesso valore di `GITHUB_WEBHOOK_SECRET` nel `.env`
   - `DIFFINDER_URL` = URL pubblico del backend (es. `https://diffinder.tuodominio.it`)
2. Copia il workflow di esempio:

```bash
mkdir -p .github/workflows
cp /percorso/diffinder/docs/github-actions-example.yml \
   .github/workflows/diffinder-cert-check.yml
```

3. Crea/aggiorna il progetto su Diffinder facendo combaciare `repository_url` con `https://github.com/<owner>/<repo>` (senza trailing slash). Questo è il campo usato dal webhook per risalire al progetto.
4. Crea su Diffinder una **release** con `branch_name` uguale all'`head_branch` della PR (es. `feature/new-checkout`).

Al primo `pull_request` aperto verso `main`, il workflow:

- compila il payload JSON `{repo, pr_number, head_sha, base_branch, head_branch, pr_url}`
- calcola `X-Hub-Signature-256: sha256=<HMAC-SHA256>` sul body
- chiama `POST /api/webhooks/github/pr`
- legge `{ passed, reason }` e fa fallire la pipeline se `passed=false`

### Testare il webhook in locale

```bash
PAYLOAD='{"repo":"https://github.com/alloy/payments-api","pr_number":42,"head_sha":"abc123","base_branch":"main","head_branch":"feature/refund-api","pr_url":"https://x"}'
SECRET="$(grep ^GITHUB_WEBHOOK_SECRET .env | cut -d= -f2)"
SIG="sha256=$(echo -n "$PAYLOAD" | openssl dgst -sha256 -hmac "$SECRET" -hex | awk '{print $2}')"

curl -i -X POST http://localhost:8080/api/webhooks/github/pr \
  -H "Content-Type: application/json" \
  -H "X-Hub-Signature-256: $SIG" \
  -d "$PAYLOAD"
```

Risposta attesa: `{"passed":false,"reason":"head ... not present in cert snapshot ..."}` (perché `abc123` non è il SHA che c'è in cert).

## 8. Build di produzione

### Backend

```bash
CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" \
  -o bin/server ./cmd/server
CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" \
  -o bin/seed ./cmd/seed
```

L'immagine prodotta dal `Dockerfile` è multi-stage, parte da `alpine:3.20`, gira come utente non-root.

### Frontend

```bash
cd frontend
npm ci
npx ng build --configuration=production
# output: dist/diffinder/browser
```

Servila con qualunque static server o usa l'immagine nginx generata dal `frontend/Dockerfile`.

## 9. Operatività

### Tail dei log

```bash
docker compose logs -f backend
```

Formato JSON strutturato (`slog`). Campi rilevanti per ogni richiesta HTTP:

```json
{"time":"2026-05-08T10:23:00Z","level":"INFO","msg":"http",
 "method":"POST","path":"/api/releases/.../deploy","status":201,
 "bytes":312,"duration_ms":18,"request_id":"abc-..."}
```

### Backup database

```bash
docker compose exec postgres pg_dump -U diffinder diffinder > backup.sql
```

### Restore

```bash
docker compose exec -T postgres psql -U diffinder diffinder < backup.sql
```

### Reset totale (sviluppo)

```bash
docker compose down -v   # rimuove anche il volume
docker compose up -d --build
docker compose exec backend /app/seed
```

## 10. Variabili d'ambiente — riferimento completo

| Variabile | Default | Obbligatoria | Descrizione |
|-----------|---------|--------------|-------------|
| `SERVER_HOST` | `0.0.0.0` | no | Bind address del server |
| `SERVER_PORT` | `8080` | no | Porta HTTP |
| `SERVER_READ_TIMEOUT` | `15s` | no | Timeout lettura richiesta |
| `SERVER_WRITE_TIMEOUT` | `15s` | no | Timeout scrittura risposta |
| `DB_HOST` | `postgres` | no | Hostname Postgres |
| `DB_PORT` | `5432` | no | Porta Postgres |
| `DB_NAME` | `diffinder` | no | Nome database |
| `DB_USER` | `diffinder` | no | Utente |
| `DB_PASSWORD` | `diffinder` | no | Password |
| `DB_SSLMODE` | `disable` | no | `disable\|require\|verify-full` |
| `DB_MAX_CONNS` | `10` | no | Pool size massimo |
| `JWT_SECRET` | — | **sì** | HS256, ≥16 caratteri |
| `JWT_ACCESS_TTL` | `15m` | no | Durata access token |
| `JWT_REFRESH_TTL` | `168h` | no | Durata refresh token (7gg) |
| `GITHUB_WEBHOOK_SECRET` | — | **sì** | Secret HMAC condiviso |
| `CORS_ALLOWED_ORIGINS` | `http://localhost:4200` | no | CSV di origini consentite |
| `LOG_LEVEL` | `info` | no | `debug\|info\|warn\|error` |
| `LOG_FORMAT` | `json` | no | `json\|text` |
| `SEED_ADMIN_EMAIL` | `admin@diffinder.local` | no | Email admin del seed |
| `SEED_ADMIN_PASSWORD` | `admin123` | no | Password admin del seed |

> Se una variabile obbligatoria manca, il backend va in `panic` allo startup con messaggio esplicito (`missing required env var: ...`).

## 11. Troubleshooting

### Il backend non si avvia: `missing required env var: JWT_SECRET`
Devi creare `.env` partendo da `.env.example` e generare un secret. Vedi punto 1.

### Il container `migrate` esce con errore "dirty database"
Una migration precedente è fallita a metà. Forza la versione:
```bash
docker compose run --rm migrate -path /migrations \
  -database "postgres://diffinder:diffinder@postgres:5432/diffinder?sslmode=disable" \
  force <version>
```

### Il frontend mostra "Errore sconosciuto" su tutte le chiamate
Probabilmente il backend non risponde. Verifica:
```bash
curl http://localhost:8080/healthz
docker compose logs backend | tail -50
```

### Webhook risponde 401 "invalid signature"
Il `GITHUB_WEBHOOK_SECRET` del backend non coincide con quello usato per firmare. Controlla che siano identici (no spazi, no newline). Per riavviare con il nuovo valore: `docker compose up -d --force-recreate backend`.

### Il cert-check ritorna sempre `passed=false`
- La release esiste su Diffinder ma non c'è stato ancora alcun deploy in `cert`? → registra un deploy in cert.
- Il `head_sha` della PR è diverso dal SHA registrato in cert e non è nello snapshot. Vedi `DOCUMENTATION.md` sezione "Logica del cert-check" per i criteri.

### Le release non si associano alla PR
La risoluzione avviene per (`repository_url` del progetto, `head_branch` della PR). Assicurati che:
- il `repository_url` del progetto su Diffinder coincida con `repo` inviato dal webhook
- esiste una release per quel progetto con `branch_name` = `head_branch`
