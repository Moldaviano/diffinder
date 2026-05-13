import { Injectable, computed, inject, signal } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, tap } from 'rxjs';
import { TokenPair, User } from '../models';
import { environment } from '../environment';

const STORAGE_KEY = 'diffinder.auth';

@Injectable({ providedIn: 'root' })
export class AuthService {
  private http = inject(HttpClient);

  // Signal-based state, esposta come read-only.
  private _token = signal<string | null>(null);
  private _refresh = signal<string | null>(null);
  private _user = signal<User | null>(null);

  readonly token = this._token.asReadonly();
  readonly user = this._user.asReadonly();
  readonly isLoggedIn = computed(() => !!this._token());
  readonly isAdmin = computed(() => this._user()?.role === 'admin');

  constructor() {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (raw) {
      try {
        const p = JSON.parse(raw) as TokenPair;
        this._token.set(p.access_token);
        this._refresh.set(p.refresh_token);
        this._user.set(p.user);
      } catch {
        localStorage.removeItem(STORAGE_KEY);
      }
    }
  }

  login(email: string, password: string): Observable<TokenPair> {
    return this.http
      .post<TokenPair>(`${environment.apiBase}/auth/login`, { email, password })
      .pipe(tap(p => this.persist(p)));
  }

  refresh(): Observable<TokenPair> {
    return this.http
      .post<TokenPair>(`${environment.apiBase}/auth/refresh`, { refresh_token: this._refresh() })
      .pipe(tap(p => this.persist(p)));
  }

  logout(): void {
    this._token.set(null);
    this._refresh.set(null);
    this._user.set(null);
    localStorage.removeItem(STORAGE_KEY);
  }

  private persist(p: TokenPair) {
    this._token.set(p.access_token);
    this._refresh.set(p.refresh_token);
    this._user.set(p.user);
    localStorage.setItem(STORAGE_KEY, JSON.stringify(p));
  }
}
