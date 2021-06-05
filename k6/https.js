import http from 'k6/http';

export let options = {
  stages: [
    { duration: '20s', target: 1 },
    { duration: '20s', target: 10 },
    { duration: '20s', target: 100 }
  ],
  thresholds: {
    http_req_duration: ['p(99)<1500'], // 99% of requests must complete below 1.5s
  },
};

export default function () {
  http.get('https://test.k6.io');
}