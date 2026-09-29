// Abnormal-cause reproducer. Uses the Responder to simulate a peer that
// answers Create Session Requests with a rejection Cause, then verifies the
// client extracts and reports that Cause exactly. This is the failure-mode
// equivalent of the session-health probe: instead of asking "did the good
// case go through", it asks "if the peer rejects, can we tell why".
//
// Handy for wiring up alerting rules: emit a rejection intentionally in a
// staging run and confirm the metric / log path picks it up before relying
// on the same signals in production.

import { check, sleep } from 'k6';
import exec from 'k6/execution';

import gtpv2 from 'k6/x/gtpv2';

// GTPv2 message and Cause codes (3GPP TS 29.274 clauses 6.1 and 8.4).
const MSG_TYPE_CREATE_SESSION_REQUEST = 32;
const MSG_TYPE_CREATE_SESSION_RESPONSE = 33;
const CAUSE_REQUEST_ACCEPTED = 16;
const CAUSE_CONTEXT_NOT_FOUND = 64;

// Interface types used by the Responder-driven abnormal peer.
const IF_TYPE_SGW_C = gtpv2.IFType.S5S8_SGW_GTPC;
const IF_TYPE_SGW_U = gtpv2.IFType.S5S8_SGW_GTPU;

export const options = {
    vus: 1,
    iterations: 1,
    thresholds: {
        // Assert that at least one Cause=64 response is observed. If the
        // extension were to drop the Cause tag we would see this hit zero.
        'gtpv2_resp_cause{msg_type:create_session,cause:64}': ['count>0'],
        checks: ['rate==1.00'],
    },
};

let responder;
let client;

export default function () {
    if (responder == null) {
        responder = new gtpv2.K6GTPv2Responder({
            listen: `127.0.0.${exec.vu.idInTest}:9123`,
            if_type_name: 'IFTypeS5S8PGWGTPC',
            handle_msg_types: [MSG_TYPE_CREATE_SESSION_REQUEST],
        });
    }
    if (client == null) {
        client = new gtpv2.K6GTPv2Client();
        client.setTimeout(1);
        // Dial the reference PGW to satisfy Dial's built-in Echo handshake
        // and leave the client's read loop running.
        client.connect({
            saddr: `127.0.0.${exec.vu.idInTest}:9124`,
            daddr: '127.0.0.1:2123',
            count: 0,
            if_type_name: 'IFTypeS5S8SGWGTPC',
        });
    }

    const { ie, msg } = gtpv2;
    const target = `127.0.0.${exec.vu.idInTest}:9123`;

    const imsi = '123451234567891';
    const csrIEs = [
        ie.newIMSI(imsi),
        ie.newMSISDN(imsi),
        ie.newMobileEquipmentIdentity(imsi),
        ie.newAccessPointName('apn'),
        ie.newServingNetwork('123', '123'),
        ie.newRATType(6),
        ie.newFullyQualifiedTEID(IF_TYPE_SGW_C, 1, `127.0.0.${exec.vu.idInTest}`, ''),
        ie.newSelectionMode(0),
        ie.newPDNType(1),
        ie.newPDNAddressAllocation('0.0.0.0'),
        ie.newAPNRestriction(2),
        ie.newAggregateMaximumBitRate(100000000, 100000000),
        ie.newBearerContext([
            ie.newEPSBearerID(1),
            ie.newFullyQualifiedTEID(IF_TYPE_SGW_U, 1, `127.0.0.${exec.vu.idInTest}`, ''),
            ie.newBearerQoS(1, 2, 1, 0xff, 0, 0, 0, 0),
        ]),
    ];
    const csr = msg.newCreateSessionRequest(0, 0, csrIEs);
    const dispatched = client.sendRaw(target, csr);
    check(dispatched, {
        'csr dispatched to abnormal peer': (h) => h.ok && h.sequence > 0,
    });

    // Responder drains the CSR, replies with Cause = Context Not Found.
    const captured = responder.nextRequest(1000);
    check(captured, {
        'responder captured csr': (r) => r.ok && r.message_type === MSG_TYPE_CREATE_SESSION_REQUEST,
    });

    const rejection = msg.newCreateSessionResponse(0, 0, [
        ie.newCause(CAUSE_CONTEXT_NOT_FOUND, 0, 0, 0),
    ]);
    const respErr = responder.respondTo(captured, rejection);
    check(respErr, {
        'respond_to succeeded': (e) => e === null || e === undefined,
    });

    // Client should see the rejection reflected in the AwaitResult.
    sleep(0.05);
    const observed = client.awaitMessage(
        MSG_TYPE_CREATE_SESSION_RESPONSE,
        dispatched.sequence,
        500,
    );
    check(observed, {
        'client observed csr response': (r) => r.ok,
        'cause is rejection not accepted': (r) => r.cause !== CAUSE_REQUEST_ACCEPTED,
        'cause is exactly context-not-found': (r) => r.cause === CAUSE_CONTEXT_NOT_FOUND,
    });

    client.close();
    responder.close();
}
