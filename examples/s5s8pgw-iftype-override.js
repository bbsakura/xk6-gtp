// Demonstrates configuring FTEID interface types explicitly through the
// S5S8SgwParams IfType overrides. Passing the S5/S8 SGW values yields the
// same wire format as the defaults; swap in S11 / S8 / S2b values from the
// gtpv2.IFType map to drive the same helpers against a different peer role.

import { check } from 'k6';
import exec from 'k6/execution';

import gtpv2 from 'k6/x/gtpv2';

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

        // Explicit interface-type overrides. Zero would fall back to the
        // S5/S8 SGW defaults; set them from gtpv2.IFType.* to point the
        // FTEIDs at a different reference point.
        iftype_cplane: gtpv2.IFType.S5S8_SGW_GTPC,
        iftype_uplane: gtpv2.IFType.S5S8_SGW_GTPU,
        iftype_cplane_peer: gtpv2.IFType.S5S8_PGW_GTPC,
    });
    check(csr, {
        'csr accepted with explicit IfTypes': (r) =>
            r.ok && r.cause === CAUSE_REQUEST_ACCEPTED,
    });

    const dsr = client.tryDeleteSessionS5S8('', {
        imsi: '123451234567891',
        epsbearerid: 1,
        iftype_cplane_peer: gtpv2.IFType.S5S8_PGW_GTPC,
    });
    check(dsr, {
        'dsr accepted with explicit peer IfType': (r) =>
            r.ok && r.cause === CAUSE_REQUEST_ACCEPTED,
    });

    client.close();
}
