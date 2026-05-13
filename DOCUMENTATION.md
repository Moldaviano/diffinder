# Diffinder — Documentazione

> Diffinder traccia il ciclo di vita di una release software attraverso gli ambienti **dev → cert → prod** e blocca le Pull Request verso produzione che contengono commit non certificati.

Per la guida di avvio operativa vedi [`SETUP.md`](SETUP.md).

---

## 1. Cosa risolve

In molti team il rilascio in produzione passa per un percorso del tipo:

1. lo sviluppatore lavora su un feature branch
2. lo testa in locale
3. lo porta in ambiente **dev** (a basso costo, dati sporchi)
4. lo porta in ambiente **cert** (dati controllati, test approfonditi)
5. dopo approvazione lo porta in **prod**

Il **rischio specifico** che Diffinder mitiga: tra il punto 4 e il punto 5, qualcuno potrebbe aggiungere altri commit al branch e poi aprire una PR verso produzione contenente codice che *non è mai stato certificato*. Questi commit possono passare inosservati nelle review umane.

Diffinder fornisce:

- una **fonte di verità persistente** di quali SHA sono passati per cert per ciascuna release
- un **webhook** chiamato da GitHub Actions all'apertura di una PR verso `main` che esegue il "cert check" e blocca la pipeline se fallisce
- una **dashboard** per visualizzare in tempo reale lo stato di release, deploy e PR bloccate
- uno **storico** di tutti i deploy, snapshot di commit, e tutti i check eseguiti

---

## 2. Architettura ad alto livello

```
┌──────────────┐      HTTPS       ┌────────────────┐      SQL       ┌─────────────┐
│  Angular SPA │ ───────────────► │   Go backend   │ ──────────────►│ PostgreSQL  │
│  (nginx)     │ ◄─────────────── │   (Chi + pgx)  │ ◄──────────────│             │
└──────────────┘    JSON/JWT      └────────────────┘                └─────────────┘
        ▲                                  ▲
        │                                  │ HMAC-SHA256
        │                                  │
        │                          ┌───────┴────────┐
        │                          │ GitHub Actions │  (workflow su PR)
        │                          └────────────────┘
        │
        └── operatori umani via browser
```

### Componenti

| Componente | Tech | Ruolo |
|------------|------|-------|
| **Frontend** | Angular 18 standalone + Material + Signals | UI, autenticazione utente, dashboard, CRUD |
| **Backend** | Go 1.22, Chi, pgx, slog | API REST `/api/*` + webhook `/api/webhooks/github/pr` |
| **DB** | PostgreSQL 16 | Persistenza |
| **Migrations** | golang-migrate | Schema evolution |
| **Auth** | JWT HS256 access + refresh | Token Bearer su tutte le route protette |
| **Webhook auth** | HMAC-SHA256 in tempo costante | `X-Hub-Signature-256` |
| **Container** | Docker Compose | Stack riproducibile |

### Architettura del backend

Tre layer rigidi, niente ORM "intelligenti":

```
HTTP request
    │
    ▼
┌────────────┐  json decoding, validazione di forma,
│  handler   │  estrazione path/query params, response writing
└─────┬──────┘
      │ ctx + DTOs
      ▼
┌────────────┐  logica di dominio, regole di business,
│  service   │  errori semantici (ErrNotFound, ErrBadRequest…)
└─────┬──────┘
      │ entità di dominio
      ▼
┌────────────┐  query SQL parametriche con pgx/v5,
│ repository │  mapping rows → struct
└─────┬──────┘
      │
      ▼
   PostgreSQL
```

Regole applicate:

- **Handler non parlano col DB.** Sempre via service.
- **Service non costruisce HTTP responses.** Ritorna entità o `*httpx.AppError` con codice semantico.
- **Repository non sa di HTTP.** Ritorna `pgx.ErrNoRows`, il service lo traduce in `ErrNotFound`.
- **Modelli di dominio** in `internal/model/` non dipendono né da HTTP né da DB.

---

## 3. Dominio: cosa rappresenta cosa

### Entità

```
User ─────┐
          │ creato_da
          ▼
        Release ◄── Project
          │
          ├── DeploymentEvent (N, ordinati nel tempo)
          ├── CommitSnapshot  (N, set di SHA presenti al deploy in cert)
          └── PullRequest (N)
                  │
                  └── CertificationCheck (N, storico)
```

