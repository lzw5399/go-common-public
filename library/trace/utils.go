package trace

import (
    "net/url"
    "sync"
)

func ParseURL(uri string) (scheme string, host string, path string) {
    u, err := url.Parse(uri)
    if err != nil {
        return "", "", ""
    }
    if u.Opaque != "" {
        // eg. jdbc:mysql://test_user:ouupppssss@localhost:3306/sakila?profileSQL=true
        u, err := url.Parse(u.Opaque)
        if err != nil {
            return "", "", ""
        }
        return u.Scheme, u.Host, u.Path
    }
    return u.Scheme, u.Host, u.Path
}

var (
    mutex      sync.Mutex
    cache      sync.Map
    mongoCache sync.Map
)

type urlParsedInfo struct {
    scheme string
    host   string
    path   string
}

func ParseURLWithCache(uri string) (scheme string, host string, path string) {
    v, ok := cache.Load(uri)
    if ok {
        info := v.(*urlParsedInfo)
        return info.scheme, info.host, info.path
    }

    mutex.Lock()
    defer mutex.Unlock()
    v, ok = cache.Load(uri)
    if ok {
        info := v.(*urlParsedInfo)
        return info.scheme, info.host, info.path
    }

    scheme, host, path = ParseURL(uri)
    info := urlParsedInfo{scheme: scheme, host: host, path: path}
    cache.Store(uri, &info)
    return info.scheme, info.host, info.path
}

func ParseMongoURLWithCache(uri string) (scheme string, host string, path string) {
    v, ok := mongoCache.Load(uri)
    if ok {
        info := v.(*urlParsedInfo)
        return info.scheme, info.host, info.path
    }

    mutex.Lock()
    defer mutex.Unlock()
    v, ok = mongoCache.Load(uri)
    if ok {
        info := v.(*urlParsedInfo)
        return info.scheme, info.host, info.path
    }

    scheme, host, path = ParseMongoURL(uri)
    info := urlParsedInfo{scheme: scheme, host: host, path: path}
    mongoCache.Store(uri, &info)
    return info.scheme, info.host, info.path
}
