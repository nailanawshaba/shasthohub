// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package libkb

import (
	b64 "encoding/base64"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	keybase1 "github.com/keybase/client/go/protocol"
	jsonw "github.com/keybase/go-jsonw"
)

// UsePvl says whether to use PVL for verifying proofs.
const UsePvl = false

// PvlSupportedVersion is which version of PVL is supported by this client.
const PvlSupportedVersion int = 1

type PvlScriptState struct {
	PC           int32
	Service      keybase1.ProofType
	Vars         PvlScriptVariables
	ActiveString string
	FetchURL     string
	HasFetched   bool
	// nil until fetched
	FetchResult *PvlFetchResult
}

type PvlScriptVariables struct {
	UsernameService string
	UsernameKeybase string
	Sig             []byte
	SigIDMedium     string
	SigIDShort      string
	Hostname        string
}

type PvlFetchResult struct {
	Mode PvlMode
	// One of these 3 must be filled.
	String string
	HTML   *goquery.Document
	JSON   *jsonw.Wrapper
}

type PvlMode string

const (
	PvlModeJSON   PvlMode = "json"
	PvlModeHTML   PvlMode = "html"
	PvlModeString PvlMode = "string"
	PvlModeDNS    PvlMode = "dns"
)

const (
	PvlAssertRegexMatch    = "assert_regex_match"
	PvlAssertFindBase64    = "assert_find_base64"
	PvlWhitespaceNormalize = "whitespace_normalize"
	PvlRegexCapture        = "regex_capture"
	PvlFetch               = "fetch"
	PvlSelectorJSON        = "selector_json"
	PvlSelectorCSS         = "selector_css"
	PvlTransformURL        = "transform_url"
)

type PvlStep func(*GlobalContext, *jsonw.Wrapper, PvlScriptState) (PvlScriptState, ProofError)

var PvlSteps = map[string]PvlStep{
	PvlAssertRegexMatch:    pvlStepAssertRegexMatch,
	PvlAssertFindBase64:    pvlStepAssertFindBase64,
	PvlWhitespaceNormalize: pvlStepWhitespaceNormalize,
	PvlRegexCapture:        pvlStepRegexCapture,
	PvlFetch:               pvlStepFetch,
	PvlSelectorJSON:        pvlStepSelectorJSON,
	PvlSelectorCSS:         pvlStepSelectorCSS,
	PvlTransformURL:        pvlStepTransformURL,
}

func CheckProof(g *GlobalContext, pvl *jsonw.Wrapper, service keybase1.ProofType, link RemoteProofChainLink, h SigHint) ProofError {
	if perr := pvlValidateChunk(pvl, service); perr != nil {
		return perr
	}

	sigBody, sigID, err := OpenSig(link.GetArmoredSig())
	if err != nil {
		return NewProofError(keybase1.ProofStatus_BAD_SIGNATURE,
			"Bad signature: %v", err)
	}

	scripts, perr := pvlChunkGetScripts(pvl, service)
	if perr != nil {
		return perr
	}

	newstate := func() PvlScriptState {
		vars := PvlScriptVariables{
			UsernameService: link.GetRemoteUsername(), // Blank for DNS-proofs
			UsernameKeybase: link.GetUsername(),
			Sig:             sigBody,
			SigIDMedium:     sigID.ToMediumID(),
			SigIDShort:      sigID.ToShortID(),
			Hostname:        link.GetHostname(), // Blank for non-{DNS/Web} proofs
		}

		// Enforce prooftype-dependent variables.
		webish := (service == keybase1.ProofType_DNS || service == keybase1.ProofType_GENERIC_WEB_SITE)
		if webish {
			vars.UsernameService = ""
		} else {
			vars.Hostname = ""
		}

		state := PvlScriptState{
			PC:           0,
			Service:      service,
			Vars:         vars,
			ActiveString: h.apiURL,
			FetchURL:     h.apiURL,
			HasFetched:   false,
			FetchResult:  nil,
		}
		return state
	}

	if service == keybase1.ProofType_DNS {
		perr = pvlRunDNS(g, scripts, newstate())
		if perr != nil {
			return perr
		}
		return nil
	}

	// Run the scripts in order.
	// If any succeed, the proof succeeds.
	// If one fails, the next takes over.
	// If all fail, report only the first error.
	var errs []ProofError
	for _, script := range scripts {
		perr = pvlRunScript(g, script, newstate())
		if perr == nil {
			return nil
		}
		errs = append(errs, perr)
	}
	return errs[0]
}