| Entità | Cardinalità | Scopo |
|--------|-------------|-------|
| **Project** | N | Un'applicazione/repo. Identificato univocamente da `name` e da `repository_url`. |
| **Release** | 1:N con Project | Una sequenza di commit (= un branch) tracciata per quel progetto. Stato di alto livello (`draft`, `in_dev`, `in_cert`, `approved`, `in_prod`, `rejected`). |
| **DeploymentEvent** | 1:N con Release | Ogni volta che la release viene portata in `dev`, `cert` o `prod` con un dato `commit_sha`. **Append-only**, mai modificato. |
| **CommitSnapshot** | 1:N con Release | Set di SHA noti al momento del deploy in cert. È la "fotografia" usata per certificare quali commit sono stati effettivamente testati. |
| **PullRequest** | 1:N con Release | PR verso produzione, registrata dal webhook o manualmente. Unique key: (`release_id`, `pr_number`). |
| **CertificationCheck** | 1:N con PR | Storico dei check (passed/failed + dettagli + SHA confrontati). |
| **User** | N | account, ruoli `admin\|developer\|viewer`. Auth via bcrypt + JWT. |

### Stati di una Release

```
       ┌─────────┐
       │  draft  │
       └────┬────┘
            │  deploy dev
            ▼
       ┌─────────┐
       │ in_dev  │
       └────┬────┘
            │  deploy cert
            ▼
       ┌─────────┐      reject
       │ in_cert │──────────────► rejected
       └────┬────┘
            │  approvazione manuale
            ▼
       ┌──────────┐
       │ approved │
       └────┬─────┘
            │  deploy prod
            ▼
       ┌─────────┐
       │ in_prod │
       └─────────┘
```

I deploy events **non resettano lo stato all'indietro**: una release `in_prod` resta `in_prod` anche se viene ri-deployata in dev. Lo stato segue il punto più avanzato raggiunto, salvo `rejected` che è terminale.

---

## 4. La logica del cert-check (cuore del sistema)

Quando GitHub Actions chiama il webhook, il backend:

1. **Verifica firma HMAC-SHA256** sull'header `X-Hub-Signature-256` (confronto in tempo costante con `hmac.Equal`). Firma non valida → `401`.
2. Risale al `Project` dal `repository_url` inviato nel payload.
3. Risale alla `Release` con (`project_id`, `head_branch`). Se non esiste → `404`.
4. **Upsert della PullRequest** su `(release_id, pr_number)`: aggiorna SHA e branch base se la PR è già presente. Questo rende il webhook idempotente: lo puoi rilanciare a ogni push.
5. Recupera l'ultimo `DeploymentEvent` con `environment=cert` per quella release. Il suo `commit_sha` è il **cert HEAD**.
6. Applica le regole nell'ordine seguente:

| Condizione | Esito | Note |
|------------|-------|------|
| Nessun deploy in cert per la release | `passed=false`, reason "no cert deployment" | Niente da certificare |
| `head_sha == cert_head_sha` | `passed=true` | Il commit della PR è esattamente quello certificato |
| `head_sha` è presente in `commit_snapshots` della release | `passed=true` | È un commit che faceva parte del set noto al momento del deploy in cert |
| Altrimenti | `passed=false`, reason esplicita | Commit aggiunto dopo cert, mai testato |

7. Salva una nuova riga in `certification_checks` (mai modifica, mai elimina — è uno storico).
8. Se `passed=false` e la PR non è in stato terminale (`merged`/`closed`), aggiorna lo status PR a `blocked`.
9. Risponde a GitHub Actions con `{ passed: bool, reason: string }`.

### Perché lo snapshot è importante

Quando una release passa per cert, registriamo l'elenco SHA presenti nel branch in quel momento (lo passa il client al momento del `POST /api/releases/:id/deploy` con `environment=cert`). Questo permette di accettare automaticamente come "certificati" anche commit più recenti del `cert HEAD`, purché fossero già noti al momento del test in cert (ad esempio: deploy in cert al commit C5, ma il branch conteneva già C6 e C7 che erano stati testati come parte del test in cert — il check passa anche su C6/C7).

### Limitazione attuale

