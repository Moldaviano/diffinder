import { Component, OnDestroy, OnInit, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MatIconModule } from '@angular/material/icon';
import { DashboardService } from '../../core/services/dashboard.service';
import { ActivityItem, Summary } from '../../core/models';

@Component({
  selector: 'df-dashboard',
  standalone: true,
  imports: [CommonModule, MatIconModule],
  template: `
    <div class="df-page">
      <div class="df-page-header">
        <div>
          <h1>Overview</h1>
          <span class="subtitle">Stato real-time delle release e dell'attività del team</span>
        </div>
        <div class="df-chip">
          <mat-icon style="font-size:14px;width:14px;height:14px">bolt</mat-icon>
          live · aggiorna ogni 30s
        </div>
      </div>

      <div class="df-card-grid">
        <div class="df-kpi">
          <div class="icon"><mat-icon>rocket_launch</mat-icon></div>
          <div class="label">Release totali</div>
          <div class="value">{{ summary()?.total_releases ?? '—' }}</div>
          <div class="delta">aggregato workspace</div>
        </div>

        <div class="df-kpi warning">
          <div class="icon"><mat-icon>verified</mat-icon></div>
          <div class="label">In certificazione</div>
          <div class="value">{{ summary()?.in_cert ?? '—' }}</div>
          <div class="delta">in attesa di QA</div>
        </div>

        <div class="df-kpi danger">
          <div class="icon"><mat-icon>block</mat-icon></div>
          <div class="label">PR bloccate</div>
          <div class="value">{{ summary()?.blocked_prs ?? '—' }}</div>
          <div class="delta">richiedono intervento</div>
        </div>

        <div class="df-kpi success">
          <div class="icon"><mat-icon>cloud_done</mat-icon></div>
          <div class="label">Deploy oggi</div>
          <div class="value">{{ summary()?.deployments_today ?? '—' }}</div>
          <div class="delta">ultime 24 ore</div>
        </div>
      </div>

      <div class="df-card" style="margin-top:28px">
        <div class="df-card-head">
          <h3>Attività recente</h3>
          <span class="df-muted" style="font-size:12px">ultimi 20 eventi</span>
        </div>
        <div class="df-card-body" style="padding-top:6px;padding-bottom:6px">
          <div class="df-activity">
            <div class="df-activity-row" *ngFor="let a of activity()">
              <span class="df-activity-icon" [ngClass]="iconClass(a)">
                <mat-icon style="font-size:18px;width:18px;height:18px">{{ iconFor(a) }}</mat-icon>
              </span>
              <div class="df-activity-body">
                <div class="df-activity-title">{{ a.title }}</div>
                <div class="df-activity-meta">
                  <span>{{ a.type }}</span>
                  <span *ngIf="a.environment"> → <strong>{{ a.environment }}</strong></span>
                  <span *ngIf="a.commit_sha"> · <code>{{ a.commit_sha?.substring(0,7) }}</code></span>
                </div>
              </div>
              <div class="df-activity-time">{{ a.at | date:'short' }}</div>
            </div>
            <div class="df-activity-row" *ngIf="!activity().length">
              <span class="df-activity-icon"><mat-icon style="font-size:18px;width:18px;height:18px">inbox</mat-icon></span>
              <div class="df-activity-body">
                <div class="df-activity-title">Nessuna attività</div>
                <div class="df-activity-meta">Gli eventi compariranno qui appena disponibili</div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  `,
  styles: [`
    .df-activity-icon.warn { background: var(--warning-bg); color: var(--warning); }
    .df-activity-icon.ok { background: var(--success-bg); color: var(--success); }
    .df-activity-icon.danger { background: var(--danger-bg); color: var(--danger); }
  `],
})
export class DashboardComponent implements OnInit, OnDestroy {
  private dash = inject(DashboardService);

  readonly summary = signal<Summary | null>(null);
  readonly activity = signal<ActivityItem[]>([]);

  private timer?: ReturnType<typeof setInterval>;

  ngOnInit() {
    this.refresh();
    this.timer = setInterval(() => this.refresh(), 30_000);
  }

  ngOnDestroy() {
    if (this.timer) clearInterval(this.timer);
  }

  private refresh() {
    this.dash.summary().subscribe(s => this.summary.set(s));
    this.dash.recent(20).subscribe(a => this.activity.set(a ?? []));
  }

  iconFor(a: ActivityItem): string {
    const t = (a.type || '').toLowerCase();
    if (t.includes('deploy')) return 'rocket_launch';
    if (t.includes('pr') || t.includes('pull')) return 'call_merge';
    if (t.includes('cert')) return 'verified';
    if (t.includes('release')) return 'inventory_2';
    return 'history';
  }
  iconClass(a: ActivityItem): string {
    const t = (a.type || '').toLowerCase();
    if (t.includes('reject') || t.includes('block') || t.includes('fail')) return 'danger';
    if (t.includes('approve') || t.includes('merge') || t.includes('done')) return 'ok';
    if (t.includes('cert')) return 'warn';
    return '';
  }
}