// Get the list of scripts for a given service.
func pvlChunkGetScripts(pvl *jsonw.Wrapper, service keybase1.ProofType) ([]*jsonw.Wrapper, ProofError) {
	serviceString, perr := pvlServiceToString(service)
	if perr != nil {
		return nil, perr
	}
	scriptsw, err := pvl.AtKey("services").AtKey(serviceString).ToArray()
	if err != nil {
		return nil, NewProofError(keybase1.ProofStatus_INVALID_PVL,
			"PVL script is not an array: %v", err)
	}

	// Check if pvl[services][service][0] is an array. If it is, this an OR of multiple scripts.
	_, err = scriptsw.AtIndex(0).ToArray()
	multiscript := err == nil
	var scripts []*jsonw.Wrapper
	if multiscript {
		scripts, err = pvlJSONUnpackArray(scriptsw)
		if err != nil {
			return nil, NewProofError(keybase1.ProofStatus_INVALID_PVL,
				"Could not unpack PVL multiscript: %v", err)
		}
	} else {
		scripts = []*jsonw.Wrapper{scriptsw}
	}
	if len(scripts) < 1 {
		return nil, NewProofError(keybase1.ProofStatus_INVALID_PVL,
			"Empty script list")
	}
	return scripts, nil
}

// Check that a chunk of PVL is valid code.
// Will always accept valid code, may not always notice invalidities.
func pvlValidateChunk(pvl *jsonw.Wrapper, service keybase1.ProofType) ProofError {
	// Check the version.
	version, err := pvl.AtKey("pvl_version").GetInt()
	if err != nil {
		return NewProofError(keybase1.ProofStatus_INVALID_PVL,
			"PVL missing version number: %v", err)
	}
	if version != PvlSupportedVersion {
		return NewProofError(keybase1.ProofStatus_INVALID_PVL,
			"PVL is for the wrong version %v != %v", version, PvlSupportedVersion)
	}

	scripts, perr := pvlChunkGetScripts(pvl, service)
	if perr != nil {
		return perr
	}

	// Scan all scripts (for the service) for errors. Report the first error.
	var errs []ProofError
	for _, script := range scripts {
		perr = pvlValidateScript(script, service)
		errs = append(errs, perr)
	}
	return errs[0]
}

