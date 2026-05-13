// Configurazione runtime: in dev punta al backend locale; in prod tipicamente
// si usa reverse proxy (es. nginx) e si setta apiBase = '/api' lasciando relativo.
export const environment = {
  apiBase: '/api',
};
