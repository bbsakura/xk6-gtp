import gtpv2 from 'k6/x/gtpv2';

export function setup() {
    gtpv2.connect({
        saddr: 'SRC_IP',
        daddr: "DST_IP",
        count: "RETRY_NUM",
        IFTypeName: "IFTypeS11MMEGTPC"
    })
    console.log("setup");
}


export default function () {
    const res = gtpv2.send_echo_request_with_check_echo_response("DST_IP")
    check (res, {
        'success': (res) => true === res,
    });
    gtpv2.close()
}
