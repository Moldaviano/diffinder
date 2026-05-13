import { Injectable, inject } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import {
  CommitSnapshot, DeploymentEvent, Environment, Paged, PullRequest, Release, ReleaseStatus,
} from '../models';
import { environment } from '../environment';

export interface DeployRequest {
  environment: Environment;
  commit_sha: string;
  notes?: string;
  commits?: Partial<CommitSnapshot>[];
}

@Injectable({ providedIn: 'root' })
export class ReleasesService {
  private http = inject(HttpClient);
  private base = `${environment.apiBase}/releases`;

  list(opts: { page?: number; limit?: number; project_id?: string; status?: ReleaseStatus } = {}) {
    let p = new HttpParams().set('page', opts.page ?? 1).set('limit', opts.limit ?? 20);
    if (opts.project_id) p = p.set('project_id', opts.project_id);
    if (opts.status) p = p.set('status', opts.status);
    return this.http.get<Paged<Release>>(this.base, { params: p });
  }
  get(id: string)                  { return this.http.get<Release>(`${this.base}/${id}`); }
  create(r: Partial<Release>)      { return this.http.post<Release>(this.base, r); }
  update(id: string, r: Partial<Release>) { return this.http.put<Release>(`${this.base}/${id}`, r); }
  deployments(id: string)          { return this.http.get<DeploymentEvent[]>(`${this.base}/${id}/deployments`); }
  pullRequests(id: string)         { return this.http.get<PullRequest[]>(`${this.base}/${id}/pull-requests`); }
  commits(id: string)              { return this.http.get<CommitSnapshot[]>(`${this.base}/${id}/commits`); }
  deploy(id: string, req: DeployRequest) {
    return this.http.post<DeploymentEvent>(`${this.base}/${id}/deploy`, req);
  }
}
