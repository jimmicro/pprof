package pprof

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHeapGCRequiresPOST(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/debug/heap/gc", nil)
	w := httptest.NewRecorder()
	heapGC(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("GET /debug/heap/gc = %d, want %d", w.Code, http.StatusMethodNotAllowed)
	}
}

func TestHeapGCReturns200(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/debug/heap/gc", nil)
	w := httptest.NewRecorder()
	heapGC(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("POST /debug/heap/gc = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestHeapFreeRequiresPOST(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/debug/heap/free", nil)
	w := httptest.NewRecorder()
	heapFree(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("GET /debug/heap/free = %d, want %d", w.Code, http.StatusMethodNotAllowed)
	}
}

func TestHeapFreeReturns200(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/debug/heap/free", nil)
	w := httptest.NewRecorder()
	heapFree(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("POST /debug/heap/free = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestHeapGCPercentRequiresPOST(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/debug/heap/gcp?v=50", nil)
	w := httptest.NewRecorder()
	heapGCPercent(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("GET /debug/heap/gcp = %d, want %d", w.Code, http.StatusMethodNotAllowed)
	}
}

func TestHeapGCPercentReturns200(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/debug/heap/gcp?v=50", nil)
	w := httptest.NewRecorder()
	heapGCPercent(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("POST /debug/heap/gcp?v=50 = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestHeapGCPercentInvalidParam(t *testing.T) {
	cases := []string{"", "abc", "1.5"}
	for _, v := range cases {
		req := httptest.NewRequest(http.MethodPost, "/debug/heap/gcp?v="+v, nil)
		w := httptest.NewRecorder()
		heapGCPercent(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("POST /debug/heap/gcp?v=%q = %d, want %d", v, w.Code, http.StatusBadRequest)
		}
	}
}

func TestHeapMemLimitRequiresPOST(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/debug/heap/memlimit?v=1073741824", nil)
	w := httptest.NewRecorder()
	heapMemLimit(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("GET /debug/heap/memlimit = %d, want %d", w.Code, http.StatusMethodNotAllowed)
	}
}

func TestHeapMemLimitReturns200(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/debug/heap/memlimit?v=1073741824", nil)
	w := httptest.NewRecorder()
	heapMemLimit(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("POST /debug/heap/memlimit?v=1073741824 = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestHeapMemLimitInvalidParam(t *testing.T) {
	cases := []string{"", "abc", "1.5"}
	for _, v := range cases {
		req := httptest.NewRequest(http.MethodPost, "/debug/heap/memlimit?v="+v, nil)
		w := httptest.NewRecorder()
		heapMemLimit(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("POST /debug/heap/memlimit?v=%q = %d, want %d", v, w.Code, http.StatusBadRequest)
		}
	}
}
