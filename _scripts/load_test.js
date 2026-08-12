import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  stages: [
    { duration: '30s', target: 10 },
    { duration: '1m', target: 50 },
    { duration: '1m', target: 100 },
    { duration: '2m', target: 500 },
    { duration: '30s', target: 0 },
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'], // 95% das requests < 500ms
    http_req_failed: ['rate<0.01'],   // Falhas < 1%
  },
};

export default function () {
  // Ajuste a URL para o endpoint desejado.
  // IMPORTANTE: odontoflow_backend é o nome do container no docker-compose.yml
  const url = 'http://odontoflow_backend:8080/api/agendamentos';

  // Como a aplicação exige JWT, substitua "SEU_TOKEN_AQUI" por um token válido real gerado na aplicação.
  const payload = null;
  const params = {
    headers: {
      'Accept': 'application/json',
      'Authorization': 'Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJjbGluaWNhSWQiOjEsImV4cCI6MTc4MjQ4OTQzNiwic3ViIjoxLCJ0aXBvVXN1YXJpbyI6IkFETUlOX1RPVEFMIn0.szxi8QJdmzIyJ3XZccaG7p6foI96bAcuM29tqW8K1zG3HhuRETw0OGneJC-ELfXuZQezf3SV0Ng59dE8P4MQYNQJcX7GKb7q9_CLYVCQErVjp1WfEUtvD-wnwAJtX20Z887RIDR4_S4JQXY-riNR3sVc3BVaANBcaIw7sDDVvTpRiBE-Wf0nFl4Zl08AHwKX991BhUWuIJk0LXI7T4YiXU3X4P7gt2OFSiokTmy-sRyJs92Hvd0CUT6hmV-qAugh__DyIKrkWa7_pXLIjUIdwKX32QbunKoBQJfvi1bvrXixyYLTzbNETxT5sdiC6u-pIC_EcKWPaBd2TBEmjxzC6Q'
    },
  };

  const res = http.get(url, params);

  check(res, {
    'status is 200': (r) => r.status === 200,
  });

  sleep(1);
}
