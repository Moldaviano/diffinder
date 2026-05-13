import { Component, OnInit, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { MatTableModule } from '@angular/material/table';
import { MatSlideToggleModule } from '@angular/material/slide-toggle';
import { MatIconModule } from '@angular/material/icon';
import { MatButtonModule } from '@angular/material/button';
import { MatPaginatorModule, PageEvent } from '@angular/material/paginator';
import { PullRequestsService } from '../../core/services/pull-requests.service';
import { CertificationCheck, PullRequest } from '../../core/models';
import { StatusBadgeComponent } from '../../shared/components/status-badge.component';
import { NotificationService } from '../../core/services/notification.service';
import { forkJoin } from 'rxjs';

interface Row extends PullRequest { last_check: CertificationCheck | null; }

@Component({
  selector: 'df-pull-requests',
  standalone: true,
  imports: [
    CommonModule, FormsModule, MatTableModule, MatSlideToggleModule,
    MatIconModule, MatButtonModule, MatPaginatorModule,
    StatusBadgeComponent,
  ],
  template: `
    <div class="df-page">
      <div class="df-page-header">
        <h1>Pull Requests</h1>
        <mat-slide-toggle [(ngModel)]="onlyBlocked" (change)="reload()">Solo bloccate</mat-slide-toggle>
      </div>

      <table mat-table [dataSource]="rows()" style="width:100%; background:white">
        <ng-container matColumnDef="pr">
          <th mat-header-cell *matHeaderCellDef>PR</th>
          <td mat-cell *matCellDef="let r">
            <a [href]="r.pr_url" target="_blank">#{{ r.pr_number }}</a>
            <div style="color:#888; font-size:12px">{{ r.head_commit_sha.substring(0,10) }} → {{ r.base_branch }}</div>
          </td>
        </ng-container>
        <ng-container matColumnDef="status">
          <th mat-header-cell *matHeaderCellDef>Stato</th>
          <td mat-cell *matCellDef="let r"><df-status-badge [status]="r.status"></df-status-badge></td>
        </ng-container>
        <ng-container matColumnDef="cert">
          <th mat-header-cell *matHeaderCellDef>Cert check</th>
          <td mat-cell *matCellDef="let r">
            <span *ngIf="!r.last_check" class="df-cert-none">non eseguito</span>
            <span *ngIf="r.last_check?.passed" class="df-cert-pass">
              <mat-icon style="vertical-align:middle; font-size:18px">check_circle</mat-icon>
              passato
            </span>
            <span *ngIf="r.last_check && !r.last_check.passed" class="df-cert-fail" [title]="r.last_check.details">
              <mat-icon style="vertical-align:middle; font-size:18px">cancel</mat-icon>
              fallito
            </span>
          </td>
        </ng-container>
        <ng-container matColumnDef="opened">
          <th mat-header-cell *matHeaderCellDef>Aperta</th>
          <td mat-cell *matCellDef="let r">{{ r.opened_at | date:'short' }}</td>
        </ng-container>
        <ng-container matColumnDef="actions">
          <th mat-header-cell *matHeaderCellDef></th>
          <td mat-cell *matCellDef="let r">
            <button mat-button color="primary" (click)="runCheck(r)">Esegui check</button>
          </td>
        </ng-container>

        <tr mat-header-row *matHeaderRowDef="cols"></tr>
        <tr mat-row *matRowDef="let row; columns: cols"></tr>
      </table>

      <mat-paginator [length]="total()" [pageSize]="limit" [pageIndex]="page - 1"
                     [pageSizeOptions]="[10,20,50]" (page)="onPage($event)"></mat-paginator>
    </div>
  `,
})
export class PullRequestsComponent implements OnInit {
  private svc = inject(PullRequestsService);
  private notify = inject(NotificationService);

  readonly cols = ['pr', 'status', 'cert', 'opened', 'actions'];
  readonly rows = signal<Row[]>([]);
  readonly total = signal(0);

  page = 1; limit = 20; onlyBlocked = false;

  ngOnInit() { this.reload(); }

  reload() {
    this.svc.list(this.page, this.limit, this.onlyBlocked).subscribe(p => {
      const items = p.items ?? [];
      if (items.length === 0) {
        this.rows.set([]); this.total.set(p.total); return;
      }
      forkJoin(items.map(i => this.svc.history(i.id))).subscribe(histories => {
        const enriched: Row[] = items.map((pr, idx) => ({
          ...pr, last_check: histories[idx]?.[0] ?? null,
        }));
        this.rows.set(enriched);
        this.total.set(p.total);
      });
    });
  }

  runCheck(r: Row) {
    this.svc.runCheck(r.id).subscribe(c => {
      this.notify.success(c.passed ? 'Check passato' : 'Check fallito');
      this.reload();
    });
  }

  onPage(e: PageEvent) {
    this.page = e.pageIndex + 1; this.limit = e.pageSize; this.reload();
  }
}
