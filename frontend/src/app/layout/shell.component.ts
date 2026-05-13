import { Component, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Router, RouterLink, RouterLinkActive, RouterOutlet } from '@angular/router';
import { MatToolbarModule } from '@angular/material/toolbar';
import { MatSidenavModule } from '@angular/material/sidenav';
import { MatListModule } from '@angular/material/list';
import { MatIconModule } from '@angular/material/icon';
import { MatButtonModule } from '@angular/material/button';
import { MatMenuModule } from '@angular/material/menu';
import { AuthService } from '../core/services/auth.service';

@Component({
  selector: 'df-shell',
  standalone: true,
  imports: [
    CommonModule, RouterOutlet, RouterLink, RouterLinkActive,
    MatToolbarModule, MatSidenavModule, MatListModule,
    MatIconModule, MatButtonModule, MatMenuModule,
  ],
  template: `
    <mat-sidenav-container class="df-shell" autosize>
      <mat-sidenav mode="side" opened class="df-sidenav">
        <div class="df-brand">
          <div class="df-brand-mark">
            <span class="df-brand-mark-inner">C</span>
          </div>
          <div class="df-brand-text">
            <div class="df-brand-name">CAME<span>·</span>Diffinder</div>
            <div class="df-brand-sub">Release Intelligence</div>
          </div>
        </div>

        <nav class="df-nav">
          <a routerLink="/dashboard" routerLinkActive="active" class="df-nav-item">
            <mat-icon>dashboard</mat-icon>
            <span>Overview</span>
          </a>
          <a routerLink="/releases" routerLinkActive="active" class="df-nav-item">
            <mat-icon>rocket_launch</mat-icon>
            <span>Release</span>
          </a>
          <a routerLink="/pull-requests" routerLinkActive="active" class="df-nav-item">
            <mat-icon>call_merge</mat-icon>
            <span>Pull Request</span>
          </a>
          <a routerLink="/projects" routerLinkActive="active" class="df-nav-item">
            <mat-icon>folder</mat-icon>
            <span>Progetti</span>
          </a>
          <a *ngIf="auth.isAdmin()" routerLink="/settings" routerLinkActive="active" class="df-nav-item">
            <mat-icon>settings</mat-icon>
            <span>Impostazioni</span>
          </a>
        </nav>

        <div class="df-sidenav-footer">
          <div class="df-sidenav-footer-row">
            <span class="df-status-dot"></span>
            <span>Sistema operativo</span>
          </div>
          <div class="df-sidenav-footer-version">v0.1 · build stable</div>
        </div>
      </mat-sidenav>

      <mat-sidenav-content>
        <header class="df-topbar">
          <div class="df-topbar-crumb">
            <mat-icon>chevron_right</mat-icon>
            <span>Workspace</span>
          </div>
          <span class="df-spacer"></span>
          <button class="df-user-pill" [matMenuTriggerFor]="menu">
            <span class="df-avatar">{{ initials() }}</span>
            <span class="df-user-name">{{ auth.user()?.username }}</span>
            <mat-icon class="df-chevron">expand_more</mat-icon>
          </button>
          <mat-menu #menu="matMenu" xPosition="before">
            <div class="df-menu-head">
              <div class="df-menu-name">{{ auth.user()?.username }}</div>
              <div class="df-menu-email">{{ auth.user()?.email }}</div>
            </div>
            <button mat-menu-item (click)="logout()">
              <mat-icon>logout</mat-icon>
              <span>Logout</span>
            </button>
          </mat-menu>
        </header>
        <main class="df-main"><router-outlet></router-outlet></main>
      </mat-sidenav-content>
    </mat-sidenav-container>
  `,
  styles: [`
    :host { display: block; height: 100vh; }
    .df-shell { height: 100vh; background: var(--bg); }

    /* ── Sidenav ────────────────────────────────────────── */
    .df-sidenav {
      width: var(--sidenav-w);
      background: linear-gradient(180deg, #0E1014 0%, #161A20 60%, #181C23 100%);
      color: #E8EDF3;
      border-right: 1px solid rgba(255,255,255,0.06);
      padding: 18px 14px;
      display: flex;
      flex-direction: column;
    }

    .df-brand {
      display: flex;
      align-items: center;
      gap: 12px;
      padding: 6px 8px 18px;
      border-bottom: 1px solid rgba(255,255,255,0.06);
      margin-bottom: 14px;
    }
    .df-brand-mark {
      width: 38px; height: 38px;
      border-radius: 10px;
      background: linear-gradient(135deg, var(--came-blue) 0%, var(--came-blue-700) 100%);
      display: inline-flex; align-items: center; justify-content: center;
      box-shadow: 0 6px 16px rgba(0, 176, 237, 0.35), inset 0 0 0 1px rgba(255,255,255,0.18);
    }
    .df-brand-mark-inner {
      color: #fff;
      font-weight: 800;
      font-size: 18px;
      letter-spacing: -0.02em;
      text-shadow: 0 1px 0 rgba(0,0,0,0.15);
    }
    .df-brand-text { line-height: 1.15; }
    .df-brand-name {
      font-weight: 700;
      font-size: 14.5px;
      letter-spacing: 0.02em;
      color: #fff;
    }
    .df-brand-name span {
      color: var(--came-blue);
      margin: 0 4px;
    }
    .df-brand-sub {
      font-size: 11px;
      color: #8089A0;
      margin-top: 2px;
      letter-spacing: 0.06em;
      text-transform: uppercase;
    }

    /* ── Nav items ──────────────────────────────────────── */
    .df-nav { display: flex; flex-direction: column; gap: 2px; flex: 1; }
    .df-nav-item {
      display: flex;
      align-items: center;
      gap: 12px;
      padding: 10px 12px;
      border-radius: 8px;
      color: #B7BFCC;
      font-size: 13.5px;
      font-weight: 500;
      letter-spacing: 0.01em;
      transition: background 140ms ease, color 140ms ease, transform 80ms ease;
      position: relative;
    }
    .df-nav-item:hover {
      background: rgba(255,255,255,0.04);
      color: #fff;
      text-decoration: none;
    }
    .df-nav-item mat-icon {
      font-size: 20px;
      width: 20px; height: 20px;
      opacity: 0.85;
    }
    .df-nav-item.active {
      background: linear-gradient(90deg, rgba(0,176,237,0.18) 0%, rgba(0,176,237,0.04) 100%);
      color: #fff;
      box-shadow: inset 0 0 0 1px rgba(0,176,237,0.25);
    }
    .df-nav-item.active::before {
      content: '';
      position: absolute;
      left: -14px; top: 8px; bottom: 8px;
      width: 3px;
      background: var(--came-blue);
      border-radius: 0 3px 3px 0;
      box-shadow: 0 0 12px var(--came-blue);
    }
    .df-nav-item.active mat-icon { color: var(--came-blue); opacity: 1; }

    /* ── Sidenav footer ─────────────────────────────────── */
    .df-sidenav-footer {
      padding: 14px 10px 6px;
      border-top: 1px solid rgba(255,255,255,0.06);
      color: #7C8598;
      font-size: 12px;
    }
    .df-sidenav-footer-row {
      display: flex; align-items: center; gap: 8px;
      color: #B7BFCC;
    }
    .df-status-dot {
      width: 8px; height: 8px; border-radius: 50%;
      background: #22C55E;
      box-shadow: 0 0 0 3px rgba(34,197,94,0.18);
    }
    .df-sidenav-footer-version {
      margin-top: 4px;
      color: #5A6478;
      font-family: 'JetBrains Mono', monospace;
      font-size: 11px;
    }

    /* ── Topbar ─────────────────────────────────────────── */
    .df-topbar {
      height: var(--topbar-h);
      background: var(--bg-elev);
      border-bottom: 1px solid var(--border);
      display: flex;
      align-items: center;
      gap: 16px;
      padding: 0 24px;
      position: sticky;
      top: 0;
      z-index: 10;
    }
    .df-topbar-crumb {
      display: flex; align-items: center; gap: 4px;
      color: var(--text-muted);
      font-size: 13px;
      font-weight: 500;
    }
    .df-topbar-crumb mat-icon { font-size: 18px; width: 18px; height: 18px; color: var(--text-soft); }

    .df-user-pill {
      display: inline-flex;
      align-items: center;
      gap: 10px;
      padding: 5px 12px 5px 5px;
      border-radius: 999px;
      background: var(--surface-2);
      border: 1px solid var(--border);
      cursor: pointer;
      color: var(--text);
      font-size: 13px;
      font-weight: 500;
      transition: background 120ms ease, border-color 120ms ease;
    }
    .df-user-pill:hover { background: var(--came-blue-50); border-color: var(--came-blue-100); }
    .df-avatar {
      width: 28px; height: 28px;
      border-radius: 50%;
      background: linear-gradient(135deg, var(--came-blue) 0%, var(--came-blue-700) 100%);
      color: #fff;
      display: inline-flex; align-items: center; justify-content: center;
      font-size: 12px; font-weight: 700;
      letter-spacing: 0.02em;
    }
    .df-user-name { line-height: 1; }
    .df-chevron { font-size: 18px; width: 18px; height: 18px; color: var(--text-soft); }

    .df-menu-head { padding: 10px 14px 8px; border-bottom: 1px solid var(--border); }
    .df-menu-name { font-weight: 600; font-size: 13px; color: var(--text); }
    .df-menu-email { font-size: 12px; color: var(--text-muted); margin-top: 2px; }

    .df-main { min-height: calc(100vh - var(--topbar-h)); }

    @media (max-width: 768px) {
      .df-sidenav { width: 72px; }
      .df-brand-text, .df-nav-item span, .df-sidenav-footer { display: none; }
      .df-nav-item { justify-content: center; }
    }
  `],
})
export class ShellComponent {
  readonly auth = inject(AuthService);
  private router = inject(Router);
  logout() {
    this.auth.logout();
    this.router.navigate(['/login']);
  }
  initials(): string {
    const name = this.auth.user()?.username || '';
    return name
      .split(/[\s._-]+/)
      .filter(Boolean)
      .slice(0, 2)
      .map(p => p[0]?.toUpperCase() ?? '')
      .join('') || '?';
  }
}
