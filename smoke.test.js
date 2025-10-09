import { check } from 'k6';
import exec from 'k6/execution';

import gtpv2 from 'k6/x/gtpv2';

let client;

export default function (){
    if (client == null) {
        client = new gtpv2.K6GTPv2Client();
        try {
        client.connect({
            saddr: `127.0.0.1:2124`,
            daddr: "127.0.0.1:2123",
            count: 0,
            IFTypeName: "IFTypeS5S8PGWGTPC"
        });
        } catch (e) {
            if (e.message.includes("i/o timeout")) {
                return 0;
            }
            throw e;
        }
    }
}
