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
            if_type_name: "IFTypeS5S8SGWGTPC"
        });
    }
    const CauseRequestAccepted = 16;
    const csr_res = client.checkSendCreateSessionRequestS5S8(
        "127.0.0.1:2123",
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
            pdntype: 1,
            epsbearerid: 1,
            uplane_ie: {
                teid: 1,
            },
            ambrul: 100000000,
            ambrdl: 100000000,
        }
    )
    check (csr_res == CauseRequestAccepted, {
        'csr is success': (res) => true === res,
    });

    const mbr_res = client.checkSendModifyBearerRequestS5S8(
        "",
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
        epsbearerid: 100,
        uplane_ie: {
            ip: "1.1.1.5",
            ip6: "fc00::5",
        },
        cplane_sgw_ie:{
            ip:  "1.1.1.4",
            ip6: "fc00::4",
        },
        ambrul: 200000000,
        ambrdl: 200000000,
    })
    check (mbr_res == CauseRequestAccepted, {
        'mbr is success': (res) => true === res,
    });

    // if not exist dst
    const dsr_res = client.checkSendDeleteSessionRequestS5S8(
        "",
        {
            imsi: "123451234567891",
            epsbearerid: 100,
        }
    )
    check (dsr_res == CauseRequestAccepted, {
        'dsr is success': (res) => true === res,
    });
    client.close()
}