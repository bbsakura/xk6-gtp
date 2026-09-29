// Session lifecycle sanity check for S5/S8 SGW → PGW: Create Session, Modify
// Bearer, Delete Session must each return Cause = 16 (Request Accepted) and
// stay under a per-step latency budget. Each step is one k6 check so the
// failing step shows up plainly in the run summary.

import { check } from 'k6';
import exec from 'k6/execution';

import gtpv2 from 'k6/x/gtpv2';

const CAUSE_REQUEST_ACCEPTED = 16;

// Per-message latency budgets (ms). Tune per environment.
const BUDGET_CSR_MS = 200;
const BUDGET_MBR_MS = 200;
const BUDGET_DSR_MS = 200;

export const options = {
    vus: 1,
    iterations: 3,
    thresholds: {
        'gtpv2_resp_cause{msg_type:create_session,cause:16}': ['count>=3'],
        'gtpv2_resp_cause{msg_type:delete_session,cause:16}': ['count>=3'],
        'gtpv2_req_duration{msg_type:create_session}': ['p(95)<200'],
        'gtpv2_timeout_total': ['count==0'],
        checks: ['rate==1.00'],
    },
};

let client;

export default function () {
    if (client == null) {
        client = new gtpv2.K6GTPv2Client();
        client.setTimeout(3);
        client.connect({
            saddr: `127.0.0.${exec.vu.idInTest}:2124`,
            daddr: '127.0.0.1:2123',
            count: 0,
            if_type_name: 'IFTypeS5S8SGWGTPC',
        });
    }

    const imsi = `12345123456789${exec.vu.iterationInInstance % 5 + 1}`;
    const base = {
        imsi,
        msisdn: imsi,
        mei: imsi,
        mcc: '123',
        mnc: '123',
        tac: 1,
        rat: 'EUTRAN',
        apn: 'apn',
        eci: 1,
        pdntype: 1,
        epsbearerid: 1,
        ambrul: 100000000,
        ambrdl: 100000000,
        uplane_ie: { teid: 1 },
    };

    const csr = client.tryCreateSessionS5S8('127.0.0.1:2123', base);
    check(csr, {
        'csr accepted': (r) => r.ok && r.cause === CAUSE_REQUEST_ACCEPTED,
        'csr latency ok': (r) => r.elapsed_ms < BUDGET_CSR_MS,
    });

    const mbr = client.tryModifyBearerS5S8('', {
        ...base,
        epsbearerid: 100,
        uplane_ie: { ip: '1.1.1.5', ip6: 'fc00::5' },
        cplane_sgw_ie: { ip: '1.1.1.4', ip6: 'fc00::4' },
        ambrul: 200000000,
        ambrdl: 200000000,
    });
    check(mbr, {
        'mbr accepted': (r) => r.ok && r.cause === CAUSE_REQUEST_ACCEPTED,
        'mbr latency ok': (r) => r.elapsed_ms < BUDGET_MBR_MS,
    });

    const dsr = client.tryDeleteSessionS5S8('', {
        imsi,
        epsbearerid: 100,
    });
    check(dsr, {
        'dsr accepted': (r) => r.ok && r.cause === CAUSE_REQUEST_ACCEPTED,
        'dsr latency ok': (r) => r.elapsed_ms < BUDGET_DSR_MS,
    });
}

export function teardown() {
    if (client) client.close();
}
