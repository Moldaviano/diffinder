import { Component, OnInit, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ActivatedRoute, RouterLink } from '@angular/router';
import { MatCardModule } from '@angular/material/card';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatDialog, MatDialogModule } from '@angular/material/dialog';
import { MatTableModule } from '@angular/material/table';
import { MatChipsModule } from '@angular/material/chips';
import { ReleasesService } from '../../core/services/releases.service';
import {
  CommitSnapshot, DeploymentEvent, PullRequest, Release,
} from '../../core/models';
import { StatusBadgeComponent } from '../../shared/components/status-badge.component';
import { DeployDialogComponent } from './deploy-dialog.component';
import { NotificationService } from '../../core/services/notification.service';

@Component({
  selector: 'df-release-detail',
  standalone: true,
  imports: [
    CommonModule, RouterLink,
    MatCardModule, MatButtonModule, MatIconModule, MatDialogModule, MatTableModule, MatChipsModule,
    StatusBadgeComponent,
  ],
  template: `
    <div class="df-page" *ngIf="release() as r">
      <div class="df-page-header">
        <div>
          <h1>{{ r.title }}</h1>
          <div style="color:#666">{{ r.branch_name }}
            <df-status-badge [status]="r.status" style="margin-left:8px"></df-status-badge>
          </div>
        </div>
        <button mat-raised-button color="primary" (click)="openDeploy()">
          <mat-icon>publish</mat-icon> Registra deploy
        </button>
      </div>

      <mat-card style="margin-bottom:24px">
        <mat-card-header><mat-card-title>Timeline deployment</mat-card-title></mat-card-header>
        <mat-card-content>
          <ul class="df-timeline">
            <li *ngFor="let d of deployments()">
              <div><strong>{{ d.environment | uppercase }}</strong> — {{ d.commit_sha.substring(0,12) }}</div>
              <div style="color:#666; font-size:13px">{{ d.deployed_at | date:'medium' }}</div>
              <div *ngIf="d.notes" style="font-style:italic; margin-top:4px">{{ d.notes }}</div>
            </li>
            <li *ngIf="!deployments().length">Nessun deploy registrato</li>
          </ul>
        </mat-card-content>
      </mat-card>

      <div style="display:grid; gap:24px; grid-template-columns: 1fr 1fr;">
        <mat-card>
          <mat-card-header><mat-card-title>Commit in cert</mat-card-title></mat-card-header>
          <mat-card-content>
            <table mat-table [dataSource]="commits()" *ngIf="commits().length; else noCommits">
              <ng-container matColumnDef="sha">
                <th mat-header-cell *matHeaderCellDef>SHA</th>
                <td mat-cell *matCellDef="let c">{{ c.commit_sha.substring(0,10) }}</td>
              </ng-container>
              <ng-container matColumnDef="msg">
                <th mat-header-cell *matHeaderCellDef>Messaggio</th>
                <td mat-cell *matCellDef="let c">{{ c.commit_message }}</td>
              </ng-container>
              <ng-container matColumnDef="when">
                <th mat-header-cell *matHeaderCellDef>Quando</th>
                <td mat-cell *matCellDef="let c">{{ c.committed_at | date:'short' }}</td>
              </ng-container>
              <tr mat-header-row *matHeaderRowDef="commitCols"></tr>
              <tr mat-row *matRowDef="let row; columns: commitCols"></tr>
            </table>
            <ng-template #noCommits><div style="color:#888">Nessun commit catturato</div></ng-template>
          </mat-card-content>
        </mat-card>

        <mat-card>
          <mat-card-header><mat-card-title>Pull Request</mat-card-title></mat-card-header>
          <mat-card-content>
            <ul style="list-style:none; padding:0; margin:0">
              <li *ngFor="let pr of prs()" style="padding:8px 0; border-bottom:1px solid #eee">
                <a [href]="pr.pr_url" target="_blank">#{{ pr.pr_number }}</a>
                — {{ pr.head_commit_sha.substring(0,10) }}
                <df-status-badge [status]="pr.status" style="margin-left:8px"></df-status-badge>
              </li>
              <li *ngIf="!prs().length" style="color:#888">Nessuna PR collegata</li>
            </ul>
          </mat-card-content>
        </mat-card>
      </div>
    </div>
  `,
})
export class ReleaseDetailComponent implements OnInit {
  private route = inject(ActivatedRoute);
  private svc = inject(ReleasesService);
  private dialog = inject(MatDialog);
  private notify = inject(NotificationService);

  readonly release = signal<Release | null>(null);
  readonly deployments = signal<DeploymentEvent[]>([]);
  readonly commits = signal<CommitSnapshot[]>([]);
  readonly prs = signal<PullRequest[]>([]);
  readonly commitCols = ['sha', 'msg', 'when'];

  private id = '';

  ngOnInit() {
    this.id = this.route.snapshot.paramMap.get('id') ?? '';
    this.reload();
  }

  private reload() {
    if (!this.id) return;
    this.svc.get(this.id).subscribe(r => this.release.set(r));
    this.svc.deployments(this.id).subscribe(d => this.deployments.set(d ?? []));
    this.svc.commits(this.id).subscribe(c => this.commits.set(c ?? []));
    this.svc.pullRequests(this.id).subscribe(p => this.prs.set(p ?? []));
  }

  openDeploy() {
    const ref = this.dialog.open(DeployDialogComponent, {
      data: { releaseId: this.id }, width: '420px',
    });
    ref.afterClosed().subscribe(req => {
      if (!req) return;
      this.svc.deploy(this.id, req).subscribe(() => {
        this.notify.success('Deploy registrato');
        this.reload();
      });
    });
  }
}
