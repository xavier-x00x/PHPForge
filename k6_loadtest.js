import http from 'k6/http';
import { check, sleep } from 'k6';

// Konfigurasi skenario load testing
export const options = {
  stages: [
    { duration: '10s', target: 300 },  // Naik perlahan ke 700 Virtual Users (VU)
    { duration: '30s', target: 300 },  // Tahan beban 700 VU
    { duration: '10s', target: 0 },   // Turun kembali ke 0 VU
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'], // 95% request di bawah 500ms
    http_req_failed: ['rate<0.01'],   // Error rate kurang dari 1%
  },
};

export default function () {
  // Gunakan 'localhost' agar Virtual Host Router dari PHPForge bisa mengenalinya
  // const res = http.get('http://localhost:8083/api_test.php');
  const res = http.get('http://192.168.3.2:8083/api_test.php');

  check(res, {
    'status is 200': (r) => r.status === 200,
    'transaction time OK': (r) => r.timings.duration < 1000,
  });

  sleep(1);
}
