// Demonstrates awaitMessageAsync for issuing multiple GTPv2 requests in
// parallel and collecting all responses via Promise.all. Correlation is per
// (msgType, sequence) inside the sending Client's sessions map, so responses
// route back to the right Promise even when they arrive out of order.
//
// This is the xk6-diameter-style pattern for concurrent request/response —
// scripts stay reactive instead of blocking one iteration on a serial wait.

import { check } from 'k6';
import exec from 'k6/execution';

import gtpv2 from 'k6/x/gtpv2';

const MSG_TYPE_ECHO_RESPONSE = 2;

export const options = {
    vus: 1,
    iterations: 1,
    thresholds: {
        checks: ['rate==1.00'],
        'gtpv2_req_duration{msg_type:echo}': ['p(95)<200'],
    },
};

let client;

export default async function () {
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

    // Fan out three echo requests without waiting for any response.
    const target = '127.0.0.1:2123';
    const handles = [];
    for (let i = 0; i < 3; i++) {
        const req = gtpv2.msg.newEchoRequest(0, [gtpv2.ie.newRecovery(0)]);
        handles.push(client.sendRaw(target, req));
    }
    check(handles, {
        'all echoes dispatched': (hs) => hs.every((h) => h.ok && h.sequence > 0),
    });

    // Collect responses concurrently. Each Promise correlates by
    // (MSG_TYPE_ECHO_RESPONSE, sequence) and resolves as soon as its
    // paired response lands in the Client's sessions map — order-independent.
    const results = await Promise.all(
        handles.map((h) => client.awaitMessageAsync(MSG_TYPE_ECHO_RESPONSE, h.sequence, 1000)),
    );
    check(results, {
        'all echoes returned': (rs) => rs.every((r) => r.ok),
        'no async timeouts': (rs) => rs.every((r) => !r.timeout),
        'each response under 200ms': (rs) => rs.every((r) => r.elapsed_ms < 200),
    });

    client.close();
}
