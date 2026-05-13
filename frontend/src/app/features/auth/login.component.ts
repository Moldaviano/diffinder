import { Component, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { Router } from '@angular/router';
import { AuthService } from '../../core/services/auth.service';

@Component({
  selector: 'df-login',
  standalone: true,
  imports: [
    CommonModule, ReactiveFormsModule,
    MatFormFieldModule, MatInputModule, MatButtonModule, MatIconModule,
  ],
  template: `
    <div class="df-login">
      <aside class="df-login-hero">
        <div class="df-login-hero-bg">
          <span class="orb orb-a"></span>
          <span class="orb orb-b"></span>
          <span class="grid"></span>
        </div>
        <div class="df-login-hero-content">
          <div class="df-login-brand">
            <div class="df-brand-mark">
              <span class="df-brand-mark-inner">C</span>
            </div>
            <div>
              <div class="df-brand-name">CAME<span>·</span>Diffinder</div>
              <div class="df-brand-sub">Release Intelligence</div>
            </div>
          </div>

          <h2>Coordinate ogni release.<br/>Senza sorprese.</h2>
          <p>Diffinder unifica pull request, certificazioni e deploy in un'unica console. Costruito sui processi di rilascio CAME.</p>

          <ul class="df-login-bullets">
            <li><mat-icon>check_circle</mat-icon> Stato real-time degli ambienti</li>
            <li><mat-icon>check_circle</mat-icon> Tracking certificazioni e approvazioni</li>
            <li><mat-icon>check_circle</mat-icon> Audit trail completo</li>
          </ul>
        </div>
        <div class="df-login-hero-footer">© {{ year }} CAME S.p.A.</div>
      </aside>

      <section class="df-login-panel">
        <div class="df-login-card">
          <h1>Bentornato</h1>
          <p class="df-muted" style="margin-top:6px">Accedi al tuo workspace Diffinder</p>

          <form [formGroup]="form" (ngSubmit)="submit()" style="margin-top:28px">
            <mat-form-field appearance="outline" style="width:100%">
              <mat-label>Email</mat-label>
              <input matInput formControlName="email" type="email" autocomplete="username" placeholder="nome.cognome@came.it" />
              <mat-icon matPrefix style="color:var(--text-soft);margin-right:6px">alternate_email</mat-icon>
            </mat-form-field>
            <mat-form-field appearance="outline" style="width:100%">
              <mat-label>Password</mat-label>
              <input matInput formControlName="password" type="password" autocomplete="current-password" />
              <mat-icon matPrefix style="color:var(--text-soft);margin-right:6px">lock</mat-icon>
            </mat-form-field>

            <button mat-raised-button color="primary" type="submit"
                    [disabled]="form.invalid || loading()"
                    class="df-login-submit">
              <mat-icon *ngIf="!loading()">login</mat-icon>
              {{ loading() ? 'Accesso in corso…' : 'Accedi' }}
            </button>
          </form>

          <div class="df-login-meta">
            <span class="df-login-secure">
              <mat-icon>shield</mat-icon> Connessione sicura · SSO compatibile
            </span>
          </div>
        </div>
      </section>
    </div>
  `,
  styles: [`
    :host { display: block; }

    .df-login {
      min-height: 100vh;
      display: grid;
      grid-template-columns: minmax(0, 1.05fr) minmax(0, 0.95fr);
      background: var(--bg);
    }

    /* ─── Hero (left) ───────────────────────────────── */
    .df-login-hero {
      position: relative;
      overflow: hidden;
      background: linear-gradient(135deg, #0B0E13 0%, #0F1620 45%, #0A2236 100%);
      color: #fff;
      padding: 56px 64px;
      display: flex;
      flex-direction: column;
      justify-content: space-between;
    }
    .df-login-hero-bg { position: absolute; inset: 0; pointer-events: none; }
    .df-login-hero-bg .orb {
      position: absolute;
      border-radius: 50%;
      filter: blur(70px);
      opacity: 0.55;
    }
    .df-login-hero-bg .orb-a {
      width: 420px; height: 420px;
      background: radial-gradient(circle, var(--came-blue) 0%, rgba(0,176,237,0) 70%);
      top: -120px; right: -80px;
    }
    .df-login-hero-bg .orb-b {
      width: 360px; height: 360px;
      background: radial-gradient(circle, #1E88E5 0%, rgba(30,136,229,0) 70%);
      bottom: -100px; left: -100px;
      opacity: 0.45;
    }
    .df-login-hero-bg .grid {
      position: absolute; inset: 0;
      background-image:
        linear-gradient(rgba(255,255,255,0.04) 1px, transparent 1px),
        linear-gradient(90deg, rgba(255,255,255,0.04) 1px, transparent 1px);
      background-size: 44px 44px;
      mask-image: radial-gradient(ellipse at 60% 40%, #000 30%, transparent 75%);
    }

    .df-login-hero-content { position: relative; z-index: 1; max-width: 480px; }
    .df-login-brand {
      display: flex; align-items: center; gap: 14px;
      margin-bottom: 80px;
    }
    .df-brand-mark {
      width: 44px; height: 44px;
      border-radius: 12px;
      background: linear-gradient(135deg, var(--came-blue) 0%, var(--came-blue-700) 100%);
      display: inline-flex; align-items: center; justify-content: center;
      box-shadow: 0 8px 24px rgba(0, 176, 237, 0.4), inset 0 0 0 1px rgba(255,255,255,0.18);
    }
    .df-brand-mark-inner { color: #fff; font-weight: 800; font-size: 20px; letter-spacing: -0.02em; }
    .df-brand-name { font-weight: 700; font-size: 16px; letter-spacing: 0.02em; }
    .df-brand-name span { color: var(--came-blue); margin: 0 5px; }
    .df-brand-sub { font-size: 11px; color: #8B95AB; letter-spacing: 0.1em; text-transform: uppercase; margin-top: 2px; }

    .df-login-hero h2 {
      font-size: 40px;
      line-height: 1.1;
      font-weight: 700;
      letter-spacing: -0.025em;
      color: #fff;
      margin: 0 0 20px;
    }
    .df-login-hero p {
      font-size: 16px;
      line-height: 1.55;
      color: #B9C2D4;
      margin: 0 0 32px;
      max-width: 440px;
    }
    .df-login-bullets {
      list-style: none;
      padding: 0; margin: 0;
      display: flex; flex-direction: column; gap: 12px;
    }
    .df-login-bullets li {
      display: flex; align-items: center; gap: 10px;
      font-size: 14px;
      color: #DCE3F0;
    }
    .df-login-bullets mat-icon {
      color: var(--came-blue);
      font-size: 20px; width: 20px; height: 20px;
    }

    .df-login-hero-footer {
      position: relative; z-index: 1;
      font-size: 12px;
      color: #6E7790;
      letter-spacing: 0.04em;
    }

    /* ─── Panel (right) ─────────────────────────────── */
    .df-login-panel {
      display: flex; align-items: center; justify-content: center;
      padding: 40px;
    }
    .df-login-card {
      width: 100%;
      max-width: 420px;
      background: var(--bg-elev);
      border: 1px solid var(--border);
      border-radius: var(--radius-lg);
      box-shadow: var(--shadow-2);
      padding: 40px;
    }
    .df-login-card h1 {
      margin: 0;
      font-size: 28px;
      font-weight: 700;
      letter-spacing: -0.02em;
    }

    .df-login-submit {
      width: 100%;
      height: 46px !important;
      border-radius: 10px !important;
      font-weight: 600 !important;
      letter-spacing: 0.01em;
      margin-top: 6px;
    }
    .df-login-submit mat-icon { margin-right: 8px; }

    .df-login-meta {
      margin-top: 22px;
      padding-top: 18px;
      border-top: 1px solid var(--border);
      text-align: center;
    }
    .df-login-secure {
      display: inline-flex; align-items: center; gap: 6px;
      color: var(--text-muted);
      font-size: 12px;
    }
    .df-login-secure mat-icon { font-size: 14px; width: 14px; height: 14px; color: var(--came-blue-700); }

    @media (max-width: 900px) {
      .df-login { grid-template-columns: 1fr; }
      .df-login-hero { padding: 36px; min-height: 280px; }
      .df-login-brand { margin-bottom: 32px; }
      .df-login-hero h2 { font-size: 28px; }
      .df-login-hero p { font-size: 14px; }
      .df-login-bullets { display: none; }
      .df-login-panel { padding: 28px 20px 40px; }
    }
  `],
})
export class LoginComponent {
  private fb = inject(FormBuilder);
  private auth = inject(AuthService);
  private router = inject(Router);
  readonly loading = signal(false);
  readonly year = new Date().getFullYear();

  readonly form = this.fb.nonNullable.group({
    email: ['', [Validators.required, Validators.email]],
    password: ['', Validators.required],
  });

  submit() {
    if (this.form.invalid) return;
    this.loading.set(true);
    const { email, password } = this.form.getRawValue();
    this.auth.login(email, password).subscribe({
      next: () => { this.loading.set(false); this.router.navigate(['/']); },
      error: () => this.loading.set(false),
    });
  }
}
