import { Component, Inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { MAT_DIALOG_DATA, MatDialogModule, MatDialogRef } from '@angular/material/dialog';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatButtonModule } from '@angular/material/button';
import { Project } from '../../core/models';

export interface ProjectFormData { project?: Project; }

@Component({
  selector: 'df-project-form',
  standalone: true,
  imports: [
    CommonModule, ReactiveFormsModule, MatDialogModule,
    MatFormFieldModule, MatInputModule, MatButtonModule,
  ],
  template: `
    <h2 mat-dialog-title>{{ data.project ? 'Modifica progetto' : 'Nuovo progetto' }}</h2>
    <mat-dialog-content>
      <form [formGroup]="form">
        <mat-form-field appearance="outline" style="width:100%">
          <mat-label>Nome</mat-label>
          <input matInput formControlName="name" />
        </mat-form-field>
        <mat-form-field appearance="outline" style="width:100%">
          <mat-label>Descrizione</mat-label>
          <input matInput formControlName="description" />
        </mat-form-field>
        <mat-form-field appearance="outline" style="width:100%">
          <mat-label>Repository URL</mat-label>
          <input matInput formControlName="repository_url" placeholder="https://github.com/org/repo" />
        </mat-form-field>
      </form>
    </mat-dialog-content>
    <mat-dialog-actions align="end">
      <button mat-button mat-dialog-close>Annulla</button>
      <button mat-raised-button color="primary" [disabled]="form.invalid" (click)="submit()">Salva</button>
    </mat-dialog-actions>
  `,
})
export class ProjectFormDialogComponent {
  constructor(
    private fb: FormBuilder,
    private ref: MatDialogRef<ProjectFormDialogComponent, Partial<Project>>,
    @Inject(MAT_DIALOG_DATA) public data: ProjectFormData,
  ) {}

  readonly form = this.fb.nonNullable.group({
    name: [this.data.project?.name ?? '', Validators.required],
    description: [this.data.project?.description ?? ''],
    repository_url: [this.data.project?.repository_url ?? '', Validators.required],
  });

  submit() { this.ref.close(this.form.getRawValue()); }
}
