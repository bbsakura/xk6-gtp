import { check } from 'k6';
import exec from 'k6/execution';
import gtpv2 from 'k6/x/gtpv2';

// sgw:8, concurrent:100 senario (vut:100)
function get_random(n, m){
   return Math.floor(Math.random() * (m + 1 - n)) + n;
};

const sgwip_addr_list = [
    "127.0.0.1",
    "127.0.0.2",
    "127.0.0.3",
    "127.0.0.4",
    "127.0.0.5",
    "127.0.0.6",
    "127.0.0.7",
    "127.0.0.8",
];

const client = {};
export default function (){
    const saddr = `${sgwip_addr_list[Math.floor(exec.vu.idInTest%sgwip_addr_list.length)]}:2124`;
    if (!(`${saddr}` in client)) {
        console.log(saddr);
        let opts = {
            saddr: saddr,
            daddr: "127.0.0.1:2123",
            count: 3,
            if_type_name: "IFTypeS5S8SGWGTPC"
        }
        client[saddr] = new gtpv2.K6GTPv2ClientWithConnect(opts);
        client[saddr].setTimeout(10);
    }
    const randkey = get_random(0, 10000)
    const teid = 10000 + randkey;
    const imsi = 123451234567890 + randkey;
    const csr_res = client[saddr].checkSendCreateSessionRequestS5S8(
        "127.0.0.1:2123",
        {
            imsi: `${imsi}`,
            msisdn: "123451234567891",
            mei: "123451234567891",
            mcc: "123",
            mnc: "123",
            tac: 1,
            rat: "EUTRAN",
            apn: "ocx",
            eci: 1,
            pdntype: 1,
            epsbearerid: 1,
            uplane_ie: {
                teid: teid,
            },
            ambrul: 100000000,
            ambrdl: 100000000,
        }
    )
    check (csr_res, {
        'csr is success': (res) => true === res,
    });
}

export function teardown(data) {
    for (let key in client) {
        client[key].close();
        client[key] = null;
    }
  }
