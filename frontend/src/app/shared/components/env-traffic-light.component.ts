import { Component, Input, computed, signal } from '@angular/core';
import { Release } from '../../core/models';

// Mostra tre dot dev/cert/prod con stato basato sulla release.
@Component({
  selector: 'df-env-traffic-light',
  standalone: true,
  template: `
    <span class="df-traffic-light" [title]="title()">
      <span class="dot" [class.done]="dev() === 'done'">D</span>
      <span class="dot"
            [class.done]="cert() === 'done'"
            [class.cur]="cert() === 'cur'"
            [class.failed]="cert() === 'failed'">C</span>
      <span class="dot" [class.done]="prod() === 'done'" [class.cur]="prod() === 'cur'">P</span>
    </span>
  `,
})
export class EnvTrafficLightComponent {
  private _release = signal<Release | null>(null);

  @Input({ required: true }) set release(r: Release) { this._release.set(r); }

  // Mappa status → stato di ciascun dot. "done" = verde, "cur" = arancio, "failed" = rosso.
  readonly dev = computed<'done' | 'idle'>(() => {
    const s = this._release()?.status;
    return s && ['in_dev','in_cert','approved','in_prod','rejected'].includes(s) ? 'done' : 'idle';
  });
  readonly cert = computed<'done' | 'cur' | 'failed' | 'idle'>(() => {
    const s = this._release()?.status;
    if (!s) return 'idle';
    if (s === 'in_cert') return 'cur';
    if (s === 'approved' || s === 'in_prod') return 'done';
    if (s === 'rejected') return 'failed';
    return 'idle';
  });
  readonly prod = computed<'done' | 'cur' | 'idle'>(() => {
    const s = this._release()?.status;
    return s === 'in_prod' ? 'done' : 'idle';
  });

  readonly title = computed(() => `dev: ${this.dev()} / cert: ${this.cert()} / prod: ${this.prod()}`);
}
