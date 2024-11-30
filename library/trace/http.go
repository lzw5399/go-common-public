package trace

import (
	"net/http"

	"github.com/SkyAPM/go2sky/propagation"
)

func HttpExtractor(req *http.Request) propagation.Extractor {
    return func() (s string, e error) {
        if req == nil || req.Header == nil {
            return "", nil
        }
        return req.Header.Get(propagation.Header), nil
    }
}

func HttpInjector(req *http.Request) propagation.Injector {
    return func(header string) error {
        if req == nil || req.Header == nil {
            return nil
        }
        req.Header.Set(propagation.Header, header)
        return nil
    }
}
