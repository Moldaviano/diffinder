import { Component, OnInit, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MatTableModule } from '@angular/material/table';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatDialog, MatDialogModule } from '@angular/material/dialog';
import { ProjectsService } from '../../core/services/projects.service';
import { Project, ProjectStats } from '../../core/models';
import { ProjectFormDialogComponent } from './project-form-dialog.component';
import { NotificationService } from '../../core/services/notification.service';
import { forkJoin, of } from 'rxjs';
import { catchError } from 'rxjs/operators';

@Component({
  selector: 'df-projects',
  standalone: true,
  imports: [CommonModule, MatTableModule, MatButtonModule, MatIconModule, MatDialogModule],
  template: `
    <div class="df-page">
      <div class="df-page-header">
        <h1>Progetti</h1>
        <button mat-raised-button color="primary" (click)="openForm()">
          <mat-icon>add</mat-icon> Nuovo progetto
        </button>
      </div>

      <table mat-table [dataSource]="rows()" style="width:100%; background:white">
        <ng-container matColumnDef="name">
          <th mat-header-cell *matHeaderCellDef>Nome</th>
          <td mat-cell *matCellDef="let r">
            <strong>{{ r.name }}</strong>
            <div style="color:#888; font-size:12px">{{ r.description }}</div>
          </td>
        </ng-container>
        <ng-container matColumnDef="repo">
          <th mat-header-cell *matHeaderCellDef>Repository</th>
          <td mat-cell *matCellDef="let r"><a [href]="r.repository_url" target="_blank">{{ r.repository_url }}</a></td>
        </ng-container>
        <ng-container matColumnDef="active">
          <th mat-header-cell *matHeaderCellDef>Release attive</th>
          <td mat-cell *matCellDef="let r">{{ stats()[r.id]?.active_releases ?? 0 }}</td>
        </ng-container>
        <ng-container matColumnDef="last">
          <th mat-header-cell *matHeaderCellDef>Ultima attività</th>
          <td mat-cell *matCellDef="let r">{{ (stats()[r.id]?.last_activity | date:'short') || '—' }}</td>
        </ng-container>
        <ng-container matColumnDef="actions">
          <th mat-header-cell *matHeaderCellDef></th>
          <td mat-cell *matCellDef="let r">
            <button mat-icon-button (click)="openForm(r)"><mat-icon>edit</mat-icon></button>
            <button mat-icon-button color="warn" (click)="remove(r)"><mat-icon>delete</mat-icon></button>
          </td>
        </ng-container>

        <tr mat-header-row *matHeaderRowDef="cols"></tr>
        <tr mat-row *matRowDef="let row; columns: cols"></tr>
      </table>
    </div>
  `,
})
export class ProjectsComponent implements OnInit {
  private svc = inject(ProjectsService);
  private dialog = inject(MatDialog);
  private notify = inject(NotificationService);

  readonly cols = ['name', 'repo', 'active', 'last', 'actions'];
  readonly rows = signal<Project[]>([]);
  readonly stats = signal<Record<string, ProjectStats>>({});

  ngOnInit() { this.reload(); }

  reload() {
    this.svc.list(1, 100).subscribe(p => {
      this.rows.set(p.items ?? []);
      if (!p.items?.length) { this.stats.set({}); return; }
      forkJoin(
        p.items.map(it => this.svc.stats(it.id).pipe(catchError(() => of(null))))
      ).subscribe(results => {
        const map: Record<string, ProjectStats> = {};
        results.forEach(r => { if (r) map[r.project_id] = r; });
        this.stats.set(map);
      });
    });
  }

  openForm(project?: Project) {
    const ref = this.dialog.open(ProjectFormDialogComponent, {
      data: { project }, width: '480px',
    });
    ref.afterClosed().subscribe(payload => {
      if (!payload) return;
      const obs = project ? this.svc.update(project.id, payload) : this.svc.create(payload);
      obs.subscribe(() => {
        this.notify.success('Progetto salvato');
        this.reload();
      });
    });
  }

  remove(p: Project) {
    if (!confirm(`Eliminare il progetto "${p.name}"? Anche le release verranno cancellate.`)) return;
    this.svc.remove(p.id).subscribe(() => {
      this.notify.success('Progetto eliminato');
      this.reload();
    });
  }
}
