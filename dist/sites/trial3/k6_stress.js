import http from 'k6/http';
import { check, sleep } from 'k6';

// -------------------------------------------------------------------------
// STRESS TEST: Menguji batas antrean FastCGI Semaphore & Concurrency Pool
// Naik hingga 200 - 300 concurrent requests untuk memastikan tidak ada 502 Bad Gateway
// -------------------------------------------------------------------------

export const options = {
  stages: [
    { duration: '15s', target: 100 }, // Naik ke 100 VU
    { duration: '30s', target: 200 }, // Dorong ke 200 VU (uji semaphore queue)
    { duration: '15s', target: 300 }, // Puncak beban 300 VU
    { duration: '15s', target: 0 },   // Turun kembali
  ],
  thresholds: {
    http_req_failed: ['rate<0.02'],     // Error rate ditargetkan < 2%
    http_req_duration: ['p(95)<1500'],  // 95% request di bawah 1.5s di bawah beban ekstrem
  },
};

const BASE_URL = __ENV.TARGET_URL || 'http://localhost:8084';

export default function () {
  // Campuran 70% transaksi POS, 30% simple ping
  const isTransaction = Math.random() < 0.7;

  if (isTransaction) {
    const res = http.get(`${BASE_URL}/api_transaction.php`);
    check(res, {
      'trx status 200': (r) => r.status === 200,
      'has grand_total': (r) => r.json('grand_total') !== undefined,
    });
  } else {
    const res = http.get(`${BASE_URL}/api_simple.php`);
    check(res, {
      'simple status 200': (r) => r.status === 200,
    });
  }

  sleep(0.3);
}

