// Demonstrates building GTPv2-C IEs from JavaScript via gtpv2.ie.* and
// dispatching them through the raw send API.
//
// The happy-path branch composes a complete Create Session Request for an
// IMSI the reference PGW recognises. The abnormal-path branch uses an IMSI
// the PGW does not know about, so it drops the request without responding
// and the client observes a receive timeout — the same signal you would
// see from a peer that failed to build a Create Session Response.

import { check } from 'k6';
import exec from 'k6/execution';

import gtpv2 from 'k6/x/gtpv2';

const CAUSE_REQUEST_ACCEPTED = 16;

// Interface types from 3GPP TS 29.274 clause 8.22.
const IF_TYPE_S5S8_SGW_GTPC = 6;
const IF_TYPE_S5S8_SGW_GTPU = 4;

let client;

export default function () {
    if (client == null) {
        client = new gtpv2.K6GTPv2Client();
        client.setTimeout(1); // seconds — keep the timeout arm short for the miss case
        client.connect({
            saddr: `127.0.0.${exec.vu.idInTest}:2124`,
            daddr: '127.0.0.1:2123',
            count: 0,
            if_type_name: 'IFTypeS5S8SGWGTPC',
        });
    }

    const { ie } = gtpv2;

    // Happy-path: fully-populated Create Session Request assembled from JS.
    const imsi = '123451234567891';
    const happyIEs = [
        ie.newIMSI(imsi),
        ie.newMSISDN(imsi),
        ie.newMobileEquipmentIdentity(imsi),
        ie.newAccessPointName('apn'),
        ie.newServingNetwork('123', '123'),
        ie.newRATType(6), // EUTRAN
        ie.newFullyQualifiedTEID(IF_TYPE_S5S8_SGW_GTPC, 1, '127.0.0.1', ''),
        ie.newSelectionMode(0),
        ie.newPDNType(1),
        ie.newPDNAddressAllocation('0.0.0.0'),
        ie.newAPNRestriction(2),
        ie.newAggregateMaximumBitRate(100000000, 100000000),
        ie.newBearerContext([
            ie.newEPSBearerID(1),
            ie.newFullyQualifiedTEID(IF_TYPE_S5S8_SGW_GTPU, 1, '127.0.0.1', ''),
            ie.newBearerQoS(1, 2, 1, 0xff, 0, 0, 0, 0),
        ]),
    ];
    const happy = client.tryCreateSessionRaw('127.0.0.1:2123', imsi, happyIEs);
    check(happy, {
        'raw csr accepted': (r) => r.ok && r.cause === CAUSE_REQUEST_ACCEPTED,
    });

    // Abnormal path: use an IMSI the reference PGW does not know about, so it
    // drops the request without responding. We expect ok=false + timeout=true.
    const unknownImsi = '999999999999999';
    const abnormalIEs = [
        ie.newIMSI(unknownImsi),
        ie.newMSISDN(unknownImsi),
        ie.newMobileEquipmentIdentity(unknownImsi),
        ie.newAccessPointName('apn'),
        ie.newServingNetwork('123', '123'),
        ie.newRATType(6),
        ie.newFullyQualifiedTEID(IF_TYPE_S5S8_SGW_GTPC, 2, '127.0.0.1', ''),
        ie.newSelectionMode(0),
        ie.newPDNType(1),
        ie.newPDNAddressAllocation('0.0.0.0'),
        ie.newAPNRestriction(2),
        ie.newAggregateMaximumBitRate(100000000, 100000000),
        ie.newBearerContext([
            ie.newEPSBearerID(1),
            ie.newFullyQualifiedTEID(IF_TYPE_S5S8_SGW_GTPU, 2, '127.0.0.1', ''),
            ie.newBearerQoS(1, 2, 1, 0xff, 0, 0, 0, 0),
        ]),
    ];
    const abnormal = client.tryCreateSessionRaw('127.0.0.1:2123', unknownImsi, abnormalIEs);
    check(abnormal, {
        'unknown IMSI produces timeout': (r) => !r.ok && r.timeout === true,
    });

    client.close();
}