Diffinder **non** chiama l'API GitHub per verificare ancestry git lato server. Se il `head_sha` della PR è un commit *successivo* al cert HEAD e *non* è nello snapshot, viene rifiutato anche se la verità git è che è un fast-forward innocuo (ad esempio un merge di `main` nel feature branch). Per un ancestor-check git-aware è possibile integrare:

- chiamata a `GET /repos/:owner/:repo/compare/{base}...{head}` (richiede token GitHub server-side)
- oppure spingere la lista commit completa dal workflow nel payload `head_commits[]`

Entrambe sono estensioni naturali ma non implementate in questa versione.

---

## 5. Cosa include il progetto — inventario

### Backend Go

```
cmd/
├── server/main.go            entrypoint API: config, DB, DI graph, http.Server
└── seed/main.go              popola DB con 4 utenti, 3 progetti, 10 release + eventi/PR/check

internal/
├── config/config.go          loader env tipizzato, fail-fast su required
├── logger/logger.go          slog JSON|text + livelli configurabili
├── httpx/
│   ├── errors.go             *AppError + helper WriteJSON / WriteError
│   └── pagination.go         ParsePage(r) + PagedResponse[T] generico
├── model/                    entità di dominio (user, project, release, deployment,
│                             commit_snapshot, pull_request, certification_check)
├── auth/
│   ├── jwt.go                Issuer HS256, access+refresh, claims tipizzate
│   └── password.go           bcrypt hash/check
├── middleware/
│   ├── auth.go               Bearer middleware + Principal in ctx + RequireRole
│   └── logging.go            slog per request con request_id e duration
├── repository/               pgx-based, una query per metodo
│   ├── db.go                 pgxpool wrapper + IsNotFound helper
│   ├── users_repo.go
│   ├── projects_repo.go      include Stats per progetto
│   ├── releases_repo.go      filtri opzionali project_id + status
│   ├── deployments_repo.go   include LatestCertDeployment e RecentActivity
│   ├── commits_repo.go       bulk upsert idempotente snapshot
│   ├── pull_requests_repo.go upsert su (release_id, pr_number), ListBlocked
│   ├── checks_repo.go        storico cert-check + LastByPR
│   └── dashboard_repo.go     summary + releases-by-status
├── service/                  logica di dominio, errori tipizzati
│   ├── auth_service.go       login + refresh
│   ├── project_service.go    CRUD + generazione webhook_token
│   ├── release_service.go    aggregato release + sub-listing
│   ├── deployment_service.go register deploy + transizione stato + snapshot commit
│   ├── pr_service.go         CRUD PR
│   ├── check_service.go      *** logica del cert-check ***
│   ├── dashboard_service.go  aggregazioni
│   └── webhook_service.go    repo→project→release, upsert PR, invoca check
└── handler/                  HTTP, una sola responsabilità: I/O
    ├── helpers.go            pathUUID, decodeJSON, principalUserUUID
    ├── auth_handler.go
    ├── projects_handler.go
    ├── releases_handler.go
    ├── pull_requests_handler.go
    ├── dashboard_handler.go
    ├── users_handler.go
    ├── webhook_handler.go    *** verifica HMAC-SHA256 ***
    └── router.go             composizione chi + DI + middleware stack

migrations/
├── 0001_init.up.sql          7 tabelle, 4 enum types, trigger updated_at
└── 0001_init.down.sql

docs/
└── github-actions-example.yml  workflow di riferimento
```

### Frontend Angular

```
src/
├── index.html                root + font + Material icons
├── main.ts                   bootstrapApplication standalone
├── styles.scss               classi SCSS shared, palette badge/semafori
└── app/
    ├── app.component.ts      shell con <router-outlet>
    ├── app.config.ts         provider router + http(interceptors) + animations
    ├── app.routes.ts         lazy loading per ogni feature, guard auth/admin
    ├── layout/
    │   └── shell.component.ts    sidenav + toolbar + menu utente
    ├── core/
    │   ├── environment.ts        apiBase = '/api'
    │   ├── models/index.ts       mirror TS dei tipi backend
    │   ├── services/             auth, projects, releases, pull-requests,
    │   │                         dashboard, users, notification (snackbar)
    │   ├── interceptors/         auth (Bearer) + error (snackbar + 401→logout)
    │   └── guards/               authGuard + adminGuard
    ├── shared/components/
    │   ├── status-badge.component.ts        badge colorato per stato
    │   └── env-traffic-light.component.ts   semaforo dev/cert/prod (computed signals)
    └── features/
        ├── auth/login.component.ts           reactive form
        ├── dashboard/dashboard.component.ts  4 card + lista attività + polling 30s
        ├── releases/
        │   ├── releases-list.component.ts    tabella filtrabile + semaforo
        │   ├── release-detail.component.ts   timeline + commit + PR
        │   └── deploy-dialog.component.ts    dialog "registra deploy"
        ├── pull-requests/pull-requests.component.ts  + cert-check column + toggle
        ├── projects/
        │   ├── projects.component.ts                CRUD + stats per riga
        │   └── project-form-dialog.component.ts
        └── settings/settings.component.ts            tab utenti + tab webhook token
```

