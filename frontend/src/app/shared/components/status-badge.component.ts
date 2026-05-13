import { Component, Input } from '@angular/core';

@Component({
  selector: 'df-status-badge',
  standalone: true,
  template: `<span class="df-status-badge df-status-{{ status }}">{{ status }}</span>`,
})
export class StatusBadgeComponent {
  @Input({ required: true }) status!: string;
}
