import { HttpInterceptorFn } from '@angular/common/http';
import { inject } from '@angular/core';
import { AuthService } from '../services/auth.service';

// Aggiunge Authorization: Bearer <token> a tutte le richieste verso il backend,
// tranne le rotte di autenticazione (login/refresh).
export const authInterceptor: HttpInterceptorFn = (req, next) => {
  const auth = inject(AuthService);
  const token = auth.token();
  const isAuthRoute = req.url.includes('/auth/login') || req.url.includes('/auth/refresh');
  if (token && !isAuthRoute) {
    req = req.clone({ setHeaders: { Authorization: `Bearer ${token}` } });
  }
  return next(req);
};