func pvlValidateScript(script *jsonw.Wrapper, service keybase1.ProofType) ProofError {
	// Scan the script.
	// Does not validate each instruction's format. (That is done when running it)

	var modeknown = false
	var mode PvlMode
	if service == keybase1.ProofType_DNS {
		modeknown = true
		mode = PvlModeDNS
	}
	scriptlen, err := script.Len()
	if err != nil {
		return NewProofError(keybase1.ProofStatus_INVALID_PVL,
			"Could not get length of script: %v", err)
	}
	if scriptlen < 1 {
		return NewProofError(keybase1.ProofStatus_INVALID_PVL,
			"Empty script")
	}

	for i := 0; i < scriptlen; i++ {
		ins := script.AtIndex(i)
		switch {
		case pvlJSONHasKey(ins, PvlAssertRegexMatch):
		case pvlJSONHasKey(ins, PvlAssertFindBase64):
		case pvlJSONHasKey(ins, PvlWhitespaceNormalize):
		case pvlJSONHasKey(ins, PvlRegexCapture):

		case pvlJSONHasKey(ins, PvlFetch):
			// A script can contain only <=1 fetches.
			// A DNS script cannot contain fetches.

			fetchType, err := ins.AtKey(PvlFetch).GetString()
			if err != nil {
				return NewProofError(keybase1.ProofStatus_INVALID_PVL,
					"Could not get fetch type %v", i)
			}

			if service == keybase1.ProofType_DNS {
				return NewProofError(keybase1.ProofStatus_INVALID_PVL,
					"DNS script cannot contain fetch instruction")
			}
			if modeknown {
				return NewProofError(keybase1.ProofStatus_INVALID_PVL,
					"Script cannot contain multiple fetch instructions")
			}
			switch PvlMode(fetchType) {
			case PvlModeString:
				modeknown = true
				mode = PvlModeString
			case PvlModeHTML:
				modeknown = true
				mode = PvlModeHTML
			case PvlModeJSON:
				modeknown = true
				mode = PvlModeJSON
			default:
				return NewProofError(keybase1.ProofStatus_INVALID_PVL,
					"Unsupported fetch type: %v", fetchType)
			}
		case pvlJSONHasKey(ins, PvlSelectorJSON):
			// Can only select after fetching.
			switch {
			case service == keybase1.ProofType_DNS:
				return NewProofError(keybase1.ProofStatus_INVALID_PVL,
					"DNS script cannot use json selector")
			case !modeknown:
				return NewProofError(keybase1.ProofStatus_INVALID_PVL,
					"Script cannot select before fetch")
			case mode != PvlModeJSON:
				return NewProofError(keybase1.ProofStatus_INVALID_PVL,
					"Script contains json selector in non-html mode")
			}
		case pvlJSONHasKey(ins, PvlSelectorCSS):
			// Can only select after fetching.
			switch {
			case service == keybase1.ProofType_DNS:
				return NewProofError(keybase1.ProofStatus_INVALID_PVL,
					"DNS script cannot use css selector")
			case !modeknown:
				return NewProofError(keybase1.ProofStatus_INVALID_PVL,
					"Script cannot select before fetch")
			case mode != PvlModeHTML:
				return NewProofError(keybase1.ProofStatus_INVALID_PVL,
					"Script contains css selector in non-html mode")
			}
		case pvlJSONHasKey(ins, PvlTransformURL):
			// Can only transform before fetching.
			switch {
			case service == keybase1.ProofType_DNS:
				return NewProofError(keybase1.ProofStatus_INVALID_PVL,
					"DNS script cannot transform url")
			case modeknown:
				return NewProofError(keybase1.ProofStatus_INVALID_PVL,
					"Script cannot transform after fetch")
			}
		default:
			return NewProofError(keybase1.ProofStatus_INVALID_PVL,
				"Unsupported PVL instruction %d", i)
		}
	}

	return nil
}

// Run each script on each TXT record of each domain.
// Succeed if any succeed.
func pvlRunDNS(g *GlobalContext, scripts []*jsonw.Wrapper, startstate PvlScriptState) ProofError {
	userdomain := startstate.Vars.Hostname
	domains := []string{userdomain, "_keybase." + userdomain}
	var firsterr ProofError
	for _, d := range domains {
		g.Log.Debug("Trying DNS: %v", d)

		err := pvlRunDNSOne(g, scripts, startstate, d)
		if err == nil {
			return nil
		}
		if firsterr == nil {
			firsterr = err
		}
	}

	return firsterr
}

func pvlRunDNSOne(g *GlobalContext, scripts []*jsonw.Wrapper, startstate PvlScriptState, domain string) ProofError {
	txts, err := net.LookupTXT(domain)
	if err != nil {
		return NewProofError(keybase1.ProofStatus_DNS_ERROR,
			"DNS failure for %s: %s", domain, err)
	}

	for _, record := range txts {
		g.Log.Debug("For %s, got TXT record: %s", domain, record)

		// Try all scripts.
		for _, script := range scripts {
			state := startstate
			state.ActiveString = record
			err = pvlRunScript(g, script, state)
			if err == nil {
				return nil
			}
		}
	}

	return NewProofError(keybase1.ProofStatus_NOT_FOUND,
		"Checked %d TXT entries of %s, but didn't find signature",
		len(txts), domain)
}

