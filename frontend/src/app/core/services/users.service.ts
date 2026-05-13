import { Injectable, inject } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Paged, User, UserRole } from '../models';
import { environment } from '../environment';

export interface CreateUserRequest {
  username: string;
  email: string;
  password: string;
  role: UserRole;
}

@Injectable({ providedIn: 'root' })
export class UsersService {
  private http = inject(HttpClient);
  private base = `${environment.apiBase}/users`;

  list(page = 1, limit = 20) {
    return this.http.get<Paged<User>>(this.base, {
      params: new HttpParams().set('page', page).set('limit', limit),
    });
  }
  create(req: CreateUserRequest) { return this.http.post<User>(this.base, req); }
}
