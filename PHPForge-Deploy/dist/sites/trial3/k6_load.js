import http from 'k6/http';
import { check, sleep } from 'k6';

// -------------------------------------------------------------------------
// LOAD TEST: Simulasi trafik operasional normal (naik hingga 50 pengguna simultan)
// -------------------------------------------------------------------------

export const options = {
  stages: [
    { duration: '10s', target: 20 },  // Ramp-up perlahan ke 20 VU
    { duration: '30s', target: 50 },  // Beban kerja normal 50 VU
    { duration: '10s', target: 0 },   // Ramp-down ke 0 VU
  ],
  thresholds: {
    http_req_failed: ['rate<0.01'],    // Kegagalan request harus < 1%
    http_req_duration: ['p(95)<800'],  // 95% request di bawah 800ms
  },
};

const BASE_URL = __ENV.TARGET_URL || 'http://localhost:8084';

export default function () {
  // Simulasi kasir melakukan transaksi sync
  const storeId = `CABANG-${Math.floor(Math.random() * 10) + 1}`;
  const res = http.get(`${BASE_URL}/api_transaction.php?store_id=${storeId}`);

  check(res, {
    'status is 200': (r) => r.status === 200,
    'transaction created': (r) => r.json('status') === 'success',
    'has valid grand_total': (r) => r.json('grand_total') > 0,
  });

  sleep(0.5);
}

