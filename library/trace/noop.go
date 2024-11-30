package trace

import (
    "github.com/SkyAPM/go2sky/propagation"
)

func NoopExtractor() propagation.Extractor {
    return func() (s string, e error) {
        return "", nil
    }
}

func NoopInjector() propagation.Injector {
    return func(header string) error {
        return nil
    }
}
