// Verifies the gtpv2_conn_reconnect_total counter fires when a Client's
// underlying gtpv2.Conn is torn down and re-established. Useful as a
// regression guard on the reconnect signal that dashboards rely on.
//
// The test intentionally reconnects twice in the second iteration so the
// counter climbs to 2 by end-of-run. The threshold below asserts that.

import { check } from 'k6';
import exec from 'k6/execution';

import gtpv2 from 'k6/x/gtpv2';

export const options = {
    vus: 1,
    iterations: 1,
    thresholds: {
        // Two intentional reconnects; threshold catches regression where the
        // metric silently stops firing.
        'gtpv2_conn_reconnect_total': ['count==2'],
        checks: ['rate==1.00'],
    },
};

export default function () {
    const daddr = '127.0.0.1:2123';
    const saddr = `127.0.0.${exec.vu.idInTest}:2124`;

    const client = new gtpv2.K6GTPv2Client();
    client.setTimeout(1);
    client.connect({
        saddr,
        daddr,
        count: 0,
        if_type_name: 'IFTypeS5S8SGWGTPC',
    });

    const first = client.tryEcho(daddr);
    check(first, { 'echo after initial connect': (r) => r.ok });

    // First reconnect.
    client.close();
    client.connect({
        saddr,
        daddr,
        count: 0,
        if_type_name: 'IFTypeS5S8SGWGTPC',
    });
    const second = client.tryEcho(daddr);
    check(second, { 'echo after 1st reconnect': (r) => r.ok });

    // Second reconnect.
    client.close();
    client.connect({
        saddr,
        daddr,
        count: 0,
        if_type_name: 'IFTypeS5S8SGWGTPC',
    });
    const third = client.tryEcho(daddr);
    check(third, { 'echo after 2nd reconnect': (r) => r.ok });

    client.close();
}
