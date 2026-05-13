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

I workflow versionati e il README operativo stanno in [`docs/github-actions/`](docs/github-actions/). Due varianti disponibili:

- **`diffinder-notify.yml`** — notify-only. Manda l'evento a Diffinder ma non blocca mai il PR (`continue-on-error: true`). **Usalo come primo step.**
- **`diffinder-cert-check.yml`** — fa fallire il PR se Diffinder risponde `passed=false`. Attivalo solo quando il sistema è stabile e i progetti/release sono configurati correttamente.

### 7.1 Prerequisito: esponi il backend pubblicamente

GitHub Actions gira sui runner cloud, non può raggiungere `localhost`. Serve un URL pubblico HTTPS.

**Sviluppo locale — tunnel rapido (Cloudflare Quick Tunnel, gratis, no account)**:

```bash
brew install cloudflared
cloudflared tunnel --url http://localhost:8080
```

Cloudflared stampa un URL `https://<random>.trycloudflare.com`. Verifica:

```bash
curl https://<...>.trycloudflare.com/healthz   # → {"status":"ok"}
```

Tieni il terminale aperto: chiudendolo il tunnel muore. L'URL **cambia ad ogni restart** (gratis): quando cambia aggiorna il secret `DIFFINDER_URL` su GitHub.

In alternativa: `ngrok http 8080` (richiede account gratuito, ma offre la dashboard live su `http://localhost:4040`).

**Produzione**: usa un dominio stabile (Cloudflare named tunnel, reverse proxy aziendale, VPS).

### 7.2 Per OGNI repo GitHub da tracciare

**Passo 1 — Secrets**: Settings → Secrets and variables → Actions → New repository secret

| Nome | Valore |
|------|--------|
| `DIFFINDER_URL` | URL pubblico del backend (es. `https://abc.trycloudflare.com` o `https://diffinder.tuodominio.it`) — senza slash finale |
| `DIFFINDER_WEBHOOK_SECRET` | **Identico** al `GITHUB_WEBHOOK_SECRET` del `.env` del backend |

Suggerimento: se le repo sono molte, usa un **Organization secret** (Org → Settings → Secrets → Actions) e abilita le repo che servono.

**Passo 2 — Workflow file su tutti i base branch**

⚠️ Regola GitHub fondamentale: per eventi `pull_request`, il workflow eseguito è quello presente **sul base branch della PR**, non sul branch di default. Quindi il file `.github/workflows/diffinder-notify.yml` deve esistere su **ogni branch** che può essere base di una PR.

Per il branch model CAME (`master`, `main`, `test/dev`, `test/staging`):

```bash
cd <repo-target>

# 1. Mettilo sul branch di default (es. master) come punto di verità
git checkout master
mkdir -p .github/workflows
cp /percorso/diffinder/docs/github-actions/diffinder-notify.yml .github/workflows/
git add .github/workflows/diffinder-notify.yml
git commit -m "ci: notifica Diffinder sulle PR"
git push origin master

# 2. Propaga lo stesso file su tutti gli altri base branch
for B in main test/dev test/staging; do
  git push origin master:$B  # se accettato come fast-forward
done

# Se il push fast-forward è rifiutato (branch con storie divergenti):
for B in main test/dev test/staging; do
  git checkout $B
  git pull
  git checkout master -- .github/workflows/diffinder-notify.yml
  git commit -m "ci: workflow Diffinder" && git push origin $B
done
```

Verifica che il file sia presente su ogni base branch:

```bash
REPO="<owner>/<repo>"
for B in master main test/dev test/staging; do
  curl -s -o /dev/null -w "$B → HTTP %{http_code}\n" \
    https://raw.githubusercontent.com/$REPO/$B/.github/workflows/diffinder-notify.yml
done
```

**Passo 3 — Configurazione lato Diffinder**

1. Crea/aggiorna il progetto in Diffinder con `repository_url = https://github.com/<owner>/<repo>` **esatto** (no trailing slash, no `.git`). Il match nel webhook è letterale: qualsiasi differenza → 404.
2. Per ogni branch da tracciare, crea una **Release** con `branch_name = head_branch` della PR (cioè il branch sorgente, non quello di destinazione). Senza una release per quel branch, il webhook risponde 404 "no release tracked for branch ...".

**Passo 4 — Trigger di prova**

```bash
git checkout master
git checkout -b feature/diffinder-smoke
echo "ping" >> README.md
git add README.md && git commit -m "test diffinder"
git push -u origin feature/diffinder-smoke
gh pr create --base master --title "Smoke Diffinder" --body "test"
```

Verifica in 4 punti:

1. Tab Actions della repo → run `diffinder-notify` con conclusion `success`
2. Dashboard tunnel (ngrok: `localhost:4040`) → richiesta in arrivo
3. Log backend: `docker compose logs -f backend | grep webhooks` → riga `path:"/api/webhooks/github/pr"`
4. UI Diffinder (`/pull-requests`) → la PR appare con stato (es. `blocked` se non c'è ancora un deploy cert)

### 7.3 Comportamento del workflow

- Eventi tracciati: `opened`, `synchronize`, `reopened`, `closed`, `ready_for_review`, `edited`
- Filtra le PR per base branch (default `master`, `main`, `test/dev`, `test/staging` — modificabile)
- `concurrency`: push consecutivi sulla stessa PR cancellano i run vecchi → niente webhook duplicati
- Su risposte non-200 dal backend logga un `::warning::` (notify-only) o un `::error::` (cert-check)

### 7.4 Testare il webhook senza GitHub (in locale)

Utile per debug isolato dal layer Actions/tunnel. Body e firma devono combaciare byte-per-byte:

```bash
SECRET="$(grep ^GITHUB_WEBHOOK_SECRET .env | cut -d= -f2-)"
BODY='{"repo":"https://github.com/alloy/payments-api","pr_number":42,"head_sha":"abc1234567890abcdef","base_branch":"master","head_branch":"feature/refund","pr_url":"https://x"}'
SIG=$(printf '%s' "$BODY" | openssl dgst -sha256 -hmac "$SECRET" -hex | awk '{print $2}')

curl -i -X POST http://localhost:8080/api/webhooks/github/pr \
  -H "Content-Type: application/json" \
  -H "X-Hub-Signature-256: sha256=$SIG" \
  -d "$BODY"
```

Esiti possibili (vedi anche `docs/github-actions/README.md` per la tabella completa):

| HTTP | Significato |
|------|-------------|
| `200 {"passed":false,"reason":"no cert deployment..."}` | Tutto OK end-to-end (il `passed:false` è atteso senza un deploy cert) |
| `401 invalid signature` | `SECRET` lato curl ≠ `GITHUB_WEBHOOK_SECRET` nel backend |
| `404 project not registered for repo` | `repository_url` non combacia con un progetto in Diffinder |
| `404 no release tracked for branch` | Manca la release con quel `branch_name` |

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
| `JWT_ACCESS_TTL` | `15m` (override compose: `2h`) | no | Durata access token |
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
