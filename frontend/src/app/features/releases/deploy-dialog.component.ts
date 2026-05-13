import { Component, Inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { MAT_DIALOG_DATA, MatDialogModule, MatDialogRef } from '@angular/material/dialog';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatSelectModule } from '@angular/material/select';
import { MatButtonModule } from '@angular/material/button';
import { Environment } from '../../core/models';
import { DeployRequest } from '../../core/services/releases.service';

export interface DeployDialogData { releaseId: string; }

@Component({
  selector: 'df-deploy-dialog',
  standalone: true,
  imports: [
    CommonModule, ReactiveFormsModule, MatDialogModule,
    MatFormFieldModule, MatInputModule, MatSelectModule, MatButtonModule,
  ],
  template: `
    <h2 mat-dialog-title>Registra deploy</h2>
    <mat-dialog-content>
      <form [formGroup]="form">
        <mat-form-field appearance="outline" style="width:100%">
          <mat-label>Ambiente</mat-label>
          <mat-select formControlName="environment">
            <mat-option value="dev">dev</mat-option>
            <mat-option value="cert">cert</mat-option>
            <mat-option value="prod">prod</mat-option>
          </mat-select>
        </mat-form-field>
        <mat-form-field appearance="outline" style="width:100%">
          <mat-label>Commit SHA</mat-label>
          <input matInput formControlName="commit_sha" placeholder="40 caratteri esadecimali" />
        </mat-form-field>
        <mat-form-field appearance="outline" style="width:100%">
          <mat-label>Note</mat-label>
          <textarea matInput formControlName="notes" rows="3"></textarea>
        </mat-form-field>
        <div *ngIf="form.value.environment === 'cert'" style="font-size:13px; color:#666">
          Suggerimento: per il deploy in cert puoi passare anche la lista dei commit del branch
          (campo <code>commits</code>) per popolare lo snapshot. In questa UI semplificata
          inviamo solo il SHA HEAD; la lista può essere fornita via API o webhook.
        </div>
      </form>
    </mat-dialog-content>
    <mat-dialog-actions align="end">
      <button mat-button mat-dialog-close>Annulla</button>
      <button mat-raised-button color="primary" [disabled]="form.invalid" (click)="submit()">Registra</button>
    </mat-dialog-actions>
  `,
})
export class DeployDialogComponent {
  constructor(
    private fb: FormBuilder,
    private dialog: MatDialogRef<DeployDialogComponent, DeployRequest>,
    @Inject(MAT_DIALOG_DATA) public data: DeployDialogData,
  ) {}

  readonly form = this.fb.nonNullable.group({
    environment: ['dev' as Environment, Validators.required],
    commit_sha: ['', [Validators.required, Validators.minLength(7)]],
    notes: [''],
  });

  submit() {
    this.dialog.close(this.form.getRawValue() as DeployRequest);
  }
}