### Container & infra

| File | Scopo |
|------|-------|
| `Dockerfile` | Multi-stage build del backend (alpine, non-root) |
| `frontend/Dockerfile` | Build Angular + nginx con proxy `/api` |
| `frontend/nginx.conf` | SPA fallback + reverse proxy `/api/*` → `backend:8080` |
| `frontend/proxy.conf.json` | Proxy dev per `ng serve` |
| `docker-compose.yml` | 4 servizi: postgres, migrate, backend, frontend; healthcheck e dependency conditions |
| `.env.example` | Template variabili d'ambiente |
| `.gitignore` | esclusioni standard Go + Angular |

---

## 6. API REST

Base: `/api`. Tutte le route non-auth e non-webhook richiedono `Authorization: Bearer <access_token>`.

### Autenticazione

| Metodo | Endpoint | Body | Risposta |
|--------|----------|------|----------|
| `POST` | `/api/auth/login` | `{email, password}` | `{access_token, refresh_token, user}` |
| `POST` | `/api/auth/refresh` | `{refresh_token}` | `{access_token, refresh_token, user}` |
| `POST` | `/api/auth/logout` | — | `{status:"logged_out"}` (no-op server-side) |

### Projects

| Metodo | Endpoint | Note |
|--------|----------|------|
| `GET`  | `/api/projects` | `?page=&limit=` |
| `POST` | `/api/projects` | crea, `webhook_token` auto-generato se non passato |
| `GET`  | `/api/projects/:id` | |
| `PUT`  | `/api/projects/:id` | |
| `DELETE` | `/api/projects/:id` | cascade su release/deploy/PR |
| `GET`  | `/api/projects/:id/releases` | paginato |
| `GET`  | `/api/projects/:id/stats` | `{active_releases, last_activity}` |

### Releases

| Metodo | Endpoint | Note |
|--------|----------|------|
| `GET`  | `/api/releases` | filtri `?project_id=&status=` |
| `POST` | `/api/releases` | `created_by` auto da JWT |
| `GET`  | `/api/releases/:id` | |
| `PUT`  | `/api/releases/:id` | |
| `GET`  | `/api/releases/:id/deployments` | |
| `GET`  | `/api/releases/:id/pull-requests` | |
| `GET`  | `/api/releases/:id/commits` | snapshot commit |
| `POST` | `/api/releases/:id/deploy` | body `{environment, commit_sha, notes, commits[]}` |

### Pull Requests

| Metodo | Endpoint | Note |
|--------|----------|------|
| `POST` | `/api/pull-requests` | registrazione manuale (alternativa al webhook) |
| `GET`  | `/api/pull-requests` | `?blocked=true` per filtrare bloccate |
| `GET`  | `/api/pull-requests/:id` | |
| `PUT`  | `/api/pull-requests/:id/status` | body `{status}` |
| `POST` | `/api/pull-requests/:id/check-cert` | esegue e salva check, ritorna risultato |
| `GET`  | `/api/pull-requests/:id/checks` | storico cronologico |

### Dashboard

| Metodo | Endpoint | Risposta |
|--------|----------|----------|
| `GET` | `/api/dashboard/summary` | `{total_releases, in_cert, blocked_prs, deployments_today}` |
| `GET` | `/api/dashboard/releases-by-status` | `[{status, count}, ...]` |
| `GET` | `/api/dashboard/recent-activity` | `?limit=` lista eventi |
| `GET` | `/api/dashboard/blocked-prs` | paginato |

### Webhook

