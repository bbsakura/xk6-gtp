import { check } from 'k6';
import gtpv2 from 'k6/x/gtpv2';
const client = new gtpv2.K6GTPv2Client();

export default function (){
    client.connect({
        saddr: "127.0.0.1:2124",
        daddr: "127.0.0.1:2123",
        count: 0,
        IFTypeName: "IFTypeS5S8PGWGTPC"
    });
    const res = client.checkSendEchoRequestWithReturnResponse("127.0.0.1:2123")
    check (res, {
        'success': (res) => true === res,
    });
    client.close()
}
