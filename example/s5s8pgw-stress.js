import gtpv2 from 'k6/x/gtpv2';

// pseudo sgw
export function setup() {
    gtpv2.connect({
        saddr: 'SRC_IP',
        daddr: "DST_IP",
        count: "RETRY_NUM",
        IFTypeName: "IFTypeS5S8PGWGTPC"
    })
    console.log("setup");
}

export default function () {
    const res = gtpv2.SendCreateSessionRequestS5S8(
        "DST_IP",
        {},
    )
    check (res, {
        'success': (res) => true === res,
    });
    gtpv2.close()
}