| Metodo | Endpoint | Auth | Note |
|--------|----------|------|------|
| `POST` | `/api/webhooks/github/pr` | HMAC-SHA256 via `X-Hub-Signature-256` | Body: `{repo, pr_number, head_sha, base_branch, head_branch?, pr_url?}` |

### Users (solo admin)

| Metodo | Endpoint | Note |
|--------|----------|------|
| `GET`  | `/api/users` | paginato |
| `POST` | `/api/users` | `{username, email, password, role}` |

### Formato errore uniforme

Tutti gli endpoint che falliscono ritornano:

```json
{
  "error": "messaggio umano",
  "code": "BAD_REQUEST|UNAUTHORIZED|FORBIDDEN|NOT_FOUND|CONFLICT|INTERNAL"
}
```

Lo `HttpStatus` corrisponde al code (400, 401, 403, 404, 409, 500). Il frontend tramite l'`errorInterceptor` mostra automaticamente il messaggio in uno snackbar Material; sul `401` fa logout e redirect a `/login`.

### Paginazione

Tutte le liste accettano `?page=N&limit=M` (default `page=1, limit=20`, `limit` clampato a 200). Risposta:

```json
{ "items": [...], "page": 1, "limit": 20, "total": 137 }
```

---

## 7. Modello dati relazionale

Schema completo in `migrations/0001_init.up.sql`. Punti chiave:

- **UUID** ovunque come PK (`gen_random_uuid()` da `pgcrypto`).
- **Enum types** Postgres nativi: `user_role`, `release_status`, `environment`, `pr_status`. Non `VARCHAR + CHECK`.
- **Indici** su tutte le FK + sui campi di filtro frequente: `releases.status`, `pull_requests.status`, `deployment_events.environment`, timestamp `DESC` per liste cronologiche.
- **Vincoli UNIQUE**:
  - `(release_id, branch_name)` per evitare release duplicate
  - `(release_id, commit_sha)` su snapshot per upsert idempotente
  - `(release_id, pr_number)` su PR per upsert dal webhook
- **Cascade**: cancellare un `project` cancella tutto a valle. È intenzionale — niente progetti orfani.
- **Trigger `set_updated_at()`** su `projects` e `releases`: aggiorna `updated_at` automaticamente su `UPDATE`.

---

## 8. Sicurezza

| Vettore | Mitigazione |
|---------|-------------|
| **SQL injection** | Tutte le query usano parametri pgx posizionali. Nessuna concatenazione. |
| **Brute force login** | bcrypt cost di default. (Rate-limit non implementato — eventuale aggiunta in middleware.) |
| **Token theft** | Access TTL corto (15m default). Refresh separato con TTL più lungo (7gg). Tipo del token discriminato nelle claims (`access` vs `refresh`). |
| **Token forging** | HS256 con secret server-side non esposto. JWT firmati e verificati con `jwt/v5`. Algorithm pinning (`SigningMethodHMAC` only). |
| **Cross-site requests** | CORS chiuso: solo origini in `CORS_ALLOWED_ORIGINS`. |
| **Webhook spoofing** | HMAC-SHA256 con `hmac.Equal` (constant-time). Body letto come bytes prima del parsing JSON così la firma include l'intero payload. |
| **Privilege escalation** | `RequireRole(admin)` middleware su tutto `/api/users/*`. Frontend `adminGuard` nasconde l'UI ma il backend è autoritativo. |
| **Password leak in log/JSON** | `User.PasswordHash` ha tag `json:"-"`. Mai esposto. |
| **CSRF** | Non vulnerabile: auth via header `Authorization`, non via cookie. |
| **XSS** | Angular escapa di default. Nessun `[innerHTML]` su contenuto non sanificato. |
| **Stack trace leak** | `Recoverer` middleware risponde 500 generico in produzione. |

### Quando il check fallisce

Il backend **non** rivela informazioni sensibili nel `reason`: solo "head not present in cert snapshot", non i SHA di altri progetti. Lo stesso `webhook_service` non rivela se un progetto/branch esiste tramite confronti di tempistica (potenziale miglioramento futuro).

---

## 9. Decisioni architetturali e tradeoff

### Chi invece di Gin
Chi è più idiomatico Go, middleware compatibili con la stdlib `net/http`, routing con sotto-router semplice. Per un'API REST classica è esattamente lo strumento giusto.

