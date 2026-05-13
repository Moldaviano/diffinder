import { Injectable, inject } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';
import { Paged, Project, ProjectStats, Release } from '../models';
import { environment } from '../environment';

@Injectable({ providedIn: 'root' })
export class ProjectsService {
  private http = inject(HttpClient);
  private base = `${environment.apiBase}/projects`;

  list(page = 1, limit = 20): Observable<Paged<Project>> {
    return this.http.get<Paged<Project>>(this.base, { params: new HttpParams().set('page', page).set('limit', limit) });
  }
  get(id: string)                       { return this.http.get<Project>(`${this.base}/${id}`); }
  create(p: Partial<Project>)           { return this.http.post<Project>(this.base, p); }
  update(id: string, p: Partial<Project>) { return this.http.put<Project>(`${this.base}/${id}`, p); }
  remove(id: string)                    { return this.http.delete<void>(`${this.base}/${id}`); }
  stats(id: string)                     { return this.http.get<ProjectStats>(`${this.base}/${id}/stats`); }
  releases(id: string, page = 1, limit = 20) {
    return this.http.get<Paged<Release>>(`${this.base}/${id}/releases`, {
      params: new HttpParams().set('page', page).set('limit', limit),
    });
  }
}