func pvlRunScript(g *GlobalContext, script *jsonw.Wrapper, startstate PvlScriptState) ProofError {
	var state = startstate
	scriptlen, err := script.Len()
	if err != nil {
		return NewProofError(keybase1.ProofStatus_INVALID_PVL,
			"Could not get length of script: %v", err)
	}
	if scriptlen < 1 {
		return NewProofError(keybase1.ProofStatus_INVALID_PVL,
			"Empty scripts are not allowed.")
	}
	for i := 0; i < scriptlen; i++ {
		ins := script.AtIndex(i)

		// Sanity check.
		if int(state.PC) != i {
			return NewProofError(keybase1.ProofStatus_INVALID_PVL,
				fmt.Sprintf("Execution failure, PC mismatch %v %v", state.PC, i))
		}

		newstate, perr := pvlStepInstruction(g, ins, state)
		state = newstate
		if perr != nil {
			if perr.GetProofStatus() == keybase1.ProofStatus_INVALID_PVL {
				perr = NewProofError(keybase1.ProofStatus_INVALID_PVL,
					fmt.Sprintf("Invalid PVL (%v): %v", state.PC, perr.GetDesc()))
			}
			return perr
		}
		state.PC++
	}

	// Script executed successfully and with no errors.
	return nil
}

func pvlStepInstruction(g *GlobalContext, ins *jsonw.Wrapper, state PvlScriptState) (PvlScriptState, ProofError) {
	for name, step := range PvlSteps {
		if pvlJSONHasKey(ins, name) {
			return step(g, ins, state)
		}
	}

	return state, NewProofError(keybase1.ProofStatus_INVALID_PVL,
		"Unsupported PVL instruction %d", state.PC)
}

func pvlStepAssertRegexMatch(g *GlobalContext, ins *jsonw.Wrapper, state PvlScriptState) (PvlScriptState, ProofError) {
	template, err := ins.AtKey(PvlAssertRegexMatch).GetString()
	if err != nil {
		return state, NewProofError(keybase1.ProofStatus_INVALID_PVL,
			"Could not get pattern %v", ins)
	}
	re, perr := pvlInterpretRegex(template, state.Vars)
	if perr != nil {
		return state, perr
	}
	if !re.MatchString(state.ActiveString) {
		g.Log.Debug("PVL regex did not match: %v %v", re, state.ActiveString)
		return state, NewProofError(keybase1.ProofStatus_CONTENT_FAILURE,
			"Regex did not match %v", re)
	}

	return state, nil
}

func pvlStepAssertFindBase64(g *GlobalContext, ins *jsonw.Wrapper, state PvlScriptState) (PvlScriptState, ProofError) {
	target, err := ins.AtKey(PvlAssertFindBase64).GetString()
	if err != nil {
		return state, NewProofError(keybase1.ProofStatus_INVALID_PVL,
			"Could not assert target %v", ins)
	}
	if target == "sig" {
		if FindBase64Block(state.ActiveString, state.Vars.Sig, false) {
			return state, nil
		}
		return state, NewProofError(keybase1.ProofStatus_TEXT_NOT_FOUND,
			"Signature not found")
	}
	return state, NewProofError(keybase1.ProofStatus_INVALID_PVL,
		"Can only assert_find_base64 for sig")
}

func pvlStepWhitespaceNormalize(g *GlobalContext, ins *jsonw.Wrapper, state PvlScriptState) (PvlScriptState, ProofError) {
	state.ActiveString = WhitespaceNormalize(state.ActiveString)
	return state, nil
}

func pvlStepRegexCapture(g *GlobalContext, ins *jsonw.Wrapper, state PvlScriptState) (PvlScriptState, ProofError) {
	template, err := ins.AtKey(PvlRegexCapture).GetString()
	if err != nil {
		return state, NewProofError(keybase1.ProofStatus_INVALID_PVL,
			"Could not get pattern %v", ins)
	}
	re, perr := pvlInterpretRegex(template, state.Vars)
	if perr != nil {
		return state, perr
	}
	match := re.FindStringSubmatch(state.ActiveString)
	if len(match) < 2 {
		g.Log.Debug("PVL regex capture did not match: %v %v", re, state.ActiveString)
		return state, NewProofError(keybase1.ProofStatus_CONTENT_FAILURE,
			"Regex capture did not match: %v", re)
	}
	state.ActiveString = match[1]
	return state, nil
}

