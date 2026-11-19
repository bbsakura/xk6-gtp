// Demonstrates the fully-generic SendRaw / AwaitMessage split so scripts can
// dispatch any GTPv2 message assembled from `gtpv2.msg.*` and `gtpv2.ie.*`
// without a matching typed helper in Go.
//
// Uses Echo Request/Response because the reference PGW implements it; the
// same pattern works for any message the peer under test handles (Modify
// Bearer Command, Bearer Resource Command, Downlink Data Notification, ...).

import { check } from 'k6';
import exec from 'k6/execution';

import gtpv2 from 'k6/x/gtpv2';

// Message type numbers from 3GPP TS 29.274 clause 6.1.
const MSG_TYPE_ECHO_RESPONSE = 2;

let client;

export default function () {
    if (client == null) {
        client = new gtpv2.K6GTPv2Client();
        client.setTimeout(1);
        client.connect({
            saddr: `127.0.0.${exec.vu.idInTest}:2124`,
            daddr: '127.0.0.1:2123',
            count: 0,
            if_type_name: 'IFTypeS5S8SGWGTPC',
        });
    }

    const { ie, msg } = gtpv2;

    const echoReq = msg.newEchoRequest(0, [ie.newRecovery(0)]);
    const handle = client.sendRaw('127.0.0.1:2123', echoReq);
    check(handle, {
        'sendRaw returns sequence': (h) => h.ok && h.sequence > 0,
    });

    const result = client.awaitMessage(MSG_TYPE_ECHO_RESPONSE, handle.sequence, 1000);
    check(result, {
        'echo response arrives': (r) => r.ok,
        'echo response under 200ms': (r) => r.elapsed_ms < 200,
        'no timeout': (r) => !r.timeout,
    });

    client.close();
}
