// Copyright 2022 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/moov-io/base/log"
	"github.com/moov-io/watchman/pkg/csl"
	"github.com/moov-io/watchman/pkg/dpl"
	"github.com/moov-io/watchman/pkg/ofac"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
)

func TestSearch__Address(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/search?address=ibex+house+minories&limit=1", nil)

	router := mux.NewRouter()
	addSearchRoutes(log.NewNopLogger(), router, addressSearcher)
	router.ServeHTTP(w, req)
	w.Flush()

	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), `"match":0.88194`)

	var wrapper struct {
		Addresses []*ofac.Address `json:"addresses"`
	}
	if err := json.NewDecoder(w.Body).Decode(&wrapper); err != nil {
		t.Fatal(err)
	}
	if len(wrapper.Addresses) == 0 {
		t.Fatal("found no addresses")
	}
	if wrapper.Addresses[0].EntityID != "173" {
		t.Errorf("%#v", wrapper.Addresses[0])
	}

	// send an empty body and get an error
	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/search?limit=1", nil)
	router.ServeHTTP(w, req)
	w.Flush()

	if w.Code != http.StatusBadRequest {
		t.Errorf("bogus status code: %d", w.Code)
	}
}

func TestSearch__AddressCountry(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/search?country=united+kingdom&limit=1", nil)

	router := mux.NewRouter()
	addSearchRoutes(log.NewNopLogger(), router, addressSearcher)
	router.ServeHTTP(w, req)
	w.Flush()

	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), `"match":1`)

}

func TestSearch__AddressMulti(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/search?address=ibex+house&country=united+kingdom&limit=1", nil)

	router := mux.NewRouter()
	addSearchRoutes(log.NewNopLogger(), router, addressSearcher)
	router.ServeHTTP(w, req)
	w.Flush()

	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), `"match":0.8847`)

}

func TestSearch__AddressProvidence(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/search?address=ibex+house&country=united+kingdom&providence=london+ec3n+1DY&limit=1", nil)

	router := mux.NewRouter()
	addSearchRoutes(log.NewNopLogger(), router, addressSearcher)
	router.ServeHTTP(w, req)
	w.Flush()

	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), `"match":0.923`)

}

func TestSearch__AddressCity(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/search?address=ibex+house&country=united+kingdom&city=london+ec3n+1DY&limit=1", nil)

	router := mux.NewRouter()
	addSearchRoutes(log.NewNopLogger(), router, addressSearcher)
	router.ServeHTTP(w, req)
	w.Flush()

	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), `"match":0.923`)

}

func TestSearch__AddressState(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/search?address=ibex+house&country=united+kingdom&state=london+ec3n+1DY&limit=1", nil)

	router := mux.NewRouter()
	addSearchRoutes(log.NewNopLogger(), router, addressSearcher)
	router.ServeHTTP(w, req)
	w.Flush()

	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), `"match":0.923`)

}

