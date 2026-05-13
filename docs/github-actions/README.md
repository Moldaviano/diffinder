# Diffinder · GitHub Actions setup

Questa cartella contiene i workflow GitHub Actions che le repo CAME devono
installare per notificare Diffinder degli eventi delle pull request.

## TL;DR (per ogni repo da osservare)

1. Copia **uno** dei due workflow nella repo target:
   - [`diffinder-notify.yml`](./diffinder-notify.yml) → solo notifica, non blocca i merge. **Inizia da qui.**
   - [`diffinder-cert-check.yml`](./diffinder-cert-check.yml) → blocca il PR se il cert check fallisce.

   Path nel repo target: `.github/workflows/diffinder-notify.yml`

2. Su GitHub aggiungi i **secret**: `Settings → Secrets and variables → Actions → New repository secret`
   - `DIFFINDER_URL` → URL pubblico del backend Diffinder (es. `https://diffinder.alloy.it`)
   - `DIFFINDER_WEBHOOK_SECRET` → la **stessa stringa** del `GITHUB_WEBHOOK_SECRET` del backend

   💡 Se le repo sono molte, usa un **Organization secret** (`Organization → Settings → Secrets → Actions`) e abilita le repo che devono accedervi: lo configuri una volta sola.

3. **Registra il progetto** in Diffinder (UI → Progetti → Nuovo) con `repository_url` esattamente uguale a `https://github.com/<owner>/<repo>`. Il match è esatto.

4. **Crea una release** per il branch su cui le PR puntano (UI → Release). Il webhook cerca la release per `head_branch` (o, in fallback, `base_branch`); se non la trova, risponde 404.

5. Apri o aggiorna una PR sul repo: dovresti vedere il job `Notify Diffinder` apparire tra i checks. Nei log del backend:

   ```bash
   docker logs -f diffinder-backend | grep webhooks
   ```

## Architettura della notifica

```
┌─────────────┐  pull_request   ┌──────────────────────┐
│   GitHub    │ ──────────────► │ GitHub Actions       │
│   (repo)    │                 │ diffinder-notify.yml │
└─────────────┘                 └──────────┬───────────┘
                                           │ HTTPS POST  +  HMAC-SHA256
                                           ▼
                                ┌──────────────────────┐
                                │ Diffinder backend    │
                                │ /api/webhooks/       │
                                │   github/pr          │
                                └──────────┬───────────┘
                                           │
                                           ▼
                              crea/aggiorna PR + cert check
```

Il body inviato a Diffinder ha questa forma:

```json
{
  "repo": "https://github.com/alloy/myrepo",
  "pr_number": 123,
  "head_sha": "abc1234...",
  "base_branch": "main",
  "head_branch": "feature/foo",
  "pr_url": "https://github.com/alloy/myrepo/pull/123"
}
```

L'header `X-Hub-Signature-256` contiene `sha256=<HMAC>` calcolato sul body con il secret condiviso. Lato server, [`webhook_handler.go`](../../internal/handler/webhook_handler.go) lo verifica in tempo costante.

## Sviluppo locale: esporre il backend con ngrok

GitHub non riesce a raggiungere `localhost:8080`. In dev espandi il backend con un tunnel:

```bash
# in un terminale
docker compose up -d backend

# in un altro terminale (richiede account ngrok gratuito)
ngrok http 8080
```

ngrok ti dà un URL del tipo `https://abc-123-456.ngrok-free.app`. Usalo come `DIFFINDER_URL`. Vedrai le richieste passare anche dal dashboard ngrok (`http://localhost:4040`) — utile per ispezionare body e headers.

> ⚠️ Ogni volta che riavvii ngrok l'URL cambia (senza piano a pagamento). Aggiorna il secret nelle repo o usa un sottodominio statico (`ngrok http --domain=...`).

## Test manuale (senza GitHub)

Per verificare che l'endpoint risponda correttamente, dalla tua macchina:

```bash
SECRET="dev-webhook-secret"   # uguale a GITHUB_WEBHOOK_SECRET
BODY='{"repo":"https://github.com/alloy/test","pr_number":1,"head_sha":"abc1234","base_branch":"main","head_branch":"feature/x","pr_url":"https://github.com/alloy/test/pull/1"}'
SIG=$(printf '%s' "$BODY" | openssl dgst -sha256 -hmac "$SECRET" -hex | awk '{print $2}')

curl -i -X POST http://localhost:8080/api/webhooks/github/pr \
  -H "Content-Type: application/json" \
  -H "X-Hub-Signature-256: sha256=$SIG" \
  -d "$BODY"
```

## Troubleshooting

| Sintomo                                                  | Causa probabile                                                                | Fix                                                                                              |
|----------------------------------------------------------|--------------------------------------------------------------------------------|--------------------------------------------------------------------------------------------------|
| Workflow non parte                                       | Manca il filtro `branches` o la PR è in draft (con `ready_for_review`)         | Verifica `on.pull_request.branches` e i `types`                                                  |
| `401 invalid signature` nei log backend                  | `DIFFINDER_WEBHOOK_SECRET` ≠ `GITHUB_WEBHOOK_SECRET`                            | Allinea i due valori. Attenzione a spazi / newline finali nei secret                             |
| `404 project not registered for repo`                    | `repository_url` del progetto in Diffinder non combacia con `repo` del payload | Registra il progetto con URL esatto `https://github.com/<owner>/<repo>`                          |
| `404 no release tracked for branch`                      | Manca la release per quel branch                                               | Crea la release in Diffinder con `branch_name` = `head_branch` della PR                          |
| `HTTP 000` nel workflow                                  | `DIFFINDER_URL` irraggiungibile dalla rete GitHub Actions                      | Esponi pubblicamente il backend (ngrok in dev, reverse proxy/Cloudflare in prod)                 |
| Nessuna riga `path:"/api/webhooks/github/pr"` nei log    | Il workflow non arriva al backend                                              | Controlla i log del run Actions; controlla DNS/TLS se è dietro un proxy                          |

## Note operative

- I workflow usano `concurrency` per cancellare run obsoleti quando arrivano push ravvicinati alla stessa PR: questo evita di inondare il backend di webhook duplicati.
- `permissions:` è ridotto al minimo (`contents: read`, `pull-requests: read`) — il workflow non scrive sulla repo.
- La firma viene mascherata nei log (`::add-mask::`) per igiene, anche se non è formalmente un segreto (è derivata, ma rivela il contenuto del secret se combinata con il body).
