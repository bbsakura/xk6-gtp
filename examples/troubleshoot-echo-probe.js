// Node monitoring / uplink health check: probe a peer with Echo Request on a
// fixed cadence and fail the run if latency degrades or requests time out.
//
// Metric-driven guardrails:
//   - gtpv2_req_duration{msg_type:"echo"} p95 must stay under 200 ms
//   - gtpv2_timeout_total{msg_type:"echo"} must stay at zero
//
// Turn this into a continuous probe by giving k6 a duration (`k6 run -d 10m`)
// or replacing the fixed iterations option with a `ramping-arrival-rate`
// executor.

import { check } from 'k6';
import exec from 'k6/execution';

import gtpv2 from 'k6/x/gtpv2';

export const options = {
    vus: 1,
    iterations: 5,
    thresholds: {
        'gtpv2_req_duration{msg_type:echo}': ['p(95)<200'],
        'gtpv2_timeout_total{msg_type:echo}': ['count==0'],
        checks: ['rate==1.00'],
    },
};

let client;

export default function () {
    if (client == null) {
        client = new gtpv2.K6GTPv2Client();
        client.setTimeout(2);
        client.connect({
            saddr: `127.0.0.${exec.vu.idInTest}:2124`,
            daddr: '127.0.0.1:2123',
            count: 0,
            if_type_name: 'IFTypeS5S8SGWGTPC',
        });
    }

    const r = client.tryEcho('127.0.0.1:2123');
    check(r, {
        'echo returned': (res) => res.ok,
        'no timeout': (res) => !res.timeout,
        'elapsed under 200ms': (res) => res.elapsed_ms < 200,
    });
}

export function teardown() {
    if (client) client.close();
}
