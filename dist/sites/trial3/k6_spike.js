import http from 'k6/http';
import { check, sleep } from 'k6';

// -------------------------------------------------------------------------
// SPIKE TEST: Lonjakan tiba-tiba (Spike) dari 10 ke 250 VU dalam waktu sangat singkat.
// Menguji apakah semaphore channel mampu menahan lonjakan tanpa membuat socket crash.
// -------------------------------------------------------------------------

export const options = {
  stages: [
    { duration: '5s', target: 10 },    // Baseline tenang
    { duration: '5s', target: 250 },   // SPIKE TIBA-TIBA KE 250 VU!
    { duration: '20s', target: 250 },  // Tahan beban spike
    { duration: '10s', target: 10 },   // Kembali ke normal
    { duration: '5s', target: 0 },     // Selesai
  ],
  thresholds: {
    http_req_failed: ['rate<0.05'],    // Error rate tetap rendah di saat lonjakan ekstrem
  },
};

const BASE_URL = __ENV.TARGET_URL || 'http://localhost:8084';

export default function () {
  const res = http.get(`${BASE_URL}/api_simple.php`);
  check(res, {
    'status is 200': (r) => r.status === 200,
  });
  sleep(0.1);
}