func pvlStepFetch(g *GlobalContext, ins *jsonw.Wrapper, state PvlScriptState) (PvlScriptState, ProofError) {
	fetchType, err := ins.AtKey(PvlFetch).GetString()
	if err != nil {
		return state, NewProofError(keybase1.ProofStatus_INVALID_PVL,
			"Could not get fetch type")
	}
	if state.FetchResult != nil {
		return state, NewProofError(keybase1.ProofStatus_INVALID_PVL,
			"Script cannot contain more than one fetch")
	}
	if state.Service == keybase1.ProofType_DNS {
		return state, NewProofError(keybase1.ProofStatus_INVALID_PVL,
			"Script cannot fetch for DNS")
	}

	switch PvlMode(fetchType) {
	case PvlModeString:
		res, err := g.XAPI.GetText(NewAPIArg(g, state.FetchURL))
		if err != nil {
			return state, XapiError(err, state.FetchURL)
		}
		state.FetchResult = &PvlFetchResult{
			Mode:   PvlModeString,
			String: res.Body,
		}
		state.ActiveString = state.FetchResult.String
		return state, nil
	case PvlModeJSON:
		res, err := g.XAPI.Get(NewAPIArg(g, state.FetchURL))
		if err != nil {
			return state, XapiError(err, state.FetchURL)
		}
		state.FetchResult = &PvlFetchResult{
			Mode: PvlModeJSON,
			JSON: res.Body,
		}
		state.ActiveString = ""
		return state, nil
	case PvlModeHTML:
		res, err := g.XAPI.GetHTML(NewAPIArg(g, state.FetchURL))
		if err != nil {
			return state, XapiError(err, state.FetchURL)
		}
		state.FetchResult = &PvlFetchResult{
			Mode: PvlModeHTML,
			HTML: res.GoQuery,
		}
		state.ActiveString = ""
		return state, nil
	default:
		return state, NewProofError(keybase1.ProofStatus_INVALID_PVL,
			"Unsupported fetch type %v", fetchType)
	}
}

func pvlStepSelectorJSON(g *GlobalContext, ins *jsonw.Wrapper, state PvlScriptState) (PvlScriptState, ProofError) {
	if state.FetchResult == nil || state.FetchResult.Mode != PvlModeJSON {
		return state, NewProofError(keybase1.ProofStatus_INVALID_PVL,
			"Cannot use json selector with non-json fetch result")
	}

	selectorsw, err := ins.AtKey(PvlSelectorJSON).ToArray()
	if err != nil {
		return state, NewProofError(keybase1.ProofStatus_INVALID_PVL,
			"Cannot use css selector with non-html fetch result")
	}

	selectors, err := pvlJSONUnpackArray(selectorsw)
	if err != nil {
		return state, NewProofError(keybase1.ProofStatus_INVALID_PVL,
			"Could not unpack json selector list: %v", err)
	}
	if len(selectors) < 1 {
		return state, NewProofError(keybase1.ProofStatus_INVALID_PVL,
			"Json selector list must contain at least 1 element")
	}

	results, perr := pvlRunSelectorJSONInner(g, state.FetchResult.JSON, selectors)
	if perr != nil {
		return state, perr
	}
	if len(results) < 1 {
		return state, NewProofError(keybase1.ProofStatus_CONTENT_FAILURE,
			"Json selector did not match any values")
	}
	s := strings.Join(results, " ")

	state.ActiveString = s
	return state, nil
}

func pvlStepSelectorCSS(g *GlobalContext, ins *jsonw.Wrapper, state PvlScriptState) (PvlScriptState, ProofError) {
	if state.FetchResult == nil || state.FetchResult.Mode != PvlModeHTML {
		return state, NewProofError(keybase1.ProofStatus_INVALID_PVL,
			"Cannot use css selector with non-html fetch result")
	}

	selectors, err := ins.AtKey(PvlSelectorCSS).ToArray()
	if err != nil {
		return state, NewProofError(keybase1.ProofStatus_INVALID_PVL,
			"CSS selectors must be an array: %v", err)
	}

	attr, err := ins.AtKey("attr").GetString()
	useAttr := err == nil

	selection, perr := pvlRunCSSSelector(g, state.FetchResult.HTML.Selection, selectors)
	if perr != nil {
		return state, perr
	}

	if selection.Size() < 1 {
		return state, NewProofError(keybase1.ProofStatus_CONTENT_FAILURE,
			"No elements matched by selector")
	}

	res, err := pvlSelectionContents(selection, useAttr, attr)
	if err != nil {
		return state, NewProofError(keybase1.ProofStatus_CONTENT_FAILURE,
			"Could not get html for selection: %v", err)
	}

	state.ActiveString = res
	return state, nil
}

