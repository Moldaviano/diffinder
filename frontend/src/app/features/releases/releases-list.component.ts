import { Component, OnInit, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterLink } from '@angular/router';
import { MatTableModule } from '@angular/material/table';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatSelectModule } from '@angular/material/select';
import { MatInputModule } from '@angular/material/input';
import { MatPaginatorModule, PageEvent } from '@angular/material/paginator';
import { ReleasesService } from '../../core/services/releases.service';
import { ProjectsService } from '../../core/services/projects.service';
import { Project, Release, ReleaseStatus } from '../../core/models';
import { StatusBadgeComponent } from '../../shared/components/status-badge.component';
import { EnvTrafficLightComponent } from '../../shared/components/env-traffic-light.component';

@Component({
  selector: 'df-releases-list',
  standalone: true,
  imports: [
    CommonModule, RouterLink,
    MatTableModule, MatFormFieldModule, MatSelectModule, MatInputModule, MatPaginatorModule,
    StatusBadgeComponent, EnvTrafficLightComponent,
  ],
  template: `
    <div class="df-page">
      <div class="df-page-header"><h1>Release</h1></div>

      <div class="df-toolbar" style="margin-bottom:16px">
        <mat-form-field appearance="outline" subscriptSizing="dynamic">
          <mat-label>Progetto</mat-label>
          <mat-select [(value)]="projectFilter" (selectionChange)="reload()">
            <mat-option [value]="null">Tutti</mat-option>
            <mat-option *ngFor="let p of projects()" [value]="p.id">{{ p.name }}</mat-option>
          </mat-select>
        </mat-form-field>
        <mat-form-field appearance="outline" subscriptSizing="dynamic">
          <mat-label>Stato</mat-label>
          <mat-select [(value)]="statusFilter" (selectionChange)="reload()">
            <mat-option [value]="null">Tutti</mat-option>
            <mat-option value="draft">draft</mat-option>
            <mat-option value="in_dev">in_dev</mat-option>
            <mat-option value="in_cert">in_cert</mat-option>
            <mat-option value="approved">approved</mat-option>
            <mat-option value="in_prod">in_prod</mat-option>
            <mat-option value="rejected">rejected</mat-option>
          </mat-select>
        </mat-form-field>
      </div>

      <table mat-table [dataSource]="rows()" style="width:100%; background:white">
        <ng-container matColumnDef="title">
          <th mat-header-cell *matHeaderCellDef>Titolo</th>
          <td mat-cell *matCellDef="let r">
            <a [routerLink]="['/releases', r.id]">{{ r.title }}</a>
            <div style="color:#888; font-size:12px">{{ r.branch_name }}</div>
          </td>
        </ng-container>
        <ng-container matColumnDef="status">
          <th mat-header-cell *matHeaderCellDef>Stato</th>
          <td mat-cell *matCellDef="let r"><df-status-badge [status]="r.status"></df-status-badge></td>
        </ng-container>
        <ng-container matColumnDef="envs">
          <th mat-header-cell *matHeaderCellDef>Ambienti</th>
          <td mat-cell *matCellDef="let r"><df-env-traffic-light [release]="r"></df-env-traffic-light></td>
        </ng-container>
        <ng-container matColumnDef="updated">
          <th mat-header-cell *matHeaderCellDef>Aggiornata</th>
          <td mat-cell *matCellDef="let r">{{ r.updated_at | date:'short' }}</td>
        </ng-container>

        <tr mat-header-row *matHeaderRowDef="cols"></tr>
        <tr mat-row *matRowDef="let row; columns: cols" [routerLink]="['/releases', row.id]"></tr>
      </table>

      <mat-paginator [length]="total()" [pageSize]="limit" [pageIndex]="page - 1"
                     [pageSizeOptions]="[10,20,50]" (page)="onPage($event)"></mat-paginator>
    </div>
  `,
})
export class ReleasesListComponent implements OnInit {
  private releases = inject(ReleasesService);
  private projectsSvc = inject(ProjectsService);

  readonly cols = ['title', 'status', 'envs', 'updated'];
  readonly projects = signal<Project[]>([]);
  readonly rows = signal<Release[]>([]);
  readonly total = signal(0);

  page = 1;
  limit = 20;
  projectFilter: string | null = null;
  statusFilter: ReleaseStatus | null = null;

  ngOnInit() {
    this.projectsSvc.list(1, 200).subscribe(p => this.projects.set(p.items));
    this.reload();
  }

  reload() {
    this.releases.list({
      page: this.page, limit: this.limit,
      project_id: this.projectFilter ?? undefined,
      status: this.statusFilter ?? undefined,
    }).subscribe(p => {
      this.rows.set(p.items ?? []);
      this.total.set(p.total);
    });
  }

  onPage(e: PageEvent) {
    this.page = e.pageIndex + 1;
    this.limit = e.pageSize;
    this.reload();
  }
}
