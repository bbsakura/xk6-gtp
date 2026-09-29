// Demonstrates the Responder MVP: a k6 script binds a GTPv2-C listener,
// pulls captured Echo Requests off the queue with next_request(), and hands
// back an Echo Response built from gtpv2.msg.* / gtpv2.ie.*.
//
// The driving client Dials the reference PGW (which handles the initial Echo
// handshake), then uses send_raw() to send an Echo Request directly to the
// Responder's address. The Responder captures it into its queue and the
// script writes back an Echo Response from JS.

import { check, sleep } from 'k6';
import exec from 'k6/execution';

import gtpv2 from 'k6/x/gtpv2';

const MSG_TYPE_ECHO_REQUEST = 1;
const MSG_TYPE_ECHO_RESPONSE = 2;

export const options = {
    vus: 1,
    iterations: 1,
};

let responder;
let client;

export default function () {
    if (responder == null) {
        responder = new gtpv2.K6GTPv2Responder({
            listen: `127.0.0.${exec.vu.idInTest}:2224`,
            if_type_name: 'IFTypeS5S8PGWGTPC',
            handle_msg_types: [MSG_TYPE_ECHO_REQUEST],
        });
    }
    if (client == null) {
        client = new gtpv2.K6GTPv2Client();
        client.setTimeout(1);
        // Dial the reference PGW to complete the mandatory Echo handshake
        // without racing our own Responder.
        client.connect({
            saddr: `127.0.0.${exec.vu.idInTest}:2225`,
            daddr: '127.0.0.1:2123',
            count: 0,
            if_type_name: 'IFTypeS5S8SGWGTPC',
        });
    }

    // Fire an echo request directly at the Responder. sendRaw sends to any
    // UDP address, so the target does not need to be the original Dial peer.
    const target = `127.0.0.${exec.vu.idInTest}:2224`;
    const handle = client.sendRaw(
        target,
        gtpv2.msg.newEchoRequest(0, [gtpv2.ie.newRecovery(0)]),
    );
    check(handle, {
        'echo dispatched': (h) => h.ok && h.sequence > 0,
    });

    // Drain one request from the responder queue.
    const req = responder.nextRequest(1000);
    check(req, {
        'responder saw echo request': (r) => r.ok && r.message_type === MSG_TYPE_ECHO_REQUEST,
        'sequence propagated': (r) => r.sequence === handle.sequence,
    });

    // Hand back an Echo Response built from JS. RespondTo aligns TEID and
    // Sequence against the captured request under the hood.
    const err = responder.respondTo(
        req,
        gtpv2.msg.newEchoResponse(0, [gtpv2.ie.newRecovery(1)]),
    );
    check(err, {
        'respond_to succeeded': (e) => e === null || e === undefined,
    });

    // Client should observe the response inside its recv budget.
    sleep(0.05);
    const recv = client.awaitMessage(MSG_TYPE_ECHO_RESPONSE, handle.sequence, 500);
    check(recv, {
        'client received echo response': (r) => r.ok,
    });

    client.close();
    responder.close();
}
