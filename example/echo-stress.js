import { check } from 'k6';
import exec from 'k6/execution';

import gtpv2 from 'k6/x/gtpv2';

let client;

export default function (){
    if (client == null) {
        client = new gtpv2.K6GTPv2Client();
        client.connect({
            saddr: `127.0.0.${exec.vu.idInTest}:2124`,
            daddr: "127.0.0.1:2123",
            count: 0,
            IFTypeName: "IFTypeS5S8PGWGTPC"
        });
    }
    const res = client.checkSendEchoRequestWithReturnResponse("127.0.0.1:2123")
    check (res, {
        'success': (res) => true === res,
    });
}
