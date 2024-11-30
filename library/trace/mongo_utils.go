package trace

import (
    "errors"
    "fmt"
    "net"
    "net/url"
    "strconv"
    "strings"
    "time"
)

type urlInfo struct {
    addrs   []string
    user    string
    pass    string
    db      string
    options map[string]string
}

func isOptSep(c rune) bool {
    return c == ';' || c == '&'
}

func extractURL(s string) (*urlInfo, error) {
    if strings.HasPrefix(s, "mongodb://") {
        s = s[10:]
    }
    info := &urlInfo{options: make(map[string]string)}
    if c := strings.Index(s, "?"); c != -1 {
        for _, pair := range strings.FieldsFunc(s[c+1:], isOptSep) {
            l := strings.SplitN(pair, "=", 2)
            if len(l) != 2 || l[0] == "" || l[1] == "" {
                return nil, errors.New("connection option must be key=value: " + pair)
            }
            info.options[l[0]] = l[1]
        }
        s = s[:c]
    }
    if c := strings.Index(s, "@"); c != -1 {
        pair := strings.SplitN(s[:c], ":", 2)
        if len(pair) > 2 || pair[0] == "" {
            return nil, errors.New("credentials must be provided as user:pass@host")
        }
        var err error
        info.user, err = url.QueryUnescape(pair[0])
        if err != nil {
            return nil, fmt.Errorf("cannot unescape username in URL: %q", pair[0])
        }
        if len(pair) > 1 {
            info.pass, err = url.QueryUnescape(pair[1])
            if err != nil {
                return nil, fmt.Errorf("cannot unescape password in URL")
            }
        }
        s = s[c+1:]
    }
    if c := strings.Index(s, "/"); c != -1 {
        info.db = s[c+1:]
        s = s[:c]
    }
    info.addrs = strings.Split(s, ",")
    return info, nil
}

func ParseMongoURL(url string) (scheme string, host string, path string) {
    dialInfo, err := doParseMongoURL(url)
    if err != nil {
        return "", "", ""
    }
    scheme = "mongodb"
    if dialInfo.Addrs == nil {
        host = ""
    } else {
        host = strings.Join(dialInfo.Addrs, ",")
    }
    path = dialInfo.Database
    return scheme, host, path
}

// ParseURL parses a MongoDB URL as accepted by the Dial function and returns
// a value suitable for providing into DialWithInfo.
//
// See Dial for more details on the format of url.
func doParseMongoURL(url string) (*DialInfo, error) {
    uinfo, err := extractURL(url)
    if err != nil {
        return nil, err
    }
    direct := false
    mechanism := ""
    service := ""
    source := ""
    setName := ""
    poolLimit := 0
    for k, v := range uinfo.options {
        switch k {
        case "authSource":
            source = v
        case "authMechanism":
            mechanism = v
        case "gssapiServiceName":
            service = v
        case "replicaSet":
            setName = v
        case "maxPoolSize":
            poolLimit, err = strconv.Atoi(v)
            if err != nil {
                return nil, errors.New("bad value for maxPoolSize: " + v)
            }
        case "connect":
            if v == "direct" {
                direct = true
                break
            }
            if v == "replicaSet" {
                break
            }
            fallthrough
        default:
            return nil, errors.New("unsupported connection URL option: " + k + "=" + v)
        }
    }
    info := DialInfo{
        Addrs:          uinfo.addrs,
        Direct:         direct,
        Database:       uinfo.db,
        Username:       uinfo.user,
        Password:       uinfo.pass,
        Mechanism:      mechanism,
        Service:        service,
        Source:         source,
        PoolLimit:      poolLimit,
        ReplicaSetName: setName,
    }
    return &info, nil
}

// DialInfo holds options for establishing a session with a MongoDB cluster.
// To use a URL, see the Dial function.
type DialInfo struct {
    // Addrs holds the addresses for the seed servers.
    Addrs []string

    // Direct informs whether to establish connections only with the
    // specified seed servers, or to obtain information for the whole
    // cluster and establish connections with further servers too.
    Direct bool

    // Timeout is the amount of time to wait for a server to respond when
    // first connecting and on follow up operations in the session. If
    // timeout is zero, the call may block forever waiting for a connection
    // to be established. Timeout does not affect logic in DialServer.
    Timeout time.Duration

    // FailFast will cause connection and query attempts to fail faster when
    // the server is unavailable, instead of retrying until the configured
    // timeout period. Note that an unavailable server may silently drop
    // packets instead of rejecting them, in which case it's impossible to
    // distinguish it from a slow server, so the timeout stays relevant.
    FailFast bool

    // Database is the default database name used when the Session.DB method
    // is called with an empty name, and is also used during the initial
    // authentication if Source is unset.
    Database string

    // ReplicaSetName, if specified, will prevent the obtained session from
    // communicating with any server which is not part of a replica set
    // with the given name. The default is to communicate with any server
    // specified or discovered via the servers contacted.
    ReplicaSetName string

    // Source is the database used to establish credentials and privileges
    // with a MongoDB server. Defaults to the value of Database, if that is
    // set, or "admin" otherwise.
    Source string

    // Service defines the service name to use when authenticating with the GSSAPI
    // mechanism. Defaults to "mongodb".
    Service string

    // ServiceHost defines which hostname to use when authenticating
    // with the GSSAPI mechanism. If not specified, defaults to the MongoDB
    // server's address.
    ServiceHost string

    // Mechanism defines the protocol for credential negotiation.
    // Defaults to "MONGODB-CR".
    Mechanism string

    // Username and Password inform the credentials for the initial authentication
    // done on the database defined by the Source field. See Session.Login.
    Username string
    Password string

    // PoolLimit defines the per-server socket pool limit. Defaults to 4096.
    // See Session.SetPoolLimit for details.
    PoolLimit int

    // DialServer optionally specifies the dial function for establishing
    // connections with the MongoDB servers.
    DialServer func(addr *ServerAddr) (net.Conn, error)

    // WARNING: This field is obsolete. See DialServer above.
    Dial func(addr net.Addr) (net.Conn, error)
}

// ServerAddr represents the address for establishing a connection to an
// individual MongoDB server.
type ServerAddr struct {
    str string
    tcp *net.TCPAddr
}