func TestSearch__NameAndAddress(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/search?name=midco&address=rue+de+rhone&limit=1", nil)

	pipe := noLogPipeliner
	s := newSearcher(log.NewNopLogger(), pipe, 1)
	s.Addresses = precomputeAddresses([]*ofac.Address{
		{
			EntityID:                    "2831",
			AddressID:                   "1965",
			Address:                     "57 Rue du Rhone",
			CityStateProvincePostalCode: "Geneva CH-1204",
			Country:                     "Switzerland",
		},
		{
			EntityID:                    "173",
			AddressID:                   "129",
			Address:                     "Ibex House, The Minories",
			CityStateProvincePostalCode: "London EC3N 1DY",
			Country:                     "United Kingdom",
		},
	})
	s.SDNs = precomputeSDNs([]*ofac.SDN{
		{
			EntityID: "2831",
			SDNName:  "MIDCO FINANCE S.A.",
			SDNType:  "individual",
			Programs: []string{"IRAQ2"},
			Remarks:  "US FEIN CH-660-0-469-982-0 (United States); Switzerland.",
		},
	}, nil, noLogPipeliner)

	router := mux.NewRouter()
	addSearchRoutes(log.NewNopLogger(), router, s)
	router.ServeHTTP(w, req)
	w.Flush()

	if w.Code != http.StatusOK {
		t.Errorf("bogus status code: %d", w.Code)
	}
	var wrapper struct {
		SDNs      []*ofac.SDN     `json:"SDNs"`
		Addresses []*ofac.Address `json:"addresses"`
	}
	if err := json.NewDecoder(w.Body).Decode(&wrapper); err != nil {
		t.Fatal(err)
	}

	if len(wrapper.SDNs) != 1 || len(wrapper.Addresses) != 1 {
		t.Fatalf("sdns=%#v addresses=%#v", wrapper.SDNs[0], wrapper.Addresses[0])
	}

	if wrapper.SDNs[0].EntityID != "2831" || wrapper.Addresses[0].EntityID != "2831" {
		t.Errorf("SDNs[0].EntityID=%s Addresses[0].EntityID=%s", wrapper.SDNs[0].EntityID, wrapper.Addresses[0].EntityID)
	}

	// request with no results
	req = httptest.NewRequest("GET", "/search?name=midco&country=United+Kingdom&limit=1", nil)

	router.ServeHTTP(w, req)
	w.Flush()

	if w.Code != http.StatusOK {
		t.Errorf("bogus status code: %d", w.Code)
	}
	if err := json.NewDecoder(w.Body).Decode(&wrapper); err != nil {
		t.Fatal(err)
	}
	if len(wrapper.SDNs) != 1 || len(wrapper.Addresses) != 1 {
		t.Errorf("sdns=%#v", wrapper.SDNs[0])
		t.Fatalf("addresses=%#v", wrapper.Addresses[0])
	}
}

func TestSearch__NameAndAltName(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/search?limit=1&q=nayif", nil)

	s := newSearcher(log.NewNopLogger(), noLogPipeliner, 1)
	// OFAC
	s.SDNs = sdnSearcher.SDNs
	s.Alts = altSearcher.Alts
	s.Addresses = addressSearcher.Addresses
	s.SSIs = ssiSearcher.SSIs
	// BIS
	s.DPs = dplSearcher.DPs
	s.BISEntities = bisEntitySearcher.BISEntities

	router := mux.NewRouter()
	addSearchRoutes(log.NewNopLogger(), router, s)
	router.ServeHTTP(w, req)
	w.Flush()

	if w.Code != http.StatusOK {
		t.Errorf("bogus status code: %d", w.Code)
	}

	// read response body
	var wrapper struct {
		// OFAC
		SDNs              []*ofac.SDN               `json:"SDNs"`
		AltNames          []*ofac.AlternateIdentity `json:"altNames"`
		Addresses         []*ofac.Address           `json:"addresses"`
		SectoralSanctions []*csl.SSI                `json:"sectoralSanctions"`
		// BIS
		DeniedPersons []*dpl.DPL `json:"deniedPersons"`
		BISEntities   []*csl.EL  `json:"bisEntities"`
	}
	if err := json.NewDecoder(w.Body).Decode(&wrapper); err != nil {
		t.Fatal(err)
	}

	// OFAC
	require.Equal(t, "2681", wrapper.SDNs[0].EntityID)
	require.Equal(t, "HAWATMA, Nayif", wrapper.SDNs[0].SDNName)
	require.Equal(t, "559", wrapper.AltNames[0].EntityID)
	require.Equal(t, "735", wrapper.Addresses[0].EntityID)
	require.Equal(t, "18736", wrapper.SectoralSanctions[0].EntityID)

	// BIS
	require.Equal(t, "P.O. BOX 28360", wrapper.DeniedPersons[0].StreetAddress)
	require.Equal(t, "Mohammad Jan Khan Mangal", wrapper.BISEntities[0].Name)
}

