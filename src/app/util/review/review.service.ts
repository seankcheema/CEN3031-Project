import { HttpClient, HttpHeaders } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';

const httpOptions = {
  headers: new HttpHeaders({
    'Content-Type': 'application/json',
  })
}

export interface Review {
  ID: number;
  CreatedAt: string;
  UpdatedAt: string;
  DeletedAt: string;
  GameName: string;
  Rating: number;
  Description: string;
  Username: string;
  PlayStatus: string;
}

@Injectable({
  providedIn: 'root'
})
export class ReviewService {

  baseURL:string = 'http://localhost:8080';

  constructor(private http : HttpClient) { }

  getRecentReviews(): Observable<Review[]> {
    return this.http.get<Review[]>(this.baseURL + '/recentreviews');
  }

  addReview(review: any): Observable<any> {
    return this.http.post<any>(this.baseURL + '/writeareview', review, httpOptions);
  }
}
