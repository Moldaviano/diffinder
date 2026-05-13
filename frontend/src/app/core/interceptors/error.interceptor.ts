import { HttpErrorResponse, HttpInterceptorFn } from '@angular/common/http';
import { inject } from '@angular/core';
import { Router } from '@angular/router';
import { catchError, throwError } from 'rxjs';
import { AuthService } from '../services/auth.service';
import { NotificationService } from '../services/notification.service';

// Intercetta errori HTTP e mostra snackbar; su 401 fa logout e redirect a /login.
export const errorInterceptor: HttpInterceptorFn = (req, next) => {
  const notify = inject(NotificationService);
  const auth = inject(AuthService);
  const router = inject(Router);

  return next(req).pipe(
    catchError((err: HttpErrorResponse) => {
      const body = err.error as { error?: string; code?: string } | null;
      const msg = body?.error ?? err.message ?? 'Errore sconosciuto';

      if (err.status === 401 && !req.url.includes('/auth/')) {
        auth.logout();
        router.navigate(['/login']);
      }
      notify.error(msg);
      return throwError(() => err);
    })
  );
};