func pvlStepTransformURL(g *GlobalContext, ins *jsonw.Wrapper, state PvlScriptState) (PvlScriptState, ProofError) {
	sourceTemplate, err := ins.AtKey(PvlTransformURL).GetString()
	if err != nil {
		return state, NewProofError(keybase1.ProofStatus_INVALID_PVL,
			"Could not get source pattern %v", ins)
	}
	destTemplate, err := ins.AtKey("to").GetString()
	if err != nil {
		return state, NewProofError(keybase1.ProofStatus_INVALID_PVL,
			"Could not get dest pattern %v", ins)
	}

	re, perr := pvlInterpretRegex(sourceTemplate, state.Vars)
	if perr != nil {
		return state, perr
	}

	match := re.FindStringSubmatch(state.FetchURL)
	if len(match) < 1 {
		g.Log.Debug("PVL regex transform did not match: %v %v", re, state.FetchURL)
		return state, NewProofError(keybase1.ProofStatus_CONTENT_FAILURE,
			"Regex transform did not match: %v", re)
	}

	newURL, err := pvlSubstitute(destTemplate, state.Vars, match)
	if err != nil {
		g.Log.Debug("PVL regex transform did not substitute: %v %v", re, state.FetchURL)
		return state, NewProofError(keybase1.ProofStatus_BAD_API_URL,
			"Regex transform did not substitute: %v %v", re, err)
	}

	state.FetchURL = newURL
	state.ActiveString = newURL
	return state, nil
}

// Run a PVL CSS selector.
// selectors is a list like [ "div .foo", 0, ".bar"] ].
// Each string runs a selector, each integer runs a Eq.
func pvlRunCSSSelector(g *GlobalContext, html *goquery.Selection, selectors *jsonw.Wrapper) (*goquery.Selection, ProofError) {
	nselectors, err := selectors.Len()
	if err != nil {
		return nil, NewProofError(keybase1.ProofStatus_INVALID_PVL,
			"Could not get length of selector list")
	}
	if nselectors < 1 {
		return nil, NewProofError(keybase1.ProofStatus_INVALID_PVL,
			"CSS selectors array must not be empty")
	}

	var selection *goquery.Selection
	selection = html

	for i := 0; i < nselectors; i++ {
		selector := selectors.AtIndex(i)

		selectorIndex, err := selector.GetInt()
		selectorIsIndex := err == nil
		selectorString, err := selector.GetString()
		selectorIsString := err == nil && !selectorIsIndex

		switch {
		case selectorIsIndex:
			selection = selection.Eq(selectorIndex)
		case selectorIsString:
			selection = selection.Find(selectorString)
		default:
			return nil, NewProofError(keybase1.ProofStatus_INVALID_PVL,
				"Selector entry string or int %v", selector)
		}
	}

	return selection, nil
}

