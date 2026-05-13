import { Injectable, inject } from '@angular/core';
import { MatSnackBar } from '@angular/material/snack-bar';

@Injectable({ providedIn: 'root' })
export class NotificationService {
  private sb = inject(MatSnackBar);
  info(msg: string)    { this.sb.open(msg, 'OK', { duration: 3000 }); }
  success(msg: string) { this.sb.open(msg, 'OK', { duration: 3000, panelClass: 'snack-success' }); }
  error(msg: string)   { this.sb.open(msg, 'Chiudi', { duration: 5000, panelClass: 'snack-error' }); }
}
