import { Injectable, inject } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { CertificationCheck, Paged, PRStatus, PullRequest } from '../models';
import { environment } from '../environment';

@Injectable({ providedIn: 'root' })
export class PullRequestsService {
  private http = inject(HttpClient);
  private base = `${environment.apiBase}/pull-requests`;

  list(page = 1, limit = 20, blocked = false) {
    let p = new HttpParams().set('page', page).set('limit', limit);
    if (blocked) p = p.set('blocked', 'true');
    return this.http.get<Paged<PullRequest>>(this.base, { params: p });
  }
  get(id: string) { return this.http.get<PullRequest>(`${this.base}/${id}`); }
  create(p: Partial<PullRequest>) { return this.http.post<PullRequest>(this.base, p); }
  updateStatus(id: string, status: PRStatus) {
    return this.http.put<PullRequest>(`${this.base}/${id}/status`, { status });
  }
  runCheck(id: string) { return this.http.post<CertificationCheck>(`${this.base}/${id}/check-cert`, {}); }
  history(id: string)  { return this.http.get<CertificationCheck[]>(`${this.base}/${id}/checks`); }
}
