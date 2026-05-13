import { Component, OnInit, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { MatTabsModule } from '@angular/material/tabs';
import { MatCardModule } from '@angular/material/card';
import { MatTableModule } from '@angular/material/table';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatSelectModule } from '@angular/material/select';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { UsersService } from '../../core/services/users.service';
import { ProjectsService } from '../../core/services/projects.service';
import { Project, User, UserRole } from '../../core/models';
import { NotificationService } from '../../core/services/notification.service';

@Component({
  selector: 'df-settings',
  standalone: true,
  imports: [
    CommonModule, ReactiveFormsModule,
    MatTabsModule, MatCardModule, MatTableModule, MatFormFieldModule,
    MatInputModule, MatSelectModule, MatButtonModule, MatIconModule,
  ],
  template: `
    <div class="df-page">
      <div class="df-page-header"><h1>Impostazioni</h1></div>

      <mat-tab-group>
        <mat-tab label="Utenti">
          <mat-card style="margin-top:16px">
            <mat-card-header><mat-card-title>Crea utente</mat-card-title></mat-card-header>
            <mat-card-content>
              <form [formGroup]="userForm" (ngSubmit)="createUser()" style="display:grid; grid-template-columns: repeat(4, 1fr); gap: 12px">
                <mat-form-field appearance="outline">
                  <mat-label>Username</mat-label>
                  <input matInput formControlName="username" />
                </mat-form-field>
                <mat-form-field appearance="outline">
                  <mat-label>Email</mat-label>
                  <input matInput formControlName="email" type="email" />
                </mat-form-field>
                <mat-form-field appearance="outline">
                  <mat-label>Password</mat-label>
                  <input matInput formControlName="password" type="password" />
                </mat-form-field>
                <mat-form-field appearance="outline">
                  <mat-label>Ruolo</mat-label>
                  <mat-select formControlName="role">
                    <mat-option value="admin">admin</mat-option>
                    <mat-option value="developer">developer</mat-option>
                    <mat-option value="viewer">viewer</mat-option>
                  </mat-select>
                </mat-form-field>
                <div style="grid-column: span 4;">
                  <button mat-raised-button color="primary" type="submit" [disabled]="userForm.invalid">Crea</button>
                </div>
              </form>
            </mat-card-content>
          </mat-card>

          <mat-card style="margin-top:16px">
            <mat-card-header><mat-card-title>Utenti</mat-card-title></mat-card-header>
            <mat-card-content>
              <table mat-table [dataSource]="users()">
                <ng-container matColumnDef="username">
                  <th mat-header-cell *matHeaderCellDef>Username</th>
                  <td mat-cell *matCellDef="let u">{{ u.username }}</td>
                </ng-container>
                <ng-container matColumnDef="email">
                  <th mat-header-cell *matHeaderCellDef>Email</th>
                  <td mat-cell *matCellDef="let u">{{ u.email }}</td>
                </ng-container>
                <ng-container matColumnDef="role">
                  <th mat-header-cell *matHeaderCellDef>Ruolo</th>
                  <td mat-cell *matCellDef="let u">{{ u.role }}</td>
                </ng-container>
                <tr mat-header-row *matHeaderRowDef="userCols"></tr>
                <tr mat-row *matRowDef="let row; columns: userCols"></tr>
              </table>
            </mat-card-content>
          </mat-card>
        </mat-tab>

        <mat-tab label="Webhook tokens">
          <mat-card style="margin-top:16px">
            <mat-card-content>
              <table mat-table [dataSource]="projects()">
                <ng-container matColumnDef="name">
                  <th mat-header-cell *matHeaderCellDef>Progetto</th>
                  <td mat-cell *matCellDef="let p">{{ p.name }}</td>
                </ng-container>
                <ng-container matColumnDef="repo">
                  <th mat-header-cell *matHeaderCellDef>Repo</th>
                  <td mat-cell *matCellDef="let p">{{ p.repository_url }}</td>
                </ng-container>
                <ng-container matColumnDef="token">
                  <th mat-header-cell *matHeaderCellDef>Token</th>
                  <td mat-cell *matCellDef="let p">
                    <code>{{ p.webhook_token }}</code>
                    <button mat-icon-button (click)="copy(p.webhook_token)"><mat-icon>content_copy</mat-icon></button>
                  </td>
                </ng-container>
                <tr mat-header-row *matHeaderRowDef="projectCols"></tr>
                <tr mat-row *matRowDef="let row; columns: projectCols"></tr>
              </table>
              <div style="margin-top:12px; color:#666; font-size:13px">
                Il webhook condivide un secret HMAC globale (env <code>GITHUB_WEBHOOK_SECRET</code>);
                il token per progetto è utilizzabile come identificativo opzionale lato GitHub Actions.
              </div>
            </mat-card-content>
          </mat-card>
        </mat-tab>
      </mat-tab-group>
    </div>
  `,
})
export class SettingsComponent implements OnInit {
  private fb = inject(FormBuilder);
  private usersSvc = inject(UsersService);
  private projectsSvc = inject(ProjectsService);
  private notify = inject(NotificationService);

  readonly userCols = ['username', 'email', 'role'];
  readonly projectCols = ['name', 'repo', 'token'];
  readonly users = signal<User[]>([]);
  readonly projects = signal<Project[]>([]);

  readonly userForm = this.fb.nonNullable.group({
    username: ['', Validators.required],
    email: ['', [Validators.required, Validators.email]],
    password: ['', [Validators.required, Validators.minLength(6)]],
    role: ['developer' as UserRole, Validators.required],
  });

  ngOnInit() {
    this.usersSvc.list(1, 100).subscribe(p => this.users.set(p.items ?? []));
    this.projectsSvc.list(1, 100).subscribe(p => this.projects.set(p.items ?? []));
  }

  createUser() {
    if (this.userForm.invalid) return;
    this.usersSvc.create(this.userForm.getRawValue()).subscribe(() => {
      this.notify.success('Utente creato');
      this.userForm.reset({ role: 'developer' });
      this.usersSvc.list(1, 100).subscribe(p => this.users.set(p.items ?? []));
    });
  }

  copy(s: string) {
    navigator.clipboard.writeText(s);
    this.notify.info('Copiato');
  }
}