func TestSearch__Name(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/search?name=Dr+AL+ZAWAHIRI&limit=1", nil)

	router := mux.NewRouter()
	combinedSearcher := newSearcher(log.NewNopLogger(), noLogPipeliner, 1)
	// OFAC
	combinedSearcher.SDNs = sdnSearcher.SDNs
	combinedSearcher.Alts = altSearcher.Alts
	combinedSearcher.SSIs = ssiSearcher.SSIs
	// BIS
	combinedSearcher.DPs = dplSearcher.DPs
	combinedSearcher.BISEntities = bisEntitySearcher.BISEntities

	addSearchRoutes(log.NewNopLogger(), router, combinedSearcher)
	router.ServeHTTP(w, req)
	w.Flush()

	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), `"match":0.95588`)
	require.Contains(t, w.Body.String(), `"matchedName":"dr ayman al zawahiri"`)

	var wrapper struct {
		// OFAC
		SDNs []*ofac.SDN               `json:"SDNs"`
		Alts []*ofac.AlternateIdentity `json:"altNames"`
		SSIs []*csl.SSI                `json:"sectoralSanctions"`
		// BIS
		DPs []*dpl.DPL `json:"deniedPersons"`
		ELs []*csl.EL  `json:"bisEntities"`
	}
	if err := json.NewDecoder(w.Body).Decode(&wrapper); err != nil {
		t.Fatal(err)
	}
	if len(wrapper.SDNs) != 1 || len(wrapper.Alts) != 1 || len(wrapper.SSIs) != 1 || len(wrapper.DPs) != 1 || len(wrapper.ELs) != 1 {
		t.Fatalf("SDNs=%d Alts=%d SSIs=%d DPs=%d ELs=%d",
			len(wrapper.SDNs), len(wrapper.Alts), len(wrapper.SSIs), len(wrapper.DPs), len(wrapper.ELs))
	}

	require.Equal(t, "2676", wrapper.SDNs[0].EntityID)
	require.Equal(t, "4691", wrapper.Alts[0].EntityID)

	require.Equal(t, "18736", wrapper.SSIs[0].EntityID)
	require.Equal(t, "AL NASER WINGS AIRLINES", wrapper.DPs[0].Name)
	require.Equal(t, "Luqman Yasin Yunus Shgragi", wrapper.ELs[0].Name)
}

func TestSearch__AltName(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/search?altName=SOGO+KENKYUSHO&limit=1", nil)

	router := mux.NewRouter()
	addSearchRoutes(log.NewNopLogger(), router, altSearcher)
	router.ServeHTTP(w, req)
	w.Flush()

	if w.Code != http.StatusOK {
		t.Errorf("bogus status code: %d", w.Code)
	}

	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), `"match":0.98`)
	require.Contains(t, w.Body.String(), `"matchedName":"i c sogo kenkyusho"`)

	var wrapper struct {
		Alts []*ofac.AlternateIdentity `json:"altNames"`
	}
	if err := json.NewDecoder(w.Body).Decode(&wrapper); err != nil {
		t.Fatal(err)
	}
	if wrapper.Alts[0].EntityID != "4691" {
		t.Errorf("%#v", wrapper.Alts[0])
	}
}

func TestSearch__ID(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/search?id=5892464&limit=2", nil)

	router := mux.NewRouter()
	addSearchRoutes(log.NewNopLogger(), router, idSearcher)
	router.ServeHTTP(w, req)
	w.Flush()

	if w.Code != http.StatusOK {
		t.Errorf("bogus status code: %d", w.Code)
	}

	if v := w.Body.String(); !strings.Contains(v, `"match":1`) {
		t.Error(v)
	}

	var wrapper struct {
		SDNs []*ofac.SDN `json:"SDNs"`
	}
	if err := json.NewDecoder(w.Body).Decode(&wrapper); err != nil {
		t.Fatal(err)
	}
	if wrapper.SDNs[0].EntityID != "22790" {
		t.Errorf("%#v", wrapper.SDNs[0])
	}
}

func TestSearch__EscapeQuery(t *testing.T) {
	req, err := http.NewRequest("GET", "/search?name=John%2BDoe", nil)
	require.NoError(t, err)

	name := req.URL.Query().Get("name")
	require.Equal(t, "John+Doe", name)

	name, _ = url.QueryUnescape(name)
	require.Equal(t, "John Doe", name)

	req, err = http.NewRequest("GET", "/search?name=John+Doe", nil)
	if err != nil {
		t.Fatal(err)
	}
	name = req.URL.Query().Get("name")
	require.Equal(t, "John Doe", name)

	name, _ = url.QueryUnescape(name)
	require.Equal(t, "John Doe", name)
}
