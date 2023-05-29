import gtpv2 from 'k6/x/gtpv2';

let conn;
export function setup() {
    // conn = gtpv2.connect({
    //     saddr: "127.0.0.1:2124",
    //     daddr: "127.0.0.1:2123",
    //     count: 0,
    //     IFTypeName: "IFTypeS5S8PGWGTPC"
    // });
    console.log("setup");
}


export default function () {
    conn = gtpv2.connect({
        saddr: "127.0.0.1:2124",
        daddr: "127.0.0.1:2123",
        count: 0,
        IFTypeName: "IFTypeS5S8PGWGTPC"
    });

    const res = gtpv2.checkSendEchoRequestWithReturnResponse(conn, "127.0.0.1:2123")
    check (res, {
        'success': (res) => true === res,
    });
    gtpv2.close()
}
