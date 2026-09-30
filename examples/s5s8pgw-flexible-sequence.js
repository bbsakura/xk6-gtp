// Demonstrates the Try*/SendResult API added for flexible sequence testing.
//
// Each Try* call returns a SendResult exposing `ok`, `sequence`, `cause`,
// `elapsed_ms`, and `timeout` so scripts can gate on Cause values, budget
// per-message latency, and distinguish send failures from receive timeouts
// without wrapping errors in try/catch.

import { check } from 'k6';
import exec from 'k6/execution';

import gtpv2 from 'k6/x/gtpv2';

export const options = {
    vus: 1,
    iterations: 1,
    thresholds: {
        // Metric-driven guardrails; the metrics ship with the extension.
        gtpv2_req_duration: ['p(95)<200'],
        'gtpv2_resp_cause{cause:16}': ['count>0'],
    },
};

const CAUSE_REQUEST_ACCEPTED = 16;

let client;

export default function () {
    if (client == null) {
        client = new gtpv2.K6GTPv2Client();
        client.connect({
            saddr: `127.0.0.${exec.vu.idInTest}:2124`,
            daddr: '127.0.0.1:2123',
            count: 0,
            if_type_name: 'IFTypeS5S8SGWGTPC',
        });
    }

    const csr = client.tryCreateSessionS5S8('127.0.0.1:2123', {
        imsi: '123451234567891',
        msisdn: '123451234567891',
        mei: '123451234567891',
        mcc: '123',
        mnc: '123',
        tac: 1,
        rat: 'EUTRAN',
        apn: 'apn',
        eci: 1,
        pdntype: 1,
        epsbearerid: 1,
        uplane_ie: { teid: 1 },
        ambrul: 100000000,
        ambrdl: 100000000,
    });
    check(csr, {
        'csr accepted': (r) => r.ok && r.cause === CAUSE_REQUEST_ACCEPTED,
        'csr under 200ms': (r) => r.elapsed_ms < 200,
        'csr did not time out': (r) => !r.timeout,
    });

    const mbr = client.tryModifyBearerS5S8('', {
        imsi: '123451234567891',
        msisdn: '123451234567891',
        mei: '123451234567891',
        mcc: '123',
        mnc: '123',
        tac: 1,
        rat: 'EUTRAN',
        apn: 'apn',
        eci: 1,
        epsbearerid: 100,
        uplane_ie: { ip: '1.1.1.5', ip6: 'fc00::5' },
        cplane_sgw_ie: { ip: '1.1.1.4', ip6: 'fc00::4' },
        ambrul: 200000000,
        ambrdl: 200000000,
    });
    check(mbr, {
        'mbr accepted': (r) => r.ok && r.cause === CAUSE_REQUEST_ACCEPTED,
    });

    const dsr = client.tryDeleteSessionS5S8('', {
        imsi: '123451234567891',
        epsbearerid: 100,
    });
    check(dsr, {
        'dsr accepted': (r) => r.ok && r.cause === CAUSE_REQUEST_ACCEPTED,
    });

    client.close();
}