### pgx diretto invece di sqlc/GORM
- **GORM**: nasconde troppo, magia su transazioni e relazioni, prestazioni meno predicibili. Non vogliamo "lazy loading" su questo dominio.
- **sqlc**: ottimo, ma richiede toolchain di code-generation + file generati committati. Eccessivo per un progetto di queste dimensioni.
- **pgx diretto**: query SQL esplicite, ogni metodo del repository è una query parametrica chiara, type-safety via `Scan(&field)`. Niente cose magiche.

### Signals invece di NgRx
NgRx richiede actions/reducers/effects/selectors per ogni feature. Per una dashboard di queste dimensioni è overkill: i signal e i servizi stateful (auth, notification) coprono tutti i casi. RxJS rimane per HTTP. Niente boilerplate inutile.

### Standalone components ovunque
Angular 18 raccomanda standalone, niente NgModules. Lazy loading più semplice (un solo file per route). Compilazione più veloce.

### Token in localStorage
Pragmatica: niente cookie → niente CSRF, integrazione triviale con interceptors. Il rischio XSS è gestito da Angular default. Se in futuro serve isolamento maggiore (es. token in httpOnly cookie con CSRF token separato), il refactor è limitato all'`AuthService` + interceptor.

### Auto-blocco PR
Quando il check fallisce, il backend **mette automaticamente la PR in stato `blocked`**. Trade-off: l'operatore può sempre re-aprire (passando a `open` via `PUT /pull-requests/:id/status`), ma di default segnaliamo subito che c'è un problema da risolvere.

### Append-only su deployment events
Una volta registrato, un `DeploymentEvent` non si modifica. Se l'utente sbaglia, registra un nuovo evento con `notes` esplicative. Questo preserva l'audit trail.

---

## 10. Test e qualità

Stato attuale: **non sono presenti test unitari/integration** nel codice consegnato. Le aree più sensibili dove introdurre test per prime:

| Cosa testare | Tipo | Dove |
|--------------|------|------|
| `check_service.RunCheck` con casi (no cert deployment, head=cert, head in snapshot, head fuori) | unit con repo mockati | `internal/service/check_service_test.go` |
| `webhook_handler.verifySig` con firma valida/invalida/malformata | unit puro | `internal/handler/webhook_handler_test.go` |
| `auth.Issuer` round-trip + token type discrimination | unit puro | `internal/auth/jwt_test.go` |
| `repository` happy path | integration con testcontainers-go o `pgxtest` | `internal/repository/*_test.go` |
| Webhook end-to-end | integration con httptest + DB reale | `cmd/server/server_test.go` |

Per il frontend: `@angular/testing` + `MockProvider` per i servizi HTTP, focus su `release-detail` e `pull-requests` (sono i due con più logica di rendering condizionale).

---

## 11. Estensioni naturali (non implementate)

Ordinate per valore atteso:

1. **Ancestor check git-aware** — chiamata a GitHub Compare API server-side, così il check passa se `head_sha` è un fast-forward del `cert HEAD`.
2. **Rate limiting** sull'endpoint `/auth/login` per mitigare brute force.
3. **WebSocket** per la dashboard, sostituendo il polling 30s con push.
4. **Audit log** dedicato (chi ha fatto cosa quando), separato dagli eventi di dominio.
5. **OpenAPI / Swagger** auto-generata dalle route Chi.
6. **Multi-tenant** se mai servisse (ora un'istanza = un'organizzazione).
7. **Notifiche** (Slack/email) su PR bloccate.

---

## 12. Glossario

- **cert HEAD**: il `commit_sha` dell'ultimo `DeploymentEvent` con `environment=cert` per una data release.
- **Snapshot**: il set di `commit_sha` salvato in `commit_snapshots` al momento del deploy in cert, fornito dal client come elenco di commit del branch in quel momento.
- **Cert check**: il processo che verifica che il `head_sha` di una PR verso prod sia stato effettivamente certificato (= match con cert HEAD oppure presenza nello snapshot).
- **Blocked PR**: PR il cui ultimo cert check ha `passed=false`. Lo stato `pr_status='blocked'` viene impostato automaticamente.
- **Release**: una sequenza di commit identificata da `(project_id, branch_name)`. Una release non è "un singolo deploy" ma il *contesto* che attraversa più ambienti.