// Most failures here log instead of returning an error. If an error occurs, ([], nil) will be returned.
// This is because a selector may descend into many subtrees and fail in all but one.
func pvlRunSelectorJSONInner(g *GlobalContext, object *jsonw.Wrapper, selectors []*jsonw.Wrapper) ([]string, ProofError) {
	if len(selectors) == 0 {
		s, err := pvlJSONStringOrMarshal(object)
		if err != nil {
			g.Log.Debug("PVL could not read object: %v", err)
			return make([]string, 0), nil
		}
		return []string{s}, nil
	}

	selector := selectors[0]
	nextselectors := selectors[1:]

	selectorIndex, err := selector.GetInt()
	selectorIsIndex := err == nil
	selectorKey, err := selector.GetString()
	selectorIsKey := err == nil && !selectorIsIndex
	allness, err := selector.AtKey("all").GetBool()
	selectorIsAll := err == nil && allness

	switch {
	case selectorIsIndex:
		object, err := object.ToArray()
		if err != nil {
			g.Log.Debug("PVL json select by index from non-array: %v", err)
			return []string{}, nil
		}

		nextobject := object.AtIndex(selectorIndex)
		return pvlRunSelectorJSONInner(g, nextobject, nextselectors)
	case selectorIsKey:
		object, err := object.ToDictionary()
		if err != nil {
			g.Log.Debug("PVL json select by key from non-map: %v", err)
			return []string{}, nil
		}

		nextobject := object.AtKey(selectorKey)
		return pvlRunSelectorJSONInner(g, nextobject, nextselectors)
	case selectorIsAll:
		children, err := pvlJSONGetChildren(object)
		if err != nil {
			g.Log.Debug("PVL json select could not get children: %v", err)
			return []string{}, nil
		}
		var results []string
		for _, child := range children {
			innerresults, perr := pvlRunSelectorJSONInner(g, child, nextselectors)
			if perr != nil {
				return nil, perr
			}
			results = append(results, innerresults...)
		}
	}
	return []string{}, NewProofError(keybase1.ProofStatus_INVALID_PVL,
		"Selector entry not recognized: %v", selector)
}

func pvlInterpretRegex(template string, vars PvlScriptVariables) (*regexp.Regexp, ProofError) {
	perr := NewProofError(keybase1.ProofStatus_INVALID_PVL,
		"Could not build regex %v", template)

	// Parse out side bars and option letters.
	if !strings.HasPrefix(template, "/") {
		return nil, perr
	}
	lastSlash := strings.LastIndex(template, "/")
	if lastSlash == -1 {
		return nil, perr
	}
	opts := template[lastSlash+1:]
	if !regexp.MustCompile("[imsU]*").MatchString(opts) {
		return nil, perr
	}
	var prefix = ""
	if len(opts) > 0 {
		prefix = "(?" + opts + ")"
	}

	// Do variable interpolation.
	prepattern, err := pvlSubstitute(template[1:lastSlash], vars, nil)
	if err != nil {
		return nil, NewProofError(keybase1.ProofStatus_BAD_API_URL, err.Error())
	}
	pattern := prefix + prepattern

	// Build the regex.
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, NewProofError(keybase1.ProofStatus_INVALID_PVL,
			"Could not compile regex (%v): %v", template, err)
	}
	return re, nil
}

// Substitute vars for %{name} in the string.
// Only substitutes whitelisted variables.
// It is an error to refer to an undefined numbered group. But not an error to refer to a poorly-named variable.
// Match is an optional slice which is a regex match.
func pvlSubstitute(template string, vars PvlScriptVariables, match []string) (string, error) {
	var outerr error
	// Regex to find %{name} occurrences.
	re := regexp.MustCompile("%{[\\w0-9]+}")
	pvlSubstituteOne := func(vartag string) string {
		// Strip off the %, {, and }
		varname := vartag[2 : len(vartag)-1]
		var value string
		switch varname {
		case "username_service":
			value = vars.UsernameService
		case "username_keybase":
			value = vars.UsernameKeybase
		case "sig":
			value = b64.StdEncoding.EncodeToString(vars.Sig)
		case "sig_id_medium":
			value = vars.SigIDMedium
		case "sig_id_short":
			value = vars.SigIDShort
		case "hostname":
			value = vars.Hostname
		default:
			var i int
			i, err := strconv.Atoi(varname)
			if err == nil {
				if i >= 0 && i < len(match) {
					value = match[i]
				} else {
					outerr = fmt.Errorf("Substitution argument %v out of range of match", i)
				}
			} else {
				// Unrecognized variable, do no substitution.
				return vartag
			}
		}
		return regexp.QuoteMeta(value)
	}
	res := re.ReplaceAllStringFunc(template, pvlSubstituteOne)
	if outerr != nil {
		return template, outerr
	}
	return res, nil
}
