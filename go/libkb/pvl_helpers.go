// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package libkb

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	keybase1 "github.com/keybase/client/go/protocol"
	jsonw "github.com/keybase/go-jsonw"
)

func pvlStringToService(service string) (keybase1.ProofType, ProofError) {
	switch service {
	case "twitter":
		return keybase1.ProofType_TWITTER, nil
	case "github":
		return keybase1.ProofType_GITHUB, nil
	case "reddit":
		return keybase1.ProofType_REDDIT, nil
	case "coinbase":
		return keybase1.ProofType_COINBASE, nil
	case "hackernews":
		return keybase1.ProofType_HACKERNEWS, nil
	case "dns":
		return keybase1.ProofType_DNS, nil
	case "rooter":
		return keybase1.ProofType_ROOTER, nil
	case "web":
		return keybase1.ProofType_GENERIC_WEB_SITE, nil
	default:
		return 0, NewProofError(keybase1.ProofStatus_INVALID_PVL,
			"Unsupported service %v", service)
	}
}

func pvlServiceToString(service keybase1.ProofType) (string, ProofError) {
	// This is not quite the same as RemoteServiceTypes due to http/https.
	switch service {
	case keybase1.ProofType_TWITTER:
		return "twitter", nil
	case keybase1.ProofType_GITHUB:
		return "github", nil
	case keybase1.ProofType_REDDIT:
		return "reddit", nil
	case keybase1.ProofType_COINBASE:
		return "coinbase", nil
	case keybase1.ProofType_HACKERNEWS:
		return "hackernews", nil
	case keybase1.ProofType_DNS:
		return "dns", nil
	case keybase1.ProofType_ROOTER:
		return "rooter", nil
	case keybase1.ProofType_GENERIC_WEB_SITE:
		return "web", nil
	default:
		return "", NewProofError(keybase1.ProofStatus_INVALID_PVL,
			"Unsupported service %v", service)
	}
}

func pvlJSONHasKey(w *jsonw.Wrapper, key string) bool {
	return !w.AtKey(key).IsNil()
}

func pvlJSONUnpackArray(w *jsonw.Wrapper) ([]*jsonw.Wrapper, error) {
	w, err := w.ToArray()
	if err != nil {
		return nil, err
	}
	length, err := w.Len()
	if err != nil {
		return nil, err
	}
	res := make([]*jsonw.Wrapper, length)
	for i := 0; i < length; i++ {
		res[i] = w.AtIndex(i)
	}
	return res, nil
}

func pvlJSONGetChildren(w *jsonw.Wrapper) ([]*jsonw.Wrapper, error) {
	dict, err := w.ToDictionary()
	isDict := err == nil
	array, err := w.ToArray()
	isArray := err == nil

	switch {
	case isDict:
		keys, err := dict.Keys()
		if err != nil {
			return nil, err
		}
		var res = make([]*jsonw.Wrapper, len(keys))
		for i, key := range keys {
			res[i] = dict.AtKey(key)
		}
		return res, nil
	case isArray:
		return pvlJSONUnpackArray(array)
	default:
		return nil, errors.New("got children of non-container")
	}
}

func pvlJSONStringOrMarshal(object *jsonw.Wrapper) (string, error) {
	s, err := object.GetString()
	if err == nil {
		return s, nil
	}
	b, err := object.Marshal()
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// Get the HTML contents of all elements in a selection, concatenated by a space.
func pvlSelectionContents(selection *goquery.Selection, useAttr bool, attr string) (string, error) {
	len := selection.Length()
	results := make([]string, len)
	errs := make([]error, len)
	var wg sync.WaitGroup
	wg.Add(len)
	selection.Each(func(i int, element *goquery.Selection) {
		if useAttr {
			res, ok := element.Attr(attr)
			results[i] = res
			if !ok {
				errs[i] = fmt.Errorf("Could not get attr %v of element", attr)
			}
		} else {
			results[i] = element.Text()
			errs[i] = nil
		}
		wg.Done()
	})
	wg.Wait()
	for _, err := range errs {
		if err != nil {
			return "", err
		}
	}
	return strings.Join(results, " "), nil
}
