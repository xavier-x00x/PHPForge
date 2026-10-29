import http from 'k6/http';
import { check, sleep } from 'k6';

// -------------------------------------------------------------------------
// SMOKE TEST: Verifikasi awal memastikan semua endpoint merespons dengan benar (200 OK)
// -------------------------------------------------------------------------

export const options = {
  vus: 5,           // 5 pengguna virtual
  duration: '15s',  // Berjalan selama 15 detik
  thresholds: {
    http_req_failed: ['rate<0.01'],    // Error rate harus 0%
    http_req_duration: ['p(95)<300'],  // 95% request harus di bawah 300ms
  },
};

const BASE_URL = __ENV.TARGET_URL || 'http://localhost:8084';

export default function () {
  // 1. Test Simple Endpoint
  const resSimple = http.get(`${BASE_URL}/api_simple.php`);
  check(resSimple, {
    'simple status 200': (r) => r.status === 200,
    'simple has status ok': (r) => r.json('status') === 'ok',
  });

  // 2. Test Transaction Endpoint
  const resTrx = http.get(`${BASE_URL}/api_transaction.php?store_id=STORE-TEST`);
  check(resTrx, {
    'trx status 200': (r) => r.status === 200,
    'trx has transaction_id': (r) => r.json('transaction_id') !== undefined,
  });

  // 3. Test POST Endpoint
  const payload = JSON.stringify({
    branch: 'JKT-01',
    timestamp: Date.now(),
    action: 'sync_stock',
    items: [101, 102, 103]
  });
  const headers = { 'Content-Type': 'application/json' };
  const resPost = http.post(`${BASE_URL}/api_post.php`, payload, { headers });
  check(resPost, {
    'post status 200': (r) => r.status === 200,
    'post response success': (r) => r.json('status') === 'success',
  });

  sleep(0.5);
}

