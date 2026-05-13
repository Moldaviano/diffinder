import { Injectable, inject } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { ActivityItem, Paged, PullRequest, StatusCount, Summary } from '../models';
import { environment } from '../environment';

@Injectable({ providedIn: 'root' })
export class DashboardService {
  private http = inject(HttpClient);
  private base = `${environment.apiBase}/dashboard`;

  summary() { return this.http.get<Summary>(`${this.base}/summary`); }
  byStatus() { return this.http.get<StatusCount[]>(`${this.base}/releases-by-status`); }
  recent(limit = 20) {
    return this.http.get<ActivityItem[]>(`${this.base}/recent-activity`, {
      params: new HttpParams().set('limit', limit),
    });
  }
  blocked(page = 1, limit = 20) {
    return this.http.get<Paged<PullRequest>>(`${this.base}/blocked-prs`, {
      params: new HttpParams().set('page', page).set('limit', limit),
    });
  }
}
