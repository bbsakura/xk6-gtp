import { check } from 'k6';
import exec from 'k6/execution';

import gtpv2 from 'k6/x/gtpv2';

let client;

export default function (){
    if (client == null) {
        client = new gtpv2.K6GTPv2Client();
        client.connect({
            saddr: `127.0.0.${exec.vu.idInTest}:2124`,
            daddr: "127.0.0.1:2125",
            count: 0,
            IFTypeName: "IFTypeS5S8PGWGTPC"
        });
    }
    const res = client.checkSendCreateSessionRequestS5S8(
        "127.0.0.1:2125",
         {
            imsi: "123451234567891",
            msisdn: "123451234567891",
            mei: "123451234567891",
            mcc: "123",
            mnc: "123",
            tac: 1,
            rat: "EUTRAN",
            apn: "apn",
            eci: 1,
            epsbearerid: 1,
            uplaneteid: 1,
            ambrul: 100000000,
            ambrdl: 100000000,
        }
    )
    check (res, {
        'success': (res) => true === res,
    });
    client.close()
}
